<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary-900 to-primary-700 px-4">
    <div class="max-w-md w-full space-y-8 bg-white rounded-lg shadow-2xl p-8">
      <!-- Logo y título -->
      <div class="text-center">
        <h1 class="text-4xl font-bold text-primary-600">AYMC</h1>
        <p class="mt-2 text-sm text-gray-600">
          Sistema de Administración de Servidores Minecraft
        </p>
      </div>

      <!-- Formulario -->
      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="rules"
        @submit.prevent="handleLogin"
        class="mt-8 space-y-6"
        label-position="top"
      >
        <el-form-item label="Usuario" prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="Ingresa tu usuario"
            size="large"
            :prefix-icon="User"
            clearable
          />
        </el-form-item>

        <el-form-item label="Contraseña" prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="Ingresa tu contraseña"
            size="large"
            :prefix-icon="Lock"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            class="w-full"
            :loading="authStore.loading"
            @click="handleLogin"
          >
            <template v-if="!authStore.loading">
              Iniciar Sesión
            </template>
            <template v-else>
              Iniciando sesión...
            </template>
          </el-button>
        </el-form-item>

        <!-- Link a registro -->
        <div class="text-center text-sm">
          <span class="text-gray-600">¿No tienes una cuenta?</span>
          <router-link
            to="/register"
            class="ml-1 font-medium text-primary-600 hover:text-primary-500"
          >
            Regístrate aquí
          </router-link>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { User, Lock } from '@element-plus/icons-vue';
import type { FormInstance, FormRules } from 'element-plus';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();
const loginFormRef = ref<FormInstance>();

const loginForm = reactive({
  username: '',
  password: '',
});

const rules: FormRules = {
  username: [
    { required: true, message: 'Por favor ingresa tu usuario', trigger: 'blur' },
    { min: 3, message: 'El usuario debe tener al menos 3 caracteres', trigger: 'blur' },
  ],
  password: [
    { required: true, message: 'Por favor ingresa tu contraseña', trigger: 'blur' },
    { min: 6, message: 'La contraseña debe tener al menos 6 caracteres', trigger: 'blur' },
  ],
};

const handleLogin = async () => {
  if (!loginFormRef.value) return;

  await loginFormRef.value.validate(async (valid) => {
    if (!valid) return;

    const success = await authStore.login(loginForm.username, loginForm.password);
    
    if (success) {
      // Redirigir a la ruta original o al dashboard
      const redirect = route.query.redirect as string || '/dashboard';
      router.push(redirect);
    }
  });
};
</script>

<style scoped>
.el-form-item {
  margin-bottom: 24px;
}
</style>
