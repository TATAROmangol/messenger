package postgres

import (
	"auth_service/internal/config"
	"context"
	"errors"
	"fmt"
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	PGXPool *pgxpool.Pool
)

func New(ctx context.Context, cfg config.PGConfig, path string) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_max_conns=%d&pool_min_conns=%d",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,

		cfg.MaxConns,
		cfg.MinConns,
	)

	var err error

	PGXPool, err = pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to database: %w", err)
	}

	m, err := migrate.New(
		"file://"+path,
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&x-migrations-table=auth-schema",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("Unable to create migrations: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("Unable to run migrations: %w", err)
	}

	return PGXPool, nil
}

func InsertUser(token, login, email, pass, name string) (string, error) {
	ret := ""
	err := PGXPool.QueryRow(context.Background(), "INSERT INTO auth_schema.users (token, login, email, pass, name) VALUES ($1, $2, $3, $4, $5) RETURNING id;", token, login, email, pass, name).Scan(&ret)
	return ret, err
}

func GetIdByToken(token string) (int, error) {
	ret := -1
	err := PGXPool.QueryRow(context.Background(), "SELECT id FROM auth_schema.users WHERE token=$1", token).Scan(&ret)
	return ret, err
}

func GetTokenByCredentialAndPassword(credential, pass string) (string, error) {
	ret := ""
	err := PGXPool.QueryRow(context.Background(), "SELECT token FROM auth_schema.users WHERE pass=$1 AND (login=$2 OR email=$2);", pass, credential).Scan(&ret)
	return ret, err
}

func DeleteUserByToken(token string) error {
	_, err := PGXPool.Exec(context.Background(), "DELETE FROM auth_schema.users WHERE token=$1;", token)
	return err
}

func GetRecordByToken(token string) (string, string, string, error) {
	type Record struct {
		Login string
		Email string
		Name  string
	}
	var ret Record
	err := PGXPool.QueryRow(context.Background(), "SELECT login, email, name FROM auth_schema.users WHERE token=$1;", token).Scan(&ret.Login, &ret.Email, &ret.Name)
	return ret.Login, ret.Email, ret.Name, err
}
