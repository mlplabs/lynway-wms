package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/mlplabs/lynway-wms/pkg/dto"
)

// ItemsSuggest поиск номенклатуры для автокомплита
func (c *HttpClient) ItemsSuggest(ctx context.Context, text string, limit int) (*dto.ItemsSuggestResponse, error) {
	result := struct {
		Data dto.ItemsSuggestResponse `json:"data"`
	}{}
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get(fmt.Sprintf("/api/v1/items/suggest?q=%s&l=%d", text, limit))

	if err != nil {
		return nil, fmt.Errorf("items suggest error: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("items suggest error: %w", errors.New(resp.Status()))
	}

	return &result.Data, nil
}

// ManufacturersSuggest поиск производителей для автокомплита
func (c *HttpClient) ManufacturersSuggest(ctx context.Context, text string, limit int) (*dto.ManufacturersSuggestResponse, error) {
	result := struct {
		Data dto.ManufacturersSuggestResponse `json:"data"`
	}{}
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get(fmt.Sprintf("/api/v1/manufacturers/suggest?q=%s&l=%d", text, limit))

	if err != nil {
		return nil, fmt.Errorf("manufacturers suggest error: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("manufacturers suggest error: %w", errors.New(resp.Status()))
	}

	return &result.Data, nil
}

// GetItemById возвращает объект номенклатуры
func (c *HttpClient) GetItemById(ctx context.Context, id int) (*dto.ItemResponse, error) {
	result := struct {
		Data dto.ItemResponse `json:"data"`
	}{}
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get(fmt.Sprintf("/api/v1/items/%d", id))

	if err != nil {
		return nil, fmt.Errorf("item by id error: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("item by id error: %w", errors.New(resp.Status()))
	}

	return &result.Data, nil
}

// GetManufacturerById возвращает объект производителя
func (c *HttpClient) GetManufacturerById(ctx context.Context, id int) (*dto.ManufacturerResponse, error) {
	result := struct {
		Data dto.ManufacturerResponse `json:"data"`
	}{}
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get(fmt.Sprintf("/api/v1/manufacturers/%d", id))

	if err != nil {
		return nil, fmt.Errorf("manufacturer by id error: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("manufacturer by id error: %w", errors.New(resp.Status()))
	}

	return &result.Data, nil
}

// UpdateItem обновляет и возвращает объект номенклатуры
func (c *HttpClient) UpdateItem(ctx context.Context, id int, req *dto.UpdateItemRequest) (*dto.ItemResponse, error) {
	result := struct {
		Data dto.ItemResponse `json:"data"`
	}{}
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Patch(fmt.Sprintf("/api/v1/items/%d", id))

	if err != nil {
		return nil, fmt.Errorf("item by id error: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("item by id error: %w", errors.New(resp.Status()))
	}

	return &result.Data, nil
}

// UpdateManufacturer обновляет и возвращает объект производителя
func (c *HttpClient) UpdateManufacturer(ctx context.Context, id int, req *dto.UpdateManufacturerRequest) (*dto.ManufacturerResponse, error) {
	result := struct {
		Data dto.ManufacturerResponse `json:"data"`
	}{}
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Patch(fmt.Sprintf("/api/v1/manufacturers/%d", id))

	if err != nil {
		return nil, fmt.Errorf("manufacturer by id error: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("manufacturer by id error: %w", errors.New(resp.Status()))
	}

	return &result.Data, nil
}

// ResolveItem  обновляет или создает новый элемент номенклатуры
func (c *HttpClient) ResolveItem(ctx context.Context, req *dto.ResolveItemRequest) (*dto.ResolveItemResponse, error) {
	result := struct {
		Data dto.ResolveItemResponse `json:"data"`
	}{}
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Post("/api/v1/items/resolve")

	if err != nil {
		return nil, fmt.Errorf("resolve item error: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("resolve item error: %w", errors.New(resp.Status()))
	}

	return &result.Data, nil
}

// ResolveManufacturer обновляет или создает новый элемент производителя
func (c *HttpClient) ResolveManufacturer(ctx context.Context, req *dto.ResolveManufacturerRequest) (*dto.ResolveManufacturerResponse, error) {
	result := struct {
		Data dto.ResolveManufacturerResponse `json:"data"`
	}{}
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Post("/api/v1/manufacturers/resolve")

	if err != nil {
		return nil, fmt.Errorf("resolve manufacturer error: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("resolve manufacturer error: %w", errors.New(resp.Status()))
	}

	return &result.Data, nil
}
