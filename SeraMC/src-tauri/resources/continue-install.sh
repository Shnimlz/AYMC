#!/bin/bash

#################################################
# AYMC Installation - Continue from PostgreSQL
# Continúa la instalación desde la configuración
# de la base de datos
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

# Verificar root
if [ "$EUID" -ne 0 ]; then 
    log_error "Este script debe ejecutarse como root (sudo)"
    exit 1
fi

CONFIG_DIR="/etc/aymc"

echo ""
log_info "═══════════════════════════════════════"
log_info "  CONTINUANDO INSTALACIÓN DE AYMC"
log_info "═══════════════════════════════════════"
echo ""

#################################################
# VERIFICAR POSTGRESQL
#################################################

log_info "Verificando PostgreSQL..."
if ! systemctl is-active --quiet postgresql; then
    log_error "PostgreSQL no está corriendo"
    log_info "Ejecuta primero:"
    echo "  sudo su -l postgres -c \"initdb --locale=C.UTF-8 --encoding=UTF8 -D '/var/lib/postgres/data'\""
    echo "  sudo systemctl start postgresql"
    exit 1
fi
log_success "PostgreSQL está corriendo"
echo ""

#################################################
# CONFIGURAR BASE DE DATOS
#################################################

log_info "═══════════════════════════════════════"
log_info "  CONFIGURANDO BASE DE DATOS"
log_info "═══════════════════════════════════════"
echo ""

# Generar contraseña segura
DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)

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
    
    # Actualizar configuración del backend usando método seguro
    grep -v "^DB_PASSWORD=" "$CONFIG_DIR/backend.env" > "$CONFIG_DIR/backend.env.tmp"
    echo "DB_PASSWORD=${DB_PASSWORD}" >> "$CONFIG_DIR/backend.env.tmp"
    mv "$CONFIG_DIR/backend.env.tmp" "$CONFIG_DIR/backend.env"
    
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
# Usar método más seguro: crear archivo temporal y reemplazar
grep -v "^JWT_SECRET=" "$CONFIG_DIR/backend.env" > "$CONFIG_DIR/backend.env.tmp"
echo "JWT_SECRET=${JWT_SECRET}" >> "$CONFIG_DIR/backend.env.tmp"
mv "$CONFIG_DIR/backend.env.tmp" "$CONFIG_DIR/backend.env"
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
User=aymc
Group=aymc
WorkingDirectory=/opt/aymc/backend
EnvironmentFile=/etc/aymc/backend.env
ExecStart=/opt/aymc/backend/aymc-backend
Restart=always
RestartSec=10
StandardOutput=append:/var/log/aymc/backend.log
StandardError=append:/var/log/aymc/backend-error.log

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/aymc /var/log/aymc

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
User=aymc
Group=aymc
WorkingDirectory=/opt/aymc/agent
ExecStart=/opt/aymc/agent/aymc-agent -config /etc/aymc/agent.json
Restart=always
RestartSec=10
StandardOutput=append:/var/log/aymc/agent.log
StandardError=append:/var/log/aymc/agent-error.log

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/aymc /var/log/aymc

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
log_success "¡Disfruta de AYMC! 🎉"
echo ""
