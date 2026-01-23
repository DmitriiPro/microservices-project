package main

import (
	"log"
	"net"

	"github.com/DmitriiPro/user-service/internal/handler"
	userv1 "github.com/DmitriiPro/user-service/internal/pb/user"
	"google.golang.org/grpc"
)

var PORT = ":50051"

func main() {
	lis, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	userv1.RegisterUserServiceServer(s, handler.NewUserHandler())

	log.Printf("User service is running on %v\n", PORT)
	
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}