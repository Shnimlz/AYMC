#!/bin/bash

#################################################
# AYMC SeraMC - Windows Build Script
# Compila la aplicación Tauri para Windows
#################################################

set -e

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Banner
clear
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
echo "║           Windows Build Script v1.0.0             ║"
echo "║                                                   ║"
echo "╚═══════════════════════════════════════════════════╝"
echo ""

log_info "Iniciando build para Windows..."
sleep 2

#################################################
# VERIFICAR REQUISITOS
#################################################

log_info "═══════════════════════════════════════"
log_info "  VERIFICANDO REQUISITOS"
log_info "═══════════════════════════════════════"
echo ""

# Verificar Node.js
if ! command -v node &> /dev/null; then
    log_error "Node.js no está instalado"
    log_info "Instala Node.js desde: https://nodejs.org/"
    exit 1
fi
NODE_VERSION=$(node -v)
log_success "Node.js instalado: $NODE_VERSION"

# Verificar npm
if ! command -v npm &> /dev/null; then
    log_error "npm no está instalado"
    exit 1
fi
NPM_VERSION=$(npm -v)
log_success "npm instalado: $NPM_VERSION"

# Verificar Rust
if ! command -v cargo &> /dev/null; then
    log_error "Rust no está instalado"
    log_info "Instala Rust desde: https://rustup.rs/"
    exit 1
fi
RUST_VERSION=$(rustc --version)
log_success "Rust instalado: $RUST_VERSION"

# Verificar target Windows
log_info "Verificando target x86_64-pc-windows-gnu..."
if ! rustup target list | grep -q "x86_64-pc-windows-gnu (installed)"; then
    log_warning "Target Windows no instalado, instalando..."
    rustup target add x86_64-pc-windows-gnu
    log_success "Target Windows instalado"
else
    log_success "Target Windows ya instalado"
fi

# Verificar MinGW
if ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    log_error "MinGW-w64 no está instalado"
    log_info "Instala con:"
    echo "  Arch: sudo pacman -S mingw-w64-gcc"
    echo "  Ubuntu: sudo apt install mingw-w64"
    exit 1
fi
log_success "MinGW-w64 instalado"

echo ""

#################################################
# LIMPIAR BUILD ANTERIOR
#################################################

log_info "═══════════════════════════════════════"
log_info "  LIMPIANDO BUILD ANTERIOR"
log_info "═══════════════════════════════════════"
echo ""

if [ -d "dist" ]; then
    log_info "Eliminando dist/..."
    rm -rf dist/
fi

if [ -d "src-tauri/target/x86_64-pc-windows-gnu" ]; then
    log_info "Eliminando target Windows anterior..."
    rm -rf src-tauri/target/x86_64-pc-windows-gnu/release/bundle/
fi

log_success "Limpieza completada"
echo ""

#################################################
# INSTALAR DEPENDENCIAS
#################################################

log_info "═══════════════════════════════════════"
log_info "  INSTALANDO DEPENDENCIAS"
log_info "═══════════════════════════════════════"
echo ""

log_info "Instalando dependencias de Node.js..."
npm install
log_success "Dependencias instaladas"
echo ""

#################################################
# BUILD FRONTEND
#################################################

log_info "═══════════════════════════════════════"
log_info "  COMPILANDO FRONTEND"
log_info "═══════════════════════════════════════"
echo ""

log_info "Ejecutando vue-tsc y vite build..."
npm run build
log_success "Frontend compilado exitosamente"
echo ""

#################################################
# BUILD TAURI PARA WINDOWS
#################################################

log_info "═══════════════════════════════════════"
log_info "  COMPILANDO TAURI PARA WINDOWS"
log_info "═══════════════════════════════════════"
echo ""

log_info "Esto puede tomar 10-15 minutos en la primera compilación..."
log_info "Compilaciones subsecuentes serán más rápidas gracias a cache."
echo ""

# Configurar linker para Windows
export CARGO_TARGET_X86_64_PC_WINDOWS_GNU_LINKER=x86_64-w64-mingw32-gcc

# Build
npm run tauri build -- --target x86_64-pc-windows-gnu

if [ $? -eq 0 ]; then
    log_success "Build de Tauri completado exitosamente"
else
    log_error "Error en el build de Tauri"
    exit 1
fi

echo ""

#################################################
# VERIFICAR RESULTADOS
#################################################

log_info "═══════════════════════════════════════"
log_info "  VERIFICANDO ARCHIVOS GENERADOS"
log_info "═══════════════════════════════════════"
echo ""

TARGET_DIR="src-tauri/target/x86_64-pc-windows-gnu/release"

# Verificar ejecutable
if [ -f "$TARGET_DIR/seramc.exe" ]; then
    EXE_SIZE=$(du -h "$TARGET_DIR/seramc.exe" | cut -f1)
    log_success "Ejecutable generado: seramc.exe ($EXE_SIZE)"
else
    log_error "No se generó seramc.exe"
fi

# Verificar instalador MSI
MSI_FILES=$(find "$TARGET_DIR/bundle/msi" -name "*.msi" 2>/dev/null)
if [ -n "$MSI_FILES" ]; then
    for msi in $MSI_FILES; do
        MSI_SIZE=$(du -h "$msi" | cut -f1)
        MSI_NAME=$(basename "$msi")
        log_success "Instalador MSI: $MSI_NAME ($MSI_SIZE)"
    done
else
    log_warning "No se generó instalador MSI"
fi

# Verificar instalador NSIS
NSIS_FILES=$(find "$TARGET_DIR/bundle/nsis" -name "*setup.exe" 2>/dev/null)
if [ -n "$NSIS_FILES" ]; then
    for nsis in $NSIS_FILES; do
        NSIS_SIZE=$(du -h "$nsis" | cut -f1)
        NSIS_NAME=$(basename "$nsis")
        log_success "Instalador NSIS: $NSIS_NAME ($NSIS_SIZE)"
    done
else
    log_warning "No se generó instalador NSIS"
fi

echo ""

#################################################
# CREAR PAQUETE DE DISTRIBUCIÓN
#################################################

log_info "═══════════════════════════════════════"
log_info "  CREANDO PAQUETE DE DISTRIBUCIÓN"
log_info "═══════════════════════════════════════"
echo ""

DIST_DIR="dist-windows"
VERSION=$(grep '"version"' src-tauri/tauri.conf.json | head -1 | sed 's/.*: "\(.*\)".*/\1/')

log_info "Versión detectada: $VERSION"

# Crear directorio de distribución
mkdir -p "$DIST_DIR"

# Copiar ejecutable
if [ -f "$TARGET_DIR/seramc.exe" ]; then
    cp "$TARGET_DIR/seramc.exe" "$DIST_DIR/"
    log_success "Ejecutable copiado"
fi

# Copiar instaladores MSI
if [ -n "$MSI_FILES" ]; then
    for msi in $MSI_FILES; do
        cp "$msi" "$DIST_DIR/"
    done
    log_success "Instaladores MSI copiados"
fi

# Copiar instaladores NSIS
if [ -n "$NSIS_FILES" ]; then
    for nsis in $NSIS_FILES; do
        cp "$nsis" "$DIST_DIR/"
    done
    log_success "Instaladores NSIS copiados"
fi

# Crear README
cat > "$DIST_DIR/README.txt" << EOF
AYMC SeraMC - Windows Distribution Package
Version: $VERSION
Build Date: $(date '+%Y-%m-%d %H:%M:%S')

Archivos incluidos:
- seramc.exe: Ejecutable standalone (no requiere instalación)
- *.msi: Instalador Windows MSI (recomendado para empresas)
- *setup.exe: Instalador NSIS (recomendado para usuarios finales)

Instalación:
1. Ejecuta el instalador MSI o NSIS
2. Sigue las instrucciones del wizard
3. La aplicación se instalará en C:\Program Files\AYMC SeraMC

Ejecución Portable:
1. Copia seramc.exe a cualquier carpeta
2. Ejecuta directamente sin instalación
3. Los datos se guardarán en: %APPDATA%\com.shni.aymc.seramc

Requisitos del Sistema:
- Windows 10/11 (64-bit)
- 4 GB RAM (mínimo)
- 200 MB espacio en disco
- WebView2 Runtime (incluido en Windows 11)

Soporte:
- GitHub: https://github.com/Shnimlz/AYMC
- Issues: https://github.com/Shnimlz/AYMC/issues
- Email: tu@email.com

Licencia: MIT
Copyright (c) 2025 AYMC Team
EOF

log_success "README.txt creado"

# Crear archivo ZIP
log_info "Creando archivo ZIP..."
ZIP_NAME="AYMC-SeraMC-v${VERSION}-Windows-x64.zip"
cd "$DIST_DIR"
zip -r "../$ZIP_NAME" ./*
cd ..

if [ -f "$ZIP_NAME" ]; then
    ZIP_SIZE=$(du -h "$ZIP_NAME" | cut -f1)
    log_success "Paquete ZIP creado: $ZIP_NAME ($ZIP_SIZE)"
else
    log_error "Error al crear ZIP"
fi

echo ""

#################################################
# RESUMEN
#################################################

log_success "═══════════════════════════════════════"
log_success "  ¡BUILD COMPLETADO EXITOSAMENTE!"
log_success "═══════════════════════════════════════"
echo ""
log_info "📦 Archivos de Distribución:"
echo ""
echo "   Carpeta: $DIST_DIR/"
echo "   Paquete: $ZIP_NAME"
echo ""
log_info "📁 Estructura generada:"
tree -h "$DIST_DIR" 2>/dev/null || ls -lh "$DIST_DIR"
echo ""
log_info "🚀 Próximos Pasos:"
echo ""
echo "   1. Probar el ejecutable en Windows:"
echo "      - Copia seramc.exe a Windows"
echo "      - Ejecuta con doble clic"
echo ""
echo "   2. Probar instaladores:"
echo "      - MSI: Para distribución corporativa"
echo "      - NSIS: Para usuarios finales"
echo ""
echo "   3. Distribución:"
echo "      - Subir $ZIP_NAME a GitHub Releases"
echo "      - Compartir link de descarga"
echo "      - Documentar proceso de instalación"
echo ""
log_warning "⚠️  IMPORTANTE:"
echo "   - El ejecutable no está firmado digitalmente"
echo "   - Windows SmartScreen puede mostrar advertencia"
echo "   - Considera firmar con certificado code-signing"
echo ""
log_info "📚 Documentación completa:"
echo "   docs/BUILD_WINDOWS.md"
echo ""
log_success "¡Disfruta de AYMC SeraMC! 🎉"
echo ""
