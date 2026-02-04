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

	"github.com/DmitriiPro/order-service/internal/cache"
	"github.com/DmitriiPro/order-service/internal/config"
	"github.com/DmitriiPro/order-service/internal/db"
	"github.com/DmitriiPro/order-service/internal/handler"
	"github.com/DmitriiPro/order-service/internal/middleware"
	orderv1 "github.com/DmitriiPro/order-service/internal/pb/order"
	"github.com/DmitriiPro/order-service/internal/repository"
	"github.com/DmitriiPro/order-service/internal/service"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/justinas/alice"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

var HTTP_PORT = ":8083"
// var SWAGGER_PORT = ":8084"

func main() {
	cfg := config.Load()

	dbConn := db.NewPostgres(cfg.PostgresDSN)
	defer dbConn.Close()

	// db redis
	redisClient := cache.NewRedis(cfg.RedisAddr)
	// client
	// clientUser := client.NewUserClient(cfg.UserSvcHostApi)

	repo := repository.NewOrderRepository(dbConn)
	service := service.NewOrderService(repo, redisClient, cfg.UserSvcHostApi)
	handler := handler.NewOrderHandler(service)

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

	orderv1.RegisterOrderServiceServer(s, handler)

	// ✅ Запускаем gRPC сервер синхронно в горутине
	go func() {
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

	mux := runtime.NewServeMux()

	mux.HandlePath("GET", "/swagger.json", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		// http.ServeFile(w, r, "../swagger/user.swagger.json")
		http.ServeFile(w, r, "./swagger/order/order.swagger.json")
	})

	// gwmux.HandlePath("/swagger/", httpSwagger.Handler(
	// 	httpSwagger.URL("http://localhost"+HTTP_PORT+"/swagger.json"),
	// ))
	// err = gwmux.HandlePath("GET", "/swagger.json", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	// 	http.ServeFile(w, r, "./swagger/order/order.swagger.json")
	// })
	// if err != nil {
	// 	log.Fatalf("failed to register swagger json: %v", err)
	// }

	//! ===== Swagger JSON endpoint =====

	// err = userv1.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	// err = userv1.RegisterUserServiceHandlerClient(ctx, mux, userv1.NewUserServiceClient(conn))
err = orderv1.RegisterOrderServiceHandlerClient(ctx, mux, orderv1.NewOrderServiceClient(conn))
	// err = orderv1.RegisterOrderServiceHandlerFromEndpoint(ctx, gwmux, grpcEndpoint, opts)
	// err = userv1.RegisterUserServiceHandlerClient(ctx, mux, userv1.NewUserServiceClient(conn))
	if err != nil {
		log.Fatalf("failed to register gRPC gateway: %v", err)
	}
	log.Println("✅ HTTP Gateway successfully connected to gRPC")

	// ✅ Прогреваем соединение тестовым запросом
	// go func() {
	// 	time.Sleep(100 * time.Millisecond)
	// 	warmupCtx, warmupCancel := context.WithTimeout(context.Background(), 2*time.Second)
	// 	defer warmupCancel()

	// 	client := userv1.NewUserServiceClient(conn)
	// 	_, _ = client.GetUserByID(warmupCtx, &userv1.GetUserByIDRequest{Id: 999999})
	// 	log.Println("✅ Connection warmed up")
	// }()

	// rootMux := http.NewServeMux()
	// Подключаем Swagger UI на путь /swagger/
	// rootMux.Handle("/swagger/", httpSwagger.Handler(
	// 	httpSwagger.URL("/swagger.json"),
	// ))


	// 1. Добавляем Swagger JSON
	// rootMux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		// http.ServeFile(w, r, "./swagger/order/order.swagger.json")
	// })
	
	// 2. Добавляем Swagger UI
	// swaggerHandler := httpSwagger.Handler(
		// httpSwagger.URL("/swagger.json"), // ✅ Указываем относительный путь
	// )
	// rootMux.Handle("/swagger/", swaggerHandler)

	// rootMux.Handle("/", gwmux)
	// Создаем цепочку middleware
	chain := alice.New(
		middleware.CORSMiddleware,         // CORS
		middleware.HTTPRecoveryMiddleware, // Восстановление после паники
		middleware.LoggingMiddleware,      // Логирование
	).Then(mux)

// 	rootMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//     if r.URL.Path == "/" || !strings.HasPrefix(r.URL.Path, "/swagger/") {
//         chain.ServeHTTP(w, r)
//         return
//     }
//     http.NotFound(w, r)
// })

	

	httpServer := &http.Server{
		Addr:         ":8084",
		Handler:      chain,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// go func() {
	// 	// log.Printf("HTTP Gateway started on %s\n", HTTP_PORT)
	// 	// if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 	// 	log.Fatalf("failed to serve HTTP: %v", err)
	// 	// }

	// 	log.Printf("HTTP Gateway & Swagger started on %s\n", HTTP_PORT)
	// 	log.Printf("Swagger UI available at: http://localhost%s/swagger/index.html", HTTP_PORT)
	// 	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 		log.Fatalf("failed to serve HTTP: %v", err)
	// 	}
	// }()

	go func() {
		log.Printf("HTTP Gateway started on %s\n", "8084")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to serve HTTP: %v", err)
		}
	}()

	//! ===== Swagger UI server =====
	// go func() {
	// 	log.Println("Swagger UI started on :8082")
	// 	err := http.ListenAndServe(":8082", httpSwagger.Handler(
	// 		httpSwagger.URL("http://localhost:8081/swagger.json"),
	// 	))
	// 	if err != nil {
	// 		log.Fatalf("failed to serve swagger: %v", err)
	// 	}
	// }()

	go func() {
		log.Println("Swagger UI started on : 8083")
		err := http.ListenAndServe(":8083", httpSwagger.Handler(
			httpSwagger.URL("http://localhost:8084/swagger.json"),
		))
		if err != nil {
			log.Fatalf("failed to serve swagger: %v", err)
		}
	}()

	//! ===== Swagger UI server =====
	// go func() {
	// 	log.Println("Swagger UI started on :8082")
	// 	err := http.ListenAndServe(HTTP_PORT, httpSwagger.Handler(
	// 		httpSwagger.URL("http://localhost:8082/swagger.json"),
	// 	))
	// 	if err != nil {
	// 		log.Fatalf("failed to serve swagger: %v", err)
	// 	}
	// }()

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
