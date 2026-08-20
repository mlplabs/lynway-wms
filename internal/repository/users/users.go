package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mlplabs/lynway-wms/internal/model"
)

type pgUsers struct {
	conn *sql.DB
}

func NewUsersRepository(conn *sql.DB) *pgUsers {
	return &pgUsers{
		conn: conn,
	}
}

func (r *pgUsers) GetById(ctx context.Context, userId int) (*model.User, error) {
	var user model.User

	query := `SELECT id, login FROM users WHERE id = $1`

	err := r.conn.QueryRowContext(ctx, query, userId).Scan(&user.Id, &user.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &user, nil
		}

		return nil, fmt.Errorf("get user by id error: %w", err)
	}

	return &user, nil
}

func (r *pgUsers) Get(ctx context.Context) ([]model.User, error) {
	query := `SELECT id, login FROM users`
	res, err := r.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get users query error: %w", err)
	}
	defer res.Close()

	var users []model.User

	for res.Next() {
		var user model.User

		err = res.Scan(&user.Id, &user.Name)
		if err != nil {
			return nil, fmt.Errorf("get users scan error: %w", err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *pgUsers) Create(ctx context.Context, user *model.User) (*int, error) {
	query := `INSERT INTO users (id, login) VALUES ($1, $2) RETURNING id`

	err := r.conn.QueryRowContext(ctx, query, user.Id, user.Name).Scan(&user.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user.Id, nil
}

func (r *pgUsers) Update(ctx context.Context, user *model.User) error {
	query := `UPDATE users SET login = $1 WHERE id = $2`

	_, err := r.conn.ExecContext(ctx, query, user.Id, user.Name)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
