import { defineStore } from "pinia";
import { ref } from "vue";
import { serversAPI } from "@/api";
import { ElMessage } from "element-plus";

export interface Server {
  id: string;
  name: string;
  type: string;
  version: string;
  port: number;
  status: "stopped" | "running" | "starting" | "stopping" | "error";
  agent_id: string;
  auto_start: boolean;
  ram_min: number;
  ram_max: number;
  java_version: string;
  created_at: string;
  updated_at: string;
}

export interface ServerCreate {
  name: string;
  type: string;
  version: string;
  port: number;
  agent_id: string;
  auto_start?: boolean;
  ram_min?: number;
  ram_max?: number;
  java_version?: string;
}

export const useServersStore = defineStore("servers", () => {
  // State
  const servers = ref<Server[]>([]);
  const selectedServer = ref<Server | null>(null);
  const loading = ref(false);

  // Actions
  async function fetchServers() {
    try {
      loading.value = true;
      const response = await serversAPI.list();
      servers.value = response.data;
      return true;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al cargar servidores";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function fetchServer(id: string) {
    try {
      loading.value = true;
      const response = await serversAPI.get(id);
      selectedServer.value = response.data;
      return response.data;
    } catch (error: any) {
      const message = error.response?.data?.error || "Error al cargar servidor";
      ElMessage.error(message);
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function createServer(data: ServerCreate) {
    try {
      loading.value = true;
      const response = await serversAPI.create(data);

      // Añadir a la lista
      servers.value.push(response.data);

      ElMessage.success("Servidor creado correctamente");
      return response.data;
    } catch (error: any) {
      const message = error.response?.data?.error || "Error al crear servidor";
      ElMessage.error(message);
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function updateServer(id: string, data: Partial<ServerCreate>) {
    try {
      loading.value = true;
      const response = await serversAPI.update(id, data);

      // Actualizar en la lista
      const index = servers.value.findIndex((s) => s.id === id);
      if (index !== -1) {
        servers.value[index] = response.data;
      }

      // Actualizar seleccionado si es el mismo
      if (selectedServer.value?.id === id) {
        selectedServer.value = response.data;
      }

      ElMessage.success("Servidor actualizado correctamente");
      return response.data;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al actualizar servidor";
      ElMessage.error(message);
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function deleteServer(id: string) {
    try {
      loading.value = true;
      await serversAPI.delete(id);

      // Eliminar de la lista
      servers.value = servers.value.filter((s) => s.id !== id);

      // Limpiar seleccionado si es el mismo
      if (selectedServer.value?.id === id) {
        selectedServer.value = null;
      }

      ElMessage.success("Servidor eliminado correctamente");
      return true;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al eliminar servidor";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function startServer(id: string) {
    try {
      loading.value = true;
      await serversAPI.start(id);

      // Actualizar estado
      await updateServerStatus(id, "starting");

      ElMessage.success("Servidor iniciando...");
      return true;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al iniciar servidor";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function stopServer(id: string) {
    try {
      loading.value = true;
      await serversAPI.stop(id);

      // Actualizar estado
      await updateServerStatus(id, "stopping");

      ElMessage.success("Servidor deteniéndose...");
      return true;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al detener servidor";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function restartServer(id: string) {
    try {
      loading.value = true;
      await serversAPI.restart(id);

      // Actualizar estado
      await updateServerStatus(id, "starting");

      ElMessage.success("Servidor reiniciando...");
      return true;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al reiniciar servidor";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function getServerStatus(id: string) {
    try {
      const response = await serversAPI.getStatus(id);

      // Actualizar en la lista
      const index = servers.value.findIndex((s) => s.id === id);
      if (index !== -1 && response.data.status) {
        servers.value[index].status = response.data.status;
      }

      // Actualizar seleccionado si es el mismo
      if (selectedServer.value?.id === id && response.data.status) {
        selectedServer.value.status = response.data.status;
      }

      return response.data;
    } catch (error: any) {
      console.error("Error al obtener estado del servidor:", error);
      return null;
    }
  }

  function updateServerStatus(id: string, status: Server["status"]) {
    // Actualizar en la lista
    const index = servers.value.findIndex((s) => s.id === id);
    if (index !== -1) {
      servers.value[index].status = status;
    }

    // Actualizar seleccionado si es el mismo
    if (selectedServer.value?.id === id) {
      selectedServer.value.status = status;
    }
  }

  function selectServer(server: Server | null) {
    selectedServer.value = server;
  }

  return {
    // State
    servers,
    selectedServer,
    loading,

    // Actions
    fetchServers,
    fetchServer,
    createServer,
    updateServer,
    deleteServer,
    startServer,
    stopServer,
    restartServer,
    getServerStatus,
    updateServerStatus,
    selectServer,
  };
});
