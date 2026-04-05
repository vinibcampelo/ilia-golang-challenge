# ms-users — run from this directory (cd ms-users && make).

.DEFAULT_GOAL := help

# Loaded by godotenv in cmd/api. Override: make run DOTENV_FILE=.env.prod
DOTENV_FILE ?= .env

.PHONY: help run run-example run-dev run-prod test fmt vet up down logs ps

help: ## Show targets and variables
	@echo "ms-users"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort
	@echo ""
	@echo "Variables:"
	@echo "  DOTENV_FILE=$(DOTENV_FILE)  (used by: run)"

run: ## Run API: go run ./cmd/api with DOTENV_FILE
	DOTENV_FILE="$(DOTENV_FILE)" go run ./cmd/api

run-example: ## Run API with DOTENV_FILE=.env.example
	DOTENV_FILE=.env.example go run ./cmd/api

run-dev: ## Run API with DOTENV_FILE=.env.develop
	DOTENV_FILE=.env.develop go run ./cmd/api

run-prod: ## Run API with DOTENV_FILE=.env.prod
	DOTENV_FILE=.env.prod go run ./cmd/api

test: ## Run all tests (no cache)
	go test ./... -count=1

fmt: ## Format all packages (go fmt)
	go fmt ./...

vet: ## Static analysis (go vet)
	go vet ./...

up: ## docker compose up -d --build
	docker compose up -d --build

down: ## docker compose down
	docker compose down

logs: ## Follow compose logs (Ctrl+C to stop)
	docker compose logs -f

ps: ## docker compose ps
	docker compose ps
