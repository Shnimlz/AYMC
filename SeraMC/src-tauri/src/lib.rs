// Módulos personalizados
mod ssh;
mod commands;
mod scripts;

use commands::SSHState;

// Learn more about Tauri commands at https://tauri.app/develop/calling-rust/
#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}! You've been greeted from Rust!", name)
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .manage(SSHState::default()) // Estado global para SSH
        .invoke_handler(tauri::generate_handler![
            greet,
            // Comandos SSH
            commands::ssh_connect,
            commands::ssh_disconnect,
            commands::ssh_is_connected,
            commands::ssh_execute_command,
            commands::ssh_check_services,
            commands::ssh_get_backend_config,
            commands::ssh_file_exists,
            commands::ssh_read_file,
            commands::ssh_upload_content,
            commands::ssh_get_host_info,
            commands::ssh_has_sudo,
            commands::ssh_execute_streaming,
            // Comandos de Scripts
            commands::list_embedded_scripts,
            commands::read_embedded_script,
            commands::ssh_install_backend,
            commands::ssh_uninstall_backend,
            // Fase 6: Comandos de Validación
            commands::ssh_check_port_available,
            commands::ssh_get_disk_space,
            commands::ssh_check_docker,
            commands::ssh_get_system_logs,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

