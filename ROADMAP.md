# 🗺️ AYMC - Roadmap de Desarrollo

## 📅 Plan: Fase A (Mejoras del Agente) → Fase B (Backend Central)

**Última actualización**: 13 de noviembre de 2024

---

## 🎯 Fase A: Mejoras del Agente (1-2 semanas)

**Objetivo**: Completar funcionalidades avanzadas del agente antes de desarrollar el backend.

### 📋 Tareas

#### 1. InstallJava Automático (2-3 días)
**Prioridad**: Alta  
**Archivo**: `agent/core/installer.go` (nuevo)

**Funcionalidades**:
- ✅ Detectar sistema operativo (Linux, Windows, macOS)
- ✅ Detectar distribución Linux (Debian/Ubuntu, RHEL/CentOS, Arch, Alpine)
- ✅ Instalar Java según el gestor de paquetes:
  - `apt-get install openjdk-21-jdk` (Debian/Ubuntu)
  - `yum install java-21-openjdk` (RHEL/CentOS)
  - `pacman -S jdk-openjdk` (Arch)
  - `choco install openjdk` (Windows con Chocolatey)
  - `brew install openjdk@21` (macOS con Homebrew)
- ✅ Verificar versión post-instalación
- ✅ Progress reporting vía gRPC

**Tests**:
- Unit tests para detección de SO
- Integration test en container Docker
- Test de rollback si falla instalación

---

#### 2. DownloadServer con Progress (2-3 días)
**Prioridad**: Alta  
**Archivo**: `agent/core/downloader.go` (nuevo)

**Funcionalidades**:
- ✅ Descargar JARs de servidores populares:
  - **Paper**: https://api.papermc.io/v2/projects/paper
  - **Spigot**: BuildTools.jar + compilación
  - **Purpur**: https://api.purpurmc.org/v2/purpur
  - **Fabric**: https://meta.fabricmc.net/v2/versions/loader
  - **Forge**: https://files.minecraftforge.net/net/minecraftforge/forge/
- ✅ Progress streaming con porcentaje de descarga
- ✅ Validación de checksums SHA256
- ✅ Retry automático con backoff exponencial
- ✅ Cache de versiones descargadas

**Tests**:
- Mock de HTTP responses
- Test de validación de checksums
- Test de progress reporting

---

#### 3. Parser de Logs Avanzado (2 días)
**Prioridad**: Media  
**Archivo**: `agent/core/logparser.go` (extender)

**Mejoras**:
- ✅ Más patrones de errores:
  - Crash reports completos
  - Plugin-specific errors (WorldEdit, EssentialsX, etc.)
  - ClassNotFoundException con sugerencias
  - Database connection errors
  - Permission issues
- ✅ Stack trace completo extraction
- ✅ Sugerencias automáticas de fix basadas en patterns
- ✅ Detección de plugins instalados desde logs

**Ejemplo de sugerencia**:
```
Error: java.lang.ClassNotFoundException: com.mysql.jdbc.Driver
Sugerencia: Instalar plugin MySQL Connector. Descarga: https://...
```

---

#### 4. Tests de Integración gRPC (1-2 días)
**Prioridad**: Alta  
**Archivo**: `agent/tests/integration_test.go` (nuevo)

**Cobertura**:
- ✅ Server completo con TLS
- ✅ Cliente gRPC conectándose
- ✅ Test de todos los 20+ métodos end-to-end
- ✅ Test de autenticación con tokens
- ✅ Test de StreamLogs bidireccional
- ✅ Test de concurrencia (múltiples clientes)

**Setup**:
- Docker Compose con agente + mock servers
- Certificados de prueba
- Data fixtures

---

#### 5. Benchmarks de Rendimiento (1 día)
**Prioridad**: Baja  
**Archivo**: `agent/benchmarks/` (nuevo)

**Benchmarks**:
```go
BenchmarkParseLog-8              500000    2345 ns/op
BenchmarkGetSystemMetrics-8      100000   12450 ns/op
BenchmarkStreamLogs-8             50000   23450 ns/op
```

**Objetivos**:
- ParseLog: < 5ms por línea
- GetSystemMetrics: < 20ms
- StreamLogs: > 1000 líneas/segundo

---

## 🏗️ Fase B: Backend Central (4-6 semanas)

**Objetivo**: Crear el cerebro del sistema que coordina agentes y frontend.

### 📋 Arquitectura

```
backend/
├── main.go                      # Entry point
├── config/
│   ├── config.go                # Configuración (env vars)
│   └── config.yaml              # Valores por defecto
├── api/
│   ├── rest/                    # REST API (Gin)
│   │   ├── server.go
│   │   ├── middleware/          # Auth, CORS, logging
│   │   ├── handlers/
│   │   │   ├── auth.go          # Login, register
│   │   │   ├── servers.go       # CRUD servidores
│   │   │   ├── agents.go        # Gestión de agentes
│   │   │   ├── plugins.go       # Marketplace
│   │   │   └── backups.go       # Backups
│   │   └── routes.go
│   ├── websocket/               # WebSocket server
│   │   ├── hub.go               # Connection pool
│   │   ├── client.go            # Client handler
│   │   └── messages.go          # Message types
│   └── grpc/                    # gRPC client
│       ├── agent_client.go      # Cliente para agentes
│       └── pool.go              # Pool de conexiones
├── services/
│   ├── auth/                    # Autenticación JWT
│   │   ├── jwt.go
│   │   ├── middleware.go
│   │   └── roles.go
│   ├── servers/                 # Lógica de servidores
│   │   ├── manager.go
│   │   └── operations.go
│   ├── agents/                  # Pool de agentes
│   │   ├── registry.go
│   │   ├── health.go
│   │   └── balancer.go
│   ├── marketplace/             # Catálogo de plugins
│   │   ├── spigot.go
│   │   ├── modrinth.go
│   │   ├── curseforge.go
│   │   └── cache.go
│   └── analyzer/                # Análisis de logs (IA)
│       ├── patterns.go
│       └── suggestions.go
├── database/
│   ├── db.go                    # Conexión
│   ├── models/
│   │   ├── user.go
│   │   ├── server.go
│   │   ├── agent.go
│   │   ├── plugin.go
│   │   └── backup.go
│   └── migrations/
│       └── 001_initial.sql
└── tests/
    ├── integration/
    └── e2e/
```

---

### 📋 Tareas de la Fase B

#### Semana 1-2: Fundamentos

##### 1. Setup del Proyecto (1 día)
- ✅ Estructura de directorios
- ✅ Go modules (`go mod init`)
- ✅ Dependencias:
  ```bash
  go get github.com/gin-gonic/gin
  go get gorm.io/gorm
  go get gorm.io/driver/postgres
  go get github.com/golang-jwt/jwt/v5
  go get github.com/gorilla/websocket
  go get google.golang.org/grpc
  go get github.com/redis/go-redis/v9
  ```
- ✅ Dockerfile + docker-compose.yml

##### 2. Base de Datos (2-3 días)
**Schema PostgreSQL**:

```sql
-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Agents
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(100) UNIQUE NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    port INT DEFAULT 50051,
    status VARCHAR(20) DEFAULT 'offline',
    version VARCHAR(20),
    os VARCHAR(50),
    last_seen TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Servers
CREATE TABLE servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    server_type VARCHAR(50), -- paper, spigot, etc.
    version VARCHAR(20),
    port INT,
    max_players INT,
    status VARCHAR(20) DEFAULT 'stopped',
    work_dir TEXT,
    java_args TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Plugins
CREATE TABLE plugins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    author VARCHAR(100),
    version VARCHAR(20),
    download_url TEXT,
    source VARCHAR(20), -- spigot, modrinth, curseforge
    downloads INT DEFAULT 0,
    rating DECIMAL(3,2),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Server_Plugins (many-to-many)
CREATE TABLE server_plugins (
    server_id UUID REFERENCES servers(id) ON DELETE CASCADE,
    plugin_id UUID REFERENCES plugins(id) ON DELETE CASCADE,
    installed_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (server_id, plugin_id)
);

-- Backups
CREATE TABLE backups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id UUID REFERENCES servers(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    size_bytes BIGINT,
    backup_type VARCHAR(20), -- full, world, plugins
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Implementar**:
- Modelos GORM
- Migraciones automáticas
- Seeders con datos de prueba

##### 3. Sistema de Autenticación (2 días)
- ✅ Registro de usuarios con bcrypt
- ✅ Login con JWT (access + refresh token)
- ✅ Middleware de autenticación
- ✅ RBAC (Role-Based Access Control)
  - `admin`: acceso total
  - `user`: solo sus servidores
  - `viewer`: solo lectura

---

#### Semana 3-4: APIs y Comunicación

##### 4. API REST (3-4 días)
**Endpoints**:

```
Auth:
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout

Servers:
GET    /api/v1/servers           # Listar todos
POST   /api/v1/servers           # Crear servidor
GET    /api/v1/servers/:id       # Ver detalles
PUT    /api/v1/servers/:id       # Actualizar
DELETE /api/v1/servers/:id       # Eliminar
POST   /api/v1/servers/:id/start
POST   /api/v1/servers/:id/stop
POST   /api/v1/servers/:id/restart
POST   /api/v1/servers/:id/command
GET    /api/v1/servers/:id/logs
GET    /api/v1/servers/:id/metrics

Agents:
GET    /api/v1/agents            # Listar agentes
GET    /api/v1/agents/:id        # Detalles del agente
GET    /api/v1/agents/:id/health

Plugins:
GET    /api/v1/plugins/search?q=worldedit
GET    /api/v1/plugins/:id
POST   /api/v1/servers/:id/plugins/:plugin_id/install

Backups:
GET    /api/v1/servers/:id/backups
POST   /api/v1/servers/:id/backups
POST   /api/v1/backups/:id/restore
DELETE /api/v1/backups/:id
```

**Documentación**:
- Swagger UI en `/api/docs`

##### 5. Cliente gRPC para Agentes (2 días)
- ✅ Pool de conexiones a múltiples agentes
- ✅ Retry automático con circuit breaker
- ✅ Load balancing entre agentes
- ✅ Health checks periódicos
- ✅ TLS mutual authentication

##### 6. WebSocket Server (2 días)
**Mensajes en tiempo real**:

```typescript
// Cliente → Servidor
{
  "type": "subscribe_logs",
  "server_id": "uuid"
}

// Servidor → Cliente
{
  "type": "log_entry",
  "server_id": "uuid",
  "timestamp": "2024-11-13T10:30:00Z",
  "level": "INFO",
  "message": "Server started"
}

{
  "type": "metrics",
  "server_id": "uuid",
  "cpu": 45.2,
  "ram": 2048,
  "players": 5
}

{
  "type": "alert",
  "severity": "error",
  "message": "OutOfMemoryError detected"
}
```

---

#### Semana 5-6: Servicios Avanzados

##### 7. Pool de Agentes (2-3 días)
- ✅ Registry de agentes conectados
- ✅ Health monitoring cada 30s
- ✅ Auto-reconnect si se cae conexión
- ✅ Failover: si un agente falla, migrar servidores a otro
- ✅ Métricas agregadas de todos los agentes

##### 8. Marketplace Service (2-3 días)
**Integraciones**:

1. **Spigot API**:
   ```
   GET https://api.spiget.org/v2/search/resources/{query}
   GET https://api.spiget.org/v2/resources/{id}
   ```

2. **Modrinth API**:
   ```
   GET https://api.modrinth.com/v2/search?query={query}
   GET https://api.modrinth.com/v2/project/{id}
   ```

3. **CurseForge API**:
   ```
   GET https://api.curseforge.com/v1/mods/search
   ```

**Cache**:
- Redis para resultados de búsqueda (TTL 1 hora)
- Cache de metadatos de plugins populares

##### 9. Analyzer Service (2 días)
- ✅ Análisis de logs con IA (opcional: OpenAI API)
- ✅ Detección de problemas recurrentes
- ✅ Sugerencias automáticas
- ✅ Generación de reportes

---

## 📊 Estimación de Tiempo

| Fase | Duración | Esfuerzo |
|------|----------|----------|
| **Fase A: Mejoras del Agente** | 1-2 semanas | ~40-60 horas |
| **Fase B: Backend Central** | 4-6 semanas | ~120-180 horas |
| **TOTAL** | **5-8 semanas** | **160-240 horas** |

---

## 🎯 Milestones

### Milestone 1: Agente Completo (Fin Semana 2)
- ✅ InstallJava funcional
- ✅ DownloadServer funcional
- ✅ Parser mejorado
- ✅ Tests de integración pasando
- ✅ Benchmarks aceptables

### Milestone 2: Backend MVP (Fin Semana 4)
- ✅ API REST completa
- ✅ Autenticación JWT
- ✅ Base de datos funcionando
- ✅ 1 agente conectado

### Milestone 3: Sistema Real-time (Fin Semana 6)
- ✅ WebSocket funcionando
- ✅ Logs en tiempo real
- ✅ Métricas en vivo

### Milestone 4: Marketplace (Fin Semana 8)
- ✅ Búsqueda de plugins
- ✅ Instalación automática
- ✅ Cache funcionando

---

## 🚀 Next Steps

### Empezar Ahora (Fase A):

```bash
cd /home/shni/Documents/GitHub/AYMC/agent

# Crear nuevos archivos
touch core/installer.go
touch core/downloader.go
touch tests/integration_test.go
touch benchmarks/parser_bench_test.go

# Actualizar Makefile
make test-integration
make bench
```

### Después (Fase B):

```bash
cd /home/shni/Documents/GitHub/AYMC

# Crear estructura backend
mkdir -p backend/{config,api/{rest,websocket,grpc},services/{auth,servers,agents,marketplace,analyzer},database/{models,migrations},tests/{integration,e2e}}

# Inicializar módulo Go
cd backend
go mod init github.com/aymc/backend
```

---

## 📝 Notas

- **Priorizar funcionalidad sobre perfección**: MVP primero, optimizaciones después
- **Tests desde el inicio**: Cada feature con sus tests
- **Documentar APIs**: Swagger/OpenAPI para REST, comentarios en proto para gRPC
- **Seguridad first**: Nunca hardcodear secrets, siempre usar env vars

---

*Roadmap actualizado el 13 de noviembre de 2024*
