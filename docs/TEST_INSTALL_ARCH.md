# 🧪 Guía de Prueba de Instalación - Arch Linux

## ⚠️ ADVERTENCIA IMPORTANTE

El instalador creará/modificará:
- `/opt/aymc` - Binarios
- `/var/aymc` - Datos de servidores
- `/etc/aymc` - Configuraciones
- `/var/log/aymc` - Logs
- `/etc/systemd/system/aymc-*.service` - Servicios
- Base de datos PostgreSQL `aymc`
- Usuario del sistema `aymc`

## 📋 Pre-requisitos

1. **PostgreSQL debe estar instalado y corriendo**:
```bash
sudo pacman -S postgresql
sudo -u postgres initdb -D /var/lib/postgres/data
sudo systemctl enable postgresql
sudo systemctl start postgresql
```

2. **Java debe estar instalado**:
```bash
sudo pacman -S jdk-openjdk
java -version
```

3. **Tienes permisos de sudo**

## 🚀 Ejecutar Instalación

### Opción 1: Instalación Completa (Recomendado para VPS)

```bash
cd /home/shni/Documents/GitHub/AYMC/build
sudo ./install-vps.sh
```

El instalador:
1. ✅ Detecta que estás en Arch Linux
2. ✅ Instala dependencias (postgresql, java)
3. ✅ Crea usuario `aymc`
4. ✅ Crea directorios en `/opt/aymc`, `/var/aymc`, `/etc/aymc`, `/var/log/aymc`
5. ✅ Copia binarios compilados
6. ✅ Configura PostgreSQL (crea base de datos + usuario)
7. ✅ Genera JWT secret aleatorio
8. ✅ Crea servicios systemd
9. ✅ Configura firewall (UFW si está instalado)
10. ✅ Inicia servicios

### Opción 2: Instalación Manual Paso a Paso (Para depuración)

Si quieres ver cada paso en detalle:

```bash
# 1. Ver lo que va a hacer
cd /home/shni/Documents/GitHub/AYMC/build
less install-vps.sh

# 2. Ejecutar con output completo
sudo ./install-vps.sh 2>&1 | tee install.log

# 3. Si algo falla, revisar el log
cat install.log
```

## 📝 Durante la Instalación

Observa estos mensajes:

### ✅ Éxito:
```
[SUCCESS] Distribución detectada: Arch Linux
[SUCCESS] Dependencias instaladas
[SUCCESS] Backend instalado
[SUCCESS] Agent instalado
[SUCCESS] Base de datos creada
[SUCCESS] Backend iniciado correctamente
[SUCCESS] Agent iniciado correctamente
```

### ⚠️ Advertencias (normales):
```
[WARNING] Usuario aymc ya existe
[WARNING] La base de datos puede que ya exista
[WARNING] No se detectó firewall (ufw/firewalld)
```

### ❌ Errores (requieren atención):
```
[ERROR] Go no está instalado
[ERROR] No se encontró el binario del backend
[ERROR] Error al iniciar el backend
[ERROR] Backend no está escuchando en puerto 8080
```

## 🔍 Verificar Instalación

### 1. Ver servicios
```bash
# Estado
sudo systemctl status aymc-backend
sudo systemctl status aymc-agent

# Logs en tiempo real
sudo journalctl -u aymc-backend -f
sudo journalctl -u aymc-agent -f
```

### 2. Ver logs de archivo
```bash
# Backend
sudo tail -f /var/log/aymc/backend.log
sudo tail -f /var/log/aymc/backend-error.log

# Agent
sudo tail -f /var/log/aymc/agent.log
sudo tail -f /var/log/aymc/agent-error.log
```

### 3. Probar API
```bash
# Health check
curl http://localhost:8080/health

# Debe responder:
# {"status":"ok"}

# Registrar usuario
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@aymc.local",
    "password": "Test123456!"
  }'
```

### 4. Verificar puertos
```bash
# Backend (8080)
sudo ss -tlnp | grep 8080

# Agent (50051)
sudo ss -tlnp | grep 50051
```

### 5. Verificar base de datos
```bash
sudo -u postgres psql aymc
\dt  # Listar tablas
SELECT * FROM users;  # Ver usuarios
\q  # Salir
```

## 🐛 Problemas Comunes

### PostgreSQL no inicia
```bash
# Verificar estado
sudo systemctl status postgresql

# Si no existe la base de datos, inicializar
sudo -u postgres initdb -D /var/lib/postgres/data
sudo systemctl start postgresql
```

### Puerto 8080 ocupado
```bash
# Ver qué proceso usa el puerto
sudo ss -tlnp | grep 8080

# Detener proceso anterior
sudo systemctl stop aymc-backend  # Si es AYMC viejo
# O cambiar puerto en /etc/aymc/backend.env
```

### Error "Permission Denied"
```bash
# Arreglar permisos
sudo chown aymc:aymc /var/aymc -R
sudo chmod 755 /var/aymc
sudo systemctl restart aymc-backend aymc-agent
```

### Backend inicia pero no responde
```bash
# Ver logs detallados
sudo journalctl -u aymc-backend -n 100

# Verificar configuración
sudo cat /etc/aymc/backend.env

# Probar conexión a PostgreSQL
sudo -u postgres psql -l
```

## 🔄 Comandos Útiles Post-Instalación

```bash
# Reiniciar servicios
sudo systemctl restart aymc-backend aymc-agent

# Ver logs
sudo journalctl -u aymc-backend -f
sudo journalctl -u aymc-agent -f

# Detener servicios
sudo systemctl stop aymc-backend aymc-agent

# Estado
sudo systemctl status aymc-backend aymc-agent

# Configuración
sudo nano /etc/aymc/backend.env
sudo nano /etc/aymc/agent.json
```

## 🗑️ Desinstalar (si algo sale mal)

```bash
cd /home/shni/Documents/GitHub/AYMC/build
sudo ./uninstall.sh
```

El desinstalador preguntará si quieres eliminar:
- Base de datos
- Datos de servidores
- Reglas de firewall

## ✅ Checklist de Verificación

Después de instalar, verifica:

- [ ] `sudo systemctl status aymc-backend` → active (running)
- [ ] `sudo systemctl status aymc-agent` → active (running)
- [ ] `curl http://localhost:8080/health` → {"status":"ok"}
- [ ] `sudo ss -tlnp | grep 8080` → muestra aymc-backend
- [ ] `sudo ss -tlnp | grep 50051` → muestra aymc-agent
- [ ] `sudo -u postgres psql -l | grep aymc` → muestra base de datos
- [ ] Logs sin errores en `/var/log/aymc/`

## 📊 Monitoreo en Tiempo Real

Para ver todo en tiempo real durante la instalación:

```bash
# Terminal 1: Ejecutar instalador
sudo ./install-vps.sh

# Terminal 2: Ver logs PostgreSQL
sudo journalctl -u postgresql -f

# Terminal 3: Ver puertos
watch -n 1 'sudo ss -tlnp | grep -E "(8080|50051)"'

# Terminal 4: Ver procesos
watch -n 1 'ps aux | grep aymc'
```

## 🎯 Siguiente Paso

Una vez instalado y funcionando:

1. **Abrir el frontend**:
   ```bash
   cd SeraMC
   npm run dev
   ```

2. **Registrar usuario** en http://localhost:1420/register

3. **Crear servidor** en la interfaz web

4. **Verificar** que el servidor se crea en `/var/aymc/servers/`

---

¿Listo para instalar? Ejecuta:

```bash
cd /home/shni/Documents/GitHub/AYMC/build
sudo ./install-vps.sh
```

Y reporta cualquier error o warning que veas! 🚀
