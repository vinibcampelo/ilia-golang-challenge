package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr               string
	OpenAPISpecPath        string
	DatabaseURL            string
	DBPingTimeout          time.Duration
	JWTSecret              string
	JWTInternalSecret      string
	UsersServiceBaseURL    string
	UsersServiceTimeout    time.Duration
}

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

	jwtInternalSecret, err := requiredEnvironmentVariable("JWT_INTERNAL_SECRET")
	if err != nil {
		return Config{}, err
	}
	usersBaseURL, err := requiredEnvironmentVariable("USERS_SERVICE_BASE_URL")
	if err != nil {
		return Config{}, err
	}
	usersTimeout := durationFromEnvironmentOrDefault("USERS_SERVICE_TIMEOUT", 5*time.Second)

	return Config{
		HTTPAddr:            fmt.Sprintf(":%s", httpPort),
		OpenAPISpecPath:     openAPISpecPath,
		DatabaseURL:         databaseURL,
		DBPingTimeout:       pingTimeout,
		JWTSecret:           jwtSecret,
		JWTInternalSecret:   jwtInternalSecret,
		UsersServiceBaseURL: strings.TrimRight(usersBaseURL, "/"),
		UsersServiceTimeout: usersTimeout,
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
