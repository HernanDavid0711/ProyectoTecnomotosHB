package ordenes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"tecnomotos/internal/auth"
	"tecnomotos/internal/shared"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterAdminRoutes(mux *http.ServeMux, middlewares ...func(http.Handler) http.Handler) {
	mux.Handle("GET /api/admin/formato-orden", wrap(http.HandlerFunc(h.GetFormato), middlewares...))
	mux.Handle("GET /api/admin/catalogo-items", wrap(http.HandlerFunc(h.AdminListCatalogoItems), middlewares...))
	mux.Handle("POST /api/admin/catalogo-items", wrap(http.HandlerFunc(h.AdminCreateCatalogoItem), middlewares...))
	mux.Handle("GET /api/admin/catalogo-items/{id}", wrap(http.HandlerFunc(h.AdminGetCatalogoItem), middlewares...))
	mux.Handle("PUT /api/admin/catalogo-items/{id}", wrap(http.HandlerFunc(h.AdminUpdateCatalogoItem), middlewares...))
	mux.Handle("DELETE /api/admin/catalogo-items/{id}", wrap(http.HandlerFunc(h.AdminDeleteCatalogoItem), middlewares...))
	mux.Handle("GET /api/admin/ordenes", wrap(http.HandlerFunc(h.AdminList), middlewares...))
	mux.Handle("POST /api/admin/ordenes", wrap(http.HandlerFunc(h.AdminCreate), middlewares...))
	mux.Handle("GET /api/admin/ordenes/{id}", wrap(http.HandlerFunc(h.AdminGetByID), middlewares...))
	mux.Handle("PATCH /api/admin/ordenes/{id}", wrap(http.HandlerFunc(h.AdminUpdate), middlewares...))
	mux.Handle("PATCH /api/admin/ordenes/{id}/asignacion", wrap(http.HandlerFunc(h.AdminAssign), middlewares...))
}

func (h *Handler) RegisterEmpleadoRoutes(mux *http.ServeMux, middlewares ...func(http.Handler) http.Handler) {
	mux.Handle("GET /api/empleado/ordenes", wrap(http.HandlerFunc(h.EmpleadoList), middlewares...))
	mux.Handle("GET /api/empleado/ordenes/{id}", wrap(http.HandlerFunc(h.EmpleadoGetByID), middlewares...))
	mux.Handle("PATCH /api/empleado/ordenes/{id}/progreso", wrap(http.HandlerFunc(h.EmpleadoUpdateProgress), middlewares...))
}

func (h *Handler) BuscarClientesConMotos(w http.ResponseWriter, r *http.Request) {
	busqueda := r.URL.Query().Get("busqueda")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	clientes, err := h.service.BuscarClientesConMotos(r.Context(), busqueda, limit)
	if err != nil {
		writeOrdenError(w, err, "error buscando clientes")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": clientes,
	})
}

func (h *Handler) ListCatalogoItems(w http.ResponseWriter, r *http.Request) {
	busqueda := r.URL.Query().Get("busqueda")
	tipoItem := r.URL.Query().Get("tipo_item")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)

	items, err := h.service.ListCatalogoItems(r.Context(), busqueda, tipoItem, limit, offset)
	if err != nil {
		writeOrdenError(w, err, "error listando catalogo de items")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       items,
		"pagination": shared.PaginationBody{Limit: limit, Offset: offset},
	})
}

func (h *Handler) AdminListCatalogoItems(w http.ResponseWriter, r *http.Request) {
	busqueda := r.URL.Query().Get("busqueda")
	tipoItem := r.URL.Query().Get("tipo_item")
	activo := r.URL.Query().Get("activo")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)

	items, err := h.service.AdminListCatalogoItems(r.Context(), busqueda, tipoItem, activo, limit, offset)
	if err != nil {
		writeOrdenError(w, err, "error listando catalogo de items")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       items,
		"pagination": shared.PaginationBody{Limit: limit, Offset: offset},
	})
}

func (h *Handler) AdminGetCatalogoItem(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetCatalogoItemByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeOrdenError(w, err, "error consultando item de catalogo")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": item,
	})
}

func (h *Handler) AdminCreateCatalogoItem(w http.ResponseWriter, r *http.Request) {
	var input GuardarCatalogoItemTrabajoEntrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	item, err := h.service.CreateCatalogoItem(r.Context(), input)
	if err != nil {
		writeOrdenError(w, err, "error creando item de catalogo")
		return
	}

	w.Header().Set("Location", "/api/admin/catalogo-items/"+item.ID)
	shared.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": item,
	})
}

func (h *Handler) AdminUpdateCatalogoItem(w http.ResponseWriter, r *http.Request) {
	var input GuardarCatalogoItemTrabajoEntrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	item, err := h.service.UpdateCatalogoItem(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeOrdenError(w, err, "error actualizando item de catalogo")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": item,
	})
}

func (h *Handler) AdminDeleteCatalogoItem(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeactivateCatalogoItem(r.Context(), r.PathValue("id")); err != nil {
		writeOrdenError(w, err, "error desactivando item de catalogo")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListEmpleadosAsignables(w http.ResponseWriter, r *http.Request) {
	busqueda := r.URL.Query().Get("busqueda")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)

	empleados, err := h.service.ListEmpleadosAsignables(r.Context(), busqueda, limit, offset)
	if err != nil {
		writeOrdenError(w, err, "error listando empleados")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       empleados,
		"pagination": shared.PaginationBody{Limit: limit, Offset: offset},
	})
}

func (h *Handler) GetFormato(w http.ResponseWriter, r *http.Request) {
	formato, err := h.service.GetFormato(r.Context())
	if err != nil {
		writeOrdenError(w, err, "error consultando formato de orden")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": formato,
	})
}

func (h *Handler) SaveFormato(w http.ResponseWriter, r *http.Request) {
	var input GuardarFormatoOrdenEntrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	formato, err := h.service.SaveFormato(r.Context(), input)
	if err != nil {
		writeOrdenError(w, err, "error guardando formato de orden")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": formato,
	})
}

func (h *Handler) CreadorCreateWithItems(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "no_autenticado", "token bearer requerido")
		return
	}

	var input CrearOrdenConItemsEntrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	orden, err := h.service.CreateWithItems(r.Context(), input, claims.UserID)
	if err != nil {
		writeOrdenError(w, err, "error creando orden de trabajo")
		return
	}

	w.Header().Set("Location", "/api/creador-ordenes/ordenes/"+orden.ID)
	shared.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": orden,
	})
}

func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) {
	busqueda := r.URL.Query().Get("busqueda")
	estado := r.URL.Query().Get("estado")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)

	ordenes, err := h.service.List(r.Context(), busqueda, estado, limit, offset)
	if err != nil {
		writeOrdenError(w, err, "error listando ordenes de trabajo")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       ordenes,
		"pagination": shared.PaginationBody{Limit: limit, Offset: offset},
	})
}

func (h *Handler) AdminCreate(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "no_autenticado", "token bearer requerido")
		return
	}

	var input CrearOrdenEntrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	orden, err := h.service.Create(r.Context(), input, claims.UserID)
	if err != nil {
		writeOrdenError(w, err, "error creando orden de trabajo")
		return
	}

	w.Header().Set("Location", "/api/admin/ordenes/"+orden.ID)
	shared.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": orden,
	})
}

func (h *Handler) AdminGetByID(w http.ResponseWriter, r *http.Request) {
	orden, err := h.service.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeOrdenError(w, err, "error consultando orden de trabajo")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": orden,
	})
}

func (h *Handler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	var input ActualizarOrdenAdminEntrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	orden, err := h.service.UpdateAdmin(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeOrdenError(w, err, "error actualizando orden de trabajo")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": orden,
	})
}

func (h *Handler) AdminAssign(w http.ResponseWriter, r *http.Request) {
	var input AsignarOrdenEntrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	orden, err := h.service.Assign(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeOrdenError(w, err, "error asignando orden de trabajo")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": orden,
	})
}

func (h *Handler) EmpleadoList(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "no_autenticado", "token bearer requerido")
		return
	}

	estado := r.URL.Query().Get("estado")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)

	ordenes, err := h.service.ListAssigned(r.Context(), claims.UserID, estado, limit, offset)
	if err != nil {
		writeOrdenError(w, err, "error listando ordenes asignadas")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       ordenes,
		"pagination": shared.PaginationBody{Limit: limit, Offset: offset},
	})
}

func (h *Handler) EmpleadoGetByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "no_autenticado", "token bearer requerido")
		return
	}

	orden, err := h.service.GetAssignedByID(r.Context(), r.PathValue("id"), claims.UserID)
	if err != nil {
		writeOrdenError(w, err, "error consultando orden asignada")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": orden,
	})
}

func (h *Handler) EmpleadoUpdateProgress(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "no_autenticado", "token bearer requerido")
		return
	}

	var input ActualizarProgresoEntrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "json_invalido", "json invalido")
		return
	}

	orden, err := h.service.UpdateProgress(r.Context(), r.PathValue("id"), claims.UserID, input)
	if err != nil {
		writeOrdenError(w, err, "error actualizando progreso")
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"data": orden,
	})
}

func writeOrdenError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, ErrOrdenInvalida):
		shared.WriteError(w, http.StatusBadRequest, "datos_invalidos", "datos de la orden de trabajo invalidos")
	case errors.Is(err, ErrOrdenNoEncontrada):
		shared.WriteError(w, http.StatusNotFound, "orden_no_encontrada", "orden de trabajo no encontrada")
	case errors.Is(err, ErrCatalogoItemNoEncontrado):
		shared.WriteError(w, http.StatusNotFound, "catalogo_item_no_encontrado", "item de catalogo no encontrado")
	case errors.Is(err, ErrCatalogoItemDuplicado):
		shared.WriteError(w, http.StatusConflict, "catalogo_item_duplicado", "ya existe un item con ese tipo y nombre")
	case errors.Is(err, ErrReferenciaNoEncontrada):
		shared.WriteError(w, http.StatusBadRequest, "referencia_invalida", "cliente, moto o usuario relacionado no encontrado")
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
