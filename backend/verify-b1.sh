#!/bin/bash

# Script de verificación de la Fase B.1

echo "🔍 Verificando Fase B.1 - Estructura y Setup"
echo "=============================================="
echo ""

# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Contadores
SUCCESS=0
FAIL=0

# Función para verificar
check() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $2"
        ((SUCCESS++))
    else
        echo -e "${RED}✗${NC} $2"
        ((FAIL++))
    fi
}

# 1. Verificar go.mod
echo "📦 Verificando módulo Go..."
[ -f "go.mod" ]
check $? "go.mod existe"

grep -q "module github.com/aymc/backend" go.mod
check $? "Módulo correcto: github.com/aymc/backend"

# 2. Verificar estructura de directorios
echo ""
echo "📁 Verificando estructura de directorios..."
dirs=(
    "cmd/server"
    "config"
    "api/rest/handlers"
    "api/rest/middleware"
    "api/websocket"
    "api/grpc"
    "services/auth"
    "services/servers"
    "services/agents"
    "database/models"
    "database/migrations"
    "pkg/logger"
    "tests/integration"
)

for dir in "${dirs[@]}"; do
    [ -d "$dir" ]
    check $? "Directorio: $dir"
done

# 3. Verificar archivos principales
echo ""
echo "📝 Verificando archivos principales..."
files=(
    "cmd/server/main.go"
    "config/config.go"
    "config/config.yaml"
    "pkg/logger/logger.go"
    ".env.example"
    ".gitignore"
    "docker-compose.yml"
    "Dockerfile"
    "Makefile"
    "README.md"
)

for file in "${files[@]}"; do
    [ -f "$file" ]
    check $? "Archivo: $file"
done

# 4. Verificar dependencias críticas
echo ""
echo "🔧 Verificando dependencias críticas..."
deps=(
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "github.com/golang-jwt/jwt/v5"
    "go.uber.org/zap"
    "github.com/spf13/viper"
    "github.com/redis/go-redis/v9"
)

for dep in "${deps[@]}"; do
    grep -q "$dep" go.mod
    check $? "Dependencia: $dep"
done

# 5. Verificar compilación
echo ""
echo "🔨 Verificando compilación..."
if [ -f "bin/aymc-backend" ]; then
    size=$(ls -lh bin/aymc-backend | awk '{print $5}')
    echo -e "${GREEN}✓${NC} Binario compilado: $size"
    ((SUCCESS++))
else
    echo -e "${YELLOW}⚠${NC} Binario no encontrado, compilando..."
    go build -o bin/aymc-backend cmd/server/main.go 2>&1 | tail -5
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC} Compilación exitosa"
        ((SUCCESS++))
    else
        echo -e "${RED}✗${NC} Error de compilación"
        ((FAIL++))
    fi
fi

# 6. Verificar sintaxis de archivos
echo ""
echo "🔍 Verificando sintaxis..."
go vet ./... 2>&1 | head -5
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓${NC} go vet: sin errores"
    ((SUCCESS++))
else
    echo -e "${YELLOW}⚠${NC} go vet: advertencias detectadas"
fi

# Resumen
echo ""
echo "=============================================="
echo -e "📊 Resumen:"
echo -e "   ${GREEN}Exitosos:${NC} $SUCCESS"
echo -e "   ${RED}Fallidos:${NC} $FAIL"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}✅ FASE B.1 VERIFICADA EXITOSAMENTE${NC}"
    echo ""
    echo "Próximos pasos:"
    echo "  1. make docker-up    # Iniciar servicios"
    echo "  2. make run          # Ejecutar servidor"
    echo "  3. Comenzar Fase B.2 # Base de Datos"
    exit 0
else
    echo -e "${RED}❌ VERIFICACIÓN FALLIDA${NC}"
    echo "Por favor, revisa los errores anteriores."
    exit 1
fi
