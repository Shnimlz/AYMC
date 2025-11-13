<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary-900 to-primary-700 px-4">
    <div class="max-w-md w-full space-y-8 bg-white rounded-lg shadow-2xl p-8">
      <!-- Logo y título -->
      <div class="text-center">
        <h1 class="text-4xl font-bold text-primary-600">AYMC</h1>
        <p class="mt-2 text-sm text-gray-600">
          Crea tu cuenta para administrar servidores Minecraft
        </p>
      </div>

      <!-- Formulario -->
      <el-form
        ref="registerFormRef"
        :model="registerForm"
        :rules="rules"
        @submit.prevent="handleRegister"
        class="mt-8 space-y-6"
        label-position="top"
      >
        <el-form-item label="Usuario" prop="username">
          <el-input
            v-model="registerForm.username"
            placeholder="Elige un nombre de usuario"
            size="large"
            :prefix-icon="User"
            clearable
          />
        </el-form-item>

        <el-form-item label="Correo Electrónico" prop="email">
          <el-input
            v-model="registerForm.email"
            type="email"
            placeholder="tu@email.com"
            size="large"
            :prefix-icon="Message"
            clearable
          />
        </el-form-item>

        <el-form-item label="Contraseña" prop="password">
          <el-input
            v-model="registerForm.password"
            type="password"
            placeholder="Mínimo 6 caracteres"
            size="large"
            :prefix-icon="Lock"
            show-password
          />
        </el-form-item>

        <el-form-item label="Confirmar Contraseña" prop="confirmPassword">
          <el-input
            v-model="registerForm.confirmPassword"
            type="password"
            placeholder="Repite tu contraseña"
            size="large"
            :prefix-icon="Lock"
            show-password
            @keyup.enter="handleRegister"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            class="w-full"
            :loading="authStore.loading"
            @click="handleRegister"
          >
            <template v-if="!authStore.loading">
              Crear Cuenta
            </template>
            <template v-else>
              Creando cuenta...
            </template>
          </el-button>
        </el-form-item>

        <!-- Link a login -->
        <div class="text-center text-sm">
          <span class="text-gray-600">¿Ya tienes una cuenta?</span>
          <router-link
            to="/login"
            class="ml-1 font-medium text-primary-600 hover:text-primary-500"
          >
            Inicia sesión aquí
          </router-link>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { User, Lock, Message } from '@element-plus/icons-vue';
import type { FormInstance, FormRules } from 'element-plus';

const router = useRouter();
const authStore = useAuthStore();
const registerFormRef = ref<FormInstance>();

const registerForm = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
});

const validatePass = (_rule: any, value: any, callback: any) => {
  if (value === '') {
    callback(new Error('Por favor confirma tu contraseña'));
  } else if (value !== registerForm.password) {
    callback(new Error('Las contraseñas no coinciden'));
  } else {
    callback();
  }
};

const rules: FormRules = {
  username: [
    { required: true, message: 'Por favor ingresa un usuario', trigger: 'blur' },
    { min: 3, message: 'El usuario debe tener al menos 3 caracteres', trigger: 'blur' },
    { max: 20, message: 'El usuario no puede exceder 20 caracteres', trigger: 'blur' },
    { 
      pattern: /^[a-zA-Z0-9_]+$/, 
      message: 'Solo se permiten letras, números y guiones bajos', 
      trigger: 'blur' 
    },
  ],
  email: [
    { required: true, message: 'Por favor ingresa tu correo', trigger: 'blur' },
    { type: 'email', message: 'Por favor ingresa un correo válido', trigger: 'blur' },
  ],
  password: [
    { required: true, message: 'Por favor ingresa una contraseña', trigger: 'blur' },
    { min: 6, message: 'La contraseña debe tener al menos 6 caracteres', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: 'Por favor confirma tu contraseña', trigger: 'blur' },
    { validator: validatePass, trigger: 'blur' },
  ],
};

const handleRegister = async () => {
  if (!registerFormRef.value) return;

  await registerFormRef.value.validate(async (valid) => {
    if (!valid) return;

    const success = await authStore.register(
      registerForm.username,
      registerForm.email,
      registerForm.password
    );
    
    if (success) {
      router.push('/dashboard');
    }
  });
};
</script>

<style scoped>
.el-form-item {
  margin-bottom: 24px;
}
</style>
