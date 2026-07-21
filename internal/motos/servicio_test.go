package motos

import "testing"

func TestValidarCrearEntrada(t *testing.T) {
	valid := CrearMotoEntrada{
		Placa:             "ABC12D",
		Marca:             "Yamaha",
		Modelo:            "FZ",
		Cilindraje:        150,
		Anio:              2022,
		KilometrajeActual: 1200,
		ClienteID:         "1",
	}
	if err := validarCrearEntrada(valid); err != nil {
		t.Fatalf("expected valid moto, got %v", err)
	}

	cases := []CrearMotoEntrada{
		{Placa: "", Marca: "Yamaha", Modelo: "FZ", Cilindraje: 150, ClienteID: "1"},
		{Placa: "ABC12D", Marca: "", Modelo: "FZ", Cilindraje: 150, ClienteID: "1"},
		{Placa: "ABC12D", Marca: "Yamaha", Modelo: "", Cilindraje: 150, ClienteID: "1"},
		{Placa: "ABC-12D", Marca: "Yamaha", Modelo: "FZ", Cilindraje: 150, ClienteID: "1"},
		{Placa: "ABC12D", Marca: "Yamaha", Modelo: "FZ", Cilindraje: 0, ClienteID: "1"},
		{Placa: "ABC12D", Marca: "Yamaha", Modelo: "FZ", Cilindraje: 150, KilometrajeActual: -1, ClienteID: "1"},
		{Placa: "ABC12D", Marca: "Yamaha", Modelo: "FZ", Cilindraje: 150, Anio: 1949, ClienteID: "1"},
		{Placa: "ABC12D", Marca: "Yamaha", Modelo: "FZ", Cilindraje: 150, ClienteID: "abc"},
	}

	for _, tc := range cases {
		if err := validarCrearEntrada(tc); err == nil {
			t.Fatalf("expected invalid moto for %+v", tc)
		}
	}
}

func TestNormalizarCrearEntrada(t *testing.T) {
	input := normalizarCrearEntrada(CrearMotoEntrada{
		Placa:     " abc-12d ",
		Marca:     " Yamaha ",
		Modelo:    " FZ ",
		Color:     " Negro ",
		ClienteID: " 001 ",
	})

	if input.Placa != "ABC12D" || input.ClienteID != "1" {
		t.Fatalf("unexpected normalized moto: %+v", input)
	}
	if input.Marca != "Yamaha" || input.Modelo != "FZ" || input.Color != "Negro" {
		t.Fatalf("unexpected normalized text: %+v", input)
	}
}
