import axios, {
  AxiosInstance,
  AxiosError,
  InternalAxiosRequestConfig,
} from "axios";
import { useAuthStore } from "@/stores/auth";
import { ElMessage } from "element-plus";

// Crear instancia de axios
const apiClient: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1",
  timeout: 30000,
  headers: {
    "Content-Type": "application/json",
  },
});

// Request interceptor - añadir token
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const authStore = useAuthStore();
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`;
    }
    return config;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  }
);

// Response interceptor - manejar errores
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const authStore = useAuthStore();

    if (error.response) {
      const status = error.response.status;

      // Token expirado o inválido
      if (status === 401) {
        authStore.logout();
        ElMessage.error(
          "Sesión expirada. Por favor, inicia sesión nuevamente."
        );
        window.location.href = "/login";
        return Promise.reject(error);
      }

      // Forbidden
      if (status === 403) {
        ElMessage.error("No tienes permisos para realizar esta acción");
        return Promise.reject(error);
      }

      // Server error
      if (status >= 500) {
        ElMessage.error("Error del servidor. Por favor, intenta más tarde.");
        return Promise.reject(error);
      }
    } else if (error.request) {
      // Request hecho pero sin respuesta
      ElMessage.error("No se pudo conectar con el servidor");
    } else {
      // Error al configurar request
      ElMessage.error("Error al realizar la petición");
    }

    return Promise.reject(error);
  }
);

export default apiClient;

// API Endpoints
export const authAPI = {
  register: (data: { username: string; email: string; password: string }) =>
    apiClient.post("/auth/register", data),

  login: (data: { username: string; password: string }) =>
    apiClient.post("/auth/login", data),

  logout: () => apiClient.post("/auth/logout"),

  getProfile: () => apiClient.get("/auth/me"),

  refreshToken: (refreshToken: string) =>
    apiClient.post("/auth/refresh", { refresh_token: refreshToken }),

  changePassword: (data: { current_password: string; new_password: string }) =>
    apiClient.post("/auth/change-password", data),
};

export const serversAPI = {
  list: () => apiClient.get("/servers"),

  get: (id: string) => apiClient.get(`/servers/${id}`),

  create: (data: any) => apiClient.post("/servers", data),

  update: (id: string, data: any) => apiClient.put(`/servers/${id}`, data),

  delete: (id: string) => apiClient.delete(`/servers/${id}`),

  start: (id: string) => apiClient.post(`/servers/${id}/start`),

  stop: (id: string) => apiClient.post(`/servers/${id}/stop`),

  restart: (id: string) => apiClient.post(`/servers/${id}/restart`),

  getStatus: (id: string) => apiClient.get(`/servers/${id}/status`),
};

export const agentsAPI = {
  list: () => apiClient.get("/agents"),

  get: (id: string) => apiClient.get(`/agents/${id}`),

  getHealth: (id: string) => apiClient.get(`/agents/${id}/health`),

  getMetrics: (id: string) => apiClient.get(`/agents/${id}/metrics`),

  getStats: () => apiClient.get("/agents/stats"),
};

export const marketplaceAPI = {
  search: (params: {
    query: string;
    source?: string;
    limit?: number;
    offset?: number;
  }) => apiClient.get("/marketplace/search", { params }),

  getPlugin: (source: string, id: string) =>
    apiClient.get(`/marketplace/${source}/${id}`),

  getVersions: (source: string, id: string) =>
    apiClient.get(`/marketplace/${source}/${id}/versions`),

  listInstalledPlugins: (serverId: string) =>
    apiClient.get(`/marketplace/servers/${serverId}/plugins`),

  installPlugin: (serverId: string, data: any) =>
    apiClient.post(`/marketplace/servers/${serverId}/plugins/install`, data),

  uninstallPlugin: (serverId: string, data: any) =>
    apiClient.post(`/marketplace/servers/${serverId}/plugins/uninstall`, data),

  updatePlugin: (serverId: string, data: any) =>
    apiClient.post(`/marketplace/servers/${serverId}/plugins/update`, data),
};

export const backupsAPI = {
  list: (serverId: string, params?: { limit?: number; offset?: number }) =>
    apiClient.get(`/servers/${serverId}/backups`, { params }),

  get: (backupId: string) => apiClient.get(`/backups/${backupId}`),

  create: (serverId: string, data: any) =>
    apiClient.post(`/servers/${serverId}/backups`, data),

  createManual: (serverId: string) =>
    apiClient.post(`/servers/${serverId}/backups/manual`),

  delete: (backupId: string) => apiClient.delete(`/backups/${backupId}`),

  restore: (backupId: string, data: any) =>
    apiClient.post(`/backups/${backupId}/restore`, data),

  getConfig: (serverId: string) =>
    apiClient.get(`/servers/${serverId}/backup-config`),

  updateConfig: (serverId: string, data: any) =>
    apiClient.put(`/servers/${serverId}/backup-config`, data),

  getStats: (serverId: string) =>
    apiClient.get(`/servers/${serverId}/backup-stats`),
};
