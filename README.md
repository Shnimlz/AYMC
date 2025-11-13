# AYMC - Advanced Your Minecraft Controller

<div align="center">

![AYMC Logo](./docs/assets/logo.png)

**Sistema completo de gestión y administración de servidores Minecraft**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![Vue 3](https://img.shields.io/badge/Vue-3.0+-4FC08D?logo=vue.js)](https://vuejs.org/)
[![Tauri](https://img.shields.io/badge/Tauri-1.5+-FFC131?logo=tauri)](https://tauri.app/)

[Características](#-características) •
[Instalación](#-instalación) •
[Uso](#-uso) •
[Documentación](#-documentación) •
[Contribuir](#-contribuir)

</div>

---

## 📋 Descripción

AYMC es una plataforma completa para gestionar servidores Minecraft desde una interfaz moderna y elegante. Incluye:

- **Frontend Desktop** (Tauri + Vue 3): Aplicación nativa para Windows, Linux y macOS
- **Backend API** (Go + Gin): API REST robusta con autenticación JWT
- **Agent** (Go + gRPC): Agente que se ejecuta en servidores remotos para gestionar instancias de Minecraft
- **Base de Datos** (PostgreSQL): Almacenamiento persistente y confiable

## ✨ Características

### 🎮 Gestión de Servidores
- ✅ Crear, iniciar, detener y reiniciar servidores Minecraft
- ✅ Soporte para múltiples versiones (Vanilla, Paper, Spigot, Purpur, Fabric, Forge)
- ✅ Configuración de memoria RAM, puertos y argumentos Java
- ✅ Auto-inicio y auto-reinicio configurable
- ✅ Logs en tiempo real vía WebSocket

### 🔌 Gestión de Plugins
- ✅ Búsqueda y descarga desde SpigotMC, Hangar, Modrinth y CurseForge
- ✅ Instalación con un solo clic
- ✅ Actualización automática de plugins
- ✅ Gestión de dependencias

### 💾 Backups Automáticos
- ✅ Backups programados con cron expressions
- ✅ Compresión inteligente (gzip, zip, tar)
- ✅ Retención configurable
- ✅ Restauración con un clic
- ✅ Backups incrementales y completos

### 📊 Monitoreo en Tiempo Real
- ✅ CPU, RAM, Disco y Red
- ✅ Jugadores conectados
- ✅ TPS (Ticks Per Second)
- ✅ Gráficos históricos
- ✅ Alertas configurables

### 👥 Multi-Usuario
- ✅ Sistema de roles (Admin, Moderador, Usuario)
- ✅ Permisos granulares por servidor
- ✅ Registro de auditoría
- ✅ Autenticación JWT segura

### 🌐 Multi-Agente
- ✅ Gestiona servidores en múltiples VPS
- ✅ Comunicación gRPC eficiente
- ✅ Health checks automáticos
- ✅ Balanceo de carga

## 🏗️ Arquitectura

```
┌─────────────────────────────────────────────────────────────┐
│                    FRONTEND (Tauri + Vue 3)                 │
│                  Desktop App (Windows/Linux/macOS)          │
└────────────────────────────┬────────────────────────────────┘
                             │ HTTPS / REST API
                             │
┌────────────────────────────▼────────────────────────────────┐
│                   BACKEND (Go + Gin)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  REST API    │  │  WebSocket   │  │  gRPC Client │     │
│  │  (82 rutas)  │  │  (Logs)      │  │  (a Agents)  │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │           PostgreSQL Database                        │  │
│  │  (Usuarios, Servidores, Plugins, Backups, Métricas) │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────┘
                             │ gRPC
              ┌──────────────┴──────────────┐
              │                             │
┌─────────────▼─────────────┐  ┌───────────▼────────────┐
│   AGENT 1 (VPS 1)         │  │   AGENT 2 (VPS 2)      │
│  ┌────────────────────┐   │  │  ┌────────────────┐    │
│  │ MC Server 1        │   │  │  │ MC Server 3    │    │
│  │ MC Server 2        │   │  │  │ MC Server 4    │    │
│  └────────────────────┘   │  │  └────────────────┘    │
└───────────────────────────┘  └────────────────────────┘
```

## 🚀 Instalación

### Requisitos Previos

#### Para el Frontend (Desktop App)
- Node.js 18+ y npm/pnpm
- Rust 1.70+ (para Tauri)

#### Para el Backend/Agent (VPS)
- Sistema Operativo: Arch Linux, Debian 11/12, Ubuntu 20.04/22.04/24.04, RHEL/CentOS/Rocky/AlmaLinux 8/9, Fedora 38+
- CPU: 2 núcleos mínimo (4 recomendado)
- RAM: 4GB mínimo (8GB recomendado)
- Disco: 20GB mínimo (50GB+ recomendado)
- PostgreSQL 13+
- Java 17+ (para servidores Minecraft)

### Instalación Automática en VPS

```bash
# 1. Descargar el paquete
wget https://github.com/tuusuario/aymc/releases/latest/download/aymc-latest-linux-amd64.tar.gz

# 2. Extraer
tar -xzf aymc-latest-linux-amd64.tar.gz
cd aymc

# 3. Ejecutar instalador (requiere sudo)
sudo ./install-vps.sh

# El instalador configura automáticamente:
# - PostgreSQL con base de datos 'aymc'
# - Backend API en puerto 8080
# - Agent gRPC en puerto 50051
# - Servicios systemd
# - Firewall (UFW/firewalld)
```

### Instalación Frontend (Aplicación de Escritorio)

```bash
# Clonar repositorio
git clone https://github.com/tuusuario/aymc.git
cd aymc/SeraMC

# Instalar dependencias
npm install

# Configurar URL del backend
# Editar src/config.ts y establecer BACKEND_URL
echo "export const BACKEND_URL = 'https://tu-vps.com'" > src/config.ts

# Desarrollo
npm run tauri dev

# Compilar para producción
npm run tauri build
```

## 📖 Uso

### 1. Primer Inicio

Al abrir la aplicación por primera vez:

1. **Configurar Backend**: La app pedirá la URL del backend (ej: `https://tu-vps.com:8080`)
2. **Registrar Usuario**: Crear cuenta con email y contraseña
3. **Iniciar Sesión**: Autenticarse con las credenciales

### 2. Registrar un Agent

Antes de crear servidores, debes registrar al menos un agent:

1. Ve a **"Agents"** en el menú lateral
2. Clic en **"Agregar Agent"**
3. Ingresa:
   - **Agent ID**: Identificador único (ej: `vps-us-east-1`)
   - **Hostname**: Nombre del servidor (ej: `mc-server-1`)
   - **IP Address**: IP pública del VPS
   - **Port**: 50051 (por defecto)

### 3. Crear un Servidor Minecraft

1. Ve a **"Servidores"** → **"Crear Servidor"**
2. Configura:
   - **Nombre**: Identificador interno
   - **Display Name**: Nombre visible
   - **Tipo**: Paper, Spigot, Vanilla, etc.
   - **Versión**: 1.20.1, 1.19.4, etc.
   - **RAM**: Mínima y máxima (MB)
   - **Puerto**: 25565 (por defecto)
   - **Agent**: Selecciona dónde se ejecutará

3. Clic en **"Crear"**

### 4. Iniciar Servidor

1. En la lista de servidores, clic en **"Iniciar"**
2. Monitorea el proceso en **"Logs"**
3. Cuando esté online, conéctate desde Minecraft con: `tu-vps.com:25565`

### 5. Instalar Plugins

1. Ve a **"Marketplace"**
2. Busca el plugin deseado (ej: "EssentialsX")
3. Selecciona el servidor destino
4. Clic en **"Instalar"**
5. Reinicia el servidor para aplicar cambios

### 6. Configurar Backups

1. Ve a **Servidor → Backups → Configuración**
2. Habilita backups automáticos
3. Configura:
   - **Frecuencia**: Cron expression (ej: `0 2 * * *` = 2 AM diario)
   - **Retención**: Cantidad de backups a mantener
   - **Compresión**: gzip, zip o tar
   - **Incluir**: World, plugins, config, logs

## 🔧 Configuración Avanzada

### Variables de Entorno (Backend)

El instalador crea `/etc/aymc/backend.env`:

```env
# Aplicación
APP_ENV=production
APP_PORT=8080

# Base de Datos
DB_HOST=localhost
DB_PORT=5432
DB_USER=aymc
DB_PASSWORD=<generada-automáticamente>
DB_NAME=aymc
DB_SSL_MODE=disable

# Seguridad
JWT_SECRET=<generada-automáticamente>
JWT_EXPIRES_IN=24h
JWT_REFRESH_EXPIRES_IN=168h

# CORS
CORS_ORIGINS=http://localhost:1420,tauri://localhost
```

### Configuración del Agent

El instalador crea `/etc/aymc/agent.json`:

```json
{
  "agent_id": "agent-1",
  "backend_url": "http://localhost:8080",
  "port": 50051,
  "work_dir": "/var/aymc/servers",
  "max_servers": 50
}
```

### Configurar HTTPS (Recomendado para Producción)

```bash
# Instalar Nginx
sudo apt install nginx certbot python3-certbot-nginx

# Configurar proxy reverso
sudo nano /etc/nginx/sites-available/aymc

# Contenido:
server {
    listen 80;
    server_name tu-dominio.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
    
    location /ws {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
    }
}

# Activar sitio
sudo ln -s /etc/nginx/sites-available/aymc /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx

# Obtener certificado SSL
sudo certbot --nginx -d tu-dominio.com
```

## 📚 Documentación

- **[Guía de Instalación VPS](./docs/INSTALL_VPS.md)** - Instalación detallada paso a paso
- **[Test de Instalación en Arch](./docs/TEST_INSTALL_ARCH.md)** - Guía de testing en Arch Linux
- **[Errores Solucionados](./docs/VPS_ERRORS_FIXED.md)** - Problemas comunes y soluciones
- **[Resumen de Instalación](./docs/INSTALLATION_SUMMARY.md)** - Vista general del proceso
- **[API Reference](./docs/API.md)** - Documentación completa de la API REST
- **[Scripts README](./scripts/README.md)** - Guía de uso de scripts

## 🔍 Monitoreo y Logs

### Ver Logs del Backend

```bash
# Logs en tiempo real
sudo journalctl -u aymc-backend -f

# Últimas 100 líneas
sudo journalctl -u aymc-backend -n 100

# Archivo de log
sudo tail -f /var/log/aymc/backend.log
```

### Ver Logs del Agent

```bash
# Logs en tiempo real
sudo journalctl -u aymc-agent -f

# Archivo de log
sudo tail -f /var/log/aymc/agent.log
```

### Estado de Servicios

```bash
# Ver estado
sudo systemctl status aymc-backend aymc-agent

# Reiniciar servicios
sudo systemctl restart aymc-backend aymc-agent

# Ver puertos activos
sudo ss -tlnp | grep -E "(8080|50051)"
```

## 🐛 Troubleshooting

### Backend no inicia

```bash
# Verificar logs
sudo journalctl -u aymc-backend -n 50

# Verificar configuración
sudo cat /etc/aymc/backend.env

# Verificar permisos
sudo ls -la /etc/aymc/
sudo ls -la /opt/aymc/

# Verificar PostgreSQL
sudo systemctl status postgresql
sudo -u postgres psql -c "\l" | grep aymc
```

### Agent no conecta al Backend

```bash
# Verificar configuración
sudo cat /etc/aymc/agent.json

# Probar conectividad
curl http://localhost:8080/health

# Verificar firewall
sudo ufw status
sudo firewall-cmd --list-all
```

### Frontend no conecta al Backend

1. Verificar que `src/config.ts` tenga la URL correcta
2. Verificar que el backend esté accesible: `curl https://tu-vps.com:8080/health`
3. Verificar CORS en `/etc/aymc/backend.env` incluya la URL del frontend
4. Verificar certificado SSL si usas HTTPS

## 🗑️ Desinstalación

```bash
cd /opt/aymc
sudo ./uninstall.sh

# El script preguntará si deseas eliminar:
# - Servicios systemd
# - Binarios
# - Configuraciones
# - Base de datos
# - Datos de servidores (/var/aymc)
# - Reglas de firewall
```

## 🛠️ Desarrollo

### Estructura del Proyecto

```
aymc/
├── backend/           # API REST (Go)
│   ├── api/          # Handlers y rutas
│   ├── database/     # Modelos y migraciones
│   ├── services/     # Lógica de negocio
│   └── cmd/server/   # Entry point
├── agent/            # Agent gRPC (Go)
│   ├── grpc/        # Servicios gRPC
│   ├── minecraft/   # Gestión de MC servers
│   └── cmd/agent/   # Entry point
├── SeraMC/           # Frontend (Vue 3 + Tauri)
│   ├── src/         # Código Vue
│   ├── src-tauri/   # Código Rust
│   └── public/      # Assets
├── docs/             # Documentación
├── scripts/          # Scripts de instalación
│   ├── build.sh           # Compilar binarios
│   ├── install-vps.sh     # Instalador automático
│   ├── continue-install.sh # Recuperación
│   └── uninstall.sh       # Desinstalador
└── README.md         # Este archivo
```

### Compilar desde Código Fuente

```bash
# Backend
cd backend
go mod download
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o aymc-backend ./cmd/server

# Agent
cd ../agent
go mod download
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o aymc-agent ./cmd/agent

# Frontend
cd ../SeraMC
npm install
npm run tauri build
```

### Ejecutar Tests

```bash
# Backend
cd backend
go test ./...

# Frontend
cd SeraMC
npm run test
```

## 🤝 Contribuir

¡Las contribuciones son bienvenidas! Por favor:

1. Fork el repositorio
2. Crea una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Ver [LICENSE](LICENSE) para más detalles.

## 👥 Autores

- **Tu Nombre** - *Desarrollo inicial* - [GitHub](https://github.com/tuusuario)

## 🙏 Agradecimientos

- [Gin](https://github.com/gin-gonic/gin) - Framework web para Go
- [GORM](https://gorm.io/) - ORM para Go
- [Vue 3](https://vuejs.org/) - Framework JavaScript progresivo
- [Tauri](https://tauri.app/) - Framework para aplicaciones de escritorio
- [PostgreSQL](https://www.postgresql.org/) - Base de datos

## 📞 Soporte

- 📧 Email: soporte@aymc.com
- 💬 Discord: [Servidor AYMC](https://discord.gg/aymc)
- 🐛 Issues: [GitHub Issues](https://github.com/tuusuario/aymc/issues)

---

<div align="center">
Hecho con ❤️ para la comunidad de Minecraft
</div>
