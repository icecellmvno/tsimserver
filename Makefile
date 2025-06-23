# TsimServer Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
SERVER_BINARY=tsimserver
MIGRATE_BINARY=tsimmigrate
SEED_BINARY=tsimseed
WEBSOCKET_BINARY=tsimwebsocket

# Directories
BIN_DIR=bin
CMD_DIR=cmd
CONFIG_FILE=config.yaml

# Build all binaries
.PHONY: all
all: clean build

# Build all applications
.PHONY: build
build: build-server build-migrate build-seed build-websocket

# Build server
.PHONY: build-server
build-server:
	@echo "Building server..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(SERVER_BINARY) $(CMD_DIR)/server/main.go

# Build migrate
.PHONY: build-migrate
build-migrate:
	@echo "Building migrate..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(MIGRATE_BINARY) $(CMD_DIR)/migrate/main.go

# Build seed
.PHONY: build-seed
build-seed:
	@echo "Building seed..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(SEED_BINARY) $(CMD_DIR)/seed/main.go

# Build websocket
.PHONY: build-websocket
build-websocket:
	@echo "Building websocket..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(WEBSOCKET_BINARY) $(CMD_DIR)/websocket/main.go

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BIN_DIR)

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Database operations
.PHONY: migrate
migrate: build-migrate
	@echo "Running database migrations..."
	./$(BIN_DIR)/$(MIGRATE_BINARY) -config=$(CONFIG_FILE)

.PHONY: migrate-reset
migrate-reset: build-migrate
	@echo "Resetting database..."
	./$(BIN_DIR)/$(MIGRATE_BINARY) -config=$(CONFIG_FILE) -reset

.PHONY: migrate-rollback
migrate-rollback: build-migrate
	@echo "Rolling back migrations..."
	./$(BIN_DIR)/$(MIGRATE_BINARY) -config=$(CONFIG_FILE) -rollback

# Seeding operations
.PHONY: seed
seed: build-seed
	@echo "Seeding database..."
	./$(BIN_DIR)/$(SEED_BINARY) -config=$(CONFIG_FILE)

.PHONY: seed-world
seed-world: build-seed
	@echo "Seeding world data..."
	./$(BIN_DIR)/$(SEED_BINARY) -config=$(CONFIG_FILE) -world

.PHONY: seed-auth
seed-auth: build-seed
	@echo "Seeding auth data..."
	./$(BIN_DIR)/$(SEED_BINARY) -config=$(CONFIG_FILE) -auth

.PHONY: seed-site
seed-site: build-seed
	@echo "Seeding site data..."
	./$(BIN_DIR)/$(SEED_BINARY) -config=$(CONFIG_FILE) -site

.PHONY: seed-verify
seed-verify: build-seed
	@echo "Verifying seeded data..."
	./$(BIN_DIR)/$(SEED_BINARY) -config=$(CONFIG_FILE) -verify

# Run applications
.PHONY: run-server
run-server: build-server
	@echo "Starting server..."
	./$(BIN_DIR)/$(SERVER_BINARY)

.PHONY: run-websocket
run-websocket: build-websocket
	@echo "Starting websocket server..."
	./$(BIN_DIR)/$(WEBSOCKET_BINARY) -port=8081

# Development setup
.PHONY: setup
setup: deps migrate seed
	@echo "Development setup completed!"

# Production setup
.PHONY: setup-prod
setup-prod: deps build migrate seed-auth
	@echo "Production setup completed!"

# Docker operations
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t tsimserver .

.PHONY: docker-run
docker-run:
	@echo "Running with Docker Compose..."
	docker-compose up -d

.PHONY: docker-stop
docker-stop:
	@echo "Stopping Docker containers..."
	docker-compose down

.PHONY: docker-logs
docker-logs:
	@echo "Showing Docker logs..."
	docker-compose logs -f

# Help
.PHONY: help
help:
	@echo "TsimServer Makefile Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build          - Build all binaries"
	@echo "  build-server   - Build server binary"
	@echo "  build-migrate  - Build migrate binary"
	@echo "  build-seed     - Build seed binary"
	@echo "  build-websocket- Build websocket binary"
	@echo ""
	@echo "Database Commands:"
	@echo "  migrate        - Run database migrations"
	@echo "  migrate-reset  - Reset database"
	@echo "  migrate-rollback - Rollback migrations"
	@echo "  seed           - Seed all data"
	@echo "  seed-world     - Seed world data only"
	@echo "  seed-auth      - Seed auth data only"
	@echo "  seed-site      - Seed site data only"
	@echo "  seed-verify    - Verify seeded data"
	@echo ""
	@echo "Run Commands:"
	@echo "  run-server     - Run main server"
	@echo "  run-websocket  - Run websocket server"
	@echo ""
	@echo "Setup Commands:"
	@echo "  setup          - Development setup"
	@echo "  setup-prod     - Production setup"
	@echo ""
	@echo "Docker Commands:"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run with Docker Compose"
	@echo "  docker-stop    - Stop Docker containers"
	@echo "  docker-logs    - Show Docker logs"
	@echo ""
	@echo "Utility Commands:"
	@echo "  deps           - Download dependencies"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  help           - Show this help" 