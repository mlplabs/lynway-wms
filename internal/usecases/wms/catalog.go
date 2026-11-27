package wms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mlplabs/lynway-wms/internal/model"
	"github.com/mlplabs/lynway-wms/internal/repository"
	"github.com/mlplabs/lynway-wms/internal/repository/dbo"
)

var (
	ErrEmptyName            = errors.New("name cannot be empty")
	ErrBarcodeAlreadyExists = errors.New("barcode already assigned to another item")
)

// SearchManufacturers — автокомплит производителей через pg_trgm
func (s *Service) SearchManufacturers(ctx context.Context, query string, limit int) ([]model.Manufacturer, error) {
	q := strings.TrimSpace(query)

	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.whs.SearchManufacturers(ctx, q, limit)
}

func (s *Service) GetManufacturerById(ctx context.Context, id int) (*model.Manufacturer, error) {
	return s.whs.GetManufacturerById(ctx, id)
}

// ResolveManufacturer — Find, Clone or Create для производителей (точное совпадение по имени без учета регистра)
func (s *Service) ResolveManufacturer(ctx context.Context, in *model.ResolveManufacturerRequest) (*model.ResolveManufacturerResponse, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, ErrEmptyName
	}

	if in.BaseID != nil && *in.BaseID > 0 {
		orig, err := s.whs.GetManufacturerById(ctx, *in.BaseID)
		if err != nil {
			return nil, err
		}

		if strings.EqualFold(strings.TrimSpace(orig.Name), name) {
			return &model.ResolveManufacturerResponse{Manufacturer: orig, Action: model.ActionFound}, nil
		}
	}

	existing, err := s.whs.FindManufacturerByName(ctx, name)
	if err == nil {
		action := model.ActionFound
		if in.BaseID != nil && *in.BaseID > 0 && existing.ID != *in.BaseID {
			action = model.ActionCloned
		}
		return &model.ResolveManufacturerResponse{Manufacturer: existing, Action: action}, nil
	}

	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	created, err := s.whs.CreateManufacturer(ctx, &dbo.CreateManufacturerParams{
		Name:      name,
		CreatedBy: model.UserIdPtrFromContext(ctx),
	})

	action := model.ActionCreated
	if in.BaseID != nil && *in.BaseID > 0 {
		action = model.ActionCloned
	}
	return &model.ResolveManufacturerResponse{Manufacturer: created, Action: action}, nil
}

// UpdateManufacturer — прямое редактирование карточки производителя (исправление опечаток без клонирования)
func (s *Service) UpdateManufacturer(ctx context.Context, id int, in *model.UpdateManufacturerRequest) (*model.Manufacturer, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, ErrEmptyName
	}

	err := s.whs.UpdateManufacturer(ctx, &dbo.UpdateManufacturerParams{
		ID:        id,
		Name:      name,
		UpdatedBy: model.UserIdPtrFromContext(ctx),
	})

	if err != nil {
		return nil, err
	}

	updated, err := s.whs.GetManufacturerById(ctx, id)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// SearchItems — поиск номенклатуры по barcode, sku, name (pg_trgm) или tags
func (s *Service) SearchItems(ctx context.Context, query string, limit int) ([]model.Item, error) {
	q := strings.TrimSpace(query)

	if limit <= 0 || limit > 50 {
		limit = 20
	}

	res, err := s.whs.SearchItems(ctx, q, limit)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) GetItemById(ctx context.Context, id int) (*model.Item, error) {
	return s.whs.GetItemById(ctx, id)
}

// ResolveItem — ядро килер-фичи «Find, Clone or Create» (сравнение по name, manufacturer_id, sku, barcode)
func (s *Service) ResolveItem(ctx context.Context, in *model.ResolveItemRequest) (*model.ResolveItemResponse, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, ErrEmptyName
	}

	sku := normalizeNullStr(in.SKU)
	barcode := normalizeNullStr(in.Barcode)

	// 1. Если выбран базовый товар из списка — проверяем, изменились ли ключевые поля
	var orig *model.Item

	if in.BaseID != nil && *in.BaseID > 0 {
		loaded, err := s.whs.GetItemById(ctx, *in.BaseID)
		if err != nil {
			return nil, err
		}

		orig = loaded

		if isSameKeyFields(orig, name, in.ManufacturerID, sku, barcode) {
			return &model.ResolveItemResponse{Item: orig, Action: model.ActionFound}, nil
		}
	}

	// 2. Если ввод с нуля — проверяем, нет ли уже полного дубля в базе
	if orig == nil {
		existing, err := s.whs.FindExactItem(ctx, name, in.ManufacturerID, sku, barcode)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}

		if err == nil {
			return &model.ResolveItemResponse{Item: existing, Action: model.ActionFound}, nil
		}
	}

	// 3. Готовим теги и атрибуты (с наследованием от оригинала при клонировании)
	finalTags := mergeTags(orig, in.Tags)

	finalAttrs, err := mergeAttributes(orig, in.Attributes)
	if err != nil {
		return nil, err
	}

	// 4. Транзакция открывается только на создание/клонирование
	var created *model.Item

	err = s.whs.WithinTx(ctx, func(txRepo repository.PgRepository) error {
		newID, err := txRepo.CreateItem(ctx, &dbo.CreateItemParams{
			Name:           name,
			ManufacturerID: in.ManufacturerID,
			SKU:            sku,
			Barcode:        barcode,
			Tags:           finalTags,
			Attributes:     finalAttrs,
			CreatedBy:      model.UserIdPtrFromContext(ctx),
		})
		if err != nil {
			return err
		}

		// В будущем здесь же в рамках одной транзакции копируем вложения (attachments) при orig != nil

		created, err = txRepo.GetItemById(ctx, newID)

		return err
	})
	if err != nil {
		return nil, err
	}

	action := model.ActionCreated
	if orig != nil {
		action = model.ActionCloned
	}
	return &model.ResolveItemResponse{Item: created, Action: action}, nil
}

// GetItemWithStock — детальная карточка товара + агрегированные остатки из storage
func (s *Service) GetItemWithStock(ctx context.Context, id int) (*model.ItemDetails, error) {
	it, err := s.whs.GetItemById(ctx, id)
	if err != nil {
		return nil, err
	}

	locations, err := s.whs.GetItemStockLocation(ctx, id)
	if err != nil {
		return nil, err
	}

	var totalAvailable int

	for _, loc := range locations {
		if loc.ZoneType != "DISPATCH" {
			totalAvailable += loc.Quantity
		}
	}

	return &model.ItemDetails{
		Item:           *it,
		TotalAvailable: totalAvailable,
		Locations:      locations,
	}, nil
}

// UpdateItem — прямое редактирование карточки товара (исправление опечаток без клонирования)
func (s *Service) UpdateItem(ctx context.Context, id int, in *model.UpdateItemRequest) (*model.Item, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, ErrEmptyName
	}

	attrs := in.Attributes
	if len(attrs) == 0 || string(attrs) == "null" {
		attrs = json.RawMessage(`{}`)
	}

	tags := in.Tags
	if tags == nil {
		tags = []string{}
	}

	err := s.whs.UpdateItem(ctx, dbo.UpdateItemParams{
		ID:             id,
		Name:           name,
		ManufacturerID: in.ManufacturerID,
		SKU:            normalizeNullStr(in.SKU),
		Barcode:        normalizeNullStr(in.Barcode),
		Tags:           tags,
		Attributes:     attrs,
		UpdatedBy:      model.UserIdPtrFromContext(ctx),
	})

	if err != nil {
		return nil, err
	}

	updated, err := s.whs.GetItemById(ctx, id)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// --- Helpers ---

func normalizeNullStr(s *string) *string {
	if s == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func equalStrPtr(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}
	return strings.EqualFold(*a, *b)
}

func equalInt64Ptr(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func isSameKeyFields(orig *model.Item, name string, mfgID *int, sku, barcode *string) bool {
	return strings.EqualFold(strings.TrimSpace(orig.Name), name) &&
		equalInt64Ptr(orig.ManufacturerID, mfgID) &&
		equalStrPtr(orig.SKU, sku) &&
		equalStrPtr(orig.Barcode, barcode)
}

func mergeTags(orig *model.Item, incoming []string) []string {
	seen := make(map[string]struct{})
	res := make([]string, 0)

	add := func(list []string) {
		for _, t := range list {
			clean := strings.TrimSpace(t)
			if clean == "" {
				continue
			}

			key := strings.ToLower(clean)
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}

				res = append(res, clean)
			}
		}
	}

	if orig != nil {
		add(orig.Tags)
	}

	add(incoming)

	return res
}

func mergeAttributes(orig *model.Item, incoming json.RawMessage) ([]byte, error) {
	base := make(map[string]any)
	if orig != nil && len(orig.Attributes) > 0 {
		_ = json.Unmarshal(orig.Attributes, &base)
	}

	if len(incoming) > 0 && string(incoming) != "null" {
		var patch map[string]any
		if err := json.Unmarshal(incoming, &patch); err != nil {
			return nil, fmt.Errorf("invalid attributes json: %w", err)
		}

		for k, v := range patch {
			if v == nil {
				delete(base, k)
			} else {
				base[k] = v
			}
		}
	}
	return json.Marshal(base)
}
