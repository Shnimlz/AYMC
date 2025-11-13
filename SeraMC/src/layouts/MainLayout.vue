<template>
  <div class="h-screen flex overflow-hidden bg-gray-100">
    <!-- Sidebar -->
    <aside
      :class="[
        'bg-primary-800 text-white transition-all duration-300',
        sidebarCollapsed ? 'w-16' : 'w-64'
      ]"
    >
      <!-- Logo -->
      <div class="h-16 flex items-center justify-center border-b border-primary-700">
        <h1 v-if="!sidebarCollapsed" class="text-2xl font-bold">AYMC</h1>
        <h1 v-else class="text-xl font-bold">A</h1>
      </div>

      <!-- Navigation -->
      <nav class="mt-6 px-3">
        <router-link
          v-for="item in menuItems"
          :key="item.path"
          :to="item.path"
          class="flex items-center px-3 py-3 mb-2 rounded-lg transition-colors hover:bg-primary-700"
          :class="{ 'bg-primary-700': $route.path === item.path }"
        >
          <component :is="item.icon" class="w-5 h-5" />
          <span v-if="!sidebarCollapsed" class="ml-3">{{ item.title }}</span>
        </router-link>
      </nav>

      <!-- Toggle button -->
      <div class="absolute bottom-4 left-0 right-0 px-3">
        <el-button
          class="w-full"
          :icon="sidebarCollapsed ? ArrowRight : ArrowLeft"
          @click="sidebarCollapsed = !sidebarCollapsed"
        >
          {{ sidebarCollapsed ? '' : 'Contraer' }}
        </el-button>
      </div>
    </aside>

    <!-- Main content -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- Top navbar -->
      <header class="bg-white shadow-sm h-16 flex items-center justify-between px-6">
        <div class="flex items-center">
          <h2 class="text-xl font-semibold text-gray-800">
            {{ currentPageTitle }}
          </h2>
        </div>

        <div class="flex items-center space-x-4">
          <!-- User menu -->
          <el-dropdown @command="handleCommand">
            <div class="flex items-center cursor-pointer hover:text-primary-600">
              <el-avatar :size="32" :icon="UserFilled" />
              <span class="ml-2 text-sm font-medium">
                {{ authStore.user?.username }}
              </span>
              <el-icon class="ml-1"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>
                  Perfil
                </el-dropdown-item>
                <el-dropdown-item command="settings">
                  <el-icon><Setting /></el-icon>
                  Configuración
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  Cerrar Sesión
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <!-- Page content -->
      <main class="flex-1 overflow-y-auto p-6">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import {
  Monitor,
  ShoppingCart,
  FolderOpened,
  User,
  UserFilled,
  Setting,
  SwitchButton,
  ArrowDown,
  ArrowLeft,
  ArrowRight,
  Odometer,
} from '@element-plus/icons-vue';
import { ElMessageBox } from 'element-plus';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();
const sidebarCollapsed = ref(false);

const menuItems = [
  {
    path: '/dashboard',
    title: 'Panel de Control',
    icon: Odometer,
  },
  {
    path: '/servers',
    title: 'Servidores',
    icon: Monitor,
  },
  {
    path: '/marketplace',
    title: 'Marketplace',
    icon: ShoppingCart,
  },
  {
    path: '/backups',
    title: 'Respaldos',
    icon: FolderOpened,
  },
];

const currentPageTitle = computed(() => {
  return route.meta.title as string || 'AYMC';
});

const handleCommand = async (command: string) => {
  switch (command) {
    case 'profile':
      // TODO: Implementar vista de perfil
      break;
    case 'settings':
      // TODO: Implementar vista de configuración
      break;
    case 'logout':
      try {
        await ElMessageBox.confirm(
          '¿Estás seguro de que deseas cerrar sesión?',
          'Confirmar',
          {
            confirmButtonText: 'Sí, cerrar sesión',
            cancelButtonText: 'Cancelar',
            type: 'warning',
          }
        );
        
        await authStore.logout();
        router.push('/login');
      } catch {
        // Usuario canceló
      }
      break;
  }
};
</script>

<style scoped>
.router-link-active {
  @apply bg-primary-700;
}
</style>
