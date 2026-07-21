package clientes

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"tecnomotos/internal/shared"
)

var ErrClienteNoEncontrado = errors.New("cliente no encontrado")
var ErrClienteInvalido = errors.New("datos del cliente invalidos")
var ErrDocumentoClienteYaExiste = errors.New("ya existe un cliente con ese numero_documento")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CrearClienteEntrada) (*Cliente, error) {
	input = normalizarCrearEntrada(input)

	if err := validarCrearEntrada(input); err != nil {
		return nil, err
	}

	cliente, err := s.repo.Create(ctx, input)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDocumentoClienteYaExiste
		}
		return nil, err
	}

	return cliente, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Cliente, error) {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return nil, ErrClienteInvalido
	}

	cliente, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClienteNoEncontrado
		}
		return nil, err
	}

	return cliente, nil
}

func (s *Service) List(ctx context.Context, busqueda string, limit, offset int) ([]Cliente, error) {
	busqueda = strings.TrimSpace(busqueda)
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)

	return s.repo.List(ctx, busqueda, limit, offset)
}

func (s *Service) Update(ctx context.Context, id string, input ActualizarClienteEntrada) (*Cliente, error) {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return nil, ErrClienteInvalido
	}

	input = normalizarActualizarEntrada(input)

	if err := validarActualizarEntrada(input); err != nil {
		return nil, err
	}

	cliente, err := s.repo.Update(ctx, id, input)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClienteNoEncontrado
		}
		if isUniqueViolation(err) {
			return nil, ErrDocumentoClienteYaExiste
		}
		return nil, err
	}

	return cliente, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return ErrClienteInvalido
	}

	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrClienteNoEncontrado
	}

	return nil
}

func normalizarCrearEntrada(input CrearClienteEntrada) CrearClienteEntrada {
	input.Nombres = normalizarTexto(input.Nombres)
	input.Apellidos = normalizarTexto(input.Apellidos)
	input.Telefono = strings.TrimSpace(input.Telefono)
	input.Correo = strings.ToLower(strings.TrimSpace(input.Correo))
	input.Direccion = strings.TrimSpace(input.Direccion)
	input.TipoDocumento = strings.ToLower(strings.TrimSpace(input.TipoDocumento))
	input.NumeroDocumento = strings.TrimSpace(input.NumeroDocumento)
	input.Notas = strings.TrimSpace(input.Notas)
	return input
}

func normalizarActualizarEntrada(input ActualizarClienteEntrada) ActualizarClienteEntrada {
	input.Nombres = normalizarTexto(input.Nombres)
	input.Apellidos = normalizarTexto(input.Apellidos)
	input.Telefono = strings.TrimSpace(input.Telefono)
	input.Correo = strings.ToLower(strings.TrimSpace(input.Correo))
	input.Direccion = strings.TrimSpace(input.Direccion)
	input.TipoDocumento = strings.ToLower(strings.TrimSpace(input.TipoDocumento))
	input.NumeroDocumento = strings.TrimSpace(input.NumeroDocumento)
	input.Notas = strings.TrimSpace(input.Notas)
	return input
}

func normalizarTexto(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func validarCrearEntrada(input CrearClienteEntrada) error {
	if input.Nombres == "" || input.Apellidos == "" {
		return ErrClienteInvalido
	}
	if !datosContactoValidos(input.Correo, input.Telefono) {
		return ErrClienteInvalido
	}
	if !documentoValido(input.TipoDocumento, input.NumeroDocumento) {
		return ErrClienteInvalido
	}

	return nil
}

func validarActualizarEntrada(input ActualizarClienteEntrada) error {
	if input.Nombres == "" || input.Apellidos == "" {
		return ErrClienteInvalido
	}
	if !datosContactoValidos(input.Correo, input.Telefono) {
		return ErrClienteInvalido
	}
	if !documentoValido(input.TipoDocumento, input.NumeroDocumento) {
		return ErrClienteInvalido
	}

	return nil
}

func datosContactoValidos(correo string, telefono string) bool {
	if correo != "" {
		address, err := mail.ParseAddress(correo)
		if err != nil || address.Address != correo {
			return false
		}
	}

	if telefono != "" && len(telefono) < 7 {
		return false
	}

	return true
}

func documentoValido(tipoDocumento string, numeroDocumento string) bool {
	if tipoDocumento == "" && numeroDocumento == "" {
		return true
	}
	if tipoDocumento == "" || numeroDocumento == "" {
		return false
	}

	tiposPermitidos := map[string]bool{
		"cc":  true,
		"ce":  true,
		"nit": true,
		"ti":  true,
		"pa":  true,
	}

	return tiposPermitidos[tipoDocumento] && len(numeroDocumento) >= 5
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
