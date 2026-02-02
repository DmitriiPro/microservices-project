package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"github.com/DmitriiPro/user-service/internal/cache"
	"github.com/DmitriiPro/user-service/internal/model"
	"github.com/DmitriiPro/user-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService interface {
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	CreateUser(ctx context.Context, email, password string) (int64, error)
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

func (s *userService) CreateUser(ctx context.Context, email, password string) (int64, error) {
	log.Printf("Service: CreateUser called for email: %s", email)
	// Проверка, есть ли пользователь с таким email
	existing, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil && err != repository.ErrNotFoundUser {
		log.Printf("Service: Error checking existing user: %v", err)
		return 0, err
	}

	if existing != nil {
		log.Printf("Service: User with email %s already exists", email)
		return 0, status.Errorf(codes.AlreadyExists, "user with email %s already exists", email)
	}

	log.Printf("Service: Hashing password...")
	// hashed password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Service: Error hashing password: %v", err)
		return 0, fmt.Errorf("error generate password %v ", err)
	}

	log.Printf("Service: Creating user in repository...")
	user, err := s.repo.CreateUser(ctx, email, string(hash))

	if err != nil {
		log.Printf("Service: Repository error: %v", err)
		return 0, err
	}

	key := fmt.Sprintf("user:%d", user.ID)

	data, _ := json.Marshal(user)
	s.cache.Set(ctx, key, string(data))
	log.Printf("Service: User created with ID: %d", user.ID)

	return user.ID, nil
}

func (s *userService) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	key := fmt.Sprintf("user:%d", id)

	// redis cache
	valueRedis, err := s.cache.Get(ctx, key)

	if err == nil {
		var user model.User
		if json.Unmarshal([]byte(valueRedis), &user) == nil {
			log.Printf("userService - GetUserByID: Cache hit for ID %d", id)
			return &user, nil
		}
		log.Printf("userService - GetUserByID: Stale cache for ID %d, deleting", id)
		_ = s.cache.Del(ctx, key) // delete stale cache
	}

	// postgres
	user, err := s.repo.GetUserByID(ctx, id)

	if err != nil {
		log.Printf("userService - GetUserByID: Error from repository for ID %d: %v", id, err)
		if err == repository.ErrNotFoundUser {
			log.Printf("userService - GetUserByID: User with id %d not found", id)
			_ = s.cache.Del(ctx, key) // delete stale cache
			return nil, status.Errorf(codes.NotFound, "user with id %d not found", id)
		}
		return nil, err
	}
	log.Printf("userService - GetUserByID: %v, CreatedAt: %v, Type: %T",
		user, user.CreatedAt, user.CreatedAt)

	// save to redis
	data, err := json.Marshal(user)
	if err != nil {
		log.Printf("userService - GetUserByID: Error marshalling user for cache: %v", err)
	} else {
		if err := s.cache.Set(ctx, key, string(data)); err != nil {
			log.Printf("userService - GetUserByID: Error saving to cache: %v", err)
		}
	}

	return user, nil
}
