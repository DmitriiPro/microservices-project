package cache

import "context"

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
	Del(ctx context.Context, key string) error
}