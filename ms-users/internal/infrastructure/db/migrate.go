package db

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	pgdriver "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationSQL embed.FS

// RunMigrations uses its own *sql.DB. Do not pass the app pool: m.Close() on golang-migrate’s
// postgres driver calls sql.DB.Close() on that instance, which would kill the API pool.
func RunMigrations(databaseURL string) error {
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("migrations open: %w", err)
	}

	sourceDriver, err := iofs.New(migrationSQL, "migrations")
	if err != nil {
		_ = sqlDB.Close()
		return fmt.Errorf("migrations source: %w", err)
	}
	dbDriver, err := pgdriver.WithInstance(sqlDB, &pgdriver.Config{})
	if err != nil {
		_ = sqlDB.Close()
		return fmt.Errorf("migrations postgres driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		_ = sqlDB.Close()
		return fmt.Errorf("migrations: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrations up: %w", err)
	}
	return nil
}
