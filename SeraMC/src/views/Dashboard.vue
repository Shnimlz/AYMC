<template>
  <div class="space-y-6">
    <!-- Stats cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <el-card shadow="hover" class="bg-gradient-to-br from-blue-500 to-blue-600 text-white">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm opacity-90">Total Servidores</p>
            <p class="text-3xl font-bold mt-1">{{ stats.totalServers }}</p>
          </div>
          <el-icon :size="48" class="opacity-80">
            <Monitor />
          </el-icon>
        </div>
      </el-card>

      <el-card shadow="hover" class="bg-gradient-to-br from-green-500 to-green-600 text-white">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm opacity-90">Servidores Activos</p>
            <p class="text-3xl font-bold mt-1">{{ stats.runningServers }}</p>
          </div>
          <el-icon :size="48" class="opacity-80">
            <VideoPlay />
          </el-icon>
        </div>
      </el-card>

      <el-card shadow="hover" class="bg-gradient-to-br from-purple-500 to-purple-600 text-white">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm opacity-90">Agentes Online</p>
            <p class="text-3xl font-bold mt-1">{{ stats.agentsOnline }}</p>
          </div>
          <el-icon :size="48" class="opacity-80">
            <Connection />
          </el-icon>
        </div>
      </el-card>

      <el-card shadow="hover" class="bg-gradient-to-br from-orange-500 to-orange-600 text-white">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm opacity-90">Total Respaldos</p>
            <p class="text-3xl font-bold mt-1">{{ stats.totalBackups }}</p>
          </div>
          <el-icon :size="48" class="opacity-80">
            <FolderOpened />
          </el-icon>
        </div>
      </el-card>
    </div>

    <!-- Quick actions -->
    <el-card>
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-lg font-semibold">Acciones Rápidas</span>
        </div>
      </template>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <el-button
          type="primary"
          size="large"
          :icon="Plus"
          @click="router.push('/servers/create')"
        >
          Crear Servidor
        </el-button>

        <el-button
          type="success"
          size="large"
          :icon="Search"
          @click="router.push('/marketplace')"
        >
          Buscar Plugins
        </el-button>

        <el-button
          type="warning"
          size="large"
          :icon="DocumentCopy"
          @click="createBackup"
        >
          Crear Respaldo
        </el-button>
      </div>
    </el-card>

    <!-- Servers list -->
    <el-card>
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-lg font-semibold">Mis Servidores</span>
          <el-button
            type="primary"
            text
            :icon="Refresh"
            @click="loadData"
          >
            Actualizar
          </el-button>
        </div>
      </template>

      <el-table
        v-loading="serversStore.loading"
        :data="serversStore.servers"
        style="width: 100%"
      >
        <el-table-column prop="name" label="Nombre" min-width="150" />
        <el-table-column prop="type" label="Tipo" width="120" />
        <el-table-column prop="version" label="Versión" width="120" />
        <el-table-column label="Estado" width="120">
          <template #default="{ row }">
            <el-tag
              :type="getStatusType(row.status)"
              effect="dark"
            >
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Acciones" width="250" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'stopped'"
              type="success"
              size="small"
              :icon="VideoPlay"
              @click="startServer(row.id)"
            >
              Iniciar
            </el-button>
            <el-button
              v-else-if="row.status === 'running'"
              type="danger"
              size="small"
              :icon="VideoPause"
              @click="stopServer(row.id)"
            >
              Detener
            </el-button>
            <el-button
              type="primary"
              size="small"
              :icon="View"
              @click="router.push(`/servers/${row.id}`)"
            >
              Ver
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useServersStore } from '@/stores/servers';
import {
  Monitor,
  VideoPlay,
  VideoPause,
  Connection,
  FolderOpened,
  Plus,
  Search,
  DocumentCopy,
  Refresh,
  View,
} from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';

const router = useRouter();
const serversStore = useServersStore();

const stats = ref({
  totalServers: 0,
  runningServers: 0,
  agentsOnline: 0,
  totalBackups: 0,
});

const getStatusType = (status: string) => {
  switch (status) {
    case 'running':
      return 'success';
    case 'stopped':
      return 'info';
    case 'starting':
    case 'stopping':
      return 'warning';
    case 'error':
      return 'danger';
    default:
      return 'info';
  }
};

const getStatusText = (status: string) => {
  switch (status) {
    case 'running':
      return 'Activo';
    case 'stopped':
      return 'Detenido';
    case 'starting':
      return 'Iniciando';
    case 'stopping':
      return 'Deteniendo';
    case 'error':
      return 'Error';
    default:
      return status;
  }
};

const loadData = async () => {
  await serversStore.fetchServers();
  
  // Calcular estadísticas
  stats.value.totalServers = serversStore.servers.length;
  stats.value.runningServers = serversStore.servers.filter(s => s.status === 'running').length;
  
  // TODO: Cargar stats de agentes y backups desde API
  stats.value.agentsOnline = 0;
  stats.value.totalBackups = 0;
};

const startServer = async (id: string) => {
  const success = await serversStore.startServer(id);
  if (success) {
    await loadData();
  }
};

const stopServer = async (id: string) => {
  const success = await serversStore.stopServer(id);
  if (success) {
    await loadData();
  }
};

const createBackup = () => {
  ElMessage.info('Selecciona un servidor primero en la vista de Respaldos');
  router.push('/backups');
};

onMounted(() => {
  loadData();
});
</script>
