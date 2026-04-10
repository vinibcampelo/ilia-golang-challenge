# ilia-golang-challenge (repo root) — run from repository root: make <target>
#
# Orchestrates both microservices by delegating to their Makefiles under services/*.
# Each service keeps its own docker-compose.yml, .env.example, and service-local targets.
#
# Go commands that apply to the whole module run from REPO_ROOT (single go.mod; no go.work).
#
# There is no docker-compose.yml at the repository root; use `make up` here or
# `make -C services/ms-users …` / `make -C services/ms-transactions …` for one stack.

SHELL         := bash
.DEFAULT_GOAL := help

REPO_ROOT            := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
MS_USERS_DIR         := $(REPO_ROOT)/services/ms-users
MS_TRANSACTIONS_DIR  := $(REPO_ROOT)/services/ms-transactions

COMPOSE      ?= docker compose
COMPOSE_FILE ?= docker-compose.yml

.PHONY: help
.PHONY: setup setup-dev setup-prod
.PHONY: up up-dev up-example up-prod down down-v build restart ps
.PHONY: logs-users logs-transactions
.PHONY: run-users run-transactions
.PHONY: test fmt vet build-api

# ------------------------------------------------------------------------------
# Help
# ------------------------------------------------------------------------------

help: ## List targets (default goal)
	@echo "ilia-golang-challenge (root orchestration)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "  %-22s%s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort -u
	@echo ""
	@echo "Variables (override: make VAR=value <target>):"
	@echo "  COMPOSE=$(COMPOSE)  COMPOSE_FILE=$(COMPOSE_FILE)"
	@echo "  REPO_ROOT=$(REPO_ROOT)"
	@echo "  MS_USERS_DIR=$(MS_USERS_DIR)"
	@echo "  MS_TRANSACTIONS_DIR=$(MS_TRANSACTIONS_DIR)"

# ------------------------------------------------------------------------------
# Orchestration (delegate to service Makefiles)
# ------------------------------------------------------------------------------

setup: ## Run setup in ms-users and ms-transactions (.env from .env.example if missing)
	$(MAKE) -C "$(MS_USERS_DIR)" setup
	$(MAKE) -C "$(MS_TRANSACTIONS_DIR)" setup

setup-dev: ## Run setup-dev in both services
	$(MAKE) -C "$(MS_USERS_DIR)" setup-dev
	$(MAKE) -C "$(MS_TRANSACTIONS_DIR)" setup-dev

setup-prod: ## Run setup-prod in both services
	$(MAKE) -C "$(MS_USERS_DIR)" setup-prod
	$(MAKE) -C "$(MS_TRANSACTIONS_DIR)" setup-prod

up: up-dev ## Alias for up-dev (both stacks; ms-users first, then ms-transactions)

up-dev: setup ## Start both stacks (Compose vars from each service .env)
	$(MAKE) -C "$(MS_USERS_DIR)" up-dev
	$(MAKE) -C "$(MS_TRANSACTIONS_DIR)" up-dev

up-example: ## Start both stacks using each service .env.example only (no setup)
	$(MAKE) -C "$(MS_USERS_DIR)" up-example
	$(MAKE) -C "$(MS_TRANSACTIONS_DIR)" up-example

up-prod: setup-prod ## Start both stacks (Compose vars from each service .env.prod)
	$(MAKE) -C "$(MS_USERS_DIR)" up-prod
	$(MAKE) -C "$(MS_TRANSACTIONS_DIR)" up-prod

down: ## Stop both stacks (keep volumes)
	$(MAKE) -C "$(MS_TRANSACTIONS_DIR)" down
	$(MAKE) -C "$(MS_USERS_DIR)" down

down-v: ## Stop both stacks and remove volumes (wipes Postgres data)
	$(MAKE) -C "$(MS_TRANSACTIONS_DIR)" down-v
	$(MAKE) -C "$(MS_USERS_DIR)" down-v

build: ## Build images for both stacks
	$(MAKE) -C "$(MS_USERS_DIR)" build
	$(MAKE) -C "$(MS_TRANSACTIONS_DIR)" build

restart: ## Restart all services in both stacks
	$(MAKE) -C "$(MS_USERS_DIR)" restart
	$(MAKE) -C "$(MS_TRANSACTIONS_DIR)" restart

ps: ## Service status for both stacks
	@echo "=== ms-users ==="
	$(MAKE) -C "$(MS_USERS_DIR)" ps
	@echo ""
	@echo "=== ms-transactions ==="
	$(MAKE) -C "$(MS_TRANSACTIONS_DIR)" ps

logs-users: ## Follow ms-users container logs (Ctrl+C to stop)
	$(MAKE) -C "$(MS_USERS_DIR)" logs

logs-transactions: ## Follow ms-transactions container logs (Ctrl+C to stop)
	$(MAKE) -C "$(MS_TRANSACTIONS_DIR)" logs

run-users: ## Run ms-users API on host (not Docker); uses that service .env via child Makefile
	$(MAKE) -C "$(MS_USERS_DIR)" run

run-transactions: ## Run ms-transactions API on host (not Docker)
	$(MAKE) -C "$(MS_TRANSACTIONS_DIR)" run

# ------------------------------------------------------------------------------
# Go (whole module)
# ------------------------------------------------------------------------------

test: ## Run all tests (no cache); repo-root module
	cd "$(REPO_ROOT)" && go test ./... -count=1

fmt: ## Format all packages in the module
	cd "$(REPO_ROOT)" && go fmt ./...

vet: ## Static analysis for the whole module
	cd "$(REPO_ROOT)" && go vet ./...

build-api: ## Build API binaries into bin/
	@mkdir -p "$(REPO_ROOT)/bin"
	cd "$(REPO_ROOT)" && go build -o bin/ms-users-api ./services/ms-users/cmd/api
	cd "$(REPO_ROOT)" && go build -o bin/ms-transactions-api ./services/ms-transactions/cmd/api
