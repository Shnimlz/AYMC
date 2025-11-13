# SeraMC - AYMC Frontend# Tauri + Vue + TypeScript



Aplicación de escritorio multiplataforma construida con Vue 3 + Tauri para gestionar servidores Minecraft.This template should help get you started developing with Vue 3 and TypeScript in Vite. The template uses Vue 3 `<script setup>` SFCs, check out the [script setup docs](https://v3.vuejs.org/api/sfc-script-setup.html#sfc-script-setup) to learn more.



## 🚀 Inicio Rápido## Recommended IDE Setup



### 1. Instalar Dependencias- [VS Code](https://code.visualstudio.com/) + [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) + [Tauri](https://marketplace.visualstudio.com/items?itemName=tauri-apps.tauri-vscode) + [rust-analyzer](https://marketplace.visualstudio.com/items?itemName=rust-lang.rust-analyzer)



```bash## Type Support For `.vue` Imports in TS

npm install

```Since TypeScript cannot handle type information for `.vue` imports, they are shimmed to be a generic Vue component type by default. In most cases this is fine if you don't really care about component prop types outside of templates. However, if you wish to get actual prop types in `.vue` imports (for example to get props validation when using manual `h(...)` calls), you can enable Volar's Take Over mode by following these steps:



### 2. Configurar Conexión al Backend1. Run `Extensions: Show Built-in Extensions` from VS Code's command palette, look for `TypeScript and JavaScript Language Features`, then right click and select `Disable (Workspace)`. By default, Take Over mode will enable itself if the default TypeScript extension is disabled.

2. Reload the VS Code window by running `Developer: Reload Window` from the command palette.

**⚠️ IMPORTANTE**: Antes de ejecutar la aplicación, debes configurar la URL del backend.

You can learn more about Take Over mode [here](https://github.com/johnsoncodehk/volar/discussions/471).

#### Opción A: Backend Local (en tu PC)

Si instalaste el backend en tu máquina local:

```bash
# El archivo .env ya está configurado correctamente
cat .env
```

Debería mostrar:
```properties
VITE_API_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080/api/v1/ws
```

#### Opción B: Backend Remoto (en VPS)

Si tienes el backend en un VPS:

1. **Edita `.env`:**
```bash
nano .env
```

2. **Con HTTPS (recomendado para producción):**
```properties
VITE_API_URL=https://tu-dominio.com/api/v1
VITE_WS_URL=wss://tu-dominio.com/api/v1/ws
```

3. **Sin HTTPS (desarrollo):**
```properties
VITE_API_URL=http://tu-vps-ip:8080/api/v1
VITE_WS_URL=ws://tu-vps-ip:8080/api/v1/ws
```

### 3. Verificar Conectividad

Antes de ejecutar la app, verifica que el backend esté accesible:

```bash
# Local
curl http://localhost:8080/health

# Remoto
curl https://tu-dominio.com/health
```

✅ Respuesta esperada:
```json
{
  "status": "healthy",
  "environment": "production",
  "timestamp": "2025-11-13T19:54:28Z"
}
```

### 4. Ejecutar Aplicación

#### Desarrollo
```bash
npm run tauri dev
```

#### Compilar para Producción
```bash
npm run tauri build
```

**Ejecutables generados:**
- Windows: `src-tauri/target/release/SeraMC.exe`
- Linux: `src-tauri/target/release/sera-mc`
- macOS: `src-tauri/target/release/bundle/dmg/SeraMC.dmg`

## 🔧 Configurar CORS en Backend

Si ves errores de CORS, configura el backend para permitir conexiones desde la app:

```bash
# En el servidor
sudo nano /etc/aymc/backend.env
```

Agrega los orígenes:
```env
CORS_ORIGINS=http://localhost:1420,tauri://localhost,https://tu-dominio.com
```

Reinicia:
```bash
sudo systemctl restart aymc-backend
```

## 🐛 Solución de Problemas

### ❌ "Network Error" al hacer login

**Causa:** Frontend no puede conectar con el backend

**Solución:**
1. Verifica que el backend esté corriendo:
   ```bash
   sudo systemctl status aymc-backend
   ```

2. Verifica la URL en `.env`

3. Prueba conexión manual:
   ```bash
   curl http://localhost:8080/health
   ```

### ❌ Error de CORS

**Causa:** Backend no permite el origen del frontend

**Solución:**
```bash
# En el servidor backend
sudo nano /etc/aymc/backend.env
# Agregar: CORS_ORIGINS=http://localhost:1420,tauri://localhost

sudo systemctl restart aymc-backend
```

### ❌ WebSocket no conecta

**Verifica:**
1. URL correcta en `.env` (debe ser `ws://` o `wss://`)
2. Si usas Nginx, configuración de proxy WebSocket:

```nginx
location /api/v1/ws {
    proxy_pass http://localhost:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "Upgrade";
}
```

### ❌ Error al compilar (Rust)

**Linux:**
```bash
sudo apt install -y libwebkit2gtk-4.0-dev build-essential curl wget libssl-dev libgtk-3-dev libayatana-appindicator3-dev librsvg2-dev
```

**Arch Linux:**
```bash
sudo pacman -S --needed webkit2gtk base-devel curl wget openssl gtk3 libappindicator-gtk3 librsvg
```

## 📁 Estructura

```
SeraMC/
├── src/
│   ├── api/              # Clientes API (axios)
│   ├── components/       # Componentes Vue
│   ├── views/            # Vistas/Páginas
│   ├── stores/           # Pinia stores
│   ├── composables/      # Composables reutilizables
│   ├── router/           # Vue Router
│   └── main.ts           # Entry point
├── src-tauri/            # Código Rust (Tauri)
├── .env                  # ⚠️ CONFIGURAR AQUÍ
├── .env.example          # Ejemplo
└── package.json
```

## 🎯 Primer Uso

1. **Iniciar Sesión**
   - Email: `admin@aymc.local`
   - Password: `Test123456!` (o la que configuraste)

2. **Registrar Agent**
   - Ve a "Agents" → "Agregar Agent"
   - Ingresa IP y puerto del VPS

3. **Crear Servidor**
   - Ve a "Servidores" → "Crear Servidor"
   - Selecciona tipo, versión, RAM y agent

4. **Instalar Plugins**
   - Ve a "Marketplace"
   - Busca e instala plugins

## 📚 Documentación

- Ver [README principal](../README.md) para documentación completa
- Ver [INSTALL_VPS.md](../docs/INSTALL_VPS.md) para instalar el backend

## 🤝 Contribuir

Consulta el README principal para guías de contribución.

---

💡 **Tip**: Si no tienes backend instalado, consulta [INSTALL_VPS.md](../docs/INSTALL_VPS.md)
