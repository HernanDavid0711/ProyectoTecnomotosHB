package usuarios

import "time"

const (
	RolAdministrador  = "administrador"
	RolCreadorOrdenes = "creador_ordenes"
	RolEmpleado       = "empleado"
)

type Rol struct {
	ID            string    `json:"id"`
	Nombre        string    `json:"nombre"`
	Descripcion   string    `json:"descripcion"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}

type Usuario struct {
	ID             string     `json:"id"`
	RolID          string     `json:"rol_id"`
	Rol            string     `json:"rol"`
	NombreUsuario  string     `json:"usuario"`
	Nombres        string     `json:"nombres"`
	Apellidos      string     `json:"apellidos"`
	Correo         string     `json:"correo"`
	Telefono       string     `json:"telefono"`
	Activo         bool       `json:"activo"`
	UltimoAccesoEn *time.Time `json:"ultimo_acceso_en,omitempty"`
	CreadoEn       time.Time  `json:"creado_en"`
	ActualizadoEn  time.Time  `json:"actualizado_en"`
}

type UsuarioConPassword struct {
	Usuario
	HashContrasena string
}

type CrearUsuarioEntrada struct {
	Rol        string `json:"rol"`
	Usuario    string `json:"usuario"`
	Nombres    string `json:"nombres"`
	Apellidos  string `json:"apellidos"`
	Correo     string `json:"correo"`
	Telefono   string `json:"telefono"`
	Contrasena string `json:"contrasena"`
}

type ActualizarUsuarioEntrada struct {
	Rol        string `json:"rol"`
	Usuario    string `json:"usuario"`
	Nombres    string `json:"nombres"`
	Apellidos  string `json:"apellidos"`
	Correo     string `json:"correo"`
	Telefono   string `json:"telefono"`
	Contrasena string `json:"contrasena,omitempty"`
	Activo     *bool  `json:"activo,omitempty"`
}

type AsignarRolEntrada struct {
	Rol string `json:"rol"`
}
