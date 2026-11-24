package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mlplabs/lynway-wms/internal/model"
	"github.com/mlplabs/lynway-wms/internal/repository/dbo"
)

type DbTx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type AuthRepository interface {
	Login(ctx context.Context, email string, password string) (*model.LoginUserResponse, error)
}

type PgRepository interface {
	WithinTx(ctx context.Context, fn func(txRepo PgRepository) error) error
	AcquireLock(ctx context.Context, lockKey string) error
	AcquireSharedLock(ctx context.Context, lockKey string) error

	GetContainerById(ctx context.Context, id int) (*model.Container, error)
	CreateContainer(ctx context.Context, name *string, containerType *int) (*model.Container, error)
	PutContainer(ctx context.Context, this *model.Container, target *model.Container) error
	GetContainer(ctx context.Context, this *model.Container) error
	GetChildrenContainers(ctx context.Context, this *model.Container) ([]model.Container, error)
	GetParentContainers(ctx context.Context, this *model.Container) ([]model.Container, error)
	GetParentContainer(ctx context.Context, this *model.Container) (*model.Container, error)
	GetContainerProducts(ctx context.Context, this *model.Container) ([]model.RemainingProduct, error)

	SearchManufacturers(ctx context.Context, query string, limit int) ([]model.Manufacturer, error)
	GetManufacturerById(ctx context.Context, baseID int) (*model.Manufacturer, error)
	FindManufacturerByName(ctx context.Context, name string) (*model.Manufacturer, error)
	CreateManufacturer(ctx context.Context, name string) (*model.Manufacturer, error)

	SearchItems(ctx context.Context, query string, limit int) ([]model.Item, error)
	GetItemById(ctx context.Context, baseID int) (*model.Item, error)
	FindExactItem(ctx context.Context, name string, mfgID *int, sku, barcode *string) (*model.Item, error)
	CreateItem(ctx context.Context, p *dbo.CreateItemParams) (int, error)
	UpdateItem(ctx context.Context, p dbo.UpdateItemParams) error

	GetItemStockLocation(ctx context.Context, id int) ([]model.ItemStockLocation, error)

	GetActiveReservedQty(ctx context.Context, loc model.Location, itemID *int) (int, error)
	HasActiveReservationsForSubtree(ctx context.Context, loc model.Location, containerIDs []int) (bool, error)
	GetZoneMeta(ctx context.Context, zoneID int) (model.ZoneMeta, error)
	GetCellMeta(ctx context.Context, cellID int) (model.CellMeta, error)
	GetContainerMeta(ctx context.Context, containerID int) (model.ContainerMeta, error)
	GetBalance(ctx context.Context, loc model.Location, itemID *int) (int, error)
	GetContainerCurrentLocation(ctx context.Context, containerID int) (*model.Location, error)
	GetRootSubtreeContainerIDs(ctx context.Context, rootID int) ([]int, error)
	GetNestedSubtreeContainerIDs(ctx context.Context, containerID int) ([]int, error)
	GetSubtreeBalances(ctx context.Context, loc model.Location, containerIDs []int) ([]model.ContainerStockItem, error)
	UpdateSubtreeHierarchy(ctx context.Context, targetID int, subtreeIDs []int, newParentID, newRootID, depthDelta int) error
	GetDispatchLocation(ctx context.Context, whsID int) (model.Location, error)
	InsertRows(ctx context.Context, rows []model.StorageRow) error
	GetInboxRecord(ctx context.Context, key string) (*model.InboxRecord, error)
	SaveInboxRecord(ctx context.Context, rec model.InboxRecord) error
}

type pgRepository struct {
	conn DbTx
}

func NewWhsRepository(conn *sql.DB) *pgRepository {
	return &pgRepository{
		conn: conn,
	}
}

// WithTx возвращает новый экземпляр репозитория, работающий внутри переданной транзакции
func (r *pgRepository) WithTx(tx *sql.Tx) PgRepository {
	return &pgRepository{conn: tx}
}

// WithinTx — хелпер, чтобы выполнять методы репозитория в транзакции
func (r *pgRepository) WithinTx(ctx context.Context, fn func(txRepo PgRepository) error) error {
	// Если мы уже внутри транзакции — просто передаем текущий экземпляр
	if _, ok := r.conn.(*sql.Tx); ok {
		return fn(r)
	}

	conn, ok := r.conn.(*sql.DB)
	if !ok {
		return fmt.Errorf("expected *sql.DB, got %T", r.conn)
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if err := fn(r.WithTx(tx)); err != nil {
		return err
	}

	return tx.Commit()
}

// AcquireLock накладывает эксклюзивную блокировку на строку
func (r *pgRepository) AcquireLock(ctx context.Context, lockKey string) error {
	_, err := r.conn.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, lockKey)
	if err != nil {
		return fmt.Errorf("acquire exclusive lock %s: %w", lockKey, err)
	}
	return nil
}

// AcquireSharedLock накладывает разделяемую блокировку на строку
func (r *pgRepository) AcquireSharedLock(ctx context.Context, lockKey string) error {
	_, err := r.conn.ExecContext(ctx, `SELECT pg_advisory_xact_lock_shared(hashtextextended($1, 0))`, lockKey)
	if err != nil {
		return fmt.Errorf("acquire shared lock %s: %w", lockKey, err)
	}
	return nil
}
