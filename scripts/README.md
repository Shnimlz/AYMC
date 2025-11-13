# 📜 Scripts de Instalación AYMC

Este directorio contiene todos los scripts necesarios para compilar, instalar y gestionar AYMC en VPS.

---

## 📦 Scripts Disponibles

### 🔨 `build.sh` - Compilador de Binarios

Compila el backend y agent en binarios optimizados y genera un tarball listo para distribución.

**Uso:**
```bash
./scripts/build.sh
```

**Output:**
- `build/aymc-YYYY.MM.DD-linux-amd64.tar.gz` (paquete completo)
- `build/backend/aymc-backend` (binario backend)
- `build/agent/aymc-agent` (binario agent)
- `build/config/backend.env` (configuración)
- `build/config/agent.json` (configuración agent)

---

### 🚀 `install-vps.sh` - Instalador VPS

Instala AYMC completamente en un VPS o servidor Linux.

**Soporta:**
- Arch Linux / Manjaro
- Debian 11/12
- Ubuntu 20.04/22.04/24.04
- RHEL 8/9
- CentOS Stream
- Rocky Linux / AlmaLinux
- Fedora 38+

**Uso:**
```bash
# En el VPS:
cd /tmp
tar -xzf aymc-*.tar.gz
sudo ./install-vps.sh
```

**Qué hace:**
1. Detecta distribución de Linux
2. Instala dependencias (PostgreSQL, Java, etc.)
3. Inicializa PostgreSQL (si es necesario)
4. Crea usuario y directorios
5. Instala binarios en `/opt/aymc`
6. Configura base de datos
7. Genera secrets aleatorios
8. Crea servicios systemd
9. Configura firewall
10. Inicia servicios

---

### 🔄 `continue-install.sh` - Continuación de Instalación

Script auxiliar para continuar la instalación si falla en PostgreSQL.

**Uso:**
```bash
# Si install-vps.sh falló en PostgreSQL:
sudo su -l postgres -c "initdb --locale=C.UTF-8 --encoding=UTF8 -D '/var/lib/postgres/data'"
sudo systemctl start postgresql
sudo ./continue-install.sh
```

---

### 🗑️ `uninstall.sh` - Desinstalador

Elimina completamente AYMC del sistema.

**Uso:**
```bash
sudo ./uninstall.sh
```

**Elimina:**
- Binarios (`/opt/aymc`)
- Servicios systemd
- Configuración (`/etc/aymc`)
- Logs (`/var/log/aymc`)
- Usuario `aymc`

**Pregunta antes de eliminar:**
- Base de datos PostgreSQL
- Datos de servidores (`/var/aymc`)
- Reglas de firewall

---

## 🎯 Flujo de Trabajo Típico

### 1. Desarrollo Local (Compilar)

```bash
cd /path/to/AYMC
./scripts/build.sh
```

### 2. Transferir a VPS

```bash
scp build/aymc-*.tar.gz user@your-vps:/tmp/
```

### 3. Instalar en VPS

```bash
ssh user@your-vps
cd /tmp
tar -xzf aymc-*.tar.gz
sudo ./install-vps.sh
```

### 4. Verificar Instalación

```bash
# Estado de servicios
sudo systemctl status aymc-backend aymc-agent

# API funcionando
curl http://localhost:8080/health

# Ver logs
sudo journalctl -u aymc-backend -f
```

---

## 🐛 Troubleshooting

### PostgreSQL no inicia (Arch Linux)

```bash
# Inicializar manualmente
sudo su -l postgres -c "initdb --locale=C.UTF-8 --encoding=UTF8 -D '/var/lib/postgres/data'"
sudo systemctl start postgresql

# Luego continuar instalación
sudo ./continue-install.sh
```

### Backend no inicia

```bash
# Ver logs detallados
sudo journalctl -u aymc-backend -n 100

# Verificar configuración
sudo cat /etc/aymc/backend.env

# Verificar puerto
sudo ss -tlnp | grep 8080
```

### Agent no inicia

```bash
# Ver logs detallados
sudo journalctl -u aymc-agent -n 100

# Verificar Java
java -version

# Verificar permisos
sudo chown aymc:aymc /var/aymc -R
```

---

## 📝 Archivos Generados

Después de la instalación:

```
/opt/aymc/
├── backend/
│   └── aymc-backend
└── agent/
    └── aymc-agent

/etc/aymc/
├── backend.env          # Configuración backend
└── agent.json          # Configuración agent

/var/aymc/
├── servers/            # Servidores Minecraft
├── backups/            # Respaldos
└── uploads/            # Archivos subidos

/var/log/aymc/
├── backend.log
├── backend-error.log
├── agent.log
└── agent-error.log

/etc/systemd/system/
├── aymc-backend.service
└── aymc-agent.service
```

---

## 🔐 Seguridad

Los scripts de instalación:

- ✅ Crean usuario dedicado `aymc` (sin login)
- ✅ Generan contraseñas aleatorias (25 caracteres)
- ✅ Generan JWT secrets aleatorios (64 caracteres)
- ✅ Establecen permisos restrictivos (750/640)
- ✅ Configuran sandboxing en servicios systemd

**IMPORTANTE:** Después de instalar:
1. Cambia las contraseñas en `/etc/aymc/backend.env`
2. Configura CORS con tu dominio real
3. Considera usar HTTPS (Nginx + Let's Encrypt)

---

## 📚 Documentación Completa

Para más detalles, consulta:

- **`docs/INSTALL_VPS.md`** - Guía completa de instalación
- **`docs/TEST_INSTALL_ARCH.md`** - Testing en Arch Linux
- **`docs/VPS_ERRORS_FIXED.md`** - Errores encontrados y soluciones
- **`docs/INSTALLATION_SUMMARY.md`** - Resumen ejecutivo

---

## ✅ Estado

- **Versión:** 1.0.0
- **Última actualización:** 13 de Noviembre 2025
- **Estado:** ✅ PRODUCCIÓN READY
- **Errores conocidos:** 0

---

## 🆘 Soporte

Si encuentras algún problema:

1. Revisa logs: `sudo journalctl -u aymc-backend -n 100`
2. Verifica estado: `sudo systemctl status aymc-backend aymc-agent`
3. Consulta documentación: `docs/INSTALL_VPS.md`
4. Reporta issue con logs completos

---

**¡Disfruta de AYMC!** 🎮🚀
