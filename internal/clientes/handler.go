package clientes

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

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/clientes", h.List)
	mux.HandleFunc("POST /api/clientes", h.Create)
	mux.HandleFunc("GET /api/clientes/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/clientes/{id}", h.Update)
	mux.HandleFunc("DELETE /api/clientes/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	busqueda := r.URL.Query().Get("busqueda")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)

	clientes, err := h.service.List(r.Context(), busqueda, limit, offset)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "error_interno", "error listando clientes")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       clientes,
		"pagination": shared.PaginationBody{Limit: limit, Offset: offset},
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CrearClienteEntrada

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	cliente, err := h.service.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrClienteInvalido) {
			shared.WriteError(w, http.StatusBadRequest, "datos_invalidos", "nombres y apellidos son obligatorios; valida correo, telefono y documento")
			return
		}

		if errors.Is(err, ErrDocumentoClienteYaExiste) {
			shared.WriteError(w, http.StatusConflict, "documento_duplicado", "ya existe un cliente con ese numero_documento")
			return
		}

		shared.WriteError(w, http.StatusInternalServerError, "error_interno", "error creando cliente")
		return
	}

	w.Header().Set("Location", "/api/clientes/"+cliente.ID)
	shared.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": cliente,
	})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	cliente, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrClienteNoEncontrado) {
			shared.WriteError(w, http.StatusNotFound, "cliente_no_encontrado", "cliente no encontrado")
			return
		}

		if errors.Is(err, ErrClienteInvalido) {
			shared.WriteError(w, http.StatusBadRequest, "id_invalido", "id invalido")
			return
		}

		shared.WriteError(w, http.StatusInternalServerError, "error_interno", "error consultando cliente")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": cliente,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var input ActualizarClienteEntrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	cliente, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, ErrClienteNoEncontrado) {
			shared.WriteError(w, http.StatusNotFound, "cliente_no_encontrado", "cliente no encontrado")
			return
		}

		if errors.Is(err, ErrClienteInvalido) {
			shared.WriteError(w, http.StatusBadRequest, "datos_invalidos", "datos invalidos")
			return
		}

		if errors.Is(err, ErrDocumentoClienteYaExiste) {
			shared.WriteError(w, http.StatusConflict, "documento_duplicado", "ya existe un cliente con ese numero_documento")
			return
		}

		shared.WriteError(w, http.StatusInternalServerError, "error_interno", "error actualizando cliente")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": cliente,
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.service.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrClienteNoEncontrado) {
			shared.WriteError(w, http.StatusNotFound, "cliente_no_encontrado", "cliente no encontrado")
			return
		}

		if errors.Is(err, ErrClienteInvalido) {
			shared.WriteError(w, http.StatusBadRequest, "id_invalido", "id invalido")
			return
		}

		shared.WriteError(w, http.StatusInternalServerError, "error_interno", "error eliminando cliente")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "cliente desactivado correctamente",
	})
}
