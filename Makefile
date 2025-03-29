.PHONY: build run test clean docker docker-compose db-setup db-migrate lint fmt help

# Default target
all: build

# Build the application
build:
	@echo "Building LlamaChat..."
	go build -o bin/llamachat ./cmd/llamachat

# Run the application
run: build
	@echo "Running LlamaChat..."
	./bin/llamachat --config config.json

# Run the application in debug mode
debug: build
	@echo "Running LlamaChat in debug mode..."
	./bin/llamachat --config config.json --debug

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t llamachat .

# Run with Docker Compose
docker-compose:
	@echo "Starting with Docker Compose..."
	docker-compose up -d

# Stop Docker Compose
docker-compose-down:
	@echo "Stopping Docker Compose..."
	docker-compose down

# Setup the database
db-setup:
	@echo "Setting up database..."
	@if [ -z "$(DB_USER)" ]; then \
		psql -U postgres -c "CREATE USER llamachat WITH PASSWORD 'llamachat';"; \
		psql -U postgres -c "CREATE DATABASE llamachat OWNER llamachat;"; \
	else \
		psql -U $(DB_USER) -c "CREATE USER llamachat WITH PASSWORD 'llamachat';"; \
		psql -U $(DB_USER) -c "CREATE DATABASE llamachat OWNER llamachat;"; \
	fi
	psql -U llamachat -d llamachat -f schema.sql

# Initialize or migrate the database
db-migrate:
	@echo "Migrating database..."
	psql -U llamachat -d llamachat -f schema.sql

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Generate documentation
docs:
	@echo "Generating documentation..."
	mkdir -p docs/api
	swag init -g cmd/llamachat/main.go -o docs/api

# Show help
help:
	@echo "LlamaChat Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build           Build the application"
	@echo "  make run             Build and run the application"
	@echo "  make debug           Run in debug mode"
	@echo "  make test            Run tests"
	@echo "  make test-coverage   Run tests with coverage report"
	@echo "  make clean           Clean build artifacts"
	@echo "  make docker          Build Docker image"
	@echo "  make docker-compose  Run with Docker Compose"
	@echo "  make docker-compose-down Stop Docker Compose"
	@echo "  make db-setup        Setup database (create user and db)"
	@echo "  make db-migrate      Apply database migrations"
	@echo "  make lint            Run linter"
	@echo "  make fmt             Format code"
	@echo "  make docs            Generate API documentation"
	@echo "  make help            Show this help message" 