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

type User struct {
	Id    int
	Name  string
	Email string
}

type Product struct {
	Id           int64         `json:"id"`
	Name         string        `json:"name"`
	ItemNumber   string        `json:"item_number"`
	Manufacturer *Manufacturer `json:"manufacturer"`
	Tags         string        `json:"tags"`
	Barcodes     []Barcode     `json:"barcodes"`
}

type Barcode struct {
	Id       int64            `json:"id"`
	Name     string           `json:"name"`
	Type     int              `json:"type"`       // Тип ШК
	OwnerId  int              `json:"owner_id"`   // ID владельца ШК
	OwnerRef dto.EntityTypeId `json:"owner_type"` // Владелец
}

type RemainingProduct struct {
	Product
	ContainerId int64 `json:"container_id"`
	Quantity    int   `json:"quantity"`
}

type Container struct {
	Id    int     // ID
	Pid   int     // Parent ID
	Rid   int     // Root ID
	Depth int     // Глубина
	Type  int     // Тип контейнера
	Name  *string // Название контейнера
}
