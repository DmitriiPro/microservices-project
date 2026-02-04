package cache

import (
	"context"
)

type OrderRequestRedis struct {
	UserId    int64
	Product   string
	Quantity  int64
}

type Cache interface {
	Get(ctx context.Context, key string) (OrderRequestRedis, error)
	Set(ctx context.Context, key string, order OrderRequestRedis) error
	Del(ctx context.Context, key string) error
}