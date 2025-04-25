package main

import (
	"net"
	"net/http"
	"fmt"
	"context"
	"auth_service/pkg/logger"
	"auth_service/internal/config"
	"auth_service/pkg/postgres"
	"auth_service/internal/transport"
	"os"
	"os/signal"
	"go.uber.org/zap"
	"errors"
	"time"
	pbapi "auth_service/pkg/api"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	ctx, _ = logger.New(ctx)
	logger.Info(ctx, "Logger started")


	logger.Info(ctx, "Loading config...")
	cfg, err := config.New("./config/config.env")
	transport.JWTKey = cfg.JWTKey
	if err != nil {
		logger.Fatal(ctx, "Failed to load config", zap.Error(err))
	}

	logger.Info(ctx, fmt.Sprintf("Config: REST_PORT=%d, GRPC_PORT=%d, JWTKEY=%s, POSTGRES: {POSTGRES_HOST=%s, POSTGRES_PORT=%d, POSTGRES_USER=%s, POSTGRES_PASS=%s, POSTGRES_MAX_CONN=%d, POSTGRES_MIN_CONN=%d}.", cfg.Port, cfg.GRPC_Port, cfg.JWTKey, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.Username, cfg.Postgres.Password, cfg.Postgres.MaxConns, cfg.Postgres.MinConns))
	
	logger.Info(ctx, "Connecting to the database...")
	pool, err := postgres.New(ctx, cfg.Postgres, "./db/migrations")
	if err != nil {
		logger.Fatal(ctx, "Failed to connect to the database", zap.Error(err))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/register", transport.Register)
	mux.HandleFunc("/auth/login", transport.Login) // TODO: temporary
	mux.HandleFunc("/auth/logout", transport.Logout)
	mux.HandleFunc("/auth/delete", transport.Delete)

	address := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	logger.Info(ctx, fmt.Sprintf("REST: Addr:%s", address))
	server := http.Server{
		Addr: address,
		Handler: transport.MiddlewareHandler(mux),
	}

	go func() {
		logger.Info(ctx, "REST: Starting Server...")
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(ctx, "REST: Failed to serve", zap.Error(err))
		}
	}()

	srv := transport.New()
	grpc_server := grpc.NewServer(grpc.UnaryInterceptor(transport.UnaryInterceptor))
	pbapi.RegisterAuthServiceServer(grpc_server, srv)

	grpc_address := fmt.Sprintf("0.0.0.0:%d", cfg.GRPC_Port)
	logger.Info(ctx, fmt.Sprintf("gRPC: Addr:%s", grpc_address))
	lis, err := net.Listen("tcp", grpc_address)
	if err != nil {
		logger.Fatal(ctx, "gRPC: Failed to listen", zap.Error(err))
	}
	go func() {
		logger.Info(ctx, "gRPC: Starting Server...")
		if err := grpc_server.Serve(lis); err != nil {
			logger.Fatal(ctx, "gRPC: Failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	logger.Info(ctx, "Ready for graceful shutdown. Press CTRL+C to execute.")
	<-quit
	logger.Info(ctx, "Gracefully shutting down...")
	timeoutctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(timeoutctx)
	logger.Info(ctx, "REST: Server gracefully stopped.")
	grpc_server.GracefulStop()
	logger.Info(ctx, "gRPC: Server gracefully stopped.")
	pool.Close()
	logger.Info(ctx, "Database connection stopped.")
}