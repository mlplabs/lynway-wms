package v1

import (
	"net/http"

	"github.com/mlplabs/lynway-wms/pkg/dto"
)

// updateAgentInfo
// @Tags Printing
// @Summary обновление данных агента печати
// @Description агент обновляет данные о себе на сервере и обновляет список принтеров
// @Router /print/agents/sync [post]
// @Param request body dto.AgentInfoRequest true "данные агента"
// @Success	200
func (h *ServiceHandler) updateAgentInfo(w http.ResponseWriter, r *http.Request) error {
	agentToken, err := getAccessToken(r)
	if err != nil {
		return err
	}

	agentInfo, err := getAgentInfoRequest(r)
	if err != nil {
		return err
	}

	agentInfo.Token = agentToken
	return h.service.UpdateAgentInfo(r.Context(), agentInfo)
}

// getPrintTask
// @Tags Printing
// @Summary получение задания на печать
// @Description агент получает задание из очереди.
// @Router /print/tasks [get]
// @Success	200 {object} wrapper.Data{data=dto.GetPrintTaskResponse} "задание для печати"
func (h *ServiceHandler) getPrintTask(w http.ResponseWriter, r *http.Request) (any, error) {
	ctx := r.Context()

	agentToken, err := getAccessToken(r)
	if err != nil {
		return nil, err
	}

	agent, err := h.service.GetAgentByToken(ctx, agentToken)
	if err != nil {
		return nil, err
	}

	task, err := h.service.GetPrintTask(ctx, agent.Id)
	if err != nil {
		return nil, err
	}
	return dto.GetPrintTaskResponse{
		Id:           task.Id,
		Type:         task.Type,
		PrinterName:  task.PrinterName,
		EntityId:     task.EntityId,
		EntityTypeId: task.EntityTypeId,
		Data:         task.Data,
		Labels:       task.Labels,
		ToFile:       task.ToFile,
	}, err
}

// updatePrintTaskStatus
// @Tags Printing
// @Summary обновление статуса задания печати на сервере
// @Description после выполнения печати, агент обновляет статус задания
// @Router /print/tasks/status [post]
// @Param request body dto.UpdatePrintTaskStatusRequest true "новый статус задания"
// @Success	200
func (h *ServiceHandler) updatePrintTaskStatus(w http.ResponseWriter, r *http.Request) error {
	status, err := getUpdateTaskStatusRequest(r)
	if err != nil {
		return err
	}
	return h.service.UpdatePrintTaskStatus(r.Context(), status)
}
