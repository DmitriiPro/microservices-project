package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DmitriiPro/user-service/internal/cache"
	"github.com/DmitriiPro/user-service/internal/model"
	"github.com/DmitriiPro/user-service/internal/repository"
	"github.com/redis/go-redis/v9"
)

type UserService interface {
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	CreateUser(ctx context.Context, email string) (int64, error)
}

type userService struct {
	repo  repository.UserRepository
	cache cache.Cache
}

type ClientWrapper struct {
	Client *redis.Client
}

func NewUserService(repo repository.UserRepository, cache cache.Cache) UserService {
	return &userService{repo: repo, cache: cache}
}

func (s *userService) CreateUser(ctx context.Context, email string) (int64, error) {
	 id, err := s.repo.CreateUser(ctx, email)

	 if err != nil {
			return 0, err
	 }
	 
	 key := fmt.Sprintf("user:%d", id)
	 user := model.User{
		 ID:    id,
		 Email: email,
	 }

	 data, _ := json.Marshal(user)
	 s.cache.Set(ctx, key, string(data))

	 return id, nil
}

func (s *userService) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	key := fmt.Sprintf("user:%d", id)

	// redis cache
	valueRedis, err := s.cache.Get(ctx, key)

	if err == nil {
		var user model.User
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
	s.cache.Set(ctx, key, string(data))

	return user, nil
}
