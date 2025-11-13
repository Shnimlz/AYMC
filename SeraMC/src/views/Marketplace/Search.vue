<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold text-gray-800">Marketplace</h1>
      <p class="text-gray-600 mt-1">Busca e instala plugins para tus servidores</p>
    </div>

    <!-- Search bar -->
    <el-card>
      <div class="flex gap-4">
        <el-input
          v-model="searchQuery"
          placeholder="Buscar plugins..."
          :prefix-icon="Search"
          size="large"
          clearable
          @keyup.enter="handleSearch"
          class="flex-1"
        />
        
        <el-select
          v-model="sourceFilter"
          placeholder="Fuente"
          size="large"
          class="w-48"
        >
          <el-option label="Todas" value="all" />
          <el-option label="Modrinth" value="modrinth" />
          <el-option label="Spigot" value="spigot" />
        </el-select>

        <el-button
          type="primary"
          size="large"
          :icon="Search"
          :loading="marketplaceStore.loading"
          @click="handleSearch"
        >
          Buscar
        </el-button>
      </div>
    </el-card>

    <!-- Quick links -->
    <div class="flex gap-4">
      <el-button :icon="Star" @click="searchQuery = 'worldedit'; handleSearch()">
        WorldEdit
      </el-button>
      <el-button :icon="Star" @click="searchQuery = 'essentials'; handleSearch()">
        Essentials
      </el-button>
      <el-button :icon="Star" @click="searchQuery = 'vault'; handleSearch()">
        Vault
      </el-button>
      <el-button :icon="Star" @click="searchQuery = 'luckperms'; handleSearch()">
        LuckPerms
      </el-button>
      <el-button
        type="success"
        :icon="FolderOpened"
        @click="router.push('/marketplace/installed')"
      >
        Ver Instalados
      </el-button>
    </div>

    <!-- Results -->
    <div v-if="marketplaceStore.searchResults.length > 0">
      <div class="mb-4 text-gray-600">
        {{ marketplaceStore.searchResults.length }} resultados encontrados
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <el-card
          v-for="plugin in marketplaceStore.searchResults"
          :key="`${plugin.source}-${plugin.id}`"
          shadow="hover"
          class="cursor-pointer hover:shadow-lg transition-shadow"
          @click="viewPlugin(plugin)"
        >
          <div class="flex gap-4">
            <!-- Icon -->
            <div class="flex-shrink-0">
              <el-avatar
                :size="64"
                :src="plugin.icon_url"
                :icon="Box"
              />
            </div>

            <!-- Info -->
            <div class="flex-1 min-w-0">
              <div class="flex items-start justify-between gap-2">
                <h3 class="font-bold text-lg truncate">{{ plugin.name }}</h3>
                <el-tag size="small">{{ plugin.source }}</el-tag>
              </div>

              <p class="text-sm text-gray-600 mt-1">
                por <span class="font-medium">{{ plugin.author }}</span>
              </p>

              <p class="text-sm text-gray-700 mt-2 line-clamp-2">
                {{ plugin.description }}
              </p>

              <div class="flex items-center gap-4 mt-3 text-sm text-gray-500">
                <div class="flex items-center gap-1">
                  <el-icon><Download /></el-icon>
                  <span>{{ formatNumber(plugin.downloads) }}</span>
                </div>
                <div v-if="plugin.rating" class="flex items-center gap-1">
                  <el-icon><Star /></el-icon>
                  <span>{{ plugin.rating.toFixed(1) }}</span>
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </div>
    </div>

    <!-- Empty state -->
    <el-empty
      v-else-if="!marketplaceStore.loading && marketplaceStore.searchQuery"
      description="No se encontraron plugins"
    >
      <el-button type="primary" @click="searchQuery = ''; marketplaceStore.searchQuery = ''">
        Limpiar búsqueda
      </el-button>
    </el-empty>

    <el-empty
      v-else-if="!marketplaceStore.loading"
      description="Busca plugins para tus servidores"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useMarketplaceStore } from '@/stores/marketplace';
import {
  Search,
  Star,
  FolderOpened,
  Box,
  Download,
} from '@element-plus/icons-vue';

const router = useRouter();
const marketplaceStore = useMarketplaceStore();

const searchQuery = ref('');
const sourceFilter = ref('all');

const handleSearch = async () => {
  if (!searchQuery.value.trim()) return;
  
  await marketplaceStore.searchPlugins(
    searchQuery.value,
    sourceFilter.value === 'all' ? undefined : sourceFilter.value
  );
};

const viewPlugin = (plugin: any) => {
  router.push(`/marketplace/${plugin.source}/${plugin.slug || plugin.id}`);
};

const formatNumber = (num: number) => {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M';
  } else if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K';
  }
  return num.toString();
};

onMounted(() => {
  // Restaurar búsqueda anterior si existe
  if (marketplaceStore.searchQuery) {
    searchQuery.value = marketplaceStore.searchQuery;
    sourceFilter.value = marketplaceStore.searchSource;
  }
});
</script>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
