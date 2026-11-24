package dto

import (
	"encoding/json"
	"time"
)

type ResolveAction string

const (
	ActionFound   ResolveAction = "FOUND"
	ActionCloned  ResolveAction = "CLONED"
	ActionCreated ResolveAction = "CREATED"
)

type ItemResponse struct {
	ID               int             `json:"id"`
	Name             string          `json:"name"`
	ManufacturerID   *int            `json:"manufacturer_id,omitempty"`
	ManufacturerName *string         `json:"manufacturer_name,omitempty"`
	SKU              *string         `json:"sku,omitempty"`
	Barcode          *string         `json:"barcode,omitempty"`
	Tags             []string        `json:"tags"`
	Attributes       json.RawMessage `json:"attributes"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type ItemSuggest struct {
	ID               int             `json:"id"`
	Name             string          `json:"name"`
	ManufacturerID   *int            `json:"manufacturer_id,omitempty"`
	ManufacturerName *string         `json:"manufacturer_name,omitempty"`
	SKU              *string         `json:"sku,omitempty"`
	Barcode          *string         `json:"barcode,omitempty"`
	Tags             []string        `json:"tags"`
	Attributes       json.RawMessage `json:"attributes"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type ItemsSuggestResponse []ItemSuggest

type ManufacturerResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ManufacturerSuggest struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ManufacturersSuggestResponse []ManufacturerSuggest

type UpdateItemRequest struct {
	Name           string          `json:"name"`
	ManufacturerID *int            `json:"manufacturer_id"`
	SKU            *string         `json:"sku"`
	Barcode        *string         `json:"barcode"`
	Tags           []string        `json:"tags"`
	Attributes     json.RawMessage `json:"attributes"`
}

type UpdateManufacturerRequest struct {
	Name string `json:"name"`
}

type ResolveItemRequest struct {
	BaseID         *int            `json:"base_id"`
	Name           string          `json:"name"`
	ManufacturerID *int            `json:"manufacturer_id"`
	SKU            *string         `json:"sku"`
	Barcode        *string         `json:"barcode"`
	Tags           []string        `json:"tags"`
	Attributes     json.RawMessage `json:"attributes"`
}

type ResolveItemResponse struct {
	Item   *ItemResponse `json:"item"`
	Action ResolveAction `json:"action"`
}

// ResolveManufacturerRequest - FCC структура запроса создания/выбора производителя
type ResolveManufacturerRequest struct {
	BaseID *int   `json:"base_id"`
	Name   string `json:"name"`
}

// ResolveManufacturerResponse - структура ответа создания/выбора производителя
type ResolveManufacturerResponse struct {
	Manufacturer *ManufacturerResponse `json:"manufacturer"`
	Action       ResolveAction         `json:"action"`
}
