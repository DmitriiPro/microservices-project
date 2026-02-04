package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Ctx = context.Background()
	TTLTimeRedis = (time.Minute * 25)
)

type redisCache struct {
	client *redis.Client
}

func NewRedis(addr string) Cache {
	rdb :=  redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if err := rdb.Ping(Ctx).Err(); err != nil {
		log.Fatalf("Redis ping failed: %v", err)
	}

	return &redisCache{client: rdb}
}

func (r *redisCache) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *redisCache) Get(ctx context.Context, key string) (OrderRequestRedis, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return OrderRequestRedis{}, err
	}

	var order OrderRequestRedis
	if err := json.Unmarshal([]byte(value), &order); err != nil {
		return OrderRequestRedis{}, err
	}
	
	return order, nil
}

func (r *redisCache) Set(ctx context.Context, key string, order OrderRequestRedis) error {
	return r.client.Set(ctx, key, order, TTLTimeRedis).Err()
}
