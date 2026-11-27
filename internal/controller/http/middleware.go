package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	pkgErr "github.com/mlplabs/common-go-pkg/pkg/http/errors"
	domainErr "github.com/mlplabs/common-go-pkg/pkg/http/errors/custom"
	"github.com/mlplabs/lynway-wms/internal/config"
	"github.com/mlplabs/lynway-wms/internal/model"
)

const (
	idempotencyKey   = "Idempotency-Key"
	authorizationKey = "Authorization"
)

type AccessTokenKey struct{}

// Language упрощенный вариант определения запрашиваемого языка
// использует первый из списка
// не определяет вес
// https://developer.mozilla.org/ru/docs/Web/HTTP/Reference/Headers/Accept-Language
func Language(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		langVal := "ru-RU"
		al := r.Header.Get("Accept-Language")

		split := strings.Split(al, ",")
		if len(split) > 0 {
			switch split[0] {
			case "en-US", "en-GB", "en":
				langVal = "en-US"
			case "ru", "ru-RU":
				langVal = "ru-RU"
			}
		}
		ctx := context.WithValue(r.Context(), "language", langVal)
		h.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

func APIKeyValidate(apiKeys map[string]struct{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-Access-Token")
			if key == "" {
				pkgErr.SetError(w, r, domainErr.NewBadRequest(errors.New("missing X-Access-Token")))
				return
			}

			_, ok := apiKeys[key]
			if !ok {
				pkgErr.SetError(w, r, domainErr.NewUnauthorized(errors.New("invalid X-Access-Token")))
				return
			}

			ctx := context.WithValue(r.Context(), AccessTokenKey{}, key)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireIdempotencyKey() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get(idempotencyKey)
			if key == "" {
				pkgErr.SetError(w, r, domainErr.NewBadRequest(fmt.Errorf("missing %s header", idempotencyKey)))
				return
			}

			ctx := model.WithIdempotencyKey(r.Context(), key)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type AuthMode string

const (
	AuthModeNone     AuthMode = "none"
	AuthModeLocal    AuthMode = "local"
	AuthModeExternal AuthMode = "external"
)

// TokenClaims — унифицированный результат валидации как локального, так и внешнего JWT.

// TokenVerifier проверяет подпись и срок действия JWT (реализуется при подключении JWT-библиотеки).
type TokenVerifier interface {
	Verify(ctx context.Context, rawToken string) (*model.TokenClaims, error)
}

// ExternalUserResolver находит или создает локального пользователя по external_sub и возвращает users.id.
type ExternalUserResolver interface {
	ResolveExternalUser(ctx context.Context, sub, name string) (int, error)
}

func Auth(cfg *config.Config, Verifier TokenVerifier, Resolver ExternalUserResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Zero-Config режим: пропускаем всех с UserID = 0 (NULL в БД)
			if cfg.IsAuthDisable() {
				next.ServeHTTP(w, r)
				return
			}

			// 2. Извлекаем Bearer-токен из заголовка Authorization
			rawToken, err := extractBearerToken(r.Header.Get(authorizationKey))
			if err != nil {
				pkgErr.SetError(w, r, domainErr.NewUnauthorized(fmt.Errorf("invalid authorization header: %w", err)))
				return
			}

			if Verifier == nil {
				http.Error(w, "auth verifier is not configured", http.StatusNotImplemented)
				return
			}

			// 3. Валидируем JWT
			claims, err := Verifier.Verify(r.Context(), rawToken)
			if err != nil {
				pkgErr.SetError(w, r, domainErr.NewUnauthorized(fmt.Errorf("invalid or expired token")))
				return
			}

			// 4. Определяем локальный users.id (int)
			var userID int
			switch cfg.Mode {
			case "local":
				userID = claims.LocalUserID
			case "external":
				if Resolver == nil {
					pkgErr.SetError(w, r, domainErr.NewServerError(fmt.Errorf("external user resolver is not configured")))

					return
				}
				userID, err = Resolver.ResolveExternalUser(r.Context(), claims.ExternalSub, claims.Name)
				if err != nil {
					pkgErr.SetError(w, r, domainErr.NewServerError(fmt.Errorf("failed to resolve user: %w", err)))

					return
				}
			default:
				pkgErr.SetError(w, r, domainErr.NewServerError(fmt.Errorf("unknown auth mode: %s", cfg.Mode)))

				return
			}

			if userID <= 0 {
				pkgErr.SetError(w, r, domainErr.NewUnauthorized(fmt.Errorf("invalid user identity in token: %w", err)))

				return
			}

			// 5. Кладем единый int UserID в контекст
			ctx := model.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(header string) (string, error) {
	if header == "" {
		return "", fmt.Errorf("missing authorization header")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return strings.TrimSpace(parts[1]), nil
}
