<template>
  <div class="terminal-container">
    <div class="terminal-header">
      <div class="terminal-title">
        <span class="terminal-icon">💻</span>
        <span>{{ title }}</span>
      </div>
      <div class="terminal-controls">
        <button 
          v-if="canClear" 
          @click="clearTerminal" 
          class="control-btn clear-btn"
          title="Limpiar terminal"
        >
          🗑️
        </button>
        <button 
          v-if="canCopy" 
          @click="copyOutput" 
          class="control-btn copy-btn"
          title="Copiar salida"
        >
          📋
        </button>
        <button 
          v-if="isRunning && canStop"
          @click="stopExecution" 
          class="control-btn stop-btn"
          title="Detener ejecución"
        >
          ⏹️
        </button>
      </div>
    </div>
    
    <div ref="terminalRef" class="terminal-content"></div>
    
    <div v-if="showStatus" class="terminal-status">
      <div class="status-item">
        <span class="status-label">Estado:</span>
        <span class="status-value" :class="statusClass">{{ statusText }}</span>
      </div>
      <div v-if="duration > 0" class="status-item">
        <span class="status-label">Duración:</span>
        <span class="status-value">{{ formattedDuration }}</span>
      </div>
      <div v-if="exitCode !== null" class="status-item">
        <span class="status-label">Código de salida:</span>
        <span class="status-value" :class="exitCodeClass">{{ exitCode }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed, watch } from 'vue';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';

interface Props {
  title?: string;
  autoFit?: boolean;
  canClear?: boolean;
  canCopy?: boolean;
  canStop?: boolean;
  showStatus?: boolean;
  theme?: 'dark' | 'light';
  fontSize?: number;
}

const props = withDefaults(defineProps<Props>(), {
  title: 'Terminal Remota',
  autoFit: true,
  canClear: true,
  canCopy: true,
  canStop: true,
  showStatus: true,
  theme: 'dark',
  fontSize: 13,
});

const emit = defineEmits<{
  ready: [];
  stop: [];
}>();

// Referencias
const terminalRef = ref<HTMLElement>();
let terminal: Terminal | null = null;
let fitAddon: FitAddon | null = null;
let resizeObserver: ResizeObserver | null = null;

// Estado
const isRunning = ref(false);
const startTime = ref<number | null>(null);
const endTime = ref<number | null>(null);
const duration = ref(0);
const exitCode = ref<number | null>(null);
const outputBuffer = ref<string[]>([]);

// Timer para actualizar duración
let durationInterval: number | null = null;

// Computed
const statusText = computed(() => {
  if (isRunning.value) return 'Ejecutando...';
  if (exitCode.value === null) return 'Esperando...';
  if (exitCode.value === 0) return 'Completado';
  return 'Error';
});

const statusClass = computed(() => {
  if (isRunning.value) return 'status-running';
  if (exitCode.value === null) return 'status-idle';
  if (exitCode.value === 0) return 'status-success';
  return 'status-error';
});

const exitCodeClass = computed(() => {
  if (exitCode.value === 0) return 'exit-success';
  return 'exit-error';
});

const formattedDuration = computed(() => {
  const seconds = Math.floor(duration.value / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  return `${minutes}m ${remainingSeconds}s`;
});

// Funciones
function initTerminal() {
  if (!terminalRef.value) return;

  // Crear terminal con tema
  terminal = new Terminal({
    cursorBlink: true,
    cursorStyle: 'block',
    fontSize: props.fontSize,
    fontFamily: 'Monaco, Menlo, "Courier New", monospace',
    theme: props.theme === 'dark' ? {
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
      brightBlack: '#666666',
      brightRed: '#f14c4c',
      brightGreen: '#23d18b',
      brightYellow: '#f5f543',
      brightBlue: '#3b8eea',
      brightMagenta: '#d670d6',
      brightCyan: '#29b8db',
      brightWhite: '#e5e5e5',
    } : {
      background: '#ffffff',
      foreground: '#000000',
      cursor: '#000000',
    },
    scrollback: 10000,
    convertEol: true,
  });

  // FitAddon para responsive
  if (props.autoFit) {
    fitAddon = new FitAddon();
    terminal.loadAddon(fitAddon);
  }

  // Abrir terminal en el elemento
  terminal.open(terminalRef.value);

  // Ajustar tamaño inicial
  if (fitAddon) {
    fitAddon.fit();
  }

  // Observer para resize automático
  if (props.autoFit && terminalRef.value) {
    resizeObserver = new ResizeObserver(() => {
      if (fitAddon && terminal) {
        try {
          fitAddon.fit();
        } catch (error) {
          console.warn('Error al ajustar terminal:', error);
        }
      }
    });
    resizeObserver.observe(terminalRef.value);
  }

  emit('ready');
}

function writeLine(text: string) {
  if (!terminal) return;
  
  // Agregar al buffer
  outputBuffer.value.push(text);
  
  // Escribir en terminal con nueva línea
  terminal.writeln(text);
}

function write(text: string) {
  if (!terminal) return;
  
  // Agregar al buffer (sin nueva línea)
  outputBuffer.value.push(text);
  
  // Escribir en terminal
  terminal.write(text);
}

function clearTerminal() {
  if (!terminal) return;
  terminal.clear();
  outputBuffer.value = [];
  resetStatus();
}

function copyOutput() {
  const text = outputBuffer.value.join('\n');
  navigator.clipboard.writeText(text).then(() => {
    // Mostrar mensaje temporal
    if (terminal) {
      terminal.writeln('\x1b[32m✓ Salida copiada al portapapeles\x1b[0m');
    }
  }).catch(error => {
    console.error('Error al copiar:', error);
    if (terminal) {
      terminal.writeln('\x1b[31m✗ Error al copiar al portapapeles\x1b[0m');
    }
  });
}

function stopExecution() {
  if (!isRunning.value) return;
  
  emit('stop');
  endExecution(1); // Exit code 1 por cancelación manual
  
  if (terminal) {
    terminal.writeln('\x1b[33m⚠ Ejecución detenida por el usuario\x1b[0m');
  }
}

function startExecution() {
  isRunning.value = true;
  startTime.value = Date.now();
  endTime.value = null;
  exitCode.value = null;
  
  // Iniciar actualización de duración
  if (durationInterval) clearInterval(durationInterval);
  durationInterval = window.setInterval(() => {
    if (startTime.value && !endTime.value) {
      duration.value = Date.now() - startTime.value;
    }
  }, 100);
}

function endExecution(code: number = 0) {
  isRunning.value = false;
  endTime.value = Date.now();
  exitCode.value = code;
  
  if (startTime.value && endTime.value) {
    duration.value = endTime.value - startTime.value;
  }
  
  // Detener actualización de duración
  if (durationInterval) {
    clearInterval(durationInterval);
    durationInterval = null;
  }
}

function resetStatus() {
  isRunning.value = false;
  startTime.value = null;
  endTime.value = null;
  duration.value = 0;
  exitCode.value = null;
  
  if (durationInterval) {
    clearInterval(durationInterval);
    durationInterval = null;
  }
}

function writeSuccess(message: string) {
  writeLine(`\x1b[32m✓ ${message}\x1b[0m`);
}

function writeError(message: string) {
  writeLine(`\x1b[31m✗ ${message}\x1b[0m`);
}

function writeWarning(message: string) {
  writeLine(`\x1b[33m⚠ ${message}\x1b[0m`);
}

function writeInfo(message: string) {
  writeLine(`\x1b[36mℹ ${message}\x1b[0m`);
}

function writeHeader(message: string) {
  writeLine(`\x1b[1m\x1b[36m${'='.repeat(60)}\x1b[0m`);
  writeLine(`\x1b[1m\x1b[36m${message}\x1b[0m`);
  writeLine(`\x1b[1m\x1b[36m${'='.repeat(60)}\x1b[0m`);
}

// Lifecycle
onMounted(() => {
  initTerminal();
});

onBeforeUnmount(() => {
  if (resizeObserver) {
    resizeObserver.disconnect();
  }
  if (durationInterval) {
    clearInterval(durationInterval);
  }
  if (terminal) {
    terminal.dispose();
  }
});

// Exponer métodos para el componente padre
defineExpose({
  write,
  writeLine,
  writeSuccess,
  writeError,
  writeWarning,
  writeInfo,
  writeHeader,
  clearTerminal,
  startExecution,
  endExecution,
  resetStatus,
  terminal,
});
</script>

<style scoped>
.terminal-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1e1e1e;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: linear-gradient(135deg, #2d2d30 0%, #252526 100%);
  border-bottom: 1px solid #3e3e42;
}

.terminal-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #cccccc;
  font-size: 14px;
}

.terminal-icon {
  font-size: 18px;
}

.terminal-controls {
  display: flex;
  gap: 8px;
}

.control-btn {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  padding: 6px 12px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  color: #cccccc;
}

.control-btn:hover {
  background: rgba(255, 255, 255, 0.15);
  border-color: rgba(255, 255, 255, 0.3);
  transform: translateY(-1px);
}

.control-btn:active {
  transform: translateY(0);
}

.stop-btn:hover {
  background: rgba(244, 67, 54, 0.2);
  border-color: #f44336;
  color: #f44336;
}

.terminal-content {
  flex: 1;
  overflow: hidden;
  padding: 8px;
}

.terminal-status {
  display: flex;
  gap: 24px;
  padding: 12px 16px;
  background: #252526;
  border-top: 1px solid #3e3e42;
  font-size: 13px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-label {
  color: #858585;
  font-weight: 500;
}

.status-value {
  color: #cccccc;
  font-weight: 600;
}

.status-running {
  color: #4fc3f7 !important;
  animation: pulse 1.5s ease-in-out infinite;
}

.status-idle {
  color: #858585 !important;
}

.status-success {
  color: #4caf50 !important;
}

.status-error {
  color: #f44336 !important;
}

.exit-success {
  color: #4caf50 !important;
}

.exit-error {
  color: #f44336 !important;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.6;
  }
}

/* Responsive */
@media (max-width: 768px) {
  .terminal-header {
    padding: 10px 12px;
  }

  .terminal-title {
    font-size: 13px;
  }

  .control-btn {
    padding: 5px 10px;
    font-size: 13px;
  }

  .terminal-status {
    flex-direction: column;
    gap: 8px;
    padding: 10px 12px;
  }

  .status-item {
    gap: 6px;
  }
}
</style>
