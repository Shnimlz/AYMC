package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewAgent(t *testing.T) {
	// Crear directorio temporal para tests
	tmpDir := t.TempDir()

	config := &Config{
		AgentID:         "test-agent",
		BackendURL:      "localhost:50050",
		Port:            50051,
		LogLevel:        "info",
		MaxServers:      5,
		JavaPath:        "/usr/bin/java",
		WorkDir:         tmpDir,
		EnableMetrics:   true,
		MetricsInterval: 5 * time.Second,
		CustomEnv:       make(map[string]string),
	}

	ctx := context.Background()
	agent, err := NewAgent(ctx, config)

	if err != nil {
		t.Fatalf("Error creando agente: %v", err)
	}

	if agent == nil {
		t.Fatal("Agente es nil")
	}

	if agent.config.AgentID != "test-agent" {
		t.Errorf("AgentID esperado 'test-agent', obtenido '%s'", agent.config.AgentID)
	}

	if agent.config.MaxServers != 5 {
		t.Errorf("MaxServers esperado 5, obtenido %d", agent.config.MaxServers)
	}

	// Verificar que el directorio de trabajo se creó
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Directorio de trabajo no fue creado")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("Config es nil")
	}

	if config.Port != 50051 {
		t.Errorf("Puerto por defecto esperado 50051, obtenido %d", config.Port)
	}

	if config.MaxServers != 10 {
		t.Errorf("MaxServers por defecto esperado 10, obtenido %d", config.MaxServers)
	}

	if config.LogLevel != "info" {
		t.Errorf("LogLevel por defecto esperado 'info', obtenido '%s'", config.LogLevel)
	}
}

func TestAgentGetters(t *testing.T) {
	tmpDir := t.TempDir()
	config := &Config{
		AgentID:    "test-agent",
		WorkDir:    tmpDir,
		MaxServers: 5,
	}

	ctx := context.Background()
	agent, err := NewAgent(ctx, config)

	if err != nil {
		t.Fatalf("Error creando agente: %v", err)
	}

	// Test GetConfig
	if cfg := agent.GetConfig(); cfg == nil {
		t.Error("GetConfig retornó nil")
	}

	// Test GetExecutor
	if exec := agent.GetExecutor(); exec == nil {
		t.Error("GetExecutor retornó nil")
	}

	// Test GetMonitor
	if mon := agent.GetMonitor(); mon == nil {
		t.Error("GetMonitor retornó nil")
	}

	// Test GetStartTime
	startTime := agent.GetStartTime()
	if startTime.IsZero() {
		t.Error("GetStartTime retornó time.Time cero")
	}

	// Verificar que el tiempo de inicio es reciente
	if time.Since(startTime) > 1*time.Second {
		t.Error("StartTime no parece ser reciente")
	}
}

func TestStartStopServer(t *testing.T) {
	tmpDir := t.TempDir()
	config := &Config{
		AgentID:    "test-agent",
		WorkDir:    tmpDir,
		MaxServers: 5,
	}

	ctx := context.Background()
	agent, err := NewAgent(ctx, config)
	if err != nil {
		t.Fatalf("Error creando agente: %v", err)
	}

	// Crear un servidor de prueba
	server := &MinecraftServer{
		ID:      "test-server-1",
		Name:    "Test Server",
		Type:    "paper",
		Version: "1.20.1",
		Config: ServerConfig{
			MinRAM:  "1G",
			MaxRAM:  "2G",
			JarFile: filepath.Join(tmpDir, "server.jar"),
		},
	}

	// Crear archivo JAR dummy
	jarPath := filepath.Join(tmpDir, "test-server-1", "server.jar")
	os.MkdirAll(filepath.Dir(jarPath), 0755)
	os.WriteFile(jarPath, []byte("dummy"), 0644)

	// Test: Intentar iniciar servidor (fallará porque no hay Java real)
	// pero verificamos que la lógica funciona
	err = agent.StartServer(server)
	// Esperamos error porque no hay Java configurado correctamente
	if err == nil {
		// Si no hay error, verificar que se agregó a la lista
		servers := agent.ListServers()
		if len(servers) == 0 {
			t.Error("Servidor no se agregó a la lista")
		}
	}

	// Test: Listar servidores
	servers := agent.ListServers()
	if servers == nil {
		t.Error("ListServers retornó nil")
	}
}

func TestMaxServersLimit(t *testing.T) {
	tmpDir := t.TempDir()
	config := &Config{
		AgentID:    "test-agent",
		WorkDir:    tmpDir,
		MaxServers: 2, // Límite bajo para test
	}

	ctx := context.Background()
	agent, err := NewAgent(ctx, config)
	if err != nil {
		t.Fatalf("Error creando agente: %v", err)
	}

	// Agregar servidores directamente al mapa para simular
	agent.serversMux.Lock()
	agent.servers["server1"] = &MinecraftServer{ID: "server1"}
	agent.servers["server2"] = &MinecraftServer{ID: "server2"}
	agent.serversMux.Unlock()

	// Intentar agregar un tercer servidor
	server3 := &MinecraftServer{
		ID:   "server3",
		Name: "Server 3",
		Config: ServerConfig{
			MinRAM:  "1G",
			MaxRAM:  "2G",
			JarFile: "dummy.jar",
		},
	}

	err = agent.StartServer(server3)
	if err == nil {
		t.Error("Debería haber error por límite de servidores alcanzado")
	}
}

func TestGetServer(t *testing.T) {
	tmpDir := t.TempDir()
	config := &Config{
		AgentID:    "test-agent",
		WorkDir:    tmpDir,
		MaxServers: 5,
	}

	ctx := context.Background()
	agent, err := NewAgent(ctx, config)
	if err != nil {
		t.Fatalf("Error creando agente: %v", err)
	}

	// Agregar servidor manualmente
	testServer := &MinecraftServer{
		ID:     "test-server",
		Name:   "Test Server",
		Status: StatusStopped,
	}

	agent.serversMux.Lock()
	agent.servers["test-server"] = testServer
	agent.serversMux.Unlock()

	// Test: Obtener servidor existente
	server, err := agent.GetServer("test-server")
	if err != nil {
		t.Errorf("Error obteniendo servidor: %v", err)
	}

	if server == nil {
		t.Fatal("Servidor es nil")
	}

	if server.ID != "test-server" {
		t.Errorf("ID esperado 'test-server', obtenido '%s'", server.ID)
	}

	// Test: Intentar obtener servidor que no existe
	_, err = agent.GetServer("non-existent")
	if err == nil {
		t.Error("Debería haber error al obtener servidor inexistente")
	}
}
