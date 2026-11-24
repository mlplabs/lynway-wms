package wms

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/mlplabs/lynway-wms/internal/model"
	"github.com/mlplabs/lynway-wms/internal/repository"
)

func (s *Service) ExecuteIdempotent(
	ctx context.Context,
	idempotencyKey string,
	op func(txRepo repository.PgRepository) (code int, body []byte, err error),
) (int, []byte, error) {
	if idempotencyKey == "" {
		return 0, nil, model.ErrInvalidMovement
	}

	var resCode int
	var resBody []byte

	err := s.whs.WithinTx(ctx, func(txRepo repository.PgRepository) error {
		// 1. Блокируем ключ, чтобы параллельные ретраи с Android-клиента ждали первый запрос
		if err := txRepo.AcquireLock(ctx, "inbox:"+idempotencyKey); err != nil {
			return err
		}

		// 2. Если операция уже была выполнена ранее — возвращаем сохраненный HTTP-ответ без повторного списания
		existing, err := txRepo.GetInboxRecord(ctx, idempotencyKey)
		if err != nil {
			return err
		}
		if existing != nil {
			resCode = existing.ResponseCode
			resBody = existing.ResponseBody
			return nil
		}

		// 3. Выполняем саму складскую операцию на том же txRepo (внутри той же транзакции)
		code, body, err := op(txRepo)
		if err != nil {
			// Если бизнес-валидация упала (например, ErrInsufficientStock) — транзакция откатится,
			// запись в operations_inbox НЕ создастся, и клиент сможет исправить запрос и повторить с тем же ключом
			return err
		}

		// 4. Фиксируем результат в operations_inbox перед коммитом транзакции
		if err := txRepo.SaveInboxRecord(ctx, model.InboxRecord{
			Key:          idempotencyKey,
			ResponseCode: code,
			ResponseBody: body,
		}); err != nil {
			return err
		}

		resCode = code
		resBody = body
		return nil
	})

	return resCode, resBody, err
}

// --- Operation 1 & 2: Приемка товара (в зону, в ячейку или в контейнер) ---

func (s *Service) ReceiveItem(ctx context.Context, loc model.Location, itemID int, qty int) (uuid.UUID, error) {
	var rowID uuid.UUID
	err := s.whs.WithinTx(ctx, func(txRepo repository.PgRepository) error {
		var err error
		rowID, err = s.ReceiveItemTx(ctx, txRepo, loc, itemID, qty)
		return err
	})
	return rowID, err
}

func (s *Service) ReceiveItemTx(ctx context.Context, txRepo repository.PgRepository, loc model.Location, itemID int, qty int) (uuid.UUID, error) {
	if qty <= 0 || itemID <= 0 || loc.WhsID <= 0 || loc.ZoneID <= 0 {
		return uuid.Nil, model.ErrInvalidMovement
	}

	targetLoc := loc

	if loc.ContainerID != nil {
		if err := txRepo.AcquireSharedLock(ctx, fmt.Sprintf("cnt_tree:%d", *loc.ContainerID)); err != nil {
			return uuid.Nil, err
		}
		resolvedLoc, err := s.resolveContainerTargetLocation(ctx, txRepo, *loc.ContainerID, loc)
		if err != nil {
			return uuid.Nil, err
		}
		targetLoc = resolvedLoc
	}

	if err := s.validateLocationIn(ctx, txRepo, targetLoc); err != nil {
		return uuid.Nil, err
	}

	rowID := uuid.New()
	row := model.StorageRow{
		RowID:       rowID,
		RowTime:     time.Now().UTC(),
		WhsID:       targetLoc.WhsID,
		ZoneID:      targetLoc.ZoneID,
		CellID:      targetLoc.CellID,
		ContainerID: targetLoc.ContainerID,
		ItemID:      &itemID,
		Quantity:    qty,
	}
	if err := txRepo.InsertRows(ctx, []model.StorageRow{row}); err != nil {
		return uuid.Nil, err
	}
	return rowID, nil
}

// --- Operation 3: Перемещение товара между любыми координатами ---

func (s *Service) MoveItem(ctx context.Context, from, to model.Location, itemID int, qty int) (uuid.UUID, error) {
	var rowID uuid.UUID
	err := s.whs.WithinTx(ctx, func(txRepo repository.PgRepository) error {
		var err error
		rowID, err = s.MoveItemTx(ctx, txRepo, from, to, itemID, qty)
		return err
	})
	return rowID, err
}

func (s *Service) MoveItemTx(ctx context.Context, txRepo repository.PgRepository, from, to model.Location, itemID int, qty int) (uuid.UUID, error) {
	if qty <= 0 || itemID <= 0 || from.WhsID <= 0 || from.ZoneID <= 0 || to.WhsID <= 0 || to.ZoneID <= 0 {
		return uuid.Nil, model.ErrInvalidMovement
	}

	targetLoc := to

	var cntIDs []int
	if from.ContainerID != nil {
		cntIDs = append(cntIDs, *from.ContainerID)
	}
	if to.ContainerID != nil && (from.ContainerID == nil || *to.ContainerID != *from.ContainerID) {
		cntIDs = append(cntIDs, *to.ContainerID)
	}
	sort.Ints(cntIDs)
	for _, id := range cntIDs {
		if err := txRepo.AcquireSharedLock(ctx, fmt.Sprintf("cnt_tree:%d", id)); err != nil {
			return uuid.Nil, err
		}
	}

	if to.ContainerID != nil {
		resolvedLoc, err := s.resolveContainerTargetLocation(ctx, txRepo, *to.ContainerID, to)
		if err != nil {
			return uuid.Nil, err
		}
		targetLoc = resolvedLoc
	}

	if err := s.validateLocationOut(ctx, txRepo, from); err != nil {
		return uuid.Nil, err
	}
	if err := s.validateLocationIn(ctx, txRepo, targetLoc); err != nil {
		return uuid.Nil, err
	}

	keys := sortedUniqueKeys(
		buildLockKey(from, &itemID),
		buildLockKey(targetLoc, &itemID),
	)
	for _, k := range keys {
		if err := txRepo.AcquireLock(ctx, k); err != nil {
			return uuid.Nil, err
		}
	}

	balance, err := txRepo.GetBalance(ctx, from, &itemID)
	if err != nil {
		return uuid.Nil, err
	}
	if balance < qty {
		return uuid.Nil, fmt.Errorf("%w: available %d, requested %d", model.ErrInsufficientStock, balance, qty)
	}

	reservedQty, err := s.whs.GetActiveReservedQty(ctx, from, &itemID)
	if err != nil {
		return uuid.Nil, err
	}
	if balance-reservedQty < qty {
		return uuid.Nil, fmt.Errorf("%w: physical %d, reserved %d, requested %d", model.ErrLocationReserved, balance, reservedQty, qty)
	}

	rowID := uuid.New()
	now := time.Now().UTC()
	rows := []model.StorageRow{
		{
			RowID:       rowID,
			RowTime:     now,
			WhsID:       from.WhsID,
			ZoneID:      from.ZoneID,
			CellID:      from.CellID,
			ContainerID: from.ContainerID,
			ItemID:      &itemID,
			Quantity:    -qty,
		},
		{
			RowID:       rowID,
			RowTime:     now,
			WhsID:       targetLoc.WhsID,
			ZoneID:      targetLoc.ZoneID,
			CellID:      targetLoc.CellID,
			ContainerID: targetLoc.ContainerID,
			ItemID:      &itemID,
			Quantity:    qty,
		},
	}
	if err := txRepo.InsertRows(ctx, rows); err != nil {
		return uuid.Nil, err
	}
	return rowID, nil
}

// --- Operation 4: Первичное размещение внешнего контейнера в зону/ячейку ---

func (s *Service) PlaceContainer(ctx context.Context, to model.Location, containerID int) (uuid.UUID, error) {
	var rowID uuid.UUID
	err := s.whs.WithinTx(ctx, func(txRepo repository.PgRepository) error {
		var err error
		rowID, err = s.PlaceContainerTx(ctx, txRepo, to, containerID)
		return err
	})
	return rowID, err
}

func (s *Service) PlaceContainerTx(ctx context.Context, txRepo repository.PgRepository, to model.Location, containerID int) (uuid.UUID, error) {
	if containerID <= 0 || to.WhsID <= 0 || to.ZoneID <= 0 {
		return uuid.Nil, model.ErrInvalidMovement
	}

	cMeta, err := txRepo.GetContainerMeta(ctx, containerID)
	if err != nil {
		return uuid.Nil, err
	}
	if cMeta.ParentID != 0 {
		return uuid.Nil, model.ErrContainerNotRoot
	}

	if err := txRepo.AcquireLock(ctx, fmt.Sprintf("cnt_tree:%d", containerID)); err != nil {
		return uuid.Nil, err
	}

	curLoc, err := txRepo.GetContainerCurrentLocation(ctx, containerID)
	if err != nil {
		return uuid.Nil, err
	}
	if curLoc != nil {
		return uuid.Nil, model.ErrContainerHasCell
	}

	if err := s.validateLocationIn(ctx, txRepo, to); err != nil {
		return uuid.Nil, err
	}

	rowID := uuid.New()
	row := model.StorageRow{
		RowID:       rowID,
		RowTime:     time.Now().UTC(),
		WhsID:       to.WhsID,
		ZoneID:      to.ZoneID,
		CellID:      to.CellID,
		ContainerID: &containerID,
		ItemID:      nil,
		Quantity:    1,
	}
	if err := txRepo.InsertRows(ctx, []model.StorageRow{row}); err != nil {
		return uuid.Nil, err
	}
	return rowID, nil
}

// --- Operation 5: Перемещение внешнего контейнера между зонами/ячейками ---

func (s *Service) MoveContainer(ctx context.Context, from, to model.Location, containerID int) (uuid.UUID, error) {
	var rowID uuid.UUID
	err := s.whs.WithinTx(ctx, func(txRepo repository.PgRepository) error {
		var err error
		rowID, err = s.MoveContainerTx(ctx, txRepo, from, to, containerID)
		return err
	})
	return rowID, err
}

func (s *Service) MoveContainerTx(ctx context.Context, txRepo repository.PgRepository, from, to model.Location, containerID int) (uuid.UUID, error) {
	if containerID <= 0 || from.WhsID <= 0 || from.ZoneID <= 0 || to.WhsID <= 0 || to.ZoneID <= 0 {
		return uuid.Nil, model.ErrInvalidMovement
	}

	cMeta, err := txRepo.GetContainerMeta(ctx, containerID)
	if err != nil {
		return uuid.Nil, err
	}
	if cMeta.ParentID != 0 {
		return uuid.Nil, model.ErrContainerNotRoot
	}

	if err := s.validateLocationOut(ctx, txRepo, from); err != nil {
		return uuid.Nil, err
	}
	if err := s.validateLocationIn(ctx, txRepo, to); err != nil {
		return uuid.Nil, err
	}

	subtreeIDs, err := txRepo.GetRootSubtreeContainerIDs(ctx, containerID)
	if err != nil {
		return uuid.Nil, err
	}

	rowID := uuid.New()
	if err := s.relocateSubtree(ctx, txRepo, rowID, subtreeIDs, from, to); err != nil {
		return uuid.Nil, err
	}
	return rowID, nil
}

// --- Operation 6: Вложение контейнера (childID) внутрь другого контейнера (targetParentID) ---

func (s *Service) NestContainer(ctx context.Context, childID, targetParentID int) (uuid.UUID, error) {
	var rowID uuid.UUID
	err := s.whs.WithinTx(ctx, func(txRepo repository.PgRepository) error {
		var err error
		rowID, err = s.NestContainerTx(ctx, txRepo, childID, targetParentID)
		return err
	})
	return rowID, err
}

func (s *Service) NestContainerTx(ctx context.Context, txRepo repository.PgRepository, childID, targetParentID int) (uuid.UUID, error) {
	if childID <= 0 || targetParentID <= 0 || childID == targetParentID {
		return uuid.Nil, model.ErrCyclicNesting
	}

	childMeta, err := txRepo.GetContainerMeta(ctx, childID)
	if err != nil {
		return uuid.Nil, err
	}
	parentMeta, err := txRepo.GetContainerMeta(ctx, targetParentID)
	if err != nil {
		return uuid.Nil, err
	}

	childSubtree, err := s.getSubtreeIDsByMeta(ctx, txRepo, childMeta)
	if err != nil {
		return uuid.Nil, err
	}

	if parentMeta.RootID == childMeta.RootID {
		for _, id := range childSubtree {
			if id == targetParentID {
				return uuid.Nil, model.ErrCyclicNesting
			}
		}
	}

	fromLoc, err := txRepo.GetContainerCurrentLocation(ctx, childMeta.RootID)
	if err != nil {
		return uuid.Nil, err
	}
	if fromLoc == nil {
		return uuid.Nil, model.ErrContainerNoCell
	}

	toLoc, err := txRepo.GetContainerCurrentLocation(ctx, parentMeta.RootID)
	if err != nil {
		return uuid.Nil, err
	}
	if toLoc == nil {
		return uuid.Nil, model.ErrContainerNoCell
	}

	rowID := uuid.New()
	if !sameLocation(*fromLoc, *toLoc) {
		if err := s.validateLocationOut(ctx, txRepo, *fromLoc); err != nil {
			return uuid.Nil, err
		}
		if err := s.validateLocationIn(ctx, txRepo, *toLoc); err != nil {
			return uuid.Nil, err
		}
		if err := s.relocateSubtree(ctx, txRepo, rowID, childSubtree, *fromLoc, *toLoc); err != nil {
			return uuid.Nil, err
		}
	} else {
		for _, id := range childSubtree {
			if err := txRepo.AcquireLock(ctx, fmt.Sprintf("cnt_tree:%d", id)); err != nil {
				return uuid.Nil, err
			}
		}
		hasRes, err := s.whs.HasActiveReservationsForSubtree(ctx, *fromLoc, childSubtree)
		if err != nil {
			return uuid.Nil, err
		}
		if hasRes {
			return uuid.Nil, model.ErrLocationReserved
		}
	}

	newDepth := parentMeta.Depth + 1
	depthDelta := newDepth - childMeta.Depth
	if err := txRepo.UpdateSubtreeHierarchy(ctx, childID, childSubtree, targetParentID, parentMeta.RootID, depthDelta); err != nil {
		return uuid.Nil, err
	}
	return rowID, nil
}

// --- Operation 7: Извлечение вложенного контейнера наружу (становится root: parent_id = 0) ---

func (s *Service) UnnestContainer(ctx context.Context, containerID int, to model.Location) (uuid.UUID, error) {
	var rowID uuid.UUID
	err := s.whs.WithinTx(ctx, func(txRepo repository.PgRepository) error {
		var err error
		rowID, err = s.UnnestContainerTx(ctx, txRepo, containerID, to)
		return err
	})
	return rowID, err
}

func (s *Service) UnnestContainerTx(ctx context.Context, txRepo repository.PgRepository, containerID int, to model.Location) (uuid.UUID, error) {
	if containerID <= 0 || to.WhsID <= 0 || to.ZoneID <= 0 {
		return uuid.Nil, model.ErrInvalidMovement
	}

	cMeta, err := txRepo.GetContainerMeta(ctx, containerID)
	if err != nil {
		return uuid.Nil, err
	}
	if cMeta.ParentID == 0 {
		return uuid.Nil, model.ErrInvalidMovement
	}

	fromLoc, err := txRepo.GetContainerCurrentLocation(ctx, cMeta.RootID)
	if err != nil {
		return uuid.Nil, err
	}
	if fromLoc == nil {
		return uuid.Nil, model.ErrContainerNoCell
	}

	subtreeIDs, err := s.getSubtreeIDsByMeta(ctx, txRepo, cMeta)
	if err != nil {
		return uuid.Nil, err
	}

	rowID := uuid.New()
	if !sameLocation(*fromLoc, to) {
		if err := s.validateLocationOut(ctx, txRepo, *fromLoc); err != nil {
			return uuid.Nil, err
		}
		if err := s.validateLocationIn(ctx, txRepo, to); err != nil {
			return uuid.Nil, err
		}
		if err := s.relocateSubtree(ctx, txRepo, rowID, subtreeIDs, *fromLoc, to); err != nil {
			return uuid.Nil, err
		}
	} else {
		for _, id := range subtreeIDs {
			if err := txRepo.AcquireLock(ctx, fmt.Sprintf("cnt_tree:%d", id)); err != nil {
				return uuid.Nil, err
			}
		}
		hasRes, err := s.whs.HasActiveReservationsForSubtree(ctx, *fromLoc, subtreeIDs)
		if err != nil {
			return uuid.Nil, err
		}
		if hasRes {
			return uuid.Nil, model.ErrLocationReserved
		}
	}

	depthDelta := -cMeta.Depth
	if err := txRepo.UpdateSubtreeHierarchy(ctx, containerID, subtreeIDs, 0, containerID, depthDelta); err != nil {
		return uuid.Nil, err
	}
	return rowID, nil
}

// --- Operation 8: Выбытие / Списание (в системную зону DISPATCH с cell_id = NULL) ---

func (s *Service) DispatchItem(ctx context.Context, from model.Location, itemID int, qty int) (uuid.UUID, error) {
	var rowID uuid.UUID
	err := s.whs.WithinTx(ctx, func(txRepo repository.PgRepository) error {
		var err error
		rowID, err = s.DispatchItemTx(ctx, txRepo, from, itemID, qty)
		return err
	})
	return rowID, err
}

func (s *Service) DispatchItemTx(ctx context.Context, txRepo repository.PgRepository, from model.Location, itemID int, qty int) (uuid.UUID, error) {
	dispatchLoc, err := txRepo.GetDispatchLocation(ctx, from.WhsID)
	if err != nil {
		return uuid.Nil, err
	}
	return s.MoveItemTx(ctx, txRepo, from, dispatchLoc, itemID, qty)
}

func (s *Service) DispatchContainer(ctx context.Context, from model.Location, containerID int) (uuid.UUID, error) {
	var rowID uuid.UUID
	err := s.whs.WithinTx(ctx, func(txRepo repository.PgRepository) error {
		var err error
		rowID, err = s.DispatchContainerTx(ctx, txRepo, from, containerID)
		return err
	})
	return rowID, err
}

func (s *Service) DispatchContainerTx(ctx context.Context, txRepo repository.PgRepository, from model.Location, containerID int) (uuid.UUID, error) {
	dispatchLoc, err := txRepo.GetDispatchLocation(ctx, from.WhsID)
	if err != nil {
		return uuid.Nil, err
	}
	return s.MoveContainerTx(ctx, txRepo, from, dispatchLoc, containerID)
}

// --- Внутренние хелперы ---

func (s *Service) getSubtreeIDsByMeta(ctx context.Context, txRepo repository.PgRepository, meta model.ContainerMeta) ([]int, error) {
	if meta.ParentID == 0 {
		return txRepo.GetRootSubtreeContainerIDs(ctx, meta.ID)
	}
	return txRepo.GetNestedSubtreeContainerIDs(ctx, meta.ID)
}

func (s *Service) resolveContainerTargetLocation(ctx context.Context, txRepo repository.PgRepository, containerID int, fallback model.Location) (model.Location, error) {
	cMeta, err := txRepo.GetContainerMeta(ctx, containerID)
	if err != nil {
		return fallback, err
	}
	rootLoc, err := txRepo.GetContainerCurrentLocation(ctx, cMeta.RootID)
	if err != nil {
		return fallback, err
	}
	res := fallback
	if rootLoc != nil {
		res.WhsID = rootLoc.WhsID
		res.ZoneID = rootLoc.ZoneID
		res.CellID = rootLoc.CellID
	}
	return res, nil
}

func (s *Service) relocateSubtree(
	ctx context.Context,
	txRepo repository.PgRepository,
	rowID uuid.UUID,
	subtreeIDs []int,
	from, to model.Location,
) error {
	for _, id := range subtreeIDs {
		if err := txRepo.AcquireLock(ctx, fmt.Sprintf("cnt_tree:%d", id)); err != nil {
			return err
		}
	}

	hasRes, err := s.whs.HasActiveReservationsForSubtree(ctx, from, subtreeIDs)
	if err != nil {
		return err
	}
	if hasRes {
		return model.ErrLocationReserved
	}

	balances, err := txRepo.GetSubtreeBalances(ctx, from, subtreeIDs)
	if err != nil {
		return err
	}
	if len(balances) == 0 {
		return model.ErrInsufficientStock
	}

	now := time.Now().UTC()
	rows := make([]model.StorageRow, 0, len(balances)*2)

	for _, b := range balances {
		cntID := b.ContainerID
		rows = append(rows,
			model.StorageRow{
				RowID:       rowID,
				RowTime:     now,
				WhsID:       from.WhsID,
				ZoneID:      from.ZoneID,
				CellID:      from.CellID,
				ContainerID: &cntID,
				ItemID:      b.ItemID,
				Quantity:    -b.Quantity,
			},
			model.StorageRow{
				RowID:       rowID,
				RowTime:     now,
				WhsID:       to.WhsID,
				ZoneID:      to.ZoneID,
				CellID:      to.CellID,
				ContainerID: &cntID,
				ItemID:      b.ItemID,
				Quantity:    b.Quantity,
			},
		)
	}

	return txRepo.InsertRows(ctx, rows)
}

func (s *Service) validateLocationIn(ctx context.Context, txRepo repository.PgRepository, loc model.Location) error {
	zMeta, err := txRepo.GetZoneMeta(ctx, loc.ZoneID)
	if err != nil {
		return err
	}
	if zMeta.WhsID != loc.WhsID {
		return model.ErrZoneWhsMismatch
	}

	if loc.CellID == nil {
		return nil
	}

	cMeta, err := txRepo.GetCellMeta(ctx, *loc.CellID)
	if err != nil {
		return err
	}
	if cMeta.WhsID != loc.WhsID || cMeta.ZoneID != loc.ZoneID {
		return model.ErrCellZoneMismatch
	}
	if cMeta.NotAllowedIn {
		return model.ErrCellNotAllowedIn
	}
	return nil
}

func (s *Service) validateLocationOut(ctx context.Context, txRepo repository.PgRepository, loc model.Location) error {
	zMeta, err := txRepo.GetZoneMeta(ctx, loc.ZoneID)
	if err != nil {
		return err
	}
	if zMeta.WhsID != loc.WhsID {
		return model.ErrZoneWhsMismatch
	}
	if zMeta.Type == "DISPATCH" {
		return model.ErrZoneNotAllowedOut
	}

	if loc.CellID == nil {
		return nil
	}

	cMeta, err := txRepo.GetCellMeta(ctx, *loc.CellID)
	if err != nil {
		return err
	}
	if cMeta.WhsID != loc.WhsID || cMeta.ZoneID != loc.ZoneID {
		return model.ErrCellZoneMismatch
	}
	if cMeta.NotAllowedOut {
		return model.ErrCellNotAllowedOut
	}

	return nil
}

func buildLockKey(loc model.Location, itemID *int) string {
	cID, cntID, iID := 0, 0, 0
	if loc.CellID != nil {
		cID = *loc.CellID
	}
	if loc.ContainerID != nil {
		cntID = *loc.ContainerID
	}
	if itemID != nil {
		iID = *itemID
	}

	return fmt.Sprintf("stk:%d:%d:%d:%d:%d", loc.WhsID, loc.ZoneID, cID, cntID, iID)
}

func sortedUniqueKeys(keys ...string) []string {
	sort.Strings(keys)
	res := make([]string, 0, len(keys))
	for i, k := range keys {
		if i == 0 || k != keys[i-1] {
			res = append(res, k)
		}
	}

	return res
}

func sameLocation(a, b model.Location) bool {
	if a.WhsID != b.WhsID || a.ZoneID != b.ZoneID {
		return false
	}
	if a.CellID == nil && b.CellID == nil {
		return true
	}
	if a.CellID != nil && b.CellID != nil && *a.CellID == *b.CellID {
		return true
	}

	return false
}
