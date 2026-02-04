package service

import (
	"context"
	"errors"
	"fmt"
	// "strconv"

	"github.com/DmitriiPro/order-service/internal/cache"
	"github.com/DmitriiPro/order-service/internal/client"
	"github.com/DmitriiPro/order-service/internal/model"
	"github.com/DmitriiPro/order-service/internal/repository"
	"github.com/redis/go-redis/v9"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userId int64, product string, quantity int64) (*model.Order, error)
}

type orderService struct {
	repo       repository.OrderRepository
	cache      cache.Cache
	clientUser *client.UserClient
}

type ClientWrapper struct {
	Client *redis.Client
}

func NewOrderService(repo repository.OrderRepository, cache cache.Cache, addr string) OrderService {
	return &orderService{repo: repo, cache: cache, clientUser: client.NewUserClient(addr)}
}

var UserNotFoundError = errors.New("user not found")

func (s *orderService) CreateOrder(ctx context.Context, userId int64, product string, quantity int64) (*model.Order, error) {

	user, err := s.clientUser.GetUser(ctx, userId)
	if err != nil {
		if err.Error() == "user not found" {
			return nil, UserNotFoundError
		}
		return nil, err
	}


	order, err := s.repo.CreateOrder(ctx, user.Id, product, quantity)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("order:%d", order.Id)

	_ = s.cache.Set(ctx, key, cache.OrderRequestRedis{
		UserId:   order.UserId,
		Product:  order.Product,
		Quantity: order.Quantity,
	})

	return order, nil
}
