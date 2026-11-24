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
