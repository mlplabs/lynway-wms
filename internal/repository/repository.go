package repository

import (
	"context"
	"database/sql"

	"github.com/mlplabs/lynway-wms/internal/model"
)

type AuthRepository interface {
	Login(ctx context.Context, email string, password string) (*model.LoginUserResponse, error)
}

type PgRepository interface {
	GetProductById(ctx context.Context, id int) (*model.Product, error)
	GetProductByBarcode(ctx context.Context, barcode string) (*model.Product, error)

	GetContainerById(ctx context.Context, id int) (*model.Container, error)
	CreateContainer(ctx context.Context, name *string, containerType *int) (*model.Container, error)
	PutContainer(ctx context.Context, this *model.Container, target *model.Container) error
	GetContainer(ctx context.Context, this *model.Container) error
	GetChildrenContainers(ctx context.Context, this *model.Container) ([]model.Container, error)
	GetParentContainers(ctx context.Context, this *model.Container) ([]model.Container, error)
	GetParentContainer(ctx context.Context, this *model.Container) (*model.Container, error)
	GetContainerProducts(ctx context.Context, this *model.Container) ([]model.RemainingProduct, error)
}

type pgRepository struct {
	conn *sql.DB
}

func NewWhsRepository(conn *sql.DB) *pgRepository {
	return &pgRepository{
		conn: conn,
	}
}

func (r *pgRepository) GetProductById(ctx context.Context, id int) (*model.Product, error) {
	return nil, nil
}

func (r *pgRepository) GetProductByBarcode(ctx context.Context, barcode string) (*model.Product, error) {
	return &model.Product{Id: 123, Name: "Test"}, nil
}
