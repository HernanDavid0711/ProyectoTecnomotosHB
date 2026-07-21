package motos

import "time"

type Moto struct {
	ID                string    `json:"id"`
	Placa             string    `json:"placa"`
	Marca             string    `json:"marca"`
	Modelo            string    `json:"modelo"`
	Cilindraje        int       `json:"cilindraje"`
	Anio              int       `json:"anio"`
	Color             string    `json:"color"`
	KilometrajeActual int       `json:"kilometraje_actual"`
	ClienteID         string    `json:"cliente_id"`
	NombreCliente     string    `json:"nombre_cliente"`
	CreadoEn          time.Time `json:"creado_en"`
	ActualizadoEn     time.Time `json:"actualizado_en"`
}

type CrearMotoEntrada struct {
	Placa             string `json:"placa"`
	Marca             string `json:"marca"`
	Modelo            string `json:"modelo"`
	Cilindraje        int    `json:"cilindraje"`
	Anio              int    `json:"anio"`
	Color             string `json:"color"`
	KilometrajeActual int    `json:"kilometraje_actual"`
	ClienteID         string `json:"cliente_id"`
}

type ActualizarMotoEntrada struct {
	Placa             string `json:"placa"`
	Marca             string `json:"marca"`
	Modelo            string `json:"modelo"`
	Cilindraje        int    `json:"cilindraje"`
	Anio              int    `json:"anio"`
	Color             string `json:"color"`
	KilometrajeActual int    `json:"kilometraje_actual"`
	ClienteID         string `json:"cliente_id"`
}

type MotoAutocompleteItem struct {
	ID                string `json:"id"`
	Placa             string `json:"placa"`
	Marca             string `json:"marca"`
	Modelo            string `json:"modelo"`
	Cilindraje        int    `json:"cilindraje"`
	Anio              int    `json:"anio"`
	Color             string `json:"color"`
	KilometrajeActual int    `json:"kilometraje_actual"`
	ClienteID         string `json:"cliente_id"`
	NombreCliente     string `json:"nombre_cliente"`
}
