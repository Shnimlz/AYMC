<template>
  <ServiceDetectionView
    @install="handleInstall"
    @continue="handleContinue"
    @restart-services="handleRestartServices"
  />
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import ServiceDetectionView from '@/components/ServiceDetectionView.vue';

const router = useRouter();

function handleInstall() {
  console.log('📦 Iniciando instalación de AYMC');
  
  // Navegar a installer
  router.push({ name: 'Installer' });
}

function handleContinue() {
  console.log('✅ Backend instalado y corriendo, continuando...');
  
  // Marcar first-time como completado
  localStorage.setItem('aymc_first_time_completed', 'true');
  
  // Navegar a login
  router.push({ name: 'Login' });
}

async function handleRestartServices() {
  console.log('🔄 Reiniciando servicios...');
  
  // Aquí podrías agregar lógica para reiniciar servicios
  // Por ahora solo mostramos un mensaje
  alert('Funcionalidad de reinicio en desarrollo');
}
</script>
