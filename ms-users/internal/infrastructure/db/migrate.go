package db

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	pgdriver "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationSQL embed.FS

// RunMigrations applies embedded SQL migrations to the given Postgres *sql.DB.
func RunMigrations(sqlDB *sql.DB) error {
	sourceDriver, err := iofs.New(migrationSQL, "migrations")
	if err != nil {
		return fmt.Errorf("migrations source: %w", err)
	}
	dbDriver, err := pgdriver.WithInstance(sqlDB, &pgdriver.Config{})
	if err != nil {
		return fmt.Errorf("migrations postgres driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("migrations: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrations up: %w", err)
	}
	return nil
}
