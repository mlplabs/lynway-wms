package model

import "context"

type userIdCtxKey struct{}
type idempotencyKeyCtxKey struct{}

func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIdCtxKey{}, userID)
}

func UserIdFromContext(ctx context.Context) int {
	v, _ := ctx.Value(userIdCtxKey{}).(int)
	return v
}

func UserIdPtrFromContext(ctx context.Context) *int {
	id := UserIdFromContext(ctx)
	if id <= 0 {
		return nil
	}
	return &id
}
func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyKeyCtxKey{}, key)
}

func IdempotencyKeyFromContext(ctx context.Context) string {
	v, _ := ctx.Value(idempotencyKeyCtxKey{}).(string)
	return v
}
