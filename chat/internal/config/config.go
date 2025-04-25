package config

import (
	"chat/internal/transport/grpc/auth"
	"chat/internal/transport/http"
	"chat/pkg/pg"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTPServer httpserver.Config
	Postgres   pg.Config
	Auth       auth.Config
}

func MustLoad() (*Config, error) {
	var cfg Config
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
