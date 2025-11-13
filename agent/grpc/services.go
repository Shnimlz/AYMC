package grpc

import (
	"archive/zip"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aymc/agent/core"
	pb "github.com/aymc/agent/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

// agentServiceImpl implementa el servicio AgentService
type agentServiceImpl struct {
	pb.UnimplementedAgentServiceServer
	agent *core.Agent
}

// GetAgentInfo retorna información del agente
func (s *agentServiceImpl) GetAgentInfo(ctx context.Context, req *pb.Empty) (*pb.AgentInfo, error) {
	log.Printf("[DEBUG] GetAgentInfo llamado")

	config := s.agent.GetConfig()
	servers := s.agent.ListServers()

	info := &pb.AgentInfo{
		AgentId:       config.AgentID,
		Version:       "0.1.0",
		Platform:      "linux", // TODO: Detectar dinámicamente
		PlatformVersion: "unknown",
		UptimeSeconds: int64(time.Since(s.agent.GetStartTime()).Seconds()),
		ActiveServers: int32(len(servers)),
		MaxServers:    int32(config.MaxServers),
	}

	return info, nil
}

// GetSystemMetrics retorna métricas del sistema
func (s *agentServiceImpl) GetSystemMetrics(ctx context.Context, req *pb.Empty) (*pb.SystemMetrics, error) {
	log.Printf("[DEBUG] GetSystemMetrics llamado")

	metrics := s.agent.GetMonitor().GetMetrics()

	openPorts, _ := s.agent.GetMonitor().GetOpenPorts()
	ports32 := make([]int32, len(openPorts))
	for i, p := range openPorts {
		ports32[i] = int32(p)
	}

	pbMetrics := &pb.SystemMetrics{
		Timestamp:     metrics.Timestamp.Unix(),
		CpuPercent:    metrics.CPUPercent,
		MemoryTotal:   metrics.MemoryTotal,
		MemoryUsed:    metrics.MemoryUsed,
		MemoryPercent: metrics.MemoryPercent,
		DiskTotal:     metrics.DiskTotal,
		DiskUsed:      metrics.DiskUsed,
		DiskPercent:   metrics.DiskPercent,
		NetworkSent:   metrics.NetworkSent,
		NetworkRecv:   metrics.NetworkRecv,
		OpenPorts:     ports32,
	}

	return pbMetrics, nil
}

// ListServers lista todos los servidores
func (s *agentServiceImpl) ListServers(ctx context.Context, req *pb.Empty) (*pb.ServerList, error) {
	log.Printf("[DEBUG] ListServers llamado")

	servers := s.agent.ListServers()
	pbServers := make([]*pb.ServerInfo, 0, len(servers))

	for _, srv := range servers {
		pbServers = append(pbServers, convertToProtoServer(srv))
	}

	return &pb.ServerList{Servers: pbServers}, nil
}

// GetServer obtiene información de un servidor específico
func (s *agentServiceImpl) GetServer(ctx context.Context, req *pb.ServerRequest) (*pb.ServerInfo, error) {
	log.Printf("[DEBUG] GetServer llamado: %s", req.ServerId)

	server, err := s.agent.GetServer(req.ServerId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "servidor no encontrado: %v", err)
	}

	return convertToProtoServer(server), nil
}

// StartServer inicia un servidor
func (s *agentServiceImpl) StartServer(ctx context.Context, req *pb.StartServerRequest) (*pb.ServerResponse, error) {
	log.Printf("[INFO] StartServer llamado: %s", req.ServerId)

	// Convertir configuración de proto a core
	config := core.ServerConfig{
		MinRAM:      req.Config.MinRam,
		MaxRAM:      req.Config.MaxRam,
		JavaArgs:    req.Config.JavaArgs,
		JarFile:     req.Config.JarFile,
		AutoRestart: req.Config.AutoRestart,
		CustomArgs:  req.Config.CustomArgs,
	}

	// Crear servidor en el agente
	server := &core.MinecraftServer{
		ID:      req.ServerId,
		Name:    req.Name,
		Type:    req.Type,
		Version: req.Version,
		Status:  core.StatusStarting,
		Config:  config,
	}

	// Iniciar servidor
	err := s.agent.StartServer(server)
	if err != nil {
		return &pb.ServerResponse{
			Success: false,
			Message: fmt.Sprintf("error iniciando servidor: %v", err),
		}, nil
	}

	return &pb.ServerResponse{
		Success: true,
		Message: "Servidor iniciado correctamente",
		Server:  convertToProtoServer(server),
	}, nil
}

// StopServer detiene un servidor
func (s *agentServiceImpl) StopServer(ctx context.Context, req *pb.ServerRequest) (*pb.ServerResponse, error) {
	log.Printf("[INFO] StopServer llamado: %s", req.ServerId)

	err := s.agent.StopServer(req.ServerId)
	if err != nil {
		return &pb.ServerResponse{
			Success: false,
			Message: fmt.Sprintf("error deteniendo servidor: %v", err),
		}, nil
	}

	return &pb.ServerResponse{
		Success: true,
		Message: "Servidor detenido correctamente",
	}, nil
}

// RestartServer reinicia un servidor
func (s *agentServiceImpl) RestartServer(ctx context.Context, req *pb.ServerRequest) (*pb.ServerResponse, error) {
	log.Printf("[INFO] RestartServer llamado: %s", req.ServerId)

	// Detener servidor
	if err := s.agent.StopServer(req.ServerId); err != nil {
		return &pb.ServerResponse{
			Success: false,
			Message: fmt.Sprintf("error deteniendo servidor: %v", err),
		}, nil
	}

	// Esperar un momento
	time.Sleep(2 * time.Second)

	// Obtener servidor y reiniciar
	server, err := s.agent.GetServer(req.ServerId)
	if err != nil {
		return &pb.ServerResponse{
			Success: false,
			Message: fmt.Sprintf("error obteniendo servidor: %v", err),
		}, nil
	}

	if err := s.agent.StartServer(server); err != nil {
		return &pb.ServerResponse{
			Success: false,
			Message: fmt.Sprintf("error iniciando servidor: %v", err),
		}, nil
	}

	return &pb.ServerResponse{
		Success: true,
		Message: "Servidor reiniciado correctamente",
	}, nil
}

// SendCommand envía un comando al servidor
func (s *agentServiceImpl) SendCommand(ctx context.Context, req *pb.CommandRequest) (*pb.CommandResponse, error) {
	log.Printf("[DEBUG] SendCommand llamado: %s -> %s", req.ServerId, req.Command)

	err := s.agent.GetExecutor().SendCommand(req.ServerId, req.Command)
	if err != nil {
		return &pb.CommandResponse{
			Success: false,
			Message: fmt.Sprintf("error enviando comando: %v", err),
		}, nil
	}

	return &pb.CommandResponse{
		Success: true,
		Message: "Comando enviado correctamente",
	}, nil
}

// StreamLogs transmite logs en tiempo real
func (s *agentServiceImpl) StreamLogs(req *pb.ServerRequest, stream pb.AgentService_StreamLogsServer) error {
	log.Printf("[INFO] StreamLogs iniciado para: %s", req.ServerId)

	logChan, err := s.agent.GetExecutor().GetLogs(req.ServerId)
	if err != nil {
		return status.Errorf(codes.NotFound, "servidor no encontrado: %v", err)
	}

	// Stream de logs
	for {
		select {
		case <-stream.Context().Done():
			log.Printf("[INFO] StreamLogs cerrado para: %s", req.ServerId)
			return nil
		case logLine, ok := <-logChan:
			if !ok {
				return nil
			}

			entry := &pb.LogEntry{
				Timestamp: time.Now().Unix(),
				ServerId:  req.ServerId,
				Level:     "INFO", // TODO: Parsear nivel
				Source:    "STDOUT",
				Message:   logLine,
			}

			if err := stream.Send(entry); err != nil {
				log.Printf("[ERROR] Error enviando log: %v", err)
				return err
			}
		}
	}
}

// ReadFile lee un archivo remoto
func (s *agentServiceImpl) ReadFile(ctx context.Context, req *pb.FileRequest) (*pb.FileContent, error) {
	log.Printf("[DEBUG] ReadFile llamado: %s", req.Path)

	// Validar ruta (seguridad básica)
	if !isValidPath(req.Path) {
		return nil, status.Errorf(codes.PermissionDenied, "ruta no válida")
	}

	// Leer archivo
	data, err := os.ReadFile(req.Path)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "error leyendo archivo: %v", err)
	}

	// Obtener info del archivo
	info, err := os.Stat(req.Path)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error obteniendo info: %v", err)
	}

	return &pb.FileContent{
		Content:      data,
		Size:         info.Size(),
		ModifiedTime: info.ModTime().Unix(),
	}, nil
}

// WriteFile escribe un archivo remoto
func (s *agentServiceImpl) WriteFile(ctx context.Context, req *pb.WriteFileRequest) (*pb.FileResponse, error) {
	log.Printf("[DEBUG] WriteFile llamado: %s", req.Path)

	// Validar ruta
	if !isValidPath(req.Path) {
		return &pb.FileResponse{
			Success: false,
			Message: "ruta no válida",
		}, nil
	}

	// Crear directorios si es necesario
	if req.CreateDirs {
		dir := filepath.Dir(req.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return &pb.FileResponse{
				Success: false,
				Message: fmt.Sprintf("error creando directorios: %v", err),
			}, nil
		}
	}

	// Escribir archivo
	if err := os.WriteFile(req.Path, req.Content, 0644); err != nil {
		return &pb.FileResponse{
			Success: false,
			Message: fmt.Sprintf("error escribiendo archivo: %v", err),
		}, nil
	}

	return &pb.FileResponse{
		Success: true,
		Message: "Archivo escrito correctamente",
	}, nil
}

// ListFiles lista archivos en un directorio
func (s *agentServiceImpl) ListFiles(ctx context.Context, req *pb.DirectoryRequest) (*pb.FileList, error) {
	log.Printf("[DEBUG] ListFiles llamado: %s", req.Path)

	// Validar ruta
	if !isValidPath(req.Path) {
		return nil, status.Errorf(codes.PermissionDenied, "ruta no válida")
	}

	var files []*pb.FileInfo

	if req.Recursive {
		// Listado recursivo
		err := filepath.Walk(req.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Continuar con siguientes archivos
			}

			files = append(files, &pb.FileInfo{
				Name:         info.Name(),
				Path:         path,
				Size:         info.Size(),
				IsDir:        info.IsDir(),
				ModifiedTime: info.ModTime().Unix(),
				Permissions:  int32(info.Mode().Perm()),
			})

			return nil
		})

		if err != nil {
			return nil, status.Errorf(codes.Internal, "error listando archivos: %v", err)
		}
	} else {
		// Listado simple
		entries, err := os.ReadDir(req.Path)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "error leyendo directorio: %v", err)
		}

		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			files = append(files, &pb.FileInfo{
				Name:         info.Name(),
				Path:         filepath.Join(req.Path, info.Name()),
				Size:         info.Size(),
				IsDir:        info.IsDir(),
				ModifiedTime: info.ModTime().Unix(),
				Permissions:  int32(info.Mode().Perm()),
			})
		}
	}

	return &pb.FileList{Files: files}, nil
}

// CheckDependencies verifica dependencias del sistema
func (s *agentServiceImpl) CheckDependencies(ctx context.Context, req *pb.Empty) (*pb.DependenciesStatus, error) {
	log.Printf("[DEBUG] CheckDependencies llamado")

	// TODO: Implementar verificación real
	return &pb.DependenciesStatus{
		JavaInstalled:   false,
		JavaVersion:     "unknown",
		JavaPaths:       []string{},
		ScreenInstalled: false,
		TmuxInstalled:   false,
		Environment:     make(map[string]string),
	}, nil
}

// InstallJava instala Java en el sistema
func (s *agentServiceImpl) InstallJava(ctx context.Context, req *pb.JavaInstallRequest) (*pb.InstallResponse, error) {
	log.Printf("[INFO] InstallJava llamado: version %s", req.Version)

	logs := []string{}
	
	// Crear instalador
	installer, err := core.NewJavaInstaller(req.Version)
	if err != nil {
		log.Printf("[ERROR] Error creando instalador: %v", err)
		return &pb.InstallResponse{
			Success: false,
			Message: fmt.Sprintf("Error detectando sistema: %v", err),
			Logs:    logs,
		}, nil
	}

	// Obtener información del sistema
	info := installer.GetInfo()
	logs = append(logs, fmt.Sprintf("Sistema operativo: %s", info["os"]))
	if info["distro"] != "" {
		logs = append(logs, fmt.Sprintf("Distribución: %s", info["distro"]))
	}
	logs = append(logs, fmt.Sprintf("Gestor de paquetes: %s", info["package_mgr"]))
	logs = append(logs, fmt.Sprintf("Paquete a instalar: %s", info["java_package"]))

	// Verificar si ya está instalado
	installed, version, _ := installer.CheckInstalled()
	if installed {
		logs = append(logs, fmt.Sprintf("Java ya está instalado: %s", version))
		return &pb.InstallResponse{
			Success: true,
			Message: fmt.Sprintf("Java ya está instalado: %s", version),
			Logs:    logs,
		}, nil
	}

	logs = append(logs, "Iniciando instalación de Java...")

	// Instalar Java
	err = installer.Install()
	if err != nil {
		log.Printf("[ERROR] Error instalando Java: %v", err)
		logs = append(logs, fmt.Sprintf("ERROR: %v", err))
		return &pb.InstallResponse{
			Success: false,
			Message: fmt.Sprintf("Error instalando Java: %v", err),
			Logs:    logs,
		}, nil
	}

	// Verificar instalación exitosa
	installed, installedVersion, err := installer.CheckInstalled()
	if !installed || err != nil {
		logs = append(logs, "Verificación post-instalación falló")
		return &pb.InstallResponse{
			Success: false,
			Message: "Java no se instaló correctamente",
			Logs:    logs,
		}, nil
	}

	logs = append(logs, fmt.Sprintf("✅ Java instalado exitosamente: %s", installedVersion))
	log.Printf("[INFO] Java instalado exitosamente: %s", installedVersion)

	return &pb.InstallResponse{
		Success: true,
		Message: fmt.Sprintf("Java %s instalado correctamente", installedVersion),
		Logs:    logs,
	}, nil
}

// DownloadServer descarga software del servidor
func (s *agentServiceImpl) DownloadServer(req *pb.DownloadRequest, stream pb.AgentService_DownloadServerServer) error {
	log.Printf("[INFO] DownloadServer llamado: %s v%s", req.ServerType, req.Version)

	// Directorio de salida (usar WorkDir del config)
	outputDir := filepath.Join(s.agent.GetConfig().WorkDir, "downloads")

	// Crear downloader
	downloader := core.NewServerDownloader(req.ServerType, req.Version, outputDir)

	// Callback para enviar progreso vía stream
	callback := func(progress core.DownloadProgress) {
		pbProgress := &pb.DownloadProgress{
			Downloaded: progress.Downloaded,
			Total:      progress.Total,
			Percent:    progress.Percent,
			Status:     progress.Message,
			Complete:   progress.Percent >= 100,
		}

		if err := stream.Send(pbProgress); err != nil {
			log.Printf("[ERROR] Error enviando progreso: %v", err)
		}
	}

	// Descargar con retry
	filePath, err := downloader.DownloadWithRetry(3, callback)
	if err != nil {
		log.Printf("[ERROR] Error descargando servidor: %v", err)
		// Enviar mensaje de error final
		stream.Send(&pb.DownloadProgress{
			Status:   fmt.Sprintf("❌ Error: %v", err),
			Complete: true,
		})
		return status.Errorf(codes.Internal, "error descargando servidor: %v", err)
	}

	// Enviar mensaje de éxito final
	stream.Send(&pb.DownloadProgress{
		Downloaded: 100,
		Total:      100,
		Percent:    100,
		Status:     fmt.Sprintf("✅ Descarga completada: %s", filePath),
		Complete:   true,
	})

	log.Printf("[INFO] Servidor descargado exitosamente: %s", filePath)
	return nil
}

// Ping responde al healthcheck
func (s *agentServiceImpl) Ping(ctx context.Context, req *pb.Empty) (*pb.PongResponse, error) {
	return &pb.PongResponse{
		Timestamp: time.Now().Unix(),
		Message:   "pong",
	}, nil
}

// HealthCheck verifica el estado del agente
func (s *agentServiceImpl) HealthCheck(ctx context.Context, req *pb.Empty) (*pb.HealthStatus, error) {
	checks := make(map[string]string)
	checks["agent"] = "ok"
	checks["executor"] = "ok"
	checks["monitor"] = "ok"

	return &pb.HealthStatus{
		Healthy:   true,
		Status:    "healthy",
		Checks:    checks,
		Timestamp: time.Now().Unix(),
	}, nil
}

// Funciones auxiliares

func convertToProtoServer(srv *core.MinecraftServer) *pb.ServerInfo {
	return &pb.ServerInfo{
		Id:          srv.ID,
		Name:        srv.Name,
		Type:        srv.Type,
		Version:     srv.Version,
		JavaVersion: srv.JavaVersion,
		Port:        int32(srv.Port),
		Status:      string(srv.Status),
		Pid:         int32(srv.PID),
		StartTime:   srv.StartTime.Unix(),
		WorkDir:     srv.WorkDir,
		Config: &pb.ServerConfig{
			MinRam:      srv.Config.MinRAM,
			MaxRam:      srv.Config.MaxRAM,
			JavaArgs:    srv.Config.JavaArgs,
			JarFile:     srv.Config.JarFile,
			AutoRestart: srv.Config.AutoRestart,
			CustomArgs:  srv.Config.CustomArgs,
		},
	}
}

func isValidPath(path string) bool {
	// Validación básica de seguridad
	// TODO: Mejorar con whitelist de directorios permitidos
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// No permitir acceso a directorios del sistema críticos
	forbidden := []string{"/etc/shadow", "/etc/passwd", "/root/.ssh"}
	for _, f := range forbidden {
		if abs == f || strings.HasPrefix(abs, f) {
			return false
		}
	}

	return true
}

// InstallPlugin instala un plugin en un servidor
func (s *agentServiceImpl) InstallPlugin(ctx context.Context, req *pb.InstallPluginRequest) (*pb.PluginResponse, error) {
	log.Printf("[INFO] InstallPlugin llamado: server=%s plugin=%s", req.ServerId, req.PluginName)

	// Obtener información del servidor
	server, err := s.agent.GetServer(req.ServerId)
	if err != nil {
		return &pb.PluginResponse{
			Success: false,
			Message: fmt.Sprintf("servidor no encontrado: %v", err),
		}, nil
	}

	// Construir ruta del directorio plugins
	pluginsDir := filepath.Join(server.WorkDir, "plugins")
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return &pb.PluginResponse{
			Success: false,
			Message: fmt.Sprintf("error creando directorio plugins: %v", err),
		}, nil
	}

	// Descargar plugin a archivo temporal
	tempFile := filepath.Join(os.TempDir(), req.FileName)
	log.Printf("[DEBUG] Descargando plugin de %s a %s", req.DownloadUrl, tempFile)

	if err := downloadFile(req.DownloadUrl, tempFile); err != nil {
		return &pb.PluginResponse{
			Success: false,
			Message: fmt.Sprintf("error descargando plugin: %v", err),
		}, nil
	}
	defer os.Remove(tempFile) // Limpiar archivo temporal

	// Validar que es un JAR válido
	if !isJarFile(tempFile) {
		return &pb.PluginResponse{
			Success: false,
			Message: "el archivo descargado no es un JAR válido",
		}, nil
	}

	// Copiar a directorio plugins
	destPath := filepath.Join(pluginsDir, req.FileName)
	if err := copyFile(tempFile, destPath); err != nil {
		return &pb.PluginResponse{
			Success: false,
			Message: fmt.Sprintf("error copiando plugin: %v", err),
		}, nil
	}

	// Leer metadata del plugin
	metadata, err := readPluginMetadata(destPath)
	if err != nil {
		log.Printf("[WARN] No se pudo leer metadata del plugin: %v", err)
		metadata = &pluginMetadata{
			Name:    req.PluginName,
			Version: req.Version,
		}
	}

	// Obtener tamaño del archivo
	fileSize, _ := getFileSize(destPath)

	// Reiniciar servidor si se solicita
	if req.AutoRestart && server.Status == core.StatusRunning {
		log.Printf("[INFO] Reiniciando servidor %s", req.ServerId)
		if err := s.agent.RestartServer(req.ServerId); err != nil {
			log.Printf("[WARN] Error reiniciando servidor: %v", err)
		}
	}

	log.Printf("[INFO] Plugin %s instalado exitosamente", req.PluginName)

	return &pb.PluginResponse{
		Success: true,
		Message: fmt.Sprintf("Plugin %s instalado correctamente", req.PluginName),
		Plugin: &pb.PluginInfo{
			Name:        metadata.Name,
			Version:     metadata.Version,
			Description: metadata.Description,
			Author:      metadata.Author,
			Enabled:     true,
			FileName:    req.FileName,
			FileSize:    fileSize,
			InstalledAt: time.Now().Unix(),
			Dependencies: metadata.Dependencies,
		},
	}, nil
}

// UninstallPlugin desinstala un plugin de un servidor
func (s *agentServiceImpl) UninstallPlugin(ctx context.Context, req *pb.UninstallPluginRequest) (*pb.PluginResponse, error) {
	log.Printf("[INFO] UninstallPlugin llamado: server=%s plugin=%s", req.ServerId, req.PluginName)

	// Obtener información del servidor
	server, err := s.agent.GetServer(req.ServerId)
	if err != nil {
		return &pb.PluginResponse{
			Success: false,
			Message: fmt.Sprintf("servidor no encontrado: %v", err),
		}, nil
	}

	pluginsDir := filepath.Join(server.WorkDir, "plugins")

	// Buscar el archivo JAR del plugin
	jarPath, err := findPluginJar(pluginsDir, req.PluginName)
	if err != nil {
		return &pb.PluginResponse{
			Success: false,
			Message: fmt.Sprintf("plugin no encontrado: %v", err),
		}, nil
	}

	// Eliminar archivo JAR
	if err := os.Remove(jarPath); err != nil {
		return &pb.PluginResponse{
			Success: false,
			Message: fmt.Sprintf("error eliminando plugin: %v", err),
		}, nil
	}

	log.Printf("[INFO] Archivo JAR eliminado: %s", jarPath)

	// Eliminar configuración si se solicita
	if req.DeleteConfig {
		configDir := filepath.Join(pluginsDir, req.PluginName)
		if _, err := os.Stat(configDir); err == nil {
			if err := os.RemoveAll(configDir); err != nil {
				log.Printf("[WARN] Error eliminando configuración: %v", err)
			} else {
				log.Printf("[INFO] Configuración eliminada: %s", configDir)
			}
		}
	}

	// Eliminar datos si se solicita
	if req.DeleteData {
		dataDir := filepath.Join(server.WorkDir, "world", "plugins", req.PluginName)
		if _, err := os.Stat(dataDir); err == nil {
			if err := os.RemoveAll(dataDir); err != nil {
				log.Printf("[WARN] Error eliminando datos: %v", err)
			} else {
				log.Printf("[INFO] Datos eliminados: %s", dataDir)
			}
		}
	}

	// Reiniciar servidor si se solicita
	if req.AutoRestart && server.Status == core.StatusRunning {
		log.Printf("[INFO] Reiniciando servidor %s", req.ServerId)
		if err := s.agent.RestartServer(req.ServerId); err != nil {
			log.Printf("[WARN] Error reiniciando servidor: %v", err)
		}
	}

	log.Printf("[INFO] Plugin %s desinstalado exitosamente", req.PluginName)

	return &pb.PluginResponse{
		Success: true,
		Message: fmt.Sprintf("Plugin %s desinstalado correctamente", req.PluginName),
	}, nil
}

// UpdatePlugin actualiza un plugin en un servidor
func (s *agentServiceImpl) UpdatePlugin(ctx context.Context, req *pb.UpdatePluginRequest) (*pb.PluginResponse, error) {
	log.Printf("[INFO] UpdatePlugin llamado: server=%s plugin=%s", req.ServerId, req.PluginName)

	// Obtener información del servidor
	server, err := s.agent.GetServer(req.ServerId)
	if err != nil {
		return &pb.PluginResponse{
			Success: false,
			Message: fmt.Sprintf("servidor no encontrado: %v", err),
		}, nil
	}

	pluginsDir := filepath.Join(server.WorkDir, "plugins")

	// Buscar el archivo JAR del plugin existente
	oldJarPath, err := findPluginJar(pluginsDir, req.PluginName)
	if err != nil {
		return &pb.PluginResponse{
			Success: false,
			Message: fmt.Sprintf("plugin no encontrado: %v", err),
		}, nil
	}

	// Crear backup del plugin antiguo
	backupPath := oldJarPath + ".backup"
	if err := copyFile(oldJarPath, backupPath); err != nil {
		return &pb.PluginResponse{
			Success: false,
			Message: fmt.Sprintf("error creando backup: %v", err),
		}, nil
	}

	log.Printf("[INFO] Backup creado: %s", backupPath)

	// Descargar nueva versión a archivo temporal
	tempFile := filepath.Join(os.TempDir(), req.FileName)
	log.Printf("[DEBUG] Descargando nueva versión de %s", req.DownloadUrl)

	if err := downloadFile(req.DownloadUrl, tempFile); err != nil {
		return &pb.PluginResponse{
			Success: false,
			Message: fmt.Sprintf("error descargando nueva versión: %v", err),
		}, nil
	}
	defer os.Remove(tempFile)

	// Validar JAR
	if !isJarFile(tempFile) {
		return &pb.PluginResponse{
			Success: false,
			Message: "el archivo descargado no es un JAR válido",
		}, nil
	}

	// Eliminar versión antigua
	if err := os.Remove(oldJarPath); err != nil {
		os.Remove(tempFile)
		return &pb.PluginResponse{
			Success: false,
			Message: fmt.Sprintf("error eliminando versión antigua: %v", err),
		}, nil
	}

	// Copiar nueva versión
	newJarPath := filepath.Join(pluginsDir, req.FileName)
	if err := copyFile(tempFile, newJarPath); err != nil {
		// Rollback: restaurar backup
		log.Printf("[ERROR] Error copiando nueva versión, restaurando backup...")
		if restoreErr := copyFile(backupPath, oldJarPath); restoreErr != nil {
			log.Printf("[CRITICAL] Error restaurando backup: %v", restoreErr)
		}
		return &pb.PluginResponse{
			Success: false,
			Message: fmt.Sprintf("error instalando nueva versión: %v", err),
		}, nil
	}

	// Eliminar backup si todo salió bien
	os.Remove(backupPath)

	// Leer metadata de la nueva versión
	metadata, err := readPluginMetadata(newJarPath)
	if err != nil {
		log.Printf("[WARN] No se pudo leer metadata: %v", err)
		metadata = &pluginMetadata{
			Name:    req.PluginName,
			Version: req.NewVersion,
		}
	}

	fileSize, _ := getFileSize(newJarPath)

	// Reiniciar servidor si se solicita
	if req.AutoRestart && server.Status == core.StatusRunning {
		log.Printf("[INFO] Reiniciando servidor %s", req.ServerId)
		if err := s.agent.RestartServer(req.ServerId); err != nil {
			log.Printf("[WARN] Error reiniciando servidor: %v", err)
		}
	}

	log.Printf("[INFO] Plugin %s actualizado exitosamente a versión %s", req.PluginName, req.NewVersion)

	return &pb.PluginResponse{
		Success: true,
		Message: fmt.Sprintf("Plugin %s actualizado a versión %s", req.PluginName, req.NewVersion),
		Plugin: &pb.PluginInfo{
			Name:        metadata.Name,
			Version:     metadata.Version,
			Description: metadata.Description,
			Author:      metadata.Author,
			Enabled:     true,
			FileName:    req.FileName,
			FileSize:    fileSize,
			InstalledAt: time.Now().Unix(),
			Dependencies: metadata.Dependencies,
		},
	}, nil
}

// ListPlugins lista todos los plugins instalados en un servidor
func (s *agentServiceImpl) ListPlugins(ctx context.Context, req *pb.ListPluginsRequest) (*pb.PluginList, error) {
	log.Printf("[INFO] ListPlugins llamado: server=%s", req.ServerId)

	// Obtener información del servidor
	server, err := s.agent.GetServer(req.ServerId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "servidor no encontrado: %v", err)
	}

	pluginsDir := filepath.Join(server.WorkDir, "plugins")

	// Verificar que el directorio existe
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		return &pb.PluginList{
			Plugins: []*pb.PluginInfo{},
			Total:   0,
		}, nil
	}

	// Listar archivos JAR
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error leyendo directorio plugins: %v", err)
	}

	var plugins []*pb.PluginInfo

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".jar") {
			continue
		}

		jarPath := filepath.Join(pluginsDir, entry.Name())

		// Leer metadata
		metadata, err := readPluginMetadata(jarPath)
		if err != nil {
			log.Printf("[WARN] Error leyendo metadata de %s: %v", entry.Name(), err)
			// Continuar con información básica
			metadata = &pluginMetadata{
				Name: strings.TrimSuffix(entry.Name(), ".jar"),
			}
		}

		info, _ := entry.Info()
		fileSize := info.Size()
		modTime := info.ModTime().Unix()

		plugins = append(plugins, &pb.PluginInfo{
			Name:        metadata.Name,
			Version:     metadata.Version,
			Description: metadata.Description,
			Author:      metadata.Author,
			Enabled:     !strings.HasSuffix(entry.Name(), ".disabled"),
			FileName:    entry.Name(),
			FileSize:    fileSize,
			InstalledAt: modTime,
			Dependencies: metadata.Dependencies,
		})
	}

	log.Printf("[INFO] Encontrados %d plugins en servidor %s", len(plugins), req.ServerId)

	return &pb.PluginList{
		Plugins: plugins,
		Total:   int32(len(plugins)),
	}, nil
}

// --- Helper functions for plugin management ---

// pluginMetadata contiene información del plugin.yml
type pluginMetadata struct {
	Name         string   `yaml:"name"`
	Version      string   `yaml:"version"`
	Main         string   `yaml:"main"`
	Description  string   `yaml:"description"`
	Author       string   `yaml:"author"`
	Authors      []string `yaml:"authors"`
	Depend       []string `yaml:"depend"`
	SoftDepend   []string `yaml:"softdepend"`
	Dependencies []string // Combinación de Depend y SoftDepend
}

// readPluginMetadata lee el plugin.yml de un archivo JAR
func readPluginMetadata(jarPath string) (*pluginMetadata, error) {
	reader, err := zip.OpenReader(jarPath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo JAR: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if file.Name == "plugin.yml" || file.Name == "paper-plugin.yml" || file.Name == "bungee.yml" {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("error abriendo %s: %w", file.Name, err)
			}
			defer rc.Close()

			var metadata pluginMetadata
			decoder := yaml.NewDecoder(rc)
			if err := decoder.Decode(&metadata); err != nil {
				return nil, fmt.Errorf("error parseando %s: %w", file.Name, err)
			}

			// Combinar dependencias
			metadata.Dependencies = append(metadata.Depend, metadata.SoftDepend...)

			// Si hay múltiples autores, usar el primero
			if len(metadata.Authors) > 0 && metadata.Author == "" {
				metadata.Author = metadata.Authors[0]
			}

			return &metadata, nil
		}
	}

	return nil, fmt.Errorf("plugin.yml no encontrado en JAR")
}

// downloadFile descarga un archivo desde una URL
func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error descargando: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creando archivo: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("error escribiendo archivo: %w", err)
	}

	return nil
}

// copyFile copia un archivo de src a dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error abriendo origen: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creando destino: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("error copiando: %w", err)
	}

	return destFile.Sync()
}

// isJarFile verifica si un archivo es un JAR válido
func isJarFile(filePath string) bool {
	if !strings.HasSuffix(strings.ToLower(filePath), ".jar") {
		return false
	}

	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return false
	}
	defer reader.Close()

	return true
}

// getFileSize retorna el tamaño de un archivo
func getFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// findPluginJar busca el archivo JAR de un plugin por nombre
func findPluginJar(pluginsDir, pluginName string) (string, error) {
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return "", fmt.Errorf("error leyendo directorio: %w", err)
	}

	pluginNameLower := strings.ToLower(pluginName)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		if !strings.HasSuffix(strings.ToLower(fileName), ".jar") {
			continue
		}

		jarPath := filepath.Join(pluginsDir, fileName)

		// Intentar leer metadata
		metadata, err := readPluginMetadata(jarPath)
		if err == nil && strings.ToLower(metadata.Name) == pluginNameLower {
			return jarPath, nil
		}

		// Fallback: comparar por nombre de archivo
		baseName := strings.TrimSuffix(strings.ToLower(fileName), ".jar")
		if strings.Contains(baseName, pluginNameLower) {
			return jarPath, nil
		}
	}

	return "", fmt.Errorf("plugin '%s' no encontrado", pluginName)
}

// validateSHA512 valida el hash SHA512 de un archivo
func validateSHA512(filePath string, expectedHash string) (bool, error) {
	if expectedHash == "" {
		return true, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("error abriendo archivo: %w", err)
	}
	defer file.Close()

	hasher := sha512.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return false, fmt.Errorf("error calculando hash: %w", err)
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	return strings.EqualFold(actualHash, expectedHash), nil
}


