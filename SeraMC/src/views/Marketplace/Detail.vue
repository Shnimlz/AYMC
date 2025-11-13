<template>
  <div v-if="plugin" class="space-y-6">
    <!-- Header -->
    <div>
      <el-button :icon="ArrowLeft" @click="router.back()" class="mb-4">
        Volver
      </el-button>

      <div class="flex gap-6">
        <el-avatar :size="120" :src="plugin.icon_url" :icon="Box" />
        
        <div class="flex-1">
          <div class="flex items-start justify-between">
            <div>
              <h1 class="text-3xl font-bold text-gray-800">{{ plugin.name }}</h1>
              <p class="text-gray-600 mt-1">
                por <span class="font-medium">{{ plugin.author }}</span>
              </p>
            </div>
            <el-tag size="large">{{ plugin.source }}</el-tag>
          </div>

          <div class="flex items-center gap-6 mt-4 text-gray-600">
            <div class="flex items-center gap-2">
              <el-icon :size="20"><Download /></el-icon>
              <span class="font-medium">{{ formatNumber(plugin.downloads) }}</span>
              <span class="text-sm">descargas</span>
            </div>
            <div v-if="plugin.rating" class="flex items-center gap-2">
              <el-icon :size="20"><Star /></el-icon>
              <span class="font-medium">{{ plugin.rating.toFixed(1) }}</span>
              <span class="text-sm">rating</span>
            </div>
            <div v-if="plugin.latest_version" class="flex items-center gap-2">
              <el-icon :size="20"><PriceTag /></el-icon>
              <span class="font-medium">{{ plugin.latest_version }}</span>
              <span class="text-sm">versión</span>
            </div>
          </div>

          <div class="flex gap-2 mt-4">
            <el-button
              v-if="plugin.website"
              :icon="Link"
              @click="openLink(plugin.website)"
            >
              Sitio Web
            </el-button>
            <el-button
              v-if="plugin.source_code"
              :icon="Link"
              @click="openLink(plugin.source_code)"
            >
              Código Fuente
            </el-button>
            <el-button
              v-if="plugin.issues"
              :icon="Link"
              @click="openLink(plugin.issues)"
            >
              Reportar Bug
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- Description -->
    <el-card>
      <template #header>
        <span class="font-semibold">Descripción</span>
      </template>
      <div class="prose max-w-none" v-html="formatDescription(plugin.long_description || plugin.description)"></div>
    </el-card>

    <!-- Categories -->
    <el-card v-if="plugin.categories && plugin.categories.length > 0">
      <template #header>
        <span class="font-semibold">Categorías</span>
      </template>
      <div class="flex flex-wrap gap-2">
        <el-tag
          v-for="category in plugin.categories"
          :key="category"
          type="info"
        >
          {{ category }}
        </el-tag>
      </div>
    </el-card>

    <!-- Versions & Install -->
    <el-card>
      <template #header>
        <span class="font-semibold">Instalación</span>
      </template>

      <el-form label-width="120px">
        <el-form-item label="Servidor">
          <el-select
            v-model="selectedServerId"
            placeholder="Selecciona un servidor"
            class="w-full"
            :loading="serversStore.loading"
          >
            <el-option
              v-for="server in serversStore.servers"
              :key="server.id"
              :label="server.name"
              :value="server.id"
              :disabled="server.status !== 'stopped'"
            >
              <div class="flex items-center justify-between">
                <span>{{ server.name }}</span>
                <el-tag
                  :type="server.status === 'stopped' ? 'success' : 'warning'"
                  size="small"
                >
                  {{ server.status }}
                </el-tag>
              </div>
            </el-option>
          </el-select>
          <div class="text-sm text-gray-500 mt-1">
            El servidor debe estar detenido para instalar plugins
          </div>
        </el-form-item>

        <el-form-item label="Versión" v-if="plugin.versions && plugin.versions.length > 0">
          <el-select
            v-model="selectedVersionId"
            placeholder="Selecciona una versión"
            class="w-full"
          >
            <el-option
              v-for="version in plugin.versions"
              :key="version.id"
              :label="version.version_number"
              :value="version.id"
            >
              <div class="flex items-center justify-between">
                <span>{{ version.version_number }}</span>
                <div class="flex items-center gap-2 text-xs text-gray-500">
                  <span>{{ version.minecraft_versions.join(', ') }}</span>
                  <span>•</span>
                  <span>{{ formatNumber(version.downloads) }} descargas</span>
                </div>
              </div>
            </el-option>
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :icon="Download"
            :loading="marketplaceStore.loading"
            :disabled="!selectedServerId || !selectedVersionId"
            @click="handleInstall"
          >
            Instalar Plugin
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- License -->
    <el-card v-if="plugin.license">
      <template #header>
        <span class="font-semibold">Licencia</span>
      </template>
      <el-tag>{{ plugin.license }}</el-tag>
    </el-card>
  </div>

  <div v-else-if="!marketplaceStore.loading" class="text-center py-12">
    <el-empty description="Plugin no encontrado" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useMarketplaceStore } from '@/stores/marketplace';
import { useServersStore } from '@/stores/servers';
import {
  ArrowLeft,
  Box,
  Download,
  Star,
  PriceTag,
  Link,
} from '@element-plus/icons-vue';
import { ElMessage, ElMessageBox } from 'element-plus';

const route = useRoute();
const router = useRouter();
const marketplaceStore = useMarketplaceStore();
const serversStore = useServersStore();

const source = route.params.source as string;
const pluginId = route.params.id as string;

const plugin = ref(marketplaceStore.selectedPlugin);
const selectedServerId = ref('');
const selectedVersionId = ref('');

const formatNumber = (num: number) => {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M';
  } else if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K';
  }
  return num.toString();
};

const formatDescription = (desc: string) => {
  // Convertir markdown simple a HTML
  return desc
    .replace(/\n/g, '<br>')
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>');
};

const openLink = (url: string) => {
  window.open(url, '_blank');
};

const handleInstall = async () => {
  if (!selectedServerId.value || !selectedVersionId.value) {
    ElMessage.warning('Por favor selecciona un servidor y una versión');
    return;
  }

  const server = serversStore.servers.find(s => s.id === selectedServerId.value);
  if (!server) return;

  try {
    await ElMessageBox.confirm(
      `¿Instalar ${plugin.value?.name} en ${server.name}?`,
      'Confirmar instalación',
      {
        confirmButtonText: 'Sí, instalar',
        cancelButtonText: 'Cancelar',
        type: 'info',
      }
    );

    const success = await marketplaceStore.installPlugin(selectedServerId.value, {
      source: source,
      plugin_id: pluginId,
      version_id: selectedVersionId.value,
    });

    if (success) {
      ElMessageBox.alert(
        'El plugin se instaló correctamente. Puedes iniciar el servidor ahora.',
        'Instalación exitosa',
        {
          confirmButtonText: 'Ver servidor',
          callback: () => {
            router.push(`/servers/${selectedServerId.value}`);
          },
        }
      );
    }
  } catch {
    // Usuario canceló
  }
};

onMounted(async () => {
  await serversStore.fetchServers();
  
  const data = await marketplaceStore.fetchPluginDetail(source, pluginId);
  if (data) {
    plugin.value = data;
    
    // Seleccionar la primera versión por defecto
    if (data.versions && data.versions.length > 0) {
      selectedVersionId.value = data.versions[0].id;
    }
  }
});
</script>

<style scoped>
.prose {
  color: #374151;
  line-height: 1.75;
}
</style>
