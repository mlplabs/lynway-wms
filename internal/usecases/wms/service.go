package wms

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/google/uuid"
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

func (s *Service) ExecuteIdempotent(
	ctx context.Context,
	op func(txRepo repository.PgRepository) (uuid.UUID, error),
) (uuid.UUID, error) {
	idempotencyKey := model.IdempotencyKeyFromContext(ctx)

	var resUUID uuid.UUID

	err := s.whs.WithinTx(ctx, func(txRepo repository.PgRepository) error {
		// Если ключ есть — включаем идемпотентность (lock + проверка/запись inbox)
		if idempotencyKey != "" {
			if err := txRepo.AcquireLock(ctx, "inbox:"+idempotencyKey); err != nil {
				return err
			}

			existing, err := txRepo.GetInboxRecord(ctx, idempotencyKey)
			if err != nil {
				return err
			}
			if existing != nil {
				if err := json.Unmarshal(existing.ResponseBody, &resUUID); err != nil {
					return fmt.Errorf("unmarshal inbox row_id: %w", err)
				}
				return nil
			}
		}

		// Выполняем саму операцию (и с ключом, и без ключа)
		rowID, err := op(txRepo)
		if err != nil {
			return err
		}

		// Сохраняем в operations_inbox только если ключ передан
		if idempotencyKey != "" {
			body, err := json.Marshal(rowID)
			if err != nil {
				return fmt.Errorf("marshal inbox row_id: %w", err)
			}

			if err := txRepo.SaveInboxRecord(ctx, model.InboxRecord{
				Key:          idempotencyKey,
				ResponseCode: 200,
				ResponseBody: body,
			}); err != nil {
				return err
			}
		}

		resUUID = rowID
		return nil
	})

	return resUUID, err
}

func (s *Service) Decode(ctx context.Context, code string) (*model.DecodeResponse, error) {
	typeId, id := s.DecodeQr(code)
	return &model.DecodeResponse{
		EntityTypeId: dto.EntityTypeId(typeId),
		EntityId:     id,
	}, nil
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
