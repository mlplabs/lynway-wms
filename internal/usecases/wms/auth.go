package wms

// Часть сервиса, отвечающая за проксирование авторизации

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mlplabs/lynway-wms/internal/model"
	"gitlab.com/mlplabs/auth-service/pkg/client/dto"
	"golang.org/x/crypto/bcrypt"
)

// Login проверяет логин (и пароль, если requirePassword == true) и выпускает JWT.
func (s *Service) Login(ctx context.Context, login, password string) (string, *model.UserInfo, error) {

	user, err := s.whs.GetUserByLogin(ctx, login)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		return "", nil, model.ErrInvalidCredentials
	}
	if !user.IsActive {
		return "", nil, model.ErrUserInactive
	}

	if s.cfg.Auth.RequirePassword {
		if user.PasswordHash == nil || *user.PasswordHash == "" {
			return "", nil, model.ErrInvalidCredentials
		}
		if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
			return "", nil, model.ErrInvalidCredentials
		}
	}

	token, err := s.issueToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("issue jwt: %w", err)
	}

	return token, user, nil
}

func (s *Service) Verify(_ context.Context, rawToken string) (*model.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(rawToken, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.cfg.Auth.JwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, model.ErrInvalidCredentials
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || claims.Subject == "" {
		return nil, model.ErrInvalidCredentials
	}

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil || userID <= 0 {
		return nil, model.ErrInvalidCredentials
	}

	return &model.TokenClaims{
		LocalUserID: userID,
	}, nil
}

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

func (s *Service) issueToken(user *model.UserInfo) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(user.ID),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.Auth.TokenTTL)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.cfg.Auth.JwtSecret)
}
