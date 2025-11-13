<template>
  <div class="detection-container">
    <div class="detection-card">
      <div class="card-header">
        <div class="header-icon">🔍</div>
        <h2>Detectando Servicios AYMC</h2>
        <p class="subtitle">
          Verificando si AYMC ya está instalado en tu VPS...
        </p>
      </div>

      <div class="detection-content">
        <!-- Scanning Animation -->
        <div v-if="scanning" class="scanning-state">
          <div class="scan-animation">
            <div class="scan-line"></div>
            <div class="scan-circle"></div>
          </div>
          <p class="scan-text">Analizando servidor...</p>
        </div>

        <!-- Results -->
        <div v-if="!scanning && status" class="results">
          <!-- Backend Status -->
          <div class="service-item">
            <div class="service-icon">
              {{ status.backend_installed ? '✅' : '❌' }}
            </div>
            <div class="service-info">
              <h3>Backend API</h3>
              <p v-if="status.backend_installed">
                <span class="status-badge success">Instalado</span>
                <span v-if="status.backend_running" class="status-badge running">
                  Corriendo
                </span>
                <span v-else class="status-badge stopped">Detenido</span>
              </p>
              <p v-else class="status-text">No instalado</p>
              <small v-if="status.backend_path">{{ status.backend_path }}</small>
            </div>
          </div>

          <!-- Agent Status -->
          <div class="service-item">
            <div class="service-icon">
              {{ status.agent_installed ? '✅' : '❌' }}
            </div>
            <div class="service-info">
              <h3>Agent gRPC</h3>
              <p v-if="status.agent_installed">
                <span class="status-badge success">Instalado</span>
                <span v-if="status.agent_running" class="status-badge running">
                  Corriendo
                </span>
                <span v-else class="status-badge stopped">Detenido</span>
              </p>
              <p v-else class="status-text">No instalado</p>
              <small v-if="status.agent_path">{{ status.agent_path }}</small>
            </div>
          </div>

          <!-- PostgreSQL Status -->
          <div class="service-item">
            <div class="service-icon">
              {{ status.postgresql_running ? '✅' : '❌' }}
            </div>
            <div class="service-info">
              <h3>PostgreSQL</h3>
              <p v-if="status.postgresql_running">
                <span class="status-badge running">Corriendo</span>
              </p>
              <p v-else class="status-text">No detectado</p>
            </div>
          </div>

          <!-- Backend Config (si existe) -->
          <div v-if="backendConfig" class="config-section">
            <h3>📋 Configuración Detectada</h3>
            <div class="config-item">
              <strong>API URL:</strong>
              <code>{{ backendConfig.api_url }}</code>
            </div>
            <div class="config-item">
              <strong>WebSocket URL:</strong>
              <code>{{ backendConfig.ws_url }}</code>
            </div>
            <div class="config-item">
              <strong>Environment:</strong>
              <span :class="['env-badge', backendConfig.environment]">
                {{ backendConfig.environment }}
              </span>
            </div>
          </div>
        </div>

        <!-- Error -->
        <div v-if="error" class="error-state">
          <span class="error-icon">⚠️</span>
          <div>
            <strong>Error al Detectar Servicios</strong>
            <p>{{ error }}</p>
          </div>
        </div>

        <!-- Actions -->
        <div v-if="!scanning && status" class="actions">
          <button
            v-if="needsInstallation"
            class="btn-primary"
            @click="$emit('install')"
          >
            📦 Instalar AYMC
          </button>

          <button
            v-else-if="allServicesRunning"
            class="btn-success"
            @click="$emit('continue')"
          >
            ✅ Continuar al Dashboard
          </button>

          <button
            v-else
            class="btn-warning"
            @click="$emit('restart-services')"
          >
            🔄 Reiniciar Servicios
          </button>

          <button class="btn-secondary" @click="rescan">
            🔍 Volver a Escanear
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { invoke } from '@tauri-apps/api/core';

interface ServiceStatus {
  backend_installed: boolean;
  agent_installed: boolean;
  backend_running: boolean;
  agent_running: boolean;
  postgresql_running: boolean;
  backend_path?: string;
  agent_path?: string;
}

interface BackendConfig {
  api_url: string;
  ws_url: string;
  environment: string;
  port: string;
}

// @ts-ignore - emit is used via $emit in template
const emit = defineEmits<{
  install: [];
  continue: [];
  'restart-services': [];
}>();

// This keeps emit from being "unused" in TypeScript
void emit;

const scanning = ref(true);
const status = ref<ServiceStatus | null>(null);
const backendConfig = ref<BackendConfig | null>(null);
const error = ref('');

const needsInstallation = computed(() => {
  return status.value && (!status.value.backend_installed || !status.value.agent_installed);
});

const allServicesRunning = computed(() => {
  return (
    status.value &&
    status.value.backend_installed &&
    status.value.agent_installed &&
    status.value.backend_running &&
    status.value.agent_running &&
    status.value.postgresql_running
  );
});

async function detectServices() {
  scanning.value = true;
  error.value = '';
  status.value = null;
  backendConfig.value = null;

  try {
    // Detectar servicios
    const serviceResponse = await invoke<{
      success: boolean;
      data?: ServiceStatus;
      error?: string;
    }>('ssh_check_services');

    if (serviceResponse.success && serviceResponse.data) {
      status.value = serviceResponse.data;

      // Si el backend está instalado, obtener configuración
      if (serviceResponse.data.backend_installed) {
        const configResponse = await invoke<{
          success: boolean;
          data?: BackendConfig;
          error?: string;
        }>('ssh_get_backend_config');

        if (configResponse.success && configResponse.data) {
          backendConfig.value = configResponse.data;
        }
      }
    } else {
      error.value = serviceResponse.error || 'Error al detectar servicios';
    }
  } catch (e) {
    error.value = String(e);
  } finally {
    scanning.value = false;
  }
}

function rescan() {
  detectServices();
}

onMounted(() => {
  detectServices();
});
</script>

<style scoped>
.detection-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.detection-card {
  background: white;
  border-radius: 20px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  max-width: 700px;
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
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
}

.card-header h2 {
  font-size: 28px;
  margin-bottom: 10px;
}

.subtitle {
  opacity: 0.9;
  font-size: 16px;
}

.detection-content {
  padding: 40px;
}

.scanning-state {
  text-align: center;
  padding: 60px 20px;
}

.scan-animation {
  position: relative;
  width: 150px;
  height: 150px;
  margin: 0 auto 30px;
}

.scan-circle {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 100px;
  height: 100px;
  border: 4px solid #667eea;
  border-radius: 50%;
  animation: pulse-circle 2s infinite;
}

@keyframes pulse-circle {
  0% {
    transform: translate(-50%, -50%) scale(1);
    opacity: 1;
  }
  100% {
    transform: translate(-50%, -50%) scale(1.5);
    opacity: 0;
  }
}

.scan-line {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 80px;
  height: 4px;
  background: linear-gradient(90deg, #667eea, #764ba2);
  border-radius: 2px;
  transform-origin: 0 50%;
  animation: rotate-line 2s linear infinite;
}

@keyframes rotate-line {
  from {
    transform: translate(0, -50%) rotate(0deg);
  }
  to {
    transform: translate(0, -50%) rotate(360deg);
  }
}

.scan-text {
  font-size: 18px;
  color: #666;
  animation: pulse-text 1.5s infinite;
}

@keyframes pulse-text {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.results {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.service-item {
  display: flex;
  gap: 20px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 12px;
  border-left: 4px solid #e0e0e0;
}

.service-icon {
  font-size: 32px;
}

.service-info {
  flex: 1;
}

.service-info h3 {
  margin: 0 0 8px 0;
  font-size: 18px;
  color: #333;
}

.status-text {
  color: #666;
  margin: 0;
}

.status-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 600;
  margin-right: 8px;
}

.status-badge.success {
  background: #e8f5e9;
  color: #2e7d32;
}

.status-badge.running {
  background: #e3f2fd;
  color: #1565c0;
}

.status-badge.stopped {
  background: #fff3e0;
  color: #e65100;
}

.service-info small {
  display: block;
  margin-top: 6px;
  color: #999;
  font-size: 12px;
}

.config-section {
  padding: 20px;
  background: #f0f4ff;
  border-radius: 12px;
  border-left: 4px solid #667eea;
}

.config-section h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
  color: #333;
}

.config-item {
  margin-bottom: 12px;
}

.config-item strong {
  display: block;
  margin-bottom: 4px;
  color: #666;
  font-size: 13px;
}

.config-item code {
  display: block;
  padding: 8px 12px;
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  color: #333;
}

.env-badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 600;
}

.env-badge.production {
  background: #ffebee;
  color: #c62828;
}

.env-badge.development {
  background: #e8f5e9;
  color: #2e7d32;
}

.error-state {
  display: flex;
  gap: 12px;
  padding: 20px;
  background: #ffebee;
  border-left: 4px solid #f44336;
  border-radius: 12px;
}

.error-icon {
  font-size: 24px;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 30px;
}

button {
  padding: 14px 28px;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  flex: 1;
  min-width: 200px;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-success {
  background: linear-gradient(135deg, #4caf50 0%, #45a049 100%);
  color: white;
}

.btn-success:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(76, 175, 80, 0.4);
}

.btn-warning {
  background: linear-gradient(135deg, #ff9800 0%, #f57c00 100%);
  color: white;
}

.btn-warning:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(255, 152, 0, 0.4);
}

.btn-secondary {
  background: #f5f5f5;
  color: #333;
}

.btn-secondary:hover {
  background: #e0e0e0;
}
</style>
