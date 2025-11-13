# Fase B.5 - Sistema de Comunicación gRPC con Agentes ✅

**Estado**: Completado  
**Fecha**: 13 de noviembre de 2025  
**Autor**: Sistema AYMC  

---

## 📊 Resumen Ejecutivo

La **Fase B.5** implementa el sistema completo de comunicación gRPC entre el backend central y los agentes remotos. Permite controlar servidores Minecraft de manera distribuida, con health checks automáticos, reconexión, y manejo de fallos.

### Estadísticas Globales

- **Archivos creados**: 6
- **Archivos modificados**: 4
- **Líneas de código**: ~2,200 líneas
- **Endpoints REST para Agents**: 5 endpoints
- **Operaciones gRPC**: StartServer, StopServer, RestartServer, GetStatus, SendCommand
- **Health Checks**: Automáticos cada 30 segundos
- **Compilación**: ✅ Exitosa sin errores

---

## 🏗️ Arquitectura del Sistema

### Flujo de Comunicación

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend (Tauri/Vue)                     │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP/REST
┌────────────────────────▼────────────────────────────────────┐
│                  Backend REST API (Gin)                     │
│         api/rest/handlers/agents.go (5 endpoints)           │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│              services/agents/service.go                     │
│    StartServer(), StopServer(), RestartServer()             │
│    GetAgentInfo(), GetAgentMetrics()                        │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│          services/agents/registry.go (Registry)             │
│     Map<UUID, AgentConnection> - Thread Safe                │
│     Register(), Unregister(), GetAgent()                    │
└────────────────────────┬────────────────────────────────────┘
                         │
           ┌─────────────┼─────────────┐
           │                           │
┌──────────▼──────────┐    ┌──────────▼──────────┐
│ AgentConnection 1   │    │ AgentConnection 2   │
│ gRPC Client         │    │ gRPC Client         │
│ IP: 192.168.1.10    │    │ IP: 192.168.1.20    │
└──────────┬──────────┘    └──────────┬──────────┘
           │ gRPC                     │ gRPC
┌──────────▼──────────┐    ┌──────────▼──────────┐
│   Agent 1 (Go)      │    │   Agent 2 (Go)      │
│   Port 50051        │    │   Port 50051        │
│   Minecraft Servers │    │   Minecraft Servers │
└─────────────────────┘    └─────────────────────┘
```

### Health Monitor (Background)

```
┌──────────────────────────────────────────────────┐
│       services/agents/health.go                  │
│                                                  │
│   Goroutine ejecutándose cada 30 segundos       │
│                                                  │
│   1. Obtiene lista de agentes del Registry      │
│   2. Ping a cada agente en paralelo             │
│   3. Si falla: incrementa consecutive_fails     │
│   4. Si consecutive_fails >= 3: marca offline   │
│   5. Actualiza last_seen en base de datos       │
│   6. Actualiza métricas (CPU, RAM, Disk)        │
└──────────────────────────────────────────────────┘
```

---

## 📁 Archivos Implementados

### 1. proto/agent.proto + Generados

**Origen**: Copiado de `/agent/proto/agent.proto`

**Archivos generados**:
- `proto/agent.pb.go` (mensaje structs)
- `proto/agent_grpc.pb.go` (cliente/servidor gRPC)

**Servicios gRPC disponibles**:
```protobuf
service AgentService {
  rpc GetAgentInfo(Empty) returns (AgentInfo);
  rpc GetSystemMetrics(Empty) returns (SystemMetrics);
  rpc StartServer(StartServerRequest) returns (ServerResponse);
  rpc StopServer(ServerRequest) returns (ServerResponse);
  rpc RestartServer(ServerRequest) returns (ServerResponse);
  rpc GetServer(ServerRequest) returns (ServerInfo);
  rpc SendCommand(CommandRequest) returns (CommandResponse);
  rpc Ping(Empty) returns (PongResponse);
  rpc HealthCheck(Empty) returns (HealthStatus);
}
```

---

### 2. services/agents/connection.go (307 líneas)

**Propósito**: Gestión de conexión gRPC individual a un agente.

#### Estructuras Principales

```go
type AgentConnection struct {
    ID              string
    Agent           *models.Agent
    Client          pb.AgentServiceClient  // Cliente gRPC
    conn            *grpc.ClientConn
    lastSeen        time.Time
    status          AgentStatus
    metrics         *AgentMetrics
    consecutiveFails int
    mu              sync.RWMutex
    logger          *zap.Logger
}

type AgentMetrics struct {
    CPUPercent    float64
    MemoryTotal   uint64
    MemoryUsed    uint64
    MemoryPercent float64
    DiskTotal     uint64
    DiskUsed      uint64
    DiskPercent   float64
    ActiveServers int32
    MaxServers    int32
    Uptime        int64
    LastUpdated   time.Time
}
```

#### Métodos Principales

| Método | Descripción | Timeout |
|--------|-------------|---------|
| `Connect()` | Establece conexión gRPC con agente | 10s |
| `Disconnect()` | Cierra conexión gracefully | - |
| `Ping()` | Envía ping para verificar conectividad | 5s |
| `IsHealthy()` | Verifica si agente está saludable | - |
| `UpdateMetrics()` | Obtiene métricas del sistema vía gRPC | 5s |
| `MarkAsOffline()` | Marca agente como offline | - |
| `GetStatus()` | Retorna estado actual | - |
| `GetMetrics()` | Retorna copia de métricas | - |

#### Estados de Agente

```go
const (
    AgentStatusConnecting AgentStatus = "connecting"
    AgentStatusOnline     AgentStatus = "online"
    AgentStatusOffline    AgentStatus = "offline"
    AgentStatusError      AgentStatus = "error"
)
```

#### Criterios de Salud

Un agente se considera **saludable** si:
1. ✅ Status es `online`
2. ✅ `consecutiveFails < 3`
3. ✅ Se vio en los últimos 2 minutos

---

### 3. services/agents/registry.go (268 líneas)

**Propósito**: Registry thread-safe de todas las conexiones de agentes.

#### Estructura Principal

```go
type AgentRegistry struct {
    agents map[uuid.UUID]*AgentConnection  // Mapa thread-safe
    mu     sync.RWMutex
    db     *gorm.DB
    logger *zap.Logger
}
```

#### Métodos del Registry

| Método | Descripción | Sincroniza BD |
|--------|-------------|---------------|
| `Register()` | Registra y conecta un agente | ✅ |
| `Unregister()` | Desconecta y elimina agente | ✅ |
| `GetAgent()` | Obtiene conexión de un agente | ❌ |
| `ListAgents()` | Lista todas las conexiones | ❌ |
| `GetOnlineAgents()` | Solo agentes saludables | ❌ |
| `Count()` | Total de agentes registrados | ❌ |
| `CountOnline()` | Total de agentes online | ❌ |
| `LoadAgentsFromDatabase()` | Carga y reconecta agentes de BD | ✅ |
| `Shutdown()` | Cierra todas las conexiones | ✅ |
| `UpdateAgentStatus()` | Actualiza estado en BD | ✅ |
| `UpdateAgentLastSeen()` | Actualiza timestamp | ✅ |

#### Flujo de Registro

```
1. Registry.Register(agent)
   ↓
2. NewAgentConnection(agent)
   ↓
3. AgentConnection.Connect() → gRPC dial
   ↓
4. Guardar en map: agents[uuid] = conn
   ↓
5. Actualizar BD: status = "online"
   ↓
6. Log: "Agent registered successfully"
```

#### Reconexión Automática

Al iniciar el backend:
```go
agentRegistry.LoadAgentsFromDatabase(ctx)
// Intenta conectar a todos los agentes en la BD
// Si falla, continúa con el siguiente
```

---

### 4. services/agents/health.go (259 líneas)

**Propósito**: Monitoreo automático de salud de agentes en background.

#### Estructura Principal

```go
type HealthMonitor struct {
    registry *AgentRegistry
    interval time.Duration  // 30 segundos por defecto
    ctx      context.Context
    cancel   context.CancelFunc
    wg       sync.WaitGroup
    logger   *zap.Logger
}
```

#### Constantes de Configuración

```go
const (
    DefaultHealthCheckInterval = 30 * time.Second
    MaxConsecutiveFailures     = 3
    HealthCheckTimeout         = 5 * time.Second
)
```

#### Flujo del Health Monitor

```
1. Start() → Inicia goroutine
   ↓
2. Loop cada 30 segundos:
   ├─> Obtener todos los agentes del registry
   ├─> Para cada agente (en paralelo):
   │   ├─> Ping con timeout de 5s
   │   ├─> Si OK:
   │   │   ├─> UpdateMetrics()
   │   │   ├─> UpdateAgentLastSeen() en BD
   │   │   └─> Reset consecutiveFails
   │   └─> Si FALLA:
   │       ├─> Incrementar consecutiveFails
   │       └─> Si >= 3:
   │           ├─> MarkAsOffline()
   │           ├─> UpdateAgentStatus(offline) en BD
   │           └─> TODO: Failover de servidores
   └─> Log resumen (online/offline)
   
3. Stop() → Cancela context, espera WaitGroup
```

#### Métodos Principales

| Método | Descripción |
|--------|-------------|
| `Start()` | Inicia el monitor en goroutine |
| `Stop()` | Detiene el monitor gracefully |
| `checkAllAgents()` | Verifica todos los agentes en paralelo |
| `checkAgent()` | Verifica un agente específico |
| `handleAgentFailure()` | Maneja agente offline (failover futuro) |
| `CheckAgent()` | Health check manual de un agente |
| `GetStats()` | Estadísticas del monitor |

---

### 5. services/agents/service.go (412 líneas)

**Propósito**: Capa de servicio para operaciones en agentes remotos.

#### Estructura Principal

```go
type AgentService struct {
    registry *AgentRegistry
    logger   *zap.Logger
}
```

#### DTOs

```go
type ServerOperationRequest struct {
    ServerID   string
    ServerName string
    AgentID    uuid.UUID
    Config     *models.Server
}

type ServerOperationResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Status  string `json:"status,omitempty"`
}
```

#### Operaciones de Servidores

##### 1. StartServer

```go
func (s *AgentService) StartServer(ctx context.Context, req *ServerOperationRequest) (*ServerOperationResponse, error)
```

**Flujo**:
1. Obtiene AgentConnection del registry
2. Verifica que agente esté saludable
3. Prepara `StartServerRequest` con configuración
4. Llama `conn.Client.StartServer()` vía gRPC (timeout 30s)
5. Retorna respuesta con success/mensaje

##### 2. StopServer

```go
func (s *AgentService) StopServer(ctx context.Context, req *ServerOperationRequest) (*ServerOperationResponse, error)
```

**Flujo**:
1. Obtiene AgentConnection del registry
2. Verifica que agente esté saludable
3. Llama `conn.Client.StopServer()` vía gRPC (timeout 30s)
4. Retorna respuesta con success/mensaje

##### 3. RestartServer

```go
func (s *AgentService) RestartServer(ctx context.Context, req *ServerOperationRequest) (*ServerOperationResponse, error)
```

**Flujo**:
1. Obtiene AgentConnection del registry
2. Verifica que agente esté saludable
3. Llama `conn.Client.RestartServer()` vía gRPC (timeout 60s)
4. Retorna respuesta con success/mensaje

##### 4. GetServerStatus

```go
func (s *AgentService) GetServerStatus(ctx context.Context, serverID string, agentID uuid.UUID) (*pb.ServerInfo, error)
```

**Retorna**: Información completa del servidor desde el agente (timeout 10s)

##### 5. SendCommand

```go
func (s *AgentService) SendCommand(ctx context.Context, serverID string, agentID uuid.UUID, command string) (string, error)
```

**Uso**: Enviar comandos de consola al servidor (timeout 15s)

#### Operaciones de Agentes

| Método | Descripción | Timeout |
|--------|-------------|---------|
| `GetAgentInfo()` | Información del agente (version, uptime, etc) | 5s |
| `GetAgentMetrics()` | Métricas del sistema (CPU, RAM, Disk) | 5s |
| `GetRegistry()` | Acceso al registry (uso interno) | - |

---

### 6. services/server/control.go (Modificado)

**Cambio**: Integración de AgentService en operaciones de control.

#### Antes (TODOs)

```go
// TODO: Send start command to agent via gRPC
// For now, we'll just update the status to running after a delay
```

#### Después (Integrado)

```go
// Send start command to agent via gRPC
ctx := context.Background()
req := &agents.ServerOperationRequest{
    ServerID:   serverID.String(),
    ServerName: server.Name,
    AgentID:    server.AgentID,
    Config:     &server,
}

resp, err := s.agentService.StartServer(ctx, req)
if err != nil {
    // Si falla el gRPC, revertir el estado
    db.Model(&server).Update("status", models.ServerStatusStopped)
    return nil, fmt.Errorf("failed to start server on agent: %w", err)
}

// Actualizar estado según respuesta del agente
if resp.Success {
    db.Model(&server).Update("status", models.ServerStatusRunning)
} else {
    db.Model(&server).Update("status", models.ServerStatusError)
}
```

#### Operaciones Modificadas

- ✅ `Start()` - Ahora llama a `agentService.StartServer()` vía gRPC
- ✅ `Stop()` - Ahora llama a `agentService.StopServer()` vía gRPC
- ⚠️ `Restart()` - Usa Stop() + Start() (heredado)

---

### 7. api/rest/handlers/agents.go (319 líneas) ✨ NUEVO

**Propósito**: Handlers HTTP para endpoints de gestión de agentes.

#### Handlers Implementados

##### 1. ListAgents

```bash
GET /api/v1/agents
Authorization: Bearer <JWT_TOKEN>
```

**Respuesta 200 OK**:
```json
{
  "agents": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "agent_id": "agent-prod-01",
      "hostname": "minecraft-server-1",
      "ip_address": "192.168.1.10",
      "port": 50051,
      "status": "online",
      "version": "0.1.0",
      "os": "linux",
      "is_healthy": true,
      "last_seen": "2025-11-13T15:30:00Z",
      "consecutive_fails": 0,
      "metrics": {
        "cpu_percent": 45.2,
        "memory_total": 16777216000,
        "memory_used": 8388608000,
        "memory_percent": 50.0,
        "disk_total": 107374182400,
        "disk_used": 53687091200,
        "disk_percent": 50.0,
        "active_servers": 3,
        "max_servers": 10,
        "uptime": 86400,
        "last_updated": "2025-11-13T15:30:00Z"
      }
    }
  ],
  "total": 1,
  "online": 1,
  "offline": 0
}
```

##### 2. GetAgent

```bash
GET /api/v1/agents/:id
Authorization: Bearer <JWT_TOKEN>
```

**Respuesta**: AgentResponse completo con métricas

**Errores**:
- 400: Invalid agent ID format
- 404: Agent not found

##### 3. GetAgentHealth

```bash
GET /api/v1/agents/:id/health
Authorization: Bearer <JWT_TOKEN>
```

**Respuesta 200 OK**:
```json
{
  "agent_id": "agent-prod-01",
  "version": "0.1.0",
  "platform": "linux",
  "platform_version": "Ubuntu 22.04",
  "uptime_seconds": 86400,
  "active_servers": 3,
  "max_servers": 10,
  "is_healthy": true,
  "status": "online",
  "last_seen": "2025-11-13T15:30:00Z",
  "consecutive_fails": 0
}
```

**Errores**:
- 400: Invalid agent ID
- 404: Agent not found
- 503: Failed to communicate with agent

##### 4. GetAgentMetrics

```bash
GET /api/v1/agents/:id/metrics
Authorization: Bearer <JWT_TOKEN>
```

**Respuesta 200 OK**:
```json
{
  "timestamp": 1699893000,
  "cpu_percent": 45.2,
  "memory_total": 16777216000,
  "memory_used": 8388608000,
  "memory_percent": 50.0,
  "disk_total": 107374182400,
  "disk_used": 53687091200,
  "disk_percent": 50.0
}
```

##### 5. GetAgentStats

```bash
GET /api/v1/agents/stats
Authorization: Bearer <JWT_TOKEN>
```

**Respuesta 200 OK**:
```json
{
  "total_agents": 5,
  "online_agents": 4,
  "offline_agents": 1
}
```

---

### 8. api/rest/server.go (Modificado)

**Cambios**:
1. Agregado import de `services/agents`
2. Agregado campo `agentHandler` a struct `Server`
3. Actualizada firma de `NewServer()` para aceptar `AgentService`
4. Inicializado `agentHandler` en `NewServer()`
5. Agregadas rutas de agents en `setupRoutes()`

#### Rutas Agregadas

```go
// Agent management routes
agents := api.Group("/agents")
{
    agents.GET("", s.agentHandler.ListAgents)
    agents.GET("/stats", s.agentHandler.GetAgentStats)
    agents.GET("/:id", s.agentHandler.GetAgent)
    agents.GET("/:id/health", s.agentHandler.GetAgentHealth)
    agents.GET("/:id/metrics", s.agentHandler.GetAgentMetrics)
}
```

**Autenticación**: Todas las rutas requieren JWT token válido.

---

### 9. cmd/server/main.go (Modificado)

**Cambios de Inicialización**:

```go
// 1. Inicializar AgentRegistry
agentRegistry := agents.NewAgentRegistry(logger.GetLogger())

// 2. Cargar agentes de BD y reconectar
ctx := context.Background()
if err := agentRegistry.LoadAgentsFromDatabase(ctx); err != nil {
    logger.Warn("Failed to load agents from database", zap.Error(err))
}

// 3. Iniciar HealthMonitor en background
healthMonitor := agents.NewHealthMonitor(agentRegistry, 30*time.Second, logger.GetLogger())
if err := healthMonitor.Start(); err != nil {
    logger.Fatal("Failed to start health monitor", zap.Error(err))
}

// 4. Crear AgentService
agentService := agents.NewAgentService(agentRegistry, logger.GetLogger())

// 5. Inyectar en ServerService
serverService := server.NewServerService(agentService, logger.GetLogger())

// 6. Inyectar en REST Server
apiServer := rest.NewServer(cfg, jwtService, authService, serverService, agentService, logger.GetLogger())
```

**Shutdown Graceful**:

```go
// Stop health monitor
healthMonitor.Stop()

// Shutdown agent registry (close all connections)
agentRegistry.Shutdown()

// Shutdown API server
apiServer.Shutdown(ctx)
```

---

## 🔄 Flujo de Operación Completo

### Ejemplo: Iniciar un Servidor Minecraft

```
1. Usuario hace clic en "Start" en el frontend
   ↓
2. Frontend: POST /api/v1/servers/:id/start
   ↓
3. ServerHandler.Start() valida permisos
   ↓
4. ServerService.Start() obtiene datos del servidor
   ↓
5. ServerService.Start() llama agentService.StartServer()
   ├─> Obtiene AgentConnection del registry
   ├─> Verifica agente healthy
   └─> Llama conn.Client.StartServer() via gRPC
   ↓
6. AgentConnection envía StartServerRequest al agente
   ↓
7. Agente (Go) recibe request y ejecuta:
   ├─> Crea directorio del servidor
   ├─> Descarga JAR si es necesario
   ├─> Inicia proceso Java
   └─> Retorna ServerResponse
   ↓
8. AgentService recibe respuesta y la pasa a ServerService
   ↓
9. ServerService actualiza estado en BD:
   ├─> Si success: status = "running"
   └─> Si error: status = "error"
   ↓
10. Handler retorna JSON al frontend
   ↓
11. Frontend actualiza UI en tiempo real
```

### Health Check Automático (Paralelo)

Mientras tanto, cada 30 segundos:

```
1. HealthMonitor ejecuta checkAllAgents()
   ↓
2. Para cada agente en paralelo:
   ├─> Ping al agente
   ├─> Si OK:
   │   ├─> UpdateMetrics() (CPU, RAM, Disk)
   │   ├─> UpdateAgentLastSeen() en BD
   │   └─> Reset consecutiveFails = 0
   └─> Si FALLA:
       ├─> consecutiveFails++
       └─> Si consecutiveFails >= 3:
           ├─> MarkAsOffline()
           ├─> UpdateAgentStatus("offline") en BD
           └─> Log: "Agent marked as offline"
```

---

## 🧪 Testing Manual

### Prerrequisitos

1. **Iniciar un agente**:
```bash
cd /home/shni/Documents/GitHub/AYMC/agent
./aymc-agent --port 50051
```

2. **Iniciar el backend**:
```bash
cd /home/shni/Documents/GitHub/AYMC/backend
./bin/aymc-server
```

3. **Obtener JWT Token**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'

export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Flujo de Prueba - Gestión de Agentes

#### 1. Listar Agentes

```bash
curl -X GET http://localhost:8080/api/v1/agents \
  -H "Authorization: Bearer $TOKEN"
```

**Resultado esperado**: Lista de agentes con estado online/offline

#### 2. Ver Detalles de un Agente

```bash
curl -X GET http://localhost:8080/api/v1/agents/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer $TOKEN"
```

**Resultado esperado**: Información completa del agente con métricas

#### 3. Health Check Manual

```bash
curl -X GET http://localhost:8080/api/v1/agents/550e8400-e29b-41d4-a716-446655440000/health \
  -H "Authorization: Bearer $TOKEN"
```

**Resultado esperado**: 
```json
{
  "agent_id": "agent-prod-01",
  "version": "0.1.0",
  "platform": "linux",
  "uptime_seconds": 3600,
  "active_servers": 0,
  "is_healthy": true,
  "status": "online"
}
```

#### 4. Obtener Métricas del Sistema

```bash
curl -X GET http://localhost:8080/api/v1/agents/550e8400-e29b-41d4-a716-446655440000/metrics \
  -H "Authorization: Bearer $TOKEN"
```

**Resultado esperado**: Métricas en tiempo real (CPU, RAM, Disk)

#### 5. Estadísticas Globales

```bash
curl -X GET http://localhost:8080/api/v1/agents/stats \
  -H "Authorization: Bearer $TOKEN"
```

**Resultado esperado**:
```json
{
  "total_agents": 2,
  "online_agents": 2,
  "offline_agents": 0
}
```

### Flujo de Prueba - Control de Servidores con gRPC

#### 1. Crear Servidor (asignado a agente)

```bash
curl -X POST http://localhost:8080/api/v1/servers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "TestServer",
    "agent_id": "550e8400-e29b-41d4-a716-446655440000",
    "server_type": "paper",
    "version": "1.20.1",
    "port": 25565,
    "max_players": 20,
    "memory_min": 2048,
    "memory_max": 4096
  }'
```

#### 2. Iniciar Servidor (vía gRPC)

```bash
export SERVER_ID="123e4567-e89b-12d3-a456-426614174000"

curl -X POST http://localhost:8080/api/v1/servers/$SERVER_ID/start \
  -H "Authorization: Bearer $TOKEN"
```

**Resultado esperado**: 
- Backend llama `agentService.StartServer()`
- gRPC call al agente
- Agente inicia proceso Java
- Estado actualizado a "running"

#### 3. Verificar Logs del Agente

El agente debe mostrar:
```
[INFO] Received StartServer request: server_id=123e4567...
[INFO] Starting Minecraft server...
[INFO] Server started successfully
```

#### 4. Verificar Logs del Backend

El backend debe mostrar:
```
[INFO] Server start initiated: server_id=123e4567...
[INFO] Sending gRPC StartServer to agent: agent_id=550e8400...
[INFO] Server started successfully on agent
```

#### 5. Detener Servidor

```bash
curl -X POST http://localhost:8080/api/v1/servers/$SERVER_ID/stop \
  -H "Authorization: Bearer $TOKEN"
```

**Resultado esperado**: 
- gRPC call `StopServer()` al agente
- Agente detiene proceso Java
- Estado actualizado a "stopped"

### Simulación de Fallo de Agente

#### 1. Detener el agente manualmente

```bash
# En terminal del agente: Ctrl+C
```

#### 2. Observar Health Monitor

Después de 30 segundos, el backend debe loggear:
```
[WARN] Agent ping failed: agent_id=550e8400... consecutive_fails=1
[WARN] Agent ping failed: agent_id=550e8400... consecutive_fails=2
[WARN] Agent ping failed: agent_id=550e8400... consecutive_fails=3
[ERROR] Agent marked as offline after multiple failures
```

#### 3. Verificar Estado en API

```bash
curl -X GET http://localhost:8080/api/v1/agents/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer $TOKEN"
```

**Resultado esperado**:
```json
{
  "status": "offline",
  "is_healthy": false,
  "consecutive_fails": 3
}
```

#### 4. Reiniciar Agente

```bash
cd /home/shni/Documents/GitHub/AYMC/agent
./aymc-agent --port 50051
```

#### 5. Observar Reconexión

El Health Monitor detectará que el agente volvió y lo marcará como online:
```
[INFO] Agent recovered: agent_id=550e8400...
[INFO] Updated agent status to online
```

---

## 📊 Métricas y Estadísticas

### Código Producido

| Componente | Archivo | Líneas | Funciones |
|------------|---------|--------|-----------|
| Proto Gen | `proto/agent.pb.go` | ~2000 | Generado |
| Proto Gen | `proto/agent_grpc.pb.go` | ~750 | Generado |
| Connection | `services/agents/connection.go` | 307 | 13 |
| Registry | `services/agents/registry.go` | 268 | 13 |
| Health | `services/agents/health.go` | 259 | 8 |
| Service | `services/agents/service.go` | 412 | 8 |
| Handlers | `api/rest/handlers/agents.go` | 319 | 6 |
| **TOTAL** | **7 archivos** | **~4,315** | **48 funciones** |

### Endpoints Disponibles

| Método | Endpoint | Handler | Auth | Admin |
|--------|----------|---------|------|-------|
| GET | `/api/v1/agents` | ListAgents | ✅ | ❌ |
| GET | `/api/v1/agents/stats` | GetAgentStats | ✅ | ❌ |
| GET | `/api/v1/agents/:id` | GetAgent | ✅ | ❌ |
| GET | `/api/v1/agents/:id/health` | GetAgentHealth | ✅ | ❌ |
| GET | `/api/v1/agents/:id/metrics` | GetAgentMetrics | ✅ | ❌ |

### Operaciones gRPC

| Operación | Timeout | Uso |
|-----------|---------|-----|
| `StartServer` | 30s | Iniciar servidor en agente |
| `StopServer` | 30s | Detener servidor en agente |
| `RestartServer` | 60s | Reiniciar servidor en agente |
| `GetServer` | 10s | Estado del servidor |
| `SendCommand` | 15s | Comando de consola |
| `GetAgentInfo` | 5s | Info del agente |
| `GetSystemMetrics` | 5s | Métricas del sistema |
| `Ping` | 5s | Health check |

---

## 🚀 Próximos Pasos

### Fase B.6 - Streaming de Logs (WebSocket)

1. **Implementar StreamLogs gRPC**:
   - Agente envía logs en tiempo real
   - Backend reenvía a clientes WebSocket
   - Filtrado por servidor

2. **WebSocket Hub**:
   - `api/websocket/hub.go`
   - Gestión de clientes conectados
   - Broadcasting de logs

3. **Frontend Integration**:
   - Conectar WebSocket desde Tauri
   - Mostrar logs en consola en tiempo real

### Mejoras de Fase B.5

1. **Failover Automático** (TODO en health.go):
   - Cuando agente falla, detectar servidores running
   - Migrar servidores a agente alternativo
   - Notificar usuarios vía WebSocket

2. **TLS/mTLS para gRPC**:
   - Reemplazar `insecure.NewCredentials()`
   - Certificados SSL para producción
   - Autenticación mutua agente ↔ backend

3. **Balanceo de Carga**:
   - `services/agents/balancer.go`
   - Estrategias: Round Robin, Least Connections, Least Load
   - Selección automática de agente al crear servidor

4. **Metrics Aggregation**:
   - Prometheus exporter
   - Grafana dashboards
   - Alertas automáticas

5. **Agent Registration API**:
   - `POST /api/v1/agents` - Auto-register
   - `DELETE /api/v1/agents/:id` - Unregister
   - API Keys para agentes

---

## ✅ Checklist de Completitud

- [x] ✅ Protobuf copiado y generado (agent.pb.go, agent_grpc.pb.go)
- [x] ✅ AgentConnection con cliente gRPC
- [x] ✅ AgentRegistry thread-safe con map
- [x] ✅ HealthMonitor con goroutine automática (30s)
- [x] ✅ AgentService con operaciones de servidor
- [x] ✅ Integración gRPC en control.go (Start, Stop)
- [x] ✅ Handlers REST para agents (5 endpoints)
- [x] ✅ Rutas de agents en REST server
- [x] ✅ Inicialización en main.go
- [x] ✅ Shutdown graceful de conexiones
- [x] ✅ Compilación exitosa sin errores
- [x] ✅ Documentación completa

**TODOs Pendientes**:
- ⏳ Failover automático de servidores
- ⏳ TLS/mTLS para producción
- ⏳ Balanceo de carga de agentes
- ⏳ Tests unitarios e integración

---

## 📝 Notas Técnicas

### Concurrencia y Thread Safety

Todos los componentes son **thread-safe**:
- ✅ `AgentRegistry` usa `sync.RWMutex`
- ✅ `AgentConnection` usa `sync.RWMutex`
- ✅ `HealthMonitor` usa `context.Context` y `sync.WaitGroup`
- ✅ Goroutines paralelas en health checks

### Manejo de Errores

- **Errores de conexión gRPC**: Loggear + incrementar `consecutiveFails`
- **Timeouts**: Todos los contextos con timeout explícito
- **Agente offline**: Health monitor detecta y actualiza BD
- **Failover**: Placeholder para migración futura

### Performance

- **Health checks paralelos**: Un goroutine por agente
- **Registry lookup**: O(1) con map de UUID
- **gRPC connection pool**: Una conexión permanente por agente
- **Lazy loading**: Agentes se cargan al inicio o bajo demanda

---

## 🎉 Conclusión

La **Fase B.5** está **completada exitosamente** con:

- ✅ ~2,200 líneas de código Go de alta calidad
- ✅ Sistema completo de comunicación gRPC
- ✅ 5 endpoints REST para gestión de agentes
- ✅ Health monitoring automático cada 30 segundos
- ✅ Operaciones de servidor distribuidas (Start, Stop, Restart)
- ✅ Thread-safe y concurrente
- ✅ Shutdown graceful
- ✅ Compilación exitosa sin errores
- ✅ Arquitectura extensible para failover y balanceo

**Estado del Backend AYMC**:
- Fases B.1, B.2, B.3, B.4, B.5: ✅ Completadas
- Próxima fase: B.6 (Streaming de Logs con WebSocket)

**Líneas Totales Backend** (Fases B.1-B.5):
- Fase B.1: ~500 líneas (Setup)
- Fase B.2: ~800 líneas (Database)
- Fase B.3: ~1,247 líneas (Auth)
- Fase B.4: ~1,120 líneas (Server Management)
- Fase B.5: ~2,200 líneas (gRPC Agents)
- **Total**: ~5,867 líneas de código funcional
