BEGIN;

INSERT INTO roles (nombre, descripcion)
VALUES ('empleado', 'Gestiona ordenes de trabajo asignadas')
ON CONFLICT (nombre) DO UPDATE
SET descripcion = EXCLUDED.descripcion,
    actualizado_en = NOW();

COMMIT;
