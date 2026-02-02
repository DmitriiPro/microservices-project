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
	"github.com/justinas/alice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
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

	// *серверные keepalive параметры
	kaep := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: true,
	}

	kasp := keepalive.ServerParameters{
		Time:    30 * time.Second,
		Timeout: 10 * time.Second,
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.RecoveryInterceptor()),
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
	)
	userv1.RegisterUserServiceServer(s, handler)

	// ✅ Запускаем gRPC сервер синхронно в горутине

	go func() {
		// Дадим время gRPC серверу запуститься
		log.Printf("gRPC server started on :%s", cfg.GRPCPort)

		if err := s.Serve(grpcLis); err != nil {
			log.Fatalf("failed to serve gRPC server: %v", err)
		}
	}()

	// Ждём готовности gRPC
	time.Sleep(200 * time.Millisecond)
	//* ================= HTTP GATEWAY =================
	// ✅ КРИТИЧНО: Используем контекст, который НЕ отменяется
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // Отменится только при завершении программы

	grpcEndpoint := fmt.Sprintf("localhost:%s", cfg.GRPCPort)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                60 * time.Second,
			Timeout:             20 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.WaitForReady(true), // ✅ Ждать готовности сервера
		),
	}

	// ✅ Устанавливаем соединение явно с блокировкой
	log.Println("Establishing gRPC connection...")
	dialCtx, dialCancel := context.WithTimeout(context.Background(), 5*time.Second)
	conn, err := grpc.DialContext(
		dialCtx,
		grpcEndpoint,
		append(opts, grpc.WithBlock())..., // ✅ Блокируем до установки
	)
	dialCancel()

	if err != nil {
		log.Fatalf("Failed to dial gRPC server: %v", err)
	}

	log.Printf("✅ gRPC connection established, state: %v", conn.GetState())

	mux := runtime.NewServeMux(
		// ✅ Добавьте обработку ошибок
		runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Gateway error: %v", err)
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
		}),
	)

	// err = userv1.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)/
	err = userv1.RegisterUserServiceHandlerClient(ctx, mux, userv1.NewUserServiceClient(conn))
	if err != nil {
		log.Fatalf("failed to register gRPC gateway: %v", err)
	}
	log.Println("✅ HTTP Gateway successfully connected to gRPC")

	// ✅ Прогреваем соединение тестовым запросом
	go func() {
		time.Sleep(100 * time.Millisecond)
		warmupCtx, warmupCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer warmupCancel()

		client := userv1.NewUserServiceClient(conn)
		_, _ = client.GetUserByID(warmupCtx, &userv1.GetUserByIDRequest{Id: 999999})
		log.Println("✅ Connection warmed up")
	}()

	// Создаем цепочку middleware в правильном порядке
	// Порядок важен: сначала recovery, потом logging
	// handlerMiddleware := middleware.LoggingMiddleware(middleware.HTTPRecoveryMiddleware(mux))

	// Создаем цепочку middleware
	chain := alice.New(
		middleware.HTTPRecoveryMiddleware, // Восстановление после паники
		// middleware.DebugMiddleware,        // Отладочное логирование
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
	
	cancel()
	conn.Close() // ✅ Закрываем соединение

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

// ✅ Функция для тестирования и прогрева gRPC соединения
func testGRPCConnection(addr string, opts []grpc.DialOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Добавляем WithBlock для синхронного подключения
	testOpts := append(opts, grpc.WithBlock())

	conn, err := grpc.DialContext(ctx, addr, testOpts...)
	if err != nil {
		return fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	// Проверяем состояние соединения
	state := conn.GetState()
	log.Printf("gRPC connection state: %v", state)

	// Создаём клиент и делаем тестовый запрос (опционально)
	client := userv1.NewUserServiceClient(conn)

	// Пробуем получить несуществующего пользователя для прогрева
	testCtx, testCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer testCancel()

	_, _ = client.GetUserByID(testCtx, &userv1.GetUserByIDRequest{Id: 999999})
	// Игнорируем ошибку - нам важно только установить соединение

	conn.Close()
	log.Println("✅ gRPC connection test successful")
	return nil
}
