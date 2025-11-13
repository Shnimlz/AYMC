<template>
  <div class="max-w-4xl">
    <!-- Header -->
    <div class="mb-6">
      <el-button :icon="ArrowLeft" @click="router.back()" class="mb-4">
        Volver
      </el-button>
      <h1 class="text-2xl font-bold text-gray-800">Crear Nuevo Servidor</h1>
      <p class="text-gray-600 mt-1">Configura un nuevo servidor de Minecraft</p>
    </div>

    <!-- Form -->
    <el-card>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="150px"
        label-position="left"
      >
        <el-divider content-position="left">Información Básica</el-divider>

        <el-form-item label="Nombre" prop="name">
          <el-input
            v-model="form.name"
            placeholder="Ej: Servidor de Supervivencia"
            clearable
          />
        </el-form-item>

        <el-form-item label="Tipo" prop="type">
          <el-select v-model="form.type" placeholder="Selecciona el tipo" class="w-full">
            <el-option label="Vanilla" value="vanilla" />
            <el-option label="Spigot" value="spigot" />
            <el-option label="Paper" value="paper" />
            <el-option label="Fabric" value="fabric" />
            <el-option label="Forge" value="forge" />
            <el-option label="Purpur" value="purpur" />
          </el-select>
        </el-form-item>

        <el-form-item label="Versión" prop="version">
          <el-input
            v-model="form.version"
            placeholder="Ej: 1.20.1"
            clearable
          />
        </el-form-item>

        <el-form-item label="Puerto" prop="port">
          <el-input-number
            v-model="form.port"
            :min="1024"
            :max="65535"
            :step="1"
            class="w-full"
          />
        </el-form-item>

        <el-form-item label="Agente" prop="agent_id">
          <el-select
            v-model="form.agent_id"
            placeholder="Selecciona el agente"
            class="w-full"
            :loading="agentsStore.loading"
          >
            <el-option
              v-for="agent in agentsStore.agents"
              :key="agent.id"
              :label="`${agent.name} (${agent.host}:${agent.port})`"
              :value="agent.id"
              :disabled="agent.status !== 'online'"
            >
              <div class="flex items-center justify-between">
                <span>{{ agent.name }}</span>
                <el-tag
                  :type="agent.status === 'online' ? 'success' : 'danger'"
                  size="small"
                >
                  {{ agent.status }}
                </el-tag>
              </div>
            </el-option>
          </el-select>
        </el-form-item>

        <el-divider content-position="left">Configuración de Recursos</el-divider>

        <el-form-item label="RAM Mínima (MB)" prop="ram_min">
          <el-input-number
            v-model="form.ram_min"
            :min="512"
            :max="32768"
            :step="512"
            class="w-full"
          />
        </el-form-item>

        <el-form-item label="RAM Máxima (MB)" prop="ram_max">
          <el-input-number
            v-model="form.ram_max"
            :min="512"
            :max="32768"
            :step="512"
            class="w-full"
          />
        </el-form-item>

        <el-form-item label="Versión de Java" prop="java_version">
          <el-select v-model="form.java_version" placeholder="Selecciona la versión" class="w-full">
            <el-option label="Java 8" value="8" />
            <el-option label="Java 11" value="11" />
            <el-option label="Java 16" value="16" />
            <el-option label="Java 17" value="17" />
            <el-option label="Java 21" value="21" />
          </el-select>
        </el-form-item>

        <el-divider content-position="left">Opciones Avanzadas</el-divider>

        <el-form-item label="Auto-inicio">
          <el-switch v-model="form.auto_start" />
          <span class="ml-2 text-sm text-gray-500">
            Iniciar automáticamente cuando el agente se inicie
          </span>
        </el-form-item>

        <!-- Actions -->
        <el-form-item>
          <div class="flex gap-2">
            <el-button
              type="primary"
              :loading="serversStore.loading"
              :icon="Check"
              @click="handleSubmit"
            >
              Crear Servidor
            </el-button>
            <el-button @click="router.back()">
              Cancelar
            </el-button>
          </div>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useServersStore } from '@/stores/servers';
import { useAgentsStore } from '@/stores/agents';
import { ArrowLeft, Check } from '@element-plus/icons-vue';
import type { FormInstance, FormRules } from 'element-plus';

const router = useRouter();
const serversStore = useServersStore();
const agentsStore = useAgentsStore();
const formRef = ref<FormInstance>();

const form = reactive({
  name: '',
  type: 'paper',
  version: '1.20.1',
  port: 25565,
  agent_id: '',
  ram_min: 1024,
  ram_max: 2048,
  java_version: '17',
  auto_start: false,
});

const validateRam = (rule: any, value: any, callback: any) => {
  if (value < form.ram_min) {
    callback(new Error('La RAM máxima debe ser mayor o igual a la RAM mínima'));
  } else {
    callback();
  }
};

const rules: FormRules = {
  name: [
    { required: true, message: 'Por favor ingresa un nombre', trigger: 'blur' },
    { min: 3, message: 'El nombre debe tener al menos 3 caracteres', trigger: 'blur' },
  ],
  type: [
    { required: true, message: 'Por favor selecciona un tipo', trigger: 'change' },
  ],
  version: [
    { required: true, message: 'Por favor ingresa una versión', trigger: 'blur' },
  ],
  port: [
    { required: true, message: 'Por favor ingresa un puerto', trigger: 'blur' },
  ],
  agent_id: [
    { required: true, message: 'Por favor selecciona un agente', trigger: 'change' },
  ],
  ram_min: [
    { required: true, message: 'Por favor ingresa la RAM mínima', trigger: 'blur' },
  ],
  ram_max: [
    { required: true, message: 'Por favor ingresa la RAM máxima', trigger: 'blur' },
    { validator: validateRam, trigger: 'blur' },
  ],
  java_version: [
    { required: true, message: 'Por favor selecciona una versión de Java', trigger: 'change' },
  ],
};

const handleSubmit = async () => {
  if (!formRef.value) return;

  await formRef.value.validate(async (valid) => {
    if (!valid) return;

    const server = await serversStore.createServer(form);
    
    if (server) {
      router.push(`/servers/${server.id}`);
    }
  });
};

onMounted(async () => {
  await agentsStore.fetchAgents();
  
  // Seleccionar el primer agente online por defecto
  const onlineAgent = agentsStore.agents.find(a => a.status === 'online');
  if (onlineAgent) {
    form.agent_id = onlineAgent.id;
  }
});
</script>
