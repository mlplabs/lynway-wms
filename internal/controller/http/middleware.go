package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	pkgErr "github.com/mlplabs/common-go-pkg/pkg/http/errors"
	domainErr "github.com/mlplabs/common-go-pkg/pkg/http/errors/custom"
)

const idempotencyKey = "Idempotency-Key"

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

type ctxKeyIdempotency struct{}

func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, ctxKeyIdempotency{}, key)
}

func RequireIdempotencyKey() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get(idempotencyKey)
			if key == "" {
				pkgErr.SetError(w, r, domainErr.NewBadRequest(fmt.Errorf("missing Idempotency-Key header")))
				return
			}
			ctx := WithIdempotencyKey(r.Context(), key)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func IdempotencyKeyFromContext(ctx context.Context) string {
	key, _ := ctx.Value(ctxKeyIdempotency{}).(string)
	return key
}
