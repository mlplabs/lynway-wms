package dto

// LocationDTO описывает координаты ячейки или контейнера на складе.
type LocationDTO struct {
	WhsID       int  `json:"whs_id"`
	ZoneID      int  `json:"zone_id"`
	CellID      *int `json:"cell_id,omitempty"`
	ContainerID *int `json:"container_id,omitempty"`
}

// StorageOperationResponse возвращает идентификатор транзакционной проводки в регистре storage.
type StorageOperationResponse struct {
	RowID string `json:"row_id"`
}

// ReceiveItemRequest запрос на приемку товара в зону, ячейку или контейнер.
type ReceiveItemRequest struct {
	ItemID   int         `json:"item_id"`
	Quantity int         `json:"quantity"`
	To       LocationDTO `json:"to"`
}

// MoveItemRequest запрос на перемещение товара между зонами, ячейками или контейнерами.
type MoveItemRequest struct {
	ItemID   int         `json:"item_id"`
	Quantity int         `json:"quantity"`
	From     LocationDTO `json:"from"`
	To       LocationDTO `json:"to"`
}

// DispatchItemRequest запрос на выбытие (списание) товара в системную зону DISPATCH склада.
type DispatchItemRequest struct {
	ItemID   int         `json:"item_id"`
	Quantity int         `json:"quantity"`
	From     LocationDTO `json:"from"`
}

// PlaceContainerRequest запрос на первичное размещение внешнего (корневого) контейнера в зону или ячейку.
type PlaceContainerRequest struct {
	ContainerID int         `json:"container_id"`
	To          LocationDTO `json:"to"`
}

// MoveContainerRequest запрос на перемещение внешнего контейнера вместе с его вложенным деревом.
type MoveContainerRequest struct {
	ContainerID int         `json:"container_id"`
	From        LocationDTO `json:"from"`
	To          LocationDTO `json:"to"`
}

// DispatchContainerRequest запрос на отгрузку (списание) внешнего контейнера в зону DISPATCH.
type DispatchContainerRequest struct {
	ContainerID int         `json:"container_id"`
	From        LocationDTO `json:"from"`
}

// NestContainerRequest запрос на вложение одного контейнера (ChildID) внутрь другого (TargetParentID).
type NestContainerRequest struct {
	ChildID        int `json:"child_id"`
	TargetParentID int `json:"target_parent_id"`
}

// UnnestContainerRequest запрос на извлечение вложенного контейнера наружу (делает его корневым с parent_id = 0).
type UnnestContainerRequest struct {
	ContainerID int         `json:"container_id"`
	To          LocationDTO `json:"to"`
}
