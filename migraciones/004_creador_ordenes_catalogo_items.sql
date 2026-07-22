BEGIN;

INSERT INTO roles (nombre, descripcion)
VALUES ('creador_ordenes', 'Crea ordenes de trabajo y las asigna a empleados')
ON CONFLICT (nombre) DO UPDATE
SET descripcion = EXCLUDED.descripcion,
    actualizado_en = NOW();

CREATE TABLE IF NOT EXISTS catalogo_items_trabajo (
  id BIGSERIAL PRIMARY KEY,
  tipo_item VARCHAR(30) NOT NULL
    CHECK (tipo_item IN ('mano_obra', 'repuesto', 'servicio_externo', 'insumo')),
  nombre VARCHAR(150) NOT NULL,
  descripcion TEXT,
  valor_base NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (valor_base >= 0),
  activo BOOLEAN NOT NULL DEFAULT TRUE,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (tipo_item, nombre)
);

CREATE INDEX IF NOT EXISTS idx_catalogo_items_trabajo_tipo
  ON catalogo_items_trabajo(tipo_item);

CREATE INDEX IF NOT EXISTS idx_catalogo_items_trabajo_activo
  ON catalogo_items_trabajo(activo);

ALTER TABLE items_orden_trabajo
ADD COLUMN IF NOT EXISTS catalogo_item_trabajo_id BIGINT REFERENCES catalogo_items_trabajo(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_items_orden_catalogo_item
  ON items_orden_trabajo(catalogo_item_trabajo_id);

DROP TRIGGER IF EXISTS trg_catalogo_items_trabajo_actualizado_en ON catalogo_items_trabajo;

CREATE TRIGGER trg_catalogo_items_trabajo_actualizado_en
BEFORE UPDATE ON catalogo_items_trabajo
FOR EACH ROW
EXECUTE FUNCTION actualizar_columna_actualizado_en();

INSERT INTO catalogo_items_trabajo (tipo_item, nombre, descripcion, valor_base) VALUES
  ('mano_obra', 'Revision general', 'Inspeccion inicial de la motocicleta', 0),
  ('mano_obra', 'Cambio de aceite', 'Cambio de aceite de motor', 0),
  ('mano_obra', 'Sincronizacion', 'Ajuste y puesta a punto del motor', 0),
  ('mano_obra', 'Revision sistema electrico', 'Diagnostico de luces, bateria y conexiones', 0),
  ('repuesto', 'Aceite motor', 'Aceite para motor de motocicleta', 0),
  ('repuesto', 'Filtro de aceite', 'Filtro de aceite segun referencia', 0),
  ('insumo', 'Limpiador de carburador', 'Insumo para limpieza de carburador o cuerpo de aceleracion', 0)
ON CONFLICT (tipo_item, nombre) DO NOTHING;

COMMIT;
