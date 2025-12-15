.PHONY: help build test lint clean docker-build docker-run docker-stop install dev local-test

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=astrometry-api-server
DOCKER_IMAGE=astrometry-api-server
DOCKER_TAG=latest
PORT=8080

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the server binary
	@echo "Building $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) ./cmd/server
	@echo "Build complete: bin/$(BINARY_NAME)"

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

test-coverage: test ## Run tests and show coverage report
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linter
	@echo "Running golangci-lint..."
	golangci-lint run --timeout=5m
	@echo "Linting complete"

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .
	@echo "Formatting complete"

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...
	@echo "Vet complete"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out coverage.html
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker build complete"

docker-run: ## Run Docker container
	@echo "Starting Docker container..."
	docker compose up -d
	@echo "Container started on http://localhost:$(PORT)"

docker-stop: ## Stop Docker container
	@echo "Stopping Docker container..."
	docker compose down
	@echo "Container stopped"

docker-logs: ## View Docker container logs
	docker compose logs -f

install: build ## Install binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	go install ./cmd/server
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

dev: ## Run server in development mode
	@echo "Starting development server..."
	go run ./cmd/server

run: build ## Build and run the server
	@echo "Running $(BINARY_NAME)..."
	./bin/$(BINARY_NAME)

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod verify
	@echo "Dependencies downloaded"

tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	go mod tidy
	@echo "Tidy complete"

swagger: ## Generate Swagger documentation from godoc comments
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/server/main.go --output ./docs
	@echo "Swagger docs generated at docs/swagger.yaml"

local-test: ## Build Docker, run integration tests locally (like CI)
	@echo "Running local integration tests..."
	@echo "Building Docker images..."
	docker compose build
	@echo "Starting containers..."
	docker compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Running health check..."
	@curl -f http://localhost:$(PORT)/health || (docker compose logs && docker compose down && exit 1)
	@echo "Health check passed"
	@echo "Running unit tests with /shared-data available..."
	@docker run --rm -v astrometry-shared:/shared-data -v $(PWD):/src -w /src golang:alpine sh -c "apk add --no-cache git make && go test -v -coverprofile=coverage-docker.out ./..." || (docker compose down && exit 1)
	@echo "Integration tests complete!"
	@echo "Stopping containers..."
	@docker compose down
	@echo "Local integration test complete"

all: clean lint test build ## Run all checks and build
	@echo "All tasks complete"
