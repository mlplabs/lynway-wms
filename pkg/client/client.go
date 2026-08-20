package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/mlplabs/lynway-wms/pkg/dto"
)

type HttpClient struct {
	client *resty.Client
}

func NewHttpClient(clientName string, cfg *Config) *HttpClient {
	client := resty.New()

	client.SetBaseURL(cfg.BaseURL).
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Access-Token", cfg.Token)

	return &HttpClient{
		client: client,
	}
}

// UpdateAgentInfo обновление информации об агенте и принтерах
func (c *HttpClient) UpdateAgentInfo(ctx context.Context, request *dto.AgentInfoRequest) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(request).
		Post("api/v1/print/agents")

	if err != nil {
		return fmt.Errorf("update agent info error: %v", err)
	}
	if resp.IsError() {
		return fmt.Errorf("update agent info error: %s", errors.New(resp.Status()))
	}

	return nil
}

// GetPrintTask получение задания на печать
func (c *HttpClient) GetPrintTask(ctx context.Context, request *dto.GetPrintTaskRequest) (*dto.GetPrintTaskResponse, error) {
	result := struct {
		Data dto.GetPrintTaskResponse `json:"data"`
	}{}
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get("api/v1/print/tasks")

	if err != nil {
		return nil, fmt.Errorf("get print tasks error: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("get print tasks error: %w", errors.New(resp.Status()))
	}

	return &result.Data, nil
}

// UpdatePrintTaskStatus обновление статуса задания на печать
func (c *HttpClient) UpdatePrintTaskStatus(ctx context.Context, request *dto.UpdatePrintTaskStatusRequest) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(request).
		Post("api/v1/print/tasks/status")

	if err != nil {
		return fmt.Errorf("get print tasks error: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("get print tasks error: %w", errors.New(resp.Status()))
	}

	return nil
}
