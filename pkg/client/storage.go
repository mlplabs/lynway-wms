package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/mlplabs/lynway-wms/pkg/dto"
)

const idempotencyKeyParam = "Idempotency-Key"

func (c *HttpClient) postStorageOp(ctx context.Context, path, idempotencyKey string, req any) (*dto.StorageOperationResponse, error) {
	result := struct {
		Data dto.StorageOperationResponse `json:"data"`
	}{}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader(idempotencyKeyParam, idempotencyKey).
		SetBody(req).
		SetResult(&result).
		Post(path)
	if err != nil {
		return nil, fmt.Errorf("storage op (%s) error: %w", path, err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("storage op (%s) error: %w", path, errors.New(resp.Status()))
	}

	return &result.Data, nil
}

// ReceiveItem выполняет приемку товара в зону, ячейку или контейнер.
func (c *HttpClient) ReceiveItem(ctx context.Context, idempotencyKey string, req *dto.ReceiveItemRequest) (*dto.StorageOperationResponse, error) {
	return c.postStorageOp(ctx, "/api/v1/storage/items/receive", idempotencyKey, req)
}

// MoveItem перемещает товар между зонами, ячейками или контейнерами.
func (c *HttpClient) MoveItem(ctx context.Context, idempotencyKey string, req *dto.MoveItemRequest) (*dto.StorageOperationResponse, error) {
	return c.postStorageOp(ctx, "/api/v1/storage/items/move", idempotencyKey, req)
}

// DispatchItem списывает товар в системную зону DISPATCH.
func (c *HttpClient) DispatchItem(ctx context.Context, idempotencyKey string, req *dto.DispatchItemRequest) (*dto.StorageOperationResponse, error) {
	return c.postStorageOp(ctx, "/api/v1/storage/items/dispatch", idempotencyKey, req)
}

// PlaceContainer выполняет первичное размещение внешнего контейнера в зону или ячейку.
func (c *HttpClient) PlaceContainer(ctx context.Context, idempotencyKey string, req *dto.PlaceContainerRequest) (*dto.StorageOperationResponse, error) {
	return c.postStorageOp(ctx, "/api/v1/storage/containers/place", idempotencyKey, req)
}

// MoveContainer перемещает внешний контейнер вместе со всем его вложенным деревом.
func (c *HttpClient) MoveContainer(ctx context.Context, idempotencyKey string, req *dto.MoveContainerRequest) (*dto.StorageOperationResponse, error) {
	return c.postStorageOp(ctx, "/api/v1/storage/containers/move", idempotencyKey, req)
}

// DispatchContainer списывает (отгружает) внешний контейнер в системную зону DISPATCH.
func (c *HttpClient) DispatchContainer(ctx context.Context, idempotencyKey string, req *dto.DispatchContainerRequest) (*dto.StorageOperationResponse, error) {
	return c.postStorageOp(ctx, "/api/v1/storage/containers/dispatch", idempotencyKey, req)
}

// NestContainer вкладывает дочерний контейнер (ChildID) внутрь родительского (TargetParentID).
func (c *HttpClient) NestContainer(ctx context.Context, idempotencyKey string, req *dto.NestContainerRequest) (*dto.StorageOperationResponse, error) {
	return c.postStorageOp(ctx, "/api/v1/storage/containers/nest", idempotencyKey, req)
}

// UnnestContainer извлекает вложенный контейнер наружу на целевые координаты.
func (c *HttpClient) UnnestContainer(ctx context.Context, idempotencyKey string, req *dto.UnnestContainerRequest) (*dto.StorageOperationResponse, error) {
	return c.postStorageOp(ctx, "/api/v1/storage/containers/unnest", idempotencyKey, req)
}
