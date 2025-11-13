<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-800">Plugins Instalados</h1>
        <p class="text-gray-600 mt-1">Gestiona los plugins de tus servidores</p>
      </div>
      <el-button
        type="primary"
        :icon="ShoppingCart"
        @click="router.push('/marketplace')"
      >
        Buscar más plugins
      </el-button>
    </div>

    <!-- Server selector -->
    <el-card>
      <el-select
        v-model="selectedServerId"
        placeholder="Selecciona un servidor"
        class="w-full"
        :loading="serversStore.loading"
        @change="loadPlugins"
        size="large"
      >
        <el-option
          v-for="server in serversStore.servers"
          :key="server.id"
          :label="server.name"
          :value="server.id"
        >
          <div class="flex items-center justify-between">
            <span>{{ server.name }}</span>
            <el-tag
              :type="getStatusType(server.status)"
              size="small"
            >
              {{ getStatusText(server.status) }}
            </el-tag>
          </div>
        </el-option>
      </el-select>
    </el-card>

    <!-- Plugins list -->
    <el-card v-if="selectedServerId">
      <div v-if="marketplaceStore.installedPlugins.length > 0">
        <div class="mb-4 text-gray-600">
          {{ marketplaceStore.installedPlugins.length }} plugins instalados
        </div>

        <el-table
          v-loading="marketplaceStore.loading"
          :data="marketplaceStore.installedPlugins"
          style="width: 100%"
        >
          <el-table-column prop="name" label="Nombre" min-width="200">
            <template #default="{ row }">
              <div class="flex items-center">
                <el-icon class="mr-2" :size="20">
                  <Box />
                </el-icon>
                <span class="font-medium">{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column prop="version" label="Versión" width="150" />

          <el-table-column label="Estado" width="120">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'">
                {{ row.enabled ? 'Habilitado' : 'Deshabilitado' }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column prop="file_name" label="Archivo" min-width="200">
            <template #default="{ row }">
              <span class="text-sm text-gray-600 font-mono">{{ row.file_name }}</span>
            </template>
          </el-table-column>

          <el-table-column label="Fuente" width="120">
            <template #default="{ row }">
              <el-tag v-if="row.source" size="small">{{ row.source }}</el-tag>
              <span v-else class="text-gray-400">-</span>
            </template>
          </el-table-column>

          <el-table-column label="Acciones" width="200" fixed="right">
            <template #default="{ row }">
              <el-button-group>
                <el-button
                  v-if="row.source && row.plugin_id"
                  type="primary"
                  size="small"
                  :icon="Refresh"
                  @click="handleUpdate(row)"
                >
                  Actualizar
                </el-button>
                <el-button
                  type="danger"
                  size="small"
                  :icon="Delete"
                  @click="handleUninstall(row)"
                >
                  Desinstalar
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <el-empty
        v-else-if="!marketplaceStore.loading"
        description="No hay plugins instalados en este servidor"
      >
        <el-button type="primary" @click="router.push('/marketplace')">
          Buscar plugins
        </el-button>
      </el-empty>
    </el-card>

    <el-empty
      v-else
      description="Selecciona un servidor para ver sus plugins instalados"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useMarketplaceStore } from '@/stores/marketplace';
import { useServersStore } from '@/stores/servers';
import {
  ShoppingCart,
  Box,
  Refresh,
  Delete,
} from '@element-plus/icons-vue';
import { ElMessage, ElMessageBox } from 'element-plus';

const router = useRouter();
const marketplaceStore = useMarketplaceStore();
const serversStore = useServersStore();

const selectedServerId = ref('');

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

const loadPlugins = async () => {
  if (!selectedServerId.value) return;
  
  const server = serversStore.servers.find(s => s.id === selectedServerId.value);
  if (server && server.status !== 'stopped') {
    ElMessage.warning('El servidor debe estar detenido para gestionar plugins');
  }
  
  await marketplaceStore.fetchInstalledPlugins(selectedServerId.value);
};

const handleUpdate = async (plugin: any) => {
  const server = serversStore.servers.find(s => s.id === selectedServerId.value);
  if (!server) return;

  if (server.status !== 'stopped') {
    ElMessage.error('El servidor debe estar detenido para actualizar plugins');
    return;
  }

  try {
    await ElMessageBox.confirm(
      `¿Actualizar ${plugin.name} a la última versión?`,
      'Confirmar actualización',
      {
        confirmButtonText: 'Sí, actualizar',
        cancelButtonText: 'Cancelar',
        type: 'info',
      }
    );

    await marketplaceStore.updatePlugin(selectedServerId.value, {
      source: plugin.source,
      plugin_id: plugin.plugin_id,
    });
  } catch {
    // Usuario canceló
  }
};

const handleUninstall = async (plugin: any) => {
  const server = serversStore.servers.find(s => s.id === selectedServerId.value);
  if (!server) return;

  if (server.status !== 'stopped') {
    ElMessage.error('El servidor debe estar detenido para desinstalar plugins');
    return;
  }

  try {
    await ElMessageBox.confirm(
      `¿Estás seguro de desinstalar ${plugin.name}? Esta acción no se puede deshacer.`,
      'Confirmar desinstalación',
      {
        confirmButtonText: 'Sí, desinstalar',
        cancelButtonText: 'Cancelar',
        type: 'error',
      }
    );

    await marketplaceStore.uninstallPlugin(selectedServerId.value, plugin.name);
  } catch {
    // Usuario canceló
  }
};

onMounted(async () => {
  await serversStore.fetchServers();
  
  // Seleccionar el primer servidor por defecto
  if (serversStore.servers.length > 0 && serversStore.servers[0]) {
    selectedServerId.value = serversStore.servers[0].id;
    await loadPlugins();
  }
});
</script>
