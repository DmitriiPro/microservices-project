package config

import (
	"github.com/redis/go-redis/v9"
)

const (
	RedisPort = "localhost:6379"
)

func NewRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     RedisPort,
		Password: "", // no password set
		DB:       0,	// use default DB
	})
	return rdb
}
