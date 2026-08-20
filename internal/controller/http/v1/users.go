package v1

import (
	"net/http"

	"github.com/mlplabs/lynway-wms/internal/model"
)

func (h *ServiceHandler) getUsers(w http.ResponseWriter, r *http.Request) (any, error) {
	return nil, nil
}

func (h *ServiceHandler) getUser(w http.ResponseWriter, r *http.Request) (any, error) {
	h.service.GetUserById(r.Context(), 0)
	return nil, nil
}

func (h *ServiceHandler) updateUser(w http.ResponseWriter, r *http.Request) (any, error) {
	h.service.UpdateUser(r.Context(), &model.User{})
	return nil, nil
}
