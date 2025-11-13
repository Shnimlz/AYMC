<template>
  <InstallationWizard
    :ssh-connected="true"
    @complete="handleComplete"
    @cancel="handleCancel"
  />
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import InstallationWizard from '@/components/InstallationWizard.vue';
import { useApiConfig } from '@/composables/useApiConfig';

const router = useRouter();
const { setConfig } = useApiConfig();

function handleComplete(apiUrl: string, wsUrl: string) {
  console.log('✅ Instalación completada:', { apiUrl, wsUrl });
  
  // Configurar API dinámicamente
  setConfig({
    apiUrl,
    wsUrl,
    environment: 'production',
  });
  
  // Marcar backend como instalado
  localStorage.setItem('aymc_backend_installed', 'true');
  
  // Marcar first-time como completado
  localStorage.setItem('aymc_first_time_completed', 'true');
  
  // Navegar a login
  router.push({ name: 'Login' });
}

function handleCancel() {
  console.log('⚠ Usuario canceló instalación');
  
  // Volver a detección
  router.push({ name: 'Detection' });
}
</script>
