#!/bin/bash

###############################################################################
# AYMC Agent Installer - Script de instalación para Linux/Unix
# Este script instala y configura el agente AYMC en sistemas Linux/Unix
###############################################################################

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Variables
AGENT_VERSION="0.1.0"
INSTALL_DIR="/opt/aymc"
CONFIG_DIR="/etc/aymc"
LOG_DIR="/var/log/aymc"
SERVICE_NAME="aymc-agent"
DOWNLOAD_URL="https://github.com/aymc/agent/releases/download/v${AGENT_VERSION}/aymc-agent-linux-amd64"

# Funciones auxiliares
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "Este script debe ejecutarse como root o con sudo"
        exit 1
    fi
}

detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
    else
        print_error "No se pudo detectar el sistema operativo"
        exit 1
    fi
    print_info "Sistema detectado: $OS $VERSION"
}

check_dependencies() {
    print_info "Verificando dependencias..."
    
    # Verificar curl o wget
    if ! command -v curl &> /dev/null && ! command -v wget &> /dev/null; then
        print_error "Se requiere curl o wget"
        exit 1
    fi
    
    # Verificar systemd
    if ! command -v systemctl &> /dev/null; then
        print_warn "systemd no encontrado, instalación de servicio omitida"
        INSTALL_SERVICE=false
    else
        INSTALL_SERVICE=true
    fi
}

install_java() {
    print_info "Verificando instalación de Java..."
    
    if command -v java &> /dev/null; then
        JAVA_VERSION=$(java -version 2>&1 | head -n 1 | cut -d'"' -f2)
        print_info "Java ya está instalado: $JAVA_VERSION"
        return 0
    fi
    
    print_info "Java no encontrado. Instalando OpenJDK 17..."
    
    case $OS in
        ubuntu|debian)
            apt-get update
            apt-get install -y openjdk-17-jre-headless
            ;;
        centos|rhel|fedora)
            yum install -y java-17-openjdk-headless
            ;;
        arch)
            pacman -S --noconfirm jre17-openjdk-headless
            ;;
        *)
            print_warn "Sistema no soportado para instalación automática de Java"
            print_warn "Por favor instale Java manualmente"
            ;;
    esac
}

install_screen() {
    print_info "Verificando instalación de screen..."
    
    if command -v screen &> /dev/null; then
        print_info "screen ya está instalado"
        return 0
    fi
    
    print_info "Instalando screen..."
    
    case $OS in
        ubuntu|debian)
            apt-get install -y screen
            ;;
        centos|rhel|fedora)
            yum install -y screen
            ;;
        arch)
            pacman -S --noconfirm screen
            ;;
        *)
            print_warn "No se pudo instalar screen automáticamente"
            ;;
    esac
}

create_directories() {
    print_info "Creando directorios..."
    
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$LOG_DIR"
    mkdir -p "$INSTALL_DIR/servers"
    
    chmod 755 "$INSTALL_DIR"
    chmod 700 "$CONFIG_DIR"
    chmod 755 "$LOG_DIR"
}

download_agent() {
    print_info "Descargando agente AYMC v${AGENT_VERSION}..."
    
    # Por ahora creamos un placeholder
    # TODO: Descomentar cuando haya releases
    # if command -v curl &> /dev/null; then
    #     curl -L -o "$INSTALL_DIR/aymc-agent" "$DOWNLOAD_URL"
    # else
    #     wget -O "$INSTALL_DIR/aymc-agent" "$DOWNLOAD_URL"
    # fi
    
    print_warn "Descarga omitida - usando binario local para desarrollo"
    
    chmod +x "$INSTALL_DIR/aymc-agent"
}

create_config() {
    print_info "Creando configuración por defecto..."
    
    cat > "$CONFIG_DIR/agent.json" <<EOF
{
  "agent_id": "$(uuidgen 2>/dev/null || echo "agent-$(date +%s)")",
  "backend_url": "localhost:50050",
  "port": 50051,
  "log_level": "info",
  "max_servers": 10,
  "java_path": "$(which java)",
  "work_dir": "$INSTALL_DIR/servers",
  "enable_metrics": true,
  "metrics_interval": "5s",
  "custom_env": {}
}
EOF
    
    chmod 600 "$CONFIG_DIR/agent.json"
}

create_systemd_service() {
    if [ "$INSTALL_SERVICE" = false ]; then
        return 0
    fi
    
    print_info "Creando servicio systemd..."
    
    cat > "/etc/systemd/system/${SERVICE_NAME}.service" <<EOF
[Unit]
Description=AYMC Agent - Advanced Minecraft Control Agent
After=network.target

[Service]
Type=simple
User=root
ExecStart=$INSTALL_DIR/aymc-agent --config=$CONFIG_DIR/agent.json
Restart=always
RestartSec=10
StandardOutput=append:$LOG_DIR/agent.log
StandardError=append:$LOG_DIR/agent-error.log

# Seguridad
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    print_info "Servicio systemd creado"
}

configure_firewall() {
    print_info "Configurando firewall..."
    
    if command -v ufw &> /dev/null; then
        ufw allow 50051/tcp comment "AYMC Agent gRPC"
        print_info "Regla UFW añadida"
    elif command -v firewall-cmd &> /dev/null; then
        firewall-cmd --permanent --add-port=50051/tcp
        firewall-cmd --reload
        print_info "Regla firewalld añadida"
    else
        print_warn "No se detectó firewall. Configure manualmente el puerto 50051/tcp"
    fi
}

print_summary() {
    echo ""
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║                                                            ║"
    echo "║  ✓ AYMC Agent instalado correctamente                     ║"
    echo "║                                                            ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""
    echo "Directorios:"
    echo "  - Instalación: $INSTALL_DIR"
    echo "  - Configuración: $CONFIG_DIR"
    echo "  - Logs: $LOG_DIR"
    echo ""
    
    if [ "$INSTALL_SERVICE" = true ]; then
        echo "Comandos útiles:"
        echo "  - Iniciar servicio:  systemctl start $SERVICE_NAME"
        echo "  - Detener servicio:  systemctl stop $SERVICE_NAME"
        echo "  - Estado:            systemctl status $SERVICE_NAME"
        echo "  - Habilitar inicio:  systemctl enable $SERVICE_NAME"
        echo "  - Ver logs:          journalctl -u $SERVICE_NAME -f"
    else
        echo "Iniciar manualmente:"
        echo "  $INSTALL_DIR/aymc-agent --config=$CONFIG_DIR/agent.json"
    fi
    echo ""
}

# Main
main() {
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║     AYMC Agent Installer v${AGENT_VERSION}                          ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""
    
    check_root
    detect_os
    check_dependencies
    install_java
    install_screen
    create_directories
    download_agent
    create_config
    create_systemd_service
    configure_firewall
    print_summary
    
    print_info "Instalación completada"
}

main "$@"
