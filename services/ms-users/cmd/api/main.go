package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"ilia-golang-challenge/services/ms-users/internal/application/user/usecase"
	"ilia-golang-challenge/services/ms-users/internal/config"
	"ilia-golang-challenge/services/ms-users/internal/infrastructure/db"
	"ilia-golang-challenge/services/ms-users/internal/infrastructure/httpapi"
	"ilia-golang-challenge/services/ms-users/internal/infrastructure/jwtissuer"
	"ilia-golang-challenge/services/ms-users/internal/infrastructure/persistence"
	"ilia-golang-challenge/services/ms-users/internal/infrastructure/walletclient"
)

func main() {
	loadEnvFile()

	configuration, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	mustRunMigrations(configuration.DatabaseURL)

	dbConn := mustOpenDB(configuration)
	defer func() {
		if err := dbConn.Close(); err != nil {
			log.Printf("database close: %v", err)
		}
	}()

	openAPISpecBytes := loadSwaggerSpecBytesOrNil(configuration.OpenAPISpecPath)
	router := buildRouter(dbConn, configuration, openAPISpecBytes)

	log.Printf("ms-users listening on %s", configuration.HTTPAddr)
	if err := http.ListenAndServe(configuration.HTTPAddr, router); err != nil {
		log.Fatalf("http serve: %v", err)
	}
}

func loadEnvFile() {
	envFilename := strings.TrimSpace(os.Getenv("DOTENV_FILE"))
	if envFilename == "" {
		envFilename = ".env"
	}
	_ = godotenv.Load(envFilename)
}

func mustOpenDB(configuration config.Config) *sql.DB {
	dbConn, err := sql.Open("pgx", configuration.DatabaseURL)
	if err != nil {
		log.Fatalf("database open: %v", err)
	}

	pingContext, cancel := context.WithTimeout(context.Background(), configuration.DBPingTimeout)
	defer cancel()
	if err := dbConn.PingContext(pingContext); err != nil {
		log.Fatalf("database ping: %v", err)
	}

	return dbConn
}

func mustRunMigrations(databaseURL string) {
	if err := db.RunMigrations(databaseURL); err != nil {
		log.Fatalf("migrations: %v", err)
	}
}

func loadSwaggerSpecBytesOrNil(openAPISpecPath string) []byte {
	openAPISpecPath = strings.TrimSpace(openAPISpecPath)
	if openAPISpecPath == "" || openAPISpecPath == "-" {
		return nil
	}
	openAPISpecBytes, err := os.ReadFile(openAPISpecPath)
	if err != nil {
		log.Printf("swagger: skipping UI (could not read %q: %v)", openAPISpecPath, err)
		return nil
	}
	log.Printf("swagger UI: /swagger/ (OpenAPI %q)", openAPISpecPath)
	return openAPISpecBytes
}

func buildRouter(dbConn *sql.DB, configuration config.Config, openAPISpecBytes []byte) http.Handler {
	jwtSecret := []byte(configuration.JWTSecret)
	repository := persistence.NewPostgresUserRepository(dbConn)
	tokenIssuer := &jwtissuer.HS256AccessTokenIssuer{
		Secret: jwtSecret,
		TTL:    configuration.JWTExpiration,
	}
	walletGate := walletclient.NewHTTPWalletZeroBalanceChecker(
		configuration.TransactionsServiceBaseURL,
		[]byte(configuration.JWTInternalSecret),
		configuration.TransactionsServiceTimeout,
	)
	getUserUC := usecase.NewGetUserUseCase(repository)
	userHandler := httpapi.NewUserHandler(
		usecase.NewCreateUserUseCase(repository, configuration.BcryptCost),
		usecase.NewListUsersUseCase(repository),
		getUserUC,
		usecase.NewUpdateUserUseCase(repository, configuration.BcryptCost),
		usecase.NewDeleteUserUseCase(repository, walletGate),
	)
	authHandler := httpapi.NewAuthHandler(usecase.NewAuthenticateUserUseCase(repository, tokenIssuer))
	internalUserHandler := httpapi.NewInternalUserHandler(getUserUC)
	return httpapi.NewRouter(
		userHandler,
		authHandler,
		internalUserHandler,
		jwtSecret,
		[]byte(configuration.JWTInternalSecret),
		openAPISpecBytes,
	)
}
