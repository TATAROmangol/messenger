package auth

import "fmt"

type Config struct {
	Host string `env:"AUTH_GRPC_HOST"`
	Port string `env:"AUTH_GRPC_PORT"`
}

func (c *Config) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
