use anyhow::{Context, Result};
use std::path::PathBuf;
use tauri::Manager;

/// Nombres de los scripts disponibles
#[derive(Debug, Clone)]
pub enum Script {
    InstallVPS,
    ContinueInstall,
    Uninstall,
    Build,
    TestAPI,
}

impl Script {
    /// Retorna el nombre del archivo del script
    pub fn filename(&self) -> &str {
        match self {
            Script::InstallVPS => "install-vps.sh",
            Script::ContinueInstall => "continue-install.sh",
            Script::Uninstall => "uninstall.sh",
            Script::Build => "build.sh",
            Script::TestAPI => "test-api.sh",
        }
    }

    /// Retorna una descripción del script
    pub fn description(&self) -> &str {
        match self {
            Script::InstallVPS => "Instalador completo de AYMC en VPS",
            Script::ContinueInstall => "Continuar instalación si fue interrumpida",
            Script::Uninstall => "Desinstalador de AYMC del VPS",
            Script::Build => "Script de compilación de binarios",
            Script::TestAPI => "Script de prueba de API",
        }
    }
}

/// Manager para acceder a los scripts embebidos
pub struct ScriptManager {
    resource_path: PathBuf,
}

impl ScriptManager {
    /// Crea un nuevo ScriptManager
    pub fn new(app_handle: &tauri::AppHandle) -> Result<Self> {
        let resource_path = app_handle
            .path()
            .resource_dir()
            .context("No se pudo obtener el directorio de recursos")?;

        Ok(Self { resource_path })
    }

    /// Obtiene la ruta completa de un script
    pub fn get_script_path(&self, script: Script) -> PathBuf {
        self.resource_path.join(script.filename())
    }

    /// Lee el contenido de un script
    pub fn read_script(&self, script: Script) -> Result<String> {
        let path = self.get_script_path(script.clone());
        std::fs::read_to_string(&path).context(format!(
            "No se pudo leer el script: {}",
            script.filename()
        ))
    }

    /// Verifica si un script existe
    pub fn script_exists(&self, script: Script) -> bool {
        self.get_script_path(script).exists()
    }

    /// Lista todos los scripts disponibles
    pub fn list_scripts(&self) -> Vec<Script> {
        vec![
            Script::InstallVPS,
            Script::ContinueInstall,
            Script::Uninstall,
            Script::Build,
            Script::TestAPI,
        ]
    }

    /// Obtiene información de todos los scripts
    pub fn get_scripts_info(&self) -> Vec<ScriptInfo> {
        self.list_scripts()
            .iter()
            .map(|script| {
                let exists = self.script_exists(script.clone());
                let size = if exists {
                    self.get_script_path(script.clone())
                        .metadata()
                        .map(|m| m.len())
                        .ok()
                } else {
                    None
                };

                ScriptInfo {
                    name: script.filename().to_string(),
                    description: script.description().to_string(),
                    exists,
                    size_bytes: size,
                }
            })
            .collect()
    }
}

/// Información de un script
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct ScriptInfo {
    pub name: String,
    pub description: String,
    pub exists: bool,
    pub size_bytes: Option<u64>,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_script_filename() {
        assert_eq!(Script::InstallVPS.filename(), "install-vps.sh");
        assert_eq!(Script::ContinueInstall.filename(), "continue-install.sh");
        assert_eq!(Script::Uninstall.filename(), "uninstall.sh");
        assert_eq!(Script::Build.filename(), "build.sh");
        assert_eq!(Script::TestAPI.filename(), "test-api.sh");
    }

    #[test]
    fn test_script_description() {
        assert!(!Script::InstallVPS.description().is_empty());
        assert!(!Script::ContinueInstall.description().is_empty());
        assert!(!Script::Uninstall.description().is_empty());
        assert!(!Script::Build.description().is_empty());
        assert!(!Script::TestAPI.description().is_empty());
    }
}
