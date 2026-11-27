package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mlplabs/lynway-wms/internal/model"
)

func (r *pgRepository) GetUserByLogin(ctx context.Context, login string) (*model.UserInfo, error) {
	const query = `
		SELECT id, login, name, password_hash, external_sub, is_active
		FROM users
		WHERE login = $1
	`
	var u model.UserInfo
	err := r.conn.QueryRowContext(ctx, query, login).Scan(
		&u.ID,
		&u.Login,
		&u.Name,
		&u.PasswordHash,
		&u.ExternalSub,
		&u.IsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
