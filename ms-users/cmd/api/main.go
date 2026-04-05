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

	"ilia-golang-challenge/ms-users/internal/application/user/usecase"
	"ilia-golang-challenge/ms-users/internal/config"
	"ilia-golang-challenge/ms-users/internal/infrastructure/db"
	"ilia-golang-challenge/ms-users/internal/infrastructure/httpapi"
	"ilia-golang-challenge/ms-users/internal/infrastructure/persistence"
)

func loadDotenv() {
	name := strings.TrimSpace(os.Getenv("DOTENV_FILE"))
	if name == "" {
		name = ".env"
	}
	_ = godotenv.Load(name)
}

func main() {
	loadDotenv()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database open: %v", err)
	}
	defer sqlDB.Close()

	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	pingCtx, cancel := context.WithTimeout(context.Background(), cfg.DBPingTimeout)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		log.Fatalf("database ping: %v", err)
	}

	if err := db.RunMigrations(sqlDB); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	var openAPISpec []byte
	path := strings.TrimSpace(cfg.OpenAPISpecPath)
	if path != "" && path != "-" {
		b, err := os.ReadFile(path)
		if err != nil {
			log.Printf("swagger: skipping UI (could not read %q: %v)", path, err)
		} else {
			openAPISpec = b
			log.Printf("swagger UI: /swagger/ (OpenAPI %q)", path)
		}
	}

	repo := persistence.NewPostgresUserRepository(sqlDB)
	createUserUseCase := usecase.NewCreateUserUseCase(repo, cfg.BcryptCost)
	userHandler := httpapi.NewUserHandler(createUserUseCase)
	router := httpapi.NewRouter(userHandler, openAPISpec)

	log.Printf("ms-users listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, router); err != nil {
		log.Fatal(err)
	}
}
