<template>
  <Teleport to="body">
    <Transition name="dialog">
      <div v-if="isOpen" class="dialog-overlay" @click.self="close">
        <div class="dialog-container" @click.stop>
          <!-- Header -->
          <div class="dialog-header">
            <div class="error-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10" />
                <path d="M12 8v4M12 16h.01" />
              </svg>
            </div>
            <h2 class="dialog-title">Error en la Instalación</h2>
            <button @click="close" class="close-button" aria-label="Cerrar">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M18 6L6 18M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Error Details -->
          <div class="dialog-content">
            <div class="error-message">
              <h3>{{ error.title || 'Error Desconocido' }}</h3>
              <p>{{ error.message }}</p>
            </div>

            <!-- Error Type Badge -->
            <div v-if="error.type" class="error-type">
              <span class="error-type-badge" :class="`type-${error.type}`">
                {{ errorTypeLabel }}
              </span>
            </div>

            <!-- Sugerencias de Solución -->
            <div v-if="suggestions.length > 0" class="suggestions">
              <h4>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
                </svg>
                Sugerencias de Solución
              </h4>
              <ul class="suggestions-list">
                <li v-for="(suggestion, index) in suggestions" :key="index">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M9 18l6-6-6-6" />
                  </svg>
                  <span>{{ suggestion }}</span>
                </li>
              </ul>
            </div>

            <!-- Diagnostics -->
            <div v-if="showDiagnostics && diagnostics" class="diagnostics">
              <button @click="toggleDiagnostics" class="diagnostics-toggle">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                  <path d="M14 2v6h6M16 13H8M16 17H8M10 9H8" />
                </svg>
                {{ diagnosticsExpanded ? 'Ocultar' : 'Ver' }} Información de Diagnóstico
              </button>
              <div v-if="diagnosticsExpanded" class="diagnostics-content">
                <pre>{{ diagnostics }}</pre>
              </div>
            </div>

            <!-- Stack Trace (solo en dev) -->
            <div v-if="error.stack && isDev" class="stack-trace">
              <button @click="toggleStackTrace" class="stack-toggle">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M10 3H6a2 2 0 0 0-2 2v14c0 1.1.9 2 2 2h4M16 17l5-5-5-5M21 12H9" />
                </svg>
                {{ stackExpanded ? 'Ocultar' : 'Ver' }} Stack Trace
              </button>
              <div v-if="stackExpanded" class="stack-content">
                <pre>{{ error.stack }}</pre>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="dialog-actions">
            <button 
              v-if="canRetry" 
              @click="handleRetry" 
              class="action-button action-retry"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M1 4v6h6M23 20v-6h-6" />
                <path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15" />
              </svg>
              Reintentar
            </button>

            <button 
              v-if="canSkip" 
              @click="handleSkip" 
              class="action-button action-skip"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M13 17l5-5-5-5M6 17l5-5-5-5" />
              </svg>
              Saltar Paso
            </button>

            <button 
              @click="handleViewLogs" 
              class="action-button action-logs"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M4 6h16M4 12h16M4 18h16" />
              </svg>
              Ver Logs
            </button>

            <button 
              @click="handleCancel" 
              class="action-button action-cancel"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M18 6L6 18M6 6l12 12" />
              </svg>
              Cancelar Instalación
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

export type ErrorType = 
  | 'network'
  | 'permission'
  | 'port'
  | 'disk'
  | 'dependency'
  | 'configuration'
  | 'unknown';

export interface InstallationError {
  title?: string;
  message: string;
  type?: ErrorType;
  stack?: string;
  step?: number;
  retryable?: boolean;
  skippable?: boolean;
}

interface Props {
  isOpen: boolean;
  error: InstallationError;
  diagnostics?: string;
  showDiagnostics?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  diagnostics: undefined,
  showDiagnostics: true,
});

const emit = defineEmits<{
  close: [];
  retry: [];
  skip: [];
  'view-logs': [];
  cancel: [];
}>();

const diagnosticsExpanded = ref(false);
const stackExpanded = ref(false);
const isDev = import.meta.env.DEV;

const errorTypeLabel = computed(() => {
  const labels: Record<ErrorType, string> = {
    network: 'Error de Red',
    permission: 'Error de Permisos',
    port: 'Error de Puerto',
    disk: 'Error de Disco',
    dependency: 'Error de Dependencias',
    configuration: 'Error de Configuración',
    unknown: 'Error Desconocido',
  };
  return labels[props.error.type || 'unknown'];
});

const suggestions = computed(() => {
  const errorSuggestions: Record<ErrorType, string[]> = {
    network: [
      'Verifica tu conexión a internet',
      'Confirma que la VPS esté accesible',
      'Revisa las reglas de firewall',
      'Intenta reconectar vía SSH',
    ],
    permission: [
      'Asegúrate de que el usuario tenga permisos sudo',
      'Verifica los permisos de archivos y directorios',
      'Ejecuta el comando con privilegios elevados',
      'Contacta al administrador del sistema',
    ],
    port: [
      'Verifica que el puerto no esté en uso',
      'Intenta usar un puerto diferente',
      'Detén el servicio que está usando el puerto',
      'Revisa las configuraciones de firewall',
    ],
    disk: [
      'Libera espacio en disco',
      'Elimina archivos temporales o logs antiguos',
      'Verifica el espacio disponible con "df -h"',
      'Considera aumentar el tamaño del disco',
    ],
    dependency: [
      'Verifica que todas las dependencias estén instaladas',
      'Actualiza los paquetes del sistema',
      'Revisa los logs de instalación de paquetes',
      'Intenta instalar las dependencias manualmente',
    ],
    configuration: [
      'Revisa los archivos de configuración',
      'Verifica las variables de entorno',
      'Confirma que los servicios estén corriendo',
      'Consulta la documentación de configuración',
    ],
    unknown: [
      'Revisa los logs del sistema para más detalles',
      'Verifica el estado de la conexión SSH',
      'Intenta reiniciar la instalación',
      'Contacta al soporte técnico si el problema persiste',
    ],
  };

  return errorSuggestions[props.error.type || 'unknown'];
});

const canRetry = computed(() => props.error.retryable !== false);
const canSkip = computed(() => props.error.skippable === true);

function toggleDiagnostics() {
  diagnosticsExpanded.value = !diagnosticsExpanded.value;
}

function toggleStackTrace() {
  stackExpanded.value = !stackExpanded.value;
}

function close() {
  emit('close');
}

function handleRetry() {
  emit('retry');
  close();
}

function handleSkip() {
  emit('skip');
  close();
}

function handleViewLogs() {
  emit('view-logs');
}

function handleCancel() {
  emit('cancel');
  close();
}
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 1rem;
}

.dialog-container {
  background: linear-gradient(135deg, #1f2937 0%, #111827 100%);
  border-radius: 1rem;
  max-width: 600px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  color: white;
}

/* Header */
.dialog-header {
  padding: 1.5rem;
  border-bottom: 1px solid rgba(239, 68, 68, 0.2);
  display: flex;
  align-items: center;
  gap: 1rem;
  position: relative;
}

.error-icon {
  width: 48px;
  height: 48px;
  background: rgba(239, 68, 68, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: pulse 2s ease-in-out infinite;
}

.error-icon svg {
  width: 28px;
  height: 28px;
  color: #ef4444;
}

.dialog-title {
  flex: 1;
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
}

.close-button {
  width: 32px;
  height: 32px;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
  color: white;
}

.close-button:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: rotate(90deg);
}

.close-button svg {
  width: 18px;
  height: 18px;
}

/* Content */
.dialog-content {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.error-message h3 {
  margin: 0 0 0.5rem 0;
  font-size: 1.2rem;
  font-weight: 600;
  color: #ef4444;
}

.error-message p {
  margin: 0;
  line-height: 1.6;
  opacity: 0.9;
}

.error-type {
  display: flex;
}

.error-type-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 999px;
  font-size: 0.85rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.type-network { background: rgba(59, 130, 246, 0.2); color: #60a5fa; }
.type-permission { background: rgba(245, 158, 11, 0.2); color: #fbbf24; }
.type-port { background: rgba(236, 72, 153, 0.2); color: #f472b6; }
.type-disk { background: rgba(168, 85, 247, 0.2); color: #c084fc; }
.type-dependency { background: rgba(34, 197, 94, 0.2); color: #4ade80; }
.type-configuration { background: rgba(251, 191, 36, 0.2); color: #fbbf24; }
.type-unknown { background: rgba(156, 163, 175, 0.2); color: #9ca3af; }

/* Suggestions */
.suggestions h4 {
  margin: 0 0 1rem 0;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1.1rem;
  color: #60a5fa;
}

.suggestions h4 svg {
  width: 20px;
  height: 20px;
}

.suggestions-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.suggestions-list li {
  display: flex;
  align-items: start;
  gap: 0.75rem;
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 0.5rem;
  transition: all 0.2s ease;
}

.suggestions-list li:hover {
  background: rgba(255, 255, 255, 0.1);
  transform: translateX(4px);
}

.suggestions-list li svg {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  margin-top: 0.1rem;
  color: #60a5fa;
}

/* Diagnostics & Stack Trace */
.diagnostics,
.stack-trace {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.diagnostics-toggle,
.stack-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 0.5rem;
  color: white;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.diagnostics-toggle:hover,
.stack-toggle:hover {
  background: rgba(255, 255, 255, 0.15);
}

.diagnostics-toggle svg,
.stack-toggle svg {
  width: 16px;
  height: 16px;
}

.diagnostics-content,
.stack-content {
  padding: 1rem;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 0.5rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  max-height: 200px;
  overflow-y: auto;
}

.diagnostics-content pre,
.stack-content pre {
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 0.85rem;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
  color: #d1d5db;
}

/* Actions */
.dialog-actions {
  padding: 1.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  justify-content: flex-end;
}

.action-button {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 0.5rem;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  transition: all 0.2s ease;
}

.action-button svg {
  width: 18px;
  height: 18px;
}

.action-retry {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: white;
}

.action-retry:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.action-skip {
  background: rgba(168, 85, 247, 0.2);
  color: #c084fc;
  border: 1px solid rgba(168, 85, 247, 0.3);
}

.action-skip:hover {
  background: rgba(168, 85, 247, 0.3);
}

.action-logs {
  background: rgba(255, 255, 255, 0.1);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.action-logs:hover {
  background: rgba(255, 255, 255, 0.15);
}

.action-cancel {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.action-cancel:hover {
  background: rgba(239, 68, 68, 0.3);
}

/* Transitions */
.dialog-enter-active,
.dialog-leave-active {
  transition: opacity 0.3s ease;
}

.dialog-enter-from,
.dialog-leave-to {
  opacity: 0;
}

.dialog-enter-active .dialog-container,
.dialog-leave-active .dialog-container {
  transition: transform 0.3s ease;
}

.dialog-enter-from .dialog-container,
.dialog-leave-to .dialog-container {
  transform: scale(0.9);
}

/* Animations */
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

/* Scrollbar */
.dialog-container::-webkit-scrollbar,
.diagnostics-content::-webkit-scrollbar,
.stack-content::-webkit-scrollbar {
  width: 8px;
}

.dialog-container::-webkit-scrollbar-track,
.diagnostics-content::-webkit-scrollbar-track,
.stack-content::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 4px;
}

.dialog-container::-webkit-scrollbar-thumb,
.diagnostics-content::-webkit-scrollbar-thumb,
.stack-content::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}

.dialog-container::-webkit-scrollbar-thumb:hover,
.diagnostics-content::-webkit-scrollbar-thumb:hover,
.stack-content::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>
