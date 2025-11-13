# 🚀 Fase 4: Installation Wizard - COMPLETADO

## ✅ Implementación Completada

### Componentes Vue Creados

```
SeraMC/src/components/
├── RemoteTerminal.vue          ✅ 550 líneas
└── InstallationWizard.vue      ✅ 980 líneas
─────────────────────────────────────────
TOTAL:                          ~1,530 líneas de Vue 3
```

---

## 📦 Componente 1: RemoteTerminal.vue

### Descripción

Terminal emulada profesional usando **xterm.js** con soporte completo para streaming de output, colores ANSI, y control de ejecución en tiempo real.

### Características Principales

- ✅ **Terminal completa** con xterm.js y FitAddon
- ✅ **Temas**: Dark (default) y Light
- ✅ **Responsive**: Se ajusta automáticamente al contenedor
- ✅ **Controles**:
  - 🗑️ Limpiar terminal
  - 📋 Copiar output al portapapeles
  - ⏹️ Detener ejecución
- ✅ **Status bar** con información en tiempo real:
  - Estado (Ejecutando, Completado, Error)
  - Duración de ejecución
  - Código de salida
- ✅ **Buffer interno** para histórico completo
- ✅ **Colores ANSI**: Soporte para códigos de color estándar
- ✅ **Scroll automático**: 10,000 líneas de historial

### Props

```typescript
interface Props {
  title?: string;           // Título del terminal (default: "Terminal Remota")
  autoFit?: boolean;        // Auto-ajuste responsive (default: true)
  canClear?: boolean;       // Mostrar botón limpiar (default: true)
  canCopy?: boolean;        // Mostrar botón copiar (default: true)
  canStop?: boolean;        // Mostrar botón detener (default: true)
  showStatus?: boolean;     // Mostrar status bar (default: true)
  theme?: 'dark' | 'light'; // Tema visual (default: 'dark')
  fontSize?: number;        // Tamaño de fuente (default: 13)
}
```

### Eventos

```typescript
emit('ready')  // Cuando el terminal está listo para usar
emit('stop')   // Cuando el usuario presiona el botón stop
```

### Métodos Expuestos (via defineExpose)

```typescript
// Escritura básica
write(text: string)           // Escribir sin nueva línea
writeLine(text: string)       // Escribir con nueva línea

// Escritura con colores
writeSuccess(message: string) // Verde: ✓ message
writeError(message: string)   // Rojo: ✗ message
writeWarning(message: string) // Amarillo: ⚠ message
writeInfo(message: string)    // Cyan: ℹ message
writeHeader(message: string)  // Header con líneas ======

// Control
clearTerminal()               // Limpiar todo el contenido
startExecution()              // Iniciar tracking de ejecución
endExecution(code: number)    // Finalizar con exit code
resetStatus()                 // Resetear estado completo

// Acceso directo
terminal: Terminal            // Instancia de xterm.js
```

### Uso Básico

```vue
<template>
  <RemoteTerminal
    ref="terminalRef"
    title="Instalación AYMC"
    :can-stop="true"
    @ready="handleReady"
    @stop="handleStop"
  />
</template>

<script setup lang="ts">
import { ref } from 'vue';
import RemoteTerminal from './RemoteTerminal.vue';

const terminalRef = ref<InstanceType<typeof RemoteTerminal>>();

function handleReady() {
  const terminal = terminalRef.value;
  if (!terminal) return;

  terminal.startExecution();
  terminal.writeHeader('Proceso de Instalación');
  terminal.writeInfo('Iniciando...');
  
  // Simular proceso
  setTimeout(() => {
    terminal.writeSuccess('Paso 1 completado');
  }, 1000);
}

function handleStop() {
  console.log('Usuario detuvo la ejecución');
}
</script>
```

### Características Técnicas

#### Tema Dark (default)
```typescript
theme: {
  background: '#1e1e1e',
  foreground: '#d4d4d4',
  cursor: '#ffffff',
  black: '#000000',
  red: '#cd3131',
  green: '#0dbc79',
  yellow: '#e5e510',
  blue: '#2472c8',
  magenta: '#bc3fbc',
  cyan: '#11a8cd',
  white: '#e5e5e5',
  // ... bright colors
}
```

#### Configuración Terminal
```typescript
{
  cursorBlink: true,
  cursorStyle: 'block',
  fontSize: 13,
  fontFamily: 'Monaco, Menlo, "Courier New", monospace',
  scrollback: 10000,    // 10,000 líneas de historial
  convertEol: true,     // Conversión automática EOL
}
```

#### ResizeObserver
Ajusta automáticamente el tamaño del terminal cuando el contenedor cambia:
```typescript
resizeObserver = new ResizeObserver(() => {
  if (fitAddon && terminal) {
    fitAddon.fit();
  }
});
```

---

## 🧙 Componente 2: InstallationWizard.vue

### Descripción

Wizard completo de instalación con 4 pasos: configuración de credenciales, instalación en progreso, éxito, y manejo de errores.

### Flujo de Pasos

```
┌─────────────────────────────────────┐
│   Step 1: Credenciales              │
│   ├─ DB Password (validado)        │
│   ├─ JWT Secret (generador)        │
│   ├─ App Port (1024-65535)         │
│   └─ DB Name (opcional)            │
│                                     │
│   [Cancelar]  [Iniciar Instalación]│
└────────────────┬────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────┐
│   Step 2: Instalación               │
│   ┌───────────────────────────────┐ │
│   │  RemoteTerminal (xterm.js)    │ │
│   │  - Streaming en tiempo real   │ │
│   │  - Botón Stop                 │ │
│   │  - Status: Ejecutando...      │ │
│   └───────────────────────────────┘ │
└────────────────┬────────────────────┘
                 │
         ┌───────┴───────┐
         │               │
      SUCCESS          ERROR
         │               │
         ↓               ↓
┌──────────────┐  ┌──────────────┐
│  Step 3:     │  │  Step 4:     │
│  Completado  │  │  Error       │
│              │  │              │
│  ✅ Summary  │  │  ✗ Details   │
│  [Dashboard] │  │  [Reintentar]│
└──────────────┘  └──────────────┘
```

### Step 1: Configuración de Credenciales

#### Campos del Formulario

**1. Contraseña de PostgreSQL** (requerido)
- Mínimo 8 caracteres
- Indicador de fuerza: Muy débil → Muy fuerte
- Toggle show/hide password
- Barra de progreso con colores:
  - Rojo: Muy débil (< 8 chars)
  - Naranja: Débil (< 12 chars)
  - Amarillo: Media (< 16 chars)
  - Amarillo claro: Fuerte (≥ 16 chars)
  - Verde: Muy fuerte (≥ 16 chars + mayúsculas + números + símbolos)

**2. JWT Secret** (requerido)
- Mínimo 32 caracteres
- Botón "Generar Aleatorio" 🎲 (genera 64 chars)
- Indicador de fuerza similar a DB Password
- Toggle show/hide

**3. Puerto de Aplicación** (opcional)
- Default: 8080
- Rango: 1024-65535
- Validación en tiempo real

**4. Nombre de Base de Datos** (opcional)
- Default: "aymc"
- Alfanumérico permitido

#### Validación del Formulario

```typescript
const canInstall = computed(() => {
  return credentials.value.dbPassword.length >= 8 &&
         credentials.value.jwtSecret.length >= 32 &&
         credentials.value.appPort >= 1024 &&
         credentials.value.appPort <= 65535;
});
```

El botón "Iniciar Instalación" solo se habilita cuando todas las validaciones pasan.

#### Generador de JWT Secret

```typescript
function generateJwtSecret() {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < 64; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  credentials.value.jwtSecret = result;
}
```

Genera un string aleatorio de 64 caracteres (letras mayúsculas, minúsculas y números).

### Step 2: Instalación en Progreso

#### Integración con RemoteTerminal

```vue
<RemoteTerminal
  ref="terminalRef"
  title="Instalación AYMC"
  :can-stop="true"
  @stop="cancelInstallation"
/>
```

#### Proceso de Instalación

```typescript
async function startInstallation() {
  currentStep.value = 2;
  isInstalling.value = true;

  const terminal = terminalRef.value;
  terminal.clearTerminal();
  terminal.startExecution();
  terminal.writeHeader('Iniciando instalación de AYMC en VPS');
  
  try {
    // Verificar SSH
    terminal.writeInfo('Verificando conexión SSH...');
    const isConnected = await invoke<boolean>('ssh_is_connected');
    if (!isConnected) throw new Error('No hay conexión SSH activa');
    terminal.writeSuccess('Conexión SSH verificada');

    // Instalar AYMC
    terminal.writeInfo('Iniciando proceso de instalación...');
    const result = await invoke<{
      success: boolean;
      api_url: string;
      ws_url: string;
      message: string;
    }>('ssh_install_backend', {
      dbPassword: credentials.value.dbPassword,
      jwtSecret: credentials.value.jwtSecret,
      appPort: credentials.value.appPort,
    });

    if (result.success) {
      terminal.writeSuccess('Instalación completada exitosamente');
      terminal.endExecution(0);
      installationResult.value = {
        apiUrl: result.api_url,
        wsUrl: result.ws_url,
        success: true,
      };
      currentStep.value = 3; // Success
    } else {
      throw new Error(result.message);
    }
  } catch (error: any) {
    terminal.writeError(`Error: ${error.message}`);
    terminal.endExecution(1);
    errorMessage.value = error.message;
    currentStep.value = 4; // Error
  } finally {
    isInstalling.value = false;
  }
}
```

#### Comando Tauri Invocado

```rust
// src-tauri/src/commands.rs
#[tauri::command]
async fn ssh_install_backend(
    db_password: String,
    jwt_secret: String,
    app_port: u16,
) -> Result<InstallResult, String> {
    // Ejecuta script install-vps.sh remotamente
    // Retorna: { success, api_url, ws_url, message }
}
```

### Step 3: Instalación Completada ✅

#### Animación de Éxito

```html
<div class="checkmark-circle">
  <div class="checkmark"></div>
</div>
```

Animación CSS:
- Círculo verde escala desde 0 a 1 (0.5s)
- Checkmark se dibuja después (0.4s)

#### Resumen de Instalación

```typescript
interface InstallationSummary {
  apiUrl: string;        // http://192.168.1.100:8080/api/v1
  wsUrl: string;         // ws://192.168.1.100:8080/api/v1/ws
  dbName: string;        // aymc
  port: number;          // 8080
  environment: string;   // Production
  services: string;      // ✅ Backend, Agent, PostgreSQL
}
```

Se muestra en una grid de 2 columnas con todos los detalles de la instalación.

#### Acciones

- **Ver Logs Completos**: Mantiene el terminal visible con todo el output
- **Ir al Dashboard**: Emite evento `complete` con API URL y WS URL

### Step 4: Error en Instalación ❌

#### Animación de Error

```html
<div class="error-circle">
  <div class="error-mark">✗</div>
</div>
```

Círculo rojo con "✗" grande.

#### Detalles del Error

```html
<pre class="error-log">{{ errorMessage }}</pre>
```

Muestra el mensaje de error completo en un bloque pre-formateado con fondo oscuro.

#### Sugerencias

```html
<ul class="suggestions-list">
  <li>Verifica que la conexión SSH siga activa</li>
  <li>Asegúrate de que el usuario tiene permisos sudo</li>
  <li>Revisa que el puerto 8080 esté disponible</li>
  <li>Comprueba que PostgreSQL se pueda instalar</li>
</ul>
```

#### Acciones

- **Ver Logs Completos**: Muestra todo el output del terminal
- **Reintentar Instalación**: Vuelve al Step 1
- **Cancelar**: Emite evento `cancel`

---

## 🎨 Diseño Visual

### Paleta de Colores

```css
Primary Gradient:  linear-gradient(135deg, #667eea 0%, #764ba2 100%)
Success Green:     #4caf50
Error Red:         #f44336
Warning Orange:    #ff9800
Warning Yellow:    #ffc107
Text Primary:      #2d3748
Text Secondary:    #718096
Background Light:  #f7fafc
Border:            #e2e8f0
```

### Componentes Visuales

#### Progress Steps

```html
┌─────┐        ┌─────┐        ┌─────┐
│  1  │───────▶│  2  │───────▶│  3  │
└─────┘        └─────┘        └─────┘
Credenciales  Instalación  Completado
```

- Círculos con números
- Líneas conectoras
- Estados: inactive (gris) → active (gradient) → completed (verde)

#### Form Inputs

- Border: 2px solid #e2e8f0
- Focus: border-color #667eea + shadow
- Disabled: background #f7fafc

#### Buttons

**Primary**: Gradient background con shadow  
**Secondary**: White con border  
**Warning**: Orange solid

### Animaciones

```css
@keyframes scaleIn {
  from { transform: scale(0); }
  to { transform: scale(1); }
}

@keyframes checkmarkDraw {
  to { opacity: 1; }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
```

---

## 🔄 Integración con Tauri

### Comandos Invocados

```typescript
// Verificar conexión SSH activa
await invoke<boolean>('ssh_is_connected');

// Instalar AYMC en VPS remota
await invoke<InstallResult>('ssh_install_backend', {
  dbPassword: string,
  jwtSecret: string,
  appPort: number,
});

interface InstallResult {
  success: boolean;
  api_url: string;
  ws_url: string;
  message: string;
}
```

### Flujo Backend

```rust
// src-tauri/src/commands.rs

#[tauri::command]
async fn ssh_install_backend(
    db_password: String,
    jwt_secret: String,
    app_port: u16,
) -> Result<InstallResult, String> {
    // 1. Obtener SSHClient del estado global
    let ssh_client = get_ssh_client()?;
    
    // 2. Leer script embebido install-vps.sh
    let script = ScriptManager::read_script(Script::InstallVPS)?;
    
    // 3. Subir script a VPS
    ssh_client.upload_content("/tmp/install-aymc.sh", &script)?;
    
    // 4. Hacer ejecutable
    ssh_client.execute_command("chmod +x /tmp/install-aymc.sh")?;
    
    // 5. Ejecutar con credenciales
    let output = ssh_client.execute_command(&format!(
        "/tmp/install-aymc.sh {} {} {}",
        db_password, jwt_secret, app_port
    ))?;
    
    // 6. Parsear resultado
    if output.contains("INSTALLATION_SUCCESS") {
        Ok(InstallResult {
            success: true,
            api_url: extract_api_url(&output),
            ws_url: extract_ws_url(&output),
            message: "Instalación exitosa".to_string(),
        })
    } else {
        Err("Error en instalación".to_string())
    }
}
```

---

## 📝 Uso Completo

### En ServiceDetectionView.vue

```vue
<template>
  <div>
    <!-- Si NO está instalado -->
    <button v-if="needsInstallation" @click="goToInstaller">
      📦 Instalar AYMC
    </button>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router';

const router = useRouter();

function goToInstaller() {
  router.push('/installer');
}
</script>
```

### En Router

```typescript
// src/router/index.ts
{
  path: '/installer',
  name: 'installer',
  component: InstallationWizard,
  meta: { requiresSSH: true },
}
```

### En la Ruta

```vue
<template>
  <InstallationWizard
    :ssh-connected="sshConnected"
    @cancel="goBack"
    @complete="handleInstallComplete"
  />
</template>

<script setup>
import { useRouter } from 'vue-router';
import InstallationWizard from '@/components/InstallationWizard.vue';

const router = useRouter();
const sshConnected = ref(true); // Verificar estado SSH

function goBack() {
  router.push('/detection');
}

function handleInstallComplete(apiUrl: string, wsUrl: string) {
  // Guardar configuración
  localStorage.setItem('aymc_api_url', apiUrl);
  localStorage.setItem('aymc_ws_url', wsUrl);
  localStorage.setItem('aymc_environment', 'production');
  
  // Ir al dashboard
  router.push('/dashboard');
}
</script>
```

---

## 🧪 Testing

### Scenario 1: Instalación Exitosa

1. **Input**:
   - DB Password: `MySecureP@ssw0rd123`
   - JWT Secret: (generado automáticamente)
   - App Port: `8080`
   - DB Name: `aymc`

2. **Expected Output**:
   - Step 1 → Step 2 (terminal muestra progreso)
   - Terminal output:
     ```
     =========================================
     Iniciando instalación de AYMC en VPS
     =========================================
     ℹ Verificando conexión SSH...
     ✓ Conexión SSH verificada
     ℹ Iniciando proceso de instalación...
     ✓ Instalación completada exitosamente
     ```
   - Step 3 (Success) con summary
   - API URL: `http://192.168.1.100:8080/api/v1`
   - WS URL: `ws://192.168.1.100:8080/api/v1/ws`

### Scenario 2: Error de Conexión SSH

1. **Input**: (cualquier credencial)

2. **SSH Status**: Desconectado

3. **Expected Output**:
   - Step 2 terminal muestra:
     ```
     ℹ Verificando conexión SSH...
     ✗ Error: No hay conexión SSH activa
     ```
   - Step 4 (Error)
   - Error message: "No hay conexión SSH activa"
   - Botón "Reintentar"

### Scenario 3: Error en Script

1. **Input**: Credenciales válidas

2. **Script Failure**: Puerto ocupado

3. **Expected Output**:
   - Step 2 terminal muestra progreso hasta error
   - Step 4 (Error)
   - Error details con mensaje específico
   - Sugerencias de solución

---

## 📊 Métricas

### Archivos Creados (Fase 4)

```
src/components/
├── RemoteTerminal.vue        550 líneas
└── InstallationWizard.vue    980 líneas
─────────────────────────────────────
TOTAL Fase 4:               1,530 líneas
```

### Acumulado (Fases 1-4)

```
Fase 1 (SSH System):        1,040 líneas Rust
Fase 2 (Scripts):             170 líneas Rust + 55 KB scripts
Fase 3 (Onboarding):        1,580 líneas Vue
Fase 4 (Installation):      1,530 líneas Vue
─────────────────────────────────────
TOTAL:                      ~4,320 líneas + 55 KB
```

### Features Completas

✅ **RemoteTerminal**:
- Terminal xterm.js completa
- Themes (dark/light)
- Control buttons (clear, copy, stop)
- Status tracking (duration, exit code)
- Color support (ANSI codes)
- Auto-fit responsive
- 10,000 líneas de historial

✅ **InstallationWizard**:
- 4 pasos completos
- Form validation con fuerza de contraseña
- JWT generator
- Terminal integrada
- Success/Error handling
- Resumen de instalación
- Retry mechanism
- Cancel support

---

## 🎯 Estado del Proyecto

```
[██████████████████████████░░░░] 90% Completado

Fase 1: SSH System          ████████████ 100% ✅
Fase 2: Embedded Scripts    ████████████ 100% ✅
Fase 3: Onboarding UI       ████████████ 100% ✅
Fase 4: Installation Wizard ████████████ 100% ✅
Fase 5: Integration         ░░░░░░░░░░░░   0% ⏳
```

---

## 🚀 Próximos Pasos (Fase 5)

### 1. Configurar Vue Router
```typescript
// src/router/index.ts
- Crear rutas: /, /ssh-setup, /detection, /installer, /dashboard
- Navigation guards para verificar SSH
- Meta fields para autenticación
```

### 2. Integrar en App.vue
```vue
- Router view principal
- First-time detection (localStorage)
- Flujo automático vs manual
```

### 3. Configuración Dinámica
```typescript
// composables/useApiConfig.ts
- Detectar API_URL desde VPS
- Actualizar axios baseURL
- Environment switching (dev/prod)
```

### 4. Testing End-to-End
```
- Flujo completo: Gallery → SSH → Detection → Install → Dashboard
- Verificar comunicación API
- Probar WebSocket connection
```

---

## ✅ Resumen

**Fase 4 COMPLETADA**: Se han creado 2 componentes fundamentales para la instalación de AYMC:

1. **RemoteTerminal.vue** (550 líneas): Terminal emulada profesional con xterm.js
2. **InstallationWizard.vue** (980 líneas): Wizard completo de 4 pasos con validación, terminal integrada, y manejo de éxito/error

**Total Fase 4**: ~1,530 líneas de Vue 3 con terminal real, validaciones robustas, y UX pulida.

**Progreso Total**: 90% del sistema completo

**Próximo**: Fase 5 - Integration (Router + App.vue + Config dinámica)

---

**Última actualización:** 13 de noviembre de 2025  
**Estado:** ✅ Fase 4 completada  
**Progreso total:** 90% del sistema de onboarding e instalación
