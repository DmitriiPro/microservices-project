package handler

import (
	"context"

	userv1 "github.com/DmitriiPro/user-service/internal/pb/user"
	"github.com/DmitriiPro/user-service/internal/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	// TODO: логика через сервис + репозиторий
	return &userv1.CreateUserResponse{Id: 1}, nil
}

func (h *UserHandler) GetUserByID(ctx context.Context, req *userv1.GetUserByIDRequest) (*userv1.GetUserResponse, error) {

	user, err := h.svc.GetUserByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}


	return &userv1.GetUserResponse{
		Id:        user.ID,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt), // пока nil, позже будет timestamppb.Now()
	}, nil
}
