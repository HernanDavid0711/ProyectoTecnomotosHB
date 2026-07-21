package clientes

import "time"

type Cliente struct {
	ID              string    `json:"id"`
	Nombres         string    `json:"nombres"`
	Apellidos       string    `json:"apellidos"`
	Telefono        string    `json:"telefono"`
	Correo          string    `json:"correo"`
	Direccion       string    `json:"direccion"`
	TipoDocumento   string    `json:"tipo_documento"`
	NumeroDocumento string    `json:"numero_documento"`
	Notas           string    `json:"notas"`
	Activo          bool      `json:"activo"`
	CreadoEn        time.Time `json:"creado_en"`
	ActualizadoEn   time.Time `json:"actualizado_en"`
}

type CrearClienteEntrada struct {
	Nombres         string `json:"nombres"`
	Apellidos       string `json:"apellidos"`
	Telefono        string `json:"telefono"`
	Correo          string `json:"correo"`
	Direccion       string `json:"direccion"`
	TipoDocumento   string `json:"tipo_documento"`
	NumeroDocumento string `json:"numero_documento"`
	Notas           string `json:"notas"`
}

type ActualizarClienteEntrada struct {
	Nombres         string `json:"nombres"`
	Apellidos       string `json:"apellidos"`
	Telefono        string `json:"telefono"`
	Correo          string `json:"correo"`
	Direccion       string `json:"direccion"`
	TipoDocumento   string `json:"tipo_documento"`
	NumeroDocumento string `json:"numero_documento"`
	Notas           string `json:"notas"`
}
