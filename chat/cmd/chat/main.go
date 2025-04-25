package main

import (
	"chat/internal/config"
	"chat/internal/service"
	"chat/internal/storage/postgresql"
	"chat/internal/transport/grpc/auth"
	"chat/internal/transport/http"
	"chat/pkg/logger"
	"chat/pkg/migrator"
	"chat/pkg/pg"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.MustLoad()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	l := logger.New()
	ctx = logger.InitFromCtx(ctx, l)

	pgConn, err := pg.New(cfg.Postgres)
	if err != nil {
		l.ErrorContext(ctx, "failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer pgConn.Close()

	m, err := migrator.New("migrations", cfg.Postgres)
	if err != nil {
		l.ErrorContext(ctx, "failed to create migrator: %v", err)
		os.Exit(1)
	}
	if err = m.Up(); err != nil {
		l.ErrorContext(ctx, "failed to apply migrations: %v", err)
		os.Exit(1)
	}

	chatStorage := postgresql.New(pgConn)
	chatService := service.New(chatStorage)

	authClient := auth.New(cfg.Auth)

	server := httpserver.New(ctx, cfg.HTTPServer, chatService, authClient)
	go server.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return
	}
}
