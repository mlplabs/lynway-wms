package wms

// Часть сервиса, отвечающая за проксирование авторизации

import (
	"context"
	"fmt"

	"github.com/mlplabs/lynway-wms/internal/model"
	"gitlab.com/mlplabs/auth-service/pkg/client/dto"
)

func (s *Service) UserLogin(ctx context.Context, login, password string) (*model.LoginUserResponse, error) {
	res, err := s.auth.Login(ctx, &dto.LoginRequest{
		Email:    login,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("user login error: %w", err)
	}

	return &model.LoginUserResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	}, nil
}

func (s *Service) UserRefresh(ctx context.Context, refreshToken string) (*model.RefreshUserResponse, error) {
	res, err := s.auth.Refresh(ctx, &dto.RefreshRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return nil, fmt.Errorf("refresh user error: %w", err)
	}
	return &model.RefreshUserResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	}, nil
}
