use crate::ssh::{BackendConfig, SSHAuth, SSHClient, SSHConfig, ServiceStatus};
use crate::scripts::{Script, ScriptInfo, ScriptManager};
use serde::{Deserialize, Serialize};
use std::sync::Mutex;
use tauri::{AppHandle, State};

/// Estado global para mantener la conexión SSH activa
pub struct SSHState {
    pub client: Mutex<Option<SSHClient>>,
}

impl Default for SSHState {
    fn default() -> Self {
        Self {
            client: Mutex::new(None),
        }
    }
}

/// Respuesta genérica para los comandos
#[derive(Debug, Serialize, Deserialize)]
pub struct CommandResponse<T> {
    pub success: bool,
    pub data: Option<T>,
    pub error: Option<String>,
}

impl<T> CommandResponse<T> {
    pub fn success(data: T) -> Self {
        Self {
            success: true,
            data: Some(data),
            error: None,
        }
    }

    pub fn error(message: String) -> Self {
        Self {
            success: false,
            data: None,
            error: Some(message),
        }
    }
}

/// Conecta al servidor SSH y guarda la sesión
#[tauri::command]
pub async fn ssh_connect(
    state: State<'_, SSHState>,
    host: String,
    port: u16,
    username: String,
    auth_type: String,
    password: Option<String>,
    private_key_path: Option<String>,
    private_key_data: Option<String>,
    passphrase: Option<String>,
) -> Result<CommandResponse<String>, String> {
    // Construir la autenticación según el tipo
    let auth = match auth_type.as_str() {
        "password" => {
            let pwd = password.ok_or("Contraseña requerida para autenticación por password")?;
            SSHAuth::Password { password: pwd }
        }
        "private_key_file" => {
            let key_path = private_key_path
                .ok_or("Ruta de clave privada requerida")?;
            SSHAuth::PrivateKey {
                private_key_path: key_path,
                passphrase,
            }
        }
        "private_key_data" => {
            let key_data = private_key_data
                .ok_or("Datos de clave privada requeridos")?;
            SSHAuth::PrivateKeyData {
                private_key_data: key_data,
                passphrase,
            }
        }
        _ => return Err("Tipo de autenticación inválido".to_string()),
    };

    let config = SSHConfig {
        host: host.clone(),
        port,
        username: username.clone(),
        auth,
    };

    // Intentar conectar
    match SSHClient::connect(config) {
        Ok(client) => {
            // Guardar la conexión en el estado
            let mut state_client = state.client.lock().unwrap();
            *state_client = Some(client);

            Ok(CommandResponse::success(format!(
                "Conectado exitosamente a {}@{}:{}",
                username, host, port
            )))
        }
        Err(e) => Ok(CommandResponse::error(format!(
            "Error al conectar: {}",
            e
        ))),
    }
}

/// Desconecta la sesión SSH actual
#[tauri::command]
pub async fn ssh_disconnect(state: State<'_, SSHState>) -> Result<CommandResponse<String>, String> {
    let mut client = state.client.lock().unwrap();
    
    if client.is_some() {
        *client = None;
        Ok(CommandResponse::success("Desconectado exitosamente".to_string()))
    } else {
        Ok(CommandResponse::error("No hay conexión SSH activa".to_string()))
    }
}

/// Verifica si hay una conexión SSH activa
#[tauri::command]
pub async fn ssh_is_connected(state: State<'_, SSHState>) -> Result<CommandResponse<bool>, String> {
    let client = state.client.lock().unwrap();
    Ok(CommandResponse::success(client.is_some()))
}

/// Ejecuta un comando remoto
#[tauri::command]
pub async fn ssh_execute_command(
    state: State<'_, SSHState>,
    command: String,
) -> Result<CommandResponse<String>, String> {
    let client = state.client.lock().unwrap();
    
    match client.as_ref() {
        Some(ssh_client) => {
            match ssh_client.execute_command(&command) {
                Ok(output) => Ok(CommandResponse::success(output)),
                Err(e) => Ok(CommandResponse::error(format!("Error al ejecutar comando: {}", e))),
            }
        }
        None => Ok(CommandResponse::error("No hay conexión SSH activa".to_string())),
    }
}

/// Verifica el estado de los servicios AYMC en la VPS
#[tauri::command]
pub async fn ssh_check_services(
    state: State<'_, SSHState>,
) -> Result<CommandResponse<ServiceStatus>, String> {
    let client = state.client.lock().unwrap();
    
    match client.as_ref() {
        Some(ssh_client) => {
            match ssh_client.check_services() {
                Ok(status) => Ok(CommandResponse::success(status)),
                Err(e) => Ok(CommandResponse::error(format!(
                    "Error al verificar servicios: {}",
                    e
                ))),
            }
        }
        None => Ok(CommandResponse::error("No hay conexión SSH activa".to_string())),
    }
}

/// Obtiene la configuración del backend desde la VPS
#[tauri::command]
pub async fn ssh_get_backend_config(
    state: State<'_, SSHState>,
) -> Result<CommandResponse<BackendConfig>, String> {
    let client = state.client.lock().unwrap();
    
    match client.as_ref() {
        Some(ssh_client) => {
            match ssh_client.get_backend_config() {
                Ok(config) => Ok(CommandResponse::success(config)),
                Err(e) => Ok(CommandResponse::error(format!(
                    "Error al obtener configuración: {}",
                    e
                ))),
            }
        }
        None => Ok(CommandResponse::error("No hay conexión SSH activa".to_string())),
    }
}

/// Verifica si un archivo existe en la VPS
#[tauri::command]
pub async fn ssh_file_exists(
    state: State<'_, SSHState>,
    path: String,
) -> Result<CommandResponse<bool>, String> {
    let client = state.client.lock().unwrap();
    
    match client.as_ref() {
        Some(ssh_client) => {
            match ssh_client.file_exists(&path) {
                Ok(exists) => Ok(CommandResponse::success(exists)),
                Err(e) => Ok(CommandResponse::error(format!(
                    "Error al verificar archivo: {}",
                    e
                ))),
            }
        }
        None => Ok(CommandResponse::error("No hay conexión SSH activa".to_string())),
    }
}

/// Lee el contenido de un archivo remoto
#[tauri::command]
pub async fn ssh_read_file(
    state: State<'_, SSHState>,
    path: String,
) -> Result<CommandResponse<String>, String> {
    let client = state.client.lock().unwrap();
    
    match client.as_ref() {
        Some(ssh_client) => {
            match ssh_client.read_file(&path) {
                Ok(content) => Ok(CommandResponse::success(content)),
                Err(e) => Ok(CommandResponse::error(format!(
                    "Error al leer archivo: {}",
                    e
                ))),
            }
        }
        None => Ok(CommandResponse::error("No hay conexión SSH activa".to_string())),
    }
}

/// Sube contenido a un archivo remoto
#[tauri::command]
pub async fn ssh_upload_content(
    state: State<'_, SSHState>,
    content: String,
    remote_path: String,
) -> Result<CommandResponse<String>, String> {
    let client = state.client.lock().unwrap();
    
    match client.as_ref() {
        Some(ssh_client) => {
            match ssh_client.upload_content(&content, &remote_path) {
                Ok(_) => Ok(CommandResponse::success(format!(
                    "Archivo subido exitosamente a {}",
                    remote_path
                ))),
                Err(e) => Ok(CommandResponse::error(format!(
                    "Error al subir archivo: {}",
                    e
                ))),
            }
        }
        None => Ok(CommandResponse::error("No hay conexión SSH activa".to_string())),
    }
}

/// Obtiene información del sistema operativo remoto
#[tauri::command]
pub async fn ssh_get_host_info(
    state: State<'_, SSHState>,
) -> Result<CommandResponse<String>, String> {
    let client = state.client.lock().unwrap();
    
    match client.as_ref() {
        Some(ssh_client) => {
            match ssh_client.get_host_info() {
                Ok(info) => Ok(CommandResponse::success(info)),
                Err(e) => Ok(CommandResponse::error(format!(
                    "Error al obtener información del host: {}",
                    e
                ))),
            }
        }
        None => Ok(CommandResponse::error("No hay conexión SSH activa".to_string())),
    }
}

/// Verifica si el usuario tiene acceso sudo
#[tauri::command]
pub async fn ssh_has_sudo(
    state: State<'_, SSHState>,
) -> Result<CommandResponse<bool>, String> {
    let client = state.client.lock().unwrap();
    
    match client.as_ref() {
        Some(ssh_client) => {
            match ssh_client.has_sudo_access() {
                Ok(has_sudo) => Ok(CommandResponse::success(has_sudo)),
                Err(e) => Ok(CommandResponse::error(format!(
                    "Error al verificar sudo: {}",
                    e
                ))),
            }
        }
        None => Ok(CommandResponse::error("No hay conexión SSH activa".to_string())),
    }
}

/// Ejecuta un comando y retorna las líneas en tiempo real
#[tauri::command]
pub async fn ssh_execute_streaming(
    state: State<'_, SSHState>,
    command: String,
) -> Result<CommandResponse<Vec<String>>, String> {
    let client = state.client.lock().unwrap();
    
    match client.as_ref() {
        Some(ssh_client) => {
            match ssh_client.execute_command_streaming(&command) {
                Ok(lines) => Ok(CommandResponse::success(lines)),
                Err(e) => Ok(CommandResponse::error(format!(
                    "Error al ejecutar comando: {}",
                    e
                ))),
            }
        }
        None => Ok(CommandResponse::error("No hay conexión SSH activa".to_string())),
    }
}

// ============================================================================
// COMANDOS DE SCRIPTS
// ============================================================================

/// Lista todos los scripts disponibles embebidos
#[tauri::command]
pub async fn list_embedded_scripts(
    app: AppHandle,
) -> Result<CommandResponse<Vec<ScriptInfo>>, String> {
    match ScriptManager::new(&app) {
        Ok(manager) => {
            let scripts = manager.get_scripts_info();
            Ok(CommandResponse::success(scripts))
        }
        Err(e) => Ok(CommandResponse::error(format!(
            "Error al listar scripts: {}",
            e
        ))),
    }
}

/// Lee el contenido de un script embebido
#[tauri::command]
pub async fn read_embedded_script(
    app: AppHandle,
    script_name: String,
) -> Result<CommandResponse<String>, String> {
    let script = match script_name.as_str() {
        "install-vps.sh" => Script::InstallVPS,
        "continue-install.sh" => Script::ContinueInstall,
        "uninstall.sh" => Script::Uninstall,
        "build.sh" => Script::Build,
        "test-api.sh" => Script::TestAPI,
        _ => {
            return Ok(CommandResponse::error(format!(
                "Script desconocido: {}",
                script_name
            )))
        }
    };

    match ScriptManager::new(&app) {
        Ok(manager) => match manager.read_script(script) {
            Ok(content) => Ok(CommandResponse::success(content)),
            Err(e) => Ok(CommandResponse::error(format!(
                "Error al leer script: {}",
                e
            ))),
        },
        Err(e) => Ok(CommandResponse::error(format!(
            "Error al acceder a scripts: {}",
            e
        ))),
    }
}

/// Instala el backend en la VPS usando SSH
/// Este comando:
/// 1. Sube el script install-vps.sh a la VPS
/// 2. Le da permisos de ejecución
/// 3. Lo ejecuta con las credenciales proporcionadas
/// 4. Retorna el output en tiempo real
#[tauri::command]
pub async fn ssh_install_backend(
    app: AppHandle,
    state: State<'_, SSHState>,
    db_password: String,
    jwt_secret: String,
    app_port: Option<String>,
) -> Result<CommandResponse<Vec<String>>, String> {
    let client = state.client.lock().unwrap();

    match client.as_ref() {
        Some(ssh_client) => {
            // Leer el script embebido
            let manager = ScriptManager::new(&app).map_err(|e| e.to_string())?;
            let script_content = manager
                .read_script(Script::InstallVPS)
                .map_err(|e| e.to_string())?;

            // Subir el script a /tmp/install-aymc.sh
            match ssh_client.upload_content(&script_content, "/tmp/install-aymc.sh") {
                Ok(_) => {
                    // Dar permisos de ejecución
                    if let Err(e) = ssh_client.execute_command("chmod +x /tmp/install-aymc.sh") {
                        return Ok(CommandResponse::error(format!(
                            "Error al dar permisos al script: {}",
                            e
                        )));
                    }

                    // Construir el comando de instalación
                    let port = app_port.unwrap_or_else(|| "8080".to_string());
                    let install_cmd = format!(
                        "DB_PASSWORD='{}' JWT_SECRET='{}' APP_PORT='{}' /tmp/install-aymc.sh",
                        db_password, jwt_secret, port
                    );

                    // Ejecutar la instalación con streaming
                    match ssh_client.execute_command_streaming(&install_cmd) {
                        Ok(lines) => Ok(CommandResponse::success(lines)),
                        Err(e) => Ok(CommandResponse::error(format!(
                            "Error durante la instalación: {}",
                            e
                        ))),
                    }
                }
                Err(e) => Ok(CommandResponse::error(format!(
                    "Error al subir script: {}",
                    e
                ))),
            }
        }
        None => Ok(CommandResponse::error(
            "No hay conexión SSH activa".to_string(),
        )),
    }
}

/// Desinstala AYMC de la VPS usando el script uninstall.sh
#[tauri::command]
pub async fn ssh_uninstall_backend(
    app: AppHandle,
    state: State<'_, SSHState>,
) -> Result<CommandResponse<Vec<String>>, String> {
    let client = state.client.lock().unwrap();

    match client.as_ref() {
        Some(ssh_client) => {
            // Leer el script embebido
            let manager = ScriptManager::new(&app).map_err(|e| e.to_string())?;
            let script_content = manager
                .read_script(Script::Uninstall)
                .map_err(|e| e.to_string())?;

            // Subir el script
            match ssh_client.upload_content(&script_content, "/tmp/uninstall-aymc.sh") {
                Ok(_) => {
                    // Dar permisos de ejecución
                    if let Err(e) = ssh_client.execute_command("chmod +x /tmp/uninstall-aymc.sh")
                    {
                        return Ok(CommandResponse::error(format!(
                            "Error al dar permisos al script: {}",
                            e
                        )));
                    }

                    // Ejecutar desinstalación
                    match ssh_client.execute_command_streaming("/tmp/uninstall-aymc.sh") {
                        Ok(lines) => Ok(CommandResponse::success(lines)),
                        Err(e) => Ok(CommandResponse::error(format!(
                            "Error durante la desinstalación: {}",
                            e
                        ))),
                    }
                }
                Err(e) => Ok(CommandResponse::error(format!(
                    "Error al subir script: {}",
                    e
                ))),
            }
        }
        None => Ok(CommandResponse::error(
            "No hay conexión SSH activa".to_string(),
        )),
    }
}

// ============================================
// FASE 6: Comandos de Validación y Diagnóstico
// ============================================

/// Información de espacio en disco
#[derive(Debug, Serialize, Deserialize)]
pub struct DiskSpace {
    pub total_mb: u64,
    pub used_mb: u64,
    pub available_mb: u64,
    pub percent_used: u8,
}

/// Verifica si un puerto específico está disponible
#[tauri::command]
pub async fn ssh_check_port_available(
    state: State<'_, SSHState>,
    port: u16,
) -> Result<bool, String> {
    let client = state.client.lock().unwrap();

    match client.as_ref() {
        Some(ssh_client) => {
            // Verificar si el puerto está en uso
            let command = format!("netstat -tuln | grep :{} || echo 'AVAILABLE'", port);
            match ssh_client.execute_command(&command) {
                Ok(output) => Ok(output.contains("AVAILABLE")),
                Err(_) => Ok(true), // Asumir disponible si no se puede verificar
            }
        }
        None => Err("No hay conexión SSH activa".to_string()),
    }
}

/// Obtiene información sobre el espacio en disco disponible
#[tauri::command]
pub async fn ssh_get_disk_space(state: State<'_, SSHState>) -> Result<DiskSpace, String> {
    let client = state.client.lock().unwrap();

    match client.as_ref() {
        Some(ssh_client) => {
            // Obtener información del disco raíz
            let output = ssh_client
                .execute_command("df -m / | tail -1")
                .map_err(|e| format!("Error obteniendo espacio en disco: {}", e))?;

            // Parsear output: Filesystem  1M-blocks  Used  Available Use% Mounted
            let parts: Vec<&str> = output.split_whitespace().collect();
            if parts.len() >= 5 {
                let total = parts[1].parse::<u64>().unwrap_or(0);
                let used = parts[2].parse::<u64>().unwrap_or(0);
                let available = parts[3].parse::<u64>().unwrap_or(0);
                
                // Calcular porcentaje usado
                let percent_used = if total > 0 {
                    ((used as f64 / total as f64) * 100.0) as u8
                } else {
                    0
                };

                Ok(DiskSpace {
                    total_mb: total,
                    used_mb: used,
                    available_mb: available,
                    percent_used,
                })
            } else {
                Err("No se pudo parsear información de disco".to_string())
            }
        }
        None => Err("No hay conexión SSH activa".to_string()),
    }
}

/// Verifica si Docker está instalado y corriendo
#[tauri::command]
pub async fn ssh_check_docker(state: State<'_, SSHState>) -> Result<bool, String> {
    let client = state.client.lock().unwrap();

    match client.as_ref() {
        Some(ssh_client) => {
            // Verificar si docker está instalado
            let docker_installed = ssh_client
                .execute_command("which docker")
                .is_ok();

            if !docker_installed {
                return Ok(false);
            }

            // Verificar si docker está corriendo
            let docker_running = ssh_client
                .execute_command("docker ps")
                .is_ok();

            Ok(docker_running)
        }
        None => Err("No hay conexión SSH activa".to_string()),
    }
}

/// Obtiene logs del sistema relacionados con AYMC
#[tauri::command]
pub async fn ssh_get_system_logs(
    state: State<'_, SSHState>,
    service: String,
    lines: u32,
) -> Result<Vec<String>, String> {
    let client = state.client.lock().unwrap();

    match client.as_ref() {
        Some(ssh_client) => {
            let command = format!("journalctl -u {} -n {} --no-pager", service, lines);
            
            match ssh_client.execute_command(&command) {
                Ok(output) => {
                    let log_lines: Vec<String> = output
                        .lines()
                        .map(|line| line.to_string())
                        .collect();
                    Ok(log_lines)
                }
                Err(e) => Err(format!("Error obteniendo logs: {}", e)),
            }
        }
        None => Err("No hay conexión SSH activa".to_string()),
    }
}

