package model

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

// Item - номенклатура
type Item struct {
	ID               int
	Name             string
	ManufacturerID   *int
	ManufacturerName *string
	SKU              *string
	Barcode          *string
	Tags             []string
	Attributes       json.RawMessage
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Manufacturer - производитель
type Manufacturer struct {
	ID   int
	Name string
}

// ResolveManufacturerRequest - FCC структура запроса создания/выбора производителя
type ResolveManufacturerRequest struct {
	BaseID *int   `json:"base_id"`
	Name   string `json:"name"`
}

// ResolveManufacturerResponse - структура ответа создания/выбора производителя
type ResolveManufacturerResponse struct {
	Manufacturer *Manufacturer `json:"manufacturer"`
	Action       ResolveAction `json:"action"`
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
	Item   *Item         `json:"item"`
	Action ResolveAction `json:"action"`
}

type UpdateItemRequest struct {
	Name           string
	ManufacturerID *int
	SKU            *string
	Barcode        *string
	Tags           []string
	Attributes     json.RawMessage
}

type UpdateManufacturerRequest struct {
	Name string
}

type ItemStockLocation struct {
	ZoneID        int     `json:"zone_id"`
	ZoneName      string  `json:"zone_name"`
	ZoneType      string  `json:"zone_type"`
	CellID        *int    `json:"cell_id,omitempty"`
	CellName      *string `json:"cell_name,omitempty"`
	ContainerID   *int    `json:"container_id,omitempty"`
	ContainerName *string `json:"container_name,omitempty"`
	Quantity      int     `json:"quantity"`
}

type ItemDetails struct {
	Item
	TotalAvailable int                 `json:"total_available"`
	Locations      []ItemStockLocation `json:"locations"`
}
