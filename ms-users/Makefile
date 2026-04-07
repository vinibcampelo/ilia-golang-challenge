# ms-users — run from this directory: cd ms-users && make <target>
#
# Docker Compose variable substitution comes from an explicit --env-file per target:
#   up-dev / up     → .env          (make setup creates .env from .env.example if missing)
#   up-example      → .env.example  (does not create .env)
#   up-prod         → .env.prod     (make setup-prod)
#
# If a root .env file exists, some Compose versions still merge it with --env-file; for a
# strict “example-only” stack, run up-example from a clone without .env or rename .env temporarily.
#
# `make run*` runs the Go binary on the host (not in Docker).

SHELL         := bash
.DEFAULT_GOAL := help

COMPOSE      ?= docker compose
COMPOSE_FILE ?= docker-compose.yml

API_PKG      := ./cmd/api
DOTENV_FILE  ?= .env

.PHONY: help setup setup-dev setup-prod
.PHONY: up up-dev up-example up-prod down build restart logs ps down-v
.PHONY: run run-example run-dev run-prod
.PHONY: test fmt vet

# ------------------------------------------------------------------------------
# Help
# ------------------------------------------------------------------------------

help: ## List targets (default goal)
	@echo "ms-users"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "  %-22s%s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort -u
	@echo ""
	@echo "Variables (override: make VAR=value <target>):"
	@echo "  COMPOSE=$(COMPOSE)  COMPOSE_FILE=$(COMPOSE_FILE)  DOTENV_FILE=$(DOTENV_FILE)"

# ------------------------------------------------------------------------------
# Environment files (godotenv / compose)
# ------------------------------------------------------------------------------

setup: ## Create .env from .env.example if missing
	@if [ -f .env ]; then echo ".env already present"; \
	else cp .env.example .env && echo "Created .env from .env.example"; fi

setup-dev: ## Create .env.develop from .env.example if missing
	@if [ -f .env.develop ]; then echo ".env.develop already present"; \
	else cp .env.example .env.develop && echo "Created .env.develop from .env.example"; fi

setup-prod: ## Create .env.prod from .env.example if missing
	@if [ -f .env.prod ]; then echo ".env.prod already present"; \
	else cp .env.example .env.prod && echo "Created .env.prod from .env.example"; fi

# ------------------------------------------------------------------------------
# Docker Compose (stack)
# ------------------------------------------------------------------------------

up: up-dev ## Alias for up-dev

up-dev: setup ## Start stack; Compose vars from .env
	$(COMPOSE) --env-file .env -f $(COMPOSE_FILE) up -d --build

up-example: ## Start stack; Compose vars from .env.example only (no setup)
	@test -f .env.example || { echo "missing .env.example"; exit 1; }
	$(COMPOSE) --env-file .env.example -f $(COMPOSE_FILE) up -d --build

up-prod: setup-prod ## Start stack; Compose vars from .env.prod
	@test -f .env.prod || { echo "missing .env.prod"; exit 1; }
	$(COMPOSE) --env-file .env.prod -f $(COMPOSE_FILE) up -d --build

down: ## Stop containers (keep named volumes / DB data)
	$(COMPOSE) -f $(COMPOSE_FILE) down

down-v: ## Stop containers and remove volumes (wipes Postgres data)
	$(COMPOSE) -f $(COMPOSE_FILE) down -v --remove-orphans

build: ## Build images only
	$(COMPOSE) -f $(COMPOSE_FILE) build

restart: ## Restart all services
	$(COMPOSE) -f $(COMPOSE_FILE) restart

logs: ## Follow container logs (Ctrl+C to stop)
	$(COMPOSE) -f $(COMPOSE_FILE) logs -f

ps: ## Service status
	$(COMPOSE) -f $(COMPOSE_FILE) ps

# ------------------------------------------------------------------------------
# API on host (Go)
# ------------------------------------------------------------------------------

run: setup ## Run API locally; uses DOTENV_FILE (default .env)
	DOTENV_FILE="$(DOTENV_FILE)" go run $(API_PKG)

run-example: ## Run API with .env.example (does not create .env)
	@test -f .env.example || { echo "missing .env.example"; exit 1; }
	DOTENV_FILE=.env.example go run $(API_PKG)

run-dev: setup-dev ## Run API with .env.develop
	DOTENV_FILE=.env.develop go run $(API_PKG)

run-prod: setup-prod ## Run API with .env.prod
	DOTENV_FILE=.env.prod go run $(API_PKG)

# ------------------------------------------------------------------------------
# Go quality
# ------------------------------------------------------------------------------

test: ## Run tests (no cache)
	go test ./... -count=1

fmt: ## Format packages
	go fmt ./...

vet: ## Static analysis
	go vet ./...
