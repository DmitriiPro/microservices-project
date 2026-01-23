package handler

import (
	"context"

	userv1 "github.com/DmitriiPro/user-service/internal/pb/user"
)

type UserHandler struct {
	userv1.UnimplementedUserServiceServer
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	// TODO: логика через сервис + репозиторий
	return &userv1.CreateUserResponse{Id: 1}, nil
}

func (h *UserHandler) GetUserByID(ctx context.Context, req *userv1.GetUserByIDRequest) (*userv1.GetUserResponse, error) {
	// TODO: логика через сервис + репозиторий
	return &userv1.GetUserResponse{
		Id:        req.Id,
		Email:     "test@example.com",
		CreatedAt: nil, // пока nil, позже будет timestamppb.Now()
	}, nil
}
