package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DmitriiPro/user-service/internal/cache"
	"github.com/DmitriiPro/user-service/internal/config"
	"github.com/DmitriiPro/user-service/internal/db"
	"github.com/DmitriiPro/user-service/internal/handler"
	"github.com/DmitriiPro/user-service/internal/middleware"
	userv1 "github.com/DmitriiPro/user-service/internal/pb/user"
	"github.com/DmitriiPro/user-service/internal/repository"
	"github.com/DmitriiPro/user-service/internal/service"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/justinas/alice"
)

var HTTP_PORT = ":8081"

func main() {
	cfg := config.Load()

	// Применяем миграции
	if err := db.RunMigrations(cfg.PostgresDSN); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// db postgres
	dbConn := db.NewPostgres(cfg.PostgresDSN)
	defer dbConn.Close()

	// db redis
	redisClient := cache.NewRedis(cfg.RedisAddr)

	repo := repository.NewUserRepository(dbConn)
	service := service.NewUserService(repo, redisClient)
	handler := handler.NewUserHandler(service)

	//* ================= gRPC SERVER =================
	grpcLis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", cfg.GRPCPort))

	if err != nil {
		log.Fatalf("failed to listen starting gRPC server: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(middleware.RecoveryInterceptor()))
	userv1.RegisterUserServiceServer(s, handler)

	go func() {
		// Дадим время gRPC серверу запуститься
		time.Sleep(1 * time.Second)
		log.Printf("gRPC server started on :%s", cfg.GRPCPort)

		if err := s.Serve(grpcLis); err != nil {
			log.Fatalf("failed to serve gRPC server: %v", err)
		}
	}()

	//* ================= HTTP GATEWAY =================
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	// wrappedMux := middleware.HTTPRecoveryMiddleware(mux)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = userv1.RegisterUserServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%s", cfg.GRPCPort), opts)
	if err != nil {
		log.Fatalf("failed to register gRPC gateway: %v", err)
	}

	// Создаем цепочку middleware в правильном порядке
	// Порядок важен: сначала recovery, потом logging
	// handlerMiddleware := middleware.LoggingMiddleware(middleware.HTTPRecoveryMiddleware(mux))

	// Создаем цепочку middleware
	chain := alice.New(
		middleware.HTTPRecoveryMiddleware, // Восстановление после паники
		middleware.LoggingMiddleware,      // Логирование
	).Then(mux)

	httpServer := &http.Server{
		Addr:         HTTP_PORT,
		Handler:      chain,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("HTTP Gateway started on %s\n", HTTP_PORT)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to serve HTTP: %v", err)
		}
	}()

	//! ================= GRACEFUL SHUTDOWN =================
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down servers...")

	//! HTTP shutdown
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()

	if err := httpServer.Shutdown(ctxShutdown); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	// gRPC shutdown
	s.GracefulStop()

	log.Println("Servers stopped")

}
