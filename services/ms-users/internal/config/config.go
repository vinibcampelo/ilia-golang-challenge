package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	HTTPAddr        string
	OpenAPISpecPath string
	DatabaseURL     string
	DBPingTimeout   time.Duration
	BcryptCost      int
	JWTSecret       string
	JWTExpiration   time.Duration
}

// Load reads configuration from the process environment (populate from a .env file via godotenv in main).
// Required: PORT, DATABASE_URL, JWT_SECRET. Optional keys are documented in .env.example.
func Load() (Config, error) {
	httpPort, err := requiredEnvironmentVariable("PORT")
	if err != nil {
		return Config{}, err
	}
	databaseURL, err := requiredEnvironmentVariable("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}
	openAPISpecPath := strings.TrimSpace(os.Getenv("OPENAPI_SPEC"))

	pingTimeout := durationFromEnvironmentOrDefault("DB_PING_TIMEOUT", 10*time.Second)

	jwtSecret, err := requiredEnvironmentVariable("JWT_SECRET")
	if err != nil {
		return Config{}, err
	}
	jwtExpiration := durationFromEnvironmentOrDefault("JWT_EXPIRATION", 24*time.Hour)

	return Config{
		HTTPAddr:        fmt.Sprintf(":%s", httpPort),
		OpenAPISpecPath: openAPISpecPath,
		DatabaseURL:     databaseURL,
		DBPingTimeout:   pingTimeout,
		BcryptCost:      bcryptCostFromEnvironment(),
		JWTSecret:       jwtSecret,
		JWTExpiration:   jwtExpiration,
	}, nil
}

func requiredEnvironmentVariable(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("config: %s is required (see .env.example)", key)
	}
	return value, nil
}

func bcryptCostFromEnvironment() int {
	rawValue := strings.TrimSpace(os.Getenv("BCRYPT_COST"))
	if rawValue == "" {
		return bcrypt.DefaultCost
	}
	cost, err := strconv.Atoi(rawValue)
	if err != nil || cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return bcrypt.DefaultCost
	}
	return cost
}

func durationFromEnvironmentOrDefault(key string, defaultValue time.Duration) time.Duration {
	rawValue := strings.TrimSpace(os.Getenv(key))
	if rawValue == "" {
		return defaultValue
	}
	parsed, err := time.ParseDuration(rawValue)
	if err != nil {
		return defaultValue
	}
	return parsed
}
