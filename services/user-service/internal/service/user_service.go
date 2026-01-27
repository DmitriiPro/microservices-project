package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DmitriiPro/user-service/internal/cache"
	"github.com/DmitriiPro/user-service/internal/repository"
	"github.com/redis/go-redis/v9"
)

type UserService struct {
	repo  *repository.UserRepository
	redis *ClientWrapper
}

type ClientWrapper struct {
	Client *redis.Client
}

func NewUserService(repo *repository.UserRepository, rdb *redis.Client) *UserService {
	return &UserService{
		repo:  repo,
		redis: &ClientWrapper{Client: rdb},
	}
}
	
var (
	TTLTimeRedis = (time.Minute * 25)
)

func (s *UserService) GetUserByID(ctx context.Context, id int64) (*repository.User, error) {
	key := fmt.Sprintf("user:%d", id)

	// redis cache
	valueRedis, err := s.redis.Client.Get(cache.Ctx, key).Result()

	if err == nil {
		var user repository.User
		json.Unmarshal([]byte(valueRedis), &user)
		return &user, nil
	}

	// postgres
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// save to redis
	data, _ := json.Marshal(user)
	s.redis.Client.Set(cache.Ctx, key, data, TTLTimeRedis)

	return user, nil
}
