<template>
  <OnboardingGallery @complete="handleComplete" />
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import OnboardingGallery from '@/components/OnboardingGallery.vue';

const router = useRouter();

function handleComplete() {
  console.log('✓ Onboarding gallery completado');
  
  // Navegar a SSH Setup
  router.push({ name: 'SSHSetup' });
}
</script>
