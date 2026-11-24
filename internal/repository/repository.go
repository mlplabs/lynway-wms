package repository

import (
	"context"
	"database/sql"

	"github.com/mlplabs/lynway-wms/internal/model"
	"github.com/mlplabs/lynway-wms/internal/repository/dbo"
)

type DbTx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type AuthRepository interface {
	Login(ctx context.Context, email string, password string) (*model.LoginUserResponse, error)
}

type PgRepository interface {
	WithinTx(ctx context.Context, fn func(txRepo PgRepository) error) error

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
}

type pgRepository struct {
	conn *sql.DB
	dbTx DbTx
}

func NewWhsRepository(conn *sql.DB) *pgRepository {
	return &pgRepository{
		conn: conn,
		dbTx: conn,
	}
}

// WithTx возвращает новый экземпляр репозитория, работающий внутри переданной транзакции
func (r *pgRepository) WithTx(tx *sql.Tx) PgRepository {
	return &pgRepository{conn: r.conn, dbTx: tx}
}

// WithinTx — хелпер, чтобы выполнять методы репозитория в транзакции
func (r *pgRepository) WithinTx(ctx context.Context, fn func(txRepo PgRepository) error) error {
	tx, err := r.conn.BeginTx(ctx, nil)
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
