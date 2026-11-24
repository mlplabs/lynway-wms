package v1

import (
	"net/http"

	"github.com/mlplabs/lynway-wms/internal/model"
	"github.com/mlplabs/lynway-wms/pkg/dto"
)

// itemsSuggest
// @Tags Catalog
// @Summary поиск номенклатуры для автокомплита
// @Description точное совпадение по `barcode`/`sku` или нечеткое по `name` и `tags` через `pg_trgm`)
// @Router /items/suggest [get]
// @Param query q string true "поисковая строка"
// @Param query l int false "количество результатов, по-умолчанию 10"
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

	return itemsToDTO(res), nil
}

// manufacturersSuggest
// @Tags Catalog
// @Summary поиск производителя для автокомплита
// @Description совпадение по `name`)
// @Router /manufacturers/suggest [get]
// @Param query q string true "поисковая строка"
// @Param query l int false "количество результатов, по-умолчанию 10"
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

	return manufacturersToDTO(res), nil
}

func itemsToDTO(in []model.Item) *dto.ItemsSuggestResponse {
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

func manufacturersToDTO(in []model.Manufacturer) *dto.ManufacturersSuggestResponse {
	res := dto.ManufacturersSuggestResponse{}
	for _, item := range in {
		res = append(res, dto.ManufacturerSuggest{
			ID:   item.ID,
			Name: item.Name,
		})
	}

	return &res
}
