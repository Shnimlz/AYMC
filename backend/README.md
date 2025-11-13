# AYMC Backend

Backend central del sistema AYMC (Advanced Yet Minecraft Controller) que coordina múltiples agentes remotos, gestiona servidores de Minecraft y provee APIs para el frontend.

## 🏗️ Arquitectura

```
┌─────────────┐      REST/WS      ┌──────────────┐      gRPC       ┌─────────────┐
│   Frontend  │ ◄──────────────── │   Backend    │ ◄────────────── │   Agentes   │
│   (Vue.js)  │                   │  (Go + Gin)  │                 │ (Go + gRPC) │
└─────────────┘                   └──────────────┘                 └─────────────┘
                                         │
                                         │
                                         ▼
                                  ┌──────────────┐
                                  │  PostgreSQL  │
                                  │  +  Redis    │
                                  └──────────────┘
```

## 🚀 Quick Start

### Requisitos Previos

- Go 1.23+
- Docker y Docker Compose
- Make

### Instalación

1. **Clonar el repositorio**
```bash
cd /path/to/AYMC/backend
```

2. **Configurar variables de entorno**
```bash
cp .env.example .env
# Editar .env con tus configuraciones
```

3. **Iniciar servicios con Docker**
```bash
make docker-up
```

4. **Ejecutar migraciones**
```bash
make migrate-up
```

5. **Insertar datos de prueba**
```bash
make seed
```

6. **Ejecutar el servidor**
```bash
make run
```

El servidor estará disponible en `http://localhost:8080`

## 📋 Comandos Disponibles

```bash
make help              # Mostrar todos los comandos disponibles
make run               # Ejecutar el servidor
make build             # Compilar el binario
make test              # Ejecutar tests
make test-coverage     # Tests con reporte de cobertura
make docker-up         # Iniciar Docker Compose
make docker-down       # Detener Docker Compose
make docker-logs       # Ver logs del contenedor
make migrate-up        # Aplicar migraciones
make migrate-down      # Revertir migraciones
make seed              # Insertar datos de prueba
make lint              # Ejecutar linters
make fmt               # Formatear código
make swagger           # Generar documentación Swagger
```

## 🗂️ Estructura del Proyecto

```
backend/
├── cmd/
│   └── server/          # Entry point de la aplicación
├── config/              # Configuración (Viper)
├── api/
│   ├── rest/            # Endpoints REST (Gin)
│   ├── websocket/       # WebSocket real-time
│   └── grpc/            # Cliente gRPC para agentes
├── services/            # Lógica de negocio
│   ├── auth/            # Autenticación JWT
│   ├── servers/         # Gestión de servidores
│   ├── agents/          # Pool de agentes
│   ├── marketplace/     # Integración con APIs externas
│   ├── backups/         # Sistema de backups
│   └── plugins/         # Gestión de plugins
├── database/
│   ├── models/          # Modelos GORM
│   ├── migrations/      # Migraciones SQL
│   └── seeders/         # Datos de prueba
├── pkg/
│   ├── logger/          # Logger (Zap)
│   └── utils/           # Utilidades
└── tests/
    ├── integration/     # Tests de integración
    └── e2e/             # Tests end-to-end
```

## 🔌 API Endpoints

### Autenticación
```
POST   /api/v1/auth/register        # Registro de usuario
POST   /api/v1/auth/login           # Login
POST   /api/v1/auth/refresh         # Refresh token
POST   /api/v1/auth/logout          # Logout
GET    /api/v1/auth/me              # Perfil actual
```

### Servidores
```
GET    /api/v1/servers              # Listar servidores
POST   /api/v1/servers              # Crear servidor
GET    /api/v1/servers/:id          # Ver servidor
PUT    /api/v1/servers/:id          # Actualizar servidor
DELETE /api/v1/servers/:id          # Eliminar servidor
POST   /api/v1/servers/:id/start    # Iniciar servidor
POST   /api/v1/servers/:id/stop     # Detener servidor
POST   /api/v1/servers/:id/restart  # Reiniciar servidor
```

### Agentes
```
GET    /api/v1/agents               # Listar agentes
GET    /api/v1/agents/:id           # Ver agente
POST   /api/v1/agents               # Registrar agente
DELETE /api/v1/agents/:id           # Desregistrar agente
GET    /api/v1/agents/:id/health    # Health check
```

### WebSocket
```
WS     /api/v1/ws?token=<jwt>       # Conexión WebSocket
```

Documentación completa en: `http://localhost:8080/swagger/index.html`

## 🗄️ Base de Datos

### Tablas Principales

- **users**: Usuarios del sistema
- **agents**: Agentes remotos conectados
- **servers**: Servidores de Minecraft
- **plugins**: Catálogo de plugins
- **server_plugins**: Relación many-to-many
- **backups**: Backups de servidores
- **server_metrics**: Métricas históricas

## 🔐 Autenticación

El sistema usa JWT (JSON Web Tokens) con:
- **Access Token**: 24 horas de validez
- **Refresh Token**: 7 días de validez
- **Roles**: admin, user, viewer

## 🧪 Testing

```bash
# Ejecutar todos los tests
make test

# Tests con cobertura
make test-coverage

# Benchmarks
make bench
```

## 🐳 Docker

### Servicios Incluidos

- **PostgreSQL 16**: Base de datos principal
- **Redis 7**: Cache y pub/sub
- **Adminer**: Administrador de DB web

### Accesos

- Backend: `http://localhost:8080`
- Adminer: `http://localhost:8081`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`

## 📊 Monitoreo

### Health Check

```bash
curl http://localhost:8080/health
```

### Métricas

```bash
# TODO: Prometheus endpoints
curl http://localhost:8080/metrics
```

## 🔧 Configuración

La configuración se carga en el siguiente orden (prioridad descendente):

1. Variables de entorno
2. Archivo `.env`
3. Archivo `config/config.yaml`
4. Valores por defecto

### Variables de Entorno Importantes

```bash
# Server
PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=aymc_db

# JWT
JWT_SECRET=your-secret-key

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
```

## 🛠️ Desarrollo

### Instalar herramientas de desarrollo

```bash
make install-tools
```

### Formatear código

```bash
make fmt
```

### Ejecutar linters

```bash
make lint
```

### Generar documentación Swagger

```bash
make swagger
```

## 📝 Roadmap

### Fase B.1 - Setup (✅ Completo)
- [x] Estructura de directorios
- [x] Configuración con Viper
- [x] Docker Compose
- [x] Logger con Zap

### Fase B.2 - Base de Datos (🚧 En progreso)
- [ ] Schema PostgreSQL
- [ ] Modelos GORM
- [ ] Migraciones
- [ ] Seeders

### Fase B.3 - Autenticación
- [ ] JWT Service
- [ ] Auth endpoints
- [ ] Middleware de auth
- [ ] RBAC

### Fase B.4 - Pool de Agentes
- [ ] Registry de agentes
- [ ] Health monitor
- [ ] Balanceador de carga
- [ ] Failover automático

### Fase B.5 - API REST
- [ ] Endpoints de servidores
- [ ] Endpoints de plugins
- [ ] Endpoints de backups
- [ ] Swagger docs

### Fase B.6 - WebSocket
- [ ] Hub de WebSocket
- [ ] Subscripciones
- [ ] Streaming de logs
- [ ] Métricas en tiempo real

### Fase B.7 - Marketplace
- [ ] Integración Spigot
- [ ] Integración Modrinth
- [ ] Integración CurseForge
- [ ] Cache con Redis

## 🤝 Contribución

Este proyecto es parte del sistema AYMC. Para contribuir:

1. Crear un branch desde `main`
2. Hacer los cambios
3. Ejecutar tests: `make test`
4. Formatear código: `make fmt`
5. Crear Pull Request

## 📄 Licencia

[Definir licencia]

## 🆘 Soporte

Para preguntas o problemas, crear un issue en el repositorio.

---

**Versión**: 0.1.0  
**Última actualización**: 13 de noviembre de 2024
