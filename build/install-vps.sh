#!/bin/bash

#################################################
# AYMC VPS Installer
# Instala Backend + Agent en tu VPS/Servidor
# Soporta: Arch Linux, Debian/Ubuntu, RHEL/CentOS
#################################################

set -e  # Exit on error

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Funciones de log
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Verificar que se ejecuta como root
if [ "$EUID" -ne 0 ]; then 
    log_error "Este script debe ejecutarse como root (sudo)"
    exit 1
fi

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
echo "║           VPS Installer v1.0.0                    ║"
echo "║                                                   ║"
echo "╚═══════════════════════════════════════════════════╝"
echo ""

log_info "Iniciando instalación de AYMC..."
sleep 2

#################################################
# DETECTAR DISTRIBUCIÓN
#################################################

log_info "Detectando distribución de Linux..."

if [ -f /etc/os-release ]; then
    . /etc/os-release
    DISTRO=$ID
    VERSION=$VERSION_ID
    log_success "Distribución detectada: $PRETTY_NAME"
else
    log_error "No se pudo detectar la distribución"
    exit 1
fi

#################################################
# VARIABLES
#################################################

INSTALL_DIR="/opt/aymc"
DATA_DIR="/var/aymc"
CONFIG_DIR="/etc/aymc"
LOG_DIR="/var/log/aymc"
USER="aymc"
GROUP="aymc"

#################################################
# INSTALAR DEPENDENCIAS
#################################################

log_info "═══════════════════════════════════════"
log_info "  INSTALANDO DEPENDENCIAS"
log_info "═══════════════════════════════════════"
echo ""

case $DISTRO in
    arch|manjaro)
        log_info "Instalando dependencias para Arch Linux..."
        pacman -Sy --noconfirm --needed \
            postgresql \
            jdk-openjdk \
            wget \
            curl \
            tar \
            gzip
        
        # Inicializar PostgreSQL en Arch si es necesario
        if [ ! -d "/var/lib/postgres/data/base" ]; then
            log_info "Inicializando base de datos PostgreSQL..."
            su -l postgres -c "initdb --locale=C.UTF-8 --encoding=UTF8 -D '/var/lib/postgres/data'"
            if [ $? -eq 0 ]; then
                log_success "PostgreSQL inicializado"
            else
                log_error "Error al inicializar PostgreSQL"
                log_info "Ejecuta manualmente: sudo su -l postgres -c \"initdb --locale=C.UTF-8 --encoding=UTF8 -D '/var/lib/postgres/data'\""
                exit 1
            fi
        fi
        ;;
    
    ubuntu|debian)
        log_info "Instalando dependencias para Debian/Ubuntu..."
        apt-get update
        apt-get install -y \
            postgresql \
            postgresql-contrib \
            openjdk-21-jdk \
            wget \
            curl \
            tar \
            gzip
        ;;
    
    rhel|centos|fedora|rocky|almalinux)
        log_info "Instalando dependencias para RHEL/CentOS..."
        yum install -y \
            postgresql-server \
            postgresql-contrib \
            java-21-openjdk \
            wget \
            curl \
            tar \
            gzip
        
        # Inicializar PostgreSQL en RHEL si es necesario
        if [ ! -d "/var/lib/pgsql/data/base" ]; then
            log_info "Inicializando base de datos PostgreSQL..."
            postgresql-setup --initdb
        fi
        ;;
    
    *)
        log_error "Distribución no soportada: $DISTRO"
        log_info "Soportadas: Arch, Debian, Ubuntu, RHEL, CentOS, Fedora"
        exit 1
        ;;
esac

log_success "Dependencias instaladas"
echo ""

#################################################
# CREAR USUARIO Y DIRECTORIOS
#################################################

log_info "═══════════════════════════════════════"
log_info "  CONFIGURANDO SISTEMA"
log_info "═══════════════════════════════════════"
echo ""

# Crear usuario aymc
if ! id "$USER" &>/dev/null; then
    log_info "Creando usuario $USER..."
    useradd -r -s /bin/false -d "$DATA_DIR" "$USER"
    log_success "Usuario creado"
else
    log_warning "Usuario $USER ya existe"
fi

# Crear directorios
log_info "Creando estructura de directorios..."
mkdir -p "$INSTALL_DIR"/{backend,agent}
mkdir -p "$DATA_DIR"/{servers,backups,uploads}
mkdir -p "$CONFIG_DIR"
mkdir -p "$LOG_DIR"

# Permisos
chown -R "$USER:$GROUP" "$DATA_DIR"
chown -R "$USER:$GROUP" "$LOG_DIR"
chmod 755 "$INSTALL_DIR"
chmod 750 "$CONFIG_DIR"

log_success "Directorios creados"
echo ""

#################################################
# COPIAR BINARIOS
#################################################

log_info "═══════════════════════════════════════"
log_info "  INSTALANDO BINARIOS"
log_info "═══════════════════════════════════════"
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Copiar backend
log_info "Instalando backend..."
if [ -f "$SCRIPT_DIR/backend/aymc-backend" ]; then
    cp "$SCRIPT_DIR/backend/aymc-backend" "$INSTALL_DIR/backend/"
    chmod +x "$INSTALL_DIR/backend/aymc-backend"
    log_success "Backend instalado"
else
    log_error "No se encontró el binario del backend"
    exit 1
fi

# Copiar agent
log_info "Instalando agent..."
if [ -f "$SCRIPT_DIR/agent/aymc-agent" ]; then
    cp "$SCRIPT_DIR/agent/aymc-agent" "$INSTALL_DIR/agent/"
    chmod +x "$INSTALL_DIR/agent/aymc-agent"
    log_success "Agent instalado"
else
    log_error "No se encontró el binario del agent"
    exit 1
fi

# Copiar configuraciones
log_info "Instalando configuraciones..."
if [ -f "$SCRIPT_DIR/config/backend.env" ]; then
    cp "$SCRIPT_DIR/config/backend.env" "$CONFIG_DIR/backend.env"
    chmod 600 "$CONFIG_DIR/backend.env"
fi
if [ -f "$SCRIPT_DIR/config/agent.json" ]; then
    cp "$SCRIPT_DIR/config/agent.json" "$CONFIG_DIR/agent.json"
    chmod 600 "$CONFIG_DIR/agent.json"
fi

log_success "Configuraciones instaladas"
echo ""

#################################################
# CONFIGURAR POSTGRESQL
#################################################

log_info "═══════════════════════════════════════"
log_info "  CONFIGURANDO BASE DE DATOS"
log_info "═══════════════════════════════════════"
echo ""

# Iniciar PostgreSQL
log_info "Iniciando servicio PostgreSQL..."
systemctl enable postgresql
systemctl start postgresql
sleep 2

# Generar contraseña segura
DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)

# Crear usuario y base de datos
log_info "Creando base de datos 'aymc'..."
sudo -u postgres psql << EOF
-- Crear usuario
CREATE USER aymc WITH PASSWORD '$DB_PASSWORD';

-- Crear base de datos
CREATE DATABASE aymc OWNER aymc;

-- Dar permisos
GRANT ALL PRIVILEGES ON DATABASE aymc TO aymc;

-- Confirmar
\l aymc
EOF

if [ $? -eq 0 ]; then
    log_success "Base de datos creada"
    
    # Actualizar configuración del backend usando reemplazo seguro
    sed -i "s|^DB_PASSWORD=.*|DB_PASSWORD=${DB_PASSWORD}|" "$CONFIG_DIR/backend.env"
    
    log_success "Contraseña de base de datos configurada"
else
    log_warning "La base de datos puede que ya exista"
fi

echo ""

#################################################
# GENERAR JWT SECRET
#################################################

log_info "Generando JWT secret..."
JWT_SECRET=$(openssl rand -base64 64 | tr -d "=+/" | cut -c1-64)
# Usar | como delimitador en sed para evitar problemas con caracteres especiales
sed -i "s|^JWT_SECRET=.*|JWT_SECRET=${JWT_SECRET}|" "$CONFIG_DIR/backend.env"
log_success "JWT secret generado"
echo ""

#################################################
# CREAR SERVICIOS SYSTEMD
#################################################

log_info "═══════════════════════════════════════"
log_info "  CONFIGURANDO SERVICIOS"
log_info "═══════════════════════════════════════"
echo ""

# Servicio del Backend
log_info "Creando servicio aymc-backend..."
cat > /etc/systemd/system/aymc-backend.service << EOF
[Unit]
Description=AYMC Backend API Server
Documentation=https://github.com/aymc/aymc
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=$USER
Group=$GROUP
WorkingDirectory=$INSTALL_DIR/backend
EnvironmentFile=$CONFIG_DIR/backend.env
ExecStart=$INSTALL_DIR/backend/aymc-backend
Restart=always
RestartSec=10
StandardOutput=append:$LOG_DIR/backend.log
StandardError=append:$LOG_DIR/backend-error.log

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$DATA_DIR $LOG_DIR

# Limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF

# Servicio del Agent
log_info "Creando servicio aymc-agent..."
cat > /etc/systemd/system/aymc-agent.service << EOF
[Unit]
Description=AYMC Agent - Minecraft Server Manager
Documentation=https://github.com/aymc/aymc
After=network.target aymc-backend.service
Wants=aymc-backend.service

[Service]
Type=simple
User=$USER
Group=$GROUP
WorkingDirectory=$INSTALL_DIR/agent
ExecStart=$INSTALL_DIR/agent/aymc-agent -config $CONFIG_DIR/agent.json
Restart=always
RestartSec=10
StandardOutput=append:$LOG_DIR/agent.log
StandardError=append:$LOG_DIR/agent-error.log

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$DATA_DIR $LOG_DIR

# Limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF

# Recargar systemd
systemctl daemon-reload
log_success "Servicios creados"
echo ""

#################################################
# CONFIGURAR FIREWALL
#################################################

log_info "═══════════════════════════════════════"
log_info "  CONFIGURANDO FIREWALL"
log_info "═══════════════════════════════════════"
echo ""

# Detectar firewall
if command -v ufw &> /dev/null; then
    log_info "Configurando UFW..."
    ufw allow 8080/tcp comment 'AYMC Backend API'
    ufw allow 50051/tcp comment 'AYMC Agent gRPC'
    ufw allow 25565:25600/tcp comment 'Minecraft Servers'
    log_success "UFW configurado"
elif command -v firewall-cmd &> /dev/null; then
    log_info "Configurando firewalld..."
    firewall-cmd --permanent --add-port=8080/tcp
    firewall-cmd --permanent --add-port=50051/tcp
    firewall-cmd --permanent --add-port=25565-25600/tcp
    firewall-cmd --reload
    log_success "firewalld configurado"
else
    log_warning "No se detectó firewall (ufw/firewalld)"
    log_info "Asegúrate de abrir manualmente los puertos:"
    log_info "  - 8080/tcp (Backend API)"
    log_info "  - 50051/tcp (Agent gRPC)"
    log_info "  - 25565-25600/tcp (Servidores Minecraft)"
fi

echo ""

#################################################
# INICIAR SERVICIOS
#################################################

log_info "═══════════════════════════════════════"
log_info "  INICIANDO SERVICIOS"
log_info "═══════════════════════════════════════"
echo ""

# Habilitar e iniciar backend
log_info "Habilitando aymc-backend..."
systemctl enable aymc-backend
log_info "Iniciando aymc-backend..."
systemctl start aymc-backend
sleep 3

if systemctl is-active --quiet aymc-backend; then
    log_success "Backend iniciado correctamente"
else
    log_error "Error al iniciar el backend"
    log_info "Ver logs: journalctl -u aymc-backend -f"
fi

# Habilitar e iniciar agent
log_info "Habilitando aymc-agent..."
systemctl enable aymc-agent
log_info "Iniciando aymc-agent..."
systemctl start aymc-agent
sleep 3

if systemctl is-active --quiet aymc-agent; then
    log_success "Agent iniciado correctamente"
else
    log_error "Error al iniciar el agent"
    log_info "Ver logs: journalctl -u aymc-agent -f"
fi

echo ""

#################################################
# VERIFICAR INSTALACIÓN
#################################################

log_info "═══════════════════════════════════════"
log_info "  VERIFICANDO INSTALACIÓN"
log_info "═══════════════════════════════════════"
echo ""

# Verificar puerto del backend
log_info "Verificando puerto 8080..."
sleep 2
if ss -tlnp | grep -q ":8080"; then
    log_success "Backend escuchando en puerto 8080"
else
    log_warning "Backend no está escuchando en puerto 8080"
fi

# Verificar puerto del agent
log_info "Verificando puerto 50051..."
if ss -tlnp | grep -q ":50051"; then
    log_success "Agent escuchando en puerto 50051"
else
    log_warning "Agent no está escuchando en puerto 50051"
fi

echo ""

#################################################
# RESUMEN
#################################################

log_success "═══════════════════════════════════════"
log_success "  ¡INSTALACIÓN COMPLETADA!"
log_success "═══════════════════════════════════════"
echo ""
log_info "📁 Instalación:"
echo "   Backend: $INSTALL_DIR/backend/aymc-backend"
echo "   Agent:   $INSTALL_DIR/agent/aymc-agent"
echo ""
log_info "📁 Datos:"
echo "   Servidores: $DATA_DIR/servers"
echo "   Backups:    $DATA_DIR/backups"
echo "   Uploads:    $DATA_DIR/uploads"
echo ""
log_info "📁 Configuración:"
echo "   Backend: $CONFIG_DIR/backend.env"
echo "   Agent:   $CONFIG_DIR/agent.json"
echo ""
log_info "📁 Logs:"
echo "   Backend: $LOG_DIR/backend.log"
echo "   Agent:   $LOG_DIR/agent.log"
echo ""
log_info "🔐 Base de Datos:"
echo "   Database: aymc"
echo "   User:     aymc"
echo "   Password: $DB_PASSWORD"
echo "   (Guardada en: $CONFIG_DIR/backend.env)"
echo ""
log_info "🌐 Endpoints:"
echo "   Backend API:  http://$(hostname -I | awk '{print $1}'):8080"
echo "   Agent gRPC:   $(hostname -I | awk '{print $1}'):50051"
echo ""
log_info "🔧 Comandos útiles:"
echo "   Ver logs backend: journalctl -u aymc-backend -f"
echo "   Ver logs agent:   journalctl -u aymc-agent -f"
echo "   Reiniciar backend: systemctl restart aymc-backend"
echo "   Reiniciar agent:   systemctl restart aymc-agent"
echo "   Estado:           systemctl status aymc-backend aymc-agent"
echo ""
log_info "📝 Siguiente paso:"
echo "   Configura tu aplicación frontend con la URL:"
echo "   http://$(hostname -I | awk '{print $1}'):8080"
echo ""
log_warning "⚠️  IMPORTANTE:"
echo "   1. Cambia las contraseñas en: $CONFIG_DIR/backend.env"
echo "   2. Configura CORS_ORIGINS según tu frontend"
echo "   3. Considera usar HTTPS en producción (Nginx + Let's Encrypt)"
echo ""
log_success "¡Disfruta de AYMC! 🎉"
echo ""
