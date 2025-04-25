package pg

type Config struct {
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port     uint16 `env:"POSTGRES_PORT" env-required:"true"`
	Username string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD"`
	Database string `env:"POSTGRES_DB" env-required:"true"`
	MaxConn  int32  `env:"POSTGRES_MAX_CONN"`
	MinConn  int32  `env:"POSTGRES_MIN_CONN"`
}