<template>
  <div class="ssh-connection-container">
    <div class="connection-card">
      <div class="card-header">
        <div class="header-icon">🔐</div>
        <h2>Conectar a tu VPS</h2>
        <p class="subtitle">
          Ingresa las credenciales de tu servidor para comenzar
        </p>
      </div>

      <form @submit.prevent="handleConnect" class="connection-form">
        <!-- Host y Puerto -->
        <div class="form-row">
          <div class="form-group flex-2">
            <label for="host">
              <span class="label-icon">🌐</span>
              Host (IP o Dominio)
            </label>
            <input
              id="host"
              v-model="config.host"
              type="text"
              placeholder="192.168.1.100 o tu-dominio.com"
              required
              :disabled="connecting"
            />
          </div>

          <div class="form-group flex-1">
            <label for="port">
              <span class="label-icon">🔌</span>
              Puerto
            </label>
            <input
              id="port"
              v-model.number="config.port"
              type="number"
              placeholder="22"
              required
              :disabled="connecting"
            />
          </div>
        </div>

        <!-- Usuario -->
        <div class="form-group">
          <label for="username">
            <span class="label-icon">👤</span>
            Usuario SSH
          </label>
          <input
            id="username"
            v-model="config.username"
            type="text"
            placeholder="root"
            required
            :disabled="connecting"
          />
          <small class="form-hint">
            Usuario con permisos sudo para instalar AYMC
          </small>
        </div>

        <!-- Método de Autenticación -->
        <div class="form-group">
          <label>
            <span class="label-icon">🔑</span>
            Método de Autenticación
          </label>
          <div class="auth-type-selector">
            <button
              type="button"
              :class="['auth-button', { active: config.authType === 'password' }]"
              @click="config.authType = 'password'"
              :disabled="connecting"
            >
              <span class="auth-icon">🔒</span>
              <span>Contraseña</span>
            </button>
            <button
              type="button"
              :class="['auth-button', { active: config.authType === 'private_key_file' }]"
              @click="config.authType = 'private_key_file'"
              :disabled="connecting"
            >
              <span class="auth-icon">🗝️</span>
              <span>Clave Privada</span>
            </button>
          </div>
        </div>

        <!-- Autenticación con Contraseña -->
        <div v-if="config.authType === 'password'" class="form-group">
          <label for="password">
            <span class="label-icon">🔐</span>
            Contraseña
          </label>
          <div class="password-input">
            <input
              id="password"
              v-model="config.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="Tu contraseña SSH"
              required
              :disabled="connecting"
            />
            <button
              type="button"
              class="toggle-password"
              @click="showPassword = !showPassword"
              :disabled="connecting"
            >
              {{ showPassword ? '👁️' : '👁️‍🗨️' }}
            </button>
          </div>
        </div>

        <!-- Autenticación con Clave Privada -->
        <div v-if="config.authType === 'private_key_file'" class="form-group">
          <label for="privateKey">
            <span class="label-icon">📄</span>
            Ruta de la Clave Privada
          </label>
          <input
            id="privateKey"
            v-model="config.privateKeyPath"
            type="text"
            placeholder="/home/user/.ssh/id_rsa"
            required
            :disabled="connecting"
          />
          <small class="form-hint">
            Ruta completa a tu archivo de clave privada SSH
          </small>

          <label for="passphrase" style="margin-top: 15px">
            <span class="label-icon">🔓</span>
            Passphrase (Opcional)
          </label>
          <input
            id="passphrase"
            v-model="config.passphrase"
            type="password"
            placeholder="Si tu clave tiene passphrase"
            :disabled="connecting"
          />
        </div>

        <!-- Error Message -->
        <div v-if="error" class="error-message">
          <span class="error-icon">⚠️</span>
          <div>
            <strong>Error de Conexión</strong>
            <p>{{ error }}</p>
          </div>
        </div>

        <!-- Success Message -->
        <div v-if="isConnected" class="success-message">
          <span class="success-icon">✅</span>
          <div>
            <strong>Conectado Exitosamente</strong>
            <p>{{ connectionStatus }}</p>
          </div>
        </div>

        <!-- Connection Test -->
        <div v-if="testing" class="testing-message">
          <div class="spinner"></div>
          <span>Probando conexión...</span>
        </div>

        <!-- Buttons -->
        <div class="button-group">
          <button
            v-if="!isConnected"
            type="button"
            class="btn-test"
            @click="testConnection"
            :disabled="connecting || !isFormValid"
          >
            🔍 Probar Conexión
          </button>

          <button
            type="submit"
            class="btn-connect"
            :disabled="connecting || !isFormValid"
          >
            <span v-if="!connecting">🚀 Conectar y Continuar</span>
            <span v-else class="connecting-text">
              <div class="spinner-small"></div>
              Conectando...
            </span>
          </button>
        </div>

        <!-- Saved Connections -->
        <div v-if="savedConnections.length > 0" class="saved-connections">
          <h4>Conexiones Guardadas</h4>
          <div class="connection-list">
            <button
              v-for="conn in savedConnections"
              :key="conn.id"
              type="button"
              class="saved-connection-item"
              @click="loadConnection(conn)"
              :disabled="connecting"
            >
              <span class="conn-icon">💾</span>
              <div class="conn-info">
                <strong>{{ conn.username }}@{{ conn.host }}</strong>
                <small>Puerto {{ conn.port }}</small>
              </div>
            </button>
          </div>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue';
import { invoke } from '@tauri-apps/api/core';

interface SSHConnectionConfig {
  host: string;
  port: number;
  username: string;
  authType: 'password' | 'private_key_file';
  password?: string;
  privateKeyPath?: string;
  passphrase?: string;
}

interface SavedConnection {
  id: string;
  host: string;
  port: number;
  username: string;
  authType: 'password' | 'private_key_file';
}

const emit = defineEmits<{
  connected: [];
}>();

const config = reactive<SSHConnectionConfig>({
  host: '',
  port: 22,
  username: 'root',
  authType: 'password',
  password: '',
  privateKeyPath: '',
  passphrase: '',
});

const connecting = ref(false);
const testing = ref(false);
const isConnected = ref(false);
const showPassword = ref(false);
const connectionStatus = ref('');
const error = ref('');
const savedConnections = ref<SavedConnection[]>([]);

const isFormValid = computed(() => {
  const hasBasicInfo = config.host && config.port && config.username;
  const hasAuth =
    config.authType === 'password'
      ? !!config.password
      : !!config.privateKeyPath;
  return hasBasicInfo && hasAuth;
});

async function testConnection() {
  testing.value = true;
  error.value = '';

  try {
    const response = await invoke<{ success: boolean; data?: string; error?: string }>(
      'ssh_execute_command',
      { command: 'echo "Conexión OK"' }
    );

    if (response.success) {
      connectionStatus.value = 'Conexión establecida correctamente';
      isConnected.value = true;
    } else {
      error.value = response.error || 'Error al probar la conexión';
    }
  } catch (e) {
    error.value = String(e);
  } finally {
    testing.value = false;
  }
}

async function handleConnect() {
  connecting.value = true;
  error.value = '';

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
        privateKeyData: undefined,
        passphrase: config.passphrase,
      }
    );

    if (response.success) {
      connectionStatus.value = response.data || 'Conectado';
      isConnected.value = true;

      // Guardar conexión (sin contraseña)
      saveConnection();

      // Emitir evento de conexión exitosa
      setTimeout(() => {
        emit('connected');
      }, 1000);
    } else {
      error.value = response.error || 'Error desconocido al conectar';
    }
  } catch (e) {
    error.value = String(e);
  } finally {
    connecting.value = false;
  }
}

function saveConnection() {
  const savedConn: SavedConnection = {
    id: `${config.host}-${config.username}-${Date.now()}`,
    host: config.host,
    port: config.port,
    username: config.username,
    authType: config.authType,
  };

  // Guardar en localStorage
  const saved = localStorage.getItem('aymc_ssh_connections');
  const connections = saved ? JSON.parse(saved) : [];
  connections.push(savedConn);
  localStorage.setItem('aymc_ssh_connections', JSON.stringify(connections));

  savedConnections.value = connections;
}

function loadConnection(conn: SavedConnection) {
  config.host = conn.host;
  config.port = conn.port;
  config.username = conn.username;
  config.authType = conn.authType;
  error.value = '';
}

// Cargar conexiones guardadas al montar
const saved = localStorage.getItem('aymc_ssh_connections');
if (saved) {
  savedConnections.value = JSON.parse(saved);
}
</script>

<style scoped>
.ssh-connection-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.connection-card {
  background: white;
  border-radius: 20px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  max-width: 600px;
  width: 100%;
  overflow: hidden;
  animation: slideIn 0.5s ease;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.card-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 40px;
  text-align: center;
}

.header-icon {
  font-size: 60px;
  margin-bottom: 20px;
}

.card-header h2 {
  font-size: 28px;
  margin-bottom: 10px;
}

.subtitle {
  opacity: 0.9;
  font-size: 16px;
}

.connection-form {
  padding: 40px;
}

.form-row {
  display: flex;
  gap: 15px;
}

.flex-1 {
  flex: 1;
}

.flex-2 {
  flex: 2;
}

.form-group {
  margin-bottom: 20px;
}

label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-weight: 600;
  color: #333;
}

.label-icon {
  font-size: 18px;
}

input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 15px;
  transition: all 0.3s ease;
}

input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

input:disabled {
  background: #f5f5f5;
  cursor: not-allowed;
}

.form-hint {
  display: block;
  margin-top: 6px;
  font-size: 13px;
  color: #666;
}

.auth-type-selector {
  display: flex;
  gap: 10px;
}

.auth-button {
  flex: 1;
  padding: 16px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  background: white;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.auth-button:hover:not(:disabled) {
  border-color: #667eea;
  transform: translateY(-2px);
}

.auth-button.active {
  border-color: #667eea;
  background: rgba(102, 126, 234, 0.1);
}

.auth-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.auth-icon {
  font-size: 24px;
}

.password-input {
  position: relative;
  display: flex;
}

.toggle-password {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  cursor: pointer;
  font-size: 20px;
}

.error-message,
.success-message {
  display: flex;
  gap: 12px;
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.error-message {
  background: #ffebee;
  border-left: 4px solid #f44336;
}

.success-message {
  background: #e8f5e9;
  border-left: 4px solid #4caf50;
}

.error-icon {
  font-size: 24px;
}

.success-icon {
  font-size: 24px;
}

.testing-message {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: #e3f2fd;
  border-radius: 8px;
  margin-bottom: 20px;
}

.button-group {
  display: flex;
  gap: 12px;
}

button[type='submit'],
.btn-test,
.btn-connect {
  padding: 14px 28px;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-test {
  flex: 1;
  background: #f5f5f5;
  color: #333;
}

.btn-test:hover:not(:disabled) {
  background: #e0e0e0;
}

.btn-connect {
  flex: 2;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-connect:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.connecting-text {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.spinner,
.spinner-small {
  border: 3px solid rgba(102, 126, 234, 0.3);
  border-top: 3px solid #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.spinner {
  width: 24px;
  height: 24px;
}

.spinner-small {
  width: 16px;
  height: 16px;
  border-width: 2px;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

.saved-connections {
  margin-top: 30px;
  padding-top: 30px;
  border-top: 1px solid #e0e0e0;
}

.saved-connections h4 {
  margin-bottom: 15px;
  color: #666;
  font-size: 14px;
  text-transform: uppercase;
}

.connection-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.saved-connection-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f8f9fa;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: left;
}

.saved-connection-item:hover:not(:disabled) {
  background: #e8f5e9;
  border-color: #4caf50;
}

.conn-icon {
  font-size: 20px;
}

.conn-info {
  flex: 1;
}

.conn-info strong {
  display: block;
  color: #333;
}

.conn-info small {
  color: #666;
}
</style>
