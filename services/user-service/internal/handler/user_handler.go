package handler

import (
	"context"
	"log"

	userv1 "github.com/DmitriiPro/user-service/internal/pb/user"
	"github.com/DmitriiPro/user-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {

	select {
	case <-ctx.Done():
		log.Printf("Context cancelled before processing: %v", ctx.Err())
		return nil, status.Error(codes.Canceled, "request cancelled")
	default:
		log.Printf("CreateUser called with email: %s", req.Email)

		if err := req.Validate(); err != nil {
			log.Printf("Validation failed: %v", err)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		log.Printf("Creating user with email: %s", req.Email)
		id, err := h.svc.CreateUser(ctx, req.Email, req.Password)

		if err != nil {
			log.Printf("CreateUser service error: %v", err)
			return nil, err
		}

		log.Printf("User created successfully with ID: %d", id)
		return &userv1.CreateUserResponse{Id: id}, nil
	}

}

func (h *UserHandler) GetUserByID(ctx context.Context, req *userv1.GetUserByIDRequest) (*userv1.GetUserResponse, error) {

	// Валидация ID
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID must be positive")
	}

	log.Printf("UserHandler - GetUserByID: Request for ID %d", req.Id)

	user, err := h.svc.GetUserByID(ctx, req.Id)
	if err != nil {
		log.Printf("UserHandler - GetUserByID: Service error for ID %d: %v", req.Id, err)
		return nil, err
	}

	log.Printf("UserHandler - GetUserByID: Success for ID %d", req.Id)

	return &userv1.GetUserResponse{
		Id:        user.ID,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, nil
}
