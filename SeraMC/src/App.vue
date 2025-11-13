<template>
  <router-view />
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { initializeApiConfig, useApiConfig } from '@/composables/useApiConfig';

const router = useRouter();
const { isBackendInstalled } = useApiConfig();

// Constante para first-time setup
const FIRST_TIME_KEY = 'aymc_first_time_completed';

/**
 * Determinar ruta inicial basado en estado de la app
 */
function determineInitialRoute() {
  const isFirstTime = localStorage.getItem(FIRST_TIME_KEY) !== 'true';
  const backendInstalled = isBackendInstalled();
  const currentPath = window.location.pathname;

  console.log('🔍 Estado de la aplicación:', {
    isFirstTime,
    backendInstalled,
    currentPath,
  });

  // Si ya estamos en una ruta específica, no redirigir
  if (currentPath !== '/' && currentPath !== '') {
    console.log('✓ Usuario ya en ruta específica, no redirigir');
    return;
  }

  // Primera vez: Onboarding completo
  if (isFirstTime) {
    console.log('👋 Primera vez: Redirigiendo a Welcome');
    router.replace({ name: 'Welcome' });
    return;
  }

  // Backend no instalado: Ir a SSH Setup
  if (!backendInstalled) {
    console.log('🔧 Backend no instalado: Redirigiendo a SSH Setup');
    router.replace({ name: 'SSHSetup' });
    return;
  }

  // Todo configurado: Ir a Login o Dashboard
  console.log('✅ Sistema configurado: Redirigiendo a Login');
  router.replace({ name: 'Login' });
}

onMounted(() => {
  // Inicializar configuración de API
  initializeApiConfig();

  // Determinar y navegar a ruta inicial
  determineInitialRoute();
});
</script>

<style>
/* Global styles */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

#app {
  width: 100%;
  height: 100vh;
  overflow: hidden;
}
</style>
