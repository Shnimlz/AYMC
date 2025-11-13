<template>
  <div class="space-y-6 max-w-4xl">
    <!-- Header -->
    <div>
      <el-button :icon="ArrowLeft" @click="router.back()" class="mb-4">
        Volver
      </el-button>
      <h1 class="text-2xl font-bold text-gray-800">Configuración de Respaldos</h1>
      <p class="text-gray-600 mt-1">Configura los respaldos automáticos para tus servidores</p>
    </div>

    <!-- Server selector -->
    <el-card>
      <el-select
        v-model="selectedServerId"
        placeholder="Selecciona un servidor"
        class="w-full"
        :loading="serversStore.loading"
        @change="loadConfig"
        size="large"
      >
        <el-option
          v-for="server in serversStore.servers"
          :key="server.id"
          :label="server.name"
          :value="server.id"
        />
      </el-select>
    </el-card>

    <!-- Config form -->
    <el-card v-if="selectedServerId && config">
      <el-form
        ref="formRef"
        :model="config"
        :rules="rules"
        label-width="200px"
        label-position="left"
      >
        <el-divider content-position="left">General</el-divider>

        <el-form-item label="Respaldos Automáticos" prop="enabled">
          <el-switch v-model="config.enabled" />
          <span class="ml-2 text-sm text-gray-500">
            Habilitar respaldos programados
          </span>
        </el-form-item>

        <el-form-item label="Programación (Cron)" prop="schedule">
          <el-input
            v-model="config.schedule"
            placeholder="Ej: 0 2 * * * (diario a las 2 AM)"
            :disabled="!config.enabled"
          >
            <template #append>
              <el-dropdown @command="handleCronPreset">
                <el-button>
                  Presets
                  <el-icon class="ml-1"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="0 2 * * *">Diario (2 AM)</el-dropdown-item>
                    <el-dropdown-item command="0 2 * * 0">Semanal (Domingo 2 AM)</el-dropdown-item>
                    <el-dropdown-item command="0 */6 * * *">Cada 6 horas</el-dropdown-item>
                    <el-dropdown-item command="0 */12 * * *">Cada 12 horas</el-dropdown-item>
                    <el-dropdown-item command="0 0 1 * *">Mensual (día 1)</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </el-input>
          <div class="text-sm text-gray-500 mt-1">
            Formato cron: minuto hora día mes día-semana
          </div>
        </el-form-item>

        <el-divider content-position="left">Retención</el-divider>

        <el-form-item label="Máximo de Respaldos" prop="max_backups">
          <el-input-number
            v-model="config.max_backups"
            :min="1"
            :max="100"
            class="w-full"
          />
          <div class="text-sm text-gray-500 mt-1">
            Los respaldos más antiguos se eliminarán automáticamente
          </div>
        </el-form-item>

        <el-form-item label="Días de Retención" prop="retention_days">
          <el-input-number
            v-model="config.retention_days"
            :min="1"
            :max="365"
            class="w-full"
          />
          <div class="text-sm text-gray-500 mt-1">
            Los respaldos más viejos que este período se eliminarán
          </div>
        </el-form-item>

        <el-divider content-position="left">Contenido a Respaldar</el-divider>

        <el-form-item label="Incluir Mundo">
          <el-switch v-model="config.include_world" />
        </el-form-item>

        <el-form-item label="Incluir Plugins">
          <el-switch v-model="config.include_plugins" />
        </el-form-item>

        <el-form-item label="Incluir Configuración">
          <el-switch v-model="config.include_config" />
        </el-form-item>

        <el-form-item label="Incluir Logs">
          <el-switch v-model="config.include_logs" />
        </el-form-item>

        <el-divider content-position="left">Exclusiones</el-divider>

        <el-form-item label="Rutas Excluidas">
          <div class="w-full space-y-2">
            <div
              v-for="(path, index) in config.exclude_paths"
              :key="index"
              class="flex gap-2"
            >
              <el-input
                v-model="config.exclude_paths[index]"
                placeholder="Ej: world/playerdata"
              />
              <el-button
                type="danger"
                :icon="Delete"
                @click="removeExcludePath(index)"
              />
            </div>
            <el-button
              type="primary"
              :icon="Plus"
              @click="addExcludePath"
            >
              Añadir Ruta
            </el-button>
          </div>
          <div class="text-sm text-gray-500 mt-1">
            Rutas relativas al directorio del servidor que no se incluirán
          </div>
        </el-form-item>

        <el-divider />

        <!-- Actions -->
        <el-form-item>
          <div class="flex gap-2">
            <el-button
              type="primary"
              :loading="backupsStore.loading"
              :icon="Check"
              @click="handleSave"
            >
              Guardar Configuración
            </el-button>
            <el-button @click="loadConfig">
              Cancelar
            </el-button>
          </div>
        </el-form-item>
      </el-form>
    </el-card>

    <el-empty
      v-else-if="!selectedServerId"
      description="Selecciona un servidor para configurar sus respaldos"
    />

    <!-- Info card -->
    <el-card>
      <template #header>
        <div class="flex items-center gap-2">
          <el-icon><InfoFilled /></el-icon>
          <span class="font-semibold">Información sobre Cron</span>
        </div>
      </template>

      <div class="space-y-2 text-sm text-gray-600">
        <p><strong>Formato:</strong> minuto hora día mes día-semana</p>
        <p><strong>Ejemplos:</strong></p>
        <ul class="list-disc list-inside pl-4 space-y-1">
          <li><code class="bg-gray-100 px-2 py-1 rounded">0 2 * * *</code> - Diariamente a las 2:00 AM</li>
          <li><code class="bg-gray-100 px-2 py-1 rounded">0 */4 * * *</code> - Cada 4 horas</li>
          <li><code class="bg-gray-100 px-2 py-1 rounded">0 0 * * 0</code> - Cada domingo a medianoche</li>
          <li><code class="bg-gray-100 px-2 py-1 rounded">30 3 1 * *</code> - El día 1 de cada mes a las 3:30 AM</li>
        </ul>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useBackupsStore } from '@/stores/backups';
import { useServersStore } from '@/stores/servers';
import {
  ArrowLeft,
  ArrowDown,
  Plus,
  Delete,
  Check,
  InfoFilled,
} from '@element-plus/icons-vue';
import type { FormInstance, FormRules } from 'element-plus';

const router = useRouter();
const backupsStore = useBackupsStore();
const serversStore = useServersStore();
const formRef = ref<FormInstance>();

const selectedServerId = ref('');
const config = reactive({
  enabled: false,
  schedule: '0 2 * * *',
  max_backups: 10,
  retention_days: 30,
  include_world: true,
  include_plugins: true,
  include_config: true,
  include_logs: false,
  exclude_paths: [] as string[],
});

const validateCron = (rule: any, value: any, callback: any) => {
  // Validación básica de formato cron (5 campos)
  const cronRegex = /^(\*|[0-9,\-*/]+)\s+(\*|[0-9,\-*/]+)\s+(\*|[0-9,\-*/]+)\s+(\*|[0-9,\-*/]+)\s+(\*|[0-9,\-*/]+)$/;
  if (!cronRegex.test(value)) {
    callback(new Error('Formato cron inválido'));
  } else {
    callback();
  }
};

const rules: FormRules = {
  schedule: [
    { required: true, message: 'Por favor ingresa una programación', trigger: 'blur' },
    { validator: validateCron, trigger: 'blur' },
  ],
  max_backups: [
    { required: true, message: 'Por favor ingresa el máximo de respaldos', trigger: 'blur' },
  ],
  retention_days: [
    { required: true, message: 'Por favor ingresa los días de retención', trigger: 'blur' },
  ],
};

const loadConfig = async () => {
  if (!selectedServerId.value) return;
  
  const data = await backupsStore.fetchConfig(selectedServerId.value);
  if (data) {
    Object.assign(config, data);
  }
};

const handleCronPreset = (command: string) => {
  config.schedule = command;
};

const addExcludePath = () => {
  config.exclude_paths.push('');
};

const removeExcludePath = (index: number) => {
  config.exclude_paths.splice(index, 1);
};

const handleSave = async () => {
  if (!formRef.value) return;

  await formRef.value.validate(async (valid) => {
    if (!valid) return;

    // Filtrar rutas vacías
    const cleanPaths = config.exclude_paths.filter(p => p.trim() !== '');
    
    await backupsStore.updateConfig(selectedServerId.value, {
      ...config,
      exclude_paths: cleanPaths,
    });
  });
};

onMounted(async () => {
  await serversStore.fetchServers();
  
  // Seleccionar el primer servidor por defecto
  if (serversStore.servers.length > 0) {
    selectedServerId.value = serversStore.servers[0].id;
    await loadConfig();
  }
});
</script>
