use anyhow::{anyhow, Context, Result};
use serde::{Deserialize, Serialize};
use ssh2::Session;
use std::io::{Read, Write};
use std::net::TcpStream;
use std::path::Path;

/// Representa los diferentes métodos de autenticación SSH
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub enum SSHAuth {
    /// Autenticación con contraseña
    Password { password: String },
    /// Autenticación con clave privada (con contraseña opcional)
    PrivateKey {
        private_key_path: String,
        passphrase: Option<String>,
    },
    /// Autenticación con clave privada desde string (para claves embebidas)
    PrivateKeyData {
        private_key_data: String,
        passphrase: Option<String>,
    },
}

/// Configuración de conexión SSH
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SSHConfig {
    pub host: String,
    pub port: u16,
    pub username: String,
    pub auth: SSHAuth,
}

/// Información del estado de los servicios en la VPS
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServiceStatus {
    pub backend_installed: bool,
    pub agent_installed: bool,
    pub backend_running: bool,
    pub agent_running: bool,
    pub postgresql_running: bool,
    pub backend_path: Option<String>,
    pub agent_path: Option<String>,
}

/// Configuración del backend leída desde la VPS
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BackendConfig {
    pub api_url: String,
    pub ws_url: String,
    pub environment: String,
    pub port: String,
}

/// Cliente SSH reutilizable
pub struct SSHClient {
    session: Session,
    config: SSHConfig,
}

impl SSHClient {
    /// Crea una nueva conexión SSH
    pub fn connect(config: SSHConfig) -> Result<Self> {
        // Conectar al host remoto
        let tcp = TcpStream::connect(format!("{}:{}", config.host, config.port))
            .context(format!("No se pudo conectar a {}:{}", config.host, config.port))?;

        // Crear sesión SSH
        let mut session = Session::new()?;
        
        session.set_tcp_stream(tcp);
        session.handshake()
            .context("Fallo en el handshake SSH")?;

        // Autenticar según el método especificado
        match &config.auth {
            SSHAuth::Password { password } => {
                session
                    .userauth_password(&config.username, password)
                    .context("Autenticación con contraseña fallida")?;
            }
            SSHAuth::PrivateKey {
                private_key_path,
                passphrase,
            } => {
                session
                    .userauth_pubkey_file(
                        &config.username,
                        None,
                        Path::new(private_key_path),
                        passphrase.as_deref(),
                    )
                    .context("Autenticación con clave privada fallida")?;
            }
            SSHAuth::PrivateKeyData {
                private_key_data,
                passphrase,
            } => {
                // Crear un archivo temporal para la clave privada
                use std::io::Write;
                let temp_dir = std::env::temp_dir();
                let temp_key_path = temp_dir.join(format!("aymc_key_{}.tmp", std::process::id()));
                
                // Escribir la clave en el archivo temporal
                let mut temp_file = std::fs::File::create(&temp_key_path)
                    .context("No se pudo crear archivo temporal para clave privada")?;
                temp_file
                    .write_all(private_key_data.as_bytes())
                    .context("No se pudo escribir clave privada temporal")?;
                
                // Establecer permisos restrictivos (solo lectura para el propietario)
                #[cfg(unix)]
                {
                    use std::os::unix::fs::PermissionsExt;
                    let mut perms = std::fs::metadata(&temp_key_path)?.permissions();
                    perms.set_mode(0o600);
                    std::fs::set_permissions(&temp_key_path, perms)?;
                }
                
                // Autenticar usando el archivo temporal
                let result = session.userauth_pubkey_file(
                    &config.username,
                    None,
                    &temp_key_path,
                    passphrase.as_deref(),
                );
                
                // Eliminar el archivo temporal inmediatamente
                let _ = std::fs::remove_file(&temp_key_path);
                
                result.context("Autenticación con clave privada (memoria) fallida")?;
            }
        }

        // Verificar que la autenticación fue exitosa
        if !session.authenticated() {
            return Err(anyhow!("Autenticación SSH fallida"));
        }

        Ok(Self { session, config })
    }

    /// Ejecuta un comando remoto y retorna el output
    pub fn execute_command(&self, command: &str) -> Result<String> {
        let mut channel = self.session.channel_session()
            .context("No se pudo crear el canal SSH")?;
        
        channel.exec(command)
            .context(format!("No se pudo ejecutar: {}", command))?;

        let mut output = String::new();
        channel.read_to_string(&mut output)
            .context("No se pudo leer el output del comando")?;

        channel.wait_close()
            .context("Error al cerrar el canal")?;

        let exit_status = channel.exit_status()
            .context("No se pudo obtener el código de salida")?;

        if exit_status != 0 {
            return Err(anyhow!(
                "Comando falló con código {}: {}",
                exit_status,
                output
            ));
        }

        Ok(output)
    }

    /// Ejecuta un comando y retorna tanto stdout como stderr
    pub fn execute_command_with_stderr(&self, command: &str) -> Result<(String, String)> {
        let mut channel = self.session.channel_session()
            .context("No se pudo crear el canal SSH")?;
        
        channel.exec(command)
            .context(format!("No se pudo ejecutar: {}", command))?;

        let mut stdout = String::new();
        channel.read_to_string(&mut stdout)
            .context("No se pudo leer stdout")?;

        let mut stderr = String::new();
        channel.stderr().read_to_string(&mut stderr)
            .context("No se pudo leer stderr")?;

        channel.wait_close()
            .context("Error al cerrar el canal")?;

        Ok((stdout, stderr))
    }

    /// Verifica si un archivo o directorio existe en la VPS
    pub fn file_exists(&self, path: &str) -> Result<bool> {
        let command = format!("test -e {} && echo 'exists' || echo 'not_exists'", path);
        let output = self.execute_command(&command)?;
        Ok(output.trim() == "exists")
    }

    /// Lee el contenido de un archivo remoto
    pub fn read_file(&self, path: &str) -> Result<String> {
        let command = format!("cat {}", path);
        self.execute_command(&command)
    }

    /// Verifica si un servicio systemd está corriendo
    pub fn is_service_running(&self, service_name: &str) -> Result<bool> {
        let command = format!("systemctl is-active {}", service_name);
        match self.execute_command(&command) {
            Ok(output) => Ok(output.trim() == "active"),
            Err(_) => Ok(false),
        }
    }

    /// Obtiene el estado de todos los servicios AYMC
    pub fn check_services(&self) -> Result<ServiceStatus> {
        let backend_installed = self.file_exists("/opt/aymc/backend/aymc-backend")?;
        let agent_installed = self.file_exists("/opt/aymc/agent/aymc-agent")?;
        
        let backend_running = self.is_service_running("aymc-backend")?;
        let agent_running = self.is_service_running("aymc-agent")?;
        let postgresql_running = self.is_service_running("postgresql")?;

        let backend_path = if backend_installed {
            Some("/opt/aymc/backend".to_string())
        } else {
            None
        };

        let agent_path = if agent_installed {
            Some("/opt/aymc/agent".to_string())
        } else {
            None
        };

        Ok(ServiceStatus {
            backend_installed,
            agent_installed,
            backend_running,
            agent_running,
            postgresql_running,
            backend_path,
            agent_path,
        })
    }

    /// Lee la configuración del backend desde /etc/aymc/backend.env
    pub fn get_backend_config(&self) -> Result<BackendConfig> {
        let env_content = self.read_file("/etc/aymc/backend.env")
            .context("No se pudo leer /etc/aymc/backend.env")?;

        let mut port = String::from("8080");
        let mut environment = String::from("production");

        // Parsear el archivo .env
        for line in env_content.lines() {
            let line = line.trim();
            if line.starts_with('#') || line.is_empty() {
                continue;
            }

            if let Some((key, value)) = line.split_once('=') {
                let key = key.trim();
                let value = value.trim().trim_matches('"');

                match key {
                    "APP_PORT" => port = value.to_string(),
                    "APP_ENV" => environment = value.to_string(),
                    _ => {}
                }
            }
        }

        // Construir URLs basadas en el host y puerto
        let host = &self.config.host;
        let api_url = format!("http://{}:{}/api/v1", host, port);
        let ws_url = format!("ws://{}:{}/api/v1/ws", host, port);

        Ok(BackendConfig {
            api_url,
            ws_url,
            environment,
            port,
        })
    }

    /// Sube un archivo local a la VPS
    pub fn upload_file(&self, local_path: &str, remote_path: &str) -> Result<()> {
        let local_content = std::fs::read(local_path)
            .context(format!("No se pudo leer el archivo local: {}", local_path))?;

        let mut remote_file = self.session.scp_send(
            Path::new(remote_path),
            0o644,
            local_content.len() as u64,
            None,
        ).context("No se pudo crear el archivo remoto")?;

        remote_file.write_all(&local_content)
            .context("No se pudo escribir en el archivo remoto")?;

        Ok(())
    }

    /// Sube contenido desde string a un archivo remoto
    pub fn upload_content(&self, content: &str, remote_path: &str) -> Result<()> {
        let bytes = content.as_bytes();
        
        let mut remote_file = self.session.scp_send(
            Path::new(remote_path),
            0o644,
            bytes.len() as u64,
            None,
        ).context("No se pudo crear el archivo remoto")?;

        remote_file.write_all(bytes)
            .context("No se pudo escribir en el archivo remoto")?;

        Ok(())
    }

    /// Ejecuta un comando con output en tiempo real (streaming)
    /// Retorna un Vec de líneas según se van generando
    pub fn execute_command_streaming(&self, command: &str) -> Result<Vec<String>> {
        let mut channel = self.session.channel_session()
            .context("No se pudo crear el canal SSH")?;
        
        channel.exec(command)
            .context(format!("No se pudo ejecutar: {}", command))?;

        let mut lines = Vec::new();
        let mut buffer = [0u8; 1024];

        loop {
            match channel.read(&mut buffer) {
                Ok(0) => break, // EOF
                Ok(n) => {
                    let text = String::from_utf8_lossy(&buffer[..n]);
                    for line in text.lines() {
                        lines.push(line.to_string());
                    }
                }
                Err(e) => return Err(anyhow!("Error al leer del canal: {}", e)),
            }
        }

        channel.wait_close()
            .context("Error al cerrar el canal")?;

        Ok(lines)
    }

    /// Obtiene la información del host remoto
    pub fn get_host_info(&self) -> Result<String> {
        let os_info = self.execute_command("cat /etc/os-release")?;
        Ok(os_info)
    }

    /// Verifica si tiene permisos sudo
    pub fn has_sudo_access(&self) -> Result<bool> {
        match self.execute_command("sudo -n true 2>&1") {
            Ok(_) => Ok(true),
            Err(_) => Ok(false),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    // Nota: Estos tests requieren un servidor SSH configurado
    // Se pueden ejecutar con: cargo test --features test-ssh

    #[test]
    #[ignore]
    fn test_ssh_connection_password() {
        let config = SSHConfig {
            host: "localhost".to_string(),
            port: 22,
            username: "testuser".to_string(),
            auth: SSHAuth::Password {
                password: "testpass".to_string(),
            },
        };

        let client = SSHClient::connect(config);
        assert!(client.is_ok());
    }

    #[test]
    #[ignore]
    fn test_execute_command() {
        let config = SSHConfig {
            host: "localhost".to_string(),
            port: 22,
            username: "testuser".to_string(),
            auth: SSHAuth::Password {
                password: "testpass".to_string(),
            },
        };

        let client = SSHClient::connect(config).unwrap();
        let output = client.execute_command("echo 'Hello, SSH!'");
        
        assert!(output.is_ok());
        assert_eq!(output.unwrap().trim(), "Hello, SSH!");
    }
}
