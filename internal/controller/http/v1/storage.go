package v1

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	domainErr "github.com/mlplabs/common-go-pkg/pkg/http/errors/custom"
	"github.com/mlplabs/lynway-wms/internal/model"
	"github.com/mlplabs/lynway-wms/pkg/dto"
)

// conflictError реализует интерфейс errors.CommonError из common-go-pkg для HTTP 409.
type conflictError struct {
	code string
	err  error
}

func (e *conflictError) StatusCode() int   { return http.StatusConflict }
func (e *conflictError) ErrorCode() string { return e.code }
func (e *conflictError) Error() string     { return e.err.Error() }

func newConflict(code string, err error) error {
	return &conflictError{code: code, err: err}
}

func locationToModel(in dto.LocationDTO) model.Location {
	return model.Location{
		WhsID:       in.WhsID,
		ZoneID:      in.ZoneID,
		CellID:      in.CellID,
		ContainerID: in.ContainerID,
	}
}

// receiveItem
// @Tags Storage
// @Summary приемка товара в зону, ячейку или контейнер
// @Router /storage/items/receive [post]
// @Param Idempotency-Key header string true "ключ идемпотентности"
// @Param request body dto.ReceiveItemRequest true "параметры приемки"
// @Success 200 {object} wrapper.Data{data=dto.StorageOperationResponse} "результат операции"
func (h *ServiceHandler) receiveItem(w http.ResponseWriter, r *http.Request) (any, error) {
	var req dto.ReceiveItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body: %w", err))
	}

	rowID, err := h.service.ReceiveItem(r.Context(), locationToModel(req.To), req.ItemID, req.Quantity)
	if err != nil {
		return nil, mapStorageError(err)
	}

	return &dto.StorageOperationResponse{RowID: rowID.String()}, nil
}

// moveItem
// @Tags Storage
// @Summary перемещение товара между зонами, ячейками или контейнерами
// @Router /storage/items/move [post]
// @Param Idempotency-Key header string true "ключ идемпотентности"
// @Param request body dto.MoveItemRequest true "параметры перемещения"
// @Success 200 {object} wrapper.Data{data=dto.StorageOperationResponse} "результат операции"
func (h *ServiceHandler) moveItem(w http.ResponseWriter, r *http.Request) (any, error) {
	var req dto.MoveItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body: %w", err))
	}

	rowID, err := h.service.MoveItem(r.Context(), locationToModel(req.From), locationToModel(req.To), req.ItemID, req.Quantity)
	if err != nil {
		return nil, mapStorageError(err)
	}

	return &dto.StorageOperationResponse{RowID: rowID.String()}, nil
}

// dispatchItem
// @Tags Storage
// @Summary выбытие (списание) товара в системную зону DISPATCH
// @Router /storage/items/dispatch [post]
// @Param Idempotency-Key header string true "ключ идемпотентности"
// @Param request body dto.DispatchItemRequest true "параметры выбытия"
// @Success 200 {object} wrapper.Data{data=dto.StorageOperationResponse} "результат операции"
func (h *ServiceHandler) dispatchItem(w http.ResponseWriter, r *http.Request) (any, error) {
	var req dto.DispatchItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body: %w", err))
	}

	rowID, err := h.service.DispatchItem(r.Context(), locationToModel(req.From), req.ItemID, req.Quantity)
	if err != nil {
		return nil, mapStorageError(err)
	}

	return &dto.StorageOperationResponse{RowID: rowID.String()}, nil
}

// placeContainer
// @Tags Storage
// @Summary первичное размещение внешнего контейнера в зону или ячейку
// @Router /storage/containers/place [post]
// @Param Idempotency-Key header string true "ключ идемпотентности"
// @Param request body dto.PlaceContainerRequest true "параметры размещения"
// @Success 200 {object} wrapper.Data{data=dto.StorageOperationResponse} "результат операции"
func (h *ServiceHandler) placeContainer(w http.ResponseWriter, r *http.Request) (any, error) {
	var req dto.PlaceContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body: %w", err))
	}

	rowID, err := h.service.PlaceContainer(r.Context(), locationToModel(req.To), req.ContainerID)
	if err != nil {
		return nil, mapStorageError(err)
	}
	return &dto.StorageOperationResponse{RowID: rowID.String()}, nil
}

// moveContainer
// @Tags Storage
// @Summary перемещение внешнего контейнера вместе с содержимым
// @Router /storage/containers/move [post]
// @Param Idempotency-Key header string true "ключ идемпотентности"
// @Param request body dto.MoveContainerRequest true "параметры перемещения"
// @Success 200 {object} wrapper.Data{data=dto.StorageOperationResponse} "результат операции"
func (h *ServiceHandler) moveContainer(w http.ResponseWriter, r *http.Request) (any, error) {
	var req dto.MoveContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body: %w", err))
	}

	rowID, err := h.service.MoveContainer(r.Context(), locationToModel(req.From), locationToModel(req.To), req.ContainerID)
	if err != nil {
		return nil, mapStorageError(err)
	}
	return &dto.StorageOperationResponse{RowID: rowID.String()}, nil
}

// dispatchContainer
// @Tags Storage
// @Summary выбытие (отгрузка) внешнего контейнера в системную зону DISPATCH
// @Router /storage/containers/dispatch [post]
// @Param Idempotency-Key header string true "ключ идемпотентности"
// @Param request body dto.DispatchContainerRequest true "параметры выбытия"
// @Success 200 {object} wrapper.Data{data=dto.StorageOperationResponse} "результат операции"
func (h *ServiceHandler) dispatchContainer(w http.ResponseWriter, r *http.Request) (any, error) {
	var req dto.DispatchContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body: %w", err))
	}

	rowID, err := h.service.DispatchContainer(r.Context(), locationToModel(req.From), req.ContainerID)
	if err != nil {
		return nil, mapStorageError(err)
	}

	return &dto.StorageOperationResponse{RowID: rowID.String()}, nil
}

// nestContainer
// @Tags Storage
// @Summary вложение одного контейнера внутрь другого
// @Router /storage/containers/nest [post]
// @Param Idempotency-Key header string true "ключ идемпотентности"
// @Param request body dto.NestContainerRequest true "параметры вложения"
// @Success 200 {object} wrapper.Data{data=dto.StorageOperationResponse} "результат операции"
func (h *ServiceHandler) nestContainer(w http.ResponseWriter, r *http.Request) (any, error) {
	var req dto.NestContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body: %w", err))
	}

	rowID, err := h.service.NestContainer(r.Context(), req.ChildID, req.TargetParentID)
	if err != nil {
		return nil, mapStorageError(err)
	}

	return &dto.StorageOperationResponse{RowID: rowID.String()}, nil
}

// unnestContainer
// @Tags Storage
// @Summary извлечение вложенного контейнера наружу
// @Router /storage/containers/unnest [post]
// @Param Idempotency-Key header string true "ключ идемпотентности"
// @Param request body dto.UnnestContainerRequest true "параметры извлечения"
// @Success 200 {object} wrapper.Data{data=dto.StorageOperationResponse} "результат операции"
func (h *ServiceHandler) unnestContainer(w http.ResponseWriter, r *http.Request) (any, error) {
	var req dto.UnnestContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body: %w", err))
	}

	rowID, err := h.service.UnnestContainer(r.Context(), req.ContainerID, locationToModel(req.To))
	if err != nil {
		return nil, mapStorageError(err)
	}

	return &dto.StorageOperationResponse{RowID: rowID.String()}, nil
}

func mapStorageError(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domainErr.NewErrorNoRows(err)
	case errors.Is(err, model.ErrInvalidMovement),
		errors.Is(err, model.ErrZoneWhsMismatch),
		errors.Is(err, model.ErrCellZoneMismatch):
		return domainErr.NewBadRequest(err)
	case errors.Is(err, model.ErrCellNotAllowedIn),
		errors.Is(err, model.ErrCellNotAllowedOut),
		errors.Is(err, model.ErrZoneNotAllowedOut):
		return domainErr.NewForbidden(err)
	case errors.Is(err, model.ErrInsufficientStock):
		return newConflict("INSUFFICIENT_STOCK", err)
	case errors.Is(err, model.ErrLocationReserved):
		return newConflict("LOCATION_RESERVED", err)
	case errors.Is(err, model.ErrContainerNotRoot):
		return newConflict("CONTAINER_NOT_ROOT", err)
	case errors.Is(err, model.ErrContainerHasCell):
		return newConflict("CONTAINER_HAS_CELL", err)
	case errors.Is(err, model.ErrContainerNoCell):
		return newConflict("CONTAINER_NO_CELL", err)
	case errors.Is(err, model.ErrCyclicNesting):
		return newConflict("CYCLIC_NESTING", err)
	default:
		return err
	}
}
