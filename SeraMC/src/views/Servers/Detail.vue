<template>
  <div v-if="server" class="space-y-6">
    <!-- Header -->
    <div>
      <el-button :icon="ArrowLeft" @click="router.push('/servers')" class="mb-4">
        Volver
      </el-button>
      
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
          <h1 class="text-2xl font-bold text-gray-800">{{ server.name }}</h1>
          <el-tag
            :type="getStatusType(server.status)"
            effect="dark"
            size="large"
          >
            {{ getStatusText(server.status) }}
          </el-tag>
        </div>

        <el-button-group>
          <el-button
            v-if="server.status === 'stopped'"
            type="success"
            :icon="VideoPlay"
            @click="startServer"
          >
            Iniciar
          </el-button>
          <el-button
            v-else-if="server.status === 'running'"
            type="danger"
            :icon="VideoPause"
            @click="stopServer"
          >
            Detener
          </el-button>
          <el-button
            v-if="server.status === 'running'"
            type="warning"
            :icon="RefreshRight"
            @click="restartServer"
          >
            Reiniciar
          </el-button>
          <el-button :icon="Setting" @click="editDialogVisible = true">
            Editar
          </el-button>
          <el-button type="danger" :icon="Delete" @click="confirmDelete">
            Eliminar
          </el-button>
        </el-button-group>
      </div>
    </div>

    <!-- Info cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <el-card shadow="hover">
        <div class="text-center">
          <p class="text-sm text-gray-600">Tipo</p>
          <p class="text-xl font-bold mt-1">{{ server.type }}</p>
        </div>
      </el-card>
      <el-card shadow="hover">
        <div class="text-center">
          <p class="text-sm text-gray-600">Versión</p>
          <p class="text-xl font-bold mt-1">{{ server.version }}</p>
        </div>
      </el-card>
      <el-card shadow="hover">
        <div class="text-center">
          <p class="text-sm text-gray-600">Puerto</p>
          <p class="text-xl font-bold mt-1">{{ server.port }}</p>
        </div>
      </el-card>
      <el-card shadow="hover">
        <div class="text-center">
          <p class="text-sm text-gray-600">RAM</p>
          <p class="text-xl font-bold mt-1">{{ server.ram_min }}-{{ server.ram_max }}MB</p>
        </div>
      </el-card>
    </div>

    <!-- Tabs -->
    <el-card>
      <el-tabs v-model="activeTab">
        <!-- Console Tab -->
        <el-tab-pane label="Consola" name="console">
          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <el-tag v-if="connected" type="success">Conectado</el-tag>
                <el-tag v-else type="danger">Desconectado</el-tag>
                <span class="text-sm text-gray-600">{{ messages.length }} mensajes</span>
              </div>
              <div class="flex gap-2">
                <el-button size="small" :icon="Refresh" @click="clearMessages">
                  Limpiar
                </el-button>
                <el-button
                  v-if="!connected"
                  size="small"
                  type="primary"
                  :icon="Connection"
                  @click="connectWebSocket"
                >
                  Conectar
                </el-button>
              </div>
            </div>

            <!-- Console output -->
            <div
              ref="consoleRef"
              class="bg-gray-900 text-green-400 p-4 rounded font-mono text-sm h-96 overflow-y-auto"
            >
              <div
                v-for="(msg, index) in messages"
                :key="index"
                class="whitespace-pre-wrap break-words"
              >
                <span class="text-gray-500">[{{ formatTime(msg.timestamp) }}]</span>
                <span :class="getMessageClass(msg.level)">{{ msg.message }}</span>
              </div>
              <div v-if="messages.length === 0" class="text-gray-600">
                No hay mensajes. {{ server.status === 'running' ? 'Esperando logs...' : 'Inicia el servidor para ver logs.' }}
              </div>
            </div>

            <!-- Command input -->
            <div v-if="server.status === 'running'" class="flex gap-2">
              <el-input
                v-model="command"
                placeholder="Ingresa un comando..."
                @keyup.enter="sendCommand"
              >
                <template #prepend>
                  <span class="font-mono">&gt;</span>
                </template>
              </el-input>
              <el-button type="primary" :icon="Promotion" @click="sendCommand">
                Enviar
              </el-button>
            </div>
          </div>
        </el-tab-pane>

        <!-- Plugins Tab -->
        <el-tab-pane label="Plugins" name="plugins">
          <div class="text-center py-8 text-gray-500">
            <p>Gestiona los plugins desde el</p>
            <el-button
              type="primary"
              :icon="ShoppingCart"
              @click="router.push('/marketplace/installed')"
              class="mt-4"
            >
              Marketplace
            </el-button>
          </div>
        </el-tab-pane>

        <!-- Backups Tab -->
        <el-tab-pane label="Respaldos" name="backups">
          <div class="text-center py-8 text-gray-500">
            <p>Gestiona los respaldos desde la sección de</p>
            <el-button
              type="primary"
              :icon="FolderOpened"
              @click="router.push('/backups')"
              class="mt-4"
            >
              Respaldos
            </el-button>
          </div>
        </el-tab-pane>

        <!-- Settings Tab -->
        <el-tab-pane label="Configuración" name="settings">
          <div class="space-y-4">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="ID">{{ server.id }}</el-descriptions-item>
              <el-descriptions-item label="Agente ID">{{ server.agent_id }}</el-descriptions-item>
              <el-descriptions-item label="Auto-inicio">
                <el-tag v-if="server.auto_start" type="success">Sí</el-tag>
                <el-tag v-else type="info">No</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="Java">v{{ server.java_version }}</el-descriptions-item>
              <el-descriptions-item label="Creado">
                {{ formatDate(server.created_at) }}
              </el-descriptions-item>
              <el-descriptions-item label="Actualizado">
                {{ formatDate(server.updated_at) }}
              </el-descriptions-item>
            </el-descriptions>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- Edit Dialog -->
    <el-dialog v-model="editDialogVisible" title="Editar Servidor" width="600px">
      <el-form :model="editForm" label-width="120px">
        <el-form-item label="Nombre">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="Puerto">
          <el-input-number v-model="editForm.port" :min="1024" :max="65535" />
        </el-form-item>
        <el-form-item label="RAM Mínima">
          <el-input-number v-model="editForm.ram_min" :min="512" :max="32768" :step="512" />
        </el-form-item>
        <el-form-item label="RAM Máxima">
          <el-input-number v-model="editForm.ram_max" :min="512" :max="32768" :step="512" />
        </el-form-item>
        <el-form-item label="Auto-inicio">
          <el-switch v-model="editForm.auto_start" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">Cancelar</el-button>
        <el-button type="primary" @click="handleUpdate">Guardar</el-button>
      </template>
    </el-dialog>
  </div>
  <div v-else class="text-center py-12">
    <el-empty description="Servidor no encontrado" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useServersStore } from '@/stores/servers';
import { useWebSocket } from '@/composables/useWebSocket';
import {
  ArrowLeft,
  VideoPlay,
  VideoPause,
  RefreshRight,
  Setting,
  Delete,
  Refresh,
  Connection,
  Promotion,
  ShoppingCart,
  FolderOpened,
} from '@element-plus/icons-vue';
import { ElMessageBox } from 'element-plus';
import dayjs from 'dayjs';

const route = useRoute();
const router = useRouter();
const serversStore = useServersStore();
const { connected, messages, connect, disconnect, subscribe, unsubscribe, sendCommand: wsSendCommand, clearMessages: wsClearMessages } = useWebSocket();

const serverId = route.params.id as string;
const server = ref(serversStore.selectedServer);
const activeTab = ref('console');
const editDialogVisible = ref(false);
const consoleRef = ref<HTMLDivElement>();
const command = ref('');

const editForm = reactive({
  name: '',
  port: 0,
  ram_min: 0,
  ram_max: 0,
  auto_start: false,
});

const getStatusType = (status: string) => {
  switch (status) {
    case 'running': return 'success';
    case 'stopped': return 'info';
    case 'starting':
    case 'stopping': return 'warning';
    case 'error': return 'danger';
    default: return 'info';
  }
};

const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    running: 'Activo',
    stopped: 'Detenido',
    starting: 'Iniciando',
    stopping: 'Deteniendo',
    error: 'Error',
  };
  return texts[status] || status;
};

const getMessageClass = (level?: string) => {
  switch (level) {
    case 'error': return 'text-red-400';
    case 'warn': return 'text-yellow-400';
    case 'info': return 'text-blue-400';
    default: return 'text-green-400';
  }
};

const formatTime = (timestamp: string) => {
  return dayjs(timestamp).format('HH:mm:ss');
};

const formatDate = (date: string) => {
  return dayjs(date).format('DD/MM/YYYY HH:mm');
};

const connectWebSocket = () => {
  connect();
  if (server.value?.status === 'running') {
    setTimeout(() => subscribe(serverId), 500);
  }
};

const clearMessages = () => {
  wsClearMessages();
};

const sendCommand = () => {
  if (!command.value.trim()) return;
  wsSendCommand(serverId, command.value);
  command.value = '';
};

const startServer = async () => {
  await serversStore.startServer(serverId);
  await loadServer();
  if (connected.value) {
    setTimeout(() => subscribe(serverId), 1000);
  }
};

const stopServer = async () => {
  if (connected.value) {
    unsubscribe(serverId);
  }
  await serversStore.stopServer(serverId);
  await loadServer();
};

const restartServer = async () => {
  await serversStore.restartServer(serverId);
  await loadServer();
};

const confirmDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `¿Estás seguro de eliminar el servidor "${server.value?.name}"? Esta acción no se puede deshacer.`,
      'Confirmar eliminación',
      {
        confirmButtonText: 'Sí, eliminar',
        cancelButtonText: 'Cancelar',
        type: 'error',
      }
    );
    
    const success = await serversStore.deleteServer(serverId);
    if (success) {
      router.push('/servers');
    }
  } catch {
    // Usuario canceló
  }
};

const handleUpdate = async () => {
  const success = await serversStore.updateServer(serverId, editForm);
  if (success) {
    editDialogVisible.value = false;
    await loadServer();
  }
};

const loadServer = async () => {
  const data = await serversStore.fetchServer(serverId);
  if (data) {
    server.value = data;
    // Actualizar form de edición
    Object.assign(editForm, {
      name: data.name,
      port: data.port,
      ram_min: data.ram_min,
      ram_max: data.ram_max,
      auto_start: data.auto_start,
    });
  }
};

// Auto-scroll console
watch(messages, async () => {
  await nextTick();
  if (consoleRef.value) {
    consoleRef.value.scrollTop = consoleRef.value.scrollHeight;
  }
});

onMounted(async () => {
  await loadServer();
  
  if (server.value?.status === 'running') {
    connectWebSocket();
  }
});

onUnmounted(() => {
  if (connected.value) {
    unsubscribe(serverId);
    disconnect();
  }
});
</script>
