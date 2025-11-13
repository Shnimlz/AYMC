import { defineStore } from "pinia";
import { ref } from "vue";
import { agentsAPI } from "@/api";
import { ElMessage } from "element-plus";

export interface Agent {
  id: string;
  name: string;
  host: string;
  port: number;
  status: "online" | "offline" | "error";
  version: string;
  created_at: string;
  updated_at: string;
}

export interface AgentMetrics {
  cpu_usage: number;
  memory_usage: number;
  memory_total: number;
  disk_usage: number;
  disk_total: number;
  uptime: number;
}

export const useAgentsStore = defineStore("agents", () => {
  // State
  const agents = ref<Agent[]>([]);
  const selectedAgent = ref<Agent | null>(null);
  const metrics = ref<AgentMetrics | null>(null);
  const loading = ref(false);

  // Actions
  async function fetchAgents() {
    try {
      loading.value = true;
      const response = await agentsAPI.list();
      agents.value = response.data;
      return true;
    } catch (error: any) {
      const message = error.response?.data?.error || "Error al cargar agentes";
      ElMessage.error(message);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function fetchAgent(id: string) {
    try {
      loading.value = true;
      const response = await agentsAPI.get(id);
      selectedAgent.value = response.data;
      return response.data;
    } catch (error: any) {
      const message = error.response?.data?.error || "Error al cargar agente";
      ElMessage.error(message);
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function fetchAgentMetrics(id: string) {
    try {
      const response = await agentsAPI.getMetrics(id);
      metrics.value = response.data;
      return response.data;
    } catch (error: any) {
      console.error("Error al cargar métricas:", error);
      return null;
    }
  }

  async function checkAgentHealth(id: string) {
    try {
      const response = await agentsAPI.getHealth(id);
      return response.data;
    } catch (error: any) {
      console.error("Error al verificar salud del agente:", error);
      return null;
    }
  }

  function selectAgent(agent: Agent | null) {
    selectedAgent.value = agent;
  }

  return {
    // State
    agents,
    selectedAgent,
    metrics,
    loading,

    // Actions
    fetchAgents,
    fetchAgent,
    fetchAgentMetrics,
    checkAgentHealth,
    selectAgent,
  };
});
