package motos

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

func (h *Handler) RegisterRoutes(mux *http.ServeMux, middlewares ...func(http.Handler) http.Handler) {
	mux.Handle("GET /api/motos", wrap(http.HandlerFunc(h.List), middlewares...))
	mux.Handle("POST /api/motos", wrap(http.HandlerFunc(h.Create), middlewares...))
	mux.Handle("GET /api/motos/autocompletar", wrap(http.HandlerFunc(h.AutocompleteByPlaca), middlewares...))
	mux.Handle("GET /api/motos/por-placa/{placa}", wrap(http.HandlerFunc(h.GetByPlaca), middlewares...))
	mux.Handle("GET /api/motos/{id}", wrap(http.HandlerFunc(h.GetByID), middlewares...))
	mux.Handle("PUT /api/motos/{id}", wrap(http.HandlerFunc(h.Update), middlewares...))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	busqueda := r.URL.Query().Get("busqueda")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)

	motos, err := h.service.List(r.Context(), busqueda, limit, offset)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "error_interno", "error listando motos")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       motos,
		"pagination": shared.PaginationBody{Limit: limit, Offset: offset},
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CrearMotoEntrada

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	moto, err := h.service.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrMotoInvalida) {
			shared.WriteError(w, http.StatusBadRequest, "datos_invalidos", "placa, marca, modelo, cilindraje y cliente_id son obligatorios; valida kilometraje_actual y anio")
			return
		}

		if errors.Is(err, ErrPlacaMotoYaExiste) {
			shared.WriteError(w, http.StatusConflict, "placa_duplicada", "ya existe una moto con esa placa")
			return
		}

		if errors.Is(err, ErrClienteMotoNoEncontrado) {
			shared.WriteError(w, http.StatusNotFound, "cliente_no_encontrado", "cliente no encontrado")
			return
		}

		shared.WriteError(w, http.StatusInternalServerError, "error_interno", "error creando moto")
		return
	}

	w.Header().Set("Location", "/api/motos/"+moto.ID)
	shared.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": moto,
	})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	moto, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrMotoNoEncontrada) {
			shared.WriteError(w, http.StatusNotFound, "moto_no_encontrada", "moto no encontrada")
			return
		}
		if errors.Is(err, ErrMotoInvalida) {
			shared.WriteError(w, http.StatusBadRequest, "id_invalido", "id invalido")
			return
		}

		shared.WriteError(w, http.StatusInternalServerError, "error_interno", "error consultando moto")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": moto,
	})
}

func (h *Handler) GetByPlaca(w http.ResponseWriter, r *http.Request) {
	placa := r.PathValue("placa")

	moto, err := h.service.GetByPlaca(r.Context(), placa)
	if err != nil {
		if errors.Is(err, ErrMotoNoEncontrada) {
			shared.WriteError(w, http.StatusNotFound, "moto_no_encontrada", "moto no encontrada")
			return
		}
		if errors.Is(err, ErrMotoInvalida) {
			shared.WriteError(w, http.StatusBadRequest, "placa_invalida", "placa invalida")
			return
		}

		shared.WriteError(w, http.StatusInternalServerError, "error_interno", "error consultando moto por placa")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": moto,
	})
}

func (h *Handler) AutocompleteByPlaca(w http.ResponseWriter, r *http.Request) {
	placa := r.URL.Query().Get("placa")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	items, err := h.service.AutocompleteByPlaca(r.Context(), placa, limit)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "error_interno", "error autocompletando placas")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var input ActualizarMotoEntrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	moto, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, ErrMotoNoEncontrada) {
			shared.WriteError(w, http.StatusNotFound, "moto_no_encontrada", "moto no encontrada")
			return
		}
		if errors.Is(err, ErrMotoInvalida) {
			shared.WriteError(w, http.StatusBadRequest, "datos_invalidos", "datos invalidos")
			return
		}
		if errors.Is(err, ErrPlacaMotoYaExiste) {
			shared.WriteError(w, http.StatusConflict, "placa_duplicada", "ya existe una moto con esa placa")
			return
		}
		if errors.Is(err, ErrClienteMotoNoEncontrado) {
			shared.WriteError(w, http.StatusNotFound, "cliente_no_encontrado", "cliente no encontrado")
			return
		}

		shared.WriteError(w, http.StatusInternalServerError, "error_interno", "error actualizando moto")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": moto,
	})
}

func wrap(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
