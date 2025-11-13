import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { authAPI } from "@/api";
import { ElMessage } from "element-plus";

export interface User {
  id: string;
  username: string;
  email: string;
  created_at: string;
}

export const useAuthStore = defineStore("auth", () => {
  // State
  const token = ref<string | null>(localStorage.getItem("token"));
  const refreshToken = ref<string | null>(localStorage.getItem("refreshToken"));
  const user = ref<User | null>(null);
  const loading = ref(false);

  // Getters
  const isAuthenticated = computed(() => !!token.value && !!user.value);

  // Actions
  async function login(username: string, password: string) {
    try {
      loading.value = true;
      const response = await authAPI.login({ username, password });

      if (response.data.token) {
        token.value = response.data.token;
        refreshToken.value = response.data.refresh_token;

        // Guardar en localStorage
        localStorage.setItem("token", response.data.token);
        localStorage.setItem("refreshToken", response.data.refresh_token);

        // Obtener perfil
        await getProfile();

        ElMessage.success("Inicio de sesión exitoso");
        return true;
      }
      return false;
    } catch (error: any) {
      const message = error.response?.data?.error || "Error al iniciar sesión";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function register(username: string, email: string, password: string) {
    try {
      loading.value = true;
      await authAPI.register({ username, email, password });

      ElMessage.success("Registro exitoso. Iniciando sesión...");

      // Auto-login después de registro
      return await login(username, password);
    } catch (error: any) {
      const message = error.response?.data?.error || "Error al registrarse";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function getProfile() {
    try {
      const response = await authAPI.getProfile();
      user.value = response.data;
      return true;
    } catch (error: any) {
      console.error("Error al obtener perfil:", error);
      return false;
    }
  }

  async function logout() {
    try {
      await authAPI.logout();
    } catch (error) {
      console.error("Error al cerrar sesión:", error);
    } finally {
      // Limpiar estado
      token.value = null;
      refreshToken.value = null;
      user.value = null;

      // Limpiar localStorage
      localStorage.removeItem("token");
      localStorage.removeItem("refreshToken");

      ElMessage.info("Sesión cerrada");
    }
  }

  async function changePassword(currentPassword: string, newPassword: string) {
    try {
      loading.value = true;
      await authAPI.changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      });

      ElMessage.success("Contraseña actualizada correctamente");
      return true;
    } catch (error: any) {
      const message =
        error.response?.data?.error || "Error al cambiar contraseña";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function refresh() {
    if (!refreshToken.value) {
      return false;
    }

    try {
      const response = await authAPI.refreshToken(refreshToken.value);

      if (response.data.token) {
        token.value = response.data.token;
        localStorage.setItem("token", response.data.token);
        return true;
      }
      return false;
    } catch (error) {
      console.error("Error al refrescar token:", error);
      logout();
      return false;
    }
  }

  // Inicializar: si hay token, obtener perfil
  if (token.value) {
    getProfile().catch(() => {
      // Si falla, limpiar token
      logout();
    });
  }

  return {
    // State
    token,
    refreshToken,
    user,
    loading,

    // Getters
    isAuthenticated,

    // Actions
    login,
    register,
    getProfile,
    logout,
    changePassword,
    refresh,
  };
});
