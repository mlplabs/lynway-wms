package v1

import (
	"errors"
	"net/http"

	domainErr "github.com/mlplabs/common-go-pkg/pkg/http/errors/custom"
	"github.com/mlplabs/common-go-pkg/pkg/http/jwtutils"
	"github.com/mlplabs/lynway-wms/pkg/dto"
)

// login
// @Tags Authorization
// @Summary вход в систему
// @Router /auth/login [post]
// @Param request body dto.LoginRequest true "логин/пароль пользователя"
// @Success	200	{object} wrapper.Data{data=dto.LoginResponse} "JWT токены"
func (h *ServiceHandler) login(w http.ResponseWriter, r *http.Request) (any, error) {
	req, err := getLoginUserRequest(r)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	resp, err := h.service.UserLogin(r.Context(), req.Email, req.Password)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return dto.LoginUserResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

// refresh
// @Tags Authorization
// @Summary обновление токенов
// @Router /auth/refresh [post]
// @Param request body dto.RefreshRequest true "refresh токен"
// @Success	200	{object} wrapper.Data{data=dto.RefreshResponse} "JWT токены"
func (h *ServiceHandler) refresh(w http.ResponseWriter, r *http.Request) (any, error) {
	req, err := getRefreshUserRequest(r)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	resp, err := h.service.UserRefresh(r.Context(), req.RefreshToken)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return dto.RefreshUserResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (h *ServiceHandler) me(w http.ResponseWriter, r *http.Request) (any, error) {
	userId, err := jwtutils.GetUserId(r)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	if userId == 1 {
		return nil, domainErr.NewBadRequest(errors.New("user id is required"))
	}
	return nil, nil
}

func (h *ServiceHandler) Logout(w http.ResponseWriter, r *http.Request) (any, error) {
	return nil, nil
}
