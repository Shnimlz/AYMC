<template>
  <SSHConnectionForm 
    @connected="handleConnected" 
    @cancel="handleCancel"
  />
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import SSHConnectionForm from '@/components/SSHConnectionForm.vue';

const router = useRouter();

function handleConnected() {
  console.log('✓ SSH conectado exitosamente');
  
  // Navegar a detección de servicios
  router.push({ name: 'Detection' });
}

function handleCancel() {
  console.log('⚠ Usuario canceló conexión SSH');
  
  // Volver al welcome o salir
  router.push({ name: 'Welcome' });
}
</script>
