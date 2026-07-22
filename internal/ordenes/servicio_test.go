package ordenes

import "testing"

func TestValidarCrearEntrada(t *testing.T) {
	valid := CrearOrdenEntrada{
		ClienteID:                  "1",
		MotoID:                     "2",
		UsuarioResponsableID:       "3",
		KilometrajeIngreso:         1200,
		NivelCombustiblePorcentaje: 50,
		DescripcionFalla:           "No enciende",
	}
	if err := validarCrearEntrada(valid); err != nil {
		t.Fatalf("expected valid orden, got %v", err)
	}

	cases := []CrearOrdenEntrada{
		{ClienteID: "", MotoID: "2", DescripcionFalla: "Falla"},
		{ClienteID: "1", MotoID: "abc", DescripcionFalla: "Falla"},
		{ClienteID: "1", MotoID: "2", UsuarioResponsableID: "abc", DescripcionFalla: "Falla"},
		{ClienteID: "1", MotoID: "2", KilometrajeIngreso: -1, DescripcionFalla: "Falla"},
		{ClienteID: "1", MotoID: "2", NivelCombustiblePorcentaje: 101, DescripcionFalla: "Falla"},
		{ClienteID: "1", MotoID: "2", DescripcionFalla: ""},
	}

	for _, tc := range cases {
		if err := validarCrearEntrada(tc); err == nil {
			t.Fatalf("expected invalid orden for %+v", tc)
		}
	}
}

func TestNormalizarEstado(t *testing.T) {
	if estado := normalizarEstado(" finalizada "); estado != EstadoCerrada {
		t.Fatalf("expected finalizada to map to cerrada, got %q", estado)
	}
}

func TestValidarProgresoEntrada(t *testing.T) {
	for _, estado := range []string{EstadoEnDiagnostico, EstadoEnProceso, EstadoListaParaEntrega, EstadoCerrada} {
		if err := validarProgresoEntrada(ActualizarProgresoEntrada{Estado: estado}); err != nil {
			t.Fatalf("expected valid estado %q, got %v", estado, err)
		}
	}

	for _, estado := range []string{"", EstadoAbierta, EstadoCancelada, "otro"} {
		if err := validarProgresoEntrada(ActualizarProgresoEntrada{Estado: estado}); err == nil {
			t.Fatalf("expected invalid estado %q", estado)
		}
	}
}
