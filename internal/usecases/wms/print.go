package wms

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	"github.com/mlplabs/lynway-wms/internal/model"
	"github.com/mlplabs/lynway-wms/pkg/dto"
)

type PrintRepository interface {
	GetAgent(ctx context.Context, token string) (*model.Agent, error)
	UpdateAgentInfo(ctx context.Context, agentInfo *model.AgentInfo) error
	UpdateAgentToken(ctx context.Context, agentId int, token string) error
	GetPrintTaskWithLock(ctx context.Context, agentId int) (int, error)
	GetPrintTaskById(ctx context.Context, taskId int) (*model.PrintQueueTask, error)
	UpdatePrintTaskStatus(ctx context.Context, status *model.PrintTaskStatus) error
}

func (s *Service) GetAgentByToken(ctx context.Context, token string) (*model.Agent, error) {
	return s.print.GetAgent(ctx, token)
}
func (s *Service) GetPrintTask(ctx context.Context, agentId int) (*model.PrintTask, error) {
	// Получаем очередную задачу с блокировкой
	taskId, err := s.print.GetPrintTaskWithLock(ctx, agentId)
	if err != nil {
		return nil, err
	}
	// Запрашиваем данные
	taskQueue, err := s.print.GetPrintTaskById(ctx, taskId)
	if err != nil {
		return nil, err
	}

	t := model.PrintTask{
		Labels: make(map[string]string),
	}

	if taskQueue.PrintType == dto.PrintTaskTypeLabel {
		// Формируем данные для печати типовой (пока) этикетки. Зависит от print_type
		switch taskQueue.EntityTypeId {
		case dto.EntityTypeIdProduct:
			t.Labels["title"] = "Title"
		case dto.EntityTypeIdContainer:
			t.Labels["title"] = "Title"
		case dto.EntityTypeIdTypeCell:
			t.Labels["title"] = "Title"
		//s.whs.GetCellById(ctx, taskQueue.EntityId)
		default:
			// попросили чтото неизвестное
		}

		t.Id = taskId                           // id задачи
		t.Type = taskQueue.PrintType            // тип печати
		t.PrinterName = taskQueue.Printer.Name  // имя принтера
		t.EntityTypeId = taskQueue.EntityTypeId // тип данных
		t.EntityId = taskQueue.EntityId         // id данных
		t.Data = s.EncodeQr(int(taskQueue.EntityTypeId), taskQueue.EntityId)
	}

	return &t, nil
}
func (s *Service) UpdatePrintTaskStatus(ctx context.Context, status *model.PrintTaskStatus) error {
	return s.print.UpdatePrintTaskStatus(ctx, status)
}
func (s *Service) UpdateAgentInfo(ctx context.Context, agentInfo *model.AgentInfo) error {
	return s.print.UpdateAgentInfo(ctx, agentInfo)
}

// GenerateAgentToken обновляет токен у выбранного агента
func (s *Service) GenerateAgentToken(ctx context.Context, agentId int) (string, error) {
	uid := uuid.New().String()
	token := base64.StdEncoding.EncodeToString([]byte(uid))

	err := s.print.UpdateAgentToken(ctx, agentId, token)
	if err != nil {
		return "", fmt.Errorf("failed to update agent token, agent %d, err: %w", agentId, err)
	}
	return token, nil
}
