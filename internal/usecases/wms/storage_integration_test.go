package wms_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/mlplabs/lynway-wms/internal/repository/bootstrap"

	"github.com/mlplabs/lynway-wms/internal/model"
	"github.com/mlplabs/lynway-wms/internal/repository"
	"github.com/mlplabs/lynway-wms/internal/usecases/wms"
)

const defaultTestDSN = "postgres://wms_test:wms_test@localhost:5433/wms_test?sslmode=disable"

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("skipping integration test: failed to open db: %v", err)
	}

	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var pingErr error
	for i := 0; i < 25; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if pingErr != nil {
		t.Fatalf("db never became ready: %v", pingErr)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	applyMigrations(t, db)
	return db
}

func applyMigrations(t *testing.T, db *sql.DB) {
	t.Helper()

	err := bootstrap.RunMigrations(db)
	if err != nil {
		log.Printf("failed to run migrations: %s", err)
		return
	}

	err = bootstrap.SeedDefaults(t.Context(), db)
	if err != nil {
		log.Printf("failed to seed defaults: %s", err)
		return
	}
}

type testEnv struct {
	whsID         int
	receivingZone int
	storageZone   int
	dispatchZone  int
}

func createIsolatedWarehouse(t *testing.T, db *sql.DB, suffix string) testEnv {
	t.Helper()

	var env testEnv
	err := db.QueryRow(
		`INSERT INTO warehouses (name) VALUES ($1) RETURNING id`,
		fmt.Sprintf("Test Whs %s %d", suffix, time.Now().UnixNano()),
	).Scan(&env.whsID)
	if err != nil {
		t.Fatalf("insert test warehouse: %v", err)
	}

	zones := []struct {
		name  string
		zType string
		dst   *int
	}{
		{"Приемка", "RECEIVING", &env.receivingZone},
		{"Хранение", "STORAGE", &env.storageZone},
		{"Отгрузка", "DISPATCH", &env.dispatchZone},
	}

	for _, z := range zones {
		err := db.QueryRow(
			`INSERT INTO zones (name, type, is_system, warehouse_id) VALUES ($1, $2, true, $3) RETURNING id`,
			z.name, z.zType, env.whsID,
		).Scan(z.dst)
		if err != nil {
			t.Fatalf("insert zone %s: %v", z.zType, err)
		}
	}

	return env
}

func createCell(t *testing.T, db *sql.DB, whsID, zoneID int, name string, notIn, notOut bool) int {
	t.Helper()
	var id int
	err := db.QueryRow(`
		INSERT INTO cells (name, whs_id, zone_id, not_allowed_in, not_allowed_out, is_service)
		VALUES ($1, $2, $3, $4, $5, false)
		RETURNING id
	`, name, whsID, zoneID, notIn, notOut).Scan(&id)
	if err != nil {
		t.Fatalf("insert cell %s: %v", name, err)
	}
	return id
}

func createItem(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	var id int
	err := db.QueryRow(`INSERT INTO items (name) VALUES ($1) RETURNING id`, name).Scan(&id)
	if err != nil {
		t.Fatalf("insert item %s: %v", name, err)
	}
	return id
}

func createContainer(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	var id int
	err := db.QueryRow(`
		INSERT INTO containers (name, type, parent_id, root_id, depth)
		VALUES ($1, 0, 0, 0, 0)
		RETURNING id
	`, name).Scan(&id)
	if err != nil {
		t.Fatalf("insert container %s: %v", name, err)
	}
	if _, err := db.Exec(`UPDATE containers SET root_id = $1 WHERE id = $1`, id); err != nil {
		t.Fatalf("set root_id for container %d: %v", id, err)
	}
	return id
}

func queryBalance(t *testing.T, db *sql.DB, loc model.Location, itemID *int) int {
	t.Helper()
	var bal int
	err := db.QueryRow(`
		SELECT COALESCE(SUM(quantity), 0)
		FROM storage
		WHERE whs_id = $1
		  AND zone_id = $2
		  AND cell_id IS NOT DISTINCT FROM $3
		  AND container_id IS NOT DISTINCT FROM $4
		  AND item_id IS NOT DISTINCT FROM $5
	`, loc.WhsID, loc.ZoneID, loc.CellID, loc.ContainerID, itemID).Scan(&bal)
	if err != nil {
		t.Fatalf("query balance: %v", err)
	}
	return bal
}

func TestStorageMatrix_ItemsAndCells(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewWhsRepository(db)
	svc := wms.NewService(nil, repo, nil, nil)
	ctx := context.Background()

	t.Run("Op1_Op3_ReceiveAndMoveItem", func(t *testing.T) {
		t.Parallel()
		env := createIsolatedWarehouse(t, db, "items")
		itemID := createItem(t, db, "Болт М8")
		cellA := createCell(t, db, env.whsID, env.storageZone, "A-01", false, false)
		cellBlockedIn := createCell(t, db, env.whsID, env.storageZone, "A-LOCK-IN", true, false)

		zoneLoc := model.Location{WhsID: env.whsID, ZoneID: env.receivingZone}
		cellALoc := model.Location{WhsID: env.whsID, ZoneID: env.storageZone, CellID: &cellA}
		blockedLoc := model.Location{WhsID: env.whsID, ZoneID: env.storageZone, CellID: &cellBlockedIn}

		// Op 1: Приемка в зону без ячейки (cell_id = nil)
		if _, err := svc.ReceiveItem(ctx, zoneLoc, itemID, 50); err != nil {
			t.Fatalf("receive to zone failed: %v", err)
		}
		if got := queryBalance(t, db, zoneLoc, &itemID); got != 50 {
			t.Fatalf("expected zone balance 50, got %d", got)
		}

		// Проверка запрета на вход
		if _, err := svc.MoveItem(ctx, zoneLoc, blockedLoc, itemID, 10); !errors.Is(err, model.ErrCellNotAllowedIn) {
			t.Fatalf("expected ErrCellNotAllowedIn, got %v", err)
		}

		// Op 3: Перемещение из зоны в ячейку A-01
		if _, err := svc.MoveItem(ctx, zoneLoc, cellALoc, itemID, 30); err != nil {
			t.Fatalf("move to cell A failed: %v", err)
		}
		if got := queryBalance(t, db, zoneLoc, &itemID); got != 20 {
			t.Fatalf("expected remaining zone balance 20, got %d", got)
		}
		if got := queryBalance(t, db, cellALoc, &itemID); got != 30 {
			t.Fatalf("expected cell A balance 30, got %d", got)
		}

		// Ошибка при попытке списать больше остатка
		if _, err := svc.MoveItem(ctx, cellALoc, zoneLoc, itemID, 31); !errors.Is(err, model.ErrInsufficientStock) {
			t.Fatalf("expected ErrInsufficientStock, got %v", err)
		}
	})
}

func TestStorageMatrix_ContainersHierarchy(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewWhsRepository(db)
	svc := wms.NewService(nil, repo, nil, nil)
	ctx := context.Background()

	t.Run("Op2_Op4_Op5_Op6_Op7_ContainerLifecycle", func(t *testing.T) {
		t.Parallel()
		env := createIsolatedWarehouse(t, db, "containers")
		itemID := createItem(t, db, "Гайка М8")
		cellA := createCell(t, db, env.whsID, env.storageZone, "B-01", false, false)
		cellB := createCell(t, db, env.whsID, env.storageZone, "B-02", false, false)

		pallet := createContainer(t, db, "Паллета-1")
		box := createContainer(t, db, "Коробка-1")

		locA := model.Location{WhsID: env.whsID, ZoneID: env.storageZone, CellID: &cellA}
		locB := model.Location{WhsID: env.whsID, ZoneID: env.storageZone, CellID: &cellB}

		// Op 4: Размещаем паллету и коробку в ячейке A
		if _, err := svc.PlaceContainer(ctx, locA, pallet); err != nil {
			t.Fatalf("place pallet failed: %v", err)
		}
		if _, err := svc.PlaceContainer(ctx, locA, box); err != nil {
			t.Fatalf("place box failed: %v", err)
		}

		// Op 2: Принимаем 100 шт товара в коробку (наследует cellA)
		boxLocA := model.Location{WhsID: env.whsID, ZoneID: env.storageZone, CellID: &cellA, ContainerID: &box}
		if _, err := svc.ReceiveItem(ctx, boxLocA, itemID, 100); err != nil {
			t.Fatalf("receive to box failed: %v", err)
		}
		if got := queryBalance(t, db, boxLocA, &itemID); got != 100 {
			t.Fatalf("expected 100 items in box at cellA, got %d", got)
		}

		// Op 6: Вкладываем коробку в паллету
		if _, err := svc.NestContainer(ctx, box, pallet); err != nil {
			t.Fatalf("nest box into pallet failed: %v", err)
		}

		// Защита от циклов (паллету в коробку нельзя)
		if _, err := svc.NestContainer(ctx, pallet, box); !errors.Is(err, model.ErrCyclicNesting) {
			t.Fatalf("expected ErrCyclicNesting, got %v", err)
		}

		// Op 5: Перемещаем паллету из ячейки A в ячейку B — содержимое должно перепровестись с cellB
		if _, err := svc.MoveContainer(ctx, locA, locB, pallet); err != nil {
			t.Fatalf("move pallet A->B failed: %v", err)
		}

		boxLocB := model.Location{WhsID: env.whsID, ZoneID: env.storageZone, CellID: &cellB, ContainerID: &box}
		if got := queryBalance(t, db, boxLocA, &itemID); got != 0 {
			t.Fatalf("expected 0 items left in cellA after pallet move, got %d", got)
		}
		if got := queryBalance(t, db, boxLocB, &itemID); got != 100 {
			t.Fatalf("expected 100 items moved to cellB with pallet, got %d", got)
		}

		// Op 7: Извлекаем коробку из паллеты обратно в ячейку A
		if _, err := svc.UnnestContainer(ctx, box, locA); err != nil {
			t.Fatalf("unnest box to cellA failed: %v", err)
		}
		if got := queryBalance(t, db, boxLocA, &itemID); got != 100 {
			t.Fatalf("expected 100 items back in cellA after unnest, got %d", got)
		}
		if got := queryBalance(t, db, boxLocB, &itemID); got != 0 {
			t.Fatalf("expected 0 items in cellB after unnest, got %d", got)
		}
	})
}

func TestStorageMatrix_DispatchAndIdempotency(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewWhsRepository(db)
	svc := wms.NewService(nil, repo, nil, nil)
	ctx := context.Background()

	t.Run("Op8_DispatchAndIdempotency", func(t *testing.T) {
		t.Parallel()
		env := createIsolatedWarehouse(t, db, "dispatch")
		itemID := createItem(t, db, "Шайба М8")
		cellA := createCell(t, db, env.whsID, env.storageZone, "C-01", false, false)

		locA := model.Location{WhsID: env.whsID, ZoneID: env.storageZone, CellID: &cellA}
		dispatchLoc := model.Location{WhsID: env.whsID, ZoneID: env.dispatchZone, CellID: nil}

		if _, err := svc.ReceiveItem(ctx, locA, itemID, 40); err != nil {
			t.Fatalf("receive failed: %v", err)
		}

		// Идемпотентное списание (Op 8) — два вызова с одним ключом спишут товар только один раз
		idemKey := fmt.Sprintf("dispatch-%d", time.Now().UnixNano())
		var firstRowID uuid.UUID
		for i := 0; i < 2; i++ {
			rowID, err := svc.ExecuteIdempotent(ctx, idemKey, func(txRepo repository.PgRepository) (uuid.UUID, error) {
				return svc.DispatchItemTx(ctx, txRepo, locA, itemID, 15)
			})
			if err != nil {
				t.Fatalf("idempotent call %d failed: err=%v", i, err)
			}
			if i == 0 {
				firstRowID = rowID
			} else if rowID != firstRowID {
				t.Fatalf("expected same rowID %s on retry, got %s", firstRowID, rowID)
			}
		}

		if got := queryBalance(t, db, locA, &itemID); got != 25 {
			t.Fatalf("expected 25 left in cellA after idempotent dispatch, got %d", got)
		}
		if got := queryBalance(t, db, dispatchLoc, &itemID); got != 15 {
			t.Fatalf("expected 15 in dispatch zone (cell_id=nil), got %d", got)
		}

		// Запрет на возврат товара из системной зоны DISPATCH
		if _, err := svc.MoveItem(ctx, dispatchLoc, locA, itemID, 5); !errors.Is(err, model.ErrZoneNotAllowedOut) {
			t.Fatalf("expected ErrZoneNotAllowedOut when moving out of DISPATCH, got %v", err)
		}
	})
}

func TestStorageMatrix_Concurrency(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewWhsRepository(db)
	svc := wms.NewService(nil, repo, nil, nil)
	ctx := context.Background()

	t.Run("ConcurrentWithdrawals_NoNegativeBalance", func(t *testing.T) {
		t.Parallel()
		env := createIsolatedWarehouse(t, db, "conc_withdraw")
		itemID := createItem(t, db, "Дефицитный товар")
		cellA := createCell(t, db, env.whsID, env.storageZone, "D-01", false, false)
		cellB := createCell(t, db, env.whsID, env.storageZone, "D-02", false, false)

		locA := model.Location{WhsID: env.whsID, ZoneID: env.storageZone, CellID: &cellA}
		locB := model.Location{WhsID: env.whsID, ZoneID: env.storageZone, CellID: &cellB}

		if _, err := svc.ReceiveItem(ctx, locA, itemID, 10); err != nil {
			t.Fatalf("initial receive failed: %v", err)
		}

		const workers = 25
		var successCount int32
		var insufficientCount int32
		var wg sync.WaitGroup

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Обязательно сохраняем результат в локальную переменную moveErr, а не во внешнюю err
				_, moveErr := svc.MoveItem(ctx, locA, locB, itemID, 1)
				if moveErr == nil {
					atomic.AddInt32(&successCount, 1)
				} else if errors.Is(moveErr, model.ErrInsufficientStock) {
					atomic.AddInt32(&insufficientCount, 1)
				} else {
					t.Errorf("unexpected error during concurrent move: %v", moveErr)
				}
			}()
		}
		wg.Wait()

		if successCount != 10 {
			t.Fatalf("expected exactly 10 successful moves, got %d", successCount)
		}
		if insufficientCount != 15 {
			t.Fatalf("expected exactly 15 ErrInsufficientStock errors, got %d", insufficientCount)
		}
		if got := queryBalance(t, db, locA, &itemID); got != 0 {
			t.Fatalf("expected final locA balance 0, got %d", got)
		}
		if got := queryBalance(t, db, locB, &itemID); got != 10 {
			t.Fatalf("expected final locB balance 10, got %d", got)
		}
	})

	t.Run("ConcurrentReceiveAndMoveContainer_NoOrphanedStock", func(t *testing.T) {
		t.Parallel()
		env := createIsolatedWarehouse(t, db, "conc_cnt")
		itemID := createItem(t, db, "Товар в коробку")
		cellA := createCell(t, db, env.whsID, env.storageZone, "E-01", false, false)
		cellB := createCell(t, db, env.whsID, env.storageZone, "E-02", false, false)
		box := createContainer(t, db, "Коробка-Гонка")

		locA := model.Location{WhsID: env.whsID, ZoneID: env.storageZone, CellID: &cellA}
		locB := model.Location{WhsID: env.whsID, ZoneID: env.storageZone, CellID: &cellB}

		if _, err := svc.PlaceContainer(ctx, locA, box); err != nil {
			t.Fatalf("place box failed: %v", err)
		}

		const packers = 12
		var wg sync.WaitGroup

		// 12 кладовщиков одновременно кладут по 5 шт товара в коробку
		for i := 0; i < packers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				target := model.Location{WhsID: env.whsID, ZoneID: env.storageZone, ContainerID: &box}
				if _, err := svc.ReceiveItem(ctx, target, itemID, 5); err != nil {
					t.Errorf("concurrent receive to container failed: %v", err)
				}
			}()
		}

		// В это же время перемещаем коробку из A в B
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(2 * time.Millisecond)
			if _, err := svc.MoveContainer(ctx, locA, locB, box); err != nil {
				t.Errorf("concurrent move container failed: %v", err)
			}
		}()

		wg.Wait()

		boxAtA := model.Location{WhsID: env.whsID, ZoneID: env.storageZone, CellID: &cellA, ContainerID: &box}
		boxAtB := model.Location{WhsID: env.whsID, ZoneID: env.storageZone, CellID: &cellB, ContainerID: &box}

		if gotA := queryBalance(t, db, boxAtA, &itemID); gotA != 0 {
			t.Fatalf("expected 0 items orphaned in cellA, got %d", gotA)
		}
		if gotB := queryBalance(t, db, boxAtB, &itemID); gotB != packers*5 {
			t.Fatalf("expected all %d items in cellB, got %d", packers*5, gotB)
		}
	})
}
