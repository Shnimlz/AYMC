<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-800">Servidores</h1>
        <p class="text-gray-600 mt-1">Gestiona tus servidores de Minecraft</p>
      </div>
      <el-button
        type="primary"
        :icon="Plus"
        @click="router.push('/servers/create')"
      >
        Crear Servidor
      </el-button>
    </div>

    <!-- Filters -->
    <el-card>
      <div class="flex gap-4">
        <el-select
          v-model="statusFilter"
          placeholder="Estado"
          clearable
          @change="applyFilters"
          class="w-48"
        >
          <el-option label="Todos" value="" />
          <el-option label="Activo" value="running" />
          <el-option label="Detenido" value="stopped" />
          <el-option label="Error" value="error" />
        </el-select>

        <el-input
          v-model="searchQuery"
          placeholder="Buscar por nombre..."
          :prefix-icon="Search"
          clearable
          @input="applyFilters"
          class="w-64"
        />

        <el-button :icon="Refresh" @click="loadServers">
          Actualizar
        </el-button>
      </div>
    </el-card>

    <!-- Servers table -->
    <el-card>
      <el-table
        v-loading="serversStore.loading"
        :data="filteredServers"
        style="width: 100%"
        :empty-text="'No hay servidores'"
      >
        <el-table-column prop="name" label="Nombre" min-width="150">
          <template #default="{ row }">
            <div class="flex items-center">
              <el-icon class="mr-2" :size="20">
                <Monitor />
              </el-icon>
              <span class="font-medium">{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="type" label="Tipo" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ row.type }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="version" label="Versión" width="120" />

        <el-table-column prop="port" label="Puerto" width="100" />

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

        <el-table-column label="RAM" width="150">
          <template #default="{ row }">
            {{ row.ram_min }}MB - {{ row.ram_max }}MB
          </template>
        </el-table-column>

        <el-table-column label="Auto-inicio" width="100" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.auto_start" :size="20" color="#67C23A">
              <Check />
            </el-icon>
            <el-icon v-else :size="20" color="#909399">
              <Close />
            </el-icon>
          </template>
        </el-table-column>

        <el-table-column label="Acciones" width="280" fixed="right">
          <template #default="{ row }">
            <el-button-group>
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
                v-if="row.status === 'running'"
                type="warning"
                size="small"
                :icon="RefreshRight"
                @click="restartServer(row.id)"
              >
                Reiniciar
              </el-button>
              <el-button
                type="primary"
                size="small"
                :icon="View"
                @click="router.push(`/servers/${row.id}`)"
              >
                Ver
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useServersStore } from '@/stores/servers';
import {
  Plus,
  Search,
  Refresh,
  Monitor,
  VideoPlay,
  VideoPause,
  RefreshRight,
  View,
  Check,
  Close,
} from '@element-plus/icons-vue';

const router = useRouter();
const serversStore = useServersStore();

const statusFilter = ref('');
const searchQuery = ref('');

const filteredServers = computed(() => {
  let servers = serversStore.servers;

  // Filtrar por estado
  if (statusFilter.value) {
    servers = servers.filter(s => s.status === statusFilter.value);
  }

  // Filtrar por búsqueda
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    servers = servers.filter(s => 
      s.name.toLowerCase().includes(query) ||
      s.type.toLowerCase().includes(query) ||
      s.version.toLowerCase().includes(query)
    );
  }

  return servers;
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

const applyFilters = () => {
  // Los filtros se aplican automáticamente a través del computed
};

const loadServers = async () => {
  await serversStore.fetchServers();
};

const startServer = async (id: string) => {
  await serversStore.startServer(id);
  await loadServers();
};

const stopServer = async (id: string) => {
  await serversStore.stopServer(id);
  await loadServers();
};

const restartServer = async (id: string) => {
  await serversStore.restartServer(id);
  await loadServers();
};

onMounted(() => {
  loadServers();
});
</script>
