package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/mlplabs/lynway-wms/internal/model"
	"github.com/mlplabs/lynway-wms/internal/repository/dbo"
)

var (
	ErrEmptyName            = errors.New("name cannot be empty")
	ErrNotFound             = errors.New("entity not found")
	ErrBarcodeAlreadyExists = errors.New("barcode already assigned to another item")
)

// SearchManufacturers - поиск производителя
func (r *pgRepository) SearchManufacturers(ctx context.Context, query string, limit int) ([]model.Manufacturer, error) {
	var rows *sql.Rows

	var err error

	if query == "" {
		rows, err = r.conn.QueryContext(ctx, `
			SELECT id, name FROM public.manufacturers 
			ORDER BY id DESC LIMIT $1`, limit)
	} else {
		rows, err = r.conn.QueryContext(ctx, `
			SELECT id, name FROM public.manufacturers 
			WHERE name % $1 OR name ILIKE '%' || $1 || '%'
			
			ORDER BY word_similarity($1, name) DESC, name ASC 
			LIMIT $2`, query, limit)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	res := make([]model.Manufacturer, 0)

	for rows.Next() {
		var m model.Manufacturer
		if err := rows.Scan(&m.ID, &m.Name); err != nil {
			return nil, err
		}

		res = append(res, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *pgRepository) GetManufacturerById(ctx context.Context, baseID int) (*model.Manufacturer, error) {
	var res model.Manufacturer
	err := r.conn.QueryRowContext(ctx, `SELECT id, name FROM public.manufacturers WHERE id = $1`, baseID).Scan(&res.ID, &res.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *pgRepository) FindManufacturerByName(ctx context.Context, name string) (*model.Manufacturer, error) {
	var res model.Manufacturer

	err := r.conn.QueryRowContext(ctx, `
		SELECT id, name FROM public.manufacturers 
		WHERE LOWER(TRIM(name)) = LOWER($1) 
		LIMIT 1`, name).Scan(&res.ID, &res.Name)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *pgRepository) CreateManufacturer(ctx context.Context, name string) (*model.Manufacturer, error) {
	var res model.Manufacturer

	err := r.conn.QueryRowContext(ctx, `
		INSERT INTO public.manufacturers (name) VALUES ($1) 
		RETURNING id, name`, name).Scan(&res.ID, &res.Name)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// SearchItems - поиск номенклатуры
func (r *pgRepository) SearchItems(ctx context.Context, query string, limit int) ([]model.Item, error) {
	var rows *sql.Rows

	var err error

	if query == "" {
		rows, err = r.conn.QueryContext(ctx, `
			SELECT i.id, i.name, i.manufacturer_id, m.name, i.sku, i.barcode, i.tags, i.attributes, i.created_at, i.updated_at
			FROM public.items i
			LEFT JOIN public.manufacturers m ON m.id = i.manufacturer_id
			ORDER BY i.updated_at DESC LIMIT $1`, limit)
	} else {
		rows, err = r.conn.QueryContext(ctx, `
			SELECT i.id, i.name, i.manufacturer_id, m.name, i.sku, i.barcode, i.tags, i.attributes, i.created_at, i.updated_at
			FROM public.items i
			LEFT JOIN public.manufacturers m ON m.id = i.manufacturer_id
			WHERE i.name % $1 
			   OR i.name ILIKE '%' || $1 || '%'
			   OR i.sku ILIKE $1 || '%'
			   OR LOWER(TRIM(i.barcode)) = LOWER($1)
			   OR EXISTS (SELECT 1 FROM unnest(i.tags) t WHERE t ILIKE '%' || $1 || '%')
			ORDER BY 
				CASE WHEN LOWER(TRIM(i.barcode)) = LOWER($1) THEN 0 WHEN i.sku ILIKE $1 THEN 1 ELSE 2 END,
				word_similarity($1, i.name) DESC,
				i.id DESC
			LIMIT $2`, query, limit)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]model.Item, 0)

	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, it)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *pgRepository) GetItemById(ctx context.Context, baseID int) (*model.Item, error) {
	row := r.conn.QueryRowContext(ctx, `
		SELECT i.id, i.name, i.manufacturer_id, m.name, i.sku, i.barcode, i.tags, i.attributes, i.created_at, i.updated_at
		FROM public.items i
		LEFT JOIN public.manufacturers m ON m.id = i.manufacturer_id
		WHERE i.id = $1`, baseID)

	it, err := scanItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return &model.Item{}, ErrNotFound
	}
	return &it, err
}

func (r *pgRepository) FindExactItem(ctx context.Context, name string, mfgID *int, sku, barcode *string) (*model.Item, error) {
	row := r.conn.QueryRowContext(ctx, `
		SELECT i.id, i.name, i.manufacturer_id, m.name, i.sku, i.barcode, i.tags, i.attributes, i.created_at, i.updated_at
		FROM public.items i
		LEFT JOIN public.manufacturers m ON m.id = i.manufacturer_id
		WHERE LOWER(TRIM(i.name)) = LOWER($1)
		  AND i.manufacturer_id IS NOT DISTINCT FROM $2
		  AND LOWER(COALESCE(i.sku, '')) = LOWER(COALESCE($3, ''))
		  AND LOWER(COALESCE(i.barcode, '')) = LOWER(COALESCE($4, ''))
		LIMIT 1`, name, mfgID, sku, barcode)

	it, err := scanItem(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &model.Item{}, ErrNotFound
		}

		return nil, err
	}

	return &it, err
}

func (r *pgRepository) CreateItem(ctx context.Context, p *dbo.CreateItemParams) (int, error) {
	var newID int

	err := r.conn.QueryRowContext(ctx, `
		INSERT INTO public.items (name, manufacturer_id, sku, barcode, tags, attributes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		p.Name, p.ManufacturerID, p.SKU, p.Barcode, pq.Array(p.Tags), p.Attributes,
	).Scan(&newID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return 0, ErrBarcodeAlreadyExists
		}

		return 0, fmt.Errorf("insert item: %w", err)
	}

	return newID, nil
}

func (r *pgRepository) GetItemStockLocation(ctx context.Context, id int) ([]model.ItemStockLocation, error) {
	res := make([]model.ItemStockLocation, 0)

	rows, err := r.conn.QueryContext(ctx, `
		SELECT 
			s.zone_id, z.name, z.type,
			s.cell_id, c.name,
			s.container_id, cnt.name,
			SUM(s.quantity) AS qty
		FROM public.storage s
		JOIN public.zones z ON z.id = s.zone_id
		LEFT JOIN public.cells c ON c.id = s.cell_id
		LEFT JOIN public.containers cnt ON cnt.id = s.container_id
		WHERE s.item_id = $1
		GROUP BY s.zone_id, z.name, z.type, s.cell_id, c.name, s.container_id, cnt.name
		HAVING SUM(s.quantity) > 0
		ORDER BY s.zone_id ASC, s.cell_id NULLS LAST, s.container_id NULLS LAST`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var loc model.ItemStockLocation
		if err := rows.Scan(
			&loc.ZoneID, &loc.ZoneName, &loc.ZoneType,
			&loc.CellID, &loc.CellName,
			&loc.ContainerID, &loc.ContainerName,
			&loc.Quantity,
		); err != nil {
			return nil, err
		}

		res = append(res, loc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *pgRepository) UpdateItem(ctx context.Context, p dbo.UpdateItemParams) error {
	res, err := r.dbTx.ExecContext(ctx, `
		UPDATE public.items
		SET name            = $1,
		    manufacturer_id = $2,
		    sku             = $3,
		    barcode         = $4,
		    tags            = $5,
		    attributes      = $6,
		    updated_at      = now()
		WHERE id = $7`,
		p.Name, p.ManufacturerID, p.SKU, p.Barcode, pq.Array(p.Tags), p.Attributes, p.ID,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrBarcodeAlreadyExists
		}
		return fmt.Errorf("update item: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanItem(s rowScanner) (model.Item, error) {
	var it model.Item

	var tags pq.StringArray

	err := s.Scan(
		&it.ID, &it.Name, &it.ManufacturerID, &it.ManufacturerName,
		&it.SKU, &it.Barcode, &tags, &it.Attributes, &it.CreatedAt, &it.UpdatedAt,
	)
	if err != nil {
		return model.Item{}, err
	}

	it.Tags = tags
	if it.Tags == nil {
		it.Tags = []string{}
	}

	if len(it.Attributes) == 0 {
		it.Attributes = json.RawMessage(`{}`)
	}
	return it, nil
}
