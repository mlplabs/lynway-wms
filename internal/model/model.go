package model

import "github.com/mlplabs/lynway-wms/pkg/dto"

type LoginUserRequest struct {
	Email    string
	Password string
}

type LoginUserResponse struct {
	AccessToken  string
	RefreshToken string
}

type RefreshUserRequest struct {
	RefreshToken string
}

type RefreshUserResponse struct {
	AccessToken  string
	RefreshToken string
}

type DecodeRequest struct {
	Code string
}

type DecodeResponse struct {
	EntityTypeId dto.EntityTypeId
	EntityId     int
}

func (m *DecodeResponse) ToDTO() *dto.DecodeResponse {
	result := dto.DecodeResponse{
		EntityTypeId: int(m.EntityTypeId),
	}

	if m.EntityId != 0 {
		mId := m.EntityId
		result.EntityId = &mId
	}
	return &result
}

type Barcode struct {
	Id       int64            `json:"id"`
	Name     string           `json:"name"`
	Type     int              `json:"type"`       // Тип ШК
	OwnerId  int              `json:"owner_id"`   // ID владельца ШК
	OwnerRef dto.EntityTypeId `json:"owner_type"` // Владелец
}
