<template>
  <div class="installation-progress">
    <!-- Header con fase actual -->
    <div class="progress-header">
      <div class="phase-indicator">
        <div class="phase-icon" :class="`phase-${currentPhase}`">
          <svg v-if="currentPhase === 'completed'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M20 6L9 17l-5-5" />
          </svg>
          <svg v-else-if="currentPhase === 'failed'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 6L6 18M6 6l12 12" />
          </svg>
          <div v-else class="spinner"></div>
        </div>
        <div class="phase-info">
          <h3 class="phase-title">{{ phaseTitle }}</h3>
          <p class="phase-message">{{ message }}</p>
        </div>
      </div>
      <div class="progress-percentage">{{ percentage }}%</div>
    </div>

    <!-- Barra de progreso general -->
    <div class="progress-bar-container">
      <div 
        class="progress-bar-fill" 
        :style="{ width: `${percentage}%` }"
        :class="{ 'progress-error': currentPhase === 'failed', 'progress-success': currentPhase === 'completed' }"
      ></div>
    </div>

    <!-- Lista de pasos -->
    <div class="steps-list">
      <div 
        v-for="step in steps" 
        :key="step.id"
        class="step-item"
        :class="`step-${step.status}`"
      >
        <div class="step-indicator">
          <div class="step-icon">
            <!-- Completado -->
            <svg v-if="step.status === 'completed'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20 6L9 17l-5-5" />
            </svg>
            <!-- Fallido -->
            <svg v-else-if="step.status === 'failed'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12" />
            </svg>
            <!-- En progreso -->
            <div v-else-if="step.status === 'running'" class="spinner-small"></div>
            <!-- Saltado -->
            <svg v-else-if="step.status === 'skipped'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M13 17l5-5-5-5M6 17l5-5-5-5" />
            </svg>
            <!-- Pendiente -->
            <div v-else class="step-number">{{ step.id }}</div>
          </div>
          <div class="step-line" v-if="step.id < steps.length"></div>
        </div>

        <div class="step-content">
          <div class="step-header">
            <h4 class="step-name">{{ step.name }}</h4>
            <span v-if="step.startTime && step.endTime" class="step-duration">
              {{ formatDuration(step.endTime - step.startTime) }}
            </span>
            <span v-else-if="step.startTime" class="step-duration">
              {{ formatDuration(Date.now() - step.startTime) }}
            </span>
          </div>
          <p class="step-description">{{ step.description }}</p>
          
          <!-- Error message -->
          <div v-if="step.error" class="step-error">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10" />
              <path d="M12 8v4M12 16h.01" />
            </svg>
            <span>{{ step.error }}</span>
          </div>

          <!-- Botón de reintento -->
          <button 
            v-if="step.status === 'failed' && step.canRetry" 
            @click="$emit('retry-step', step.id)"
            class="retry-button"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M1 4v6h6M23 20v-6h-6" />
              <path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15" />
            </svg>
            Reintentar
          </button>
        </div>
      </div>
    </div>

    <!-- Tiempo estimado -->
    <div v-if="estimatedTimeRemaining" class="estimated-time">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10" />
        <path d="M12 6v6l4 2" />
      </svg>
      <span>Tiempo estimado restante: {{ estimatedTimeRemaining }}</span>
    </div>

    <!-- Controles -->
    <div v-if="showControls" class="progress-controls">
      <button 
        v-if="currentPhase !== 'completed' && currentPhase !== 'failed'" 
        @click="$emit('pause')"
        class="control-button"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="6" y="4" width="4" height="16" />
          <rect x="14" y="4" width="4" height="16" />
        </svg>
        Pausar
      </button>
      
      <button 
        @click="$emit('cancel')"
        class="control-button control-cancel"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M18 6L6 18M6 6l12 12" />
        </svg>
        Cancelar
      </button>

      <button 
        v-if="currentPhase === 'failed'" 
        @click="$emit('view-logs')"
        class="control-button"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
          <path d="M14 2v6h6M16 13H8M16 17H8M10 9H8" />
        </svg>
        Ver Logs
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

export interface InstallationStep {
  id: number;
  name: string;
  description: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped';
  progress: number;
  startTime?: number;
  endTime?: number;
  error?: string;
  canRetry: boolean;
}

export type InstallationPhase = 
  | 'validation'
  | 'preparation'
  | 'installation'
  | 'configuration'
  | 'verification'
  | 'completed'
  | 'failed';

interface Props {
  phase: InstallationPhase;
  percentage: number;
  message: string;
  steps: InstallationStep[];
  showControls?: boolean;
  estimatedTimeRemaining?: string;
}

const props = withDefaults(defineProps<Props>(), {
  showControls: true,
  estimatedTimeRemaining: undefined,
});

defineEmits<{
  'retry-step': [stepId: number];
  'pause': [];
  'cancel': [];
  'view-logs': [];
}>();

const currentPhase = computed(() => props.phase);

const phaseTitle = computed(() => {
  const titles: Record<InstallationPhase, string> = {
    validation: 'Validando Pre-requisitos',
    preparation: 'Preparando Instalación',
    installation: 'Instalando AYMC',
    configuration: 'Configurando Servicios',
    verification: 'Verificando Instalación',
    completed: 'Instalación Completada',
    failed: 'Instalación Fallida',
  };
  return titles[props.phase];
});

function formatDuration(ms: number): string {
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  return `${minutes}m ${remainingSeconds}s`;
}
</script>

<style scoped>
.installation-progress {
  width: 100%;
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 1rem;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  color: white;
}

/* Header */
.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.phase-indicator {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.phase-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px);
}

.phase-icon svg {
  width: 32px;
  height: 32px;
  color: white;
}

.phase-icon.phase-completed {
  background: #10b981;
  animation: pulse 2s ease-in-out infinite;
}

.phase-icon.phase-failed {
  background: #ef4444;
  animation: shake 0.5s ease-in-out;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.phase-info {
  flex: 1;
}

.phase-title {
  font-size: 1.5rem;
  font-weight: 700;
  margin: 0 0 0.25rem 0;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.phase-message {
  margin: 0;
  opacity: 0.9;
  font-size: 0.95rem;
}

.progress-percentage {
  font-size: 2.5rem;
  font-weight: 700;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

/* Progress Bar */
.progress-bar-container {
  height: 8px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 999px;
  overflow: hidden;
  margin-bottom: 2rem;
}

.progress-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #60a5fa, #3b82f6);
  border-radius: 999px;
  transition: width 0.3s ease-out;
  animation: shimmer 2s ease-in-out infinite;
}

.progress-bar-fill.progress-success {
  background: linear-gradient(90deg, #34d399, #10b981);
}

.progress-bar-fill.progress-error {
  background: linear-gradient(90deg, #f87171, #ef4444);
}

@keyframes shimmer {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.8; }
}

/* Steps List */
.steps-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.step-item {
  display: flex;
  gap: 1rem;
  padding: 1rem;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 0.75rem;
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
}

.step-item.step-running {
  background: rgba(59, 130, 246, 0.2);
  box-shadow: 0 0 20px rgba(59, 130, 246, 0.3);
}

.step-item.step-completed {
  background: rgba(16, 185, 129, 0.2);
}

.step-item.step-failed {
  background: rgba(239, 68, 68, 0.2);
  box-shadow: 0 0 20px rgba(239, 68, 68, 0.3);
}

.step-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
}

.step-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 1;
}

.step-item.step-completed .step-icon {
  background: #10b981;
}

.step-item.step-failed .step-icon {
  background: #ef4444;
}

.step-item.step-running .step-icon {
  background: #3b82f6;
}

.step-icon svg {
  width: 20px;
  height: 20px;
  color: white;
}

.step-number {
  font-weight: 700;
  font-size: 1.1rem;
}

.spinner-small {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.step-line {
  flex: 1;
  width: 2px;
  background: rgba(255, 255, 255, 0.2);
  margin-top: 0.5rem;
}

.step-content {
  flex: 1;
}

.step-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.step-name {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
}

.step-duration {
  font-size: 0.85rem;
  opacity: 0.7;
  font-family: monospace;
}

.step-description {
  margin: 0 0 0.5rem 0;
  opacity: 0.8;
  font-size: 0.9rem;
}

.step-error {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem;
  background: rgba(239, 68, 68, 0.2);
  border: 1px solid rgba(239, 68, 68, 0.4);
  border-radius: 0.5rem;
  margin-top: 0.75rem;
  font-size: 0.9rem;
}

.step-error svg {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.retry-button {
  margin-top: 0.75rem;
  padding: 0.5rem 1rem;
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 0.5rem;
  color: white;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  transition: all 0.2s ease;
}

.retry-button:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.retry-button svg {
  width: 16px;
  height: 16px;
}

/* Estimated Time */
.estimated-time {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 0.75rem;
  margin-bottom: 1.5rem;
}

.estimated-time svg {
  width: 20px;
  height: 20px;
}

/* Controls */
.progress-controls {
  display: flex;
  gap: 1rem;
  justify-content: center;
}

.control-button {
  padding: 0.75rem 1.5rem;
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 0.75rem;
  color: white;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  transition: all 0.2s ease;
}

.control-button:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.control-button.control-cancel {
  background: rgba(239, 68, 68, 0.3);
  border-color: rgba(239, 68, 68, 0.5);
}

.control-button.control-cancel:hover {
  background: rgba(239, 68, 68, 0.5);
}

.control-button svg {
  width: 18px;
  height: 18px;
}

/* Animations */
@keyframes spin {
  to { transform: rotate(360deg); }
}

@keyframes pulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.05); }
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-10px); }
  75% { transform: translateX(10px); }
}
</style>
