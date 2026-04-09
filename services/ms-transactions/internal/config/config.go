package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr        string
	OpenAPISpecPath string
	DatabaseURL     string
	DBPingTimeout   time.Duration
	JWTSecret       string
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

	return Config{
		HTTPAddr:        fmt.Sprintf(":%s", httpPort),
		OpenAPISpecPath: openAPISpecPath,
		DatabaseURL:     databaseURL,
		DBPingTimeout:   pingTimeout,
		JWTSecret:       jwtSecret,
	}, nil
}

func requiredEnvironmentVariable(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("config: %s is required (see .env.example)", key)
	}
	return value, nil
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
