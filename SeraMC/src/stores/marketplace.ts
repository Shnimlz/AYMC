import { defineStore } from "pinia";
import { ref } from "vue";
import { marketplaceAPI } from "@/api";
import { ElMessage } from "element-plus";

export interface Plugin {
  id: string;
  name: string;
  slug: string;
  description: string;
  author: string;
  downloads: number;
  rating: number;
  icon_url?: string;
  source: "modrinth" | "spigot";
  latest_version?: string;
  categories?: string[];
}

export interface PluginDetail extends Plugin {
  long_description: string;
  website?: string;
  source_code?: string;
  issues?: string;
  license?: string;
  versions: PluginVersion[];
}

export interface PluginVersion {
  id: string;
  version_number: string;
  minecraft_versions: string[];
  release_date: string;
  downloads: number;
  file_url: string;
}

export interface InstalledPlugin {
  name: string;
  version: string;
  enabled: boolean;
  file_name: string;
  source?: string;
  plugin_id?: string;
}

export const useMarketplaceStore = defineStore("marketplace", () => {
  // State
  const searchResults = ref<Plugin[]>([]);
  const selectedPlugin = ref<PluginDetail | null>(null);
  const installedPlugins = ref<InstalledPlugin[]>([]);
  const loading = ref(false);
  const searchQuery = ref("");
  const searchSource = ref<"all" | "modrinth" | "spigot">("all");

  // Actions
  async function searchPlugins(query: string, source?: string) {
    try {
      loading.value = true;
      searchQuery.value = query;
      if (source) searchSource.value = source as any;

      const params: any = { query };
      if (source && source !== "all") {
        params.source = source;
      }

      const response = await marketplaceAPI.search(params);
      searchResults.value = response.data;
      return true;
    } catch (error: any) {
      const message = error.response?.data?.error || "Error al buscar plugins";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function fetchPluginDetail(source: string, id: string) {
    try {
      loading.value = true;
      const response = await marketplaceAPI.getPlugin(source, id);
      selectedPlugin.value = response.data;
      return response.data;
    } catch (error: any) {
      const message = error.response?.data?.error || "Error al cargar plugin";
      ElMessage.error(message);
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function fetchInstalledPlugins(serverId: string) {
    try {
      loading.value = true;
      const response = await marketplaceAPI.listInstalledPlugins(serverId);
      installedPlugins.value = response.data;
      return response.data;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al cargar plugins instalados";
      ElMessage.error(message);
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function installPlugin(serverId: string, data: any) {
    try {
      loading.value = true;
      await marketplaceAPI.installPlugin(serverId, data);
      ElMessage.success("Plugin instalado correctamente");

      // Recargar plugins instalados
      await fetchInstalledPlugins(serverId);
      return true;
    } catch (error: any) {
      const message = error.response?.data?.error || "Error al instalar plugin";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function uninstallPlugin(serverId: string, pluginName: string) {
    try {
      loading.value = true;
      await marketplaceAPI.uninstallPlugin(serverId, {
        plugin_name: pluginName,
      });
      ElMessage.success("Plugin desinstalado correctamente");

      // Recargar plugins instalados
      await fetchInstalledPlugins(serverId);
      return true;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al desinstalar plugin";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function updatePlugin(serverId: string, data: any) {
    try {
      loading.value = true;
      await marketplaceAPI.updatePlugin(serverId, data);
      ElMessage.success("Plugin actualizado correctamente");

      // Recargar plugins instalados
      await fetchInstalledPlugins(serverId);
      return true;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al actualizar plugin";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  return {
    // State
    searchResults,
    selectedPlugin,
    installedPlugins,
    loading,
    searchQuery,
    searchSource,

    // Actions
    searchPlugins,
    fetchPluginDetail,
    fetchInstalledPlugins,
    installPlugin,
    uninstallPlugin,
    updatePlugin,
  };
});
