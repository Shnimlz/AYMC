.PHONY: help run build test clean docker-up docker-down docker-build migrate-up migrate-down seed lint fmt

# Variables
APP_NAME=aymc-backend
BINARY_DIR=bin
BINARY_PATH=$(BINARY_DIR)/$(APP_NAME)
DOCKER_IMAGE=aymc/backend:latest

# Colors for output
GREEN=\033[0;32m
NC=\033[0m # No Color

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${GREEN}%-15s${NC} %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the application locally
	@echo "${GREEN}Starting application...${NC}"
	go run cmd/server/main.go

build: ## Build the application binary
	@echo "${GREEN}Building application...${NC}"
	@mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_PATH) cmd/server/main.go
	@echo "${GREEN}Binary built at $(BINARY_PATH)${NC}"

test: ## Run tests
	@echo "${GREEN}Running tests...${NC}"
	go test -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	@echo "${GREEN}Running tests with coverage...${NC}"
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "${GREEN}Coverage report generated at coverage.html${NC}"

bench: ## Run benchmarks
	@echo "${GREEN}Running benchmarks...${NC}"
	go test -bench=. -benchmem ./...

clean: ## Clean build artifacts
	@echo "${GREEN}Cleaning...${NC}"
	rm -rf $(BINARY_DIR)
	rm -f coverage.txt coverage.html
	go clean -testcache

docker-up: ## Start all services with Docker Compose
	@echo "${GREEN}Starting Docker services...${NC}"
	docker-compose up -d
	@echo "${GREEN}Services started. Access:${NC}"
	@echo "  - Backend: http://localhost:8080"
	@echo "  - Adminer: http://localhost:8081"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Redis: localhost:6379"

docker-down: ## Stop all Docker services
	@echo "${GREEN}Stopping Docker services...${NC}"
	docker-compose down

docker-clean: ## Stop and remove all Docker volumes
	@echo "${GREEN}Cleaning Docker resources...${NC}"
	docker-compose down -v
	docker system prune -f

docker-build: ## Build Docker image
	@echo "${GREEN}Building Docker image...${NC}"
	docker build -t $(DOCKER_IMAGE) .

docker-logs: ## Show Docker logs
	docker-compose logs -f backend

migrate-up: ## Run database migrations
	@echo "${GREEN}Running migrations...${NC}"
	go run cmd/db/main.go migrate -up

migrate-down: ## Rollback database migrations
	@echo "${GREEN}Rolling back migrations...${NC}"
	go run cmd/db/main.go migrate -down

seed: ## Seed the database with test data
	@echo "${GREEN}Seeding database...${NC}"
	go run cmd/db/main.go seed

lint: ## Run linters
	@echo "${GREEN}Running linters...${NC}"
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	golangci-lint run ./...

fmt: ## Format code
	@echo "${GREEN}Formatting code...${NC}"
	go fmt ./...
	gofmt -s -w .

vet: ## Run go vet
	@echo "${GREEN}Running go vet...${NC}"
	go vet ./...

tidy: ## Tidy go modules
	@echo "${GREEN}Tidying go modules...${NC}"
	go mod tidy

deps: ## Download dependencies
	@echo "${GREEN}Downloading dependencies...${NC}"
	go mod download

install-tools: ## Install development tools
	@echo "${GREEN}Installing development tools...${NC}"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest

swagger: ## Generate Swagger documentation
	@echo "${GREEN}Generating Swagger docs...${NC}"
	@command -v swag >/dev/null 2>&1 || { echo "swag not installed. Run: make install-tools"; exit 1; }
	swag init -g cmd/server/main.go -o ./docs

dev: docker-up run ## Start Docker services and run the application

all: clean fmt vet test build ## Run all checks and build

.DEFAULT_GOAL := help
