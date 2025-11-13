# AYMC Agent

**Advanced Minecraft Control Agent** - Agente de control remoto para servidores de Minecraft

## 📋 Descripción

El agente AYMC es un componente crítico del sistema AMCP (Advanced Minecraft Control Panel). Se ejecuta en las VPS donde están alojados los servidores de Minecraft y proporciona:

- ✅ Ejecución y gestión de servidores Minecraft (Paper, Purpur, Velocity, etc.)
- ✅ Monitoreo de recursos en tiempo real (CPU, RAM, disco, red)
- ✅ Captura y streaming de logs estructurados
- ✅ Comunicación segura vía gRPC + TLS 1.3
- ✅ Instalación automática de dependencias (Java, screen, etc.)
- ✅ Gestión remota de archivos
- ✅ Sistema de comandos en tiempo real

## 🏗️ Arquitectura

```
┌─────────────────────────────────────────┐
│         AYMC Frontend (Tauri)           │
│              WebSocket                   │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│      Backend Central (Go + gRPC)        │
│        WebSocket + REST API              │
└──────────────┬──────────────────────────┘
               │
               ▼ gRPC (TLS 1.3)
┌─────────────────────────────────────────┐
│          AYMC Agent (Go)                │
│   ┌──────────────────────────────┐     │
│   │  Core Engine                 │     │
│   │  - Executor                  │     │
│   │  - Monitor                   │     │
│   │  - File Manager              │     │
│   └──────────────────────────────┘     │
│                                         │
│   ┌──────────────────────────────┐     │
│   │  gRPC Server                 │     │
│   │  - AgentService              │     │
│   │  - StreamLogs                │     │
│   └──────────────────────────────┘     │
│                                         │
│   ┌──────────────────────────────┐     │
│   │  Security Manager            │     │
│   │  - TLS/Certificates          │     │
│   │  - Token Authentication      │     │
│   └──────────────────────────────┘     │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│    Servidores Minecraft (Procesos)      │
│    - Paper, Purpur, Velocity, etc.      │
└─────────────────────────────────────────┘
```

## 🚀 Instalación

### Linux/Unix (Bash)

```bash
curl -fsSL https://raw.githubusercontent.com/aymc/agent/main/installer/install_agent.sh | sudo bash
```

O manual:

```bash
sudo bash install_agent.sh
```

### Windows (PowerShell como Administrador)

```powershell
iwr -useb https://raw.githubusercontent.com/aymc/agent/main/installer/install_agent.ps1 | iex
```

O manual:

```powershell
.\install_agent.ps1
```

## ⚙️ Configuración

El archivo de configuración se encuentra en:
- **Linux**: `/etc/aymc/agent.json`
- **Windows**: `C:\ProgramData\AYMC\agent.json`

### Ejemplo de configuración:

```json
{
  "agent_id": "agent-unique-id",
  "backend_url": "backend.example.com:50050",
  "port": 50051,
  "log_level": "info",
  "max_servers": 10,
  "java_path": "/usr/bin/java",
  "work_dir": "/var/aymc/servers",
  "enable_metrics": true,
  "metrics_interval": "5s",
  "custom_env": {}
}
```

## 📦 Compilación desde el código fuente

### Requisitos

- Go 1.23+
- Protocol Buffers compiler (protoc)
- Make (opcional)

### Pasos

1. **Clonar el repositorio**:
```bash
git clone https://github.com/aymc/agent.git
cd agent
```

2. **Instalar dependencias**:
```bash
go mod download
```

3. **Generar código protobuf**:
```bash
# Instalar protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generar código
protoc --go_out=. --go-grpc_out=. proto/agent.proto
```

4. **Compilar**:
```bash
go build -o aymc-agent main.go
```

5. **Compilar para múltiples plataformas**:
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o aymc-agent-linux-amd64 main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o aymc-agent-windows-amd64.exe main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o aymc-agent-darwin-amd64 main.go
GOOS=darwin GOARCH=arm64 go build -o aymc-agent-darwin-arm64 main.go
```

## 🔧 Uso

### Como servicio (systemd - Linux)

```bash
# Iniciar
sudo systemctl start aymc-agent

# Detener
sudo systemctl stop aymc-agent

# Estado
sudo systemctl status aymc-agent

# Habilitar inicio automático
sudo systemctl enable aymc-agent

# Ver logs
sudo journalctl -u aymc-agent -f
```

### Como servicio (Windows)

```powershell
# Iniciar
Start-Service AYMCAgent

# Detener
Stop-Service AYMCAgent

# Estado
Get-Service AYMCAgent
```

### Ejecución manual

```bash
# Linux/macOS
./aymc-agent --config=/etc/aymc/agent.json --port=50051

# Windows
aymc-agent.exe --config=C:\ProgramData\AYMC\agent.json --port=50051
```

### Opciones de línea de comandos

```
--config <path>    Ruta al archivo de configuración (default: /etc/aymc/agent.json)
--port <number>    Puerto gRPC (default: 50051)
--cert <path>      Ruta al certificado TLS
--key <path>       Ruta a la clave TLS
--debug            Habilitar modo debug
```

## 🔒 Seguridad

### TLS/Certificados

El agente soporta TLS 1.3 con certificados:

1. **Certificados propios** (producción):
```bash
./aymc-agent --cert=/path/to/cert.pem --key=/path/to/key.pem
```

2. **Certificados autofirmados** (desarrollo):
El agente genera automáticamente certificados autofirmados si no se proporcionan.

### Autenticación

- Token-based authentication para cada request gRPC
- Validación de identidad del cliente
- Rate limiting y protección contra ataques

## 📊 API gRPC

### Servicios disponibles

- `GetAgentInfo` - Información del agente
- `GetSystemMetrics` - Métricas del sistema
- `ListServers` - Listar servidores
- `StartServer` - Iniciar servidor
- `StopServer` - Detener servidor
- `SendCommand` - Enviar comando
- `StreamLogs` - Stream de logs en tiempo real
- `ReadFile/WriteFile` - Gestión de archivos
- `DownloadServer` - Descargar software del servidor
- `CheckDependencies` - Verificar dependencias

Ver [proto/agent.proto](proto/agent.proto) para la definición completa.

## 🧪 Testing

```bash
# Tests unitarios
go test ./...

# Tests con cobertura
go test -cover ./...

# Tests de integración
go test -tags=integration ./tests/...
```

## 📝 Logs

Los logs se escriben en:
- **Linux**: `/var/log/aymc/agent.log`
- **Windows**: `C:\ProgramData\AYMC\logs\agent.log`
- **Stdout**: Cuando se ejecuta manualmente

### Niveles de log

- `debug` - Información detallada para debugging
- `info` - Información general de operación
- `warn` - Advertencias que no impiden la operación
- `error` - Errores que requieren atención

## 🛠️ Troubleshooting

### El agente no inicia

1. Verificar logs: `journalctl -u aymc-agent -n 50`
2. Verificar permisos del directorio de trabajo
3. Verificar que el puerto 50051 no esté en uso

### No se puede conectar al agente

1. Verificar firewall (puerto 50051/TCP debe estar abierto)
2. Verificar certificados TLS
3. Verificar conectividad de red

### Servidor de Minecraft no inicia

1. Verificar que Java esté instalado: `java -version`
2. Verificar permisos del archivo JAR
3. Revisar logs del servidor en el directorio de trabajo

## 🤝 Contribuir

Ver [CONTRIBUTING.md](../CONTRIBUTING.md) para guías de contribución.

## 📄 Licencia

[Pendiente definir]

## 🔗 Enlaces

- [Documentación completa](../docs/)
- [Topología del sistema](../.github/prompts/topología.prompt.md)
- [Plan de tiempos](../.github/prompts/plan_de_tiempos.prompt.md)
- [Issues](https://github.com/aymc/agent/issues)
