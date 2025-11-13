<template>
  <div class="onboarding-container">
    <div class="onboarding-gallery">
      <!-- Swiper Container -->
      <swiper
        :modules="modules"
        :pagination="{ clickable: true }"
        :navigation="true"
        :space-between="50"
        @swiper="onSwiper"
        @slideChange="onSlideChange"
        class="onboarding-swiper"
      >
        <!-- Slide 1: Bienvenida -->
        <swiper-slide>
          <div class="slide-content">
            <div class="slide-icon">🎮</div>
            <h1>Bienvenido a AYMC</h1>
            <p class="slide-description">
              Advanced Yet Manageable Control - El panel de control definitivo
              para gestionar tus servidores de Minecraft desde una sola aplicación.
            </p>
            <div class="slide-features">
              <div class="feature-badge">
                <span class="badge-icon">🚀</span>
                <span>Rápido y Eficiente</span>
              </div>
              <div class="feature-badge">
                <span class="badge-icon">🔒</span>
                <span>Seguro</span>
              </div>
              <div class="feature-badge">
                <span class="badge-icon">🌐</span>
                <span>Multi-Servidor</span>
              </div>
            </div>
          </div>
        </swiper-slide>

        <!-- Slide 2: Gestión de Servidores -->
        <swiper-slide>
          <div class="slide-content">
            <div class="slide-icon">🎯</div>
            <h2>Gestión Centralizada</h2>
            <p class="slide-description">
              Administra múltiples servidores de Minecraft desde una interfaz
              intuitiva. Crea, inicia, detén y configura servidores con unos pocos clics.
            </p>
            <ul class="feature-list">
              <li>✅ Soporte para Paper, Spigot, Purpur y Vanilla</li>
              <li>✅ Configuración automática de puertos</li>
              <li>✅ Gestión de recursos (RAM, CPU)</li>
              <li>✅ Múltiples versiones de Minecraft</li>
            </ul>
          </div>
        </swiper-slide>

        <!-- Slide 3: Marketplace de Plugins -->
        <swiper-slide>
          <div class="slide-content">
            <div class="slide-icon">🔌</div>
            <h2>Marketplace Integrado</h2>
            <p class="slide-description">
              Accede a miles de plugins desde SpigotMC, Hangar, Modrinth y
              CurseForge. Instala y actualiza plugins sin salir de la aplicación.
            </p>
            <ul class="feature-list">
              <li>✅ Búsqueda unificada en 4 marketplaces</li>
              <li>✅ Instalación con un clic</li>
              <li>✅ Actualizaciones automáticas</li>
              <li>✅ Gestión de dependencias</li>
            </ul>
          </div>
        </swiper-slide>

        <!-- Slide 4: Backups Automáticos -->
        <swiper-slide>
          <div class="slide-content">
            <div class="slide-icon">💾</div>
            <h2>Backups Inteligentes</h2>
            <p class="slide-description">
              Protege tus mundos con backups automáticos programables.
              Backups completos, incrementales y restauración con un clic.
            </p>
            <ul class="feature-list">
              <li>✅ Backups programados automáticamente</li>
              <li>✅ Backups incrementales (ahorra espacio)</li>
              <li>✅ Restauración rápida</li>
              <li>✅ Almacenamiento local o remoto</li>
            </ul>
          </div>
        </swiper-slide>

        <!-- Slide 5: Monitoreo en Tiempo Real -->
        <swiper-slide>
          <div class="slide-content">
            <div class="slide-icon">📊</div>
            <h2>Monitoreo en Vivo</h2>
            <p class="slide-description">
              Observa el rendimiento de tus servidores en tiempo real.
              CPU, RAM, jugadores conectados, TPS y logs instantáneos.
            </p>
            <ul class="feature-list">
              <li>✅ Gráficas de rendimiento en tiempo real</li>
              <li>✅ Logs con búsqueda y filtrado</li>
              <li>✅ Alertas configurables</li>
              <li>✅ Vista de jugadores conectados</li>
            </ul>
          </div>
        </swiper-slide>

        <!-- Slide 6: Llamado a la Acción -->
        <swiper-slide>
          <div class="slide-content slide-cta">
            <div class="slide-icon">🚀</div>
            <h2>¡Empecemos!</h2>
            <p class="slide-description">
              Para comenzar, necesitamos conectarnos a tu servidor VPS.
              Prepara tus credenciales SSH y te guiaremos en el proceso.
            </p>
            <button @click="startSetup" class="cta-button">
              Comenzar Configuración
            </button>
            <p class="slide-note">
              💡 Tip: Ten a mano la IP de tu VPS y tus credenciales SSH
            </p>
          </div>
        </swiper-slide>
      </swiper>

      <!-- Progress Indicator -->
      <div class="progress-bar">
        <div class="progress-fill" :style="{ width: progressWidth }"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { Swiper, SwiperSlide } from 'swiper/vue';
import { Navigation, Pagination } from 'swiper/modules';
import type { Swiper as SwiperType } from 'swiper';

// Importar estilos de Swiper
import 'swiper/css';
import 'swiper/css/navigation';
import 'swiper/css/pagination';

const emit = defineEmits<{
  complete: [];
}>();

const modules = [Navigation, Pagination];
const currentSlide = ref(0);
const totalSlides = ref(6);
const swiperInstance = ref<SwiperType | null>(null);

const progressWidth = computed(() => {
  return `${((currentSlide.value + 1) / totalSlides.value) * 100}%`;
});

function onSwiper(swiper: SwiperType) {
  swiperInstance.value = swiper;
  totalSlides.value = swiper.slides.length;
}

function onSlideChange(swiper: SwiperType) {
  currentSlide.value = swiper.activeIndex;
}

function startSetup() {
  // Navegar a SSH Setup
  emit('complete');
  
  // Opcional: Usar router directamente
  // import { useRouter } from 'vue-router';
  // const router = useRouter();
  // router.push({ name: 'SSHSetup' });
}
</script>

<style scoped>
.onboarding-container {
  width: 100%;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  overflow: hidden;
}

.onboarding-gallery {
  flex: 1;
  display: flex;
  flex-direction: column;
  position: relative;
}

.onboarding-swiper {
  flex: 1;
  width: 100%;
}

.slide-content {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 40px;
  color: white;
  text-align: center;
}

.slide-icon {
  font-size: 80px;
  margin-bottom: 30px;
  animation: bounce 2s infinite;
}

@keyframes bounce {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-20px);
  }
}

h1 {
  font-size: 48px;
  font-weight: bold;
  margin-bottom: 20px;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
}

h2 {
  font-size: 36px;
  font-weight: bold;
  margin-bottom: 20px;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
}

.slide-description {
  font-size: 20px;
  line-height: 1.6;
  max-width: 700px;
  margin-bottom: 40px;
  opacity: 0.95;
}

.slide-features {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
  justify-content: center;
}

.feature-badge {
  display: flex;
  align-items: center;
  gap: 10px;
  background: rgba(255, 255, 255, 0.2);
  padding: 12px 24px;
  border-radius: 30px;
  backdrop-filter: blur(10px);
  font-weight: 600;
}

.badge-icon {
  font-size: 24px;
}

.feature-list {
  list-style: none;
  padding: 0;
  max-width: 600px;
  text-align: left;
}

.feature-list li {
  font-size: 18px;
  line-height: 2;
  padding-left: 10px;
}

.slide-cta {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 20px;
  backdrop-filter: blur(10px);
  margin: 40px;
  padding: 60px 40px;
}

.cta-button {
  background: white;
  color: #667eea;
  border: none;
  padding: 16px 48px;
  border-radius: 30px;
  font-size: 20px;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.3);
}

.cta-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.4);
}

.slide-note {
  margin-top: 30px;
  font-size: 16px;
  opacity: 0.8;
}

.progress-bar {
  height: 6px;
  background: rgba(255, 255, 255, 0.3);
  position: relative;
}

.progress-fill {
  height: 100%;
  background: white;
  transition: width 0.3s ease;
  box-shadow: 0 0 10px rgba(255, 255, 255, 0.5);
}

/* Estilos personalizados de Swiper */
:deep(.swiper-button-next),
:deep(.swiper-button-prev) {
  color: white;
  background: rgba(255, 255, 255, 0.2);
  width: 50px;
  height: 50px;
  border-radius: 50%;
  backdrop-filter: blur(10px);
}

:deep(.swiper-button-next):hover,
:deep(.swiper-button-prev):hover {
  background: rgba(255, 255, 255, 0.3);
}

:deep(.swiper-button-next::after),
:deep(.swiper-button-prev::after) {
  font-size: 20px;
}

:deep(.swiper-pagination-bullet) {
  background: white;
  opacity: 0.5;
  width: 12px;
  height: 12px;
}

:deep(.swiper-pagination-bullet-active) {
  opacity: 1;
  background: white;
}

/* Responsive */
@media (max-width: 768px) {
  h1 {
    font-size: 32px;
  }

  h2 {
    font-size: 28px;
  }

  .slide-description {
    font-size: 16px;
  }

  .slide-icon {
    font-size: 60px;
  }

  .feature-list li {
    font-size: 16px;
  }

  .slide-content {
    padding: 40px 20px;
  }
}
</style>
