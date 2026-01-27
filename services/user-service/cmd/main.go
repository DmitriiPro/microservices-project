package main

import (
	"fmt"
	"log"
	"net"

	"github.com/DmitriiPro/user-service/internal/cache"
	"github.com/DmitriiPro/user-service/internal/config"
	"github.com/DmitriiPro/user-service/internal/db"
	"github.com/DmitriiPro/user-service/internal/handler"
	userv1 "github.com/DmitriiPro/user-service/internal/pb/user"
	"github.com/DmitriiPro/user-service/internal/repository"
	"github.com/DmitriiPro/user-service/internal/service"
	"google.golang.org/grpc"
)

// var GrpcServerPORT = 50051

func main() {
	cfg := config.Load()

	// db postgres
	dbConn := db.NewPostgres(cfg.PostgresDSN)
	defer dbConn.Close()

	// db redis
	redisClient := cache.NewRedis(cfg.RedisAddr)

	repo := repository.NewUserRepository(dbConn)
	service := service.NewUserService(repo, redisClient)
	handler := handler.NewUserHandler(service)

	//* starting gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", cfg.GRPCPort))

	if err != nil {
		log.Fatalf("failed to listen starting gRPC server: %v", err)
	}

	s := grpc.NewServer()

	userv1.RegisterUserServiceServer(s, handler)

	log.Printf("User service is running on %v\n", fmt.Sprintf("localhost:%d", cfg.GRPCPort))

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
