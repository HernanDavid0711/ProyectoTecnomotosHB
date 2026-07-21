package motos

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, input CrearMotoEntrada) (*Moto, error) {
	const query = `
		INSERT INTO motos (
			placa,
			marca,
			modelo,
			cilindraje,
			anio,
			color,
			kilometraje_actual,
			cliente_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8::BIGINT)
		RETURNING id::text
	`

	var id string

	err := r.db.QueryRow(
		ctx,
		query,
		input.Placa,
		input.Marca,
		input.Modelo,
		input.Cilindraje,
		nullableInt(input.Anio),
		nullableString(input.Color),
		input.KilometrajeActual,
		input.ClienteID,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Moto, error) {
	const query = `
		SELECT
			m.id::text,
			COALESCE(m.placa, ''),
			m.marca,
			m.modelo,
			m.cilindraje,
			COALESCE(m.anio, 0),
			COALESCE(m.color, ''),
			m.kilometraje_actual,
			COALESCE(m.cliente_id::text, ''),
			COALESCE(CONCAT_WS(' ', c.nombres, c.apellidos), ''),
			m.creado_en,
			m.actualizado_en
		FROM motos m
		LEFT JOIN clientes c ON c.id = m.cliente_id
		WHERE m.id = $1 AND m.activa = TRUE
	`

	var moto Moto

	err := r.db.QueryRow(ctx, query, id).Scan(
		&moto.ID,
		&moto.Placa,
		&moto.Marca,
		&moto.Modelo,
		&moto.Cilindraje,
		&moto.Anio,
		&moto.Color,
		&moto.KilometrajeActual,
		&moto.ClienteID,
		&moto.NombreCliente,
		&moto.CreadoEn,
		&moto.ActualizadoEn,
	)
	if err != nil {
		return nil, err
	}

	return &moto, nil
}

func (r *Repository) GetByPlaca(ctx context.Context, placa string) (*Moto, error) {
	const query = `
		SELECT
			m.id::text,
			COALESCE(m.placa, ''),
			m.marca,
			m.modelo,
			m.cilindraje,
			COALESCE(m.anio, 0),
			COALESCE(m.color, ''),
			m.kilometraje_actual,
			COALESCE(m.cliente_id::text, ''),
			COALESCE(CONCAT_WS(' ', c.nombres, c.apellidos), ''),
			m.creado_en,
			m.actualizado_en
		FROM motos m
		LEFT JOIN clientes c ON c.id = m.cliente_id
		WHERE m.placa = $1 AND m.activa = TRUE
	`

	var moto Moto

	err := r.db.QueryRow(ctx, query, placa).Scan(
		&moto.ID,
		&moto.Placa,
		&moto.Marca,
		&moto.Modelo,
		&moto.Cilindraje,
		&moto.Anio,
		&moto.Color,
		&moto.KilometrajeActual,
		&moto.ClienteID,
		&moto.NombreCliente,
		&moto.CreadoEn,
		&moto.ActualizadoEn,
	)
	if err != nil {
		return nil, err
	}

	return &moto, nil
}

func (r *Repository) List(ctx context.Context, busqueda string, limit, offset int) ([]Moto, error) {
	const query = `
		SELECT
			m.id::text,
			COALESCE(m.placa, ''),
			m.marca,
			m.modelo,
			m.cilindraje,
			COALESCE(m.anio, 0),
			COALESCE(m.color, ''),
			m.kilometraje_actual,
			COALESCE(m.cliente_id::text, ''),
			COALESCE(CONCAT_WS(' ', c.nombres, c.apellidos), ''),
			m.creado_en,
			m.actualizado_en
		FROM motos m
		LEFT JOIN clientes c ON c.id = m.cliente_id
		WHERE
			m.activa = TRUE
			AND (
				$1 = ''
				OR m.placa ILIKE '%' || $1 || '%'
				OR m.marca ILIKE '%' || $1 || '%'
				OR m.modelo ILIKE '%' || $1 || '%'
				OR m.color ILIKE '%' || $1 || '%'
				OR CONCAT_WS(' ', c.nombres, c.apellidos) ILIKE '%' || $1 || '%'
			)
		ORDER BY m.creado_en DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, busqueda, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	motos := make([]Moto, 0)

	for rows.Next() {
		var moto Moto

		if err := rows.Scan(
			&moto.ID,
			&moto.Placa,
			&moto.Marca,
			&moto.Modelo,
			&moto.Cilindraje,
			&moto.Anio,
			&moto.Color,
			&moto.KilometrajeActual,
			&moto.ClienteID,
			&moto.NombreCliente,
			&moto.CreadoEn,
			&moto.ActualizadoEn,
		); err != nil {
			return nil, err
		}

		motos = append(motos, moto)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return motos, nil
}

func (r *Repository) AutocompleteByPlaca(ctx context.Context, placa string, limit int) ([]MotoAutocompleteItem, error) {
	const query = `
		SELECT
			m.id::text,
			COALESCE(m.placa, ''),
			m.marca,
			m.modelo,
			m.cilindraje,
			COALESCE(m.anio, 0),
			COALESCE(m.color, ''),
			m.kilometraje_actual,
			COALESCE(m.cliente_id::text, ''),
			COALESCE(CONCAT_WS(' ', c.nombres, c.apellidos), '')
		FROM motos m
		LEFT JOIN clientes c ON c.id = m.cliente_id
		WHERE m.activa = TRUE AND m.placa ILIKE $1 || '%'
		ORDER BY m.placa ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, placa, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]MotoAutocompleteItem, 0)

	for rows.Next() {
		var item MotoAutocompleteItem

		if err := rows.Scan(
			&item.ID,
			&item.Placa,
			&item.Marca,
			&item.Modelo,
			&item.Cilindraje,
			&item.Anio,
			&item.Color,
			&item.KilometrajeActual,
			&item.ClienteID,
			&item.NombreCliente,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) Update(ctx context.Context, id string, input ActualizarMotoEntrada) (*Moto, error) {
	const query = `
		UPDATE motos
		SET
			placa = $2,
			marca = $3,
			modelo = $4,
			cilindraje = $5,
			anio = $6,
			color = $7,
			kilometraje_actual = $8,
			cliente_id = $9::BIGINT,
			actualizado_en = NOW()
		WHERE id = $1 AND activa = TRUE
		RETURNING id::text
	`

	var updatedID string

	err := r.db.QueryRow(
		ctx,
		query,
		id,
		input.Placa,
		input.Marca,
		input.Modelo,
		input.Cilindraje,
		nullableInt(input.Anio),
		nullableString(input.Color),
		input.KilometrajeActual,
		input.ClienteID,
	).Scan(&updatedID)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, updatedID)
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}
