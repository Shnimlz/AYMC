# 📖 Guía de Instalación VPS - AYMC

## 🎯 Objetivo

Esta guía te ayudará a instalar AYMC (Backend + Agent) en tu servidor VPS o máquina local Linux.

---

## 📋 Requisitos Previos

### Hardware Mínimo
- **CPU**: 2 cores
- **RAM**: 4 GB (2 GB para AYMC + 2 GB para servidores Minecraft)
- **Disco**: 20 GB libre
- **Red**: Conexión estable a Internet

### Hardware Recomendado
- **CPU**: 4+ cores
- **RAM**: 8+ GB
- **Disco**: 50+ GB SSD
- **Red**: 100+ Mbps

### Software
- **OS**: Linux (Arch, Debian, Ubuntu, RHEL, CentOS, Fedora, Rocky, AlmaLinux)
- **Acceso**: Root o sudo
- **Abiertos**: Puertos 8080, 50051, 25565-25600

---

## 🚀 Instalación Rápida

### Paso 1: Compilar Binarios

En tu máquina de desarrollo (donde tienes el código fuente):

```bash
cd /path/to/AYMC
./scripts/build.sh
```

Esto generará un tarball en `build/aymc-YYYY.MM.DD-linux-amd64.tar.gz`

### Paso 2: Transferir al Servidor

```bash
# Copiar tarball al VPS
scp build/aymc-*.tar.gz user@your-vps:/tmp/

# Conectar al VPS
ssh user@your-vps
```

### Paso 3: Instalar

```bash
# En el VPS
cd /tmp
tar -xzf aymc-*.tar.gz
sudo ./install-vps.sh
```

¡Listo! AYMC está instalado y corriendo.

---

## 📝 Instalación Detallada

### 1. Preparar el Sistema

#### Arch Linux

```bash
# Actualizar sistema
sudo pacman -Syu

# Instalar dependencias
sudo pacman -S postgresql jdk-openjdk wget curl

# Iniciar PostgreSQL
sudo systemctl enable postgresql
sudo systemctl start postgresql
```

#### Debian/Ubuntu

```bash
# Actualizar sistema
sudo apt update && sudo apt upgrade -y

# Instalar dependencias
sudo apt install -y postgresql postgresql-contrib openjdk-21-jdk wget curl

# PostgreSQL ya inicia automáticamente
```

#### RHEL/CentOS/Rocky/AlmaLinux

```bash
# Actualizar sistema
sudo yum update -y

# Instalar dependencias
sudo yum install -y postgresql-server postgresql-contrib java-21-openjdk wget curl

# Inicializar PostgreSQL
sudo postgresql-setup --initdb
sudo systemctl enable postgresql
sudo systemctl start postgresql
```

### 2. Compilar desde Código Fuente

```bash
# En tu máquina de desarrollo
cd ~/AYMC

# Ejecutar script de build
./scripts/build.sh

# Verificar archivos generados
ls -lh build/
```

**Salida esperada:**
```
aymc-2025.11.13-linux-amd64.tar.gz
aymc-2025.11.13-linux-amd64.tar.gz.sha256
backend/aymc-backend
agent/aymc-agent
config/backend.env
config/agent.json
install-vps.sh
uninstall.sh
```

### 3. Transferir Archivos

#### Opción A: SCP (Recomendado)

```bash
scp build/aymc-*.tar.gz user@your-vps.com:/tmp/
```

#### Opción B: SFTP

```bash
sftp user@your-vps.com
put build/aymc-*.tar.gz /tmp/
exit
```

#### Opción C: Servidor Local (desarrollo)

Si estás probando en tu Arch Linux:

```bash
# No necesitas transferir, solo extrae directamente
cd ~/AYMC/build
tar -xzf aymc-*.tar.gz
```

### 4. Ejecutar Instalador

```bash
# Conectar al servidor
ssh user@your-vps.com

# O si es local, solo:
cd /tmp

# Extraer tarball
tar -xzf aymc-*.tar.gz

# Ejecutar instalador con sudo
sudo ./install-vps.sh
```

### 5. Seguir el Instalador

El instalador te pedirá confirmación en algunos pasos:

1. **Dependencias**: Se instalan automáticamente
2. **PostgreSQL**: Se configura automáticamente
3. **Firewall**: Se configura automáticamente (UFW o firewalld)
4. **Servicios**: Se habilitan e inician automáticamente

---

## 🔧 Configuración Post-Instalación

### 1. Verificar Estado

```bash
# Ver estado de servicios
sudo systemctl status aymc-backend
sudo systemctl status aymc-agent

# Ver logs en tiempo real
sudo journalctl -u aymc-backend -f
sudo journalctl -u aymc-agent -f
```

### 2. Configurar Backend

Edita `/etc/aymc/backend.env`:

```bash
sudo nano /etc/aymc/backend.env
```

**Importante cambiar:**

```env
# Base de datos (ya configurado automáticamente)
DB_PASSWORD=<generado-automáticamente>

# JWT Secret (ya generado automáticamente)
JWT_SECRET=<generado-automáticamente>

# CORS - Agregar tu dominio
CORS_ORIGINS=http://localhost:1420,https://tu-dominio.com

# Opcional: Cambiar puerto
APP_PORT=8080
```

Después de cambios:

```bash
sudo systemctl restart aymc-backend
```

### 3. Configurar Agent

Edita `/etc/aymc/agent.json`:

```bash
sudo nano /etc/aymc/agent.json
```

```json
{
  "agent_id": "agent-1",
  "backend_url": "http://localhost:8080",
  "port": 50051,
  "log_level": "info",
  "max_servers": 50,
  "java_path": "/usr/bin/java",
  "work_dir": "/var/aymc/servers",
  "enable_metrics": true,
  "metrics_interval": 30000000000
}
```

Después de cambios:

```bash
sudo systemctl restart aymc-agent
```

### 4. Abrir Puertos en el Firewall

El instalador configura esto automáticamente, pero si usas un firewall externo (cloud provider):

**AWS Security Group / DigitalOcean / Vultr:**
- `8080/tcp` - Backend API
- `50051/tcp` - Agent gRPC
- `25565-25600/tcp` - Servidores Minecraft

---

## 🧪 Probar la Instalación

### 1. Verificar Backend API

```bash
# Prueba de health check
curl http://localhost:8080/health

# Debe responder:
# {"status":"ok"}
```

### 2. Verificar Agent gRPC

```bash
# Verificar que el puerto está abierto
ss -tlnp | grep 50051

# Debe mostrar:
# LISTEN 0 128 [::]:50051 [::]:* users:(("aymc-agent",pid=XXXX,fd=3))
```

### 3. Registrar Usuario

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@aymc.local",
    "password": "SecurePass123!"
  }'

# Respuesta esperada:
# {"token":"eyJhbGc...","user":{...}}
```

### 4. Registrar Agent

```bash
# Obtener el token del paso anterior
TOKEN="eyJhbGc..."

curl -X POST http://localhost:8080/api/v1/agents \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Local Agent",
    "host": "localhost",
    "port": 50051
  }'

# Respuesta esperada:
# {"id":"...","name":"Local Agent","status":"online"}
```

### 5. Crear Servidor de Prueba

```bash
curl -X POST http://localhost:8080/api/v1/servers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Test Server",
    "type": "paper",
    "version": "1.21",
    "port": 25565,
    "agent_id": "<agent-id-del-paso-anterior>",
    "ram_min": 1024,
    "ram_max": 2048,
    "java_version": "21"
  }'
```

---

## 📂 Estructura de Directorios

```
/opt/aymc/                    # Binarios
├── backend/
│   └── aymc-backend
└── agent/
    └── aymc-agent

/etc/aymc/                    # Configuración
├── backend.env
└── agent.json

/var/aymc/                    # Datos
├── servers/                  # Servidores Minecraft
│   ├── test-server/
│   └── ...
├── backups/                  # Respaldos
└── uploads/                  # Archivos subidos

/var/log/aymc/                # Logs
├── backend.log
├── backend-error.log
├── agent.log
└── agent-error.log
```

---

## 🔄 Comandos Útiles

### Gestión de Servicios

```bash
# Estado
sudo systemctl status aymc-backend
sudo systemctl status aymc-agent

# Iniciar
sudo systemctl start aymc-backend
sudo systemctl start aymc-agent

# Detener
sudo systemctl stop aymc-backend
sudo systemctl stop aymc-agent

# Reiniciar
sudo systemctl restart aymc-backend
sudo systemctl restart aymc-agent

# Habilitar al inicio
sudo systemctl enable aymc-backend
sudo systemctl enable aymc-agent

# Deshabilitar
sudo systemctl disable aymc-backend
sudo systemctl disable aymc-agent
```

### Logs

```bash
# Ver logs en tiempo real
sudo journalctl -u aymc-backend -f
sudo journalctl -u aymc-agent -f

# Ver últimas 100 líneas
sudo journalctl -u aymc-backend -n 100
sudo journalctl -u aymc-agent -n 100

# Ver logs de hoy
sudo journalctl -u aymc-backend --since today
sudo journalctl -u aymc-agent --since today

# Ver logs con errores
sudo journalctl -u aymc-backend -p err
sudo journalctl -u aymc-agent -p err

# Logs de archivo
sudo tail -f /var/log/aymc/backend.log
sudo tail -f /var/log/aymc/agent.log
```

### Base de Datos

```bash
# Conectar a PostgreSQL
sudo -u postgres psql aymc

# Comandos útiles en psql:
\dt                    # Listar tablas
\d users              # Describir tabla users
SELECT * FROM users;  # Ver usuarios
\q                    # Salir
```

### Backups

```bash
# Backup manual de la base de datos
sudo -u postgres pg_dump aymc > aymc-backup-$(date +%Y%m%d).sql

# Restaurar backup
sudo -u postgres psql aymc < aymc-backup-20251113.sql

# Backup de servidores
sudo tar -czf servers-backup-$(date +%Y%m%d).tar.gz /var/aymc/servers
```

---

## 🐛 Troubleshooting

### Backend no inicia

**Síntoma:** `systemctl status aymc-backend` muestra "failed"

**Solución:**

```bash
# Ver logs completos
sudo journalctl -u aymc-backend -n 50

# Errores comunes:
# 1. Puerto 8080 ocupado
sudo ss -tlnp | grep 8080
# Solución: Cambiar APP_PORT en /etc/aymc/backend.env

# 2. No puede conectar a PostgreSQL
sudo systemctl status postgresql
sudo -u postgres psql -l
# Solución: Verificar que PostgreSQL esté corriendo

# 3. Permisos incorrectos
sudo chown aymc:aymc /var/aymc -R
sudo chmod 755 /var/aymc
```

### Agent no inicia

**Síntoma:** `systemctl status aymc-agent` muestra "failed"

**Solución:**

```bash
# Ver logs
sudo journalctl -u aymc-agent -n 50

# Errores comunes:
# 1. Puerto 50051 ocupado
sudo ss -tlnp | grep 50051

# 2. No puede crear directorio de trabajo
sudo mkdir -p /var/aymc/servers
sudo chown aymc:aymc /var/aymc/servers
sudo chmod 755 /var/aymc/servers

# 3. Java no encontrado
which java
# Si no existe: sudo pacman -S jdk-openjdk (Arch)
```

### Error de conexión desde el frontend

**Síntoma:** Frontend no puede conectarse al backend

**Solución:**

```bash
# 1. Verificar que el backend esté corriendo
curl http://localhost:8080/health

# 2. Verificar firewall
sudo ufw status                    # Ubuntu/Debian
sudo firewall-cmd --list-all       # RHEL/CentOS

# 3. Si usas un VPS, verifica el firewall del proveedor
# AWS: Security Groups
# DigitalOcean: Cloud Firewall
# Vultr: Firewall Rules

# 4. Verificar CORS
sudo nano /etc/aymc/backend.env
# Asegúrate de que CORS_ORIGINS incluya tu dominio
# Ejemplo: CORS_ORIGINS=http://localhost:1420,https://tu-app.com
sudo systemctl restart aymc-backend
```

### Servidor Minecraft no inicia

**Síntoma:** El agente reporta error al iniciar servidor

**Solución:**

```bash
# 1. Verificar logs del servidor
sudo ls -la /var/aymc/servers/
sudo tail -f /var/aymc/servers/<nombre-servidor>/logs/latest.log

# 2. Verificar Java
java -version

# 3. Verificar permisos
sudo chown aymc:aymc /var/aymc/servers/<nombre-servidor> -R

# 4. Verificar RAM disponible
free -h

# 5. Verificar puerto disponible
sudo ss -tlnp | grep <puerto>
```

### Error "Permission Denied"

```bash
# Arreglar permisos de AYMC
sudo chown aymc:aymc /var/aymc -R
sudo chown aymc:aymc /var/log/aymc -R
sudo chmod 755 /var/aymc
sudo chmod 755 /var/log/aymc

# Reiniciar servicios
sudo systemctl restart aymc-backend aymc-agent
```

### Base de datos corrupta

```bash
# Verificar estado de PostgreSQL
sudo systemctl status postgresql
sudo -u postgres psql -c "SELECT version();"

# Si hay errores, intentar reparar
sudo -u postgres vacuumdb --all --analyze --verbose

# En caso extremo, restaurar desde backup
sudo -u postgres psql aymc < /path/to/backup.sql
```

---

## 🔒 Seguridad en Producción

### 1. HTTPS con Nginx

Instalar Nginx como reverse proxy:

```bash
# Arch
sudo pacman -S nginx certbot certbot-nginx

# Debian/Ubuntu
sudo apt install nginx certbot python3-certbot-nginx
```

Configurar Nginx (`/etc/nginx/sites-available/aymc`):

```nginx
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
    
    location /api/v1/ws {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
    }
}
```

Obtener certificado SSL:

```bash
sudo certbot --nginx -d tu-dominio.com
```

### 2. Firewall Estricto

```bash
# UFW (Debian/Ubuntu/Arch)
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 25565:25600/tcp
sudo ufw enable

# No expongas 8080 y 50051 directamente
# Usa Nginx como reverse proxy
```

### 3. Actualizar Contraseñas

```bash
# Editar configuración
sudo nano /etc/aymc/backend.env

# Cambiar:
DB_PASSWORD=<nueva-contraseña-fuerte>
JWT_SECRET=<nuevo-secret-aleatorio-64-chars>

# Actualizar PostgreSQL
sudo -u postgres psql
ALTER USER aymc WITH PASSWORD 'nueva-contraseña-fuerte';
\q

# Reiniciar
sudo systemctl restart aymc-backend
```

### 4. Fail2Ban (Opcional)

Proteger contra ataques de fuerza bruta:

```bash
# Instalar
sudo apt install fail2ban  # Debian/Ubuntu
sudo pacman -S fail2ban     # Arch

# Configurar para AYMC
sudo nano /etc/fail2ban/jail.local
```

```ini
[aymc-backend]
enabled = true
port = 80,443
filter = aymc-backend
logpath = /var/log/aymc/backend.log
maxretry = 5
bantime = 3600
```

---

## 🔄 Actualización

Para actualizar AYMC a una nueva versión:

```bash
# 1. Compilar nueva versión
cd ~/AYMC
git pull
./scripts/build.sh

# 2. Transferir al VPS
scp build/aymc-*.tar.gz user@your-vps:/tmp/

# 3. En el VPS
ssh user@your-vps
cd /tmp
tar -xzf aymc-*.tar.gz

# 4. Detener servicios
sudo systemctl stop aymc-backend aymc-agent

# 5. Backup de configuración
sudo cp /etc/aymc/backend.env /tmp/backend.env.bak
sudo cp /etc/aymc/agent.json /tmp/agent.json.bak

# 6. Copiar nuevos binarios
sudo cp backend/aymc-backend /opt/aymc/backend/
sudo cp agent/aymc-agent /opt/aymc/agent/

# 7. Restaurar configuración
sudo cp /tmp/backend.env.bak /etc/aymc/backend.env
sudo cp /tmp/agent.json.bak /etc/aymc/agent.json

# 8. Reiniciar servicios
sudo systemctl start aymc-backend aymc-agent

# 9. Verificar
sudo systemctl status aymc-backend aymc-agent
```

---

## 🗑️ Desinstalación

Para eliminar completamente AYMC:

```bash
sudo ./uninstall.sh
```

El desinstalador preguntará si quieres eliminar:
- Base de datos
- Datos de servidores (backups)
- Reglas de firewall

---

## 📊 Monitoreo

### Recursos del Sistema

```bash
# CPU y RAM
htop

# Espacio en disco
df -h /var/aymc

# Procesos de AYMC
ps aux | grep aymc

# Conexiones de red
sudo ss -tlnp | grep -E "(8080|50051|25565)"
```

### Logs Centralizados (Opcional)

Configurar Loki + Promtail para centralizar logs:

```bash
# TODO: Documentar integración con Grafana Loki
```

---

## 🆘 Soporte

- **Documentación**: https://github.com/aymc/aymc/tree/main/docs
- **Issues**: https://github.com/aymc/aymc/issues
- **Discord**: (Próximamente)

---

## 📝 Notas Finales

- **Backups**: Configura backups automáticos de `/var/aymc` y la base de datos
- **Monitoreo**: Considera usar herramientas como Grafana, Prometheus
- **Logs**: Rota logs regularmente para evitar llenar el disco
- **Actualizaciones**: Mantén el sistema y AYMC actualizados

¡Disfruta de AYMC! 🎉
