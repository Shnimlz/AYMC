<template>
  <div class="installation-wizard">
    <!-- Header -->
    <div class="wizard-header">
      <div class="header-content">
        <h1 class="wizard-title">
          <span class="title-icon">🚀</span>
          Instalación de AYMC
        </h1>
        <p class="wizard-subtitle">
          Configuraremos AYMC en tu VPS de forma automática
        </p>
      </div>
      
      <!-- Pasos de progreso -->
      <div class="progress-steps">
        <div 
          v-for="step in steps" 
          :key="step.id"
          class="step"
          :class="{ 
            active: step.id === currentStep, 
            completed: step.id < currentStep 
          }"
        >
          <div class="step-number">
            <span v-if="step.id < currentStep">✓</span>
            <span v-else>{{ step.id }}</span>
          </div>
          <span class="step-label">{{ step.label }}</span>
        </div>
      </div>
    </div>

    <!-- Contenido principal -->
    <div class="wizard-content">
      <!-- Step 1: Configuración de credenciales -->
      <div v-show="currentStep === 1" class="step-content">
        <div class="credentials-form">
          <h2 class="form-title">
            <span class="icon">🔐</span>
            Configuración de Credenciales
          </h2>
          
          <div class="form-description">
            <p>
              AYMC necesita algunas credenciales para configurar la base de datos 
              y la seguridad de la aplicación.
            </p>
          </div>

          <div class="form-grid">
            <!-- Database Password -->
            <div class="form-group">
              <label for="dbPassword" class="form-label">
                <span class="label-icon">🗄️</span>
                Contraseña de PostgreSQL *
              </label>
              <div class="input-with-button">
                <input
                  id="dbPassword"
                  v-model="credentials.dbPassword"
                  :type="showDbPassword ? 'text' : 'password'"
                  class="form-input"
                  placeholder="Contraseña segura para PostgreSQL"
                  :disabled="isInstalling"
                />
                <button 
                  type="button"
                  @click="showDbPassword = !showDbPassword"
                  class="toggle-password-btn"
                  :disabled="isInstalling"
                >
                  {{ showDbPassword ? '👁️' : '👁️‍🗨️' }}
                </button>
              </div>
              <div class="password-strength">
                <div class="strength-bar" :class="`strength-${dbPasswordStrength}`">
                  <div class="strength-fill"></div>
                </div>
                <span class="strength-text">{{ dbPasswordStrengthText }}</span>
              </div>
            </div>

            <!-- JWT Secret -->
            <div class="form-group">
              <label for="jwtSecret" class="form-label">
                <span class="label-icon">🔑</span>
                JWT Secret *
              </label>
              <div class="input-with-button">
                <input
                  id="jwtSecret"
                  v-model="credentials.jwtSecret"
                  :type="showJwtSecret ? 'text' : 'password'"
                  class="form-input"
                  placeholder="Clave secreta para tokens JWT"
                  :disabled="isInstalling"
                />
                <button 
                  type="button"
                  @click="showJwtSecret = !showJwtSecret"
                  class="toggle-password-btn"
                  :disabled="isInstalling"
                >
                  {{ showJwtSecret ? '👁️' : '👁️‍🗨️' }}
                </button>
                <button 
                  type="button"
                  @click="generateJwtSecret"
                  class="generate-btn"
                  title="Generar JWT aleatorio"
                  :disabled="isInstalling"
                >
                  🎲
                </button>
              </div>
              <div class="password-strength">
                <div class="strength-bar" :class="`strength-${jwtSecretStrength}`">
                  <div class="strength-fill"></div>
                </div>
                <span class="strength-text">{{ jwtSecretStrengthText }}</span>
              </div>
            </div>

            <!-- App Port -->
            <div class="form-group">
              <label for="appPort" class="form-label">
                <span class="label-icon">🔌</span>
                Puerto de la Aplicación
              </label>
              <input
                id="appPort"
                v-model.number="credentials.appPort"
                type="number"
                class="form-input"
                placeholder="8080"
                min="1024"
                max="65535"
                :disabled="isInstalling"
              />
              <span class="form-hint">
                Puerto donde escuchará la API (por defecto: 8080)
              </span>
            </div>

            <!-- Database Name -->
            <div class="form-group">
              <label for="dbName" class="form-label">
                <span class="label-icon">💾</span>
                Nombre de la Base de Datos
              </label>
              <input
                id="dbName"
                v-model="credentials.dbName"
                type="text"
                class="form-input"
                placeholder="aymc"
                :disabled="isInstalling"
              />
              <span class="form-hint">
                Nombre de la base de datos PostgreSQL (por defecto: aymc)
              </span>
            </div>
          </div>

          <div class="form-actions">
            <button 
              @click="$emit('cancel')" 
              class="btn btn-secondary"
              :disabled="isInstalling"
            >
              <span class="btn-icon">❌</span>
              Cancelar
            </button>
            <button 
              @click="startInstallation" 
              class="btn btn-primary"
              :disabled="!canInstall || isInstalling"
            >
              <span class="btn-icon">🚀</span>
              {{ isInstalling ? 'Instalando...' : 'Iniciar Instalación' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Step 2: Instalación en progreso -->
      <div v-show="currentStep === 2" class="step-content">
        <div class="installation-progress">
          <h2 class="form-title">
            <span class="icon spinning">⚙️</span>
            Instalación en Progreso
          </h2>
          
          <div class="progress-info">
            <p>Por favor espera mientras instalamos AYMC en tu VPS...</p>
            <p class="progress-warning">⚠️ No cierres esta ventana</p>
          </div>

          <!-- Terminal integrada -->
          <RemoteTerminal
            ref="terminalRef"
            title="Instalación AYMC"
            :can-stop="true"
            @stop="cancelInstallation"
          />
        </div>
      </div>

      <!-- Step 3: Instalación completada -->
      <div v-show="currentStep === 3" class="step-content">
        <div class="installation-complete">
          <div class="success-animation">
            <div class="checkmark-circle">
              <div class="checkmark"></div>
            </div>
          </div>
          
          <h2 class="success-title">
            ¡Instalación Completada!
          </h2>
          
          <p class="success-message">
            AYMC ha sido instalado correctamente en tu VPS y está listo para usar.
          </p>

          <div class="installation-summary">
            <h3 class="summary-title">📋 Resumen de la Instalación</h3>
            <div class="summary-grid">
              <div class="summary-item">
                <span class="summary-label">API URL:</span>
                <span class="summary-value">{{ installationResult?.apiUrl || 'N/A' }}</span>
              </div>
              <div class="summary-item">
                <span class="summary-label">WebSocket URL:</span>
                <span class="summary-value">{{ installationResult?.wsUrl || 'N/A' }}</span>
              </div>
              <div class="summary-item">
                <span class="summary-label">Base de Datos:</span>
                <span class="summary-value">{{ credentials.dbName }}</span>
              </div>
              <div class="summary-item">
                <span class="summary-label">Puerto:</span>
                <span class="summary-value">{{ credentials.appPort }}</span>
              </div>
              <div class="summary-item">
                <span class="summary-label">Environment:</span>
                <span class="summary-value">Production</span>
              </div>
              <div class="summary-item">
                <span class="summary-label">Servicios:</span>
                <span class="summary-value">✅ Backend, Agent, PostgreSQL</span>
              </div>
            </div>
          </div>

          <div class="form-actions">
            <button 
              @click="viewLogs" 
              class="btn btn-secondary"
            >
              <span class="btn-icon">📜</span>
              Ver Logs Completos
            </button>
            <button 
              @click="goToDashboard" 
              class="btn btn-primary btn-large"
            >
              <span class="btn-icon">🎉</span>
              Ir al Dashboard
            </button>
          </div>
        </div>
      </div>

      <!-- Step 4: Error en instalación -->
      <div v-show="currentStep === 4" class="step-content">
        <div class="installation-error">
          <div class="error-animation">
            <div class="error-circle">
              <div class="error-mark">✗</div>
            </div>
          </div>
          
          <h2 class="error-title">
            Error en la Instalación
          </h2>
          
          <p class="error-message">
            Ocurrió un problema durante la instalación de AYMC.
          </p>

          <div class="error-details">
            <h3 class="details-title">📝 Detalles del Error</h3>
            <pre class="error-log">{{ errorMessage }}</pre>
          </div>

          <div class="error-actions">
            <p class="help-text">
              💡 <strong>Sugerencias:</strong>
            </p>
            <ul class="suggestions-list">
              <li>Verifica que la conexión SSH siga activa</li>
              <li>Asegúrate de que el usuario tiene permisos sudo</li>
              <li>Revisa que el puerto {{ credentials.appPort }} esté disponible</li>
              <li>Comprueba que PostgreSQL se pueda instalar en tu sistema</li>
            </ul>
          </div>

          <div class="form-actions">
            <button 
              @click="viewErrorLogs" 
              class="btn btn-secondary"
            >
              <span class="btn-icon">📜</span>
              Ver Logs Completos
            </button>
            <button 
              @click="retryInstallation" 
              class="btn btn-warning"
            >
              <span class="btn-icon">🔄</span>
              Reintentar Instalación
            </button>
            <button 
              @click="$emit('cancel')" 
              class="btn btn-secondary"
            >
              <span class="btn-icon">❌</span>
              Cancelar
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { invoke } from '@tauri-apps/api/core';
import RemoteTerminal from './RemoteTerminal.vue';

interface Props {
  sshConnected?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  sshConnected: false,
});

const emit = defineEmits<{
  cancel: [];
  complete: [apiUrl: string, wsUrl: string];
}>();

// Pasos del wizard
const steps = [
  { id: 1, label: 'Credenciales' },
  { id: 2, label: 'Instalación' },
  { id: 3, label: 'Completado' },
];

// Estado
const currentStep = ref(1);
const isInstalling = ref(false);
const terminalRef = ref<InstanceType<typeof RemoteTerminal>>();

// Credenciales
const credentials = ref({
  dbPassword: '',
  jwtSecret: '',
  appPort: 8080,
  dbName: 'aymc',
});

// Visibilidad de contraseñas
const showDbPassword = ref(false);
const showJwtSecret = ref(false);

// Resultado de instalación
const installationResult = ref<{
  apiUrl: string;
  wsUrl: string;
  success: boolean;
} | null>(null);

// Error
const errorMessage = ref('');

// Computed
const canInstall = computed(() => {
  return credentials.value.dbPassword.length >= 8 &&
         credentials.value.jwtSecret.length >= 32 &&
         credentials.value.appPort >= 1024 &&
         credentials.value.appPort <= 65535;
});

// Fuerza de contraseña DB
const dbPasswordStrength = computed(() => {
  const pw = credentials.value.dbPassword;
  if (pw.length < 8) return 0;
  if (pw.length < 12) return 1;
  if (pw.length < 16) return 2;
  if (/[A-Z]/.test(pw) && /[a-z]/.test(pw) && /[0-9]/.test(pw) && /[^A-Za-z0-9]/.test(pw)) return 4;
  return 3;
});

const dbPasswordStrengthText = computed(() => {
  const strength = dbPasswordStrength.value;
  if (strength === 0) return 'Muy débil (mín. 8 caracteres)';
  if (strength === 1) return 'Débil';
  if (strength === 2) return 'Media';
  if (strength === 3) return 'Fuerte';
  return 'Muy fuerte';
});

// Fuerza de JWT Secret
const jwtSecretStrength = computed(() => {
  const jwt = credentials.value.jwtSecret;
  if (jwt.length < 32) return 0;
  if (jwt.length < 48) return 1;
  if (jwt.length < 64) return 2;
  if (jwt.length >= 64 && /[A-Z]/.test(jwt) && /[a-z]/.test(jwt) && /[0-9]/.test(jwt)) return 4;
  return 3;
});

const jwtSecretStrengthText = computed(() => {
  const strength = jwtSecretStrength.value;
  if (strength === 0) return 'Muy débil (mín. 32 caracteres)';
  if (strength === 1) return 'Débil';
  if (strength === 2) return 'Media';
  if (strength === 3) return 'Fuerte';
  return 'Muy fuerte';
});

// Funciones
function generateJwtSecret() {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < 64; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  credentials.value.jwtSecret = result;
}

async function startInstallation() {
  if (!canInstall.value || isInstalling.value) return;

  currentStep.value = 2;
  isInstalling.value = true;
  errorMessage.value = '';

  // Obtener referencia al terminal
  const terminal = terminalRef.value;
  if (!terminal) return;

  terminal.clearTerminal();
  terminal.startExecution();
  terminal.writeHeader('Iniciando instalación de AYMC en VPS');
  
  try {
    // Paso 1: Verificar conexión SSH
    terminal.writeInfo('Verificando conexión SSH...');
    const isConnected = await invoke<boolean>('ssh_is_connected');
    
    if (!isConnected) {
      throw new Error('No hay conexión SSH activa');
    }
    terminal.writeSuccess('Conexión SSH verificada');

    // Paso 2: Iniciar instalación
    terminal.writeInfo('Iniciando proceso de instalación...');
    terminal.writeLine('');
    
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
      terminal.writeLine('');
      terminal.writeSuccess('Instalación completada exitosamente');
      terminal.endExecution(0);
      
      installationResult.value = {
        apiUrl: result.api_url,
        wsUrl: result.ws_url,
        success: true,
      };
      
      currentStep.value = 3;
    } else {
      throw new Error(result.message || 'Error desconocido en la instalación');
    }
  } catch (error: any) {
    terminal.writeLine('');
    terminal.writeError(`Error: ${error.message || error}`);
    terminal.endExecution(1);
    
    errorMessage.value = error.message || String(error);
    currentStep.value = 4;
  } finally {
    isInstalling.value = false;
  }
}

function cancelInstallation() {
  if (!isInstalling.value) return;
  
  // Aquí podrías agregar lógica para cancelar el proceso remoto
  isInstalling.value = false;
  currentStep.value = 1;
  
  if (terminalRef.value) {
    terminalRef.value.writeWarning('Instalación cancelada por el usuario');
    terminalRef.value.endExecution(1);
  }
}

function retryInstallation() {
  currentStep.value = 1;
  errorMessage.value = '';
  isInstalling.value = false;
}

function viewLogs() {
  // Aquí podrías abrir una modal con los logs completos
  if (terminalRef.value) {
    // Terminal ya muestra los logs
  }
}

function viewErrorLogs() {
  // Similar a viewLogs pero para errores
  if (terminalRef.value) {
    // Terminal ya muestra los logs de error
  }
}

function goToDashboard() {
  if (installationResult.value) {
    emit('complete', installationResult.value.apiUrl, installationResult.value.wsUrl);
  }
}
</script>

<style scoped>
.installation-wizard {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  overflow: hidden;
}

.wizard-header {
  background: rgba(255, 255, 255, 0.95);
  padding: 24px 32px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header-content {
  margin-bottom: 24px;
}

.wizard-title {
  font-size: 32px;
  font-weight: 700;
  color: #2d3748;
  margin: 0 0 8px 0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.title-icon {
  font-size: 36px;
}

.wizard-subtitle {
  font-size: 16px;
  color: #718096;
  margin: 0;
}

.progress-steps {
  display: flex;
  gap: 32px;
  justify-content: center;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  position: relative;
  opacity: 0.5;
  transition: opacity 0.3s;
}

.step.active,
.step.completed {
  opacity: 1;
}

.step-number {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #e2e8f0;
  color: #718096;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 18px;
  transition: all 0.3s;
}

.step.active .step-number {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.step.completed .step-number {
  background: #4caf50;
  color: white;
}

.step-label {
  font-size: 14px;
  font-weight: 600;
  color: #2d3748;
}

.wizard-content {
  flex: 1;
  overflow-y: auto;
  padding: 32px;
}

.step-content {
  max-width: 900px;
  margin: 0 auto;
}

/* Credentials Form */
.credentials-form {
  background: white;
  border-radius: 16px;
  padding: 32px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.form-title {
  font-size: 24px;
  font-weight: 700;
  color: #2d3748;
  margin: 0 0 16px 0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.icon {
  font-size: 28px;
}

.form-description {
  margin-bottom: 32px;
  color: #718096;
  line-height: 1.6;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 24px;
  margin-bottom: 32px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  font-weight: 600;
  color: #2d3748;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.label-icon {
  font-size: 16px;
}

.form-input {
  padding: 12px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
  width: 100%;
}

.form-input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.form-input:disabled {
  background: #f7fafc;
  cursor: not-allowed;
}

.input-with-button {
  display: flex;
  gap: 8px;
}

.input-with-button .form-input {
  flex: 1;
}

.toggle-password-btn,
.generate-btn {
  padding: 12px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  background: white;
  cursor: pointer;
  font-size: 16px;
  transition: all 0.2s;
}

.toggle-password-btn:hover,
.generate-btn:hover {
  background: #f7fafc;
  border-color: #cbd5e0;
}

.toggle-password-btn:disabled,
.generate-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.password-strength {
  display: flex;
  align-items: center;
  gap: 12px;
}

.strength-bar {
  flex: 1;
  height: 6px;
  background: #e2e8f0;
  border-radius: 3px;
  overflow: hidden;
  position: relative;
}

.strength-fill {
  height: 100%;
  transition: all 0.3s;
  background: #cbd5e0;
}

.strength-0 .strength-fill {
  width: 0%;
}

.strength-1 .strength-fill {
  width: 25%;
  background: #f44336;
}

.strength-2 .strength-fill {
  width: 50%;
  background: #ff9800;
}

.strength-3 .strength-fill {
  width: 75%;
  background: #ffc107;
}

.strength-4 .strength-fill {
  width: 100%;
  background: #4caf50;
}

.strength-text {
  font-size: 12px;
  color: #718096;
  white-space: nowrap;
}

.form-hint {
  font-size: 12px;
  color: #a0aec0;
}

.form-actions {
  display: flex;
  gap: 16px;
  justify-content: flex-end;
}

.btn {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-icon {
  font-size: 16px;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.4);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: white;
  color: #2d3748;
  border: 2px solid #e2e8f0;
}

.btn-secondary:hover:not(:disabled) {
  background: #f7fafc;
  border-color: #cbd5e0;
}

.btn-warning {
  background: #ff9800;
  color: white;
}

.btn-warning:hover {
  background: #f57c00;
  transform: translateY(-2px);
}

.btn-large {
  padding: 16px 32px;
  font-size: 16px;
}

/* Installation Progress */
.installation-progress {
  background: white;
  border-radius: 16px;
  padding: 32px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.progress-info {
  text-align: center;
  color: #718096;
  line-height: 1.6;
}

.progress-warning {
  color: #ff9800;
  font-weight: 600;
  margin-top: 8px;
}

.spinning {
  display: inline-block;
  animation: spin 2s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* Installation Complete */
.installation-complete {
  background: white;
  border-radius: 16px;
  padding: 48px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  text-align: center;
}

.success-animation {
  margin-bottom: 32px;
}

.checkmark-circle {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  background: #4caf50;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: scaleIn 0.5s ease-out;
}

.checkmark {
  width: 40px;
  height: 60px;
  border: solid white;
  border-width: 0 6px 6px 0;
  transform: rotate(45deg);
  animation: checkmarkDraw 0.4s 0.3s ease-out forwards;
  opacity: 0;
}

@keyframes scaleIn {
  from {
    transform: scale(0);
  }
  to {
    transform: scale(1);
  }
}

@keyframes checkmarkDraw {
  to {
    opacity: 1;
  }
}

.success-title {
  font-size: 32px;
  font-weight: 700;
  color: #2d3748;
  margin: 0 0 16px 0;
}

.success-message {
  font-size: 16px;
  color: #718096;
  margin: 0 0 32px 0;
}

.installation-summary {
  background: #f7fafc;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 32px;
  text-align: left;
}

.summary-title {
  font-size: 18px;
  font-weight: 700;
  color: #2d3748;
  margin: 0 0 16px 0;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.summary-label {
  font-size: 12px;
  font-weight: 600;
  color: #a0aec0;
  text-transform: uppercase;
}

.summary-value {
  font-size: 14px;
  color: #2d3748;
  font-weight: 600;
}

/* Installation Error */
.installation-error {
  background: white;
  border-radius: 16px;
  padding: 48px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  text-align: center;
}

.error-animation {
  margin-bottom: 32px;
}

.error-circle {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  background: #f44336;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: scaleIn 0.5s ease-out;
}

.error-mark {
  font-size: 60px;
  color: white;
  font-weight: 700;
}

.error-title {
  font-size: 32px;
  font-weight: 700;
  color: #2d3748;
  margin: 0 0 16px 0;
}

.error-message {
  font-size: 16px;
  color: #718096;
  margin: 0 0 32px 0;
}

.error-details {
  background: #f7fafc;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
  text-align: left;
}

.details-title {
  font-size: 18px;
  font-weight: 700;
  color: #2d3748;
  margin: 0 0 16px 0;
}

.error-log {
  background: #2d3748;
  color: #f44336;
  padding: 16px;
  border-radius: 8px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.error-actions {
  background: #fff3cd;
  border: 2px solid #ffc107;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 32px;
  text-align: left;
}

.help-text {
  color: #856404;
  margin: 0 0 12px 0;
}

.suggestions-list {
  margin: 0;
  padding-left: 24px;
  color: #856404;
  line-height: 1.8;
}

/* Responsive */
@media (max-width: 768px) {
  .wizard-header {
    padding: 16px 20px;
  }

  .wizard-title {
    font-size: 24px;
  }

  .progress-steps {
    gap: 16px;
  }

  .step-number {
    width: 32px;
    height: 32px;
    font-size: 16px;
  }

  .wizard-content {
    padding: 16px;
  }

  .credentials-form,
  .installation-progress,
  .installation-complete,
  .installation-error {
    padding: 24px;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>
