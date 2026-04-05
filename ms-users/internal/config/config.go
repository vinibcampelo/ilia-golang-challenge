package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	HTTPAddr          string
	OpenAPISpecPath   string
	DatabaseURL       string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBPingTimeout     time.Duration
	BcryptCost        int
}

// Load reads configuration from the process environment (populate from a .env file via godotenv in main).
// Required: PORT, DATABASE_URL. Optional keys are documented in .env.example; pool tuning falls back to
// sensible values when unset so docker-compose does not need every knob.
func Load() (Config, error) {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		return Config{}, errors.New("config: PORT is required (see .env.example)")
	}
	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dbURL == "" {
		return Config{}, errors.New("config: DATABASE_URL is required (see .env.example)")
	}
	openAPIPath := strings.TrimSpace(os.Getenv("OPENAPI_SPEC"))

	maxOpen := atoiPositiveOr(os.Getenv("DB_MAX_OPEN_CONNS"), 10)
	maxIdle := atoiPositiveOr(os.Getenv("DB_MAX_IDLE_CONNS"), 5)
	connLife := parseDurationOr(os.Getenv("DB_CONN_MAX_LIFETIME"), time.Hour)
	pingTO := parseDurationOr(os.Getenv("DB_PING_TIMEOUT"), 10*time.Second)

	return Config{
		HTTPAddr:          fmt.Sprintf(":%s", port),
		OpenAPISpecPath:   openAPIPath,
		DatabaseURL:       dbURL,
		DBMaxOpenConns:    maxOpen,
		DBMaxIdleConns:    maxIdle,
		DBConnMaxLifetime: connLife,
		DBPingTimeout:     pingTO,
		BcryptCost:        bcryptCostFromEnv(),
	}, nil
}

func bcryptCostFromEnv() int {
	raw := strings.TrimSpace(os.Getenv("BCRYPT_COST"))
	if raw == "" {
		return bcrypt.DefaultCost
	}
	c, err := strconv.Atoi(raw)
	if err != nil || c < bcrypt.MinCost || c > bcrypt.MaxCost {
		return bcrypt.DefaultCost
	}
	return c
}

func atoiPositiveOr(s string, fallback int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

func parseDurationOr(s string, fallback time.Duration) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return fallback
	}
	return d
}
