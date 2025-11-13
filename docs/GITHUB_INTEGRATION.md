# 🐙 Integración con GitHub - AYMC

## 📋 Descripción

Este documento detalla los cambios realizados para permitir que los scripts de instalación descarguen automáticamente el backend y el agent desde el repositorio público de GitHub en lugar de copiarlos localmente.

## 🔗 Repositorio Público

**URL:** https://github.com/Shnimlz/AYMC

## 📝 Cambios Realizados

### 1. Script `install-vps.sh`

**Ubicación:** `/SeraMC/src-tauri/resources/install-vps.sh`

#### Cambios Principales:

1. **Dependencia Git Agregada** (Líneas 87-119)
   - Se agregó `git` a las dependencias instaladas en todas las distribuciones:
     - Arch/Manjaro: `pacman -Sy --noconfirm --needed git ...`
     - Debian/Ubuntu: `apt-get install -y git ...`
     - RHEL/CentOS: `yum install -y git ...`

2. **Nueva Sección: Descarga desde GitHub** (Reemplaza líneas 189-231)

```bash
#################################################
# DESCARGAR E INSTALAR DESDE GITHUB
#################################################

GITHUB_REPO="https://github.com/Shnimlz/AYMC.git"
TEMP_DIR="/tmp/aymc-install-$$"

# Clonar repositorio
git clone --depth 1 "$GITHUB_REPO" "$TEMP_DIR/aymc"

# Instalar backend y agent desde el repositorio clonado
# Con búsqueda automática de binarios si no están en la ubicación esperada

# Limpiar directorio temporal al finalizar
rm -rf "$TEMP_DIR"
```

**Características del nuevo sistema:**

✅ **Clonación Shallow** (`--depth 1`) - Solo la última versión, ahorra ancho de banda
✅ **Directorio Temporal** - Usa PID único para evitar conflictos
✅ **Verificación Git** - Instala git automáticamente si no está presente
✅ **Búsqueda Inteligente** - Si los binarios no están en la ruta esperada, busca en subdirectorios
✅ **Configuración Por Defecto** - Crea archivos `.env` y `.json` si no existen en el repo
✅ **Limpieza Automática** - Elimina archivos temporales al finalizar
✅ **Manejo de Errores** - Exit codes y mensajes detallados en caso de fallo

3. **Mensaje Final Actualizado** (Líneas 510-533)
   - Agregada referencia al repositorio GitHub
   - Nota sobre origen de los binarios

### 2. Script `continue-install.sh`

**Ubicación:** `/SeraMC/src-tauri/resources/continue-install.sh`

**Estado:** ✅ No requiere cambios

Este script solo continúa la instalación desde la configuración de PostgreSQL y no descarga binarios. Asume que `install-vps.sh` ya instaló los binarios correctamente.

## 🏗️ Estructura Esperada del Repositorio GitHub

Para que los scripts funcionen correctamente, el repositorio debe tener esta estructura:

```
AYMC/
├── backend/
│   └── aymc-backend          # Binario del backend (Linux x86_64)
├── agent/
│   └── aymc-agent            # Binario del agent (Linux x86_64)
└── config/                   # Opcional - configuraciones por defecto
    ├── backend.env           # Variables de entorno del backend
    └── agent.json            # Configuración del agent
```

### Configuraciones Por Defecto

Si el repositorio no incluye archivos de configuración, el script creará estos por defecto:

**`backend.env`:**
```bash
DB_HOST=localhost
DB_PORT=5432
DB_NAME=aymc
DB_USER=aymc
DB_PASSWORD=changeme
JWT_SECRET=changeme
CORS_ORIGINS=http://localhost:3000,http://localhost:5173
API_PORT=8080
LOG_LEVEL=info
```

**`agent.json`:**
```json
{
  "grpc_port": 50051,
  "backend_url": "http://localhost:8080",
  "data_dir": "/var/aymc",
  "log_level": "info",
  "max_servers": 20
}
```

## 🔄 Proceso de Instalación Actualizado

### Antes (Sistema Local):
1. Usuario ejecuta script embebido en Tauri
2. Script copia binarios desde `$SCRIPT_DIR/backend/` y `$SCRIPT_DIR/agent/`
3. Instala en `/opt/aymc/`

### Ahora (Sistema GitHub):
1. Usuario ejecuta script embebido en Tauri
2. Script verifica/instala Git
3. Clona repositorio GitHub en directorio temporal
4. Copia binarios desde repositorio clonado a `/opt/aymc/`
5. Copia/crea configuraciones
6. Limpia archivos temporales

## 📦 Ventajas del Nuevo Sistema

✅ **Actualizaciones Centralizadas** - Un solo lugar para actualizar binarios
✅ **Sin Embedimiento** - No necesita embeberlos en la app Tauri (reduce tamaño)
✅ **Versionado** - Git tags/releases pueden controlar versiones específicas
✅ **Transparencia** - Usuario puede ver exactamente qué se está instalando
✅ **Distribución Más Fácil** - Solo compartir URL del repositorio
✅ **CI/CD Ready** - GitHub Actions puede compilar y pushear binarios automáticamente

## 🚀 Uso

### Instalación Normal:
```bash
# Desde SeraMC (app Tauri), el usuario hace clic en "Instalar en VPS"
# El script se ejecuta automáticamente vía SSH y descarga desde GitHub
```

### Instalación Manual:
```bash
# Si el usuario quiere ejecutar el script manualmente:
curl -fsSL https://raw.githubusercontent.com/Shnimlz/AYMC/main/scripts/install-vps.sh | sudo bash
```

## 🔐 Consideraciones de Seguridad

⚠️ **IMPORTANTE:** Actualmente los binarios se descargan vía HTTPS sin verificación de firma.

### Recomendaciones para Producción:

1. **Firmar Binarios:**
   ```bash
   # Generar firma SHA512
   sha512sum aymc-backend > aymc-backend.sha512
   ```

2. **Verificar en Script:**
   ```bash
   # Agregar verificación después de clonar
   sha512sum -c aymc-backend.sha512 || exit 1
   ```

3. **GitHub Releases:**
   - Usar releases en lugar de branch main
   - Incluir checksums en cada release
   - Firmar releases con GPG

4. **Subresource Integrity (SRI):**
   - Implementar verificación de integridad
   - Hash hardcoded en script para versión específica

## 🧪 Testing

### Verificar que el script funciona:

```bash
# 1. Clonar tu repositorio manualmente
git clone --depth 1 https://github.com/Shnimlz/AYMC.git /tmp/test-aymc

# 2. Verificar que existen los binarios
ls -lh /tmp/test-aymc/backend/aymc-backend
ls -lh /tmp/test-aymc/agent/aymc-agent

# 3. Verificar que son ejecutables
file /tmp/test-aymc/backend/aymc-backend
file /tmp/test-aymc/agent/aymc-agent

# 4. Limpiar
rm -rf /tmp/test-aymc
```

### Test de instalación completa (VPS limpia):

```bash
# Ejecutar script de instalación
sudo bash /path/to/install-vps.sh

# Verificar servicios
systemctl status aymc-backend
systemctl status aymc-agent

# Verificar logs
journalctl -u aymc-backend -n 50
journalctl -u aymc-agent -n 50
```

## 📚 Documentación Relacionada

- [FASE_2_SCRIPTS_EMBEBIDOS_COMPLETADO.md](./FASE_2_SCRIPTS_EMBEBIDOS_COMPLETADO.md)
- [PROYECTO_COMPLETO_FASES_1-6.md](./PROYECTO_COMPLETO_FASES_1-6.md)
- [main.instructions.md](../.github/instructions/main.instructions.md)

## 🔄 Changelog

### v1.1.0 (Fecha Actual)
- ✅ Integración con repositorio GitHub público
- ✅ Instalación automática de Git si no está presente
- ✅ Búsqueda inteligente de binarios
- ✅ Configuraciones por defecto si no existen en repo
- ✅ Limpieza automática de temporales
- ✅ Manejo robusto de errores

### v1.0.0
- Instalación desde archivos locales embebidos

## 🤝 Contribuciones

Para actualizar los binarios en el repositorio:

1. Compilar backend y agent
2. Copiar a sus respectivas carpetas
3. Commit y push:
   ```bash
   git add backend/aymc-backend agent/aymc-agent
   git commit -m "Update binaries to version X.Y.Z"
   git push origin main
   ```

## 📞 Soporte

Si tienes problemas con la instalación desde GitHub:

1. Verifica conectividad: `curl -I https://github.com/Shnimlz/AYMC`
2. Verifica que Git esté instalado: `git --version`
3. Revisa logs de instalación en `/var/log/aymc/`
4. Abre un issue en GitHub: https://github.com/Shnimlz/AYMC/issues

---

**Última actualización:** Enero 2025
**Autor:** AYMC Team
**Licencia:** [Especificar licencia del proyecto]
