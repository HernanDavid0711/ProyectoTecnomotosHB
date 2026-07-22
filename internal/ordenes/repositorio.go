package ordenes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) BuscarClientesConMotos(ctx context.Context, busqueda string, limit int) ([]ClienteConMotos, error) {
	const query = `
		SELECT
			c.id::text,
			c.nombres,
			c.apellidos,
			CONCAT_WS(' ', c.nombres, c.apellidos),
			COALESCE(c.telefono, ''),
			COALESCE(c.correo, ''),
			COALESCE(c.tipo_documento, ''),
			COALESCE(c.numero_documento, ''),
			COALESCE(m.id::text, ''),
			COALESCE(m.placa, ''),
			COALESCE(m.marca, ''),
			COALESCE(m.modelo, ''),
			COALESCE(m.cilindraje, 0),
			COALESCE(m.anio, 0),
			COALESCE(m.color, ''),
			COALESCE(m.kilometraje_actual, 0)
		FROM clientes c
		LEFT JOIN motos m ON m.cliente_id = c.id AND m.activa = TRUE
		WHERE
			c.activo = TRUE
			AND (
				c.nombres ILIKE '%' || $1 || '%'
				OR c.apellidos ILIKE '%' || $1 || '%'
				OR CONCAT_WS(' ', c.nombres, c.apellidos) ILIKE '%' || $1 || '%'
				OR COALESCE(c.telefono, '') ILIKE '%' || $1 || '%'
				OR COALESCE(c.numero_documento, '') ILIKE '%' || $1 || '%'
				OR COALESCE(m.placa, '') ILIKE '%' || $1 || '%'
			)
		ORDER BY c.apellidos, c.nombres, m.placa
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, busqueda, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clientesByID := make(map[string]*ClienteConMotos)
	orderedIDs := make([]string, 0)

	for rows.Next() {
		var cliente ClienteConMotos
		var moto MotoResumen

		err := rows.Scan(
			&cliente.ID,
			&cliente.Nombres,
			&cliente.Apellidos,
			&cliente.NombreCompleto,
			&cliente.Telefono,
			&cliente.Correo,
			&cliente.TipoDocumento,
			&cliente.NumeroDocumento,
			&moto.ID,
			&moto.Placa,
			&moto.Marca,
			&moto.Modelo,
			&moto.Cilindraje,
			&moto.Anio,
			&moto.Color,
			&moto.KilometrajeActual,
		)
		if err != nil {
			return nil, err
		}

		current, exists := clientesByID[cliente.ID]
		if !exists {
			cliente.Motos = make([]MotoResumen, 0)
			clientesByID[cliente.ID] = &cliente
			orderedIDs = append(orderedIDs, cliente.ID)
			current = &cliente
		}

		if moto.ID != "" {
			current.Motos = append(current.Motos, moto)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	clientes := make([]ClienteConMotos, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		clientes = append(clientes, *clientesByID[id])
	}

	return clientes, nil
}

func (r *Repository) ListCatalogoItems(ctx context.Context, busqueda string, tipoItem string, limit int, offset int) ([]CatalogoItemTrabajo, error) {
	const query = `
		SELECT
			id::text,
			tipo_item,
			nombre,
			COALESCE(descripcion, ''),
			valor_base::float8,
			activo,
			creado_en,
			actualizado_en
		FROM catalogo_items_trabajo
		WHERE
			activo = TRUE
			AND ($1 = '' OR tipo_item = $1)
			AND (
				$2 = ''
				OR nombre ILIKE '%' || $2 || '%'
				OR COALESCE(descripcion, '') ILIKE '%' || $2 || '%'
			)
		ORDER BY tipo_item, nombre
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, tipoItem, busqueda, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]CatalogoItemTrabajo, 0)
	for rows.Next() {
		var item CatalogoItemTrabajo
		if err := rows.Scan(
			&item.ID,
			&item.TipoItem,
			&item.Nombre,
			&item.Descripcion,
			&item.ValorBase,
			&item.Activo,
			&item.CreadoEn,
			&item.ActualizadoEn,
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

func (r *Repository) AdminListCatalogoItems(ctx context.Context, busqueda string, tipoItem string, activo string, limit int, offset int) ([]CatalogoItemTrabajo, error) {
	const query = `
		SELECT
			id::text,
			tipo_item,
			nombre,
			COALESCE(descripcion, ''),
			valor_base::float8,
			activo,
			creado_en,
			actualizado_en
		FROM catalogo_items_trabajo
		WHERE
			($1 = '' OR tipo_item = $1)
			AND ($2 = '' OR activo = ($2 = 'true'))
			AND (
				$3 = ''
				OR nombre ILIKE '%' || $3 || '%'
				OR COALESCE(descripcion, '') ILIKE '%' || $3 || '%'
			)
		ORDER BY activo DESC, tipo_item, nombre
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.Query(ctx, query, tipoItem, activo, busqueda, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCatalogoItems(rows)
}

func (r *Repository) GetCatalogoItemByID(ctx context.Context, id string) (*CatalogoItemTrabajo, error) {
	const query = `
		SELECT
			id::text,
			tipo_item,
			nombre,
			COALESCE(descripcion, ''),
			valor_base::float8,
			activo,
			creado_en,
			actualizado_en
		FROM catalogo_items_trabajo
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	return scanCatalogoItem(row)
}

func (r *Repository) CreateCatalogoItem(ctx context.Context, input GuardarCatalogoItemTrabajoEntrada) (*CatalogoItemTrabajo, error) {
	const query = `
		INSERT INTO catalogo_items_trabajo (
			tipo_item,
			nombre,
			descripcion,
			valor_base,
			activo
		) VALUES ($1, $2, $3, 0, TRUE)
		RETURNING id::text
	`

	var id string
	if err := r.db.QueryRow(ctx, query, input.TipoItem, input.Nombre, nullableString(input.Descripcion)).Scan(&id); err != nil {
		return nil, err
	}

	return r.GetCatalogoItemByID(ctx, id)
}

func (r *Repository) UpdateCatalogoItem(ctx context.Context, id string, input GuardarCatalogoItemTrabajoEntrada) (*CatalogoItemTrabajo, error) {
	const query = `
		UPDATE catalogo_items_trabajo
		SET
			tipo_item = $2,
			nombre = $3,
			descripcion = $4,
			activo = COALESCE($5, activo),
			actualizado_en = NOW()
		WHERE id = $1
		RETURNING id::text
	`

	var updatedID string
	err := r.db.QueryRow(
		ctx,
		query,
		id,
		input.TipoItem,
		input.Nombre,
		nullableString(input.Descripcion),
		input.Activo,
	).Scan(&updatedID)
	if err != nil {
		return nil, err
	}

	return r.GetCatalogoItemByID(ctx, updatedID)
}

func (r *Repository) DeactivateCatalogoItem(ctx context.Context, id string) (bool, error) {
	const query = `
		UPDATE catalogo_items_trabajo
		SET activo = FALSE, actualizado_en = NOW()
		WHERE id = $1 AND activo = TRUE
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

func (r *Repository) ListEmpleadosAsignables(ctx context.Context, busqueda string, limit int, offset int) ([]EmpleadoAsignable, error) {
	const query = `
		SELECT
			u.id::text,
			u.nombre_usuario,
			u.nombres,
			u.apellidos,
			CONCAT_WS(' ', u.nombres, u.apellidos),
			u.correo
		FROM usuarios u
		INNER JOIN roles r ON r.id = u.rol_id
		WHERE
			u.activo = TRUE
			AND r.nombre = 'empleado'
			AND (
				$1 = ''
				OR u.nombre_usuario ILIKE '%' || $1 || '%'
				OR u.nombres ILIKE '%' || $1 || '%'
				OR u.apellidos ILIKE '%' || $1 || '%'
				OR u.correo ILIKE '%' || $1 || '%'
			)
		ORDER BY u.nombres, u.apellidos
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, busqueda, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	empleados := make([]EmpleadoAsignable, 0)
	for rows.Next() {
		var empleado EmpleadoAsignable
		if err := rows.Scan(
			&empleado.ID,
			&empleado.NombreUsuario,
			&empleado.Nombres,
			&empleado.Apellidos,
			&empleado.NombreCompleto,
			&empleado.Correo,
		); err != nil {
			return nil, err
		}
		empleados = append(empleados, empleado)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return empleados, nil
}

func (r *Repository) Create(ctx context.Context, input CrearOrdenEntrada, usuarioRecibeID string) (*OrdenTrabajo, error) {
	const query = `
		INSERT INTO ordenes_trabajo (
			cliente_id,
			moto_id,
			usuario_recibe_id,
			usuario_responsable_id,
			fecha_prometida,
			kilometraje_ingreso,
			nivel_combustible_porcentaje,
			tipo_servicio,
			descripcion_falla,
			observaciones,
			campos_formato
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb)
		RETURNING id::text
	`

	var id string
	err := r.db.QueryRow(
		ctx,
		query,
		input.ClienteID,
		input.MotoID,
		usuarioRecibeID,
		nullableString(input.UsuarioResponsableID),
		input.FechaPrometida,
		input.KilometrajeIngreso,
		input.NivelCombustiblePorcentaje,
		nullableString(input.TipoServicio),
		input.DescripcionFalla,
		nullableString(input.Observaciones),
		jsonMapString(input.CamposFormato),
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) CreateWithItems(ctx context.Context, input CrearOrdenConItemsEntrada, usuarioRecibeID string) (*OrdenTrabajo, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	ordenID, err := insertOrden(ctx, tx, CrearOrdenEntrada{
		ClienteID:                  input.ClienteID,
		MotoID:                     input.MotoID,
		UsuarioResponsableID:       input.UsuarioResponsableID,
		FechaPrometida:             input.FechaPrometida,
		KilometrajeIngreso:         input.KilometrajeIngreso,
		NivelCombustiblePorcentaje: input.NivelCombustiblePorcentaje,
		TipoServicio:               input.TipoServicio,
		DescripcionFalla:           input.DescripcionFalla,
		Observaciones:              input.Observaciones,
		CamposFormato:              input.CamposFormato,
	}, usuarioRecibeID)
	if err != nil {
		return nil, err
	}

	for index, item := range input.Items {
		if err := insertOrdenItem(ctx, tx, ordenID, item, index); err != nil {
			return nil, err
		}
	}

	if err := recalculateOrdenTotals(ctx, tx, ordenID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, ordenID)
}

func insertOrden(ctx context.Context, tx pgx.Tx, input CrearOrdenEntrada, usuarioRecibeID string) (string, error) {
	const query = `
		INSERT INTO ordenes_trabajo (
			cliente_id,
			moto_id,
			usuario_recibe_id,
			usuario_responsable_id,
			fecha_prometida,
			kilometraje_ingreso,
			nivel_combustible_porcentaje,
			tipo_servicio,
			descripcion_falla,
			observaciones,
			campos_formato
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb)
		RETURNING id::text
	`

	var id string
	err := tx.QueryRow(
		ctx,
		query,
		input.ClienteID,
		input.MotoID,
		usuarioRecibeID,
		nullableString(input.UsuarioResponsableID),
		input.FechaPrometida,
		input.KilometrajeIngreso,
		input.NivelCombustiblePorcentaje,
		nullableString(input.TipoServicio),
		input.DescripcionFalla,
		nullableString(input.Observaciones),
		jsonMapString(input.CamposFormato),
	).Scan(&id)
	return id, err
}

func insertOrdenItem(ctx context.Context, tx pgx.Tx, ordenID string, input CrearItemOrdenEntrada, position int) error {
	const query = `
		INSERT INTO items_orden_trabajo (
			orden_trabajo_id,
			catalogo_item_trabajo_id,
			tipo_item,
			descripcion,
			cantidad,
			valor_unitario,
			descuento,
			total_linea,
			posicion
		)
		SELECT
			$1,
			cit.id,
			cit.tipo_item,
			COALESCE(NULLIF(cit.descripcion, ''), cit.nombre),
			$3,
			CASE WHEN $4::NUMERIC > 0 THEN $4 ELSE cit.valor_base END,
			$5,
			GREATEST(($3 * (CASE WHEN $4::NUMERIC > 0 THEN $4 ELSE cit.valor_base END)) - $5, 0),
			$6
		FROM catalogo_items_trabajo cit
		WHERE cit.id = $2 AND cit.activo = TRUE
		RETURNING id
	`

	var id string
	err := tx.QueryRow(
		ctx,
		query,
		ordenID,
		input.CatalogoItemTrabajoID,
		input.Cantidad,
		input.ValorUnitario,
		input.Descuento,
		position,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: catalogo_item_trabajo_id %s", pgx.ErrNoRows, input.CatalogoItemTrabajoID)
	}
	return err
}

func recalculateOrdenTotals(ctx context.Context, tx pgx.Tx, ordenID string) error {
	const query = `
		UPDATE ordenes_trabajo
		SET
			subtotal = COALESCE((
				SELECT SUM(total_linea)
				FROM items_orden_trabajo
				WHERE orden_trabajo_id = $1
			), 0),
			total = COALESCE((
				SELECT SUM(total_linea)
				FROM items_orden_trabajo
				WHERE orden_trabajo_id = $1
			), 0) - descuento,
			actualizado_en = NOW()
		WHERE id = $1
	`

	_, err := tx.Exec(ctx, query, ordenID)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id string) (*OrdenTrabajo, error) {
	return r.getByQuery(ctx, `WHERE ot.id = $1`, id)
}

func (r *Repository) GetAssignedByID(ctx context.Context, id string, usuarioID string) (*OrdenTrabajo, error) {
	return r.getByQuery(ctx, `WHERE ot.id = $1 AND ot.usuario_responsable_id = $2`, id, usuarioID)
}

func (r *Repository) List(ctx context.Context, busqueda string, estado string, limit int, offset int) ([]OrdenTrabajo, error) {
	const query = selectOrdenBase + `
		WHERE
			($1 = '' OR ot.estado = $1)
			AND (
				$2 = ''
				OR ot.numero::text ILIKE '%' || $2 || '%'
				OR c.nombres ILIKE '%' || $2 || '%'
				OR c.apellidos ILIKE '%' || $2 || '%'
				OR COALESCE(m.placa, '') ILIKE '%' || $2 || '%'
				OR COALESCE(ur.nombres || ' ' || ur.apellidos, '') ILIKE '%' || $2 || '%'
			)
		ORDER BY ot.fecha_ingreso DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, estado, busqueda, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanOrdenes(rows)
}

func (r *Repository) ListAssigned(ctx context.Context, usuarioID string, estado string, limit int, offset int) ([]OrdenTrabajo, error) {
	const query = selectOrdenBase + `
		WHERE
			ot.usuario_responsable_id = $1
			AND ($2 = '' OR ot.estado = $2)
		ORDER BY ot.fecha_ingreso DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, usuarioID, estado, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanOrdenes(rows)
}

func (r *Repository) Assign(ctx context.Context, id string, usuarioResponsableID string) (*OrdenTrabajo, error) {
	const query = `
		UPDATE ordenes_trabajo
		SET usuario_responsable_id = $2, actualizado_en = NOW()
		WHERE id = $1
		RETURNING id::text
	`

	var updatedID string
	err := r.db.QueryRow(ctx, query, id, usuarioResponsableID).Scan(&updatedID)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, updatedID)
}

func (r *Repository) UpdateAdmin(ctx context.Context, id string, input ActualizarOrdenAdminEntrada) (*OrdenTrabajo, error) {
	const query = `
		UPDATE ordenes_trabajo
		SET
			usuario_responsable_id = $2,
			fecha_prometida = $3,
			kilometraje_ingreso = $4,
			nivel_combustible_porcentaje = $5,
			estado = $6,
			tipo_servicio = $7,
			descripcion_falla = $8,
			diagnostico = $9,
			trabajos_realizados = $10,
			observaciones = $11,
			fecha_cierre = CASE
				WHEN $6 = 'cerrada' AND fecha_cierre IS NULL THEN NOW()
				WHEN $6 <> 'cerrada' THEN NULL
				ELSE fecha_cierre
			END,
			actualizado_en = NOW()
		WHERE id = $1
		RETURNING id::text
	`

	var updatedID string
	err := r.db.QueryRow(
		ctx,
		query,
		id,
		nullableString(input.UsuarioResponsableID),
		input.FechaPrometida,
		input.KilometrajeIngreso,
		input.NivelCombustiblePorcentaje,
		input.Estado,
		nullableString(input.TipoServicio),
		input.DescripcionFalla,
		nullableString(input.Diagnostico),
		nullableString(input.TrabajosRealizados),
		nullableString(input.Observaciones),
	).Scan(&updatedID)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, updatedID)
}

func (r *Repository) GetFormato(ctx context.Context) (*FormatoOrdenTrabajo, error) {
	const query = `
		SELECT id::text, nombre, campos, activo, creado_en, actualizado_en
		FROM formatos_orden_trabajo
		WHERE nombre = 'orden_trabajo_principal' AND activo = TRUE
	`

	var formato FormatoOrdenTrabajo
	var camposBytes []byte
	err := r.db.QueryRow(ctx, query).Scan(
		&formato.ID,
		&formato.Nombre,
		&camposBytes,
		&formato.Activo,
		&formato.CreadoEn,
		&formato.ActualizadoEn,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(camposBytes, &formato.Campos); err != nil {
		return nil, err
	}

	return &formato, nil
}

func (r *Repository) SaveFormato(ctx context.Context, input GuardarFormatoOrdenEntrada) (*FormatoOrdenTrabajo, error) {
	const query = `
		INSERT INTO formatos_orden_trabajo (nombre, campos, activo)
		VALUES ('orden_trabajo_principal', $1::jsonb, TRUE)
		ON CONFLICT (nombre) DO UPDATE
		SET campos = EXCLUDED.campos,
		    activo = TRUE,
		    actualizado_en = NOW()
		RETURNING id::text
	`

	camposBytes, err := json.Marshal(input.Campos)
	if err != nil {
		return nil, err
	}

	var id string
	if err := r.db.QueryRow(ctx, query, string(camposBytes)).Scan(&id); err != nil {
		return nil, err
	}

	return r.GetFormato(ctx)
}

func (r *Repository) UpdateProgress(ctx context.Context, id string, usuarioID string, input ActualizarProgresoEntrada) (*OrdenTrabajo, error) {
	const query = `
		UPDATE ordenes_trabajo
		SET
			estado = $3,
			diagnostico = $4,
			trabajos_realizados = $5,
			observaciones = $6,
			fecha_cierre = CASE WHEN $3 = 'cerrada' THEN NOW() ELSE fecha_cierre END,
			actualizado_en = NOW()
		WHERE id = $1 AND usuario_responsable_id = $2
		RETURNING id::text
	`

	var updatedID string
	err := r.db.QueryRow(
		ctx,
		query,
		id,
		usuarioID,
		input.Estado,
		nullableString(input.Diagnostico),
		nullableString(input.TrabajosRealizados),
		nullableString(input.Observaciones),
	).Scan(&updatedID)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, updatedID)
}

const selectOrdenBase = `
	SELECT
		ot.id::text,
		ot.numero,
		ot.cliente_id::text,
		CONCAT_WS(' ', c.nombres, c.apellidos),
		ot.moto_id::text,
		COALESCE(m.placa, ''),
		COALESCE(ot.usuario_recibe_id::text, ''),
		COALESCE(CONCAT_WS(' ', urec.nombres, urec.apellidos), ''),
		COALESCE(ot.usuario_responsable_id::text, ''),
		COALESCE(CONCAT_WS(' ', ur.nombres, ur.apellidos), ''),
		ot.fecha_ingreso,
		ot.fecha_prometida,
		ot.fecha_cierre,
		ot.kilometraje_ingreso,
		ot.nivel_combustible_porcentaje,
		ot.estado,
		COALESCE(ot.tipo_servicio, ''),
		ot.descripcion_falla,
		COALESCE(ot.diagnostico, ''),
		COALESCE(ot.trabajos_realizados, ''),
		COALESCE(ot.observaciones, ''),
		COALESCE(ot.campos_formato, '{}'::JSONB),
		ot.creado_en,
		ot.actualizado_en
	FROM ordenes_trabajo ot
	INNER JOIN clientes c ON c.id = ot.cliente_id
	INNER JOIN motos m ON m.id = ot.moto_id
	LEFT JOIN usuarios urec ON urec.id = ot.usuario_recibe_id
	LEFT JOIN usuarios ur ON ur.id = ot.usuario_responsable_id
`

func (r *Repository) getByQuery(ctx context.Context, where string, args ...any) (*OrdenTrabajo, error) {
	row := r.db.QueryRow(ctx, selectOrdenBase+" "+where, args...)

	var orden OrdenTrabajo
	err := row.Scan(
		&orden.ID,
		&orden.Numero,
		&orden.ClienteID,
		&orden.NombreCliente,
		&orden.MotoID,
		&orden.PlacaMoto,
		&orden.UsuarioRecibeID,
		&orden.NombreUsuarioRecibe,
		&orden.UsuarioResponsableID,
		&orden.NombreUsuarioResponsable,
		&orden.FechaIngreso,
		&orden.FechaPrometida,
		&orden.FechaCierre,
		&orden.KilometrajeIngreso,
		&orden.NivelCombustiblePorcentaje,
		&orden.Estado,
		&orden.TipoServicio,
		&orden.DescripcionFalla,
		&orden.Diagnostico,
		&orden.TrabajosRealizados,
		&orden.Observaciones,
		&orden.CamposFormato,
		&orden.CreadoEn,
		&orden.ActualizadoEn,
	)
	if err != nil {
		return nil, err
	}

	return &orden, nil
}

type rowScanner interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanOrdenes(rows rowScanner) ([]OrdenTrabajo, error) {
	ordenes := make([]OrdenTrabajo, 0)
	for rows.Next() {
		var orden OrdenTrabajo
		if err := rows.Scan(
			&orden.ID,
			&orden.Numero,
			&orden.ClienteID,
			&orden.NombreCliente,
			&orden.MotoID,
			&orden.PlacaMoto,
			&orden.UsuarioRecibeID,
			&orden.NombreUsuarioRecibe,
			&orden.UsuarioResponsableID,
			&orden.NombreUsuarioResponsable,
			&orden.FechaIngreso,
			&orden.FechaPrometida,
			&orden.FechaCierre,
			&orden.KilometrajeIngreso,
			&orden.NivelCombustiblePorcentaje,
			&orden.Estado,
			&orden.TipoServicio,
			&orden.DescripcionFalla,
			&orden.Diagnostico,
			&orden.TrabajosRealizados,
			&orden.Observaciones,
			&orden.CamposFormato,
			&orden.CreadoEn,
			&orden.ActualizadoEn,
		); err != nil {
			return nil, err
		}
		ordenes = append(ordenes, orden)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ordenes, nil
}

type rowScannerOnce interface {
	Scan(dest ...any) error
}

func scanCatalogoItem(row rowScannerOnce) (*CatalogoItemTrabajo, error) {
	var item CatalogoItemTrabajo
	err := row.Scan(
		&item.ID,
		&item.TipoItem,
		&item.Nombre,
		&item.Descripcion,
		&item.ValorBase,
		&item.Activo,
		&item.CreadoEn,
		&item.ActualizadoEn,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func scanCatalogoItems(rows rowScanner) ([]CatalogoItemTrabajo, error) {
	items := make([]CatalogoItemTrabajo, 0)
	for rows.Next() {
		item, err := scanCatalogoItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func jsonMapString(value JSONMap) string {
	if value == nil {
		value = JSONMap{}
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
