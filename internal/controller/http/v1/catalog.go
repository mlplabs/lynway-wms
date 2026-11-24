package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	domainErr "github.com/mlplabs/common-go-pkg/pkg/http/errors/custom"
	"github.com/mlplabs/lynway-wms/internal/model"
	"github.com/mlplabs/lynway-wms/pkg/dto"
)

// itemsSuggest
// @Tags Catalog
// @Summary поиск номенклатуры для автокомплита
// @Description точное совпадение по `barcode`/`sku` или нечеткое по `name` и `tags` через `pg_trgm`)
// @Router /items/suggest [get]
// @Param q query string true "поисковая строка"
// @Param l query int false "количество результатов, по-умолчанию 10"
// @Success	200 {object} wrapper.Data{data=dto.ItemsSuggestResponse} "список номенклатуры"
func (h *ServiceHandler) itemsSuggest(w http.ResponseWriter, r *http.Request) (any, error) {
	search := getQueryParam(r)
	limit, err := getLimitSuggestParam(r)
	if err != nil {
		return nil, err
	}
	res, err := h.service.SearchItems(r.Context(), search, limit)
	if err != nil {
		return nil, err
	}

	return itemsSuggestToDTO(res), nil
}

// manufacturersSuggest
// @Tags Catalog
// @Summary поиск производителя для автокомплита
// @Description совпадение по `name`)
// @Router /manufacturers/suggest [get]
// @Param q query string true "поисковая строка"
// @Param l query int false "количество результатов, по-умолчанию 10"
// @Success	200 {object} wrapper.Data{data=dto.ManufacturersSuggestResponse} "список производителей"
func (h *ServiceHandler) manufacturersSuggest(w http.ResponseWriter, r *http.Request) (any, error) {
	search := getQueryParam(r)
	limit, err := getLimitSuggestParam(r)
	if err != nil {
		return nil, err
	}
	res, err := h.service.SearchManufacturers(r.Context(), search, limit)
	if err != nil {
		return nil, err
	}

	return manufacturersSuggestToDTO(res), nil
}

// getItemById
// @Tags Catalog
// @Summary возвращает элемент номенклатуры
// @Router /items/{id} [get]
// @Param id path int true "идентификатор номенклатуры"
// @Success	200 {object} wrapper.Data{data=dto.ItemResponse} "элемент номенклатуры"
func (h *ServiceHandler) getItemById(w http.ResponseWriter, r *http.Request) (any, error) {
	id, err := getIdParam(r)
	if err != nil {
		return nil, err
	}
	item, err := h.service.GetItemById(r.Context(), id)
	if err != nil {
		return nil, err
	}

	return itemToDTO(item), nil
}

// getManufacturerById
// @Tags Catalog
// @Summary возвращает элемент производителя
// @Router /manufacturers/{id} [get]
// @Param id path int true "идентификатор производителя"
// @Success	200 {object} wrapper.Data{data=dto.ManufacturerResponse} "элемент производителя"
func (h *ServiceHandler) getManufacturerById(w http.ResponseWriter, r *http.Request) (any, error) {
	id, err := getIdParam(r)
	if err != nil {
		return nil, err
	}
	item, err := h.service.GetManufacturerById(r.Context(), id)
	if err != nil {
		return nil, err
	}

	return manufacturerToDTO(item), nil
}

// updateItem
// @Tags Catalog
// @Summary обновляет элемент номенклатуры
// @Router /items/{id} [patch]
// @Param id path int true "идентификатор номенклатуры"
// @Param request body dto.UpdateItemRequest true "данные номенклатуры"
// @Success	200 {object} wrapper.Data{data=dto.ItemResponse} "элемент номенклатуры"
func (h *ServiceHandler) updateItem(w http.ResponseWriter, r *http.Request) (any, error) {
	id, err := getIdParam(r)
	if err != nil {
		return nil, err
	}

	req, err := getUpdateItemRequest(r)
	if err != nil {
		return nil, err
	}

	res, err := h.service.UpdateItem(r.Context(), id, req)
	if err != nil {
		return nil, err
	}

	return itemToDTO(res), nil
}

// updateManufacturer
// @Tags Catalog
// @Summary обновляет элемент производителя
// @Router /manufacturers/{id} [patch]
// @Param id path int true "идентификатор производителя"
// @Param request body dto.UpdateManufacturerRequest true "данные производителя"
// @Success	200 {object} wrapper.Data{data=dto.ManufacturerResponse} "элемент производителя"
func (h *ServiceHandler) updateManufacturer(w http.ResponseWriter, r *http.Request) (any, error) {
	id, err := getIdParam(r)
	if err != nil {
		return nil, err
	}

	req, err := getUpdateManufacturerRequest(r)
	if err != nil {
		return nil, err
	}

	res, err := h.service.UpdateManufacturer(r.Context(), id, req)
	if err != nil {
		return nil, err
	}

	return manufacturerToDTO(res), nil
}

// itemsResolve
// @Tags Catalog
// @Summary обновляет или создает новый элемент номенклатуры
// @Description **«FCC»**:
// @Description   - Передан `base_id` + ключевые поля (`name`, `sku`, `barcode`) **изменены** -> создает новый `Item`, наследуя `image_path`, `tags`, `attributes` из оригинала.
// @Description   - Передан `base_id` + ключевые поля **не изменены** -> возвращает оригинал.
// @Router /items/resolve [post]
// @Param request body dto.ResolveItemRequest true "данные производителя"
// @Success	200 {object} wrapper.Data{data=dto.ResolveItemResponse} "элемент производителя"
func (h *ServiceHandler) itemsResolve(w http.ResponseWriter, r *http.Request) (any, error) {
	req, err := getResolveItemRequest(r)
	if err != nil {
		return nil, err
	}

	res, err := h.service.ResolveItem(r.Context(), req)
	if err != nil {
		return nil, err
	}

	return &dto.ResolveItemResponse{
		Item:   itemToDTO(res.Item),
		Action: dto.ResolveAction(res.Action),
	}, nil
}

// manufacturersResolve
// @Tags Catalog
// @Summary обновляет или создает новый элемент производителя
// @Description **«FCC»**:
// @Description   - Передан `base_id` + ключевые поля (`name`) **изменены** -> создает новый `Manufacturer`.
// @Description   - Передан `base_id` + ключевые поля **не изменены** -> возвращает оригинал.
// @Router /manufacturers/resolve [post]
// @Param request body dto.ResolveManufacturerRequest true "данные производителя"
// @Success	200 {object} wrapper.Data{data=dto.ResolveManufacturerResponse} "элемент производителя"
func (h *ServiceHandler) manufacturersResolve(w http.ResponseWriter, r *http.Request) (any, error) {
	req, err := getResolveManufacturerRequest(r)
	if err != nil {
		return nil, err
	}
	
	res, err := h.service.ResolveManufacturer(r.Context(), req)
	if err != nil {
		return nil, err
	}

	return &dto.ResolveManufacturerResponse{
		Manufacturer: manufacturerToDTO(res.Manufacturer),
		Action:       dto.ResolveAction(res.Action),
	}, nil
}

func itemToDTO(in *model.Item) *dto.ItemResponse {
	if in == nil {
		return nil
	}

	return &dto.ItemResponse{
		ID:               in.ID,
		Name:             in.Name,
		ManufacturerID:   in.ManufacturerID,
		ManufacturerName: in.ManufacturerName,
		SKU:              in.SKU,
		Barcode:          in.Barcode,
		Tags:             in.Tags,
		Attributes:       in.Attributes,
		CreatedAt:        in.CreatedAt,
		UpdatedAt:        in.UpdatedAt,
	}
}

func itemsSuggestToDTO(in []model.Item) *dto.ItemsSuggestResponse {
	res := dto.ItemsSuggestResponse{}
	for _, item := range in {
		res = append(res, dto.ItemSuggest{
			ID:               item.ID,
			Name:             item.Name,
			ManufacturerID:   item.ManufacturerID,
			ManufacturerName: item.ManufacturerName,
			SKU:              item.SKU,
			Barcode:          item.Barcode,
			Tags:             item.Tags,
			Attributes:       item.Attributes,
			CreatedAt:        item.CreatedAt,
			UpdatedAt:        item.UpdatedAt,
		})
	}

	return &res
}

func manufacturerToDTO(in *model.Manufacturer) *dto.ManufacturerResponse {
	if in == nil {
		return nil
	}
	return &dto.ManufacturerResponse{
		ID:   in.ID,
		Name: in.Name,
	}
}

func manufacturersSuggestToDTO(in []model.Manufacturer) *dto.ManufacturersSuggestResponse {
	res := dto.ManufacturersSuggestResponse{}
	for _, item := range in {
		res = append(res, dto.ManufacturerSuggest{
			ID:   item.ID,
			Name: item.Name,
		})
	}

	return &res
}

func getUpdateItemRequest(r *http.Request) (*model.UpdateItemRequest, error) {
	req := dto.UpdateItemRequest{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body, %w", err))
	}
	return &model.UpdateItemRequest{
		Name:           req.Name,
		ManufacturerID: req.ManufacturerID,
		SKU:            req.SKU,
		Barcode:        req.Barcode,
		Tags:           req.Tags,
		Attributes:     req.Attributes,
	}, nil
}

func getUpdateManufacturerRequest(r *http.Request) (*model.UpdateManufacturerRequest, error) {
	req := dto.UpdateManufacturerRequest{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body, %w", err))
	}
	return &model.UpdateManufacturerRequest{
		Name: req.Name,
	}, nil
}

func getResolveItemRequest(r *http.Request) (*model.ResolveItemRequest, error) {
	req := dto.ResolveItemRequest{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body, %w", err))
	}
	return &model.ResolveItemRequest{
		BaseID:         req.BaseID,
		Name:           req.Name,
		ManufacturerID: req.ManufacturerID,
		SKU:            req.SKU,
		Barcode:        req.Barcode,
		Tags:           req.Tags,
		Attributes:     req.Attributes,
	}, nil
}

func getResolveManufacturerRequest(r *http.Request) (*model.ResolveManufacturerRequest, error) {
	req := dto.ResolveManufacturerRequest{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body, %w", err))
	}
	return &model.ResolveManufacturerRequest{
		BaseID: req.BaseID,
		Name:   req.Name,
	}, nil
}
