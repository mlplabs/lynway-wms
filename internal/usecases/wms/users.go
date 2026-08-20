package wms

import (
	"context"
	"fmt"

	"github.com/mlplabs/lynway-wms/internal/model"
)

type UsersRepository interface {
	Get(ctx context.Context) ([]model.User, error)
	GetById(ctx context.Context, userId int) (*model.User, error)
	Create(ctx context.Context, user *model.User) (*int, error)
	Update(ctx context.Context, user *model.User) error
}

func (s *Service) GetUsers(ctx context.Context) ([]model.User, error) {
	user, err := s.users.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}
	return user, nil
}

func (s *Service) GetUserById(ctx context.Context, userId int) (*model.User, error) {
	user, err := s.users.GetById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}
	return user, nil
}

func (s *Service) CreateUser(ctx context.Context, user *model.User) (*int, error) {
	res, err := s.users.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return res, nil
}

func (s *Service) UpdateUser(ctx context.Context, user *model.User) error {
	err := s.users.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	return nil
}
