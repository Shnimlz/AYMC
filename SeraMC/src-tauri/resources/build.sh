#!/bin/bash

#################################################
# AYMC Build Script
# Compila backend y agent para producción
#################################################

set -e  # Exit on error

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Función de log
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Banner
echo ""
echo "╔═══════════════════════════════════════════════════╗"
echo "║                                                   ║"
echo "║     █████╗ ██╗   ██╗███╗   ███╗ ██████╗         ║"
echo "║    ██╔══██╗╚██╗ ██╔╝████╗ ████║██╔════╝         ║"
echo "║    ███████║ ╚████╔╝ ██╔████╔██║██║              ║"
echo "║    ██╔══██║  ╚██╔╝  ██║╚██╔╝██║██║              ║"
echo "║    ██║  ██║   ██║   ██║ ╚═╝ ██║╚██████╗         ║"
echo "║    ╚═╝  ╚═╝   ╚═╝   ╚═╝     ╚═╝ ╚═════╝         ║"
echo "║                                                   ║"
echo "║            Build Script v1.0.0                    ║"
echo "║                                                   ║"
echo "╚═══════════════════════════════════════════════════╝"
echo ""

# Variables
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="${PROJECT_ROOT}/build"
VERSION=$(date +%Y.%m.%d)
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

log_info "Directorio del proyecto: ${PROJECT_ROOT}"
log_info "Directorio de build: ${BUILD_DIR}"
log_info "Versión: ${VERSION}"
log_info "Commit: ${GIT_COMMIT}"
echo ""

# Verificar Go
log_info "Verificando instalación de Go..."
if ! command -v go &> /dev/null; then
    log_error "Go no está instalado. Por favor instálalo desde https://golang.org/dl/"
    exit 1
fi
GO_VERSION=$(go version | awk '{print $3}')
log_success "Go encontrado: ${GO_VERSION}"
echo ""

# Crear directorio de build
log_info "Preparando directorio de build..."
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"/{backend,agent,config}
log_success "Directorio de build creado"
echo ""

#################################################
# COMPILAR BACKEND
#################################################

log_info "═══════════════════════════════════════"
log_info "  COMPILANDO BACKEND"
log_info "═══════════════════════════════════════"
echo ""

cd "${PROJECT_ROOT}/backend"

# Verificar dependencias
log_info "Descargando dependencias del backend..."
go mod download
go mod verify
log_success "Dependencias verificadas"

# Build flags
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# Compilar
log_info "Compilando binario del backend..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "${LDFLAGS}" \
    -o "${BUILD_DIR}/backend/aymc-backend" \
    ./cmd/server/main.go

if [ $? -eq 0 ]; then
    log_success "Backend compilado exitosamente"
    BACKEND_SIZE=$(du -h "${BUILD_DIR}/backend/aymc-backend" | cut -f1)
    log_info "Tamaño del binario: ${BACKEND_SIZE}"
else
    log_error "Error al compilar el backend"
    exit 1
fi
echo ""

#################################################
# COMPILAR AGENT
#################################################

log_info "═══════════════════════════════════════"
log_info "  COMPILANDO AGENT"
log_info "═══════════════════════════════════════"
echo ""

cd "${PROJECT_ROOT}/agent"

# Verificar dependencias
log_info "Descargando dependencias del agent..."
go mod download
go mod verify
log_success "Dependencias verificadas"

# Build flags
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# Compilar
log_info "Compilando binario del agent..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "${LDFLAGS}" \
    -o "${BUILD_DIR}/agent/aymc-agent" \
    ./main.go

if [ $? -eq 0 ]; then
    log_success "Agent compilado exitosamente"
    AGENT_SIZE=$(du -h "${BUILD_DIR}/agent/aymc-agent" | cut -f1)
    log_info "Tamaño del binario: ${AGENT_SIZE}"
else
    log_error "Error al compilar el agent"
    exit 1
fi
echo ""

#################################################
# COPIAR ARCHIVOS DE CONFIGURACIÓN
#################################################

log_info "═══════════════════════════════════════"
log_info "  COPIANDO ARCHIVOS"
log_info "═══════════════════════════════════════"
echo ""

# Copiar configs de ejemplo
log_info "Copiando archivos de configuración..."

# Backend config
cat > "${BUILD_DIR}/config/backend.env" << 'EOF'
# Backend Configuration
APP_ENV=production
APP_PORT=8080
APP_LOG_LEVEL=info

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=aymc
DB_PASSWORD=CHANGE_THIS_PASSWORD
DB_NAME=aymc
DB_SSL_MODE=disable

# JWT
JWT_SECRET=CHANGE_THIS_SECRET_KEY_TO_SOMETHING_RANDOM
JWT_EXPIRES_IN=24h
JWT_REFRESH_EXPIRES_IN=168h

# CORS
CORS_ORIGINS=http://localhost:1420,tauri://localhost

# Agent Communication
AGENT_TIMEOUT=30s

# File Upload
MAX_UPLOAD_SIZE=104857600
UPLOAD_DIR=/var/aymc/uploads
EOF

# Agent config
cat > "${BUILD_DIR}/config/agent.json" << 'EOF'
{
  "agent_id": "agent-1",
  "backend_url": "http://localhost:8080",
  "port": 50051,
  "log_level": "info",
  "max_servers": 50,
  "java_path": "/usr/bin/java",
  "work_dir": "/var/aymc/servers",
  "enable_metrics": true,
  "metrics_interval": 30000000000,
  "custom_env": {}
}
EOF

log_success "Archivos de configuración creados"
echo ""

# Copiar scripts
log_info "Copiando scripts de instalación..."
cp "${PROJECT_ROOT}/scripts/install-vps.sh" "${BUILD_DIR}/"
cp "${PROJECT_ROOT}/scripts/uninstall.sh" "${BUILD_DIR}/"
chmod +x "${BUILD_DIR}"/*.sh
log_success "Scripts copiados"
echo ""

#################################################
# CREAR TARBALL
#################################################

log_info "═══════════════════════════════════════"
log_info "  CREANDO PAQUETE"
log_info "═══════════════════════════════════════"
echo ""

cd "${PROJECT_ROOT}"
TARBALL="aymc-${VERSION}-linux-amd64.tar.gz"

log_info "Creando tarball: ${TARBALL}..."
tar -czf "${BUILD_DIR}/${TARBALL}" -C "${BUILD_DIR}" \
    backend/aymc-backend \
    agent/aymc-agent \
    config/ \
    install-vps.sh \
    uninstall.sh

TARBALL_SIZE=$(du -h "${BUILD_DIR}/${TARBALL}" | cut -f1)
log_success "Tarball creado: ${TARBALL} (${TARBALL_SIZE})"
echo ""

#################################################
# CREAR CHECKSUMS
#################################################

log_info "Generando checksums..."
cd "${BUILD_DIR}"
sha256sum "${TARBALL}" > "${TARBALL}.sha256"
log_success "Checksum SHA256 creado"
echo ""

#################################################
# RESUMEN
#################################################

log_success "═══════════════════════════════════════"
log_success "  BUILD COMPLETADO"
log_success "═══════════════════════════════════════"
echo ""
log_info "Archivos generados:"
echo "  📦 Paquete: ${BUILD_DIR}/${TARBALL}"
echo "  🔒 Checksum: ${BUILD_DIR}/${TARBALL}.sha256"
echo "  🖥️  Backend: ${BUILD_DIR}/backend/aymc-backend (${BACKEND_SIZE})"
echo "  🤖 Agent: ${BUILD_DIR}/agent/aymc-agent (${AGENT_SIZE})"
echo ""
log_info "Para instalar en tu VPS:"
echo "  1. Copia el tarball al servidor:"
echo "     scp ${TARBALL} user@your-vps:/tmp/"
echo ""
echo "  2. En el servidor, extrae e instala:"
echo "     cd /tmp"
echo "     tar -xzf ${TARBALL}"
echo "     sudo ./install-vps.sh"
echo ""
log_success "¡Build completado exitosamente! 🎉"
