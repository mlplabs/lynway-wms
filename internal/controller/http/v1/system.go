package v1

import (
	"net/http"

	"github.com/mlplabs/lynway-wms/pkg/dto"
)

// getSystemInfo godoc
//
//	@Summary		Получить системную информацию и настройки авторизации
//	@Description	Возвращает версию API, текущий режим авторизации (none, local, external) и необходимость ввода пароля при входе.
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	object{data=SystemInfoResponse}
//	@Router			/api/v1/system/info [get]
func (h *ServiceHandler) getSystemInfo(w http.ResponseWriter, r *http.Request) (any, error) {

	si := h.service.GetSystemInfo()

	return &dto.SystemInfoResponse{
		Version:         si.Version,
		AuthMode:        si.AuthMode,
		RequirePassword: si.RequirePassword,
	}, nil
}
