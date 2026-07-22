package ordenes

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"tecnomotos/internal/shared"
)

var ErrOrdenNoEncontrada = errors.New("orden de trabajo no encontrada")
var ErrOrdenInvalida = errors.New("datos de la orden de trabajo invalidos")
var ErrReferenciaNoEncontrada = errors.New("cliente, moto o usuario relacionado no encontrado")
var ErrCatalogoItemNoEncontrado = errors.New("item de catalogo no encontrado")
var ErrCatalogoItemDuplicado = errors.New("ya existe un item con ese tipo y nombre")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CrearOrdenEntrada, usuarioRecibeID string) (*OrdenTrabajo, error) {
	input = normalizarCrearEntrada(input)
	var ok bool
	usuarioRecibeID, ok = shared.NormalizeNumericID(usuarioRecibeID)
	if !ok {
		return nil, ErrOrdenInvalida
	}
	if err := validarCrearEntrada(input); err != nil {
		return nil, err
	}

	orden, err := s.repo.Create(ctx, input, usuarioRecibeID)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	return orden, nil
}

func (s *Service) BuscarClientesConMotos(ctx context.Context, busqueda string, limit int) ([]ClienteConMotos, error) {
	busqueda = strings.TrimSpace(busqueda)
	if len(busqueda) < 2 {
		return []ClienteConMotos{}, nil
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 25 {
		limit = 25
	}

	return s.repo.BuscarClientesConMotos(ctx, busqueda, limit)
}

func (s *Service) ListCatalogoItems(ctx context.Context, busqueda string, tipoItem string, limit int, offset int) ([]CatalogoItemTrabajo, error) {
	busqueda = strings.TrimSpace(busqueda)
	tipoItem = strings.ToLower(strings.TrimSpace(tipoItem))
	if tipoItem != "" && !tipoItemValido(tipoItem) {
		return nil, ErrOrdenInvalida
	}
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)

	return s.repo.ListCatalogoItems(ctx, busqueda, tipoItem, limit, offset)
}

func (s *Service) AdminListCatalogoItems(ctx context.Context, busqueda string, tipoItem string, activo string, limit int, offset int) ([]CatalogoItemTrabajo, error) {
	busqueda = strings.TrimSpace(busqueda)
	tipoItem = strings.ToLower(strings.TrimSpace(tipoItem))
	activo = strings.ToLower(strings.TrimSpace(activo))
	if tipoItem != "" && !tipoItemValido(tipoItem) {
		return nil, ErrOrdenInvalida
	}
	if activo != "" && activo != "true" && activo != "false" {
		return nil, ErrOrdenInvalida
	}
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)

	return s.repo.AdminListCatalogoItems(ctx, busqueda, tipoItem, activo, limit, offset)
}

func (s *Service) GetCatalogoItemByID(ctx context.Context, id string) (*CatalogoItemTrabajo, error) {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return nil, ErrOrdenInvalida
	}

	item, err := s.repo.GetCatalogoItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCatalogoItemNoEncontrado
		}
		return nil, mapRepositoryError(err)
	}
	return item, nil
}

func (s *Service) CreateCatalogoItem(ctx context.Context, input GuardarCatalogoItemTrabajoEntrada) (*CatalogoItemTrabajo, error) {
	input = normalizarCatalogoItemEntrada(input)
	if err := validarCatalogoItemEntrada(input); err != nil {
		return nil, err
	}

	item, err := s.repo.CreateCatalogoItem(ctx, input)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return item, nil
}

func (s *Service) UpdateCatalogoItem(ctx context.Context, id string, input GuardarCatalogoItemTrabajoEntrada) (*CatalogoItemTrabajo, error) {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return nil, ErrOrdenInvalida
	}
	input = normalizarCatalogoItemEntrada(input)
	if err := validarCatalogoItemEntrada(input); err != nil {
		return nil, err
	}

	item, err := s.repo.UpdateCatalogoItem(ctx, id, input)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCatalogoItemNoEncontrado
		}
		return nil, mapRepositoryError(err)
	}
	return item, nil
}

func (s *Service) DeactivateCatalogoItem(ctx context.Context, id string) error {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return ErrOrdenInvalida
	}

	deactivated, err := s.repo.DeactivateCatalogoItem(ctx, id)
	if err != nil {
		return mapRepositoryError(err)
	}
	if !deactivated {
		return ErrCatalogoItemNoEncontrado
	}
	return nil
}

func (s *Service) ListEmpleadosAsignables(ctx context.Context, busqueda string, limit int, offset int) ([]EmpleadoAsignable, error) {
	busqueda = strings.TrimSpace(busqueda)
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)

	return s.repo.ListEmpleadosAsignables(ctx, busqueda, limit, offset)
}

func (s *Service) GetFormato(ctx context.Context) (*FormatoOrdenTrabajo, error) {
	formato, err := s.repo.GetFormato(ctx)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return formato, nil
}

func (s *Service) SaveFormato(ctx context.Context, input GuardarFormatoOrdenEntrada) (*FormatoOrdenTrabajo, error) {
	input.Campos = normalizarCamposFormato(input.Campos)
	if err := validarCamposFormato(input.Campos); err != nil {
		return nil, err
	}

	formato, err := s.repo.SaveFormato(ctx, input)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return formato, nil
}

func (s *Service) CreateWithItems(ctx context.Context, input CrearOrdenConItemsEntrada, usuarioRecibeID string) (*OrdenTrabajo, error) {
	input = normalizarCrearConItemsEntrada(input)
	var ok bool
	usuarioRecibeID, ok = shared.NormalizeNumericID(usuarioRecibeID)
	if !ok {
		return nil, ErrOrdenInvalida
	}
	if err := validarCrearConItemsEntrada(input); err != nil {
		return nil, err
	}

	orden, err := s.repo.CreateWithItems(ctx, input, usuarioRecibeID)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	return orden, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*OrdenTrabajo, error) {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return nil, ErrOrdenInvalida
	}

	orden, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	return orden, nil
}

func (s *Service) GetAssignedByID(ctx context.Context, id string, usuarioID string) (*OrdenTrabajo, error) {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return nil, ErrOrdenInvalida
	}
	usuarioID, ok = shared.NormalizeNumericID(usuarioID)
	if !ok {
		return nil, ErrOrdenInvalida
	}

	orden, err := s.repo.GetAssignedByID(ctx, id, usuarioID)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	return orden, nil
}

func (s *Service) List(ctx context.Context, busqueda string, estado string, limit int, offset int) ([]OrdenTrabajo, error) {
	busqueda = strings.TrimSpace(busqueda)
	estado = normalizarEstado(estado)
	if estado != "" && !estadoValido(estado) {
		return nil, ErrOrdenInvalida
	}
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)

	return s.repo.List(ctx, busqueda, estado, limit, offset)
}

func (s *Service) ListAssigned(ctx context.Context, usuarioID string, estado string, limit int, offset int) ([]OrdenTrabajo, error) {
	var ok bool
	usuarioID, ok = shared.NormalizeNumericID(usuarioID)
	if !ok {
		return nil, ErrOrdenInvalida
	}
	estado = normalizarEstado(estado)
	if estado != "" && !estadoValido(estado) {
		return nil, ErrOrdenInvalida
	}
	limit, offset = shared.NormalizePagination(limit, offset, 20, 100)

	return s.repo.ListAssigned(ctx, usuarioID, estado, limit, offset)
}

func (s *Service) Assign(ctx context.Context, id string, input AsignarOrdenEntrada) (*OrdenTrabajo, error) {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return nil, ErrOrdenInvalida
	}
	input.UsuarioResponsableID, ok = shared.NormalizeNumericID(input.UsuarioResponsableID)
	if !ok {
		return nil, ErrOrdenInvalida
	}

	orden, err := s.repo.Assign(ctx, id, input.UsuarioResponsableID)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	return orden, nil
}

func (s *Service) UpdateAdmin(ctx context.Context, id string, input ActualizarOrdenAdminEntrada) (*OrdenTrabajo, error) {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return nil, ErrOrdenInvalida
	}

	input = normalizarAdminEntrada(input)
	if err := validarAdminEntrada(input); err != nil {
		return nil, err
	}

	orden, err := s.repo.UpdateAdmin(ctx, id, input)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	return orden, nil
}

func (s *Service) UpdateProgress(ctx context.Context, id string, usuarioID string, input ActualizarProgresoEntrada) (*OrdenTrabajo, error) {
	var ok bool
	id, ok = shared.NormalizeNumericID(id)
	if !ok {
		return nil, ErrOrdenInvalida
	}
	usuarioID, ok = shared.NormalizeNumericID(usuarioID)
	if !ok {
		return nil, ErrOrdenInvalida
	}

	input = normalizarProgresoEntrada(input)
	if err := validarProgresoEntrada(input); err != nil {
		return nil, err
	}

	orden, err := s.repo.UpdateProgress(ctx, id, usuarioID, input)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	return orden, nil
}

func normalizarCrearEntrada(input CrearOrdenEntrada) CrearOrdenEntrada {
	if id, ok := shared.NormalizeNumericID(input.ClienteID); ok {
		input.ClienteID = id
	}
	if id, ok := shared.NormalizeNumericID(input.MotoID); ok {
		input.MotoID = id
	}
	if id, ok := shared.NormalizeNumericID(input.UsuarioResponsableID); ok {
		input.UsuarioResponsableID = id
	}
	input.TipoServicio = strings.TrimSpace(input.TipoServicio)
	input.DescripcionFalla = strings.TrimSpace(input.DescripcionFalla)
	input.Observaciones = strings.TrimSpace(input.Observaciones)
	return input
}

func normalizarCrearConItemsEntrada(input CrearOrdenConItemsEntrada) CrearOrdenConItemsEntrada {
	if id, ok := shared.NormalizeNumericID(input.ClienteID); ok {
		input.ClienteID = id
	}
	if id, ok := shared.NormalizeNumericID(input.MotoID); ok {
		input.MotoID = id
	}
	if id, ok := shared.NormalizeNumericID(input.UsuarioResponsableID); ok {
		input.UsuarioResponsableID = id
	}
	input.TipoServicio = strings.TrimSpace(input.TipoServicio)
	input.DescripcionFalla = strings.TrimSpace(input.DescripcionFalla)
	input.Observaciones = strings.TrimSpace(input.Observaciones)
	for index := range input.Items {
		if id, ok := shared.NormalizeNumericID(input.Items[index].CatalogoItemTrabajoID); ok {
			input.Items[index].CatalogoItemTrabajoID = id
		}
	}
	return input
}

func normalizarProgresoEntrada(input ActualizarProgresoEntrada) ActualizarProgresoEntrada {
	input.Estado = normalizarEstado(input.Estado)
	input.Diagnostico = strings.TrimSpace(input.Diagnostico)
	input.TrabajosRealizados = strings.TrimSpace(input.TrabajosRealizados)
	input.Observaciones = strings.TrimSpace(input.Observaciones)
	return input
}

func normalizarAdminEntrada(input ActualizarOrdenAdminEntrada) ActualizarOrdenAdminEntrada {
	if id, ok := shared.NormalizeNumericID(input.UsuarioResponsableID); ok {
		input.UsuarioResponsableID = id
	}
	input.Estado = normalizarEstado(input.Estado)
	input.TipoServicio = strings.TrimSpace(input.TipoServicio)
	input.DescripcionFalla = strings.TrimSpace(input.DescripcionFalla)
	input.Diagnostico = strings.TrimSpace(input.Diagnostico)
	input.TrabajosRealizados = strings.TrimSpace(input.TrabajosRealizados)
	input.Observaciones = strings.TrimSpace(input.Observaciones)
	return input
}

func normalizarCatalogoItemEntrada(input GuardarCatalogoItemTrabajoEntrada) GuardarCatalogoItemTrabajoEntrada {
	input.TipoItem = strings.ToLower(strings.TrimSpace(input.TipoItem))
	input.Nombre = strings.Join(strings.Fields(strings.TrimSpace(input.Nombre)), " ")
	input.Descripcion = strings.TrimSpace(input.Descripcion)
	return input
}

func normalizarCamposFormato(campos []CampoFormato) []CampoFormato {
	normalizados := make([]CampoFormato, 0, len(campos))
	for _, campo := range campos {
		campo.ID = strings.ToLower(strings.TrimSpace(campo.ID))
		campo.Label = strings.TrimSpace(campo.Label)
		campo.Tipo = strings.ToLower(strings.TrimSpace(campo.Tipo))
		if campo.ID == "" {
			campo.ID = campoIDFromLabel(campo.Label)
		}
		normalizados = append(normalizados, campo)
	}
	return normalizados
}

func validarCrearEntrada(input CrearOrdenEntrada) error {
	if _, ok := shared.NormalizeNumericID(input.ClienteID); !ok {
		return ErrOrdenInvalida
	}
	if _, ok := shared.NormalizeNumericID(input.MotoID); !ok {
		return ErrOrdenInvalida
	}
	if input.UsuarioResponsableID != "" {
		if _, ok := shared.NormalizeNumericID(input.UsuarioResponsableID); !ok {
			return ErrOrdenInvalida
		}
	}
	if input.DescripcionFalla == "" || input.KilometrajeIngreso < 0 {
		return ErrOrdenInvalida
	}
	if input.NivelCombustiblePorcentaje < 0 || input.NivelCombustiblePorcentaje > 100 {
		return ErrOrdenInvalida
	}
	return nil
}

func validarCrearConItemsEntrada(input CrearOrdenConItemsEntrada) error {
	baseInput := CrearOrdenEntrada{
		ClienteID:                  input.ClienteID,
		MotoID:                     input.MotoID,
		UsuarioResponsableID:       input.UsuarioResponsableID,
		FechaPrometida:             input.FechaPrometida,
		KilometrajeIngreso:         input.KilometrajeIngreso,
		NivelCombustiblePorcentaje: input.NivelCombustiblePorcentaje,
		TipoServicio:               input.TipoServicio,
		DescripcionFalla:           input.DescripcionFalla,
		Observaciones:              input.Observaciones,
	}
	if err := validarCrearEntrada(baseInput); err != nil {
		return err
	}
	if input.UsuarioResponsableID == "" {
		return ErrOrdenInvalida
	}
	if len(input.Items) == 0 {
		return ErrOrdenInvalida
	}
	if len(input.Items) > 60 {
		return ErrOrdenInvalida
	}
	for _, item := range input.Items {
		if _, ok := shared.NormalizeNumericID(item.CatalogoItemTrabajoID); !ok {
			return ErrOrdenInvalida
		}
		if item.Cantidad <= 0 || item.Descuento < 0 || item.ValorUnitario < 0 {
			return ErrOrdenInvalida
		}
	}
	return nil
}

func validarAdminEntrada(input ActualizarOrdenAdminEntrada) error {
	if input.UsuarioResponsableID != "" {
		if _, ok := shared.NormalizeNumericID(input.UsuarioResponsableID); !ok {
			return ErrOrdenInvalida
		}
	}
	if !estadoValido(input.Estado) {
		return ErrOrdenInvalida
	}
	if input.DescripcionFalla == "" || input.KilometrajeIngreso < 0 {
		return ErrOrdenInvalida
	}
	if input.NivelCombustiblePorcentaje < 0 || input.NivelCombustiblePorcentaje > 100 {
		return ErrOrdenInvalida
	}
	return nil
}

func validarCatalogoItemEntrada(input GuardarCatalogoItemTrabajoEntrada) error {
	if input.Nombre == "" || len(input.Nombre) > 150 || !tipoItemValido(input.TipoItem) {
		return ErrOrdenInvalida
	}
	return nil
}

func validarCamposFormato(campos []CampoFormato) error {
	if len(campos) == 0 || len(campos) > 40 {
		return ErrOrdenInvalida
	}
	usados := map[string]bool{}
	for _, campo := range campos {
		if campo.ID == "" || campo.Label == "" || usados[campo.ID] {
			return ErrOrdenInvalida
		}
		if !tipoCampoFormatoValido(campo.Tipo) {
			return ErrOrdenInvalida
		}
		usados[campo.ID] = true
	}
	return nil
}

func validarProgresoEntrada(input ActualizarProgresoEntrada) error {
	if !estadoProgresoEmpleado(input.Estado) {
		return ErrOrdenInvalida
	}
	return nil
}

func normalizarEstado(estado string) string {
	estado = strings.ToLower(strings.TrimSpace(estado))
	if estado == "finalizada" || estado == "finalizado" {
		return EstadoCerrada
	}
	return estado
}

func estadoValido(estado string) bool {
	switch estado {
	case EstadoAbierta, EstadoEnDiagnostico, EstadoAprobacionPendiente, EstadoEnProceso, EstadoListaParaEntrega, EstadoCerrada, EstadoCancelada:
		return true
	default:
		return false
	}
}

func estadoProgresoEmpleado(estado string) bool {
	switch estado {
	case EstadoEnDiagnostico, EstadoEnProceso, EstadoListaParaEntrega, EstadoCerrada:
		return true
	default:
		return false
	}
}

func tipoItemValido(tipoItem string) bool {
	switch tipoItem {
	case "mano_obra", "repuesto", "servicio_externo", "insumo":
		return true
	default:
		return false
	}
}

func tipoCampoFormatoValido(tipo string) bool {
	switch tipo {
	case "texto", "numero", "area", "fecha", "checkbox":
		return true
	default:
		return false
	}
}

func campoIDFromLabel(label string) string {
	value := strings.ToLower(label)
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n",
		"Á", "a", "É", "e", "Í", "i", "Ó", "o", "Ú", "u", "Ñ", "n",
	)
	value = replacer.Replace(value)
	builder := strings.Builder{}
	lastUnderscore := false
	for _, char := range value {
		isLetter := char >= 'a' && char <= 'z'
		isNumber := char >= '0' && char <= '9'
		if isLetter || isNumber {
			builder.WriteRune(char)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			builder.WriteRune('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(builder.String(), "_")
}

func mapRepositoryError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrOrdenNoEncontrada
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23503" {
			return ErrReferenciaNoEncontrada
		}
		if pgErr.Code == "23505" {
			return ErrCatalogoItemDuplicado
		}
	}

	return err
}
