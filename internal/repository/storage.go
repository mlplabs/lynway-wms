package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/mlplabs/lynway-wms/internal/model"
)

func (r *pgRepository) GetZoneMeta(ctx context.Context, zoneID int) (model.ZoneMeta, error) {
	var m model.ZoneMeta
	err := r.conn.QueryRowContext(ctx, `
		SELECT id, warehouse_id, type, is_system
		FROM zones
		WHERE id = $1
	`, zoneID).Scan(&m.ID, &m.WhsID, &m.Type, &m.IsSystem)
	if err != nil {
		return m, fmt.Errorf("get zone %d: %w", zoneID, err)
	}
	return m, nil
}

func (r *pgRepository) GetCellMeta(ctx context.Context, cellID int) (model.CellMeta, error) {
	var m model.CellMeta
	err := r.conn.QueryRowContext(ctx, `
		SELECT id, whs_id, zone_id, not_allowed_in, not_allowed_out
		FROM cells
		WHERE id = $1
	`, cellID).Scan(&m.ID, &m.WhsID, &m.ZoneID, &m.NotAllowedIn, &m.NotAllowedOut)
	if err != nil {
		return m, fmt.Errorf("get cell %d: %w", cellID, err)
	}

	return m, nil
}

func (r *pgRepository) GetContainerMeta(ctx context.Context, containerID int) (model.ContainerMeta, error) {
	var m model.ContainerMeta
	err := r.conn.QueryRowContext(ctx, `
		SELECT id, parent_id, root_id, depth
		FROM containers
		WHERE id = $1
	`, containerID).Scan(&m.ID, &m.ParentID, &m.RootID, &m.Depth)
	if err != nil {
		return m, fmt.Errorf("get container %d: %w", containerID, err)
	}

	return m, nil
}

// GetBalance считает остаток по точному совпадению координат (включая NULL).
func (r *pgRepository) GetBalance(ctx context.Context, loc model.Location, itemID *int) (int, error) {
	query := `
		SELECT COALESCE(SUM(quantity), 0)
		FROM storage
		WHERE whs_id = $1
		  AND zone_id = $2
		  AND cell_id IS NOT DISTINCT FROM $3
		  AND container_id IS NOT DISTINCT FROM $4
		  AND item_id IS NOT DISTINCT FROM $5
	`
	var balance int
	err := r.conn.QueryRowContext(ctx, query, loc.WhsID, loc.ZoneID, loc.CellID, loc.ContainerID, itemID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("get balance: %w", err)
	}

	return balance, nil
}

// GetContainerCurrentLocation возвращает текущие координаты, где числится внешний контейнер (если остаток > 0).
func (r *pgRepository) GetContainerCurrentLocation(ctx context.Context, containerID int) (*model.Location, error) {
	query := `
		SELECT whs_id, zone_id, cell_id
		FROM storage
		WHERE container_id = $1 AND item_id IS NULL
		GROUP BY whs_id, zone_id, cell_id
		HAVING SUM(quantity) > 0
		LIMIT 1
	`
	var loc model.Location
	err := r.conn.QueryRowContext(ctx, query, containerID).Scan(&loc.WhsID, &loc.ZoneID, &loc.CellID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("get container location: %w", err)
	}

	return &loc, nil
}

// GetRootSubtreeContainerIDs возвращает отсортированный список ID всего дерева корневого контейнера
// плоским индексным запросом без рекурсии (работает при parent_id = 0, где root_id = id).
func (r *pgRepository) GetRootSubtreeContainerIDs(ctx context.Context, rootID int) ([]int, error) {
	rows, err := r.conn.QueryContext(ctx, `SELECT id FROM containers WHERE root_id = $1 ORDER BY id`, rootID)
	if err != nil {
		return nil, fmt.Errorf("get root subtree containers: %w", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// GetNestedSubtreeContainerIDs возвращает отсортированный список ID вложенного контейнера и всех его потомков
// через рекурсивный CTE (используется только когда извлекается или перекладывается не корневой узел).
func (r *pgRepository) GetNestedSubtreeContainerIDs(ctx context.Context, containerID int) ([]int, error) {
	query := `
		WITH RECURSIVE tree AS (
			SELECT id FROM containers WHERE id = $1
			UNION ALL
			SELECT c.id FROM containers c
			INNER JOIN tree t ON c.parent_id = t.id
		)
		SELECT id FROM tree ORDER BY id
	`
	rows, err := r.conn.QueryContext(ctx, query, containerID)
	if err != nil {
		return nil, fmt.Errorf("get nested subtree containers: %w", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// GetSubtreeBalances возвращает все ненулевые остатки (товары и контейнеры) внутри поддерева по заданным координатам.
func (r *pgRepository) GetSubtreeBalances(ctx context.Context, loc model.Location, containerIDs []int) ([]model.ContainerStockItem, error) {
	if len(containerIDs) == 0 {
		return nil, nil
	}
	query := `
		SELECT container_id, item_id, SUM(quantity) as qty
		FROM storage
		WHERE whs_id = $1
		  AND zone_id = $2
		  AND cell_id IS NOT DISTINCT FROM $3
		  AND container_id = ANY($4)
		GROUP BY container_id, item_id
		HAVING SUM(quantity) > 0
	`
	rows, err := r.conn.QueryContext(ctx, query, loc.WhsID, loc.ZoneID, loc.CellID, pq.Array(containerIDs))
	if err != nil {
		return nil, fmt.Errorf("get subtree balances: %w", err)
	}
	defer rows.Close()

	var items []model.ContainerStockItem
	for rows.Next() {
		var it model.ContainerStockItem
		if err := rows.Scan(&it.ContainerID, &it.ItemID, &it.Quantity); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// UpdateSubtreeHierarchy переподчиняет ветку контейнеров новому родителю в дереве тар (используется в Nest и Unnest):
//  1. Привязывает корень переносимой ветки (targetID) к новому родителю (newParentID, либо 0 при извлечении наружу).
//  2. Пересчитывает денормализованные поля дерева (root_id и depth) для самого targetID и всех его вложенных потомков (subtreeIDs).
//
// Параметры:
//   - targetID: ID контейнера, который вкладывают или извлекают;
//   - subtreeIDs: заранее собранный список ID самого targetID и всех его потомков (чтобы не запускать повторный CTE);
//   - newParentID: ID нового родителя (или 0, если контейнер становится самостоятельным корневым);
//   - newRootID: ID нового корневого контейнера дерева (или сам targetID при извлечении наружу);
//   - depthDelta: смещение уровня вложенности (например, +1 при вложении корня в другой корень или -2 при извлечении со 2-го уровня наружу).
func (r *pgRepository) UpdateSubtreeHierarchy(
	ctx context.Context,
	targetID int,
	subtreeIDs []int,
	newParentID, newRootID, depthDelta int,
) error {
	if _, err := r.conn.ExecContext(ctx, `UPDATE containers SET parent_id = $1 WHERE id = $2`, newParentID, targetID); err != nil {
		return fmt.Errorf("update target parent_id: %w", err)
	}

	if len(subtreeIDs) == 0 {
		return nil
	}

	_, err := r.conn.ExecContext(ctx, `
		UPDATE containers
		SET root_id = $1,
		    depth = depth + $2
		WHERE id = ANY($3)
	`, newRootID, depthDelta, pq.Array(subtreeIDs))
	if err != nil {
		return fmt.Errorf("update subtree root_id and depth: %w", err)
	}

	return nil
}

// GetDispatchLocation возвращает координаты системной зоны DISPATCH/экспедиции (cell_id = nil) для списания/выбытия.
//
//	Пока не реализованы механизмы отгрузки и размещения в зоне отгрузке, считаем, что ячеек нет.
//	В будущих версиях здесь будут подбираться ячейки, в зависимости от адресов, направлений, смен и способов сборки.. и т.д.
func (r *pgRepository) GetDispatchLocation(ctx context.Context, whsID int) (model.Location, error) {
	var loc model.Location
	query := `
		SELECT warehouse_id, id
		FROM zones
		WHERE warehouse_id = $1 AND type = 'DISPATCH' AND is_system = true
		LIMIT 1
	`
	err := r.conn.QueryRowContext(ctx, query, whsID).Scan(&loc.WhsID, &loc.ZoneID)
	if err != nil {
		return loc, fmt.Errorf("get system dispatch zone for whs %d: %w", whsID, err)
	}
	loc.CellID = nil

	return loc, nil
}

func (r *pgRepository) InsertRows(ctx context.Context, rows []model.StorageRow) error {
	stmt, err := r.conn.PrepareContext(ctx, `
		INSERT INTO storage (row_id, row_time, whs_id, zone_id, cell_id, container_id, item_id, quantity)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert storage: %w", err)
	}
	defer stmt.Close()

	for _, row := range rows {
		_, err := stmt.ExecContext(ctx,
			row.RowID, row.RowTime, row.WhsID, row.ZoneID,
			row.CellID, row.ContainerID, row.ItemID, row.Quantity,
		)
		if err != nil {
			return fmt.Errorf("exec insert storage: %w", err)
		}
	}
	return nil
}

func (r *pgRepository) GetInboxRecord(ctx context.Context, key string) (*model.InboxRecord, error) {
	var rec model.InboxRecord
	err := r.conn.QueryRowContext(ctx, `
		SELECT key, response_code, response_body
		FROM operations_inbox
		WHERE key = $1
	`, key).Scan(&rec.Key, &rec.ResponseCode, &rec.ResponseBody)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get inbox record: %w", err)
	}
	return &rec, nil
}

func (r *pgRepository) SaveInboxRecord(ctx context.Context, rec model.InboxRecord) error {
	body := "{}"
	if len(rec.ResponseBody) > 0 {
		body = string(rec.ResponseBody)
	}
	_, err := r.conn.ExecContext(ctx,
		`INSERT INTO operations_inbox (key, response_code, response_body) VALUES ($1, $2, $3::jsonb)`,
		rec.Key, rec.ResponseCode, body,
	)
	if err != nil {
		return fmt.Errorf("save inbox record: %w", err)
	}
	return nil
}

func (r *pgRepository) GetActiveReservedQty(ctx context.Context, loc model.Location, itemID *int) (int, error) {
	// TODO: заменить на запрос к таблице заданий на отбор / резервов после её внедрения
	return 0, nil
}

func (r *pgRepository) HasActiveReservationsForSubtree(ctx context.Context, loc model.Location, containerIDs []int) (bool, error) {
	// TODO: заменить на проверку активных резервов по поддереву контейнеров
	return false, nil
}
