package client

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"gitlab.com/mlplabs/auth-service/pkg/client/dto"
)

type HttpClient struct {
	client *resty.Client
}

func NewHttpClient(clientName string, cfg *Config) *HttpClient {
	client := resty.New()

	client.SetBaseURL(cfg.BaseURL).
		SetHeader("Content-Type", "application/json")

	return &HttpClient{
		client: client,
	}
}

func (c *HttpClient) Login(ctx context.Context, request *dto.LoginRequest) (*dto.LoginResponse, error) {
	result := struct {
		Data dto.LoginResponse `json:"data"`
	}{}
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(request).
		SetResult(&result).
		Post("api/v1/auth/login")

	if err != nil {
		return nil, fmt.Errorf("login error: %w", err)
	}
	if resp.IsError() {
	}

	return &result.Data, nil
}

func (c *HttpClient) Refresh(ctx context.Context, request *dto.RefreshRequest) (*dto.RefreshResponse, error) {
	result := struct {
		Data dto.RefreshResponse `json:"data"`
	}{}
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(request).
		SetResult(&result).
		Post("api/v1/auth/refresh")

	if err != nil {
		return nil, fmt.Errorf("refresh error: %w", err)
	}
	if resp.IsError() {
	}

	return &result.Data, nil
}
