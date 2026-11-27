package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	lynway "github.com/mlplabs/lynway-wms"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultLogin     = "admin"
	defaultPassword  = "admin"
	defaultAdminName = "Администратор"
	defaultRole      = "admin"
)

func RunMigrations(db *sql.DB) error {
	source, err := iofs.New(lynway.MigrationsFS, "migrations/wms_db")
	if err != nil {
		return fmt.Errorf("init embedded migrations: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres migrate driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	// Не вызываем m.Close(), так как он закроет переданный инстанс *sql.DB

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}

	return nil
}

// seedDefaults инициализирует пустую базу: 1 склад, 3 системные зоны с сервисными ячейками и admin/admin.
func SeedDefaults(ctx context.Context, db *sql.DB) error {
	var whsCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM public.warehouses`).Scan(&whsCount); err != nil {
		return fmt.Errorf("check warehouses count: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin seed tx: %w", err)
	}
	defer tx.Rollback()

	if whsCount == 0 {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO public.warehouses (id, name) VALUES (1, 'Мой склад')
		`); err != nil {
			return fmt.Errorf("seed warehouse: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO public.zones (id, name, type, is_system, warehouse_id) VALUES
				(1, 'Приемка',  'RECEIVING', true, 1),
				(2, 'Хранение', 'STORAGE',   true, 1),
				(3, 'Отгрузка', 'DISPATCH',  true, 1)
		`); err != nil {
			return fmt.Errorf("seed zones: %w", err)
		}

		sequences := []string{
			`SELECT setval('public.warehouses_id_seq', (SELECT COALESCE(MAX(id), 1) FROM public.warehouses))`,
			`SELECT setval('public.zones_id_seq', (SELECT COALESCE(MAX(id), 1) FROM public.zones))`,
			`SELECT setval('public.cells_id_seq', (SELECT COALESCE(MAX(id), 1) FROM public.cells))`,
		}
		for _, q := range sequences {
			if _, err := tx.ExecContext(ctx, q); err != nil {
				return fmt.Errorf("sync sequence: %w", err)
			}
		}
	}

	var usersCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM public.users`).Scan(&usersCount); err != nil {
		return fmt.Errorf("check users count: %w", err)
	}

	if usersCount == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash default password: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO public.users (login, name, password_hash)
			VALUES ($1, $2, $3)
		`, defaultLogin, defaultAdminName, string(hash)); err != nil {
			return fmt.Errorf("seed default admin: %w", err)
		}
	}

	return tx.Commit()
}
