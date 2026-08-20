package wms

import (
	"context"
	"math"

	"github.com/mlplabs/common-go-pkg/pkg/utils"
	"github.com/mlplabs/lynway-wms/internal/model"
	"github.com/mlplabs/lynway-wms/internal/repository"
	"github.com/mlplabs/lynway-wms/pkg/dto"
	authClient "gitlab.com/mlplabs/auth-service/pkg/client"
)

type AuthService interface {
	Login(ctx context.Context, login string, password string) (*model.LoginUserResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*model.RefreshUserResponse, error)
}

type Service struct {
	auth  *authClient.HttpClient
	users UsersRepository
	print PrintRepository
	whs   repository.PgRepository
}

func NewService(auth *authClient.HttpClient, whsRepository repository.PgRepository, users UsersRepository, print PrintRepository) *Service {
	return &Service{
		auth:  auth,
		whs:   whsRepository,
		users: users,
		print: print,
	}
}

func (s *Service) Decode(ctx context.Context, code string) (*model.DecodeResponse, error) {
	typeId, id := s.DecodeQr(code)
	return &model.DecodeResponse{
		EntityTypeId: dto.EntityTypeId(typeId),
		EntityId:     id,
	}, nil
}

func (s *Service) GetContainerById(ctx context.Context, id int) (*model.Container, error) {
	return s.whs.GetContainerById(ctx, id)
}

func (s *Service) DecodeQr(code string) (int, int) {
	val := utils.NewEnc62("").Decode(code)
	entityTypeId := val % 100
	entityId := math.Floor(float64(val / 100))
	return int(entityTypeId), int(entityId)
}

func (s *Service) EncodeQr(entityTypeId int, entityId int) string {
	var val int64
	val = int64(entityId*100 + entityTypeId)
	return utils.NewEnc62("").Encode(val)
}

func (s *Service) CreateContainer(ctx context.Context, name *string, containerType *int) (*model.Container, error) {
	return s.whs.CreateContainer(ctx, name, containerType)
}
