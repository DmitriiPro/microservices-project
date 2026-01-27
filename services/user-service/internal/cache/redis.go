package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func NewRedis(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}
