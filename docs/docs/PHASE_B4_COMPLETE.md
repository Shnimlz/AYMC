# Fase B.4 - Sistema de Gestión de Servidores Minecraft ✅

**Estado**: Completado  
**Fecha**: 2024  
**Autor**: Sistema AYMC  

---

## 📊 Resumen Ejecutivo

La **Fase B.4** implementa el sistema completo de gestión de servidores Minecraft con operaciones CRUD y control del ciclo de vida (start, stop, restart). El sistema está integrado con autenticación JWT, control de permisos basado en roles (RBAC) y validación de agentes.

### Estadísticas Globales

- **Archivos creados**: 3
- **Archivos modificados**: 2
- **Líneas de código**: ~1,120 líneas
- **Endpoints REST**: 9 endpoints
- **Operaciones CRUD**: 5 (Create, Read, List, Update, Delete)
- **Operaciones de Control**: 4 (Start, Stop, Restart, GetStatus)
- **DTOs**: 5 estructuras de transferencia de datos
- **Validaciones**: Permisos, estado, agente online, unicidad de nombres

---

## 🏗️ Arquitectura del Sistema

### Capas Implementadas

```
┌─────────────────────────────────────────────────────┐
│              REST API Endpoints                     │
│         GET/POST/PUT/DELETE /api/v1/servers         │
└──────────────────┬──────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────┐
│            HTTP Handlers Layer                      │
│     api/rest/handlers/server.go (495 lines)         │
│  Create, Get, List, Update, Delete, Start, Stop,    │
│            Restart, GetStatus                       │
└──────────────────┬──────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────┐
│          Business Logic Layer                       │
│    services/server/service.go (418 lines)           │
│    services/server/control.go (207 lines)           │
│  Permission checks, validations, state management   │
└──────────────────┬──────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────┘
│            Database Layer (GORM)                    │
│   models.Server, models.Agent, models.User          │
│   Relationships: BelongsTo Agent/User               │
└─────────────────────────────────────────────────────┘
```

---

## 📁 Archivos Implementados

### 1. services/server/service.go (418 líneas)

**Propósito**: Lógica de negocio para operaciones CRUD de servidores.

#### Estructuras de Datos

```go
// ServerService - Servicio principal
type ServerService struct {
    logger *zap.Logger
}

// CreateServerRequest - DTO para crear servidor
type CreateServerRequest struct {
    Name       string `json:"name" validate:"required,min=3,max=50"`
    AgentID    string `json:"agent_id" validate:"required,uuid"`
    ServerType string `json:"server_type" validate:"required,oneof=vanilla paper spigot fabric forge"`
    Version    string `json:"version" validate:"required"`
    Port       int    `json:"port" validate:"required,min=1024,max=65535"`
    MaxPlayers int    `json:"max_players" validate:"required,min=1,max=1000"`
    MemoryMin  int    `json:"memory_min" validate:"required,min=512"`
    MemoryMax  int    `json:"memory_max" validate:"required,min=512"`
}

// UpdateServerRequest - DTO para actualizar servidor (campos opcionales)
type UpdateServerRequest struct {
    Name       *string `json:"name,omitempty" validate:"omitempty,min=3,max=50"`
    ServerType *string `json:"server_type,omitempty" validate:"omitempty,oneof=vanilla paper spigot fabric forge"`
    Version    *string `json:"version,omitempty"`
    Port       *int    `json:"port,omitempty" validate:"omitempty,min=1024,max=65535"`
    MaxPlayers *int    `json:"max_players,omitempty" validate:"omitempty,min=1,max=1000"`
    MemoryMin  *int    `json:"memory_min,omitempty" validate:"omitempty,min=512"`
    MemoryMax  *int    `json:"memory_max,omitempty" validate:"omitempty,min=512"`
}

// ServerResponse - Respuesta con información completa
type ServerResponse struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    AgentID     string    `json:"agent_id"`
    UserID      string    `json:"user_id"`
    ServerType  string    `json:"server_type"`
    Version     string    `json:"version"`
    Port        int       `json:"port"`
    Status      string    `json:"status"`
    MaxPlayers  int       `json:"max_players"`
    MemoryMin   int       `json:"memory_min"`
    MemoryMax   int       `json:"memory_max"`
    LastStarted *time.Time `json:"last_started,omitempty"`
    LastStopped *time.Time `json:"last_stopped,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    Agent       AgentInfo `json:"agent"`
    User        UserInfo  `json:"user"`
}

// ServerListResponse - Respuesta paginada
type ServerListResponse struct {
    Servers []ServerResponse `json:"servers"`
    Total   int64            `json:"total"`
    Page    int              `json:"page"`
    PerPage int              `json:"per_page"`
}
```

#### Métodos Principales

| Método | Descripción | Validaciones |
|--------|-------------|--------------|
| `Create()` | Crea un nuevo servidor | Verifica agente online, nombre único por usuario, memoria mínima |
| `GetByID()` | Obtiene servidor por ID | Permisos: usuario dueño o admin |
| `List()` | Lista servidores paginados | Filtra por userID si no es admin |
| `Update()` | Actualiza parcialmente servidor | No puede actualizar si está en ejecución |
| `Delete()` | Elimina servidor | No puede eliminar si está en ejecución |

#### Lógica de Permisos

```go
// Usuario normal: solo ve sus servidores
if !isAdmin {
    query = query.Where("user_id = ?", userID)
}

// Admin: ve todos los servidores
// No se agrega filtro adicional
```

---

### 2. services/server/control.go (207 líneas)

**Propósito**: Control del ciclo de vida de servidores (start, stop, restart).

#### Métodos de Control

```go
// Start - Inicia un servidor
func (s *ServerService) Start(serverID, userID string, isAdmin bool) error
    // 1. Verifica permisos
    // 2. Valida que el servidor puede iniciarse (CanStart)
    // 3. Verifica que el agente está online
    // 4. Actualiza status a "starting"
    // 5. Registra timestamp last_started
    // TODO: Comunicación gRPC con agente

// Stop - Detiene un servidor
func (s *ServerService) Stop(serverID, userID string, isAdmin bool) error
    // 1. Verifica permisos
    // 2. Valida que el servidor puede detenerse (CanStop)
    // 3. Actualiza status a "stopping"
    // 4. Registra timestamp last_stopped
    // TODO: Comunicación gRPC con agente

// Restart - Reinicia un servidor
func (s *ServerService) Restart(serverID, userID string, isAdmin bool) error
    // 1. Llama a Stop() si está running
    // 2. Llama a Start()
    // Maneja caso de servidor ya detenido

// GetStatus - Obtiene estado actual
func (s *ServerService) GetStatus(serverID, userID string, isAdmin bool) (*ServerStatusResponse, error)
    // Retorna: server_id, status, is_running, last_started, last_stopped, agent_online
```

#### Estados de Servidor

| Estado | Descripción | Puede Start | Puede Stop |
|--------|-------------|-------------|------------|
| `stopped` | Detenido | ✅ | ❌ |
| `starting` | Iniciando | ❌ | ❌ |
| `running` | En ejecución | ❌ | ✅ |
| `stopping` | Deteniéndose | ❌ | ❌ |
| `error` | Error | ✅ | ❌ |

---

### 3. api/rest/handlers/server.go (495 líneas)

**Propósito**: Handlers HTTP para endpoints de servidores.

#### CRUD Handlers

##### 1. Create (POST /api/v1/servers)

```bash
curl -X POST http://localhost:8080/api/v1/servers \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Mi Servidor de Survival",
    "agent_id": "550e8400-e29b-41d4-a716-446655440000",
    "server_type": "paper",
    "version": "1.20.1",
    "port": 25565,
    "max_players": 20,
    "memory_min": 2048,
    "memory_max": 4096
  }'
```

**Respuesta 201 Created**:
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Mi Servidor de Survival",
  "agent_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "789e4567-e89b-12d3-a456-426614174000",
  "server_type": "paper",
  "version": "1.20.1",
  "port": 25565,
  "status": "stopped",
  "max_players": 20,
  "memory_min": 2048,
  "memory_max": 4096,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "agent": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Agente Principal",
    "status": "online"
  },
  "user": {
    "id": "789e4567-e89b-12d3-a456-426614174000",
    "username": "admin",
    "email": "admin@aymc.local"
  }
}
```

##### 2. List (GET /api/v1/servers?page=1&per_page=20)

```bash
curl -X GET "http://localhost:8080/api/v1/servers?page=1&per_page=20" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

**Respuesta 200 OK**:
```json
{
  "servers": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Mi Servidor de Survival",
      "status": "running",
      "server_type": "paper",
      "version": "1.20.1",
      "port": 25565,
      "max_players": 20,
      "agent": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "Agente Principal",
        "status": "online"
      }
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 20
}
```

##### 3. Get (GET /api/v1/servers/:id)

```bash
curl -X GET http://localhost:8080/api/v1/servers/123e4567-e89b-12d3-a456-426614174000 \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

**Respuesta**: ServerResponse completo (igual que Create)

##### 4. Update (PUT /api/v1/servers/:id)

```bash
curl -X PUT http://localhost:8080/api/v1/servers/123e4567-e89b-12d3-a456-426614174000 \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Servidor Actualizado",
    "max_players": 30
  }'
```

**Nota**: Solo campos enviados se actualizan (partial update).

##### 5. Delete (DELETE /api/v1/servers/:id)

```bash
curl -X DELETE http://localhost:8080/api/v1/servers/123e4567-e89b-12d3-a456-426614174000 \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

**Respuesta 200 OK**:
```json
{
  "message": "Server deleted successfully"
}
```

#### Control Handlers

##### 6. Start (POST /api/v1/servers/:id/start)

```bash
curl -X POST http://localhost:8080/api/v1/servers/123e4567-e89b-12d3-a456-426614174000/start \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

**Respuesta 200 OK**:
```json
{
  "message": "Server start initiated"
}
```

##### 7. Stop (POST /api/v1/servers/:id/stop)

```bash
curl -X POST http://localhost:8080/api/v1/servers/123e4567-e89b-12d3-a456-426614174000/stop \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

##### 8. Restart (POST /api/v1/servers/:id/restart)

```bash
curl -X POST http://localhost:8080/api/v1/servers/123e4567-e89b-12d3-a456-426614174000/restart \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

##### 9. GetStatus (GET /api/v1/servers/:id/status)

```bash
curl -X GET http://localhost:8080/api/v1/servers/123e4567-e89b-12d3-a456-426614174000/status \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

**Respuesta 200 OK**:
```json
{
  "server_id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "running",
  "is_running": true,
  "last_started": "2024-01-15T11:00:00Z",
  "last_stopped": null,
  "agent_online": true
}
```

---

## 🔒 Sistema de Permisos

### Roles Implementados

| Rol | Permisos |
|-----|----------|
| **User** | - Ver solo sus propios servidores<br>- Crear servidores bajo su cuenta<br>- Modificar/eliminar solo sus servidores<br>- Controlar (start/stop/restart) solo sus servidores |
| **Admin** | - Ver todos los servidores del sistema<br>- Modificar cualquier servidor<br>- Eliminar cualquier servidor<br>- Controlar cualquier servidor |

### Flujo de Autenticación

```
1. Cliente envía JWT en header: Authorization: Bearer <token>
2. AuthMiddleware valida token
3. AuthMiddleware inyecta userID e isAdmin en contexto
4. Handler extrae userID e isAdmin
5. Service valida permisos antes de operación
```

---

## ⚙️ Validaciones Implementadas

### A Nivel de DTO (Validator)

```go
Name:       required, min=3, max=50
AgentID:    required, uuid
ServerType: required, oneof=vanilla|paper|spigot|fabric|forge
Version:    required
Port:       required, min=1024, max=65535
MaxPlayers: required, min=1, max=1000
MemoryMin:  required, min=512 (MB)
MemoryMax:  required, min=512 (MB)
```

### A Nivel de Business Logic

- ✅ **Agente existe y está online** antes de crear servidor
- ✅ **Nombre único por usuario** (no puede haber duplicados)
- ✅ **Servidor no en ejecución** antes de actualizar/eliminar
- ✅ **Estado válido** antes de start/stop (CanStart, CanStop)
- ✅ **Permisos de usuario** (ownership o admin)

---

## 🧪 Testing Manual

### Prerrequisitos

1. **Iniciar el servidor**:
```bash
cd /home/shni/Documents/GitHub/AYMC/backend
./bin/aymc-server
```

2. **Obtener JWT Token** (usando usuario seeded):
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'

# Guardar el token retornado
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Flujo de Prueba Completo

```bash
# 1. Listar servidores (debe estar vacío inicialmente)
curl -X GET "http://localhost:8080/api/v1/servers" \
  -H "Authorization: Bearer $TOKEN"

# 2. Crear servidor (necesita agent_id de seeds)
curl -X POST http://localhost:8080/api/v1/servers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Server",
    "agent_id": "550e8400-e29b-41d4-a716-446655440000",
    "server_type": "paper",
    "version": "1.20.1",
    "port": 25565,
    "max_players": 20,
    "memory_min": 2048,
    "memory_max": 4096
  }'

# Guardar el server_id retornado
export SERVER_ID="123e4567-e89b-12d3-a456-426614174000"

# 3. Obtener detalles del servidor
curl -X GET "http://localhost:8080/api/v1/servers/$SERVER_ID" \
  -H "Authorization: Bearer $TOKEN"

# 4. Iniciar servidor
curl -X POST "http://localhost:8080/api/v1/servers/$SERVER_ID/start" \
  -H "Authorization: Bearer $TOKEN"

# 5. Verificar estado
curl -X GET "http://localhost:8080/api/v1/servers/$SERVER_ID/status" \
  -H "Authorization: Bearer $TOKEN"

# 6. Actualizar configuración
curl -X PUT "http://localhost:8080/api/v1/servers/$SERVER_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "max_players": 30
  }'

# 7. Detener servidor
curl -X POST "http://localhost:8080/api/v1/servers/$SERVER_ID/stop" \
  -H "Authorization: Bearer $TOKEN"

# 8. Eliminar servidor
curl -X DELETE "http://localhost:8080/api/v1/servers/$SERVER_ID" \
  -H "Authorization: Bearer $TOKEN"
```

---

## 📊 Métricas y Estadísticas

### Código Producido

| Componente | Archivo | Líneas | Funciones/Métodos |
|------------|---------|--------|-------------------|
| Service | `services/server/service.go` | 418 | 5 (CRUD) + 4 DTOs |
| Control | `services/server/control.go` | 207 | 4 (Start, Stop, Restart, GetStatus) |
| Handlers | `api/rest/handlers/server.go` | 495 | 9 handlers |
| **TOTAL** | **3 archivos** | **1,120** | **18 funciones** |

### Endpoints Disponibles

| Método | Endpoint | Handler | Auth | Permisos |
|--------|----------|---------|------|----------|
| GET | `/api/v1/servers` | List | ✅ | User/Admin |
| POST | `/api/v1/servers` | Create | ✅ | User/Admin |
| GET | `/api/v1/servers/:id` | Get | ✅ | Owner/Admin |
| PUT | `/api/v1/servers/:id` | Update | ✅ | Owner/Admin |
| DELETE | `/api/v1/servers/:id` | Delete | ✅ | Owner/Admin |
| POST | `/api/v1/servers/:id/start` | Start | ✅ | Owner/Admin |
| POST | `/api/v1/servers/:id/stop` | Stop | ✅ | Owner/Admin |
| POST | `/api/v1/servers/:id/restart` | Restart | ✅ | Owner/Admin |
| GET | `/api/v1/servers/:id/status` | GetStatus | ✅ | Owner/Admin |

---

## 🚀 Próximos Pasos

### Fase B.5 - Sistema de Agentes (Comunicación gRPC)

1. **Definir protobuf** para comunicación Backend ↔ Agent:
   - `StartServer(server_id)`
   - `StopServer(server_id)`
   - `GetServerStatus(server_id)`
   - `StreamLogs(server_id)`

2. **Implementar Agent Service**:
   - `services/agent/service.go`
   - Métodos para comunicar con agentes vía gRPC
   - Manejo de reconexión y timeouts

3. **Integrar gRPC en Control Layer**:
   - Reemplazar TODOs en `control.go`
   - `Start()` → llamar `agentService.StartServer()`
   - `Stop()` → llamar `agentService.StopServer()`

4. **Implementar heartbeat de agentes**:
   - Actualizar `agent.last_heartbeat` periódicamente
   - Marcar agentes como offline si no responden

### Fase B.6 - Sistema de Logs

1. **Streaming de logs de servidores**:
   - Endpoint WebSocket: `ws://localhost:8080/api/v1/servers/:id/logs`
   - Agent envía logs vía gRPC stream
   - Backend reenvía a clientes WebSocket

2. **Almacenamiento de logs**:
   - Opcional: Guardar últimas N líneas en base de datos
   - Rotación de logs

### Fase B.7 - Métricas y Monitoreo

1. **Métricas de rendimiento**:
   - CPU, RAM, TPS del servidor
   - Jugadores conectados en tiempo real

2. **Dashboard**:
   - Estadísticas agregadas
   - Gráficos de uso de recursos

---

## ✅ Checklist de Completitud

- [x] ✅ Servicio de servidores con CRUD completo
- [x] ✅ Control de ciclo de vida (start, stop, restart)
- [x] ✅ 9 endpoints REST implementados
- [x] ✅ Validación de permisos (RBAC)
- [x] ✅ Validación de datos (go-playground/validator)
- [x] ✅ Manejo de errores HTTP (400, 401, 403, 404, 409, 500)
- [x] ✅ Integración con autenticación JWT
- [x] ✅ Paginación en listados
- [x] ✅ Preload de relaciones (Agent, User)
- [x] ✅ Estados de servidor con validaciones
- [x] ✅ Verificación de agente online
- [x] ✅ Integración en main.go
- [x] ✅ Compilación exitosa del binario
- [x] ✅ Documentación completa

---

## 📝 Notas Técnicas

### TODOs Pendientes

1. **Comunicación gRPC con Agentes** (Fase B.5):
   - `services/server/control.go` líneas con `// TODO: Communicate with agent via gRPC`
   - Requiere implementar AgentService y definir protobuf

2. **Tests Unitarios**:
   - `services/server/service_test.go`
   - `services/server/control_test.go`
   - `api/rest/handlers/server_test.go`

3. **Optimizaciones**:
   - Caché de servidores frecuentemente accedidos
   - Índices en base de datos para queries de listado
   - Rate limiting en endpoints de control

---

## 🎉 Conclusión

La **Fase B.4** está **completada exitosamente** con:

- ✅ 1,120 líneas de código Go de alta calidad
- ✅ 9 endpoints REST funcionales
- ✅ Sistema CRUD completo para servidores
- ✅ Control de ciclo de vida (start/stop/restart)
- ✅ Sistema de permisos RBAC integrado
- ✅ Validaciones robustas a múltiples niveles
- ✅ Arquitectura en capas (Service → Handler → Router)
- ✅ Compilación exitosa sin errores
- ✅ Preparado para integración gRPC en Fase B.5

**Estado del Backend AYMC**: 
- Fases B.1, B.2, B.3, B.4: ✅ Completadas
- Próxima fase: B.5 (Comunicación gRPC con Agentes)
