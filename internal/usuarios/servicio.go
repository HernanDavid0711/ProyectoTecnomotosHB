package usuarios

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"tecnomotos/internal/auth"
	"tecnomotos/internal/shared"
)

var ErrUsuarioNoEncontrado = errors.New("usuario no encontrado")
var ErrUsuarioInvalido = errors.New("datos del usuario invalidos")
var ErrCorreoUsuarioYaExiste = errors.New("ya existe un usuario con ese correo")
var ErrRolNoEncontrado = errors.New("rol no encontrado")
var ErrCredencialesInvalidas = errors.New("credenciales invalidas")
var ErrUsuariosYaExisten = errors.New("ya existen usuarios registrados")

var usuarioRegexp = regexp.MustCompile(`^[a-z0-9._-]{3,40}$`)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) BootstrapAdmin(ctx context.Context, input CrearUsuarioEntrada) (*Usuario, error) {
	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, err
	}
	if total > 0 {
		return nil, ErrUsuariosYaExisten
	}

	input.Rol = RolAdministrador
	if strings.TrimSpace(input.Usuario) == "" {
		input.Usuario = "admin"
	}
	return s.Create(ctx, input)
}

func (s *Service) Create(ctx context.Context, input CrearUsuarioEntrada) (*Usuario, error) {
	input = normalizarCrearEntrada(input)
	if err := validarCrearEntrada(input); err != nil {
		return nil, err
	}

	hashContrasena, err := auth.HashPassword(input.Contrasena)
	if err != nil {
		return nil, ErrUsuarioInvalido
	}

	usuario, err := s.repo.Create(ctx, input, hashContrasena)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	return usuario, nil
}

func (s *Service) ListRoles(ctx context.Context) ([]Rol, error) {
	return s.repo.ListRoles(ctx)
}

func (s *Service) EnsureDefaultRoles(ctx context.Context) error {
	return s.repo.EnsureDefaultRoles(ctx)
}

func (s *Service) Login(ctx context.Context, identificador string, contrasena string) (*Usuario, error) {
	identificador = strings.ToLower(strings.TrimSpace(identificador))
	if identificador == "" || contrasena == "" {
		return nil, ErrCredencialesInvalidas
	}

	usuario, err := s.repo.GetByIdentificador(ctx, identificador)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCredencialesInvalidas
		}
		return nil, err
	}

	if !usuario.Activo || !auth.VerifyPassword(contrasena, usuario.HashContrasena) {
		return nil, ErrCredencialesInvalidas
	}

	if err := s.repo.TouchLastLogin(ctx, usuario.ID); err != nil {
		return nil, err
	}

	return &usuario.Usuario, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Usuario, error) {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return nil, ErrUsuarioInvalido
	}

	usuario, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUsuarioNoEncontrado
		}
		return nil, err
	}

	return usuario, nil
}

func (s *Service) List(ctx context.Context, busqueda string, limit int, offset int) ([]Usuario, error) {
	busqueda = strings.TrimSpace(busqueda)
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)
	return s.repo.List(ctx, busqueda, limit, offset)
}

func (s *Service) Update(ctx context.Context, id string, input ActualizarUsuarioEntrada) (*Usuario, error) {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return nil, ErrUsuarioInvalido
	}

	input = normalizarActualizarEntrada(input)
	if err := validarActualizarEntrada(input); err != nil {
		return nil, err
	}

	var hashContrasena *string
	if input.Contrasena != "" {
		hash, err := auth.HashPassword(input.Contrasena)
		if err != nil {
			return nil, ErrUsuarioInvalido
		}
		hashContrasena = &hash
	}

	usuario, err := s.repo.Update(ctx, id, input, hashContrasena)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	return usuario, nil
}

func (s *Service) AssignRole(ctx context.Context, id string, input AsignarRolEntrada) (*Usuario, error) {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return nil, ErrUsuarioInvalido
	}

	rol := strings.ToLower(strings.TrimSpace(input.Rol))
	if !rolValido(rol) {
		return nil, ErrUsuarioInvalido
	}

	usuario, err := s.repo.AssignRole(ctx, id, rol)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	return usuario, nil
}

func (s *Service) Deactivate(ctx context.Context, id string) error {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return ErrUsuarioInvalido
	}

	deactivated, err := s.repo.Deactivate(ctx, id)
	if err != nil {
		return err
	}
	if !deactivated {
		return ErrUsuarioNoEncontrado
	}

	return nil
}

func normalizarCrearEntrada(input CrearUsuarioEntrada) CrearUsuarioEntrada {
	input.Rol = strings.ToLower(strings.TrimSpace(input.Rol))
	input.Usuario = strings.ToLower(strings.TrimSpace(input.Usuario))
	input.Nombres = normalizarTexto(input.Nombres)
	input.Apellidos = normalizarTexto(input.Apellidos)
	input.Correo = strings.ToLower(strings.TrimSpace(input.Correo))
	input.Telefono = strings.TrimSpace(input.Telefono)
	input.Contrasena = strings.TrimSpace(input.Contrasena)
	return input
}

func normalizarActualizarEntrada(input ActualizarUsuarioEntrada) ActualizarUsuarioEntrada {
	input.Rol = strings.ToLower(strings.TrimSpace(input.Rol))
	input.Usuario = strings.ToLower(strings.TrimSpace(input.Usuario))
	input.Nombres = normalizarTexto(input.Nombres)
	input.Apellidos = normalizarTexto(input.Apellidos)
	input.Correo = strings.ToLower(strings.TrimSpace(input.Correo))
	input.Telefono = strings.TrimSpace(input.Telefono)
	input.Contrasena = strings.TrimSpace(input.Contrasena)
	return input
}

func normalizarTexto(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func validarCrearEntrada(input CrearUsuarioEntrada) error {
	if input.Usuario == "" || input.Nombres == "" || input.Apellidos == "" || input.Correo == "" || input.Contrasena == "" {
		return ErrUsuarioInvalido
	}
	if !usuarioValido(input.Usuario) || !rolValido(input.Rol) || !correoValido(input.Correo) || !telefonoValido(input.Telefono) || len(input.Contrasena) < 8 {
		return ErrUsuarioInvalido
	}
	return nil
}

func validarActualizarEntrada(input ActualizarUsuarioEntrada) error {
	if input.Usuario == "" || input.Nombres == "" || input.Apellidos == "" || input.Correo == "" {
		return ErrUsuarioInvalido
	}
	if !usuarioValido(input.Usuario) || !rolValido(input.Rol) || !correoValido(input.Correo) || !telefonoValido(input.Telefono) {
		return ErrUsuarioInvalido
	}
	if input.Contrasena != "" && len(input.Contrasena) < 8 {
		return ErrUsuarioInvalido
	}
	return nil
}

func usuarioValido(usuario string) bool {
	return usuarioRegexp.MatchString(usuario)
}

func rolValido(rol string) bool {
	switch rol {
	case "administrador", "creador_ordenes", "empleado", "mecanico", "recepcion":
		return true
	default:
		return false
	}
}

func correoValido(correo string) bool {
	address, err := mail.ParseAddress(correo)
	return err == nil && address.Address == correo
}

func telefonoValido(telefono string) bool {
	return telefono == "" || len(telefono) >= 7
}

func mapRepositoryError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrRolNoEncontrado
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return ErrCorreoUsuarioYaExiste
		}
		if pgErr.Code == "23503" {
			return ErrRolNoEncontrado
		}
	}
	return err
}
