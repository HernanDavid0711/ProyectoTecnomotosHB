package sesion

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"tecnomotos/internal/auth"
	"tecnomotos/internal/config"
	"tecnomotos/internal/shared"
	"tecnomotos/internal/usuarios"
)

type Handler struct {
	cfg     *config.Config
	service *usuarios.Service
}

type loginEntrada struct {
	Usuario    string `json:"usuario"`
	Correo     string `json:"correo"`
	Contrasena string `json:"contrasena"`
}

func NewHandler(cfg *config.Config, service *usuarios.Service) *Handler {
	return &Handler{
		cfg:     cfg,
		service: service,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	mux.HandleFunc("POST /api/auth/bootstrap-admin", h.BootstrapAdmin)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.Handle("GET /api/auth/me", authMiddleware(http.HandlerFunc(h.Me)))
}

func (h *Handler) BootstrapAdmin(w http.ResponseWriter, r *http.Request) {
	var input usuarios.CrearUsuarioEntrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	usuario, err := h.service.BootstrapAdmin(r.Context(), input)
	if err != nil {
		writeUsuarioError(w, err, "error creando administrador inicial")
		return
	}

	w.Header().Set("Location", "/api/admin/usuarios/"+usuario.ID)
	shared.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": usuario,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input loginEntrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	identificador := input.Usuario
	if identificador == "" {
		identificador = input.Correo
	}

	usuario, err := h.service.Login(r.Context(), identificador, input.Contrasena)
	if err != nil {
		if errors.Is(err, usuarios.ErrCredencialesInvalidas) {
			shared.WriteError(w, http.StatusUnauthorized, "credenciales_invalidas", "usuario o contrasena incorrectos")
			return
		}

		log.Printf("error iniciando sesion para %q: %v", identificador, err)
		shared.WriteError(w, http.StatusInternalServerError, "error_interno", "error iniciando sesion")
		return
	}

	token, expiresAt, err := auth.GenerateToken(
		h.cfg.JWTSecret,
		h.cfg.JWTIssuer,
		h.cfg.JWTDuration,
		usuario.ID,
		usuario.Correo,
		usuario.Rol,
	)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "error_interno", "error generando token")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"expires_at": expiresAt,
		"data":       usuario,
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "no_autenticado", "token bearer requerido")
		return
	}

	usuario, err := h.service.GetByID(r.Context(), claims.UserID)
	if err != nil {
		writeUsuarioError(w, err, "error consultando usuario autenticado")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": usuario,
	})
}

func writeUsuarioError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, usuarios.ErrUsuariosYaExisten):
		shared.WriteError(w, http.StatusConflict, "usuarios_ya_existen", "ya existen usuarios registrados")
	case errors.Is(err, usuarios.ErrUsuarioInvalido):
		shared.WriteError(w, http.StatusBadRequest, "datos_invalidos", "datos del usuario invalidos")
	case errors.Is(err, usuarios.ErrCorreoUsuarioYaExiste):
		shared.WriteError(w, http.StatusConflict, "correo_duplicado", "ya existe un usuario con ese correo")
	case errors.Is(err, usuarios.ErrRolNoEncontrado):
		shared.WriteError(w, http.StatusBadRequest, "rol_invalido", "rol no encontrado")
	case errors.Is(err, usuarios.ErrUsuarioNoEncontrado):
		shared.WriteError(w, http.StatusNotFound, "usuario_no_encontrado", "usuario no encontrado")
	default:
		shared.WriteError(w, http.StatusInternalServerError, "error_interno", fallback)
	}
}
