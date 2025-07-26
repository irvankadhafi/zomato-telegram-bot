# Variables
APP_NAME=telegram-bot-api
BIN_DIR=bin
CMD_DIR=cmd/api
DOCKER_IMAGE=telegram-bot-api
DOCKER_TAG=latest

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-a -installsuffix cgo

.PHONY: all build clean test coverage deps fmt vet lint run dev docker help

# Default target
all: clean deps fmt vet test build

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)/main.go
	@CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BIN_DIR)/cli cmd/cli/main.go
	@echo "Build completed: $(BIN_DIR)/$(APP_NAME)"

# Build for current OS
build-local:
	@echo "Building $(APP_NAME) for local OS..."
	@mkdir -p $(BIN_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)/main.go
	@$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/cli cmd/cli/main.go
	@echo "Local build completed: $(BIN_DIR)/$(APP_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean completed"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "Dependencies updated"

# Format code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) -s -w .
	@echo "Code formatted"

# Run go vet
vet:
	@echo "Running go vet..."
	@$(GOVET) ./...
	@echo "Vet completed"

# Run tests
test: build-local
	@echo "Running tests..."
	@./$(BIN_DIR)/cli test unit
	@echo "Tests completed"

# Run all tests
test-all: build-local
	@echo "Running all tests..."
	@./$(BIN_DIR)/cli test all

# Run tests with coverage
coverage: build-local
	@echo "Running tests with coverage..."
	@./$(BIN_DIR)/cli test coverage --html
	@echo "Coverage report generated"

# Run integration tests
test-integration: build-local
	@echo "Running integration tests..."
	@./$(BIN_DIR)/cli test integration

# Run benchmark tests
test-benchmark: build-local
	@echo "Running benchmark tests..."
	@./$(BIN_DIR)/cli test benchmark

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run
	@echo "Linting completed"

# Run the application
run: build-local
	@echo "Starting $(APP_NAME)..."
	@./$(BIN_DIR)/$(APP_NAME)

# Run CLI
run-cli: build-local
	@echo "Running CLI..."
	@./$(BIN_DIR)/cli

# Start server via CLI
server: build-local
	@echo "Starting server via CLI..."
	@./$(BIN_DIR)/cli server

# Run in development mode with live reload (requires air)
dev:
	@echo "Starting development server with live reload..."
	@air

# Database operations
db-migrate: build-local
	@echo "Running database migrations..."
	@./$(BIN_DIR)/cli migrate up
	@echo "Database migration completed"

db-migrate-down: build-local
	@echo "Rolling back last migration..."
	@./$(BIN_DIR)/cli migrate down

db-reset: build-local
	@echo "Resetting database..."
	@./$(BIN_DIR)/cli migrate reset --force
	@echo "Database reset completed"

db-status: build-local
	@echo "Checking migration status..."
	@./$(BIN_DIR)/cli migrate status

# Docker operations
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose-up:
	@echo "Starting services with docker-compose..."
	@docker-compose up -d
	@echo "Services started"

docker-compose-down:
	@echo "Stopping services with docker-compose..."
	@docker-compose down
	@echo "Services stopped"

docker-compose-logs:
	@docker-compose logs -f

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/api/main.go -o docs
	@echo "Swagger documentation generated"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2
	@echo "Development tools installed"

# Security scan
security:
	@echo "Running security scan..."
	@gosec ./...
	@echo "Security scan completed"

# Performance benchmark
bench:
	@echo "Running benchmarks..."
	@$(GOTEST) -bench=. -benchmem ./...
	@echo "Benchmarks completed"

# Check for updates
check-updates:
	@echo "Checking for dependency updates..."
	@go list -u -m all

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application for Linux"
	@echo "  build-local    - Build the application for current OS"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  test           - Run tests"
	@echo "  coverage       - Run tests with coverage"
	@echo "  lint           - Run linter"
	@echo "  run            - Build and run the application"
	@echo "  dev            - Run in development mode with live reload"
	@echo "  db-migrate     - Run database migrations"
	@echo "  db-reset       - Reset database"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-compose-up   - Start services with docker-compose"
	@echo "  docker-compose-down - Stop services with docker-compose"
	@echo "  swagger        - Generate Swagger documentation"
	@echo "  install-tools  - Install development tools"
	@echo "  security       - Run security scan"
	@echo "  bench          - Run benchmarks"
	@echo "  check-updates  - Check for dependency updates"
	@echo "  help           - Show this help message"