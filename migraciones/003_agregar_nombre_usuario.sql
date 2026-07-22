BEGIN;

ALTER TABLE usuarios
ADD COLUMN IF NOT EXISTS nombre_usuario VARCHAR(80);

UPDATE usuarios
SET nombre_usuario = LOWER(SPLIT_PART(correo, '@', 1))
WHERE nombre_usuario IS NULL OR nombre_usuario = '';

ALTER TABLE usuarios
ALTER COLUMN nombre_usuario SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_usuarios_nombre_usuario
ON usuarios(nombre_usuario);

COMMIT;
