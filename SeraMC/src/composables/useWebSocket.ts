import { ref, onUnmounted } from "vue";
import { useAuthStore } from "@/stores/auth";
import { ElMessage } from "element-plus";

export interface WebSocketMessage {
  type: "log" | "status" | "error" | "info";
  server_id?: string;
  timestamp: string;
  message: string;
  level?: "info" | "warn" | "error" | "debug";
}

export function useWebSocket() {
  const authStore = useAuthStore();
  const ws = ref<WebSocket | null>(null);
  const connected = ref(false);
  const messages = ref<WebSocketMessage[]>([]);
  const reconnectAttempts = ref(0);
  const maxReconnectAttempts = 5;
  const reconnectDelay = 2000; // 2 segundos

  const getWebSocketUrl = () => {
    const baseUrl =
      import.meta.env.VITE_WS_URL || "ws://localhost:8080/api/v1/ws";
    const token = authStore.token;
    return `${baseUrl}?token=${token}`;
  };

  const connect = () => {
    if (ws.value?.readyState === WebSocket.OPEN) {
      console.log("WebSocket ya está conectado");
      return;
    }

    try {
      const url = getWebSocketUrl();
      ws.value = new WebSocket(url);

      ws.value.onopen = () => {
        console.log("✅ WebSocket conectado");
        connected.value = true;
        reconnectAttempts.value = 0;
        ElMessage.success("Conectado al servidor de logs");
      };

      ws.value.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          messages.value.push(data);

          // Mantener solo los últimos 1000 mensajes
          if (messages.value.length > 1000) {
            messages.value = messages.value.slice(-1000);
          }
        } catch (error) {
          console.error("Error al parsear mensaje WebSocket:", error);
        }
      };

      ws.value.onerror = (error) => {
        console.error("❌ Error WebSocket:", error);
        connected.value = false;
      };

      ws.value.onclose = (event) => {
        console.log("🔌 WebSocket desconectado:", event.code, event.reason);
        connected.value = false;

        // Intentar reconectar si no fue un cierre intencional
        if (
          event.code !== 1000 &&
          reconnectAttempts.value < maxReconnectAttempts
        ) {
          reconnectAttempts.value++;
          console.log(
            `🔄 Reintentando conexión (${reconnectAttempts.value}/${maxReconnectAttempts})...`
          );

          setTimeout(() => {
            connect();
          }, reconnectDelay * reconnectAttempts.value);
        } else if (reconnectAttempts.value >= maxReconnectAttempts) {
          ElMessage.error(
            "No se pudo conectar al servidor de logs. Recarga la página."
          );
        }
      };
    } catch (error) {
      console.error("Error al crear WebSocket:", error);
      ElMessage.error("Error al conectar con el servidor de logs");
    }
  };

  const disconnect = () => {
    if (ws.value) {
      ws.value.close(1000, "Desconexión intencional");
      ws.value = null;
      connected.value = false;
      reconnectAttempts.value = maxReconnectAttempts; // Evitar reconexión automática
    }
  };

  const subscribe = (serverId: string) => {
    if (ws.value?.readyState === WebSocket.OPEN) {
      const message = JSON.stringify({
        action: "subscribe",
        server_id: serverId,
      });
      ws.value.send(message);
      console.log(`📡 Suscrito a logs del servidor ${serverId}`);
    } else {
      console.error("WebSocket no está conectado");
    }
  };

  const unsubscribe = (serverId: string) => {
    if (ws.value?.readyState === WebSocket.OPEN) {
      const message = JSON.stringify({
        action: "unsubscribe",
        server_id: serverId,
      });
      ws.value.send(message);
      console.log(`📡 Desuscrito de logs del servidor ${serverId}`);
    }
  };

  const sendCommand = (serverId: string, command: string) => {
    if (ws.value?.readyState === WebSocket.OPEN) {
      const message = JSON.stringify({
        action: "command",
        server_id: serverId,
        command: command,
      });
      ws.value.send(message);
      console.log(`📤 Comando enviado: ${command}`);
    } else {
      ElMessage.error("No estás conectado al servidor");
    }
  };

  const clearMessages = () => {
    messages.value = [];
  };

  // Limpiar al desmontar el componente
  onUnmounted(() => {
    disconnect();
  });

  return {
    // State
    ws,
    connected,
    messages,
    reconnectAttempts,

    // Methods
    connect,
    disconnect,
    subscribe,
    unsubscribe,
    sendCommand,
    clearMessages,
  };
}
