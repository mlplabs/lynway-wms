package dto

import (
	"encoding/json"
	"time"
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
