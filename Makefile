.PHONY: all test swagger help

# Tools
GOCMD=go
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Swagger
SWAG=swag

all: test swagger

test:
	$(GOTEST) -v -race -cover ./...

swagger:
	$(SWAG) init -g cmd/app/main.go -o docs

local:
	$(GOCMD) run ./cmd/app -e=local

init:
	$(GOCMD) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Database Migrations
migrate-up:
	$(GOCMD) run ./cmd/db up -e local

migrate-down:
	$(GOCMD) run ./cmd/db down -e local

migrate-create:
	@NAME=$(name); \
	if [ -z "$$NAME" ]; then \
		NAME=$(filter-out $@,$(MAKECMDGOALS)); \
	fi; \
	if [ -z "$$NAME" ]; then \
		echo "Usage: make migrate-create <name>"; \
		exit 1; \
	fi; \
	$(GOCMD) run ./cmd/db create -n $$NAME

%:
	@:

migrate-force:
	@if [ -z "$(version)" ]; then echo "Usage: make migrate-force version=1"; exit 1; fi
	$(GOCMD) run ./cmd/db force --version $(version) -e local

migrate-version:
	$(GOCMD) run ./cmd/db version -e local

help:
	@echo "Available commands:"
	@echo "  make test            - Run unit tests"
	@echo "  make swagger         - Generate Swagger documentation"
	@echo "  make all             - Run test and swagger"
	@echo "  make local           - Run the application in local mode"
	@echo "  make init            - Initialize the application"
	@echo ""
	@echo "Database Migrations:"
	@echo "  make migrate-create <name>  - Create a new migration file"
	@echo "  make migrate-up             - Run all pending up migrations"
	@echo "  make migrate-down           - Rollback the last migration"
	@echo "  make migrate-force version=x- Force migration to a specific version (use when dirty)"
	@echo "  make migrate-version        - Check the current migration version"
