import { defineStore } from "pinia";
import { ref } from "vue";
import { backupsAPI } from "@/api";
import { ElMessage } from "element-plus";

export interface Backup {
  id: string;
  server_id: string;
  filename: string;
  file_path: string;
  size: number;
  type: "manual" | "scheduled";
  status: "pending" | "in_progress" | "completed" | "failed";
  created_at: string;
  completed_at?: string;
}

export interface BackupConfig {
  id: string;
  server_id: string;
  enabled: boolean;
  schedule: string;
  max_backups: number;
  retention_days: number;
  include_world: boolean;
  include_plugins: boolean;
  include_config: boolean;
  include_logs: boolean;
  exclude_paths: string[];
  created_at: string;
  updated_at: string;
}

export interface BackupStats {
  total_backups: number;
  total_size: number;
  oldest_backup: string;
  newest_backup: string;
  last_backup_date?: string;
  next_scheduled_backup?: string;
}

export const useBackupsStore = defineStore("backups", () => {
  // State
  const backups = ref<Backup[]>([]);
  const selectedBackup = ref<Backup | null>(null);
  const config = ref<BackupConfig | null>(null);
  const stats = ref<BackupStats | null>(null);
  const loading = ref(false);

  // Actions
  async function fetchBackups(serverId: string) {
    try {
      loading.value = true;
      const response = await backupsAPI.list(serverId);
      backups.value = response.data;
      return true;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al cargar respaldos";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function fetchBackup(backupId: string) {
    try {
      loading.value = true;
      const response = await backupsAPI.get(backupId);
      selectedBackup.value = response.data;
      return response.data;
    } catch (error: any) {
      const message = error.response?.data?.error || "Error al cargar respaldo";
      ElMessage.error(message);
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function createBackup(serverId: string, data: any) {
    try {
      loading.value = true;
      const response = await backupsAPI.create(serverId, data);

      // Añadir a la lista
      backups.value.unshift(response.data);

      ElMessage.success("Respaldo creado correctamente");
      return response.data;
    } catch (error: any) {
      const message = error.response?.data?.error || "Error al crear respaldo";
      ElMessage.error(message);
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function createManualBackup(serverId: string) {
    try {
      loading.value = true;
      const response = await backupsAPI.createManual(serverId);

      // Añadir a la lista
      backups.value.unshift(response.data);

      ElMessage.success("Respaldo manual iniciado");
      return response.data;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al crear respaldo manual";
      ElMessage.error(message);
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function deleteBackup(backupId: string) {
    try {
      loading.value = true;
      await backupsAPI.delete(backupId);

      // Eliminar de la lista
      backups.value = backups.value.filter((b) => b.id !== backupId);

      ElMessage.success("Respaldo eliminado correctamente");
      return true;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al eliminar respaldo";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function restoreBackup(backupId: string, data: any) {
    try {
      loading.value = true;
      await backupsAPI.restore(backupId, data);
      ElMessage.success("Restauración iniciada. El servidor se reiniciará.");
      return true;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al restaurar respaldo";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function fetchConfig(serverId: string) {
    try {
      loading.value = true;
      const response = await backupsAPI.getConfig(serverId);
      config.value = response.data;
      return response.data;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al cargar configuración";
      ElMessage.error(message);
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function updateConfig(serverId: string, data: Partial<BackupConfig>) {
    try {
      loading.value = true;
      const response = await backupsAPI.updateConfig(serverId, data);
      config.value = response.data;
      ElMessage.success("Configuración actualizada correctamente");
      return response.data;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al actualizar configuración";
      ElMessage.error(message);
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function fetchStats(serverId: string) {
    try {
      const response = await backupsAPI.getStats(serverId);
      stats.value = response.data;
      return response.data;
    } catch (error: any) {
      console.error("Error al cargar estadísticas:", error);
      return null;
    }
  }

  return {
    // State
    backups,
    selectedBackup,
    config,
    stats,
    loading,

    // Actions
    fetchBackups,
    fetchBackup,
    createBackup,
    createManualBackup,
    deleteBackup,
    restoreBackup,
    fetchConfig,
    updateConfig,
    fetchStats,
  };
});
