# 🪟 Guía de Compilación para Windows - AYMC SeraMC

Esta guía te ayudará a compilar la aplicación **AYMC SeraMC** para Windows desde tu sistema Linux actual o directamente en Windows.

---

## 📋 Tabla de Contenidos

1. [Requisitos Previos](#requisitos-previos)
2. [Opción 1: Compilación Cruzada desde Linux](#opción-1-compilación-cruzada-desde-linux-recomendado)
3. [Opción 2: Compilación Nativa en Windows](#opción-2-compilación-nativa-en-windows)
4. [Configuración del Proyecto](#configuración-del-proyecto)
5. [Proceso de Compilación](#proceso-de-compilación)
6. [Distribución](#distribución)
7. [Troubleshooting](#troubleshooting)

---

## 📦 Requisitos Previos

### Para Linux (Compilación Cruzada)

```bash
# 1. Instalar Rust con soporte Windows
rustup target add x86_64-pc-windows-gnu

# 2. Instalar MinGW (compilador Windows)
# Arch Linux:
sudo pacman -S mingw-w64-gcc

# Ubuntu/Debian:
sudo apt install mingw-w64

# 3. Instalar dependencias adicionales
# Arch:
sudo pacman -S wine

# Ubuntu/Debian:
sudo apt install wine64
```

### Para Windows (Compilación Nativa)

1. **Node.js** (v18 o superior): https://nodejs.org/
2. **Rust**: https://rustup.rs/
3. **Visual Studio Build Tools** (MSVC): 
   - Descarga: https://visualstudio.microsoft.com/downloads/
   - Instala "Desktop development with C++"
4. **WebView2**: Preinstalado en Windows 11, descarga para Windows 10: https://go.microsoft.com/fwlink/p/?LinkId=2124703

---

## 🔨 Opción 1: Compilación Cruzada desde Linux (Recomendado)

### Paso 1: Configurar Target de Windows

```bash
cd /home/shni/Documents/GitHub/AYMC/SeraMC

# Agregar target Windows
rustup target add x86_64-pc-windows-gnu
```

### Paso 2: Configurar Cargo para Cross-Compilation

Crea o edita `~/.cargo/config.toml`:

```toml
[target.x86_64-pc-windows-gnu]
linker = "x86_64-w64-mingw32-gcc"
ar = "x86_64-w64-mingw32-ar"
```

### Paso 3: Instalar Dependencias de Node

```bash
npm install
```

### Paso 4: Compilar para Windows

```bash
# Build de producción para Windows
npm run tauri build -- --target x86_64-pc-windows-gnu
```

**Resultado esperado:**
```
/home/shni/Documents/GitHub/AYMC/SeraMC/src-tauri/target/x86_64-pc-windows-gnu/release/
├── seramc.exe              # Ejecutable principal
└── bundle/
    ├── msi/
    │   └── seramc_0.1.0_x64_en-US.msi    # Instalador MSI
    └── nsis/
        └── seramc_0.1.0_x64-setup.exe    # Instalador NSIS
```

### Limitaciones de Compilación Cruzada

⚠️ **IMPORTANTE:** La compilación cruzada desde Linux puede tener problemas con:
- Dependencias nativas de Windows (ssh2, tokio)
- Generación de instaladores MSI/NSIS
- Iconos y recursos embebidos

**Solución:** Si falla, usa la **Opción 2** (compilación nativa en Windows).

---

## 🪟 Opción 2: Compilación Nativa en Windows

### Paso 1: Preparar el Proyecto

**En tu Linux actual:**

```bash
cd /home/shni/Documents/GitHub/AYMC

# Crear un archivo .zip del proyecto
zip -r SeraMC-windows-build.zip SeraMC/ \
  -x "SeraMC/node_modules/*" \
  -x "SeraMC/dist/*" \
  -x "SeraMC/src-tauri/target/*"

# Transferir a Windows (USB, email, cloud, etc.)
```

**En Windows:**

1. Extrae `SeraMC-windows-build.zip`
2. Abre PowerShell o CMD como Administrador
3. Navega a la carpeta extraída

### Paso 2: Instalar Dependencias de Node

```powershell
cd C:\path\to\SeraMC

# Instalar dependencias
npm install
```

### Paso 3: Compilar la Aplicación

```powershell
# Build de desarrollo (rápido, para testing)
npm run tauri:dev

# Build de producción (optimizado, para distribución)
npm run tauri:build
```

### Paso 4: Resultado de la Compilación

**Ubicación de archivos:**
```
C:\path\to\SeraMC\src-tauri\target\release\
├── seramc.exe              # Ejecutable standalone (23-35 MB)
└── bundle\
    ├── msi\
    │   └── seramc_0.1.0_x64_en-US.msi        # Instalador MSI (25-40 MB)
    └── nsis\
        └── seramc_0.1.0_x64-setup.exe        # Instalador NSIS (25-40 MB)
```

---

## ⚙️ Configuración del Proyecto

### 1. Actualizar Metadatos de la App

Edita `/SeraMC/src-tauri/tauri.conf.json`:

```json
{
  "productName": "AYMC SeraMC",
  "version": "1.0.0",
  "identifier": "com.shni.aymc.seramc",
  "bundle": {
    "active": true,
    "targets": ["msi", "nsis"],
    "windows": {
      "certificateThumbprint": null,
      "digestAlgorithm": "sha256",
      "timestampUrl": ""
    },
    "icon": [
      "icons/32x32.png",
      "icons/128x128.png",
      "icons/128x128@2x.png",
      "icons/icon.icns",
      "icons/icon.ico"
    ]
  }
}
```

### 2. Actualizar Información del Cargo.toml

Edita `/SeraMC/src-tauri/Cargo.toml`:

```toml
[package]
name = "seramc"
version = "1.0.0"
description = "AYMC - Advanced Minecraft Control Panel"
authors = ["Shni <tu@email.com>"]
edition = "2021"

[package.metadata.bundle]
name = "AYMC SeraMC"
identifier = "com.shni.aymc.seramc"
```

### 3. Iconos para Windows

Asegúrate de tener un buen `icon.ico` en `/SeraMC/src-tauri/icons/`:

```bash
# Generar icon.ico desde PNG (en Linux con ImageMagick)
cd /home/shni/Documents/GitHub/AYMC/SeraMC/src-tauri/icons
convert icon.png -define icon:auto-resize=256,128,64,48,32,16 icon.ico
```

---

## 🚀 Proceso de Compilación Completo

### Build Script Automatizado

Crea `/SeraMC/build-windows.sh`:

```bash
#!/bin/bash

echo "🏗️  AYMC SeraMC - Windows Build Script"
echo "========================================"
echo ""

# 1. Limpiar builds anteriores
echo "🧹 Limpiando builds anteriores..."
rm -rf dist/
rm -rf src-tauri/target/x86_64-pc-windows-gnu/release/bundle/

# 2. Instalar/actualizar dependencias
echo "📦 Instalando dependencias de Node..."
npm install

# 3. Build del frontend
echo "🎨 Compilando frontend Vue..."
npm run build

# 4. Build de Tauri para Windows
echo "🪟 Compilando Tauri para Windows..."
npm run tauri build -- --target x86_64-pc-windows-gnu

# 5. Verificar resultado
echo ""
echo "✅ Build completado!"
echo ""
echo "📁 Archivos generados:"
echo "   Ejecutable: src-tauri/target/x86_64-pc-windows-gnu/release/seramc.exe"
echo "   Instalador MSI: src-tauri/target/x86_64-pc-windows-gnu/release/bundle/msi/"
echo "   Instalador NSIS: src-tauri/target/x86_64-pc-windows-gnu/release/bundle/nsis/"
echo ""
```

Dale permisos de ejecución:

```bash
chmod +x build-windows.sh
./build-windows.sh
```

---

## 📦 Distribución

### Opción 1: Instalador MSI (Recomendado para Empresas)

**Ventajas:**
- Instalación silenciosa (`msiexec /i seramc.msi /quiet`)
- Desinstalación desde Panel de Control
- Soporte para Group Policy

**Usar cuando:**
- Distribución corporativa
- Necesitas instalación automatizada
- Usuarios finales con permisos de administrador

### Opción 2: Instalador NSIS

**Ventajas:**
- Interfaz de instalación personalizable
- Menor tamaño
- Más opciones de configuración durante instalación

**Usar cuando:**
- Distribución pública
- Necesitas wizard de instalación
- Quieres seleccionar componentes opcionales

### Opción 3: Ejecutable Portable

**Ventajas:**
- No requiere instalación
- Ejecutar desde USB
- No requiere permisos de administrador

**Crear versión portable:**

```bash
# Copiar ejecutable y dependencias
mkdir AYMC-SeraMC-Portable
cp src-tauri/target/release/seramc.exe AYMC-SeraMC-Portable/
cp -r src-tauri/resources/ AYMC-SeraMC-Portable/resources/

# Crear archivo de configuración portable
cat > AYMC-SeraMC-Portable/portable.txt << EOF
Este archivo indica que AYMC se ejecuta en modo portable.
Los datos se guardarán en la carpeta 'data' junto al ejecutable.
EOF

# Comprimir
zip -r AYMC-SeraMC-v1.0.0-Portable.zip AYMC-SeraMC-Portable/
```

---

## 🎯 Firmar la Aplicación (Opcional pero Recomendado)

### ¿Por qué firmar?

- Windows SmartScreen no bloqueará tu app
- Usuarios verán "Editor verificado"
- Mayor confianza y profesionalismo

### Cómo obtener certificado:

1. **Certificado de Code Signing**
   - DigiCert: ~$300/año
   - Sectigo: ~$200/año
   - Let's Encrypt: No ofrece code signing

2. **Firmar el ejecutable**

```powershell
# En Windows con SignTool (viene con Visual Studio)
signtool sign /f "certificado.pfx" /p "contraseña" /t "http://timestamp.digicert.com" seramc.exe
```

3. **Configurar en tauri.conf.json**

```json
{
  "bundle": {
    "windows": {
      "certificateThumbprint": "THUMBPRINT_DEL_CERTIFICADO",
      "digestAlgorithm": "sha256",
      "timestampUrl": "http://timestamp.digicert.com"
    }
  }
}
```

---

## 🐛 Troubleshooting

### Error: "rustup target not found"

```bash
rustup update
rustup target add x86_64-pc-windows-gnu
```

### Error: "linker `x86_64-w64-mingw32-gcc` not found"

```bash
# Arch:
sudo pacman -S mingw-w64-gcc

# Ubuntu/Debian:
sudo apt install gcc-mingw-w64-x86-64
```

### Error: "failed to run custom build command for `openssl-sys`"

**Solución 1:** Usar vendored OpenSSL

Agrega a `Cargo.toml`:

```toml
[dependencies]
openssl = { version = "0.10", features = ["vendored"] }
```

**Solución 2:** Instalar OpenSSL para MinGW

```bash
# Arch:
sudo pacman -S mingw-w64-openssl

# Ubuntu:
sudo apt install libssl-dev:i386
```

### Error: "WebView2 not found"

**En Windows:**
- Descarga e instala: https://go.microsoft.com/fwlink/p/?LinkId=2124703

### Error: "Resource not found: resources/*.sh"

Los scripts `.sh` son para Linux. En Windows, Tauri los ignora automáticamente, pero puedes actualizar `tauri.conf.json`:

```json
{
  "bundle": {
    "resources": {
      "resources/*.sh": "./resources/"
    }
  }
}
```

### Build muy lento

**Optimizaciones:**

1. **Usar cache de Cargo:**

```bash
# Instalar sccache
cargo install sccache

# Configurar
export RUSTC_WRAPPER=sccache
```

2. **Parallel compilation:**

```bash
# En ~/.cargo/config.toml
[build]
jobs = 8  # Número de núcleos de tu CPU
```

3. **Link-time optimization (LTO):**

Edita `src-tauri/Cargo.toml`:

```toml
[profile.release]
lto = true
codegen-units = 1
opt-level = "z"  # Optimizar para tamaño
```

---

## 📊 Comparación de Métodos

| Método | Tiempo | Complejidad | Compatibilidad | Recomendado |
|--------|--------|-------------|----------------|-------------|
| **Cross-compile (Linux → Windows)** | 5-10 min | Alta | Media | Testing rápido |
| **Native build (Windows)** | 10-15 min | Baja | Alta | ✅ Producción |
| **GitHub Actions CI/CD** | 15-20 min | Media | Alta | ✅ Automatización |

---

## 🤖 Bonus: Automatización con GitHub Actions

Crea `.github/workflows/build-windows.yml`:

```yaml
name: Build Windows

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: windows-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: 20
      
      - name: Setup Rust
        uses: dtolnay/rust-toolchain@stable
      
      - name: Install dependencies
        run: npm install
        working-directory: ./SeraMC
      
      - name: Build Tauri app
        run: npm run tauri:build
        working-directory: ./SeraMC
      
      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: windows-build
          path: |
            SeraMC/src-tauri/target/release/bundle/msi/*.msi
            SeraMC/src-tauri/target/release/bundle/nsis/*.exe
```

---

## 📝 Checklist Final

Antes de distribuir tu aplicación:

- [ ] Actualizar versión en `tauri.conf.json` y `Cargo.toml`
- [ ] Probar instalador MSI en Windows limpia
- [ ] Probar instalador NSIS en Windows limpia
- [ ] Verificar que los scripts SSH funcionen
- [ ] Verificar que la conexión a backend funcione
- [ ] Crear documentación de usuario
- [ ] Firmar ejecutable (opcional)
- [ ] Crear release notes
- [ ] Subir a GitHub Releases

---

## 🎉 ¡Listo!

Tu aplicación **AYMC SeraMC** estará lista para distribuir en Windows. Los usuarios podrán:

1. Descargar el instalador MSI/NSIS
2. Instalar con doble clic
3. Ejecutar desde el menú Inicio
4. Conectarse a sus VPS remotos
5. Instalar y gestionar servidores Minecraft

---

**¿Necesitas ayuda?** Abre un issue en: https://github.com/Shnimlz/AYMC/issues

**Última actualización:** Enero 2025  
**Autor:** AYMC Team
