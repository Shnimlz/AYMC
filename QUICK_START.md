# 🚀 Guía de Inicio Rápido - AYMC

## 📌 Para Usuarios Finales

Esta guía te ayudará a instalar y usar AYMC en **menos de 10 minutos**.

### ¿Qué es AYMC?

AYMC es un gestor de servidores Minecraft que te permite:
- ✅ Crear y gestionar múltiples servidores Minecraft desde una interfaz moderna
- ✅ Instalar plugins con un clic
- ✅ Hacer backups automáticos
- ✅ Monitorear CPU, RAM y jugadores en tiempo real
- ✅ Gestionar servidores en múltiples VPS

---

## 🖥️ Paso 1: Instalar el Backend (Servidor)

El backend debe instalarse en un **VPS** (servidor en la nube) o en tu **PC** si quieres usarlo localmente.

### Opción A: Instalación Rápida (VPS con Linux)

```bash
# Descargar
wget https://github.com/tuusuario/aymc/releases/latest/download/aymc-latest-linux-amd64.tar.gz

# Extraer
tar -xzf aymc-latest-linux-amd64.tar.gz
cd aymc

# Instalar (requiere sudo)
sudo ./install-vps.sh
```

**Sistemas soportados:**
- Ubuntu 20.04/22.04/24.04
- Debian 11/12
- Arch Linux / Manjaro
- RHEL / CentOS / Rocky / AlmaLinux 8/9
- Fedora 38+

### ¿Qué hace el instalador?

1. ✅ Instala PostgreSQL (base de datos)
2. ✅ Instala Java 17 (para Minecraft)
3. ✅ Configura Backend API (puerto 8080)
4. ✅ Configura Agent (puerto 50051)
5. ✅ Crea servicios systemd (auto-inicio)
6. ✅ Configura firewall
7. ✅ Genera contraseñas seguras

**Tiempo estimado:** 3-5 minutos

### Verificar instalación

```bash
# Ver estado de servicios
sudo systemctl status aymc-backend aymc-agent

# Probar API
curl http://localhost:8080/health
# Debe devolver: {"status":"healthy"}
```

---

## 💻 Paso 2: Instalar la Aplicación de Escritorio

### Windows

1. Descarga: `SeraMC-Setup.exe`
2. Ejecuta el instalador
3. Sigue el asistente

### Linux

```bash
# Debian/Ubuntu
wget https://github.com/tuusuario/aymc/releases/latest/download/sera-mc_1.0.0_amd64.deb
sudo dpkg -i sera-mc_1.0.0_amd64.deb

# Arch Linux
wget https://github.com/tuusuario/aymc/releases/latest/download/sera-mc-1.0.0-1-x86_64.pkg.tar.zst
sudo pacman -U sera-mc-1.0.0-1-x86_64.pkg.tar.zst

# AppImage (universal)
wget https://github.com/tuusuario/aymc/releases/latest/download/sera-mc_1.0.0_amd64.AppImage
chmod +x sera-mc_1.0.0_amd64.AppImage
./sera-mc_1.0.0_amd64.AppImage
```

### macOS

1. Descarga: `SeraMC.dmg`
2. Abre el DMG
3. Arrastra SeraMC a Aplicaciones

---

## 🔗 Paso 3: Conectar la App al Backend

### Primera vez que abres la app

1. **Se abrirá una ventana de configuración**
2. **Ingresa la URL del backend:**

   **Si el backend está en tu PC:**
   ```
   http://localhost:8080
   ```

   **Si el backend está en un VPS:**
   ```
   https://tu-dominio.com
   # o
   http://tu-vps-ip:8080
   ```

3. **Clic en "Guardar"**

### Probar conexión

La app intentará conectar automáticamente. Si aparece "✅ Conectado", ¡estás listo!

---

## 👤 Paso 4: Crear tu Cuenta

1. **Clic en "Registrarse"**
2. **Ingresa tus datos:**
   - Usuario: `admin`
   - Email: `admin@tudominio.com`
   - Contraseña: (mínimo 8 caracteres, 1 mayúscula, 1 número)
3. **Clic en "Registrar"**

---

## 🎮 Paso 5: Crear tu Primer Servidor

### 5.1 Registrar un Agent

Antes de crear servidores, necesitas registrar dónde se ejecutarán:

1. Ve a **"Agents"** (menú lateral)
2. Clic en **"+ Agregar Agent"**
3. Ingresa:
   - **Agent ID**: `mi-vps` (o cualquier nombre)
   - **Hostname**: `servidor-minecraft-1`
   - **IP Address**: IP de tu VPS (o `127.0.0.1` si es local)
   - **Port**: `50051`
4. Clic en **"Guardar"**

### 5.2 Crear el Servidor

1. Ve a **"Servidores"** → **"+ Crear Servidor"**
2. Configura:
   - **Nombre**: `survival` (identificador interno)
   - **Display Name**: `Mi Servidor Survival`
   - **Tipo**: `Paper` (recomendado)
   - **Versión**: `1.20.1`
   - **Puerto**: `25565`
   - **RAM Mínima**: `1024` MB
   - **RAM Máxima**: `2048` MB
   - **Agent**: Selecciona el agent que registraste
3. Clic en **"Crear"**

### 5.3 Iniciar el Servidor

1. En la lista de servidores, busca tu servidor
2. Clic en **"▶ Iniciar"**
3. Ve a **"Logs"** para ver el progreso
4. Cuando veas `Done! For help, type "help"`, ¡está listo!

### 5.4 Conectarse desde Minecraft

1. Abre Minecraft Java Edition
2. Multiplayer → Add Server
3. **Server Address:**
   - Local: `localhost:25565`
   - VPS: `tu-vps-ip:25565` o `tu-dominio.com:25565`
4. ¡Conéctate y juega!

---

## 🔌 Paso 6: Instalar Plugins

1. Ve a **"Marketplace"** (menú lateral)
2. Busca el plugin deseado, por ejemplo: `EssentialsX`
3. Clic en **"Ver Detalles"**
4. Selecciona tu servidor
5. Clic en **"Instalar"**
6. Espera a que termine la descarga
7. **Reinicia el servidor** para que el plugin se cargue

### Plugins populares recomendados:

- **EssentialsX**: Comandos básicos (/home, /spawn, /warp)
- **LuckPerms**: Sistema de permisos
- **WorldEdit**: Editor de mundos
- **Vault**: Economía y permisos
- **CoreProtect**: Protección contra griefing

---

## 💾 Paso 7: Configurar Backups Automáticos

1. Ve a tu servidor → **"Backups"**
2. Clic en **"Configuración"**
3. Activa **"Backups Automáticos"**
4. Configura:
   - **Frecuencia**: `0 2 * * *` (todos los días a las 2 AM)
   - **Retención**: `7` (mantener 7 backups)
   - **Incluir**: World ✅, Plugins ✅, Config ✅
5. Clic en **"Guardar"**

### Hacer backup manual

1. Ve a **"Backups"** → **"Crear Backup"**
2. Ingresa un nombre: `antes-de-actualizar`
3. Clic en **"Crear"**

### Restaurar un backup

1. **⚠️ DETÉN el servidor primero**
2. Ve a **"Backups"**
3. Busca el backup a restaurar
4. Clic en **"⚙️"** → **"Restaurar"**
5. Confirma la operación
6. **Inicia el servidor**

---

## 📊 Paso 8: Monitorear tu Servidor

1. Ve a tu servidor → **"Dashboard"**
2. Verás en tiempo real:
   - **CPU**: Uso del procesador
   - **RAM**: Memoria utilizada
   - **Jugadores**: Cantidad online
   - **TPS**: Ticks Por Segundo (salud del servidor)

### Alertas

Si el TPS baja de 15 o la RAM supera 90%, recibirás una alerta.

---

## ❓ FAQ - Preguntas Frecuentes

### ¿Puedo gestionar varios servidores?

**Sí**, puedes crear ilimitados servidores (limitado por los recursos de tu VPS).

### ¿Puedo tener servidores en diferentes VPS?

**Sí**, registra un Agent por cada VPS y selecciona el correspondiente al crear el servidor.

### ¿Es gratis?

**Sí**, AYMC es de código abierto y gratuito. Solo pagas por tu VPS.

### ¿Funciona con Bedrock?

**No**, actualmente solo soporta Minecraft Java Edition.

### ¿Qué versiones de Minecraft soporta?

Desde 1.8 hasta la última versión (1.20+).

### ¿Puedo migrar mi servidor existente?

**Sí**:
1. Detén tu servidor actual
2. Copia la carpeta del servidor a `/var/aymc/servers/nombre-servidor/`
3. Crea el servidor en AYMC con el mismo nombre
4. Inicia el servidor

### Mi servidor no inicia, ¿qué hago?

1. Ve a **"Logs"** para ver el error
2. Verifica que tengas suficiente RAM
3. Verifica que el puerto no esté en uso
4. Revisa que Java esté instalado: `java -version`

### ¿Cómo actualizo AYMC?

```bash
# Backend (en el VPS)
cd /opt/aymc
sudo ./uninstall.sh
# Descargar nueva versión e instalar

# Frontend (en tu PC)
# Descarga el nuevo instalador y ejecuta
```

---

## 🆘 Soporte

### Algo no funciona:

1. **Revisa los logs:**
   ```bash
   sudo journalctl -u aymc-backend -n 50
   sudo journalctl -u aymc-agent -n 50
   ```

2. **Consulta la documentación completa:**
   - [README Principal](./README.md)
   - [Guía de Instalación VPS](./docs/INSTALL_VPS.md)
   - [Troubleshooting](./docs/VPS_ERRORS_FIXED.md)

3. **Contacta:**
   - 📧 Email: soporte@aymc.com
   - 💬 Discord: [Servidor AYMC](https://discord.gg/aymc)
   - 🐛 GitHub Issues: [Reportar problema](https://github.com/tuusuario/aymc/issues)

---

## 🎉 ¡Listo!

Ahora tienes tu servidor Minecraft funcionando con AYMC. Disfruta de:

- ✅ Gestión moderna y fácil
- ✅ Backups automáticos
- ✅ Plugins con un clic
- ✅ Monitoreo en tiempo real
- ✅ Sin consolas de comandos

**¡Diviértete jugando!** 🎮

---

💡 **Tip Pro**: Únete a nuestro Discord para compartir tu experiencia y obtener ayuda de la comunidad.
