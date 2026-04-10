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

	"ilia-golang-challenge/services/ms-transactions/internal/application/transaction/usecase"
	"ilia-golang-challenge/services/ms-transactions/internal/config"
	"ilia-golang-challenge/services/ms-transactions/internal/infrastructure/db"
	"ilia-golang-challenge/services/ms-transactions/internal/infrastructure/httpapi"
	"ilia-golang-challenge/services/ms-transactions/internal/infrastructure/persistence"
	"ilia-golang-challenge/services/ms-transactions/internal/infrastructure/userdirectory"
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

	log.Printf("ms-transactions listening on %s", configuration.HTTPAddr)
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
	internalSecret := []byte(configuration.JWTInternalSecret)
	repository := persistence.NewPostgresTransactionRepository(dbConn)
	balanceUC := usecase.NewGetBalanceUseCase(repository)
	userGate := userdirectory.NewHTTPUserActiveGate(
		configuration.UsersServiceBaseURL,
		internalSecret,
		configuration.UsersServiceTimeout,
	)
	handler := httpapi.NewTransactionHandler(
		usecase.NewCreateTransactionUseCase(repository, userGate),
		usecase.NewListTransactionsUseCase(repository),
		balanceUC,
	)
	internalWallet := httpapi.NewInternalWalletHandler(balanceUC)
	return httpapi.NewRouter(
		handler,
		internalWallet,
		jwtSecret,
		internalSecret,
		openAPISpecBytes,
	)
}
