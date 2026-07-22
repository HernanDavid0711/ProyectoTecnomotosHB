BEGIN;

CREATE TABLE IF NOT EXISTS formatos_orden_trabajo (
  id BIGSERIAL PRIMARY KEY,
  nombre VARCHAR(120) NOT NULL UNIQUE,
  campos JSONB NOT NULL DEFAULT '[]'::JSONB,
  activo BOOLEAN NOT NULL DEFAULT TRUE,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE ordenes_trabajo
ADD COLUMN IF NOT EXISTS campos_formato JSONB NOT NULL DEFAULT '{}'::JSONB;

DROP TRIGGER IF EXISTS trg_formatos_orden_trabajo_actualizado_en ON formatos_orden_trabajo;

CREATE TRIGGER trg_formatos_orden_trabajo_actualizado_en
BEFORE UPDATE ON formatos_orden_trabajo
FOR EACH ROW
EXECUTE FUNCTION actualizar_columna_actualizado_en();

INSERT INTO formatos_orden_trabajo (nombre, campos, activo)
VALUES (
  'orden_trabajo_principal',
  '[
    {"id":"cliente","label":"Cliente","tipo":"texto","requerido":true,"sistema":true},
    {"id":"numero_telefono","label":"Numero de telefono","tipo":"texto","requerido":true,"sistema":true},
    {"id":"moto","label":"Moto","tipo":"texto","requerido":true,"sistema":true},
    {"id":"placa","label":"Placa","tipo":"texto","requerido":true,"sistema":true},
    {"id":"modelo","label":"Modelo","tipo":"texto","requerido":true,"sistema":true},
    {"id":"kilometraje_ingreso","label":"Kilometraje","tipo":"numero","requerido":true,"sistema":true},
    {"id":"observaciones_motocicleta","label":"Observaciones de la motocicleta","tipo":"area","requerido":true,"sistema":true}
  ]'::JSONB,
  TRUE
)
ON CONFLICT (nombre) DO UPDATE
SET campos = EXCLUDED.campos,
    activo = TRUE,
    actualizado_en = NOW();

COMMIT;
