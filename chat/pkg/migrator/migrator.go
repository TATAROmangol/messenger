package migrator

import (
	"errors"
	"fmt"
	"os"

	"chat/pkg/pg"

	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
)

type Migrator struct {
	m *migrate.Migrate
}

// dirPath - dir with migrate files
func New(dirPath string, cfg pg.Config) (*Migrator, error) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory does not exist: %s", dirPath)
	}
	address := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&x-migrations-table=chat_migrations",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
	m, err := migrate.New("file://"+dirPath, address)
	if err != nil {
		return nil, fmt.Errorf("failed create migrator, err: %v", err)
	}

	return &Migrator{m}, nil
}

func (mig *Migrator) Up() error {
	defer mig.m.Close()

	err := mig.m.Up()
	if err == nil || errors.Is(err, migrate.ErrNoChange) {
		return nil
	}

	version, _, _ := mig.m.Version()
	vers := int(version) - 1
	if err := mig.m.Force(vers); err != nil {
		return fmt.Errorf("failed rollback migration: err=%v", err)
	}

	return fmt.Errorf("migrations are not applied: current version=%v, err=%v", vers, err)
}
