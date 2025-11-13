<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-800">Respaldos</h1>
        <p class="text-gray-600 mt-1">Gestiona los respaldos de tus servidores</p>
      </div>
      <el-button
        type="primary"
        :icon="Setting"
        @click="router.push('/backups/config')"
      >
        Configuración
      </el-button>
    </div>

    <!-- Server selector -->
    <el-card>
      <div class="flex gap-4">
        <el-select
          v-model="selectedServerId"
          placeholder="Selecciona un servidor"
          class="flex-1"
          :loading="serversStore.loading"
          @change="loadBackups"
          size="large"
        >
          <el-option
            v-for="server in serversStore.servers"
            :key="server.id"
            :label="server.name"
            :value="server.id"
          />
        </el-select>

        <el-button
          type="success"
          size="large"
          :icon="Plus"
          :loading="backupsStore.loading"
          :disabled="!selectedServerId"
          @click="createManualBackup"
        >
          Crear Respaldo Manual
        </el-button>

        <el-button
          size="large"
          :icon="Refresh"
          :disabled="!selectedServerId"
          @click="loadBackups"
        >
          Actualizar
        </el-button>
      </div>
    </el-card>

    <!-- Stats -->
    <div v-if="selectedServerId && backupsStore.stats" class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <el-card shadow="hover">
        <div class="text-center">
          <p class="text-sm text-gray-600">Total Respaldos</p>
          <p class="text-2xl font-bold mt-1">{{ backupsStore.stats.total_backups }}</p>
        </div>
      </el-card>
      <el-card shadow="hover">
        <div class="text-center">
          <p class="text-sm text-gray-600">Tamaño Total</p>
          <p class="text-2xl font-bold mt-1">{{ formatBytes(backupsStore.stats.total_size) }}</p>
        </div>
      </el-card>
      <el-card shadow="hover">
        <div class="text-center">
          <p class="text-sm text-gray-600">Último Respaldo</p>
          <p class="text-lg font-bold mt-1">
            {{ backupsStore.stats.last_backup_date ? formatDate(backupsStore.stats.last_backup_date) : 'N/A' }}
          </p>
        </div>
      </el-card>
      <el-card shadow="hover">
        <div class="text-center">
          <p class="text-sm text-gray-600">Próximo Programado</p>
          <p class="text-lg font-bold mt-1">
            {{ backupsStore.stats.next_scheduled_backup ? formatDate(backupsStore.stats.next_scheduled_backup) : 'N/A' }}
          </p>
        </div>
      </el-card>
    </div>

    <!-- Backups table -->
    <el-card v-if="selectedServerId">
      <el-table
        v-loading="backupsStore.loading"
        :data="backupsStore.backups"
        style="width: 100%"
        :empty-text="'No hay respaldos'"
      >
        <el-table-column prop="filename" label="Nombre" min-width="250">
          <template #default="{ row }">
            <div class="flex items-center">
              <el-icon class="mr-2" :size="20">
                <Document />
              </el-icon>
              <span class="font-mono text-sm">{{ row.filename }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="Tipo" width="120">
          <template #default="{ row }">
            <el-tag :type="row.type === 'manual' ? 'primary' : 'info'" size="small">
              {{ row.type === 'manual' ? 'Manual' : 'Programado' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="Tamaño" width="120">
          <template #default="{ row }">
            {{ formatBytes(row.size) }}
          </template>
        </el-table-column>

        <el-table-column label="Estado" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="Fecha" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column label="Acciones" width="200" fixed="right">
          <template #default="{ row }">
            <el-button-group v-if="row.status === 'completed'">
              <el-button
                type="success"
                size="small"
                :icon="RefreshLeft"
                @click="handleRestore(row)"
              >
                Restaurar
              </el-button>
              <el-button
                type="danger"
                size="small"
                :icon="Delete"
                @click="handleDelete(row.id)"
              >
                Eliminar
              </el-button>
            </el-button-group>
            <el-tag v-else-if="row.status === 'in_progress'" type="warning">
              En progreso...
            </el-tag>
            <el-tag v-else-if="row.status === 'failed'" type="danger">
              Fallido
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-empty
      v-else
      description="Selecciona un servidor para ver sus respaldos"
    />

    <!-- Restore Dialog -->
    <el-dialog v-model="restoreDialogVisible" title="Restaurar Respaldo" width="600px">
      <el-alert
        type="warning"
        :closable="false"
        show-icon
        class="mb-4"
      >
        <template #title>
          ⚠️ Advertencia
        </template>
        Esta acción reemplazará todos los archivos del servidor con los del respaldo.
        El servidor se reiniciará automáticamente.
      </el-alert>

      <el-form label-width="180px">
        <el-form-item label="Respaldo">
          <span class="font-mono">{{ selectedBackup?.filename }}</span>
        </el-form-item>

        <el-form-item label="Tamaño">
          <span>{{ selectedBackup ? formatBytes(selectedBackup.size) : '' }}</span>
        </el-form-item>

        <el-form-item label="Fecha de creación">
          <span>{{ selectedBackup ? formatDate(selectedBackup.created_at) : '' }}</span>
        </el-form-item>

        <el-divider />

        <el-form-item label="Restaurar mundo">
          <el-switch v-model="restoreOptions.restore_world" />
        </el-form-item>

        <el-form-item label="Restaurar plugins">
          <el-switch v-model="restoreOptions.restore_plugins" />
        </el-form-item>

        <el-form-item label="Restaurar config">
          <el-switch v-model="restoreOptions.restore_config" />
        </el-form-item>

        <el-form-item label="Restaurar logs">
          <el-switch v-model="restoreOptions.restore_logs" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="restoreDialogVisible = false">Cancelar</el-button>
        <el-button
          type="danger"
          :loading="backupsStore.loading"
          @click="confirmRestore"
        >
          Restaurar Ahora
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useBackupsStore } from '@/stores/backups';
import { useServersStore } from '@/stores/servers';
import {
  Setting,
  Plus,
  Refresh,
  Document,
  RefreshLeft,
  Delete,
} from '@element-plus/icons-vue';
import { ElMessageBox } from 'element-plus';
import dayjs from 'dayjs';
import type { Backup } from '@/stores/backups';

const router = useRouter();
const backupsStore = useBackupsStore();
const serversStore = useServersStore();

const selectedServerId = ref('');
const restoreDialogVisible = ref(false);
const selectedBackup = ref<Backup | null>(null);
const restoreOptions = reactive({
  restore_world: true,
  restore_plugins: true,
  restore_config: true,
  restore_logs: false,
});

const getStatusType = (status: string) => {
  switch (status) {
    case 'completed': return 'success';
    case 'in_progress': return 'warning';
    case 'failed': return 'danger';
    default: return 'info';
  }
};

const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    completed: 'Completado',
    in_progress: 'En progreso',
    failed: 'Fallido',
    pending: 'Pendiente',
  };
  return texts[status] || status;
};

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
};

const formatDate = (date: string) => {
  return dayjs(date).format('DD/MM/YYYY HH:mm');
};

const loadBackups = async () => {
  if (!selectedServerId.value) return;
  await backupsStore.fetchBackups(selectedServerId.value);
  await backupsStore.fetchStats(selectedServerId.value);
};

const createManualBackup = async () => {
  if (!selectedServerId.value) return;
  
  const server = serversStore.servers.find(s => s.id === selectedServerId.value);
  if (!server) return;

  try {
    await ElMessageBox.confirm(
      `¿Crear un respaldo manual de "${server.name}"?`,
      'Confirmar',
      {
        confirmButtonText: 'Sí, crear',
        cancelButtonText: 'Cancelar',
        type: 'info',
      }
    );

    await backupsStore.createManualBackup(selectedServerId.value);
    await loadBackups();
  } catch {
    // Usuario canceló
  }
};

const handleRestore = (backup: Backup) => {
  selectedBackup.value = backup;
  restoreDialogVisible.value = true;
};

const confirmRestore = async () => {
  if (!selectedBackup.value) return;

  const success = await backupsStore.restoreBackup(
    selectedBackup.value.id,
    restoreOptions
  );

  if (success) {
    restoreDialogVisible.value = false;
    await loadBackups();
  }
};

const handleDelete = async (backupId: string) => {
  try {
    await ElMessageBox.confirm(
      '¿Estás seguro de eliminar este respaldo? Esta acción no se puede deshacer.',
      'Confirmar eliminación',
      {
        confirmButtonText: 'Sí, eliminar',
        cancelButtonText: 'Cancelar',
        type: 'error',
      }
    );

    await backupsStore.deleteBackup(backupId);
  } catch {
    // Usuario canceló
  }
};

onMounted(async () => {
  await serversStore.fetchServers();
  
  // Seleccionar el primer servidor por defecto
  if (serversStore.servers.length > 0) {
    selectedServerId.value = serversStore.servers[0].id;
    await loadBackups();
  }
});
</script>
