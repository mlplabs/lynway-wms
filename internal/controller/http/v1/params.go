package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	domainErr "github.com/mlplabs/common-go-pkg/pkg/http/errors/custom"
	httpInternal "github.com/mlplabs/lynway-wms/internal/controller/http"
	"github.com/mlplabs/lynway-wms/internal/model"
	"github.com/mlplabs/lynway-wms/pkg/dto"
)

const (
	idParam             = "id"
	queryParam          = "q"
	suggestLimitParam   = "l"
	defaultLimitSuggest = 10
)

func getQueryParam(r *http.Request) string {
	return r.URL.Query().Get(queryParam)
}

func getLimitSuggestParam(r *http.Request) (int, error) {
	limitParam := r.URL.Query().Get(suggestLimitParam)
	if limitParam == "" {
		return defaultLimitSuggest, nil
	}

	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		return 0, domainErr.NewBadRequest(fmt.Errorf("%s parameter is invalid: %w", suggestLimitParam, err))
	}

	return limit, nil
}

func getLoginUserRequest(r *http.Request) (*model.LoginUserRequest, error) {
	req := dto.LoginUserRequest{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body, %w", err))
	}

	return &model.LoginUserRequest{
		Email:    req.Email,
		Password: req.Password,
	}, nil
}

func getRefreshUserRequest(r *http.Request) (*model.RefreshUserRequest, error) {
	req := dto.RefreshUserRequest{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body, %w", err))
	}

	return &model.RefreshUserRequest{
		RefreshToken: req.RefreshToken,
	}, nil
}

func getDecodeRequest(r *http.Request) (*dto.DecodeRequest, error) {
	data := dto.DecodeRequest{}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return nil, domainErr.NewBadRequest(fmt.Errorf("failed to decode request body: %w", err))
	}

	return &data, nil
}

func getIdParam(r *http.Request) (int, error) {
	param := chi.URLParam(r, idParam)
	if param == "" {
		return 0, domainErr.NewBadRequest(fmt.Errorf("%s parameter is required", idParam))
	}

	id, err := strconv.Atoi(param)
	if err != nil {
		return 0, domainErr.NewBadRequest(fmt.Errorf("%s parameter is invalid: %w", idParam, err))
	}

	return id, nil
}

func getAgentInfoRequest(r *http.Request) (*model.AgentInfo, error) {
	req := dto.AgentInfoRequest{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body, %w", err))
	}

	res := model.AgentInfo{}
	return res.FromDTO(&req), nil
}

func getUpdateTaskStatusRequest(r *http.Request) (*model.PrintTaskStatus, error) {
	req := dto.UpdatePrintTaskStatusRequest{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err != io.EOF {
		return nil, domainErr.NewBadRequest(fmt.Errorf("invalid request body, %w", err))
	}

	res := model.PrintTaskStatus{}
	return res.FromDTO(&req), nil
}

func getAccessToken(r *http.Request) (string, error) {
	val := r.Context().Value(httpInternal.AccessTokenKey{})

	tokenStr, ok := val.(string)
	if !ok {
		return "", domainErr.NewUnauthorized(errors.New("missing access token"))
	}
	return tokenStr, nil
}
