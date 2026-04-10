# ms-users — run from this directory: cd services/ms-users && make <target>
#
# Go commands run from the repository root (single go.mod at repo root; no go.work).
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

SERVICE_DIR  := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
REPO_ROOT    := $(abspath $(SERVICE_DIR)/../..)

COMPOSE      ?= docker compose
COMPOSE_FILE ?= docker-compose.yml

API_PKG      := ./services/ms-users/cmd/api
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
	@echo "  REPO_ROOT=$(REPO_ROOT)"

# ------------------------------------------------------------------------------
# Environment files (godotenv / compose)
# ------------------------------------------------------------------------------

setup: ## Create .env from .env.example if missing
	@if [ -f "$(SERVICE_DIR)/.env" ]; then echo ".env already present"; \
	else cp "$(SERVICE_DIR)/.env.example" "$(SERVICE_DIR)/.env" && echo "Created .env from .env.example"; fi

setup-dev: ## Create .env.develop from .env.example if missing
	@if [ -f "$(SERVICE_DIR)/.env.develop" ]; then echo ".env.develop already present"; \
	else cp "$(SERVICE_DIR)/.env.example" "$(SERVICE_DIR)/.env.develop" && echo "Created .env.develop from .env.example"; fi

setup-prod: ## Create .env.prod from .env.example if missing
	@if [ -f "$(SERVICE_DIR)/.env.prod" ]; then echo ".env.prod already present"; \
	else cp "$(SERVICE_DIR)/.env.example" "$(SERVICE_DIR)/.env.prod" && echo "Created .env.prod from .env.example"; fi

# ------------------------------------------------------------------------------
# Docker Compose (stack)
# ------------------------------------------------------------------------------

up: up-dev ## Alias for up-dev

up-dev: setup ## Start stack; Compose vars from .env
	@docker network inspect ilia-challenge-internal >/dev/null 2>&1 || docker network create ilia-challenge-internal
	$(COMPOSE) --project-directory "$(SERVICE_DIR)" --env-file "$(SERVICE_DIR)/.env" -f "$(SERVICE_DIR)/$(COMPOSE_FILE)" up -d --build

up-example: ## Start stack; Compose vars from .env.example only (no setup)
	@test -f "$(SERVICE_DIR)/.env.example" || { echo "missing .env.example"; exit 1; }
	@docker network inspect ilia-challenge-internal >/dev/null 2>&1 || docker network create ilia-challenge-internal
	$(COMPOSE) --project-directory "$(SERVICE_DIR)" --env-file "$(SERVICE_DIR)/.env.example" -f "$(SERVICE_DIR)/$(COMPOSE_FILE)" up -d --build

up-prod: setup-prod ## Start stack; Compose vars from .env.prod
	@test -f "$(SERVICE_DIR)/.env.prod" || { echo "missing .env.prod"; exit 1; }
	@docker network inspect ilia-challenge-internal >/dev/null 2>&1 || docker network create ilia-challenge-internal
	$(COMPOSE) --project-directory "$(SERVICE_DIR)" --env-file "$(SERVICE_DIR)/.env.prod" -f "$(SERVICE_DIR)/$(COMPOSE_FILE)" up -d --build

down: ## Stop containers (keep named volumes / DB data)
	$(COMPOSE) --project-directory "$(SERVICE_DIR)" -f "$(SERVICE_DIR)/$(COMPOSE_FILE)" down

down-v: ## Stop containers and remove volumes (wipes Postgres data)
	$(COMPOSE) --project-directory "$(SERVICE_DIR)" -f "$(SERVICE_DIR)/$(COMPOSE_FILE)" down -v --remove-orphans

build: ## Build images only
	$(COMPOSE) --project-directory "$(SERVICE_DIR)" -f "$(SERVICE_DIR)/$(COMPOSE_FILE)" build

restart: ## Restart all services
	$(COMPOSE) --project-directory "$(SERVICE_DIR)" -f "$(SERVICE_DIR)/$(COMPOSE_FILE)" restart

logs: ## Follow container logs (Ctrl+C to stop)
	$(COMPOSE) --project-directory "$(SERVICE_DIR)" -f "$(SERVICE_DIR)/$(COMPOSE_FILE)" logs -f

ps: ## Service status
	$(COMPOSE) --project-directory "$(SERVICE_DIR)" -f "$(SERVICE_DIR)/$(COMPOSE_FILE)" ps

# ------------------------------------------------------------------------------
# API on host (Go)
# ------------------------------------------------------------------------------

run: setup ## Run API locally; uses DOTENV_FILE (default .env) under this service directory
	cd "$(REPO_ROOT)" && DOTENV_FILE="$(SERVICE_DIR)/$(DOTENV_FILE)" go run $(API_PKG)

run-example: ## Run API with .env.example (does not create .env)
	@test -f "$(SERVICE_DIR)/.env.example" || { echo "missing .env.example"; exit 1; }
	cd "$(REPO_ROOT)" && DOTENV_FILE="$(SERVICE_DIR)/.env.example" go run $(API_PKG)

run-dev: setup-dev ## Run API with .env.develop
	cd "$(REPO_ROOT)" && DOTENV_FILE="$(SERVICE_DIR)/.env.develop" go run $(API_PKG)

run-prod: setup-prod ## Run API with .env.prod
	cd "$(REPO_ROOT)" && DOTENV_FILE="$(SERVICE_DIR)/.env.prod" go run $(API_PKG)

# ------------------------------------------------------------------------------
# Go quality
# ------------------------------------------------------------------------------

test: ## Run tests (no cache); uses repo-root module
	cd "$(REPO_ROOT)" && go test ./services/ms-users/... -count=1

fmt: ## Format packages under this service
	cd "$(REPO_ROOT)" && go fmt ./services/ms-users/...

vet: ## Static analysis
	cd "$(REPO_ROOT)" && go vet ./services/ms-users/...
