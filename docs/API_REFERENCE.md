# API Reference - AYMC Backend

**Base URL:** `http://localhost:8080`  
**API Version:** v1  
**Authentication:** JWT Bearer Token

## 📑 Tabla de Contenidos

1. [Autenticación](#autenticación)
2. [Servidores](#servidores)
3. [Agentes](#agentes)
4. [Marketplace](#marketplace)
5. [Backups](#backups)
6. [WebSocket](#websocket)
7. [Códigos de Error](#códigos-de-error)

---

## 🔐 Autenticación

Todos los endpoints protegidos requieren un header:
```
Authorization: Bearer <token>
```

### POST /api/v1/auth/register

Registrar nuevo usuario.

**Request:**
```json
{
  "username": "admin",
  "email": "admin@example.com",
  "password": "SecurePass123!"
}
```

**Response 201:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "admin",
  "email": "admin@example.com",
  "role": "user",
  "created_at": "2025-11-13T12:00:00Z"
}
```

**Errores:**
- `400`: Validación fallida (username ya existe, email inválido, etc.)

---

### POST /api/v1/auth/login

Iniciar sesión.

**Request:**
```json
{
  "username": "admin",
  "password": "SecurePass123!"
}
```

**Response 200:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "admin",
    "email": "admin@example.com",
    "role": "user"
  }
}
```

**Errores:**
- `401`: Credenciales inválidas

---

### POST /api/v1/auth/refresh

Refrescar token de acceso.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response 200:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

---

### GET /api/v1/auth/me

Obtener perfil del usuario autenticado.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "admin",
  "email": "admin@example.com",
  "role": "user",
  "created_at": "2025-11-13T12:00:00Z"
}
```

---

### POST /api/v1/auth/logout

Cerrar sesión (invalida el token).

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "message": "Logged out successfully"
}
```

---

### POST /api/v1/auth/change-password

Cambiar contraseña del usuario.

**Headers:**
```
Authorization: Bearer <token>
```

**Request:**
```json
{
  "current_password": "OldPass123!",
  "new_password": "NewPass456!"
}
```

**Response 200:**
```json
{
  "message": "Password changed successfully"
}
```

**Errores:**
- `401`: Contraseña actual incorrecta
- `400`: Nueva contraseña no cumple requisitos

---

## 🖥️ Servidores

### GET /api/v1/servers

Listar todos los servidores del usuario.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "servers": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Survival Server",
      "type": "paper",
      "version": "1.20.2",
      "status": "running",
      "port": 25565,
      "max_players": 20,
      "agent_id": "agent-001",
      "created_at": "2025-11-01T10:00:00Z",
      "updated_at": "2025-11-13T12:00:00Z"
    }
  ],
  "total": 1
}
```

---

### POST /api/v1/servers

Crear nuevo servidor.

**Headers:**
```
Authorization: Bearer <token>
```

**Request:**
```json
{
  "name": "Creative Server",
  "type": "paper",
  "version": "1.20.2",
  "port": 25566,
  "max_players": 10,
  "agent_id": "agent-001",
  "java_version": "17",
  "min_ram": "2G",
  "max_ram": "4G"
}
```

**Response 201:**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "name": "Creative Server",
  "type": "paper",
  "version": "1.20.2",
  "status": "stopped",
  "port": 25566,
  "max_players": 10,
  "agent_id": "agent-001",
  "created_at": "2025-11-13T12:30:00Z"
}
```

**Errores:**
- `400`: Validación fallida
- `404`: Agente no encontrado
- `409`: Puerto ya en uso

---

### GET /api/v1/servers/:id

Obtener detalles de un servidor.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Survival Server",
  "type": "paper",
  "version": "1.20.2",
  "status": "running",
  "port": 25565,
  "max_players": 20,
  "agent_id": "agent-001",
  "java_version": "17",
  "min_ram": "2G",
  "max_ram": "4G",
  "auto_start": true,
  "created_at": "2025-11-01T10:00:00Z",
  "updated_at": "2025-11-13T12:00:00Z",
  "agent": {
    "id": "agent-001",
    "name": "Main Agent",
    "status": "online"
  }
}
```

---

### PUT /api/v1/servers/:id

Actualizar configuración del servidor.

**Headers:**
```
Authorization: Bearer <token>
```

**Request:**
```json
{
  "name": "Survival Server Updated",
  "max_players": 30,
  "max_ram": "6G"
}
```

**Response 200:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Survival Server Updated",
  "max_players": 30,
  "max_ram": "6G",
  "updated_at": "2025-11-13T12:45:00Z"
}
```

---

### DELETE /api/v1/servers/:id

Eliminar servidor.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "message": "Server deleted successfully"
}
```

**Errores:**
- `400`: Servidor está corriendo (debe detenerse primero)
- `404`: Servidor no encontrado

---

### POST /api/v1/servers/:id/start

Iniciar servidor.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "success": true,
  "message": "Server started successfully",
  "status": "starting"
}
```

---

### POST /api/v1/servers/:id/stop

Detener servidor.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "success": true,
  "message": "Server stopped successfully",
  "status": "stopping"
}
```

---

### POST /api/v1/servers/:id/restart

Reiniciar servidor.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "success": true,
  "message": "Server restarted successfully",
  "status": "restarting"
}
```

---

### GET /api/v1/servers/:id/status

Obtener estado actual del servidor.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "uptime_seconds": 3600,
  "players_online": 5,
  "cpu_usage": 35.2,
  "memory_usage": 2048,
  "last_updated": "2025-11-13T12:50:00Z"
}
```

---

## 🤖 Agentes

### GET /api/v1/agents

Listar todos los agentes.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "agents": [
    {
      "id": "agent-001",
      "name": "Main Agent",
      "host": "192.168.1.100",
      "port": 50051,
      "status": "online",
      "version": "0.1.0",
      "platform": "linux",
      "active_servers": 3,
      "max_servers": 10,
      "last_heartbeat": "2025-11-13T12:49:30Z"
    }
  ],
  "total": 1
}
```

---

### GET /api/v1/agents/:id

Obtener detalles de un agente.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "id": "agent-001",
  "name": "Main Agent",
  "host": "192.168.1.100",
  "port": 50051,
  "status": "online",
  "version": "0.1.0",
  "platform": "linux",
  "uptime_seconds": 86400,
  "active_servers": 3,
  "max_servers": 10,
  "last_heartbeat": "2025-11-13T12:49:30Z"
}
```

---

### GET /api/v1/agents/:id/health

Verificar salud del agente.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "status": "healthy",
  "response_time_ms": 12,
  "timestamp": "2025-11-13T12:50:00Z"
}
```

---

### GET /api/v1/agents/:id/metrics

Obtener métricas del sistema del agente.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "timestamp": "2025-11-13T12:50:00Z",
  "cpu_percent": 25.5,
  "memory_total": 16777216000,
  "memory_used": 8388608000,
  "memory_percent": 50.0,
  "disk_total": 1000000000000,
  "disk_used": 500000000000,
  "disk_percent": 50.0,
  "network_sent": 1048576000,
  "network_recv": 2097152000,
  "open_ports": [25565, 25566, 50051]
}
```

---

### GET /api/v1/agents/stats

Obtener estadísticas globales de agentes.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "total_agents": 2,
  "online_agents": 2,
  "offline_agents": 0,
  "total_servers": 5,
  "running_servers": 3,
  "stopped_servers": 2
}
```

---

## 🛒 Marketplace

### GET /api/v1/marketplace/search

Buscar plugins en Modrinth y Spigot.

**Headers:**
```
Authorization: Bearer <token>
```

**Query Parameters:**
- `query`: Término de búsqueda (requerido)
- `source`: `modrinth`, `spigot`, o `all` (default: `all`)
- `limit`: Número de resultados (default: 20, max: 100)
- `offset`: Paginación (default: 0)

**Ejemplo:**
```
GET /api/v1/marketplace/search?query=worldedit&source=all&limit=10
```

**Response 200:**
```json
{
  "plugins": [
    {
      "id": "worldedit",
      "source": "modrinth",
      "name": "WorldEdit",
      "description": "In-game map editor for Minecraft",
      "author": "sk89q",
      "downloads": 50000000,
      "rating": 4.8,
      "latest_version": "7.2.16",
      "icon_url": "https://...",
      "project_url": "https://modrinth.com/plugin/worldedit"
    },
    {
      "id": "1347",
      "source": "spigot",
      "name": "WorldEdit",
      "description": "In-game map editor",
      "author": "sk89q",
      "downloads": 10000000,
      "rating": 4.9,
      "latest_version": "7.2.16",
      "icon_url": "https://...",
      "project_url": "https://www.spigotmc.org/resources/1347/"
    }
  ],
  "total": 2,
  "source": "all"
}
```

---

### GET /api/v1/marketplace/:source/:id

Obtener detalles de un plugin específico.

**Headers:**
```
Authorization: Bearer <token>
```

**Parámetros:**
- `source`: `modrinth` o `spigot`
- `id`: ID del plugin en la fuente

**Ejemplo:**
```
GET /api/v1/marketplace/modrinth/worldedit
```

**Response 200:**
```json
{
  "id": "worldedit",
  "source": "modrinth",
  "name": "WorldEdit",
  "description": "WorldEdit is an in-game map editor for Minecraft...",
  "author": "sk89q",
  "downloads": 50000000,
  "followers": 125000,
  "rating": 4.8,
  "latest_version": "7.2.16",
  "icon_url": "https://...",
  "project_url": "https://modrinth.com/plugin/worldedit",
  "source_url": "https://github.com/EngineHub/WorldEdit",
  "license": "GPL-3.0",
  "categories": ["utility", "world-generation"],
  "game_versions": ["1.20.2", "1.20.1", "1.19.4"],
  "loaders": ["paper", "spigot", "bukkit"],
  "created_at": "2010-09-18T00:00:00Z",
  "updated_at": "2025-11-01T10:00:00Z"
}
```

---

### GET /api/v1/marketplace/:source/:id/versions

Obtener versiones disponibles de un plugin.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "versions": [
    {
      "id": "abc123",
      "version_number": "7.2.16",
      "game_versions": ["1.20.2", "1.20.1"],
      "loaders": ["paper", "spigot"],
      "downloads": 1250000,
      "date_published": "2025-10-15T10:00:00Z",
      "changelog": "- Fixed bugs\n- Added features",
      "files": [
        {
          "filename": "worldedit-bukkit-7.2.16.jar",
          "url": "https://...",
          "size": 5242880,
          "sha512": "abc123..."
        }
      ]
    }
  ],
  "total": 1
}
```

---

### GET /api/v1/marketplace/servers/:server_id/plugins

Listar plugins instalados en un servidor.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "plugins": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "WorldEdit",
      "version": "7.2.16",
      "enabled": true,
      "source": "modrinth",
      "source_id": "worldedit",
      "installed_at": "2025-11-10T15:00:00Z",
      "file_size": 5242880
    }
  ],
  "total": 1
}
```

---

### POST /api/v1/marketplace/servers/:server_id/plugins/install

Instalar plugin en un servidor.

**Headers:**
```
Authorization: Bearer <token>
```

**Request:**
```json
{
  "source": "modrinth",
  "plugin_id": "worldedit",
  "version_id": "abc123",
  "restart_server": false
}
```

**Response 200:**
```json
{
  "success": true,
  "message": "Plugin installed successfully",
  "plugin": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "WorldEdit",
    "version": "7.2.16",
    "enabled": true,
    "installed_at": "2025-11-13T13:00:00Z"
  }
}
```

**Errores:**
- `400`: Versión incompatible con el servidor
- `404`: Plugin o servidor no encontrado
- `409`: Plugin ya instalado

---

### POST /api/v1/marketplace/servers/:server_id/plugins/uninstall

Desinstalar plugin de un servidor.

**Headers:**
```
Authorization: Bearer <token>
```

**Request:**
```json
{
  "plugin_id": "550e8400-e29b-41d4-a716-446655440000",
  "delete_config": false,
  "delete_data": false,
  "restart_server": false
}
```

**Response 200:**
```json
{
  "success": true,
  "message": "Plugin uninstalled successfully"
}
```

---

### POST /api/v1/marketplace/servers/:server_id/plugins/update

Actualizar plugin en un servidor.

**Headers:**
```
Authorization: Bearer <token>
```

**Request:**
```json
{
  "plugin_id": "550e8400-e29b-41d4-a716-446655440000",
  "version_id": "xyz789",
  "restart_server": false
}
```

**Response 200:**
```json
{
  "success": true,
  "message": "Plugin updated successfully",
  "old_version": "7.2.15",
  "new_version": "7.2.16"
}
```

---

## 💾 Backups

### GET /api/v1/servers/:server_id/backups

Listar backups de un servidor.

**Headers:**
```
Authorization: Bearer <token>
```

**Query Parameters:**
- `limit`: Número de resultados (default: 20, max: 100)
- `offset`: Paginación (default: 0)

**Response 200:**
```json
{
  "backups": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440000",
      "server_id": "550e8400-e29b-41d4-a716-446655440000",
      "filename": "backup-2025-11-13.tar.gz",
      "path": "/backups/backup-2025-11-13.tar.gz",
      "size_bytes": 524288000,
      "backup_type": "full",
      "status": "completed",
      "compression": "gzip",
      "created_by": "admin",
      "created_at": "2025-11-13T03:00:00Z",
      "completed_at": "2025-11-13T03:05:30Z"
    }
  ],
  "total": 1,
  "total_size_bytes": 524288000,
  "oldest_backup": "2025-11-01T03:00:00Z",
  "latest_backup": "2025-11-13T03:00:00Z"
}
```

---

### POST /api/v1/servers/:server_id/backups

Crear nuevo backup.

**Headers:**
```
Authorization: Bearer <token>
```

**Request:**
```json
{
  "filename": "manual-backup-2025-11-13",
  "backup_type": "full",
  "compression": "gzip"
}
```

**Response 201:**
```json
{
  "id": "880e8400-e29b-41d4-a716-446655440000",
  "server_id": "550e8400-e29b-41d4-a716-446655440000",
  "filename": "manual-backup-2025-11-13.tar.gz",
  "status": "pending",
  "backup_type": "full",
  "created_at": "2025-11-13T13:30:00Z"
}
```

---

### POST /api/v1/servers/:server_id/backups/manual

Ejecutar backup manual inmediato.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 201:**
```json
{
  "id": "990e8400-e29b-41d4-a716-446655440000",
  "filename": "manual-backup-1699876800.tar.gz",
  "status": "in_progress",
  "created_at": "2025-11-13T13:35:00Z"
}
```

---

### GET /api/v1/backups/:backup_id

Obtener detalles de un backup.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440000",
  "server_id": "550e8400-e29b-41d4-a716-446655440000",
  "filename": "backup-2025-11-13.tar.gz",
  "path": "/backups/backup-2025-11-13.tar.gz",
  "size_bytes": 524288000,
  "backup_type": "full",
  "status": "completed",
  "compression": "gzip",
  "created_by": "admin",
  "created_at": "2025-11-13T03:00:00Z",
  "completed_at": "2025-11-13T03:05:30Z",
  "server": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Survival Server"
  }
}
```

---

### DELETE /api/v1/backups/:backup_id

Eliminar backup.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "message": "Backup eliminado exitosamente"
}
```

---

### POST /api/v1/backups/:backup_id/restore

Restaurar backup.

**Headers:**
```
Authorization: Bearer <token>
```

**Request:**
```json
{
  "server_id": "550e8400-e29b-41d4-a716-446655440000",
  "stop_server": true,
  "backup_before_restore": true,
  "restore_world": true,
  "restore_plugins": true,
  "restore_config": false
}
```

**Response 200:**
```json
{
  "message": "Restauración de backup iniciada"
}
```

---

### GET /api/v1/servers/:server_id/backup-config

Obtener configuración de backups automáticos.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "server_id": "550e8400-e29b-41d4-a716-446655440000",
  "enabled": true,
  "auto_backup": true,
  "schedule": "0 3 * * *",
  "backup_type": "full",
  "max_backups": 7,
  "retention_days": 30,
  "compress_backups": true,
  "include_world": true,
  "include_plugins": true,
  "include_config": true,
  "include_logs": false,
  "exclude_paths": [],
  "notify_on_complete": false,
  "notify_on_failure": true,
  "storage_type": "local",
  "last_backup_at": "2025-11-13T03:00:00Z",
  "next_backup_at": "2025-11-14T03:00:00Z"
}
```

---

### PUT /api/v1/servers/:server_id/backup-config

Actualizar configuración de backups.

**Headers:**
```
Authorization: Bearer <token>
```

**Request:**
```json
{
  "enabled": true,
  "auto_backup": true,
  "schedule": "0 2 * * *",
  "max_backups": 10,
  "retention_days": 60
}
```

**Response 200:**
```json
{
  "server_id": "550e8400-e29b-41d4-a716-446655440000",
  "enabled": true,
  "auto_backup": true,
  "schedule": "0 2 * * *",
  "max_backups": 10,
  "retention_days": 60,
  "next_backup_at": "2025-11-14T02:00:00Z"
}
```

---

### GET /api/v1/servers/:server_id/backup-stats

Obtener estadísticas de backups.

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "total_backups": 15,
  "total_size_bytes": 7864320000,
  "successful_backups": 14,
  "failed_backups": 1,
  "oldest_backup": "2025-10-01T03:00:00Z",
  "latest_backup": "2025-11-13T03:00:00Z",
  "average_size_bytes": 524288000,
  "average_duration_ms": 330000
}
```

---

## 🔌 WebSocket

### GET /api/v1/ws

Conectar al servidor WebSocket para logs en tiempo real.

**Query Parameters:**
- `token`: JWT token para autenticación

**Ejemplo:**
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws?token=YOUR_JWT_TOKEN');

ws.onopen = () => {
  // Suscribirse a logs de un servidor
  ws.send(JSON.stringify({
    type: 'subscribe',
    server_id: '550e8400-e29b-41d4-a716-446655440000'
  }));
};

ws.onmessage = (event) => {
  const log = JSON.parse(event.data);
  console.log(log);
};
```

**Mensaje de log recibido:**
```json
{
  "type": "log",
  "server_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-11-13T13:45:23Z",
  "level": "INFO",
  "source": "STDOUT",
  "message": "[13:45:23] [Server thread/INFO]: Player joined the game",
  "plugin": "",
  "file": "",
  "line": 0
}
```

**Tipos de mensajes:**

1. **subscribe**: Suscribirse a logs
```json
{
  "type": "subscribe",
  "server_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

2. **unsubscribe**: Desuscribirse
```json
{
  "type": "unsubscribe",
  "server_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

3. **ping**: Mantener conexión viva
```json
{
  "type": "ping"
}
```

---

## ❌ Códigos de Error

### Respuestas de Error

Formato estándar:
```json
{
  "error": "Mensaje de error descriptivo",
  "code": "ERROR_CODE",
  "details": {
    "field": "additional info"
  }
}
```

### Códigos HTTP

| Código | Significado | Uso |
|--------|-------------|-----|
| 200 | OK | Solicitud exitosa |
| 201 | Created | Recurso creado exitosamente |
| 400 | Bad Request | Validación fallida, parámetros incorrectos |
| 401 | Unauthorized | Token inválido o expirado |
| 403 | Forbidden | Sin permisos para realizar la acción |
| 404 | Not Found | Recurso no encontrado |
| 409 | Conflict | Conflicto (recurso duplicado, estado inválido) |
| 500 | Internal Server Error | Error del servidor |

### Ejemplos de Errores Comunes

**401 Unauthorized:**
```json
{
  "error": "Invalid or expired token"
}
```

**400 Bad Request:**
```json
{
  "error": "Validation failed",
  "details": {
    "username": "Username already exists",
    "email": "Invalid email format"
  }
}
```

**404 Not Found:**
```json
{
  "error": "Server not found"
}
```

**409 Conflict:**
```json
{
  "error": "Port 25565 is already in use"
}
```

---

## 📝 Notas Adicionales

### Paginación

Los endpoints que retornan listas soportan paginación:

- `limit`: Número de resultados (default: 20, max: 100)
- `offset`: Posición inicial (default: 0)

### Expresiones Cron

Para backups automáticos, usa formato cron:
- `0 3 * * *` - Diario a las 3 AM
- `0 */6 * * *` - Cada 6 horas
- `0 0 * * 0` - Semanal (Domingos a medianoche)
- `0 0 1 * *` - Mensual (Primer día del mes)

### Tipos de Servidor

Valores válidos para `type`:
- `paper`
- `purpur`
- `velocity`
- `waterfall`
- `spigot`
- `bukkit`

### Tipos de Backup

Valores válidos para `backup_type`:
- `full` - Backup completo
- `world` - Solo mundos
- `plugins` - Solo plugins
- `config` - Solo configuración

### Niveles de Log

- `INFO` - Información general
- `WARN` - Advertencias
- `ERROR` - Errores
- `DEBUG` - Información de depuración

---

## 🧪 Testing con cURL

### Registro y Login

```bash
# Registrar usuario
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"Test123!"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"Test123!"}' \
  | jq -r '.access_token' > token.txt

# Guardar token en variable
export TOKEN=$(cat token.txt)
```

### Servidores

```bash
# Listar servidores
curl http://localhost:8080/api/v1/servers \
  -H "Authorization: Bearer $TOKEN"

# Crear servidor
curl -X POST http://localhost:8080/api/v1/servers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Server",
    "type": "paper",
    "version": "1.20.2",
    "port": 25565,
    "agent_id": "agent-001",
    "max_ram": "4G"
  }'

# Iniciar servidor
curl -X POST http://localhost:8080/api/v1/servers/SERVER_ID/start \
  -H "Authorization: Bearer $TOKEN"
```

### Marketplace

```bash
# Buscar plugins
curl "http://localhost:8080/api/v1/marketplace/search?query=worldedit" \
  -H "Authorization: Bearer $TOKEN"

# Instalar plugin
curl -X POST http://localhost:8080/api/v1/marketplace/servers/SERVER_ID/plugins/install \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source": "modrinth",
    "plugin_id": "worldedit",
    "version_id": "VERSION_ID"
  }'
```

### Backups

```bash
# Crear backup
curl -X POST http://localhost:8080/api/v1/servers/SERVER_ID/backups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"filename":"test-backup","backup_type":"full"}'

# Listar backups
curl http://localhost:8080/api/v1/servers/SERVER_ID/backups \
  -H "Authorization: Bearer $TOKEN"

# Configurar backups automáticos
curl -X PUT http://localhost:8080/api/v1/servers/SERVER_ID/backup-config \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "auto_backup": true,
    "schedule": "0 3 * * *",
    "max_backups": 7
  }'
```

---

**Última actualización:** 13 de noviembre de 2025  
**Versión:** 1.0
