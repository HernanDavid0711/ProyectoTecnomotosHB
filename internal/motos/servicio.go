package motos

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"tecnomotos/internal/shared"
)

var ErrMotoNoEncontrada = errors.New("moto no encontrada")
var ErrMotoInvalida = errors.New("datos de la moto invalidos")
var ErrPlacaMotoYaExiste = errors.New("ya existe una moto con esa placa")
var ErrClienteMotoNoEncontrado = errors.New("cliente de la moto no encontrado")

var placaRegexp = regexp.MustCompile(`^[A-Z0-9]{5,6}$`)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CrearMotoEntrada) (*Moto, error) {
	input = normalizarCrearEntrada(input)

	if err := validarCrearEntrada(input); err != nil {
		return nil, err
	}

	moto, err := s.repo.Create(ctx, input)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrPlacaMotoYaExiste
		}
		if isForeignKeyViolation(err) {
			return nil, ErrClienteMotoNoEncontrado
		}
		return nil, err
	}

	return moto, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Moto, error) {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return nil, ErrMotoInvalida
	}

	moto, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMotoNoEncontrada
		}
		return nil, err
	}

	return moto, nil
}

func (s *Service) GetByPlaca(ctx context.Context, placa string) (*Moto, error) {
	placa = normalizarPlaca(placa)
	if placa == "" {
		return nil, ErrMotoInvalida
	}

	moto, err := s.repo.GetByPlaca(ctx, placa)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMotoNoEncontrada
		}
		return nil, err
	}

	return moto, nil
}

func (s *Service) List(ctx context.Context, busqueda string, limit, offset int) ([]Moto, error) {
	busqueda = strings.TrimSpace(busqueda)
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)

	return s.repo.List(ctx, busqueda, limit, offset)
}

func (s *Service) AutocompleteByPlaca(ctx context.Context, placa string, limit int) ([]MotoAutocompleteItem, error) {
	placa = normalizarPlaca(placa)

	if placa == "" {
		return []MotoAutocompleteItem{}, nil
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}

	return s.repo.AutocompleteByPlaca(ctx, placa, limit)
}

func (s *Service) Update(ctx context.Context, id string, input ActualizarMotoEntrada) (*Moto, error) {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return nil, ErrMotoInvalida
	}

	input = normalizarActualizarEntrada(input)

	if err := validarActualizarEntrada(input); err != nil {
		return nil, err
	}

	moto, err := s.repo.Update(ctx, id, input)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMotoNoEncontrada
		}
		if isUniqueViolation(err) {
			return nil, ErrPlacaMotoYaExiste
		}
		if isForeignKeyViolation(err) {
			return nil, ErrClienteMotoNoEncontrado
		}
		return nil, err
	}

	return moto, nil
}

func normalizarCrearEntrada(input CrearMotoEntrada) CrearMotoEntrada {
	input.Placa = normalizarPlaca(input.Placa)
	input.Marca = strings.TrimSpace(input.Marca)
	input.Modelo = strings.TrimSpace(input.Modelo)
	input.Color = strings.TrimSpace(input.Color)
	input.ClienteID = strings.TrimSpace(input.ClienteID)
	if clienteID, ok := shared.NormalizeNumericID(input.ClienteID); ok {
		input.ClienteID = clienteID
	}
	return input
}

func normalizarActualizarEntrada(input ActualizarMotoEntrada) ActualizarMotoEntrada {
	input.Placa = normalizarPlaca(input.Placa)
	input.Marca = strings.TrimSpace(input.Marca)
	input.Modelo = strings.TrimSpace(input.Modelo)
	input.Color = strings.TrimSpace(input.Color)
	input.ClienteID = strings.TrimSpace(input.ClienteID)
	if clienteID, ok := shared.NormalizeNumericID(input.ClienteID); ok {
		input.ClienteID = clienteID
	}
	return input
}

func normalizarPlaca(value string) string {
	value = strings.TrimSpace(strings.ToUpper(value))
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "-", "")
	return value
}

func validarCrearEntrada(input CrearMotoEntrada) error {
	if input.Placa == "" || input.Marca == "" || input.Modelo == "" {
		return ErrMotoInvalida
	}
	if !placaRegexp.MatchString(input.Placa) {
		return ErrMotoInvalida
	}
	if _, ok := shared.NormalizeNumericID(input.ClienteID); !ok {
		return ErrMotoInvalida
	}
	if input.Cilindraje <= 0 {
		return ErrMotoInvalida
	}
	if input.KilometrajeActual < 0 {
		return ErrMotoInvalida
	}
	if input.Anio < 0 {
		return ErrMotoInvalida
	}
	if input.Anio != 0 && (input.Anio < 1950 || input.Anio > 2100) {
		return ErrMotoInvalida
	}
	return nil
}

func validarActualizarEntrada(input ActualizarMotoEntrada) error {
	if input.Placa == "" || input.Marca == "" || input.Modelo == "" {
		return ErrMotoInvalida
	}
	if !placaRegexp.MatchString(input.Placa) {
		return ErrMotoInvalida
	}
	if _, ok := shared.NormalizeNumericID(input.ClienteID); !ok {
		return ErrMotoInvalida
	}
	if input.Cilindraje <= 0 {
		return ErrMotoInvalida
	}
	if input.KilometrajeActual < 0 {
		return ErrMotoInvalida
	}
	if input.Anio < 0 {
		return ErrMotoInvalida
	}
	if input.Anio != 0 && (input.Anio < 1950 || input.Anio > 2100) {
		return ErrMotoInvalida
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}
