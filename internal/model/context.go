package model

import "context"

type idempotencyKeyCtxKey struct{}

func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyKeyCtxKey{}, key)
}

func IdempotencyKeyFromContext(ctx context.Context) string {
	v, _ := ctx.Value(idempotencyKeyCtxKey{}).(string)
	return v
}
