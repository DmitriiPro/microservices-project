package handler

import (
	"context"
	"errors"
	"fmt"
	"log"

	orderv1 "github.com/DmitriiPro/order-service/internal/pb/order"
	"github.com/DmitriiPro/order-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderHandler struct {
	orderv1.UnimplementedOrderServiceServer
	svc service.OrderService
}

func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.OrderResponse, error) {

	if err := req.Validate(); err != nil {
		log.Printf("Validation failed: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	fmt.Println("req.UserId ", req.UserId)
	order, err := h.svc.CreateOrder(ctx, req.UserId, req.Product, req.Quantity)
	if err != nil {
		log.Printf("Service error: %v", err)
		// if err == service.UserNotFoundError {
		// 	return nil, status.Error(codes.NotFound, "user not found")
		// }
		if errors.Is(err, service.UserNotFoundError) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		
		return nil, status.Error(codes.Internal, "failed to create order")
	}

	return &orderv1.OrderResponse{Id: order.Id, UserId: order.UserId, Product: order.Product, Quantity: order.Quantity, CreatedAt: timestamppb.New(order.CreatedAt)}, nil
}
