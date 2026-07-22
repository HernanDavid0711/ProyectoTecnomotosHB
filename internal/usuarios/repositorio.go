package usuarios

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

func (r *Repository) Count(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM usuarios`

	var total int
	err := r.db.QueryRow(ctx, query).Scan(&total)
	return total, err
}

func (r *Repository) ListRoles(ctx context.Context) ([]Rol, error) {
	const query = `
		SELECT
			id::text,
			nombre,
			COALESCE(descripcion, ''),
			creado_en,
			actualizado_en
		FROM roles
		ORDER BY nombre ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]Rol, 0)
	for rows.Next() {
		var rol Rol
		if err := rows.Scan(
			&rol.ID,
			&rol.Nombre,
			&rol.Descripcion,
			&rol.CreadoEn,
			&rol.ActualizadoEn,
		); err != nil {
			return nil, err
		}
		roles = append(roles, rol)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *Repository) EnsureDefaultRoles(ctx context.Context) error {
	const query = `
		INSERT INTO roles (nombre, descripcion) VALUES
			('administrador', 'Acceso total al sistema'),
			('empleado', 'Gestiona ordenes de trabajo asignadas'),
			('creador_ordenes', 'Crea ordenes de trabajo y las asigna a empleados')
		ON CONFLICT (nombre) DO UPDATE
		SET descripcion = EXCLUDED.descripcion,
		    actualizado_en = NOW()
	`

	_, err := r.db.Exec(ctx, query)
	return err
}

func (r *Repository) Create(ctx context.Context, input CrearUsuarioEntrada, hashContrasena string) (*Usuario, error) {
	const query = `
		INSERT INTO usuarios (
			rol_id,
			nombre_usuario,
			nombres,
			apellidos,
			correo,
			telefono,
			hash_contrasena,
			activo
		)
		SELECT id, $2, $3, $4, $5, $6, $7, TRUE
		FROM roles
		WHERE nombre = $1
		RETURNING usuarios.id::text
	`

	var id string
	err := r.db.QueryRow(
		ctx,
		query,
		input.Rol,
		input.Usuario,
		input.Nombres,
		input.Apellidos,
		input.Correo,
		nullableString(input.Telefono),
		hashContrasena,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Usuario, error) {
	const query = `
		SELECT
			u.id::text,
			u.rol_id::text,
			r.nombre,
			u.nombre_usuario,
			u.nombres,
			u.apellidos,
			u.correo,
			COALESCE(u.telefono, ''),
			u.activo,
			u.ultimo_acceso_en,
			u.creado_en,
			u.actualizado_en
		FROM usuarios u
		INNER JOIN roles r ON r.id = u.rol_id
		WHERE u.id = $1
	`

	var usuario Usuario
	err := r.db.QueryRow(ctx, query, id).Scan(
		&usuario.ID,
		&usuario.RolID,
		&usuario.Rol,
		&usuario.NombreUsuario,
		&usuario.Nombres,
		&usuario.Apellidos,
		&usuario.Correo,
		&usuario.Telefono,
		&usuario.Activo,
		&usuario.UltimoAccesoEn,
		&usuario.CreadoEn,
		&usuario.ActualizadoEn,
	)
	if err != nil {
		return nil, err
	}

	return &usuario, nil
}

func (r *Repository) GetByIdentificador(ctx context.Context, identificador string) (*UsuarioConPassword, error) {
	const query = `
		SELECT
			u.id::text,
			u.rol_id::text,
			r.nombre,
			u.nombre_usuario,
			u.nombres,
			u.apellidos,
			u.correo,
			COALESCE(u.telefono, ''),
			u.activo,
			u.ultimo_acceso_en,
			u.creado_en,
			u.actualizado_en,
			u.hash_contrasena
		FROM usuarios u
		INNER JOIN roles r ON r.id = u.rol_id
		WHERE u.correo = $1 OR u.nombre_usuario = $1
	`

	var usuario UsuarioConPassword
	err := r.db.QueryRow(ctx, query, identificador).Scan(
		&usuario.ID,
		&usuario.RolID,
		&usuario.Rol,
		&usuario.NombreUsuario,
		&usuario.Nombres,
		&usuario.Apellidos,
		&usuario.Correo,
		&usuario.Telefono,
		&usuario.Activo,
		&usuario.UltimoAccesoEn,
		&usuario.CreadoEn,
		&usuario.ActualizadoEn,
		&usuario.HashContrasena,
	)
	if err != nil {
		return nil, err
	}

	return &usuario, nil
}

func (r *Repository) List(ctx context.Context, busqueda string, limit int, offset int) ([]Usuario, error) {
	const query = `
		SELECT
			u.id::text,
			u.rol_id::text,
			r.nombre,
			u.nombre_usuario,
			u.nombres,
			u.apellidos,
			u.correo,
			COALESCE(u.telefono, ''),
			u.activo,
			u.ultimo_acceso_en,
			u.creado_en,
			u.actualizado_en
		FROM usuarios u
		INNER JOIN roles r ON r.id = u.rol_id
		WHERE
			$1 = ''
			OR u.nombres ILIKE '%' || $1 || '%'
			OR u.nombre_usuario ILIKE '%' || $1 || '%'
			OR u.apellidos ILIKE '%' || $1 || '%'
			OR u.correo ILIKE '%' || $1 || '%'
			OR r.nombre ILIKE '%' || $1 || '%'
		ORDER BY u.creado_en DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, busqueda, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usuarios := make([]Usuario, 0)
	for rows.Next() {
		var usuario Usuario
		if err := rows.Scan(
			&usuario.ID,
			&usuario.RolID,
			&usuario.Rol,
			&usuario.NombreUsuario,
			&usuario.Nombres,
			&usuario.Apellidos,
			&usuario.Correo,
			&usuario.Telefono,
			&usuario.Activo,
			&usuario.UltimoAccesoEn,
			&usuario.CreadoEn,
			&usuario.ActualizadoEn,
		); err != nil {
			return nil, err
		}
		usuarios = append(usuarios, usuario)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return usuarios, nil
}

func (r *Repository) Update(ctx context.Context, id string, input ActualizarUsuarioEntrada, hashContrasena *string) (*Usuario, error) {
	const query = `
		UPDATE usuarios
		SET
			rol_id = roles.id,
			nombre_usuario = $3,
			nombres = $4,
			apellidos = $5,
			correo = $6,
			telefono = $7,
			hash_contrasena = COALESCE($8, hash_contrasena),
			activo = COALESCE($9, activo),
			actualizado_en = NOW()
		FROM roles
		WHERE usuarios.id = $1 AND roles.nombre = $2
		RETURNING usuarios.id::text
	`

	var updatedID string
	err := r.db.QueryRow(
		ctx,
		query,
		id,
		input.Rol,
		input.Usuario,
		input.Nombres,
		input.Apellidos,
		input.Correo,
		nullableString(input.Telefono),
		hashContrasena,
		input.Activo,
	).Scan(&updatedID)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, updatedID)
}

func (r *Repository) AssignRole(ctx context.Context, id string, rol string) (*Usuario, error) {
	const query = `
		UPDATE usuarios
		SET
			rol_id = roles.id,
			actualizado_en = NOW()
		FROM roles
		WHERE usuarios.id = $1 AND roles.nombre = $2
		RETURNING usuarios.id::text
	`

	var updatedID string
	err := r.db.QueryRow(ctx, query, id, rol).Scan(&updatedID)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, updatedID)
}

func (r *Repository) Deactivate(ctx context.Context, id string) (bool, error) {
	const query = `
		UPDATE usuarios
		SET activo = FALSE, actualizado_en = NOW()
		WHERE id = $1 AND activo = TRUE
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return false, err
	}

	return result.RowsAffected() > 0, nil
}

func (r *Repository) TouchLastLogin(ctx context.Context, id string) error {
	const query = `
		UPDATE usuarios
		SET ultimo_acceso_en = NOW(), actualizado_en = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	return err
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
