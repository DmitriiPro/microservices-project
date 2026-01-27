package main

import (
	"fmt"
	"log"
	"net"

	"github.com/DmitriiPro/user-service/internal/handler"
	userv1 "github.com/DmitriiPro/user-service/internal/pb/user"
	"google.golang.org/grpc"
)

var GrpcServerPORT = 50051

func main() {
	//* starting gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", GrpcServerPORT))

	if err != nil {
		log.Fatalf("failed to listen starting gRPC server: %v", err)
	}

	s := grpc.NewServer()

	userv1.RegisterUserServiceServer(s, handler.NewUserHandler())

	log.Printf("User service is running on %v\n", fmt.Sprintf("localhost:%d", GrpcServerPORT))
	
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
