package repository

import (
	"context"
	"fmt"

	"github.com/mlplabs/lynway-wms/internal/repository/dbo"
)

func (r *pgRepository) GetUsers(ctx context.Context) ([]dbo.User, error) {
	query := `SELECT id, login, name, is_active FROM users`
	res, err := r.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get users query error: %w", err)
	}
	defer res.Close()

	var users []dbo.User

	for res.Next() {
		user := dbo.User{}

		err = res.Scan(&user.ID, &user.Login, &user.Name, &user.IsActive)
		if err != nil {
			return nil, fmt.Errorf("get users scan error: %w", err)
		}

		users = append(users, user)
	}

	if err = res.Err(); err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}

	return users, nil
}

func (r *pgRepository) CreateUser(ctx context.Context, p *dbo.CreateUserParams) (int, error) {
	var res int
	query := `INSERT INTO users (login, name, is_active, created_by) VALUES ($1, $2, $3, $4) RETURNING id`

	err := r.conn.QueryRowContext(ctx, query, p.Login, p.Name, p.IsActive, p.CreatedBy).Scan(&res)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return res, nil
}

func (r *pgRepository) UpdateUser(ctx context.Context, p *dbo.UpdateUserParams) error {
	query := `UPDATE users SET login = $1, name = $2, is_active = $3, updated_by = $4 WHERE id = $5`

	_, err := r.conn.ExecContext(ctx, query, p.Login, p.Name, p.IsActive, p.UpdatedBy, p.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
