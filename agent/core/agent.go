package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Agent es el núcleo del agente AYMC
type Agent struct {
	ctx        context.Context
	config     *Config
	executor   *Executor
	monitor    *SystemMonitor
	servers    map[string]*MinecraftServer
	serversMux sync.RWMutex
	startTime  time.Time
}

// Config representa la configuración del agente
type Config struct {
	AgentID        string            `json:"agent_id"`
	BackendURL     string            `json:"backend_url"`
	Port           int               `json:"port"`
	LogLevel       string            `json:"log_level"`
	MaxServers     int               `json:"max_servers"`
	JavaPath       string            `json:"java_path"`
	WorkDir        string            `json:"work_dir"`
	EnableMetrics  bool              `json:"enable_metrics"`
	MetricsInterval time.Duration    `json:"metrics_interval"`
	CustomEnv      map[string]string `json:"custom_env"`
}

// MinecraftServer representa una instancia de servidor
type MinecraftServer struct {
	ID          string
	Name        string
	Type        string // paper, purpur, velocity, etc.
	Version     string
	JavaVersion string
	Port        int
	Status      ServerStatus
	PID         int
	StartTime   time.Time
	WorkDir     string
	Config      ServerConfig
}

// ServerStatus representa el estado del servidor
type ServerStatus string

const (
	StatusStopped  ServerStatus = "stopped"
	StatusStarting ServerStatus = "starting"
	StatusRunning  ServerStatus = "running"
	StatusStopping ServerStatus = "stopping"
	StatusCrashed  ServerStatus = "crashed"
)

// ServerConfig configuración específica del servidor
type ServerConfig struct {
	MinRAM      string            `json:"min_ram"`
	MaxRAM      string            `json:"max_ram"`
	JavaArgs    []string          `json:"java_args"`
	JarFile     string            `json:"jar_file"`
	AutoRestart bool              `json:"auto_restart"`
	CustomArgs  map[string]string `json:"custom_args"`
}

// NewAgent crea una nueva instancia del agente
func NewAgent(ctx context.Context, config *Config) (*Agent, error) {
	if config == nil {
		return nil, fmt.Errorf("configuración no puede ser nil")
	}

	// Validar y crear directorio de trabajo
	if err := os.MkdirAll(config.WorkDir, 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio de trabajo: %w", err)
	}

	// Inicializar executor
	executor, err := NewExecutor(config.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("error inicializando executor: %w", err)
	}

	// Inicializar monitor de sistema
	monitor := NewSystemMonitor()

	agent := &Agent{
		ctx:       ctx,
		config:    config,
		executor:  executor,
		monitor:   monitor,
		servers:   make(map[string]*MinecraftServer),
		startTime: time.Now(),
	}

	log.Printf("[INFO] Agente inicializado correctamente")
	log.Printf("[INFO] Directorio de trabajo: %s", config.WorkDir)
	log.Printf("[INFO] Max servidores: %d", config.MaxServers)

	return agent, nil
}

// LoadConfig carga la configuración desde un archivo
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// DefaultConfig retorna una configuración por defecto
func DefaultConfig() *Config {
	return &Config{
		AgentID:         generateAgentID(),
		BackendURL:      "localhost:50050",
		Port:            50051,
		LogLevel:        "info",
		MaxServers:      10,
		JavaPath:        "/usr/bin/java",
		WorkDir:         "/var/aymc/servers",
		EnableMetrics:   true,
		MetricsInterval: 5 * time.Second,
		CustomEnv:       make(map[string]string),
	}
}

// StartMonitoring inicia el monitoreo del sistema
func (a *Agent) StartMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("[INFO] Monitoreo de sistema iniciado (interval: %v)", interval)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[INFO] Monitoreo de sistema detenido")
			return
		case <-ticker.C:
			metrics := a.monitor.GetMetrics()
			a.processMetrics(metrics)
		}
	}
}

// processMetrics procesa y registra las métricas del sistema
func (a *Agent) processMetrics(metrics *SystemMetrics) {
	if !a.config.EnableMetrics {
		return
	}

	// Log de métricas solo en modo debug
	if a.config.LogLevel == "debug" {
		log.Printf("[DEBUG] CPU: %.2f%%, RAM: %.2f%%, Disk: %.2f%%",
			metrics.CPUPercent, metrics.MemoryPercent, metrics.DiskPercent)
	}

	// Aquí se enviarían las métricas al backend vía gRPC
}

// GetServer obtiene un servidor por ID
func (a *Agent) GetServer(id string) (*MinecraftServer, error) {
	a.serversMux.RLock()
	defer a.serversMux.RUnlock()

	server, exists := a.servers[id]
	if !exists {
		return nil, fmt.Errorf("servidor no encontrado: %s", id)
	}

	return server, nil
}

// ListServers lista todos los servidores
func (a *Agent) ListServers() []*MinecraftServer {
	a.serversMux.RLock()
	defer a.serversMux.RUnlock()

	servers := make([]*MinecraftServer, 0, len(a.servers))
	for _, server := range a.servers {
		servers = append(servers, server)
	}

	return servers
}

// Shutdown detiene el agente gracefully
func (a *Agent) Shutdown() {
	log.Printf("[INFO] Deteniendo agente...")

	// Detener todos los servidores
	a.serversMux.Lock()
	defer a.serversMux.Unlock()

	for id, server := range a.servers {
		if server.Status == StatusRunning {
			log.Printf("[INFO] Deteniendo servidor: %s", id)
			// TODO: implementar stopServer
		}
	}

	log.Printf("[INFO] Agente detenido")
}

// Métodos de acceso (getters)

// GetConfig retorna la configuración del agente
func (a *Agent) GetConfig() *Config {
	return a.config
}

// GetExecutor retorna el executor
func (a *Agent) GetExecutor() *Executor {
	return a.executor
}

// GetMonitor retorna el monitor
func (a *Agent) GetMonitor() *SystemMonitor {
	return a.monitor
}

// GetStartTime retorna el tiempo de inicio del agente
func (a *Agent) GetStartTime() time.Time {
	return a.startTime
}

// Métodos de gestión de servidores

// StartServer inicia un servidor de Minecraft
func (a *Agent) StartServer(server *MinecraftServer) error {
	a.serversMux.Lock()
	defer a.serversMux.Unlock()

	// Verificar si ya existe
	if _, exists := a.servers[server.ID]; exists {
		return fmt.Errorf("el servidor ya existe: %s", server.ID)
	}

	// Verificar límite de servidores
	if len(a.servers) >= a.config.MaxServers {
		return fmt.Errorf("límite de servidores alcanzado: %d", a.config.MaxServers)
	}

	// Iniciar el servidor
	if err := a.executor.StartServer(server.ID, server.Config); err != nil {
		return fmt.Errorf("error iniciando servidor: %w", err)
	}

	// Actualizar estado
	server.Status = StatusRunning
	server.StartTime = time.Now()
	a.servers[server.ID] = server

	log.Printf("[INFO] Servidor %s iniciado correctamente", server.ID)
	return nil
}

// StopServer detiene un servidor de Minecraft
func (a *Agent) StopServer(serverID string) error {
	a.serversMux.Lock()
	defer a.serversMux.Unlock()

	server, exists := a.servers[serverID]
	if !exists {
		return fmt.Errorf("servidor no encontrado: %s", serverID)
	}

	// Detener el servidor
	if err := a.executor.StopServer(serverID); err != nil {
		return fmt.Errorf("error deteniendo servidor: %w", err)
	}

	// Actualizar estado
	server.Status = StatusStopped
	delete(a.servers, serverID)

	log.Printf("[INFO] Servidor %s detenido correctamente", serverID)
	return nil
}

// RestartServer reinicia un servidor
func (a *Agent) RestartServer(serverID string) error {
	a.serversMux.Lock()
	server, exists := a.servers[serverID]
	if !exists {
		a.serversMux.Unlock()
		return fmt.Errorf("servidor no encontrado: %s", serverID)
	}

	// Guardar configuración antes de detener
	config := server.Config
	serverCopy := *server
	a.serversMux.Unlock()

	// Detener servidor
	if err := a.StopServer(serverID); err != nil {
		return fmt.Errorf("error deteniendo servidor: %w", err)
	}

	// Esperar un momento para asegurar cierre limpio
	time.Sleep(2 * time.Second)

	// Reiniciar con la misma configuración
	serverCopy.Status = StatusStarting
	serverCopy.Config = config

	if err := a.StartServer(&serverCopy); err != nil {
		return fmt.Errorf("error iniciando servidor: %w", err)
	}

	log.Printf("[INFO] Servidor %s reiniciado correctamente", serverID)
	return nil
}

// generateAgentID genera un ID único para el agente
func generateAgentID() string {
	// TODO: Implementar generación segura de ID
	return fmt.Sprintf("agent-%d", time.Now().Unix())
}
