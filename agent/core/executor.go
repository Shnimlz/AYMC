package core

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// Executor maneja la ejecución de procesos de servidores Minecraft
type Executor struct {
	workDir   string
	processes map[string]*Process
	mu        sync.RWMutex
}

// Process representa un proceso en ejecución
type Process struct {
	ID        string
	Cmd       *exec.Cmd
	Stdin     io.WriteCloser
	Stdout    io.ReadCloser
	Stderr    io.ReadCloser
	StartTime time.Time
	LogChan   chan string
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewExecutor crea un nuevo executor
func NewExecutor(workDir string) (*Executor, error) {
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio: %w", err)
	}

	return &Executor{
		workDir:   workDir,
		processes: make(map[string]*Process),
	}, nil
}

// StartServer inicia un servidor de Minecraft
func (e *Executor) StartServer(serverID string, config ServerConfig) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Verificar si ya está en ejecución
	if _, exists := e.processes[serverID]; exists {
		return fmt.Errorf("el servidor ya está en ejecución")
	}

	// Preparar directorio del servidor
	serverDir := filepath.Join(e.workDir, serverID)
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio del servidor: %w", err)
	}

	// Construir comando Java
	args := e.buildJavaCommand(config)
	
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "java", args...)
	cmd.Dir = serverDir

	// Configurar pipes para I/O
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("error creando stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("error creando stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("error creando stderr pipe: %w", err)
	}

	// Crear proceso
	process := &Process{
		ID:        serverID,
		Cmd:       cmd,
		Stdin:     stdin,
		Stdout:    stdout,
		Stderr:    stderr,
		StartTime: time.Now(),
		LogChan:   make(chan string, 1000),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Iniciar proceso
	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("error iniciando servidor: %w", err)
	}

	e.processes[serverID] = process

	// Iniciar captura de logs
	go e.captureLogs(process, stdout, "STDOUT")
	go e.captureLogs(process, stderr, "STDERR")

	// Monitorear proceso
	go e.monitorProcess(process)

	log.Printf("[INFO] Servidor %s iniciado con PID %d", serverID, cmd.Process.Pid)

	return nil
}

// StopServer detiene un servidor
func (e *Executor) StopServer(serverID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	process, exists := e.processes[serverID]
	if !exists {
		return fmt.Errorf("servidor no encontrado: %s", serverID)
	}

	log.Printf("[INFO] Deteniendo servidor %s...", serverID)

	// Intentar detener gracefully con comando "stop"
	if _, err := process.Stdin.Write([]byte("stop\n")); err != nil {
		log.Printf("[WARN] Error enviando comando stop: %v", err)
	}

	// Esperar 30 segundos para shutdown graceful
	gracefulTimer := time.NewTimer(30 * time.Second)
	defer gracefulTimer.Stop()

	done := make(chan struct{})
	go func() {
		process.Cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("[INFO] Servidor %s detenido gracefully", serverID)
	case <-gracefulTimer.C:
		log.Printf("[WARN] Servidor %s no respondió, forzando detención...", serverID)
		if err := process.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.Printf("[ERROR] Error enviando SIGTERM: %v", err)
			process.Cmd.Process.Kill()
		}
	}

	// Cancelar contexto y limpiar
	process.cancel()
	close(process.LogChan)
	delete(e.processes, serverID)

	return nil
}

// SendCommand envía un comando al servidor
func (e *Executor) SendCommand(serverID string, command string) error {
	e.mu.RLock()
	process, exists := e.processes[serverID]
	e.mu.RUnlock()

	if !exists {
		return fmt.Errorf("servidor no encontrado: %s", serverID)
	}

	if _, err := process.Stdin.Write([]byte(command + "\n")); err != nil {
		return fmt.Errorf("error enviando comando: %w", err)
	}

	log.Printf("[DEBUG] Comando enviado a %s: %s", serverID, command)
	return nil
}

// GetLogs retorna el canal de logs del servidor
func (e *Executor) GetLogs(serverID string) (<-chan string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	process, exists := e.processes[serverID]
	if !exists {
		return nil, fmt.Errorf("servidor no encontrado: %s", serverID)
	}

	return process.LogChan, nil
}

// buildJavaCommand construye los argumentos para el comando Java
func (e *Executor) buildJavaCommand(config ServerConfig) []string {
	args := []string{
		fmt.Sprintf("-Xms%s", config.MinRAM),
		fmt.Sprintf("-Xmx%s", config.MaxRAM),
	}

	// Agregar argumentos personalizados de Java
	args = append(args, config.JavaArgs...)

	// Agregar flags de optimización comunes
	args = append(args,
		"-XX:+UseG1GC",
		"-XX:+ParallelRefProcEnabled",
		"-XX:MaxGCPauseMillis=200",
		"-XX:+UnlockExperimentalVMOptions",
		"-XX:+DisableExplicitGC",
		"-XX:+AlwaysPreTouch",
		"-XX:G1NewSizePercent=30",
		"-XX:G1MaxNewSizePercent=40",
		"-XX:G1HeapRegionSize=8M",
		"-XX:G1ReservePercent=20",
		"-XX:G1HeapWastePercent=5",
		"-XX:G1MixedGCCountTarget=4",
		"-XX:InitiatingHeapOccupancyPercent=15",
		"-XX:G1MixedGCLiveThresholdPercent=90",
		"-XX:G1RSetUpdatingPauseTimePercent=5",
		"-XX:SurvivorRatio=32",
		"-XX:+PerfDisableSharedMem",
		"-XX:MaxTenuringThreshold=1",
	)

	// Agregar JAR y argumentos finales
	args = append(args, "-jar", config.JarFile, "nogui")

	return args
}

// captureLogs captura los logs del proceso
func (e *Executor) captureLogs(process *Process, reader io.Reader, source string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		logEntry := fmt.Sprintf("[%s] %s", source, line)
		
		select {
		case process.LogChan <- logEntry:
		default:
			// Canal lleno, descartar log antiguo
			<-process.LogChan
			process.LogChan <- logEntry
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[ERROR] Error leyendo logs de %s: %v", process.ID, err)
	}
}

// monitorProcess monitorea el estado del proceso
func (e *Executor) monitorProcess(process *Process) {
	err := process.Cmd.Wait()
	
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	log.Printf("[INFO] Proceso %s terminó con código %d", process.ID, exitCode)

	// Limpiar proceso de la lista
	e.mu.Lock()
	delete(e.processes, process.ID)
	e.mu.Unlock()

	// TODO: Implementar auto-restart si está configurado
}

// IsRunning verifica si un servidor está en ejecución
func (e *Executor) IsRunning(serverID string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, exists := e.processes[serverID]
	return exists
}

// GetPID retorna el PID del proceso
func (e *Executor) GetPID(serverID string) (int, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	process, exists := e.processes[serverID]
	if !exists {
		return 0, fmt.Errorf("servidor no encontrado")
	}

	if process.Cmd.Process == nil {
		return 0, fmt.Errorf("proceso no iniciado")
	}

	return process.Cmd.Process.Pid, nil
}
