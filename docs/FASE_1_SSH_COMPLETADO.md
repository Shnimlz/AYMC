# 🔐 Fase 1: Sistema SSH Completado

## ✅ Implementación Completada

### 1. Dependencias Rust Agregadas

**Archivo:** `SeraMC/src-tauri/Cargo.toml`

```toml
# SSH Connection
ssh2 = "0.9"
# Async runtime for SSH operations
tokio = { version = "1", features = ["full"] }
# Base64 encoding for private keys
base64 = "0.21"
# Error handling
anyhow = "1.0"
thiserror = "1.0"
```

---

### 2. Módulo SSH Core (`src-tauri/src/ssh.rs`)

Implementa toda la lógica de conexión SSH:

#### Estructuras principales:

- **`SSHAuth`**: Métodos de autenticación
  - `Password`: Usuario + contraseña
  - `PrivateKey`: Usuario + archivo de clave privada + passphrase opcional
  - `PrivateKeyData`: Usuario + contenido de clave + passphrase opcional

- **`SSHConfig`**: Configuración de conexión
  - `host`: IP o dominio
  - `port`: Puerto SSH (default: 22)
  - `username`: Usuario SSH
  - `auth`: Método de autenticación

- **`ServiceStatus`**: Estado de servicios AYMC
  - `backend_installed`: ¿Backend instalado?
  - `agent_installed`: ¿Agent instalado?
  - `backend_running`: ¿Backend corriendo?
  - `agent_running`: ¿Agent corriendo?
  - `postgresql_running`: ¿PostgreSQL corriendo?

- **`BackendConfig`**: Configuración del backend
  - `api_url`: URL de la API REST
  - `ws_url`: URL del WebSocket
  - `environment`: production/development
  - `port`: Puerto del backend

#### Funciones principales:

```rust
// Conexión
SSHClient::connect(config) -> Result<SSHClient>

// Ejecución de comandos
client.execute_command(command) -> Result<String>
client.execute_command_streaming(command) -> Result<Vec<String>>

// Sistema de archivos
client.file_exists(path) -> Result<bool>
client.read_file(path) -> Result<String>
client.upload_file(local, remote) -> Result<()>
client.upload_content(content, remote) -> Result<()>

// Servicios
client.check_services() -> Result<ServiceStatus>
client.is_service_running(service_name) -> Result<bool>
client.get_backend_config() -> Result<BackendConfig>

// Información del sistema
client.get_host_info() -> Result<String>
client.has_sudo_access() -> Result<bool>
```

---

### 3. Comandos Tauri (`src-tauri/src/commands.rs`)

Expone la funcionalidad SSH a Vue mediante comandos:

#### Comandos disponibles:

1. **`ssh_connect`**: Conecta al servidor SSH
2. **`ssh_disconnect`**: Cierra la conexión SSH
3. **`ssh_is_connected`**: Verifica si hay conexión activa
4. **`ssh_execute_command`**: Ejecuta un comando remoto
5. **`ssh_execute_streaming`**: Ejecuta comando con output en tiempo real
6. **`ssh_check_services`**: Verifica estado de servicios AYMC
7. **`ssh_get_backend_config`**: Obtiene configuración del backend
8. **`ssh_file_exists`**: Verifica si existe un archivo
9. **`ssh_read_file`**: Lee contenido de archivo remoto
10. **`ssh_upload_content`**: Sube contenido a archivo remoto
11. **`ssh_get_host_info`**: Información del SO remoto
12. **`ssh_has_sudo`**: Verifica permisos sudo

#### Formato de respuesta:

Todos los comandos retornan:

```typescript
interface CommandResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}
```

---

## 📝 Uso desde Vue/TypeScript

### 1. Crear un Composable SSH

**Archivo:** `SeraMC/src/composables/useSSH.ts`

```typescript
import { invoke } from '@tauri-apps/api/core';
import { ref } from 'vue';

export interface SSHConnectionConfig {
  host: string;
  port: number;
  username: string;
  authType: 'password' | 'private_key_file' | 'private_key_data';
  password?: string;
  privateKeyPath?: string;
  privateKeyData?: string;
  passphrase?: string;
}

export interface ServiceStatus {
  backend_installed: boolean;
  agent_installed: boolean;
  backend_running: boolean;
  agent_running: boolean;
  postgresql_running: boolean;
  backend_path?: string;
  agent_path?: string;
}

export interface BackendConfig {
  api_url: string;
  ws_url: string;
  environment: string;
  port: string;
}

export function useSSH() {
  const isConnected = ref(false);
  const connectionStatus = ref('');
  const error = ref('');

  // Conectar al servidor SSH
  async function connect(config: SSHConnectionConfig) {
    try {
      const response = await invoke<{ success: boolean; data?: string; error?: string }>(
        'ssh_connect',
        {
          host: config.host,
          port: config.port,
          username: config.username,
          authType: config.authType,
          password: config.password,
          privateKeyPath: config.privateKeyPath,
          privateKeyData: config.privateKeyData,
          passphrase: config.passphrase,
        }
      );

      if (response.success) {
        isConnected.value = true;
        connectionStatus.value = response.data || 'Conectado';
        error.value = '';
        return true;
      } else {
        error.value = response.error || 'Error desconocido';
        return false;
      }
    } catch (e) {
      error.value = String(e);
      return false;
    }
  }

  // Desconectar
  async function disconnect() {
    try {
      await invoke('ssh_disconnect');
      isConnected.value = false;
      connectionStatus.value = '';
    } catch (e) {
      error.value = String(e);
    }
  }

  // Verificar servicios AYMC
  async function checkServices(): Promise<ServiceStatus | null> {
    try {
      const response = await invoke<{ success: boolean; data?: ServiceStatus; error?: string }>(
        'ssh_check_services'
      );

      if (response.success && response.data) {
        return response.data;
      } else {
        error.value = response.error || 'Error al verificar servicios';
        return null;
      }
    } catch (e) {
      error.value = String(e);
      return null;
    }
  }

  // Obtener configuración del backend
  async function getBackendConfig(): Promise<BackendConfig | null> {
    try {
      const response = await invoke<{ success: boolean; data?: BackendConfig; error?: string }>(
        'ssh_get_backend_config'
      );

      if (response.success && response.data) {
        return response.data;
      } else {
        error.value = response.error || 'Error al obtener configuración';
        return null;
      }
    } catch (e) {
      error.value = String(e);
      return null;
    }
  }

  // Ejecutar comando remoto
  async function executeCommand(command: string): Promise<string | null> {
    try {
      const response = await invoke<{ success: boolean; data?: string; error?: string }>(
        'ssh_execute_command',
        { command }
      );

      if (response.success) {
        return response.data || '';
      } else {
        error.value = response.error || 'Error al ejecutar comando';
        return null;
      }
    } catch (e) {
      error.value = String(e);
      return null;
    }
  }

  // Ejecutar comando con streaming
  async function executeCommandStreaming(command: string): Promise<string[] | null> {
    try {
      const response = await invoke<{ success: boolean; data?: string[]; error?: string }>(
        'ssh_execute_streaming',
        { command }
      );

      if (response.success && response.data) {
        return response.data;
      } else {
        error.value = response.error || 'Error al ejecutar comando';
        return null;
      }
    } catch (e) {
      error.value = String(e);
      return null;
    }
  }

  // Leer archivo remoto
  async function readFile(path: string): Promise<string | null> {
    try {
      const response = await invoke<{ success: boolean; data?: string; error?: string }>(
        'ssh_read_file',
        { path }
      );

      if (response.success) {
        return response.data || '';
      } else {
        error.value = response.error || 'Error al leer archivo';
        return null;
      }
    } catch (e) {
      error.value = String(e);
      return null;
    }
  }

  // Subir contenido a archivo remoto
  async function uploadContent(content: string, remotePath: string): Promise<boolean> {
    try {
      const response = await invoke<{ success: boolean; error?: string }>(
        'ssh_upload_content',
        { content, remotePath }
      );

      if (response.success) {
        return true;
      } else {
        error.value = response.error || 'Error al subir archivo';
        return false;
      }
    } catch (e) {
      error.value = String(e);
      return false;
    }
  }

  return {
    // Estado
    isConnected,
    connectionStatus,
    error,
    // Funciones
    connect,
    disconnect,
    checkServices,
    getBackendConfig,
    executeCommand,
    executeCommandStreaming,
    readFile,
    uploadContent,
  };
}
```

---

### 2. Ejemplo de Uso en un Componente Vue

**Archivo:** `SeraMC/src/components/SSHConnectionForm.vue`

```vue
<template>
  <div class="ssh-connection-form">
    <h2>Conectar a VPS</h2>

    <!-- Formulario de conexión -->
    <form @submit.prevent="handleConnect">
      <div class="form-group">
        <label>Host (IP o dominio):</label>
        <input v-model="config.host" type="text" placeholder="192.168.1.100" required />
      </div>

      <div class="form-group">
        <label>Puerto:</label>
        <input v-model.number="config.port" type="number" placeholder="22" required />
      </div>

      <div class="form-group">
        <label>Usuario:</label>
        <input v-model="config.username" type="text" placeholder="root" required />
      </div>

      <div class="form-group">
        <label>Método de autenticación:</label>
        <select v-model="config.authType">
          <option value="password">Contraseña</option>
          <option value="private_key_file">Clave Privada (archivo)</option>
          <option value="private_key_data">Clave Privada (texto)</option>
        </select>
      </div>

      <!-- Password -->
      <div v-if="config.authType === 'password'" class="form-group">
        <label>Contraseña:</label>
        <input v-model="config.password" type="password" required />
      </div>

      <!-- Private Key File -->
      <div v-if="config.authType === 'private_key_file'" class="form-group">
        <label>Ruta de la clave privada:</label>
        <input v-model="config.privateKeyPath" type="text" placeholder="/home/user/.ssh/id_rsa" required />
        
        <label>Passphrase (opcional):</label>
        <input v-model="config.passphrase" type="password" />
      </div>

      <!-- Private Key Data -->
      <div v-if="config.authType === 'private_key_data'" class="form-group">
        <label>Contenido de la clave privada:</label>
        <textarea v-model="config.privateKeyData" rows="10" required></textarea>
        
        <label>Passphrase (opcional):</label>
        <input v-model="config.passphrase" type="password" />
      </div>

      <button type="submit" :disabled="loading">
        {{ loading ? 'Conectando...' : 'Conectar' }}
      </button>

      <!-- Estado -->
      <div v-if="ssh.error.value" class="error">
        {{ ssh.error.value }}
      </div>

      <div v-if="ssh.isConnected.value" class="success">
        ✅ {{ ssh.connectionStatus.value }}
      </div>
    </form>

    <!-- Verificación de servicios -->
    <div v-if="ssh.isConnected.value" class="services-check">
      <button @click="checkServices" :disabled="checkingServices">
        {{ checkingServices ? 'Verificando...' : 'Verificar Servicios' }}
      </button>

      <div v-if="serviceStatus" class="service-status">
        <h3>Estado de Servicios:</h3>
        <ul>
          <li>Backend instalado: {{ serviceStatus.backend_installed ? '✅' : '❌' }}</li>
          <li>Agent instalado: {{ serviceStatus.agent_installed ? '✅' : '❌' }}</li>
          <li>Backend corriendo: {{ serviceStatus.backend_running ? '✅' : '❌' }}</li>
          <li>Agent corriendo: {{ serviceStatus.agent_running ? '✅' : '❌' }}</li>
          <li>PostgreSQL corriendo: {{ serviceStatus.postgresql_running ? '✅' : '❌' }}</li>
        </ul>

        <div v-if="backendConfig" class="backend-config">
          <h3>Configuración del Backend:</h3>
          <p><strong>API URL:</strong> {{ backendConfig.api_url }}</p>
          <p><strong>WebSocket URL:</strong> {{ backendConfig.ws_url }}</p>
          <p><strong>Environment:</strong> {{ backendConfig.environment }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { useSSH, type ServiceStatus, type BackendConfig } from '@/composables/useSSH';

const ssh = useSSH();

const config = reactive({
  host: '',
  port: 22,
  username: 'root',
  authType: 'password' as 'password' | 'private_key_file' | 'private_key_data',
  password: '',
  privateKeyPath: '',
  privateKeyData: '',
  passphrase: '',
});

const loading = ref(false);
const checkingServices = ref(false);
const serviceStatus = ref<ServiceStatus | null>(null);
const backendConfig = ref<BackendConfig | null>(null);

async function handleConnect() {
  loading.value = true;
  const success = await ssh.connect(config);
  loading.value = false;

  if (success) {
    // Automáticamente verificar servicios después de conectar
    await checkServices();
  }
}

async function checkServices() {
  checkingServices.value = true;
  
  // Verificar estado de servicios
  serviceStatus.value = await ssh.checkServices();
  
  // Si el backend está instalado, obtener su configuración
  if (serviceStatus.value?.backend_installed) {
    backendConfig.value = await ssh.getBackendConfig();
  }
  
  checkingServices.value = false;
}
</script>

<style scoped>
.ssh-connection-form {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
}

.form-group {
  margin-bottom: 15px;
}

label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}

input, select, textarea {
  width: 100%;
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
}

button {
  padding: 10px 20px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-top: 10px;
}

button:disabled {
  background-color: #ccc;
  cursor: not-allowed;
}

.error {
  color: red;
  margin-top: 10px;
  padding: 10px;
  background-color: #ffebee;
  border-radius: 4px;
}

.success {
  color: green;
  margin-top: 10px;
  padding: 10px;
  background-color: #e8f5e9;
  border-radius: 4px;
}

.services-check {
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid #ccc;
}

.service-status ul {
  list-style: none;
  padding: 0;
}

.service-status li {
  padding: 5px 0;
}
</style>
```

---

## 🚀 Próximos Pasos

### Fase 2: Embeber Scripts de Instalación

1. Crear carpeta `SeraMC/src-tauri/resources/`
2. Copiar scripts desde `/scripts/`:
   - `install-vps.sh`
   - `continue-install.sh`
   - `uninstall.sh`
   - `build.sh`
3. Configurar `tauri.conf.json` para incluir recursos
4. Crear comando Tauri para acceder a los scripts

### Fase 3: Onboarding UI

1. Instalar Swiper.js: `npm install swiper`
2. Crear `OnboardingGallery.vue`
3. Crear flujo: Gallery → SSH Setup → Service Detection → Dashboard

### Fase 4: Terminal Emulada

1. Instalar xterm.js: `npm install xterm`
2. Crear `RemoteTerminal.vue`
3. Integrar con `ssh_execute_streaming`

### Fase 5: Installation Wizard

1. Crear `InstallationWizard.vue`
2. Formulario para credenciales (.env values)
3. Ejecutar scripts vía SSH con progreso en tiempo real

---

## 📦 Compilación

Para compilar el proyecto necesitas tener Rust instalado:

```bash
# Instalar Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Compilar Tauri
cd SeraMC
npm install
npm run tauri build
```

---

## ✅ Resumen de Archivos Creados/Modificados

### Creados:
- ✅ `SeraMC/src-tauri/src/ssh.rs` (432 líneas)
- ✅ `SeraMC/src-tauri/src/commands.rs` (272 líneas)
- ✅ Este documento de documentación

### Modificados:
- ✅ `SeraMC/src-tauri/Cargo.toml` (agregadas dependencias)
- ✅ `SeraMC/src-tauri/src/lib.rs` (registrados módulos y comandos)

### Total:
- **~700 líneas de código Rust**
- **12 comandos Tauri expuestos**
- **API completa para SSH desde Vue**

---

## 🎯 Estado Actual

**Fase 1: ✅ COMPLETADA**

- ✅ Dependencias configuradas
- ✅ Módulo SSH implementado
- ✅ Comandos Tauri creados
- ✅ Documentación de uso
- ⏳ Compilación pendiente (requiere Rust instalado)

**Próximo paso:** Instalar Rust y compilar, o continuar con Fase 2 (embeber scripts).
