#!/bin/bash

#################################################
# AYMC Uninstaller
# Elimina completamente AYMC del sistema
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
echo "║             Uninstaller v1.0.0                    ║"
echo "║                                                   ║"
echo "╚═══════════════════════════════════════════════════╝"
echo ""

log_warning "⚠️  ADVERTENCIA: Esto eliminará COMPLETAMENTE AYMC del sistema"
log_warning "Esto incluye:"
echo "  - Binarios del backend y agent"
echo "  - Servicios systemd"
echo "  - Configuraciones"
echo "  - Logs"
echo ""
log_error "⚠️  LOS DATOS DE SERVIDORES Y BACKUPS PERMANECERÁN EN /var/aymc"
log_error "Si quieres eliminarlos también, hazlo manualmente después"
echo ""

read -p "¿Estás seguro de que quieres desinstalar AYMC? (escribe 'SI' para confirmar): " confirm

if [ "$confirm" != "SI" ]; then
    log_info "Desinstalación cancelada"
    exit 0
fi

echo ""
log_info "Iniciando desinstalación..."
sleep 2

#################################################
# DETENER SERVICIOS
#################################################

log_info "═══════════════════════════════════════"
log_info "  DETENIENDO SERVICIOS"
log_info "═══════════════════════════════════════"
echo ""

if systemctl is-active --quiet aymc-backend; then
    log_info "Deteniendo aymc-backend..."
    systemctl stop aymc-backend
    log_success "Backend detenido"
fi

if systemctl is-active --quiet aymc-agent; then
    log_info "Deteniendo aymc-agent..."
    systemctl stop aymc-agent
    log_success "Agent detenido"
fi

echo ""

#################################################
# DESHABILITAR SERVICIOS
#################################################

log_info "Deshabilitando servicios..."

if systemctl is-enabled --quiet aymc-backend 2>/dev/null; then
    systemctl disable aymc-backend
    log_success "Backend deshabilitado"
fi

if systemctl is-enabled --quiet aymc-agent 2>/dev/null; then
    systemctl disable aymc-agent
    log_success "Agent deshabilitado"
fi

echo ""

#################################################
# ELIMINAR SERVICIOS SYSTEMD
#################################################

log_info "═══════════════════════════════════════"
log_info "  ELIMINANDO SERVICIOS"
log_info "═══════════════════════════════════════"
echo ""

if [ -f /etc/systemd/system/aymc-backend.service ]; then
    rm /etc/systemd/system/aymc-backend.service
    log_success "Servicio backend eliminado"
fi

if [ -f /etc/systemd/system/aymc-agent.service ]; then
    rm /etc/systemd/system/aymc-agent.service
    log_success "Servicio agent eliminado"
fi

systemctl daemon-reload
log_success "Systemd recargado"
echo ""

#################################################
# ELIMINAR BINARIOS
#################################################

log_info "═══════════════════════════════════════"
log_info "  ELIMINANDO BINARIOS"
log_info "═══════════════════════════════════════"
echo ""

if [ -d /opt/aymc ]; then
    rm -rf /opt/aymc
    log_success "Binarios eliminados: /opt/aymc"
fi

echo ""

#################################################
# ELIMINAR CONFIGURACIONES
#################################################

log_info "═══════════════════════════════════════"
log_info "  ELIMINANDO CONFIGURACIONES"
log_info "═══════════════════════════════════════"
echo ""

if [ -d /etc/aymc ]; then
    rm -rf /etc/aymc
    log_success "Configuraciones eliminadas: /etc/aymc"
fi

echo ""

#################################################
# ELIMINAR LOGS
#################################################

log_info "═══════════════════════════════════════"
log_info "  ELIMINANDO LOGS"
log_info "═══════════════════════════════════════"
echo ""

if [ -d /var/log/aymc ]; then
    rm -rf /var/log/aymc
    log_success "Logs eliminados: /var/log/aymc"
fi

echo ""

#################################################
# ELIMINAR USUARIO
#################################################

log_info "═══════════════════════════════════════"
log_info "  ELIMINANDO USUARIO"
log_info "═══════════════════════════════════════"
echo ""

if id "aymc" &>/dev/null; then
    userdel aymc 2>/dev/null || log_warning "No se pudo eliminar el usuario aymc"
    log_success "Usuario aymc eliminado"
fi

echo ""

#################################################
# ELIMINAR BASE DE DATOS (OPCIONAL)
#################################################

log_info "═══════════════════════════════════════"
log_info "  BASE DE DATOS"
log_info "═══════════════════════════════════════"
echo ""

log_warning "La base de datos PostgreSQL 'aymc' aún existe"
read -p "¿Quieres eliminar la base de datos también? (s/N): " delete_db

if [[ "$delete_db" =~ ^[Ss]$ ]]; then
    log_info "Eliminando base de datos..."
    sudo -u postgres psql << EOF
DROP DATABASE IF EXISTS aymc;
DROP USER IF EXISTS aymc;
EOF
    log_success "Base de datos eliminada"
else
    log_info "Base de datos conservada"
fi

echo ""

#################################################
# ELIMINAR DATOS (OPCIONAL)
#################################################

log_info "═══════════════════════════════════════"
log_info "  DATOS DE SERVIDORES"
log_info "═══════════════════════════════════════"
echo ""

if [ -d /var/aymc ]; then
    TOTAL_SIZE=$(du -sh /var/aymc 2>/dev/null | cut -f1)
    log_warning "Los datos de servidores y backups ocupan: $TOTAL_SIZE"
    log_warning "Ubicación: /var/aymc"
    echo ""
    read -p "¿Quieres eliminar TODOS los datos? (s/N): " delete_data
    
    if [[ "$delete_data" =~ ^[Ss]$ ]]; then
        log_error "⚠️  ÚLTIMA ADVERTENCIA: Esto eliminará TODOS los servidores y backups"
        read -p "Escribe 'ELIMINAR TODO' para confirmar: " final_confirm
        
        if [ "$final_confirm" = "ELIMINAR TODO" ]; then
            rm -rf /var/aymc
            log_success "Datos eliminados: /var/aymc"
        else
            log_info "Eliminación de datos cancelada"
        fi
    else
        log_info "Datos conservados en /var/aymc"
        log_info "Puedes eliminarlos manualmente con: sudo rm -rf /var/aymc"
    fi
fi

echo ""

#################################################
# LIMPIAR FIREWALL (OPCIONAL)
#################################################

log_info "═══════════════════════════════════════"
log_info "  FIREWALL"
log_info "═══════════════════════════════════════"
echo ""

read -p "¿Quieres eliminar las reglas del firewall? (s/N): " clean_firewall

if [[ "$clean_firewall" =~ ^[Ss]$ ]]; then
    if command -v ufw &> /dev/null; then
        log_info "Eliminando reglas de UFW..."
        ufw delete allow 8080/tcp 2>/dev/null || true
        ufw delete allow 50051/tcp 2>/dev/null || true
        ufw delete allow 25565:25600/tcp 2>/dev/null || true
        log_success "Reglas de UFW eliminadas"
    elif command -v firewall-cmd &> /dev/null; then
        log_info "Eliminando reglas de firewalld..."
        firewall-cmd --permanent --remove-port=8080/tcp 2>/dev/null || true
        firewall-cmd --permanent --remove-port=50051/tcp 2>/dev/null || true
        firewall-cmd --permanent --remove-port=25565-25600/tcp 2>/dev/null || true
        firewall-cmd --reload
        log_success "Reglas de firewalld eliminadas"
    fi
else
    log_info "Reglas de firewall conservadas"
fi

echo ""

#################################################
# RESUMEN
#################################################

log_success "═══════════════════════════════════════"
log_success "  DESINSTALACIÓN COMPLETADA"
log_success "═══════════════════════════════════════"
echo ""
log_info "✅ Eliminado:"
echo "   - Binarios (/opt/aymc)"
echo "   - Servicios systemd"
echo "   - Configuraciones (/etc/aymc)"
echo "   - Logs (/var/log/aymc)"
echo "   - Usuario 'aymc'"

if [[ "$delete_db" =~ ^[Ss]$ ]]; then
    echo "   - Base de datos PostgreSQL"
fi

if [[ "$delete_data" =~ ^[Ss]$ ]] && [ "$final_confirm" = "ELIMINAR TODO" ]; then
    echo "   - Datos de servidores (/var/aymc)"
fi

echo ""

if [ -d /var/aymc ]; then
    log_warning "❗ Conservado:"
    echo "   - Datos de servidores: /var/aymc ($TOTAL_SIZE)"
    echo "   - Para eliminar: sudo rm -rf /var/aymc"
    echo ""
fi

log_success "AYMC ha sido desinstalado del sistema"
echo ""
