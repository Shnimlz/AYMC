# 📋 FASE B - Plan Detallado de Desarrollo del Backend

**Fecha de inicio**: 13 de noviembre de 2024  
**Duración estimada**: 4-6 semanas (120-180 horas)  
**Objetivo**: Crear el backend central que conecta agentes con frontend

---

## 🎯 Visión General

El backend actuará como el **cerebro del sistema AYMC**, coordinando:
- Múltiples agentes remotos (vía gRPC)
- Frontend web (vía REST API + WebSocket)
- Base de datos central (PostgreSQL)
- Sistema de autenticación (JWT)
- Marketplace de plugins (APIs externas)

---

## 📊 Orden de Ejecución Recomendado

### Semana 1: Fundamentos (Tasks 1-2)
**Objetivo**: Base sólida del proyecto

1. **Estructura y Setup** (2 días)
2. **Base de Datos** (3 días)

### Semana 2-3: Core Services (Tasks 3-4)
**Objetivo**: Funcionalidades críticas

3. **Sistema de Autenticación** (3-4 días)
4. **Pool de Agentes** (4-5 días)

### Semana 3-4: APIs (Task 5-6)
**Objetivo**: Interfaces de comunicación

5. **API REST** (4-5 días)
6. **WebSocket Real-time** (3-4 días)

### Semana 5-6: Servicios Avanzados (Task 7)
**Objetivo**: Features de valor añadido

7. **Marketplace Service** (3-4 días)
8. **Testing y Documentación** (2-3 días)

---

## 📝 Desglose Detallado por Tarea

### Task 1: Estructura y Setup (2 días)
**Prioridad**: 🔴 CRÍTICA  
**Dependencias**: Ninguna  
**Estimación**: 12-16 horas

#### Objetivos
- ✅ Crear estructura de directorios
- ✅ Inicializar go.mod
- ✅ Instalar dependencias principales
- ✅ Configurar entorno (.env)
- ✅ Setup de logging
- ✅ Dockerfile y docker-compose.yml

#### Estructura de Directorios
```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── config/
│   ├── config.go                # Configuración
│   └── config.yaml              # Valores por defecto
├── api/
│   ├── rest/
│   │   ├── server.go            # Gin server
│   │   ├── middleware/          # Auth, CORS, logging
│   │   ├── handlers/            # HTTP handlers
│   │   └── routes.go            # Definición de rutas
│   ├── websocket/
│   │   ├── hub.go               # WS hub
│   │   ├── client.go            # Cliente WS
│   │   └── messages.go          # Tipos de mensajes
│   └── grpc/
│       ├── client.go            # Cliente gRPC para agentes
│       └── pool.go              # Pool de conexiones
├── services/
│   ├── auth/                    # Autenticación JWT
│   ├── servers/                 # Gestión de servidores
│   ├── agents/                  # Pool de agentes
│   ├── marketplace/             # Plugins/mods
│   └── analyzer/                # Análisis de logs
├── database/
│   ├── db.go                    # Conexión
│   ├── models/                  # Modelos GORM
│   └── migrations/              # SQL migrations
├── pkg/
│   ├── logger/                  # Logger custom
│   └── utils/                   # Utilidades
├── tests/
│   ├── integration/
│   └── e2e/
├── .env.example
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

#### Dependencias a Instalar
```bash
# Web framework
go get github.com/gin-gonic/gin

# ORM
go get gorm.io/gorm
go get gorm.io/driver/postgres

# JWT
go get github.com/golang-jwt/jwt/v5

# WebSocket
go get github.com/gorilla/websocket

# gRPC client
go get google.golang.org/grpc
go get google.golang.org/protobuf

# Redis
go get github.com/redis/go-redis/v9

# Logging
go get go.uber.org/zap

# Config
go get github.com/spf13/viper

# UUID
go get github.com/google/uuid

# Password hashing
go get golang.org/x/crypto/bcrypt

# Validation
go get github.com/go-playground/validator/v10

# Testing
go get github.com/stretchr/testify
```

#### Archivos a Crear

**1. `cmd/server/main.go`** (50 líneas)
- Inicialización del servidor
- Setup de configuración
- Graceful shutdown

**2. `config/config.go`** (100 líneas)
- Struct de configuración
- Carga desde env/yaml
- Validación de config

**3. `.env.example`** (30 líneas)
```env
# Server
PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=aymc
DB_PASSWORD=secret
DB_NAME=aymc_db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# Agent gRPC
AGENT_GRPC_TIMEOUT=30s
```

**4. `docker-compose.yml`** (80 líneas)
- PostgreSQL
- Redis
- Backend service
- Adminer (DB admin)

**5. `Makefile`** (60 líneas)
- `make run` - Iniciar servidor
- `make test` - Ejecutar tests
- `make migrate-up` - Aplicar migraciones
- `make docker-up` - Iniciar stack completo

#### Entregables
- ✅ Proyecto compilable
- ✅ Docker Compose funcional
- ✅ Configuración cargada desde .env
- ✅ Logger configurado (Zap)
- ✅ README.md con instrucciones de setup

---

### Task 2: Base de Datos (3 días)
**Prioridad**: 🔴 CRÍTICA  
**Dependencias**: Task 1  
**Estimación**: 18-24 horas

#### Objetivos
- ✅ Diseñar schema PostgreSQL
- ✅ Crear modelos GORM
- ✅ Implementar migraciones
- ✅ Seeders de datos de prueba
- ✅ Repositorios (DAOs)

#### Schema PostgreSQL

**Tablas principales**:

**1. `users`** (Sistema de autenticación)
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user', -- admin, user, viewer
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
```

**2. `agents`** (Agentes remotos conectados)
```sql
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(100) UNIQUE NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    port INT DEFAULT 50051,
    status VARCHAR(20) DEFAULT 'offline', -- online, offline, error
    version VARCHAR(20),
    os VARCHAR(50),
    cpu_cores INT,
    memory_total BIGINT,
    disk_total BIGINT,
    last_seen TIMESTAMP,
    health_check_interval INT DEFAULT 30,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_agents_agent_id ON agents(agent_id);
CREATE INDEX idx_agents_status ON agents(status);
CREATE INDEX idx_agents_last_seen ON agents(last_seen);
```

**3. `servers`** (Servidores de Minecraft)
```sql
CREATE TABLE servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100),
    server_type VARCHAR(50), -- paper, spigot, purpur, vanilla, fabric, forge
    version VARCHAR(20),
    port INT,
    max_players INT DEFAULT 20,
    status VARCHAR(20) DEFAULT 'stopped', -- running, stopped, starting, stopping, error
    work_dir TEXT,
    java_args TEXT,
    auto_start BOOLEAN DEFAULT false,
    auto_restart BOOLEAN DEFAULT true,
    memory_min INT DEFAULT 1024, -- MB
    memory_max INT DEFAULT 2048, -- MB
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_started TIMESTAMP,
    last_stopped TIMESTAMP
);

CREATE INDEX idx_servers_agent_id ON servers(agent_id);
CREATE INDEX idx_servers_user_id ON servers(user_id);
CREATE INDEX idx_servers_status ON servers(status);
CREATE INDEX idx_servers_name ON servers(name);
```

**4. `server_metrics`** (Métricas históricas)
```sql
CREATE TABLE server_metrics (
    id BIGSERIAL PRIMARY KEY,
    server_id UUID REFERENCES servers(id) ON DELETE CASCADE,
    timestamp TIMESTAMP DEFAULT NOW(),
    cpu_percent FLOAT,
    memory_used BIGINT,
    players_online INT,
    tps FLOAT,
    uptime_seconds BIGINT
);

CREATE INDEX idx_metrics_server_timestamp ON server_metrics(server_id, timestamp DESC);

-- Particionamiento por fecha (opcional)
-- CREATE TABLE server_metrics_2024_11 PARTITION OF server_metrics
-- FOR VALUES FROM ('2024-11-01') TO ('2024-12-01');
```

**5. `plugins`** (Catálogo de plugins)
```sql
CREATE TABLE plugins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    author VARCHAR(100),
    version VARCHAR(20),
    download_url TEXT,
    icon_url TEXT,
    source VARCHAR(20), -- spigot, modrinth, curseforge, github
    source_id VARCHAR(100), -- ID en la plataforma origen
    category VARCHAR(50),
    downloads BIGINT DEFAULT 0,
    rating DECIMAL(3,2),
    minecraft_versions JSONB, -- ["1.20.1", "1.20.2"]
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_plugins_slug ON plugins(slug);
CREATE INDEX idx_plugins_source ON plugins(source, source_id);
CREATE INDEX idx_plugins_category ON plugins(category);
CREATE INDEX idx_plugins_name ON plugins USING gin(to_tsvector('english', name));
```

**6. `server_plugins`** (Many-to-many)
```sql
CREATE TABLE server_plugins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id UUID REFERENCES servers(id) ON DELETE CASCADE,
    plugin_id UUID REFERENCES plugins(id) ON DELETE CASCADE,
    version VARCHAR(20),
    is_enabled BOOLEAN DEFAULT true,
    installed_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(server_id, plugin_id)
);

CREATE INDEX idx_server_plugins_server ON server_plugins(server_id);
CREATE INDEX idx_server_plugins_plugin ON server_plugins(plugin_id);
```

**7. `backups`** (Backups de servidores)
```sql
CREATE TABLE backups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id UUID REFERENCES servers(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    path TEXT NOT NULL,
    size_bytes BIGINT,
    backup_type VARCHAR(20), -- full, world, plugins, config
    status VARCHAR(20) DEFAULT 'pending', -- pending, in_progress, completed, failed
    compression VARCHAR(10) DEFAULT 'gzip', -- gzip, zip, tar
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE INDEX idx_backups_server ON backups(server_id);
CREATE INDEX idx_backups_created_at ON backups(created_at DESC);
```

**8. `logs`** (Logs centralizados - opcional)
```sql
CREATE TABLE logs (
    id BIGSERIAL PRIMARY KEY,
    server_id UUID REFERENCES servers(id) ON DELETE CASCADE,
    timestamp TIMESTAMP DEFAULT NOW(),
    level VARCHAR(20), -- INFO, WARN, ERROR, etc.
    source VARCHAR(100),
    message TEXT,
    exception TEXT,
    stack_trace TEXT
);

-- Particionamiento recomendado para logs
CREATE INDEX idx_logs_server_timestamp ON logs(server_id, timestamp DESC);
```

**9. `sessions`** (Sesiones de usuario - opcional)
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token_hash);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
```

#### Modelos GORM

**Archivos a crear**:
- `database/models/user.go`
- `database/models/agent.go`
- `database/models/server.go`
- `database/models/plugin.go`
- `database/models/backup.go`
- `database/models/metrics.go`

**Ejemplo: `database/models/server.go`**
```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type ServerStatus string

const (
    StatusRunning  ServerStatus = "running"
    StatusStopped  ServerStatus = "stopped"
    StatusStarting ServerStatus = "starting"
    StatusStopping ServerStatus = "stopping"
    StatusError    ServerStatus = "error"
)

type Server struct {
    ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    AgentID       uuid.UUID      `gorm:"type:uuid;not null" json:"agent_id"`
    UserID        uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
    Name          string         `gorm:"size:100;not null" json:"name"`
    DisplayName   string         `gorm:"size:100" json:"display_name"`
    ServerType    string         `gorm:"size:50" json:"server_type"`
    Version       string         `gorm:"size:20" json:"version"`
    Port          int            `json:"port"`
    MaxPlayers    int            `gorm:"default:20" json:"max_players"`
    Status        ServerStatus   `gorm:"type:varchar(20);default:stopped" json:"status"`
    WorkDir       string         `gorm:"type:text" json:"work_dir"`
    JavaArgs      string         `gorm:"type:text" json:"java_args"`
    AutoStart     bool           `gorm:"default:false" json:"auto_start"`
    AutoRestart   bool           `gorm:"default:true" json:"auto_restart"`
    MemoryMin     int            `gorm:"default:1024" json:"memory_min"`
    MemoryMax     int            `gorm:"default:2048" json:"memory_max"`
    CreatedAt     time.Time      `json:"created_at"`
    UpdatedAt     time.Time      `json:"updated_at"`
    LastStarted   *time.Time     `json:"last_started,omitempty"`
    LastStopped   *time.Time     `json:"last_stopped,omitempty"`
    
    // Relaciones
    Agent         Agent          `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
    User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Plugins       []Plugin       `gorm:"many2many:server_plugins" json:"plugins,omitempty"`
    Backups       []Backup       `gorm:"foreignKey:ServerID" json:"backups,omitempty"`
}

func (Server) TableName() string {
    return "servers"
}
```

#### Migraciones

**`database/migrations/001_initial.sql`**
```sql
-- Ejecutar todos los CREATE TABLE de arriba
```

**`database/migrations/migrate.go`**
```go
package migrations

import (
    "github.com/aymc/backend/database/models"
    "gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
    return db.AutoMigrate(
        &models.User{},
        &models.Agent{},
        &models.Server{},
        &models.Plugin{},
        &models.ServerPlugin{},
        &models.Backup{},
        &models.ServerMetrics{},
    )
}
```

#### Seeders

**`database/seeders/seed.go`**
```go
package seeders

func SeedUsers(db *gorm.DB) error {
    users := []models.User{
        {
            Username:     "admin",
            Email:        "admin@aymc.local",
            PasswordHash: "$2a$10$...", // bcrypt de "admin123"
            Role:         "admin",
        },
        {
            Username:     "demo",
            Email:        "demo@aymc.local",
            PasswordHash: "$2a$10$...", // bcrypt de "demo123"
            Role:         "user",
        },
    }
    return db.Create(&users).Error
}
```

#### Entregables
- ✅ Schema PostgreSQL completo
- ✅ 9 modelos GORM implementados
- ✅ Sistema de migraciones funcional
- ✅ Seeders con datos de prueba
- ✅ Tests de conexión DB

---

### Task 3: Sistema de Autenticación (3-4 días)
**Prioridad**: 🔴 CRÍTICA  
**Dependencias**: Task 1, 2  
**Estimación**: 20-28 horas

#### Objetivos
- ✅ Implementar registro de usuarios
- ✅ Implementar login con JWT
- ✅ Refresh token mechanism
- ✅ Middleware de autenticación
- ✅ RBAC (Role-Based Access Control)
- ✅ Password reset (email)

#### Archivos a Crear

**1. `services/auth/jwt.go`** (150 líneas)
```go
type JWTService struct {
    secretKey     string
    accessExpiry  time.Duration
    refreshExpiry time.Duration
}

func (j *JWTService) GenerateTokenPair(user *models.User) (TokenPair, error)
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error)
func (j *JWTService) RefreshAccessToken(refreshToken string) (string, error)
```

**2. `services/auth/service.go`** (200 líneas)
```go
type AuthService struct {
    db         *gorm.DB
    jwtService *JWTService
}

func (a *AuthService) Register(req RegisterRequest) (*User, error)
func (a *AuthService) Login(email, password string) (*TokenPair, error)
func (a *AuthService) Logout(userID uuid.UUID) error
func (a *AuthService) ChangePassword(userID uuid.UUID, oldPass, newPass string) error
func (a *AuthService) ResetPassword(email string) error
```

**3. `api/rest/middleware/auth.go`** (100 líneas)
```go
func AuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc
func RequireRole(roles ...string) gin.HandlerFunc
func GetUserFromContext(c *gin.Context) (*models.User, error)
```

**4. `api/rest/handlers/auth.go`** (180 líneas)
```go
type AuthHandler struct {
    authService *auth.AuthService
}

func (h *AuthHandler) Register(c *gin.Context)
func (h *AuthHandler) Login(c *gin.Context)
func (h *AuthHandler) Refresh(c *gin.Context)
func (h *AuthHandler) Logout(c *gin.Context)
func (h *AuthHandler) GetProfile(c *gin.Context)
func (h *AuthHandler) UpdateProfile(c *gin.Context)
```

#### Endpoints

```
POST   /api/v1/auth/register       # Registro
POST   /api/v1/auth/login          # Login
POST   /api/v1/auth/refresh        # Refresh token
POST   /api/v1/auth/logout         # Logout
GET    /api/v1/auth/me             # Profile actual
PUT    /api/v1/auth/me             # Actualizar profile
POST   /api/v1/auth/change-password
POST   /api/v1/auth/reset-password
```

#### Roles y Permisos

| Rol | Permisos |
|-----|----------|
| **admin** | Acceso total, gestión de usuarios, ver todos los servidores |
| **user** | Gestionar sus propios servidores, instalar plugins |
| **viewer** | Solo lectura, ver métricas y logs |

#### Tests
- ✅ Test de registro
- ✅ Test de login exitoso/fallido
- ✅ Test de validación de JWT
- ✅ Test de refresh token
- ✅ Test de middleware de auth

#### Entregables
- ✅ Sistema de autenticación completo
- ✅ JWT con refresh token
- ✅ Middleware funcional
- ✅ RBAC implementado
- ✅ Tests pasando

---

### Task 4: Pool de Agentes (4-5 días)
**Prioridad**: 🟠 ALTA  
**Dependencias**: Task 1, 2, 3  
**Estimación**: 28-36 horas

#### Objetivos
- ✅ Registry de agentes conectados
- ✅ Health checks automáticos
- ✅ Balanceo de carga
- ✅ Failover automático
- ✅ Proxy gRPC transparente

#### Archivos a Crear

**1. `services/agents/registry.go`** (250 líneas)
```go
type AgentRegistry struct {
    agents     map[uuid.UUID]*AgentConnection
    mu         sync.RWMutex
    db         *gorm.DB
    healthTick time.Duration
}

func (r *AgentRegistry) Register(agent *models.Agent) error
func (r *AgentRegistry) Unregister(agentID uuid.UUID) error
func (r *AgentRegistry) GetAgent(agentID uuid.UUID) (*AgentConnection, error)
func (r *AgentRegistry) ListAgents() []*AgentConnection
func (r *AgentRegistry) SelectAgent(criteria LoadBalanceCriteria) (*AgentConnection, error)
```

**2. `services/agents/health.go`** (150 líneas)
```go
type HealthMonitor struct {
    registry *AgentRegistry
    interval time.Duration
}

func (h *HealthMonitor) Start() error
func (h *HealthMonitor) Stop()
func (h *HealthMonitor) CheckAgent(agent *AgentConnection) error
func (h *HealthMonitor) HandleFailure(agentID uuid.UUID) error
```

**3. `services/agents/connection.go`** (200 líneas)
```go
type AgentConnection struct {
    ID         uuid.UUID
    Agent      *models.Agent
    Client     pb.AgentServiceClient
    conn       *grpc.ClientConn
    lastSeen   time.Time
    status     AgentStatus
    metrics    *AgentMetrics
    mu         sync.RWMutex
}

func (ac *AgentConnection) Connect() error
func (ac *AgentConnection) Disconnect() error
func (ac *AgentConnection) IsHealthy() bool
func (ac *AgentConnection) GetClient() pb.AgentServiceClient
```

**4. `services/agents/balancer.go`** (120 líneas)
```go
type LoadBalancer struct {
    strategy LoadBalanceStrategy
}

func (lb *LoadBalancer) SelectAgent(agents []*AgentConnection) *AgentConnection
// Estrategias: RoundRobin, LeastConnections, LeastLoad
```

**5. `api/rest/handlers/agents.go`** (180 líneas)
```go
func (h *AgentHandler) ListAgents(c *gin.Context)
func (h *AgentHandler) GetAgent(c *gin.Context)
func (h *AgentHandler) RegisterAgent(c *gin.Context)
func (h *AgentHandler) UnregisterAgent(c *gin.Context)
func (h *AgentHandler) GetAgentMetrics(c *gin.Context)
func (h *AgentHandler) GetAgentHealth(c *gin.Context)
```

#### Endpoints

```
GET    /api/v1/agents              # Listar agentes
GET    /api/v1/agents/:id          # Ver agente específico
POST   /api/v1/agents              # Registrar agente
DELETE /api/v1/agents/:id          # Desregistrar agente
GET    /api/v1/agents/:id/health   # Health check manual
GET    /api/v1/agents/:id/metrics  # Métricas del agente
```

#### Health Check Automático

Cada 30 segundos:
1. Ping al agente (gRPC)
2. Verificar latencia < 5s
3. Actualizar `last_seen`
4. Si falla 3 veces → marcar como `offline`
5. Si offline > 5 min → failover de servidores

#### Failover Automático

Si un agente falla:
1. Detectar servidores asignados al agente
2. Seleccionar agente alternativo (balancer)
3. Migrar configuración
4. Notificar al usuario (WebSocket)

#### Entregables
- ✅ Registry de agentes funcional
- ✅ Health checks cada 30s
- ✅ Balanceador con 3 estrategias
- ✅ Failover automático
- ✅ Tests de conexión/desconexión

---

### Task 5: API REST (4-5 días)
**Prioridad**: 🟠 ALTA  
**Dependencias**: Task 1, 2, 3, 4  
**Estimación**: 28-36 horas

#### Objetivos
- ✅ CRUD completo de servidores
- ✅ Gestión de plugins
- ✅ Gestión de backups
- ✅ Configuraciones de servidor
- ✅ Swagger/OpenAPI documentation

#### Endpoints Completos

**Servers**
```
GET    /api/v1/servers             # Listar servidores del usuario
POST   /api/v1/servers             # Crear servidor
GET    /api/v1/servers/:id         # Ver servidor
PUT    /api/v1/servers/:id         # Actualizar servidor
DELETE /api/v1/servers/:id         # Eliminar servidor
POST   /api/v1/servers/:id/start   # Iniciar servidor
POST   /api/v1/servers/:id/stop    # Detener servidor
POST   /api/v1/servers/:id/restart # Reiniciar servidor
POST   /api/v1/servers/:id/command # Enviar comando
GET    /api/v1/servers/:id/logs    # Obtener logs (paginados)
GET    /api/v1/servers/:id/metrics # Métricas del servidor
GET    /api/v1/servers/:id/status  # Estado en tiempo real
```

**Plugins**
```
GET    /api/v1/plugins/search?q=worldedit&page=1
GET    /api/v1/plugins/:id
GET    /api/v1/servers/:id/plugins
POST   /api/v1/servers/:id/plugins/:plugin_id/install
DELETE /api/v1/servers/:id/plugins/:plugin_id
PUT    /api/v1/servers/:id/plugins/:plugin_id/enable
PUT    /api/v1/servers/:id/plugins/:plugin_id/disable
```

**Backups**
```
GET    /api/v1/servers/:id/backups
POST   /api/v1/servers/:id/backups
GET    /api/v1/backups/:id
DELETE /api/v1/backups/:id
POST   /api/v1/backups/:id/restore
POST   /api/v1/backups/:id/download
```

**Config Files**
```
GET    /api/v1/servers/:id/files?path=/plugins
GET    /api/v1/servers/:id/files/server.properties
PUT    /api/v1/servers/:id/files/server.properties
POST   /api/v1/servers/:id/files
DELETE /api/v1/servers/:id/files
```

**Users** (Admin only)
```
GET    /api/v1/users               # Listar usuarios
GET    /api/v1/users/:id
PUT    /api/v1/users/:id
DELETE /api/v1/users/:id
```

#### Swagger Documentation

Usar `swaggo/swag` para generar documentación automática:

```bash
go get github.com/swaggo/swag/cmd/swag
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
```

Acceso: `http://localhost:8080/swagger/index.html`

#### Archivos a Crear

**Handlers** (~1500 líneas totales):
- `api/rest/handlers/servers.go` (400 líneas)
- `api/rest/handlers/plugins.go` (250 líneas)
- `api/rest/handlers/backups.go` (200 líneas)
- `api/rest/handlers/files.go` (250 líneas)
- `api/rest/handlers/users.go` (200 líneas)

**Services** (~1200 líneas):
- `services/servers/service.go` (400 líneas)
- `services/plugins/service.go` (300 líneas)
- `services/backups/service.go` (250 líneas)
- `services/files/service.go` (250 líneas)

#### Entregables
- ✅ 35+ endpoints REST
- ✅ Swagger UI funcional
- ✅ Validación de requests
- ✅ Paginación en lists
- ✅ Filtros y búsqueda
- ✅ Tests de integración

---

### Task 6: WebSocket Real-time (3-4 días)
**Prioridad**: 🟡 MEDIA  
**Dependencias**: Task 1, 2, 3, 5  
**Estimación**: 20-28 horas

#### Objetivos
- ✅ Hub de WebSocket centralizado
- ✅ Autenticación vía token
- ✅ Push notifications
- ✅ Logs en tiempo real
- ✅ Métricas en tiempo real

#### Archivos a Crear

**1. `api/websocket/hub.go`** (200 líneas)
```go
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan Message
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
}

func (h *Hub) Run()
func (h *Hub) BroadcastToUser(userID uuid.UUID, msg Message)
func (h *Hub) BroadcastToServer(serverID uuid.UUID, msg Message)
```

**2. `api/websocket/client.go`** (180 líneas)
```go
type Client struct {
    hub      *Hub
    conn     *websocket.Conn
    send     chan []byte
    user     *models.User
    subscriptions map[string]bool
}

func (c *Client) ReadPump()
func (c *Client) WritePump()
func (c *Client) Subscribe(channel string)
func (c *Client) Unsubscribe(channel string)
```

**3. `api/websocket/messages.go`** (150 líneas)
```go
type MessageType string

const (
    MessageTypeLogEntry     MessageType = "log_entry"
    MessageTypeMetrics      MessageType = "metrics"
    MessageTypeServerStatus MessageType = "server_status"
    MessageTypeAlert        MessageType = "alert"
    MessageTypeNotification MessageType = "notification"
)

type Message struct {
    Type      MessageType     `json:"type"`
    Channel   string          `json:"channel"`
    Data      interface{}     `json:"data"`
    Timestamp time.Time       `json:"timestamp"`
}
```

#### Canales de Suscripción

```typescript
// Cliente se suscribe a canales específicos
{
  "type": "subscribe",
  "channels": [
    "server:uuid-123:logs",
    "server:uuid-123:metrics",
    "server:uuid-456:status",
    "user:notifications"
  ]
}

// Servidor envía mensajes
{
  "type": "log_entry",
  "channel": "server:uuid-123:logs",
  "data": {
    "timestamp": "2024-11-13T10:30:00Z",
    "level": "INFO",
    "message": "Server started"
  }
}
```

#### Integración con Agentes

Cuando un agente envía logs vía gRPC:
1. Backend recibe el stream
2. Hub broadcasta a clientes suscritos
3. Almacena en DB (opcional)

#### Endpoint

```
WS /api/v1/ws?token=jwt_token_here
```

#### Entregables
- ✅ Hub de WebSocket funcional
- ✅ Sistema de suscripciones
- ✅ 5 tipos de mensajes
- ✅ Autenticación con JWT
- ✅ Reconexión automática (client-side)
- ✅ Tests con gorilla/websocket

---

### Task 7: Marketplace Service (3-4 días)
**Prioridad**: 🟢 BAJA  
**Dependencias**: Task 1, 2, 5  
**Estimación**: 20-28 horas

#### Objetivos
- ✅ Integración con APIs externas
- ✅ Cache de resultados (Redis)
- ✅ Búsqueda unificada
- ✅ Instalación automática

#### APIs a Integrar

**1. Spigot (Spiget API)**
```
GET https://api.spiget.org/v2/search/resources/{query}
GET https://api.spiget.org/v2/resources/{id}
GET https://api.spiget.org/v2/resources/{id}/download
```

**2. Modrinth**
```
GET https://api.modrinth.com/v2/search?query={query}&facets=[[%22project_type:mod%22]]
GET https://api.modrinth.com/v2/project/{id}
GET https://api.modrinth.com/v2/project/{id}/version
```

**3. CurseForge**
```
GET https://api.curseforge.com/v1/mods/search?gameId=432&searchFilter={query}
GET https://api.curseforge.com/v1/mods/{id}
```

**Nota**: CurseForge requiere API key.

#### Archivos a Crear

**1. `services/marketplace/service.go`** (200 líneas)
```go
type MarketplaceService struct {
    spigot     *SpigotClient
    modrinth   *ModrinthClient
    curseforge *CurseForgeClient
    cache      *redis.Client
}

func (m *MarketplaceService) Search(query string, source string) ([]Plugin, error)
func (m *MarketplaceService) GetPlugin(id string, source string) (*Plugin, error)
func (m *MarketplaceService) InstallPlugin(serverID, pluginID uuid.UUID) error
```

**2. `services/marketplace/spigot.go`** (150 líneas)
**3. `services/marketplace/modrinth.go`** (150 líneas)
**4. `services/marketplace/curseforge.go`** (150 líneas)

#### Cache Strategy (Redis)

```go
// Cache de búsquedas (TTL 1 hora)
key: "marketplace:search:{source}:{query}"
value: JSON serializado de resultados

// Cache de plugins (TTL 24 horas)
key: "marketplace:plugin:{source}:{id}"
value: JSON serializado del plugin
```

#### Entregables
- ✅ 3 integraciones de APIs
- ✅ Búsqueda unificada
- ✅ Cache con Redis
- ✅ Instalación automática
- ✅ Tests con mocks

---

## 📅 Cronograma Detallado

### Semana 1 (10-12 Nov)
- **Día 1-2**: Task 1 - Estructura y Setup
  - Lunes: Estructura + dependencias + config
  - Martes: Docker Compose + Makefile + README

- **Día 3-5**: Task 2 - Base de Datos
  - Miércoles: Schema + modelos User/Agent/Server
  - Jueves: Modelos Plugin/Backup + migraciones
  - Viernes: Seeders + tests + repositorios

### Semana 2 (13-17 Nov)
- **Día 1-3**: Task 3 - Autenticación
  - Lunes: JWT service + modelos
  - Martes: Auth service + handlers
  - Miércoles: Middleware + RBAC + tests

- **Día 4-5**: Task 4 - Pool de Agentes (inicio)
  - Jueves: Registry + health monitor
  - Viernes: Connections + inicio de balancer

### Semana 3 (18-22 Nov)
- **Día 1-2**: Task 4 - Pool de Agentes (continuación)
  - Lunes: Balancer + failover
  - Martes: Handlers + tests completos

- **Día 3-5**: Task 5 - API REST (inicio)
  - Miércoles: Handlers de servers
  - Jueves: Services de servers + plugins
  - Viernes: Handlers de plugins + backups

### Semana 4 (25-29 Nov)
- **Día 1-2**: Task 5 - API REST (continuación)
  - Lunes: Files service + handlers
  - Martes: Swagger + validaciones + tests

- **Día 3-5**: Task 6 - WebSocket
  - Miércoles: Hub + client + messages
  - Jueves: Integración con agentes
  - Viernes: Tests + documentación

### Semana 5 (2-6 Dic)
- **Día 1-3**: Task 7 - Marketplace
  - Lunes: Spigot + Modrinth clients
  - Martes: CurseForge + cache Redis
  - Miércoles: Instalación automática + tests

- **Día 4-5**: Testing E2E y Documentación
  - Jueves: Tests de integración completos
  - Viernes: Documentación + deployment guide

### Semana 6 (9-10 Dic) - Buffer
- Refinamiento
- Bugs fixes
- Optimizaciones

---

## 🎯 Milestones

**Milestone 1** (Fin Semana 1): Base sólida
- ✅ Proyecto compilable
- ✅ DB funcionando
- ✅ Docker Compose operativo

**Milestone 2** (Fin Semana 2): Core funcional
- ✅ Autenticación completa
- ✅ Registry de agentes funcional

**Milestone 3** (Fin Semana 3): APIs operativas
- ✅ REST API completa
- ✅ CRUD de servidores funcional

**Milestone 4** (Fin Semana 4): Real-time
- ✅ WebSocket funcionando
- ✅ Logs en tiempo real

**Milestone 5** (Fin Semana 5): Feature complete
- ✅ Marketplace integrado
- ✅ Todos los tests pasando

**Milestone 6** (Fin Semana 6): Production ready
- ✅ Documentación completa
- ✅ Deployment probado

---

## 📊 Métricas de Éxito

| Métrica | Objetivo |
|---------|----------|
| Cobertura de tests | > 70% |
| Endpoints REST | 35+ |
| Tiempo de respuesta API | < 100ms (p95) |
| WebSocket concurrentes | > 100 clientes |
| Uptime | 99.9% |
| Documentación | 100% de endpoints |

---

## 🚀 Comandos Rápidos

```bash
# Iniciar desarrollo
make docker-up
make migrate-up
make seed
make run

# Tests
make test
make test-integration
make test-e2e

# Build
make build
make docker-build

# Deploy
make deploy-staging
make deploy-prod
```

---

## 📝 Notas Importantes

1. **Priorizar funcionalidad sobre perfección**: MVP primero
2. **Tests desde el día 1**: TDD cuando sea posible
3. **Documentar APIs**: Swagger automático
4. **Seguridad first**: Validación, sanitización, rate limiting
5. **Logs estructurados**: Zap con niveles apropiados
6. **Metrics**: Prometheus/Grafana (opcional Semana 6)
7. **CI/CD**: GitHub Actions (opcional Semana 6)

---

**¿Listo para empezar? 🚀**

El plan está completo y detallado. Podemos comenzar con Task 1 cuando estés listo.

---

*Plan creado el 13 de noviembre de 2024*
