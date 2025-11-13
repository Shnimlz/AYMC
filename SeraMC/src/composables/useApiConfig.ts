import { ref, computed } from "vue";
import { invoke } from "@tauri-apps/api/core";

/**
 * Configuración de API dinámica
 *
 * Detecta automáticamente si estamos en:
 * - Development: localhost (VITE_API_URL manual)
 * - Production: VPS remota (API_URL detectada vía SSH)
 */

interface BackendConfig {
  api_url: string;
  ws_url: string;
  environment: string;
  port: string;
}

// Estado global de configuración
const apiUrl = ref<string>("");
const wsUrl = ref<string>("");
const environment = ref<"development" | "production">("development");
const isConfigured = ref(false);
const isLoading = ref(false);
const error = ref<string | null>(null);

// Storage keys
const STORAGE_KEYS = {
  API_URL: "aymc_api_url",
  WS_URL: "aymc_ws_url",
  ENVIRONMENT: "aymc_environment",
  BACKEND_INSTALLED: "aymc_backend_installed",
  SSH_HOST: "aymc_ssh_host",
} as const;

/**
 * Hook principal para configuración de API
 */
export function useApiConfig() {
  /**
   * Inicializar configuración desde localStorage o defaults
   */
  function initFromStorage() {
    try {
      const storedApiUrl = localStorage.getItem(STORAGE_KEYS.API_URL);
      const storedWsUrl = localStorage.getItem(STORAGE_KEYS.WS_URL);
      const storedEnv = localStorage.getItem(STORAGE_KEYS.ENVIRONMENT);

      if (storedApiUrl && storedWsUrl) {
        apiUrl.value = storedApiUrl;
        wsUrl.value = storedWsUrl;
        environment.value = (storedEnv as any) || "development";
        isConfigured.value = true;

        console.log("✓ Configuración cargada desde localStorage:", {
          apiUrl: apiUrl.value,
          wsUrl: wsUrl.value,
          environment: environment.value,
        });
      } else {
        // Usar valores por defecto de environment variables
        const defaultApiUrl =
          import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1";
        const defaultWsUrl =
          import.meta.env.VITE_WS_URL || "ws://localhost:8080/api/v1/ws";

        apiUrl.value = defaultApiUrl;
        wsUrl.value = defaultWsUrl;
        environment.value = "development";
        isConfigured.value = false;

        console.log("⚠ Usando configuración por defecto (development)");
      }
    } catch (err) {
      console.error("Error inicializando configuración:", err);
      error.value = "Error al cargar configuración guardada";
    }
  }

  /**
   * Detectar configuración desde VPS vía SSH
   */
  async function detectFromVPS() {
    isLoading.value = true;
    error.value = null;

    try {
      // Verificar conexión SSH
      const isConnected = await invoke<boolean>("ssh_is_connected");

      if (!isConnected) {
        throw new Error("No hay conexión SSH activa");
      }

      // Obtener configuración del backend
      const config = await invoke<BackendConfig>("ssh_get_backend_config");

      if (!config.api_url || !config.ws_url) {
        throw new Error("Backend no tiene configuración válida");
      }

      // Actualizar estado
      apiUrl.value = config.api_url;
      wsUrl.value = config.ws_url;
      environment.value =
        config.environment === "production" ? "production" : "development";
      isConfigured.value = true;

      // Guardar en localStorage
      saveToStorage();

      // Marcar backend como instalado
      localStorage.setItem(STORAGE_KEYS.BACKEND_INSTALLED, "true");

      console.log("✓ Configuración detectada desde VPS:", {
        apiUrl: apiUrl.value,
        wsUrl: wsUrl.value,
        environment: environment.value,
      });

      return {
        apiUrl: apiUrl.value,
        wsUrl: wsUrl.value,
        environment: environment.value,
      };
    } catch (err: any) {
      console.error("Error detectando configuración desde VPS:", err);
      error.value = err.message || "Error al detectar configuración";
      throw err;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Configurar manualmente (útil después de instalación)
   */
  function setConfig(config: {
    apiUrl: string;
    wsUrl: string;
    environment?: "development" | "production";
  }) {
    apiUrl.value = config.apiUrl;
    wsUrl.value = config.wsUrl;
    environment.value = config.environment || "production";
    isConfigured.value = true;

    saveToStorage();

    console.log("✓ Configuración manual aplicada:", {
      apiUrl: apiUrl.value,
      wsUrl: wsUrl.value,
      environment: environment.value,
    });
  }

  /**
   * Guardar configuración en localStorage
   */
  function saveToStorage() {
    try {
      localStorage.setItem(STORAGE_KEYS.API_URL, apiUrl.value);
      localStorage.setItem(STORAGE_KEYS.WS_URL, wsUrl.value);
      localStorage.setItem(STORAGE_KEYS.ENVIRONMENT, environment.value);
    } catch (err) {
      console.error("Error guardando configuración:", err);
    }
  }

  /**
   * Limpiar configuración (logout o reset)
   */
  function clearConfig() {
    apiUrl.value = "";
    wsUrl.value = "";
    environment.value = "development";
    isConfigured.value = false;
    error.value = null;

    try {
      localStorage.removeItem(STORAGE_KEYS.API_URL);
      localStorage.removeItem(STORAGE_KEYS.WS_URL);
      localStorage.removeItem(STORAGE_KEYS.ENVIRONMENT);
      localStorage.removeItem(STORAGE_KEYS.BACKEND_INSTALLED);

      console.log("✓ Configuración limpiada");
    } catch (err) {
      console.error("Error limpiando configuración:", err);
    }
  }

  /**
   * Verificar si el backend está instalado
   */
  function isBackendInstalled(): boolean {
    return localStorage.getItem(STORAGE_KEYS.BACKEND_INSTALLED) === "true";
  }

  /**
   * Obtener URL completa para una ruta
   */
  function getApiUrl(path: string = ""): string {
    if (!apiUrl.value) {
      console.warn("API URL no configurada, usando default");
      return `http://localhost:8080/api/v1${path}`;
    }

    const base = apiUrl.value.endsWith("/")
      ? apiUrl.value.slice(0, -1)
      : apiUrl.value;
    const cleanPath = path.startsWith("/") ? path : `/${path}`;

    return `${base}${cleanPath}`;
  }

  /**
   * Obtener WebSocket URL
   */
  function getWsUrl(): string {
    if (!wsUrl.value) {
      console.warn("WS URL no configurada, usando default");
      return "ws://localhost:8080/api/v1/ws";
    }
    return wsUrl.value;
  }

  // Computed
  const isDevelopment = computed(() => environment.value === "development");
  const isProduction = computed(() => environment.value === "production");
  const currentApiUrl = computed(() => apiUrl.value);
  const currentWsUrl = computed(() => wsUrl.value);

  return {
    // Estado
    apiUrl: currentApiUrl,
    wsUrl: currentWsUrl,
    environment,
    isConfigured,
    isLoading,
    error,
    isDevelopment,
    isProduction,

    // Métodos
    initFromStorage,
    detectFromVPS,
    setConfig,
    clearConfig,
    saveToStorage,
    isBackendInstalled,
    getApiUrl,
    getWsUrl,
  };
}

/**
 * Inicializar configuración al cargar la app
 * Llamar en App.vue o main.ts
 */
export function initializeApiConfig() {
  const { initFromStorage } = useApiConfig();
  initFromStorage();
}
