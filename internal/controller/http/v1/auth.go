package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	domainErr "github.com/mlplabs/common-go-pkg/pkg/http/errors/custom"
	"github.com/mlplabs/common-go-pkg/pkg/http/jwtutils"
	"github.com/mlplabs/lynway-wms/internal/model"
	"github.com/mlplabs/lynway-wms/pkg/dto"
)

// login godoc
//
//	@Summary		Локальная авторизация пользователя
//	@Description	Выполняет вход по логину/табельному номеру (и паролю, если require_password=true) в режиме auth.mode=local и возвращает JWT-токен.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginRequest	true	"Учетные данные пользователя"
//	@Success		200		{object}	object{data=LoginResponse}
//	@Failure		400		{object}	object{error=object{code=string,message=string}}	"Некорректный запрос или локальный вход отключен"
//	@Failure		401		{object}	object{error=object{code=string,message=string}}	"Неверный логин или пароль"
//	@Failure		403		{object}	object{error=object{code=string,message=string}}	"Учетная запись заблокирована"
//	@Router			/api/v1/auth/login [post]
func (h *ServiceHandler) login(w http.ResponseWriter, r *http.Request) (any, error) {
	authMode := "local"
	requiredPassword := false
	if authMode != "local" {
		return nil, domainErr.NewBadRequest(fmt.Errorf("local login is disabled in %q auth mode", authMode))
	}

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body: %w", err))
	}

	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" {
		return nil, domainErr.NewBadRequest(fmt.Errorf("login is required"))
	}
	if requiredPassword && req.Password == "" {
		return nil, domainErr.NewBadRequest(fmt.Errorf("password is required"))
	}

	token, user, err := h.service.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		return nil, mapAuthError(err)
	}

	return &dto.LoginResponse{
		Token: token,
		User: dto.LoginUserDTO{
			ID:   user.ID,
			Name: user.Name,
		},
	}, nil
}

func mapAuthError(err error) error {
	switch {
	case errors.Is(err, model.ErrInvalidCredentials):
		return domainErr.NewUnauthorized(err)
	case errors.Is(err, model.ErrUserInactive):
		return domainErr.NewForbidden(err)
	default:
		return err
	}
}

// login_deprecated
// @Tags Authorization
// @Summary вход в систему
// @Router /auth/login_deprecated [post]
// @Param request body dto.LoginRequest true "логин/пароль пользователя"
// @Success	200	{object} wrapper.Data{data=dto.LoginResponse} "JWT токены"
func (h *ServiceHandler) login_deprecated(w http.ResponseWriter, r *http.Request) (any, error) {
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

// refresh_deprecated
// @Tags Authorization
// @Summary обновление токенов
// @Router /auth/refresh_deprecated [post]
// @Param request body dto.RefreshRequest true "refresh_deprecated токен"
// @Success	200	{object} wrapper.Data{data=dto.RefreshResponse} "JWT токены"
func (h *ServiceHandler) refresh_deprecated(w http.ResponseWriter, r *http.Request) (any, error) {
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
