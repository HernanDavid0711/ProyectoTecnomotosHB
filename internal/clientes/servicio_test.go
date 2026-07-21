package clientes

import "testing"

func TestValidarCrearEntrada(t *testing.T) {
	valid := CrearClienteEntrada{
		Nombres:         "Ana",
		Apellidos:       "Perez",
		Telefono:        "3001234567",
		Correo:          "ana@example.com",
		TipoDocumento:   "cc",
		NumeroDocumento: "123456",
	}
	if err := validarCrearEntrada(valid); err != nil {
		t.Fatalf("expected valid cliente, got %v", err)
	}

	cases := []CrearClienteEntrada{
		{Nombres: "", Apellidos: "Perez"},
		{Nombres: "Ana", Apellidos: ""},
		{Nombres: "Ana", Apellidos: "Perez", Correo: "correo-malo"},
		{Nombres: "Ana", Apellidos: "Perez", Telefono: "123"},
		{Nombres: "Ana", Apellidos: "Perez", TipoDocumento: "cc"},
		{Nombres: "Ana", Apellidos: "Perez", TipoDocumento: "otro", NumeroDocumento: "123456"},
	}

	for _, tc := range cases {
		if err := validarCrearEntrada(tc); err == nil {
			t.Fatalf("expected invalid cliente for %+v", tc)
		}
	}
}

func TestNormalizarCrearEntrada(t *testing.T) {
	input := normalizarCrearEntrada(CrearClienteEntrada{
		Nombres:       "  Ana   Maria ",
		Apellidos:     " Perez  Ruiz ",
		Correo:        " ANA@EXAMPLE.COM ",
		TipoDocumento: " CC ",
	})

	if input.Nombres != "Ana Maria" || input.Apellidos != "Perez Ruiz" {
		t.Fatalf("unexpected normalized name: %+v", input)
	}
	if input.Correo != "ana@example.com" || input.TipoDocumento != "cc" {
		t.Fatalf("unexpected normalized contact data: %+v", input)
	}
}
