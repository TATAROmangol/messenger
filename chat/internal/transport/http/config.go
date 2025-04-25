package httpserver

import "time"

type Config struct {
	Host        string        `env:"HTTP_HOST"`
	Port        string        `env:"HTTP_PORT"`
	Timeout     time.Duration `env:"TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
}

func (c *Config) Address() string {
	// return c.Host + ":" + c.Port
	return "0.0.0.0:" + c.Port
}
