package wms

import (
	"context"
	"fmt"

	"github.com/mlplabs/lynway-wms/internal/model"
	"github.com/mlplabs/lynway-wms/internal/repository/dbo"
	"github.com/mlplabs/lynway-wms/pkg/dto"
)

func (s *Service) GetUsers(ctx context.Context) ([]dto.User, error) {
	users, err := s.whs.GetUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}

	res := make([]dto.User, 0)

	for _, u := range users {
		res = append(res, dto.User{
			ID:       u.ID,
			Login:    u.Login,
			Name:     u.Name,
			IsActive: u.IsActive,
		})
	}
	return res, nil
}

func (s *Service) CreateUser(ctx context.Context, user *dto.CreateUserRequest) (int, error) {
	res, err := s.whs.CreateUser(ctx, &dbo.CreateUserParams{
		Login:     user.Login,
		Name:      user.Name,
		IsActive:  user.IsActive,
		CreatedBy: model.UserIdPtrFromContext(ctx),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	return res, nil
}

func (s *Service) UpdateUser(ctx context.Context, user *dto.UpdateUserRequest) error {
	err := s.whs.UpdateUser(ctx, &dbo.UpdateUserParams{
		ID:        user.ID,
		Login:     user.Login,
		Name:      user.Name,
		IsActive:  user.IsActive,
		UpdatedBy: model.UserIdPtrFromContext(ctx),
	})
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	return nil
}
