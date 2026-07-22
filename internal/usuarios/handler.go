package usuarios

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"tecnomotos/internal/shared"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterAdminRoutes(mux *http.ServeMux, middlewares ...func(http.Handler) http.Handler) {
	mux.Handle("GET /api/admin/roles", wrap(http.HandlerFunc(h.ListRoles), middlewares...))
	mux.Handle("GET /api/admin/usuarios", wrap(http.HandlerFunc(h.List), middlewares...))
	mux.Handle("POST /api/admin/usuarios", wrap(http.HandlerFunc(h.Create), middlewares...))
	mux.Handle("GET /api/admin/usuarios/{id}", wrap(http.HandlerFunc(h.GetByID), middlewares...))
	mux.Handle("PUT /api/admin/usuarios/{id}", wrap(http.HandlerFunc(h.Update), middlewares...))
	mux.Handle("PATCH /api/admin/usuarios/{id}/rol", wrap(http.HandlerFunc(h.AssignRole), middlewares...))
	mux.Handle("DELETE /api/admin/usuarios/{id}", wrap(http.HandlerFunc(h.Deactivate), middlewares...))
}

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.service.ListRoles(r.Context())
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "error_interno", "error listando roles")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": roles,
	})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	busqueda := r.URL.Query().Get("busqueda")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)

	usuarios, err := h.service.List(r.Context(), busqueda, limit, offset)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "error_interno", "error listando usuarios")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       usuarios,
		"pagination": shared.PaginationBody{Limit: limit, Offset: offset},
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CrearUsuarioEntrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	usuario, err := h.service.Create(r.Context(), input)
	if err != nil {
		writeUsuarioError(w, err, "error creando usuario")
		return
	}

	w.Header().Set("Location", "/api/admin/usuarios/"+usuario.ID)
	shared.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": usuario,
	})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	usuario, err := h.service.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeUsuarioError(w, err, "error consultando usuario")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": usuario,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var input ActualizarUsuarioEntrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	usuario, err := h.service.Update(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeUsuarioError(w, err, "error actualizando usuario")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": usuario,
	})
}

func (h *Handler) AssignRole(w http.ResponseWriter, r *http.Request) {
	var input AsignarRolEntrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	usuario, err := h.service.AssignRole(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeUsuarioError(w, err, "error asignando rol")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": usuario,
	})
}

func (h *Handler) Deactivate(w http.ResponseWriter, r *http.Request) {
	err := h.service.Deactivate(r.Context(), r.PathValue("id"))
	if err != nil {
		writeUsuarioError(w, err, "error desactivando usuario")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "usuario desactivado correctamente",
	})
}

func writeUsuarioError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, ErrUsuarioInvalido):
		shared.WriteError(w, http.StatusBadRequest, "datos_invalidos", "datos del usuario invalidos")
	case errors.Is(err, ErrUsuarioNoEncontrado):
		shared.WriteError(w, http.StatusNotFound, "usuario_no_encontrado", "usuario no encontrado")
	case errors.Is(err, ErrCorreoUsuarioYaExiste):
		shared.WriteError(w, http.StatusConflict, "correo_duplicado", "ya existe un usuario con ese correo")
	case errors.Is(err, ErrRolNoEncontrado):
		shared.WriteError(w, http.StatusBadRequest, "rol_invalido", "rol no encontrado")
	default:
		shared.WriteError(w, http.StatusInternalServerError, "error_interno", fallback)
	}
}

func wrap(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
