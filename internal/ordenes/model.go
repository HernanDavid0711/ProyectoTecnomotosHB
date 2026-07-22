package ordenes

import (
	"encoding/json"
	"time"
)

const (
	EstadoAbierta             = "abierta"
	EstadoEnDiagnostico       = "en_diagnostico"
	EstadoAprobacionPendiente = "aprobacion_pendiente"
	EstadoEnProceso           = "en_proceso"
	EstadoListaParaEntrega    = "lista_para_entrega"
	EstadoCerrada             = "cerrada"
	EstadoCancelada           = "cancelada"
)

type OrdenTrabajo struct {
	ID                         string     `json:"id"`
	Numero                     int64      `json:"numero"`
	ClienteID                  string     `json:"cliente_id"`
	NombreCliente              string     `json:"nombre_cliente"`
	MotoID                     string     `json:"moto_id"`
	PlacaMoto                  string     `json:"placa_moto"`
	UsuarioRecibeID            string     `json:"usuario_recibe_id"`
	NombreUsuarioRecibe        string     `json:"nombre_usuario_recibe"`
	UsuarioResponsableID       string     `json:"usuario_responsable_id"`
	NombreUsuarioResponsable   string     `json:"nombre_usuario_responsable"`
	FechaIngreso               time.Time  `json:"fecha_ingreso"`
	FechaPrometida             *time.Time `json:"fecha_prometida,omitempty"`
	FechaCierre                *time.Time `json:"fecha_cierre,omitempty"`
	KilometrajeIngreso         int64      `json:"kilometraje_ingreso"`
	NivelCombustiblePorcentaje int        `json:"nivel_combustible_porcentaje"`
	Estado                     string     `json:"estado"`
	TipoServicio               string     `json:"tipo_servicio"`
	DescripcionFalla           string     `json:"descripcion_falla"`
	Diagnostico                string     `json:"diagnostico"`
	TrabajosRealizados         string     `json:"trabajos_realizados"`
	Observaciones              string     `json:"observaciones"`
	CamposFormato              JSONMap    `json:"campos_formato"`
	CreadoEn                   time.Time  `json:"creado_en"`
	ActualizadoEn              time.Time  `json:"actualizado_en"`
}

type JSONMap map[string]any

func (m *JSONMap) Scan(value any) error {
	if value == nil {
		*m = JSONMap{}
		return nil
	}
	var bytes []byte
	switch typed := value.(type) {
	case []byte:
		bytes = typed
	case string:
		bytes = []byte(typed)
	default:
		return nil
	}
	return json.Unmarshal(bytes, m)
}

type FormatoOrdenTrabajo struct {
	ID            string         `json:"id"`
	Nombre        string         `json:"nombre"`
	Campos        []CampoFormato `json:"campos"`
	Activo        bool           `json:"activo"`
	CreadoEn      time.Time      `json:"creado_en"`
	ActualizadoEn time.Time      `json:"actualizado_en"`
}

type CampoFormato struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Tipo      string `json:"tipo"`
	Requerido bool   `json:"requerido"`
	Sistema   bool   `json:"sistema"`
}

type MotoResumen struct {
	ID                string `json:"id"`
	Placa             string `json:"placa"`
	Marca             string `json:"marca"`
	Modelo            string `json:"modelo"`
	Cilindraje        int    `json:"cilindraje"`
	Anio              int    `json:"anio"`
	Color             string `json:"color"`
	KilometrajeActual int    `json:"kilometraje_actual"`
}

type ClienteConMotos struct {
	ID              string        `json:"id"`
	Nombres         string        `json:"nombres"`
	Apellidos       string        `json:"apellidos"`
	NombreCompleto  string        `json:"nombre_completo"`
	Telefono        string        `json:"telefono"`
	Correo          string        `json:"correo"`
	TipoDocumento   string        `json:"tipo_documento"`
	NumeroDocumento string        `json:"numero_documento"`
	Motos           []MotoResumen `json:"motos"`
}

type EmpleadoAsignable struct {
	ID             string `json:"id"`
	NombreUsuario  string `json:"usuario"`
	Nombres        string `json:"nombres"`
	Apellidos      string `json:"apellidos"`
	NombreCompleto string `json:"nombre_completo"`
	Correo         string `json:"correo"`
}

type CatalogoItemTrabajo struct {
	ID            string    `json:"id"`
	TipoItem      string    `json:"tipo_item"`
	Nombre        string    `json:"nombre"`
	Descripcion   string    `json:"descripcion"`
	ValorBase     float64   `json:"valor_base"`
	Activo        bool      `json:"activo"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}

type GuardarCatalogoItemTrabajoEntrada struct {
	TipoItem    string `json:"tipo_item"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Activo      *bool  `json:"activo,omitempty"`
}

type ItemOrdenTrabajo struct {
	ID                    string  `json:"id"`
	CatalogoItemTrabajoID string  `json:"catalogo_item_trabajo_id"`
	TipoItem              string  `json:"tipo_item"`
	Descripcion           string  `json:"descripcion"`
	Cantidad              float64 `json:"cantidad"`
	ValorUnitario         float64 `json:"valor_unitario"`
	Descuento             float64 `json:"descuento"`
	TotalLinea            float64 `json:"total_linea"`
	Posicion              int     `json:"posicion"`
}

type CrearOrdenEntrada struct {
	ClienteID                  string     `json:"cliente_id"`
	MotoID                     string     `json:"moto_id"`
	UsuarioResponsableID       string     `json:"usuario_responsable_id"`
	FechaPrometida             *time.Time `json:"fecha_prometida,omitempty"`
	KilometrajeIngreso         int64      `json:"kilometraje_ingreso"`
	NivelCombustiblePorcentaje int        `json:"nivel_combustible_porcentaje"`
	TipoServicio               string     `json:"tipo_servicio"`
	DescripcionFalla           string     `json:"descripcion_falla"`
	Observaciones              string     `json:"observaciones"`
	CamposFormato              JSONMap    `json:"campos_formato"`
}

type CrearOrdenConItemsEntrada struct {
	ClienteID                  string                  `json:"cliente_id"`
	MotoID                     string                  `json:"moto_id"`
	UsuarioResponsableID       string                  `json:"usuario_responsable_id"`
	FechaPrometida             *time.Time              `json:"fecha_prometida,omitempty"`
	KilometrajeIngreso         int64                   `json:"kilometraje_ingreso"`
	NivelCombustiblePorcentaje int                     `json:"nivel_combustible_porcentaje"`
	TipoServicio               string                  `json:"tipo_servicio"`
	DescripcionFalla           string                  `json:"descripcion_falla"`
	Observaciones              string                  `json:"observaciones"`
	CamposFormato              JSONMap                 `json:"campos_formato"`
	Items                      []CrearItemOrdenEntrada `json:"items"`
}

type GuardarFormatoOrdenEntrada struct {
	Campos []CampoFormato `json:"campos"`
}

type CrearItemOrdenEntrada struct {
	CatalogoItemTrabajoID string  `json:"catalogo_item_trabajo_id"`
	Cantidad              float64 `json:"cantidad"`
	ValorUnitario         float64 `json:"valor_unitario"`
	Descuento             float64 `json:"descuento"`
}

type AsignarOrdenEntrada struct {
	UsuarioResponsableID string `json:"usuario_responsable_id"`
}

type ActualizarOrdenAdminEntrada struct {
	UsuarioResponsableID       string     `json:"usuario_responsable_id"`
	FechaPrometida             *time.Time `json:"fecha_prometida,omitempty"`
	KilometrajeIngreso         int64      `json:"kilometraje_ingreso"`
	NivelCombustiblePorcentaje int        `json:"nivel_combustible_porcentaje"`
	Estado                     string     `json:"estado"`
	TipoServicio               string     `json:"tipo_servicio"`
	DescripcionFalla           string     `json:"descripcion_falla"`
	Diagnostico                string     `json:"diagnostico"`
	TrabajosRealizados         string     `json:"trabajos_realizados"`
	Observaciones              string     `json:"observaciones"`
}

type ActualizarProgresoEntrada struct {
	Estado             string `json:"estado"`
	Diagnostico        string `json:"diagnostico"`
	TrabajosRealizados string `json:"trabajos_realizados"`
	Observaciones      string `json:"observaciones"`
}
