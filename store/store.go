package store

import "context"

type Store interface {
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttlSeconds int64) error
	Get(ctx context.Context, key string) (int64, error)
	Set(ctx context.Context, key string, value int64, ttlSeconds int64) error
	Del(ctx context.Context, key string) error
}
