BEGIN;

-- =========================================================
-- FUNCION GENERICA PARA actualizado_en
-- =========================================================
CREATE OR REPLACE FUNCTION actualizar_columna_actualizado_en()
RETURNS TRIGGER AS $$
BEGIN
  NEW.actualizado_en = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =========================================================
-- TABLA: roles
-- =========================================================
CREATE TABLE roles (
  id BIGSERIAL PRIMARY KEY,
  nombre VARCHAR(50) NOT NULL UNIQUE,
  descripcion TEXT,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =========================================================
-- TABLA: usuarios
-- =========================================================
CREATE TABLE usuarios (
  id BIGSERIAL PRIMARY KEY,
  rol_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
  nombres VARCHAR(100) NOT NULL,
  apellidos VARCHAR(100) NOT NULL,
  correo VARCHAR(150) NOT NULL UNIQUE,
  telefono VARCHAR(30),
  hash_contrasena TEXT NOT NULL,
  activo BOOLEAN NOT NULL DEFAULT TRUE,
  ultimo_acceso_en TIMESTAMPTZ,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_usuarios_rol_id ON usuarios(rol_id);

-- =========================================================
-- TABLA: clientes
-- =========================================================
CREATE TABLE clientes (
  id BIGSERIAL PRIMARY KEY,
  nombres VARCHAR(100) NOT NULL,
  apellidos VARCHAR(100) NOT NULL,
  telefono VARCHAR(30),
  correo VARCHAR(150),
  direccion TEXT,
  tipo_documento VARCHAR(30),
  numero_documento VARCHAR(50),
  notas TEXT,
  activo BOOLEAN NOT NULL DEFAULT TRUE,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_clientes_numero_documento
  ON clientes(numero_documento)
  WHERE numero_documento IS NOT NULL;

CREATE INDEX idx_clientes_apellidos_nombres ON clientes(apellidos, nombres);
CREATE INDEX idx_clientes_telefono ON clientes(telefono);

-- =========================================================
-- TABLA: motos
-- =========================================================
CREATE TABLE motos (
  id BIGSERIAL PRIMARY KEY,
  cliente_id BIGINT NOT NULL REFERENCES clientes(id) ON DELETE RESTRICT,
  placa VARCHAR(20),
  vin VARCHAR(50),
  numero_motor VARCHAR(50),
  marca VARCHAR(80) NOT NULL,
  modelo VARCHAR(120) NOT NULL,
  cilindraje INTEGER NOT NULL CHECK (cilindraje > 0),
  anio INTEGER CHECK (anio IS NULL OR anio BETWEEN 1950 AND 2100),
  color VARCHAR(50),
  kilometraje_actual BIGINT NOT NULL DEFAULT 0 CHECK (kilometraje_actual >= 0),
  observaciones TEXT,
  activa BOOLEAN NOT NULL DEFAULT TRUE,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_motos_cliente_id ON motos(cliente_id);
CREATE INDEX idx_motos_marca_modelo ON motos(marca, modelo);

CREATE UNIQUE INDEX uq_motos_placa
  ON motos(placa)
  WHERE placa IS NOT NULL;

CREATE UNIQUE INDEX uq_motos_vin
  ON motos(vin)
  WHERE vin IS NOT NULL;

CREATE UNIQUE INDEX uq_motos_numero_motor
  ON motos(numero_motor)
  WHERE numero_motor IS NOT NULL;

-- =========================================================
-- TABLA: historial_propietarios_moto
-- =========================================================
CREATE TABLE historial_propietarios_moto (
  id BIGSERIAL PRIMARY KEY,
  moto_id BIGINT NOT NULL REFERENCES motos(id) ON DELETE CASCADE,
  cliente_id BIGINT NOT NULL REFERENCES clientes(id) ON DELETE RESTRICT,
  fecha_inicio DATE NOT NULL,
  fecha_fin DATE,
  es_propietario_actual BOOLEAN NOT NULL DEFAULT TRUE,
  observaciones TEXT,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_historial_fechas
    CHECK (fecha_fin IS NULL OR fecha_fin >= fecha_inicio)
);

CREATE INDEX idx_historial_propietarios_moto_moto_id
  ON historial_propietarios_moto(moto_id);

CREATE INDEX idx_historial_propietarios_moto_cliente_id
  ON historial_propietarios_moto(cliente_id);

CREATE INDEX idx_historial_propietarios_moto_actual
  ON historial_propietarios_moto(moto_id, es_propietario_actual);

-- =========================================================
-- TABLA: ordenes_trabajo
-- =========================================================
CREATE TABLE ordenes_trabajo (
  id BIGSERIAL PRIMARY KEY,
  numero BIGINT GENERATED ALWAYS AS IDENTITY UNIQUE,
  cliente_id BIGINT NOT NULL REFERENCES clientes(id) ON DELETE RESTRICT,
  moto_id BIGINT NOT NULL REFERENCES motos(id) ON DELETE RESTRICT,
  usuario_recibe_id BIGINT REFERENCES usuarios(id) ON DELETE SET NULL,
  usuario_responsable_id BIGINT REFERENCES usuarios(id) ON DELETE SET NULL,

  fecha_ingreso TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  fecha_prometida TIMESTAMPTZ,
  fecha_cierre TIMESTAMPTZ,

  kilometraje_ingreso BIGINT NOT NULL DEFAULT 0 CHECK (kilometraje_ingreso >= 0),
  nivel_combustible_porcentaje SMALLINT NOT NULL DEFAULT 0
    CHECK (nivel_combustible_porcentaje BETWEEN 0 AND 100),

  estado VARCHAR(30) NOT NULL DEFAULT 'abierta'
    CHECK (estado IN (
      'abierta',
      'en_diagnostico',
      'aprobacion_pendiente',
      'en_proceso',
      'lista_para_entrega',
      'cerrada',
      'cancelada'
    )),

  tipo_servicio VARCHAR(50),
  descripcion_falla TEXT NOT NULL,
  diagnostico TEXT,
  trabajos_realizados TEXT,
  observaciones TEXT,

  subtotal NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (subtotal >= 0),
  descuento NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (descuento >= 0),
  total NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (total >= 0),

  creado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT chk_ordenes_trabajo_descuento
    CHECK (descuento <= subtotal),

  CONSTRAINT chk_ordenes_trabajo_fecha_cierre
    CHECK (fecha_cierre IS NULL OR fecha_cierre >= fecha_ingreso)
);

CREATE INDEX idx_ordenes_trabajo_cliente_id ON ordenes_trabajo(cliente_id);
CREATE INDEX idx_ordenes_trabajo_moto_id ON ordenes_trabajo(moto_id);
CREATE INDEX idx_ordenes_trabajo_estado ON ordenes_trabajo(estado);
CREATE INDEX idx_ordenes_trabajo_fecha_ingreso ON ordenes_trabajo(fecha_ingreso);
CREATE INDEX idx_ordenes_trabajo_usuario_recibe_id ON ordenes_trabajo(usuario_recibe_id);
CREATE INDEX idx_ordenes_trabajo_usuario_responsable_id ON ordenes_trabajo(usuario_responsable_id);

-- =========================================================
-- TABLA: items_orden_trabajo
-- =========================================================
CREATE TABLE items_orden_trabajo (
  id BIGSERIAL PRIMARY KEY,
  orden_trabajo_id BIGINT NOT NULL REFERENCES ordenes_trabajo(id) ON DELETE CASCADE,
  tipo_item VARCHAR(30) NOT NULL
    CHECK (tipo_item IN ('mano_obra', 'repuesto', 'servicio_externo', 'insumo')),
  descripcion TEXT NOT NULL,
  cantidad NUMERIC(10,2) NOT NULL DEFAULT 1 CHECK (cantidad > 0),
  valor_unitario NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (valor_unitario >= 0),
  descuento NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (descuento >= 0),
  total_linea NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (total_linea >= 0),
  posicion INTEGER NOT NULL DEFAULT 0,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_items_orden_trabajo_orden_trabajo_id
  ON items_orden_trabajo(orden_trabajo_id);

CREATE INDEX idx_items_orden_trabajo_tipo_item
  ON items_orden_trabajo(tipo_item);

-- =========================================================
-- TABLA: fotos_orden_trabajo
-- =========================================================
CREATE TABLE fotos_orden_trabajo (
  id BIGSERIAL PRIMARY KEY,
  orden_trabajo_id BIGINT NOT NULL REFERENCES ordenes_trabajo(id) ON DELETE CASCADE,
  url_archivo TEXT NOT NULL,
  nombre_archivo VARCHAR(255),
  descripcion TEXT,
  tipo_foto VARCHAR(20) NOT NULL DEFAULT 'general'
    CHECK (tipo_foto IN ('general', 'antes', 'durante', 'despues', 'evidencia')),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fotos_orden_trabajo_orden_trabajo_id
  ON fotos_orden_trabajo(orden_trabajo_id);

CREATE INDEX idx_fotos_orden_trabajo_tipo_foto
  ON fotos_orden_trabajo(tipo_foto);

-- =========================================================
-- TABLA: comentarios_orden_trabajo
-- =========================================================
CREATE TABLE comentarios_orden_trabajo (
  id BIGSERIAL PRIMARY KEY,
  orden_trabajo_id BIGINT NOT NULL REFERENCES ordenes_trabajo(id) ON DELETE CASCADE,
  usuario_id BIGINT REFERENCES usuarios(id) ON DELETE SET NULL,
  comentario TEXT NOT NULL,
  es_interno BOOLEAN NOT NULL DEFAULT TRUE,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comentarios_orden_trabajo_orden_trabajo_id
  ON comentarios_orden_trabajo(orden_trabajo_id);

CREATE INDEX idx_comentarios_orden_trabajo_usuario_id
  ON comentarios_orden_trabajo(usuario_id);

-- =========================================================
-- TABLA: ingresos
-- =========================================================
CREATE TABLE ingresos (
  id BIGSERIAL PRIMARY KEY,
  cliente_id BIGINT REFERENCES clientes(id) ON DELETE SET NULL,
  orden_trabajo_id BIGINT REFERENCES ordenes_trabajo(id) ON DELETE SET NULL,
  usuario_id BIGINT REFERENCES usuarios(id) ON DELETE SET NULL,
  fecha_ingreso TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  concepto VARCHAR(150) NOT NULL,
  metodo_pago VARCHAR(30) NOT NULL
    CHECK (metodo_pago IN ('efectivo', 'transferencia', 'tarjeta', 'qr', 'otro')),
  referencia VARCHAR(100),
  valor NUMERIC(12,2) NOT NULL CHECK (valor > 0),
  observaciones TEXT,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ingresos_cliente_id ON ingresos(cliente_id);
CREATE INDEX idx_ingresos_orden_trabajo_id ON ingresos(orden_trabajo_id);
CREATE INDEX idx_ingresos_usuario_id ON ingresos(usuario_id);
CREATE INDEX idx_ingresos_fecha_ingreso ON ingresos(fecha_ingreso);
CREATE INDEX idx_ingresos_metodo_pago ON ingresos(metodo_pago);

-- =========================================================
-- TABLA: recordatorios_servicio
-- =========================================================
CREATE TABLE recordatorios_servicio (
  id BIGSERIAL PRIMARY KEY,
  cliente_id BIGINT NOT NULL REFERENCES clientes(id) ON DELETE CASCADE,
  moto_id BIGINT NOT NULL REFERENCES motos(id) ON DELETE CASCADE,
  orden_trabajo_id BIGINT REFERENCES ordenes_trabajo(id) ON DELETE SET NULL,
  fecha_recordatorio DATE NOT NULL,
  kilometraje_recordatorio BIGINT CHECK (kilometraje_recordatorio IS NULL OR kilometraje_recordatorio >= 0),
  tipo_recordatorio VARCHAR(50) NOT NULL,
  canal VARCHAR(20) NOT NULL DEFAULT 'whatsapp'
    CHECK (canal IN ('whatsapp', 'llamada', 'correo', 'sms', 'otro')),
  mensaje TEXT,
  estado VARCHAR(20) NOT NULL DEFAULT 'pendiente'
    CHECK (estado IN ('pendiente', 'enviado', 'cancelado')),
  enviado_en TIMESTAMPTZ,
  observaciones TEXT,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recordatorios_servicio_cliente_id
  ON recordatorios_servicio(cliente_id);

CREATE INDEX idx_recordatorios_servicio_moto_id
  ON recordatorios_servicio(moto_id);

CREATE INDEX idx_recordatorios_servicio_fecha_recordatorio
  ON recordatorios_servicio(fecha_recordatorio);

CREATE INDEX idx_recordatorios_servicio_estado
  ON recordatorios_servicio(estado);

-- =========================================================
-- TRIGGERS updated_at -> actualizado_en
-- =========================================================
CREATE TRIGGER trg_roles_actualizado_en
BEFORE UPDATE ON roles
FOR EACH ROW
EXECUTE FUNCTION actualizar_columna_actualizado_en();

CREATE TRIGGER trg_usuarios_actualizado_en
BEFORE UPDATE ON usuarios
FOR EACH ROW
EXECUTE FUNCTION actualizar_columna_actualizado_en();

CREATE TRIGGER trg_clientes_actualizado_en
BEFORE UPDATE ON clientes
FOR EACH ROW
EXECUTE FUNCTION actualizar_columna_actualizado_en();

CREATE TRIGGER trg_motos_actualizado_en
BEFORE UPDATE ON motos
FOR EACH ROW
EXECUTE FUNCTION actualizar_columna_actualizado_en();

CREATE TRIGGER trg_ordenes_trabajo_actualizado_en
BEFORE UPDATE ON ordenes_trabajo
FOR EACH ROW
EXECUTE FUNCTION actualizar_columna_actualizado_en();

CREATE TRIGGER trg_items_orden_trabajo_actualizado_en
BEFORE UPDATE ON items_orden_trabajo
FOR EACH ROW
EXECUTE FUNCTION actualizar_columna_actualizado_en();

CREATE TRIGGER trg_ingresos_actualizado_en
BEFORE UPDATE ON ingresos
FOR EACH ROW
EXECUTE FUNCTION actualizar_columna_actualizado_en();

CREATE TRIGGER trg_recordatorios_servicio_actualizado_en
BEFORE UPDATE ON recordatorios_servicio
FOR EACH ROW
EXECUTE FUNCTION actualizar_columna_actualizado_en();

-- =========================================================
-- DATOS INICIALES
-- =========================================================
INSERT INTO roles (nombre, descripcion) VALUES
  ('administrador', 'Acceso total al sistema'),
  ('mecanico', 'Gestiona diagnosticos y trabajos'),
  ('recepcion', 'Recibe motos y registra informacion');

COMMIT;