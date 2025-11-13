import { invoke } from "@tauri-apps/api/core";

/**
 * Servicio de Instalación Remota con manejo de errores y reintentos
 * Fase 6: Instalación Remota Avanzada
 */

export interface InstallationCredentials {
  dbPassword: string;
  jwtSecret: string;
  appPort: number;
  dbName?: string;
  dbUser?: string;
}

export interface InstallationResult {
  success: boolean;
  api_url: string;
  ws_url: string;
  message: string;
  error?: string;
}

export interface PreRequisiteCheck {
  name: string;
  description: string;
  passed: boolean;
  required: boolean;
  error?: string;
}

export interface InstallationStep {
  id: number;
  name: string;
  description: string;
  status: "pending" | "running" | "completed" | "failed" | "skipped";
  progress: number;
  startTime?: number;
  endTime?: number;
  error?: string;
  canRetry: boolean;
}

export type InstallationPhase =
  | "validation"
  | "preparation"
  | "installation"
  | "configuration"
  | "verification"
  | "completed"
  | "failed";

export interface InstallationProgress {
  phase: InstallationPhase;
  currentStep: number;
  totalSteps: number;
  percentage: number;
  message: string;
  steps: InstallationStep[];
}

/**
 * Clase principal del servicio de instalación
 */
export class RemoteInstallationService {
  private maxRetries = 3;
  private retryDelay = 2000; // 2 segundos
  private currentAttempt = 0;
  private abortController: AbortController | null = null;
  private progressCallback: ((progress: InstallationProgress) => void) | null =
    null;
  private logCallback:
    | ((
        message: string,
        type: "info" | "error" | "warning" | "success"
      ) => void)
    | null = null;

  /**
   * Configurar callback de progreso
   */
  onProgress(callback: (progress: InstallationProgress) => void) {
    this.progressCallback = callback;
  }

  /**
   * Configurar callback de logs
   */
  onLog(
    callback: (
      message: string,
      type: "info" | "error" | "warning" | "success"
    ) => void
  ) {
    this.logCallback = callback;
  }

  /**
   * Emitir progreso
   */
  private emitProgress(progress: InstallationProgress) {
    if (this.progressCallback) {
      this.progressCallback(progress);
    }
  }

  /**
   * Emitir log
   */
  private log(
    message: string,
    type: "info" | "error" | "warning" | "success" = "info"
  ) {
    console.log(`[Installation] ${type.toUpperCase()}: ${message}`);
    if (this.logCallback) {
      this.logCallback(message, type);
    }
  }

  /**
   * Validar pre-requisitos antes de instalar
   */
  async validatePreRequisites(): Promise<PreRequisiteCheck[]> {
    this.log("Validando pre-requisitos...", "info");

    const checks: PreRequisiteCheck[] = [];

    try {
      // Check 1: Conexión SSH activa
      try {
        const isConnected = await invoke<boolean>("ssh_is_connected");
        checks.push({
          name: "ssh_connection",
          description: "Conexión SSH activa",
          passed: isConnected,
          required: true,
          error: isConnected ? undefined : "No hay conexión SSH activa",
        });
      } catch (error: any) {
        checks.push({
          name: "ssh_connection",
          description: "Conexión SSH activa",
          passed: false,
          required: true,
          error: error.message || "Error verificando SSH",
        });
      }

      // Check 2: Permisos sudo
      try {
        const hasSudo = await invoke<boolean>("ssh_has_sudo");
        checks.push({
          name: "sudo_permissions",
          description: "Permisos de administrador (sudo)",
          passed: hasSudo,
          required: true,
          error: hasSudo ? undefined : "Usuario no tiene permisos sudo",
        });
      } catch (error: any) {
        checks.push({
          name: "sudo_permissions",
          description: "Permisos de administrador (sudo)",
          passed: false,
          required: true,
          error: error.message || "Error verificando sudo",
        });
      }

      // Check 3: Puerto disponible
      try {
        const portAvailable = await this.checkPortAvailable(8080);
        checks.push({
          name: "port_available",
          description: "Puerto 8080 disponible",
          passed: portAvailable,
          required: true,
          error: portAvailable ? undefined : "Puerto 8080 ya está en uso",
        });
      } catch (error: any) {
        checks.push({
          name: "port_available",
          description: "Puerto 8080 disponible",
          passed: false,
          required: false,
          error: error.message || "Error verificando puerto",
        });
      }

      // Check 4: Espacio en disco (mínimo 2GB)
      try {
        const hasSpace = await this.checkDiskSpace(2048); // 2GB en MB
        checks.push({
          name: "disk_space",
          description: "Espacio en disco suficiente (2GB)",
          passed: hasSpace,
          required: true,
          error: hasSpace ? undefined : "Espacio en disco insuficiente",
        });
      } catch (error: any) {
        checks.push({
          name: "disk_space",
          description: "Espacio en disco suficiente (2GB)",
          passed: false,
          required: false,
          error: "No se pudo verificar espacio en disco",
        });
      }

      // Check 5: Sistema operativo compatible
      try {
        const osInfo = await invoke<{ os: string; version: string }>(
          "ssh_get_host_info"
        );
        const isCompatible = ["ubuntu", "debian", "centos", "rhel"].some((os) =>
          osInfo.os.toLowerCase().includes(os)
        );
        checks.push({
          name: "os_compatible",
          description: "Sistema operativo compatible",
          passed: isCompatible,
          required: false,
          error: isCompatible
            ? undefined
            : `SO no oficialmente soportado: ${osInfo.os}`,
        });
      } catch (error: any) {
        checks.push({
          name: "os_compatible",
          description: "Sistema operativo compatible",
          passed: false,
          required: false,
          error: "No se pudo detectar el sistema operativo",
        });
      }

      const requiredChecksPassed = checks
        .filter((check) => check.required)
        .every((check) => check.passed);

      if (requiredChecksPassed) {
        this.log(
          "✓ Todos los pre-requisitos necesarios están cumplidos",
          "success"
        );
      } else {
        this.log("✗ Algunos pre-requisitos necesarios no se cumplen", "error");
      }

      return checks;
    } catch (error: any) {
      this.log(
        `Error en validación de pre-requisitos: ${error.message}`,
        "error"
      );
      throw error;
    }
  }

  /**
   * Verificar si un puerto está disponible
   */
  private async checkPortAvailable(port: number): Promise<boolean> {
    try {
      const output = await invoke<string>("ssh_execute_command", {
        command: `netstat -tuln | grep :${port} || echo "AVAILABLE"`,
      });
      return output.includes("AVAILABLE");
    } catch {
      return true; // Asumir disponible si no se puede verificar
    }
  }

  /**
   * Verificar espacio en disco
   */
  private async checkDiskSpace(minMB: number): Promise<boolean> {
    try {
      const output = await invoke<string>("ssh_execute_command", {
        command: `df -m / | tail -1 | awk '{print $4}'`,
      });
      const availableMB = parseInt(output.trim());
      return availableMB >= minMB;
    } catch {
      return true; // Asumir suficiente si no se puede verificar
    }
  }

  /**
   * Instalar AYMC con reintentos automáticos
   */
  async install(
    credentials: InstallationCredentials,
    options: {
      maxRetries?: number;
      retryDelay?: number;
      validateFirst?: boolean;
    } = {}
  ): Promise<InstallationResult> {
    // Configurar opciones
    this.maxRetries = options.maxRetries ?? 3;
    this.retryDelay = options.retryDelay ?? 2000;
    this.currentAttempt = 0;
    this.abortController = new AbortController();

    // Validar pre-requisitos si está habilitado
    if (options.validateFirst !== false) {
      this.log("Validando pre-requisitos antes de instalar...", "info");
      const checks = await this.validatePreRequisites();

      const requiredFailed = checks.filter(
        (check) => check.required && !check.passed
      );
      if (requiredFailed.length > 0) {
        const errors = requiredFailed.map((check) => check.error).join(", ");
        throw new Error(`Pre-requisitos no cumplidos: ${errors}`);
      }
    }

    // Intentar instalación con reintentos
    return await this.installWithRetry(credentials);
  }

  /**
   * Instalar con reintentos automáticos
   */
  private async installWithRetry(
    credentials: InstallationCredentials
  ): Promise<InstallationResult> {
    while (this.currentAttempt < this.maxRetries) {
      this.currentAttempt++;

      try {
        this.log(
          `Intento ${this.currentAttempt} de ${this.maxRetries}...`,
          "info"
        );

        const result = await this.executeInstallation(credentials);

        if (result.success) {
          this.log("✓ Instalación completada exitosamente", "success");
          return result;
        } else {
          throw new Error(
            result.message || "Error desconocido en la instalación"
          );
        }
      } catch (error: any) {
        this.log(
          `✗ Intento ${this.currentAttempt} falló: ${error.message}`,
          "error"
        );

        // Si es el último intento, lanzar el error
        if (this.currentAttempt >= this.maxRetries) {
          this.log("✗ Se agotaron todos los intentos", "error");
          throw new Error(
            `Instalación falló después de ${this.maxRetries} intentos: ${error.message}`
          );
        }

        // Esperar antes de reintentar
        this.log(
          `Esperando ${this.retryDelay / 1000}s antes de reintentar...`,
          "warning"
        );
        await this.sleep(this.retryDelay);
      }
    }

    throw new Error("Instalación falló: se agotaron los intentos");
  }

  /**
   * Ejecutar la instalación
   */
  private async executeInstallation(
    credentials: InstallationCredentials
  ): Promise<InstallationResult> {
    try {
      // Crear progreso inicial
      const progress: InstallationProgress = {
        phase: "preparation",
        currentStep: 0,
        totalSteps: 5,
        percentage: 0,
        message: "Preparando instalación...",
        steps: this.createInstallationSteps(),
      };

      this.emitProgress(progress);

      // Verificar conexión SSH
      this.log("Verificando conexión SSH...", "info");
      const isConnected = await invoke<boolean>("ssh_is_connected");
      if (!isConnected) {
        throw new Error("No hay conexión SSH activa");
      }

      this.updateStepStatus(progress, 0, "completed");
      this.emitProgress(progress);

      // Ejecutar instalación vía comando Tauri
      this.log("Iniciando instalación remota...", "info");
      progress.phase = "installation";
      progress.message = "Ejecutando instalación en VPS...";
      this.updateStepStatus(progress, 1, "running");
      this.emitProgress(progress);

      const result = await invoke<InstallationResult>("ssh_install_backend", {
        dbPassword: credentials.dbPassword,
        jwtSecret: credentials.jwtSecret,
        appPort: credentials.appPort,
      });

      if (result.success) {
        this.updateStepStatus(progress, 1, "completed");
        progress.phase = "verification";
        progress.message = "Verificando instalación...";
        this.emitProgress(progress);

        // Verificar servicios instalados
        const verified = await this.verifyInstallation();
        if (verified) {
          this.updateStepStatus(progress, 2, "completed");
          progress.phase = "completed";
          progress.percentage = 100;
          progress.message = "Instalación completada exitosamente";
          this.emitProgress(progress);
        }

        return result;
      } else {
        throw new Error(result.message || "Error en la instalación");
      }
    } catch (error: any) {
      this.log(`Error en instalación: ${error.message}`, "error");
      throw error;
    }
  }

  /**
   * Crear pasos de instalación
   */
  private createInstallationSteps(): InstallationStep[] {
    return [
      {
        id: 1,
        name: "Validación",
        description: "Verificar conexión SSH y pre-requisitos",
        status: "pending",
        progress: 0,
        canRetry: true,
      },
      {
        id: 2,
        name: "Instalación",
        description: "Ejecutar script de instalación en VPS",
        status: "pending",
        progress: 0,
        canRetry: true,
      },
      {
        id: 3,
        name: "Verificación",
        description: "Verificar que los servicios estén corriendo",
        status: "pending",
        progress: 0,
        canRetry: true,
      },
      {
        id: 4,
        name: "Configuración",
        description: "Configurar API URL y WebSocket",
        status: "pending",
        progress: 0,
        canRetry: false,
      },
      {
        id: 5,
        name: "Finalización",
        description: "Completar instalación y guardar configuración",
        status: "pending",
        progress: 0,
        canRetry: false,
      },
    ];
  }

  /**
   * Actualizar estado de un paso
   */
  private updateStepStatus(
    progress: InstallationProgress,
    stepIndex: number,
    status: InstallationStep["status"],
    error?: string
  ) {
    if (progress.steps[stepIndex]) {
      progress.steps[stepIndex].status = status;

      if (status === "running") {
        progress.steps[stepIndex].startTime = Date.now();
      } else if (status === "completed" || status === "failed") {
        progress.steps[stepIndex].endTime = Date.now();
      }

      if (error) {
        progress.steps[stepIndex].error = error;
      }

      // Calcular progreso general
      const completedSteps = progress.steps.filter(
        (s) => s.status === "completed"
      ).length;
      progress.percentage = Math.round(
        (completedSteps / progress.steps.length) * 100
      );
    }
  }

  /**
   * Verificar que la instalación fue exitosa
   */
  private async verifyInstallation(): Promise<boolean> {
    try {
      this.log("Verificando servicios instalados...", "info");

      const status = await invoke<{
        backend_installed: boolean;
        agent_installed: boolean;
        backend_running: boolean;
        agent_running: boolean;
        postgresql_running: boolean;
      }>("ssh_check_services");

      const allRunning =
        status.backend_installed &&
        status.agent_installed &&
        status.backend_running &&
        status.agent_running &&
        status.postgresql_running;

      if (allRunning) {
        this.log("✓ Todos los servicios están corriendo", "success");
      } else {
        this.log(
          "⚠ Algunos servicios no están corriendo correctamente",
          "warning"
        );
      }

      return allRunning;
    } catch (error: any) {
      this.log(`Error verificando instalación: ${error.message}`, "error");
      return false;
    }
  }

  /**
   * Cancelar instalación en progreso
   */
  abort() {
    if (this.abortController) {
      this.abortController.abort();
      this.log("Instalación cancelada por el usuario", "warning");
    }
  }

  /**
   * Reintentar un paso específico
   */
  async retryStep(
    stepId: number,
    credentials: InstallationCredentials
  ): Promise<boolean> {
    this.log(`Reintentando paso ${stepId}...`, "info");

    try {
      // Lógica para reintentar paso específico
      // Por ahora, reintentar toda la instalación
      const result = await this.executeInstallation(credentials);
      return result.success;
    } catch (error: any) {
      this.log(`Error reintentando paso: ${error.message}`, "error");
      return false;
    }
  }

  /**
   * Utilidad: Sleep
   */
  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  /**
   * Obtener información de diagnóstico
   */
  async getDiagnostics(): Promise<string> {
    try {
      const diagnostics: string[] = [];

      // SSH Status
      const sshConnected = await invoke<boolean>("ssh_is_connected");
      diagnostics.push(`SSH Connected: ${sshConnected ? "Yes" : "No"}`);

      // Host Info
      try {
        const hostInfo = await invoke<{ os: string; version: string }>(
          "ssh_get_host_info"
        );
        diagnostics.push(`OS: ${hostInfo.os} ${hostInfo.version}`);
      } catch {
        diagnostics.push("OS: Unable to detect");
      }

      // Services Status
      try {
        const services = await invoke<any>("ssh_check_services");
        diagnostics.push(
          `Backend: ${
            services.backend_installed ? "Installed" : "Not installed"
          }`
        );
        diagnostics.push(
          `Agent: ${services.agent_installed ? "Installed" : "Not installed"}`
        );
        diagnostics.push(
          `PostgreSQL: ${
            services.postgresql_running ? "Running" : "Not running"
          }`
        );
      } catch {
        diagnostics.push("Services: Unable to check");
      }

      return diagnostics.join("\n");
    } catch (error: any) {
      return `Error getting diagnostics: ${error.message}`;
    }
  }
}

/**
 * Instancia singleton del servicio
 */
export const installationService = new RemoteInstallationService();
