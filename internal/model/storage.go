package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock balance")
	ErrLocationReserved  = errors.New("location or item is reserved by an active picking task")
	ErrCellNotAllowedIn  = errors.New("cell does not allow incoming movements")
	ErrCellNotAllowedOut = errors.New("cell does not allow outgoing movements")
	ErrZoneNotAllowedOut = errors.New("outgoing movements from DISPATCH zone are forbidden")
	ErrCellZoneMismatch  = errors.New("cell does not belong to specified warehouse/zone")
	ErrZoneWhsMismatch   = errors.New("zone does not belong to specified warehouse")
	ErrInvalidMovement   = errors.New("invalid movement parameters")
	ErrContainerNotRoot  = errors.New("container must be a root container (parent_id = 0)")
	ErrContainerHasCell  = errors.New("container is already placed in a cell/zone")
	ErrContainerNoCell   = errors.New("container is not placed in any cell/zone")
	ErrCyclicNesting     = errors.New("cyclic container nesting detected")
)

// Location координаты на складе
type Location struct {
	WhsID       int
	ZoneID      int
	CellID      *int
	ContainerID *int
}

// StorageRow строка регистра хранения
type StorageRow struct {
	RowID       uuid.UUID
	RowTime     time.Time
	WhsID       int
	ZoneID      int
	CellID      *int
	ContainerID *int
	ItemID      *int
	Quantity    int
}

// ZoneMeta метаданные зоны
type ZoneMeta struct {
	ID       int
	WhsID    int
	Type     string
	IsSystem bool
}

// CellMeta метаданные ячейки
type CellMeta struct {
	ID            int
	WhsID         int
	ZoneID        int
	NotAllowedIn  bool
	NotAllowedOut bool
}

// ContainerMeta метаданные контейнера
type ContainerMeta struct {
	ID       int
	ParentID int
	RootID   int
	Depth    int
}

// ContainerStockItem элемент хранения остатков
type ContainerStockItem struct {
	ContainerID int
	ItemID      *int
	Quantity    int
}

type InboxRecord struct {
	Key          string
	ResponseCode int
	ResponseBody []byte
}
