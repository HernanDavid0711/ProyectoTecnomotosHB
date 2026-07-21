package clientes

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

func (r *Repository) Create(ctx context.Context, input CrearClienteEntrada) (*Cliente, error) {
	const query = `
		INSERT INTO clientes (
			nombres,
			apellidos,
			telefono,
			correo,
			direccion,
			tipo_documento,
			numero_documento,
			notas,
			activo
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, TRUE)
		RETURNING id::text
	`

	var id string

	err := r.db.QueryRow(
		ctx,
		query,
		input.Nombres,
		input.Apellidos,
		nullableString(input.Telefono),
		nullableString(input.Correo),
		nullableString(input.Direccion),
		nullableString(input.TipoDocumento),
		nullableString(input.NumeroDocumento),
		nullableString(input.Notas),
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Cliente, error) {
	const query = `
		SELECT
			id::text,
			nombres,
			apellidos,
			COALESCE(telefono, ''),
			COALESCE(correo, ''),
			COALESCE(direccion, ''),
			COALESCE(tipo_documento, ''),
			COALESCE(numero_documento, ''),
			COALESCE(notas, ''),
			activo,
			creado_en,
			actualizado_en
		FROM clientes
		WHERE id = $1 AND activo = TRUE
	`

	var cliente Cliente

	err := r.db.QueryRow(ctx, query, id).Scan(
		&cliente.ID,
		&cliente.Nombres,
		&cliente.Apellidos,
		&cliente.Telefono,
		&cliente.Correo,
		&cliente.Direccion,
		&cliente.TipoDocumento,
		&cliente.NumeroDocumento,
		&cliente.Notas,
		&cliente.Activo,
		&cliente.CreadoEn,
		&cliente.ActualizadoEn,
	)
	if err != nil {
		return nil, err
	}

	return &cliente, nil
}

func (r *Repository) List(ctx context.Context, busqueda string, limit, offset int) ([]Cliente, error) {
	const query = `
		SELECT
			id::text,
			nombres,
			apellidos,
			COALESCE(telefono, ''),
			COALESCE(correo, ''),
			COALESCE(direccion, ''),
			COALESCE(tipo_documento, ''),
			COALESCE(numero_documento, ''),
			COALESCE(notas, ''),
			activo,
			creado_en,
			actualizado_en
		FROM clientes
		WHERE
			activo = TRUE
			AND (
				$1 = ''
				OR nombres ILIKE '%' || $1 || '%'
				OR apellidos ILIKE '%' || $1 || '%'
				OR COALESCE(telefono, '') ILIKE '%' || $1 || '%'
				OR COALESCE(correo, '') ILIKE '%' || $1 || '%'
				OR COALESCE(numero_documento, '') ILIKE '%' || $1 || '%'
			)
		ORDER BY creado_en DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, busqueda, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clientes := make([]Cliente, 0)

	for rows.Next() {
		var cliente Cliente

		if err := rows.Scan(
			&cliente.ID,
			&cliente.Nombres,
			&cliente.Apellidos,
			&cliente.Telefono,
			&cliente.Correo,
			&cliente.Direccion,
			&cliente.TipoDocumento,
			&cliente.NumeroDocumento,
			&cliente.Notas,
			&cliente.Activo,
			&cliente.CreadoEn,
			&cliente.ActualizadoEn,
		); err != nil {
			return nil, err
		}

		clientes = append(clientes, cliente)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clientes, nil
}

func (r *Repository) Update(ctx context.Context, id string, input ActualizarClienteEntrada) (*Cliente, error) {
	const query = `
		UPDATE clientes
		SET
			nombres = $2,
			apellidos = $3,
			telefono = $4,
			correo = $5,
			direccion = $6,
			tipo_documento = $7,
			numero_documento = $8,
			notas = $9,
			actualizado_en = NOW()
		WHERE id = $1 AND activo = TRUE
		RETURNING id::text
	`

	var updatedID string

	err := r.db.QueryRow(
		ctx,
		query,
		id,
		input.Nombres,
		input.Apellidos,
		nullableString(input.Telefono),
		nullableString(input.Correo),
		nullableString(input.Direccion),
		nullableString(input.TipoDocumento),
		nullableString(input.NumeroDocumento),
		nullableString(input.Notas),
	).Scan(&updatedID)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, updatedID)
}

func (r *Repository) Delete(ctx context.Context, id string) (bool, error) {
	const query = `
		UPDATE clientes
		SET
			activo = FALSE,
			actualizado_en = NOW()
		WHERE id = $1 AND activo = TRUE
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return false, err
	}

	return result.RowsAffected() > 0, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
