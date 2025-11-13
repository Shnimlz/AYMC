.PHONY: all build clean test proto install run dev help

# Variables
BINARY_NAME=aymc-agent
VERSION=0.1.0
BUILD_DIR=bin
PROTO_DIR=proto
GRPC_DIR=grpc/pb
GO=go
PROTOC=protoc

# Rutas
INSTALL_PATH_LINUX=/opt/aymc
CONFIG_PATH_LINUX=/etc/aymc

# Colores para output
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m

all: clean proto build

## help: Muestra esta ayuda
help:
	@echo "Makefile para AYMC Agent"
	@echo ""
	@echo "Uso: make [target]"
	@echo ""
	@echo "Targets disponibles:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /'
	@echo ""

## build: Compila el agente para la plataforma actual
build:
	@echo "$(GREEN)Compilando $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags="-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "$(GREEN)✓ Compilación exitosa: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## build-all: Compila para todas las plataformas
build-all: clean proto
	@echo "$(GREEN)Compilando para todas las plataformas...$(NC)"
	@mkdir -p $(BUILD_DIR)
	
	@echo "  → Linux amd64"
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	
	@echo "  → Linux arm64"
	GOOS=linux GOARCH=arm64 $(GO) build -ldflags="-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 main.go
	
	@echo "  → Windows amd64"
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags="-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go
	
	@echo "  → macOS amd64"
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	
	@echo "  → macOS arm64"
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	
	@echo "$(GREEN)✓ Compilación multi-plataforma completada$(NC)"

## proto: Genera código a partir de archivos .proto
proto:
	@echo "$(GREEN)Generando código protobuf...$(NC)"
	@mkdir -p $(GRPC_DIR)
	@export PATH="$$PATH:$$(go env GOPATH)/bin" && \
	$(PROTOC) --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/agent.proto
	@echo "$(GREEN)✓ Código protobuf generado$(NC)"

## test: Ejecuta tests unitarios
test:
	@echo "$(GREEN)Ejecutando tests...$(NC)"
	$(GO) test -v ./...

## test-coverage: Ejecuta tests con reporte de cobertura
test-coverage:
	@echo "$(GREEN)Ejecutando tests con cobertura...$(NC)"
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Reporte generado: coverage.html$(NC)"

## clean: Limpia archivos generados
clean:
	@echo "$(YELLOW)Limpiando...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Limpieza completada$(NC)"

## install: Instala el agente en el sistema (Linux)
install: build
	@echo "$(GREEN)Instalando $(BINARY_NAME)...$(NC)"
	@sudo mkdir -p $(INSTALL_PATH_LINUX)
	@sudo mkdir -p $(CONFIG_PATH_LINUX)
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH_LINUX)/
	@sudo chmod +x $(INSTALL_PATH_LINUX)/$(BINARY_NAME)
	@echo "$(GREEN)✓ Agente instalado en $(INSTALL_PATH_LINUX)$(NC)"
	@echo "$(YELLOW)Ejecuta 'make install-service' para instalar el servicio systemd$(NC)"

## install-service: Instala el servicio systemd (Linux)
install-service:
	@echo "$(GREEN)Instalando servicio systemd...$(NC)"
	@sudo bash installer/install_agent.sh
	@echo "$(GREEN)✓ Servicio instalado$(NC)"

## run: Ejecuta el agente en modo desarrollo
run: build
	@echo "$(GREEN)Ejecutando $(BINARY_NAME)...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) --debug --config=./agent.json

## dev: Ejecuta el agente en modo watch (requiere air)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(YELLOW)air no instalado. Instalando...$(NC)"; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

## lint: Ejecuta linters de código
lint:
	@echo "$(GREEN)Ejecutando linters...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)golangci-lint no instalado$(NC)"; \
		echo "Instala con: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin"; \
	fi

## fmt: Formatea el código
fmt:
	@echo "$(GREEN)Formateando código...$(NC)"
	$(GO) fmt ./...
	@echo "$(GREEN)✓ Código formateado$(NC)"

## deps: Descarga e instala dependencias
deps:
	@echo "$(GREEN)Descargando dependencias...$(NC)"
	$(GO) mod download
	$(GO) mod tidy
	@echo "$(GREEN)✓ Dependencias instaladas$(NC)"

## docker-build: Construye imagen Docker
docker-build:
	@echo "$(GREEN)Construyendo imagen Docker...$(NC)"
	docker build -t aymc/agent:$(VERSION) .
	docker tag aymc/agent:$(VERSION) aymc/agent:latest
	@echo "$(GREEN)✓ Imagen Docker construida$(NC)"

## version: Muestra la versión
version:
	@echo "AYMC Agent v$(VERSION)"
