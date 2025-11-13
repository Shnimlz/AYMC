package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/aymc/backend/database/models"
	pb "github.com/aymc/backend/proto"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AgentService proporciona operaciones para interactuar con agentes remotos
type AgentService struct {
	registry *AgentRegistry
	logger   *zap.Logger
}

// NewAgentService crea un nuevo servicio de agentes
func NewAgentService(registry *AgentRegistry, logger *zap.Logger) *AgentService {
	return &AgentService{
		registry: registry,
		logger:   logger.With(zap.String("service", "agent_service")),
	}
}

// ServerOperationRequest representa una solicitud de operación en un servidor
type ServerOperationRequest struct {
	ServerID   string
	ServerName string
	AgentID    uuid.UUID
	Config     *models.Server
}

// ServerOperationResponse representa la respuesta de una operación en un servidor
type ServerOperationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Status  string `json:"status,omitempty"`
}

// StartServer inicia un servidor en un agente
func (s *AgentService) StartServer(ctx context.Context, req *ServerOperationRequest) (*ServerOperationResponse, error) {
	s.logger.Info("Starting server on agent",
		zap.String("server_id", req.ServerID),
		zap.String("agent_id", req.AgentID.String()),
	)

	// Obtener conexión al agente
	conn, err := s.registry.GetAgent(req.AgentID)
	if err != nil {
		s.logger.Error("Agent not found in registry",
			zap.String("agent_id", req.AgentID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// Verificar que el agente esté saludable
	if !conn.IsHealthy() {
		s.logger.Warn("Agent is not healthy",
			zap.String("agent_id", req.AgentID.String()),
			zap.String("status", string(conn.GetStatus())),
		)
		return nil, fmt.Errorf("agent is not healthy")
	}

	// Preparar solicitud gRPC
	grpcReq := &pb.StartServerRequest{
		ServerId: req.ServerID,
		Config: &pb.ServerConfig{
			MinRam:  fmt.Sprintf("%dM", req.Config.MemoryMin),
			MaxRam:  fmt.Sprintf("%dM", req.Config.MemoryMax),
			JarFile: fmt.Sprintf("%s-%s.jar", req.Config.ServerType, req.Config.Version),
		},
	}

	// Timeout para la operación
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Llamar al agente vía gRPC
	resp, err := conn.Client.StartServer(ctx, grpcReq)
	if err != nil {
		s.logger.Error("Failed to start server on agent",
			zap.String("server_id", req.ServerID),
			zap.String("agent_id", req.AgentID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	s.logger.Info("Server started successfully",
		zap.String("server_id", req.ServerID),
		zap.Bool("success", resp.Success),
		zap.String("message", resp.Message),
	)

	return &ServerOperationResponse{
		Success: resp.Success,
		Message: resp.Message,
		Status:  resp.Server.Status,
	}, nil
}

// StopServer detiene un servidor en un agente
func (s *AgentService) StopServer(ctx context.Context, req *ServerOperationRequest) (*ServerOperationResponse, error) {
	s.logger.Info("Stopping server on agent",
		zap.String("server_id", req.ServerID),
		zap.String("agent_id", req.AgentID.String()),
	)

	// Obtener conexión al agente
	conn, err := s.registry.GetAgent(req.AgentID)
	if err != nil {
		s.logger.Error("Agent not found in registry",
			zap.String("agent_id", req.AgentID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// Verificar que el agente esté saludable
	if !conn.IsHealthy() {
		s.logger.Warn("Agent is not healthy",
			zap.String("agent_id", req.AgentID.String()),
			zap.String("status", string(conn.GetStatus())),
		)
		return nil, fmt.Errorf("agent is not healthy")
	}

	// Preparar solicitud gRPC
	grpcReq := &pb.ServerRequest{
		ServerId: req.ServerID,
	}

	// Timeout para la operación
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Llamar al agente vía gRPC
	resp, err := conn.Client.StopServer(ctx, grpcReq)
	if err != nil {
		s.logger.Error("Failed to stop server on agent",
			zap.String("server_id", req.ServerID),
			zap.String("agent_id", req.AgentID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to stop server: %w", err)
	}

	s.logger.Info("Server stopped successfully",
		zap.String("server_id", req.ServerID),
		zap.Bool("success", resp.Success),
		zap.String("message", resp.Message),
	)

	return &ServerOperationResponse{
		Success: resp.Success,
		Message: resp.Message,
		Status:  resp.Server.Status,
	}, nil
}

// RestartServer reinicia un servidor en un agente
func (s *AgentService) RestartServer(ctx context.Context, req *ServerOperationRequest) (*ServerOperationResponse, error) {
	s.logger.Info("Restarting server on agent",
		zap.String("server_id", req.ServerID),
		zap.String("agent_id", req.AgentID.String()),
	)

	// Obtener conexión al agente
	conn, err := s.registry.GetAgent(req.AgentID)
	if err != nil {
		s.logger.Error("Agent not found in registry",
			zap.String("agent_id", req.AgentID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// Verificar que el agente esté saludable
	if !conn.IsHealthy() {
		s.logger.Warn("Agent is not healthy",
			zap.String("agent_id", req.AgentID.String()),
			zap.String("status", string(conn.GetStatus())),
		)
		return nil, fmt.Errorf("agent is not healthy")
	}

	// Preparar solicitud gRPC
	grpcReq := &pb.ServerRequest{
		ServerId: req.ServerID,
	}

	// Timeout para la operación (restart puede tomar más tiempo)
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Llamar al agente vía gRPC
	resp, err := conn.Client.RestartServer(ctx, grpcReq)
	if err != nil {
		s.logger.Error("Failed to restart server on agent",
			zap.String("server_id", req.ServerID),
			zap.String("agent_id", req.AgentID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to restart server: %w", err)
	}

	s.logger.Info("Server restarted successfully",
		zap.String("server_id", req.ServerID),
		zap.Bool("success", resp.Success),
		zap.String("message", resp.Message),
	)

	return &ServerOperationResponse{
		Success: resp.Success,
		Message: resp.Message,
		Status:  resp.Server.Status,
	}, nil
}

// GetServerStatus obtiene el estado actual de un servidor desde el agente
func (s *AgentService) GetServerStatus(ctx context.Context, serverID string, agentID uuid.UUID) (*pb.ServerInfo, error) {
	s.logger.Debug("Getting server status from agent",
		zap.String("server_id", serverID),
		zap.String("agent_id", agentID.String()),
	)

	// Obtener conexión al agente
	conn, err := s.registry.GetAgent(agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// Preparar solicitud gRPC
	grpcReq := &pb.ServerRequest{
		ServerId: serverID,
	}

	// Timeout para la operación
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Llamar al agente vía gRPC
	resp, err := conn.Client.GetServer(ctx, grpcReq)
	if err != nil {
		s.logger.Error("Failed to get server status from agent",
			zap.String("server_id", serverID),
			zap.String("agent_id", agentID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get server status: %w", err)
	}

	return resp, nil
}

// SendCommand envía un comando a un servidor en un agente
func (s *AgentService) SendCommand(ctx context.Context, serverID string, agentID uuid.UUID, command string) (string, error) {
	s.logger.Info("Sending command to server",
		zap.String("server_id", serverID),
		zap.String("agent_id", agentID.String()),
		zap.String("command", command),
	)

	// Obtener conexión al agente
	conn, err := s.registry.GetAgent(agentID)
	if err != nil {
		return "", fmt.Errorf("agent not found: %w", err)
	}

	// Verificar que el agente esté saludable
	if !conn.IsHealthy() {
		return "", fmt.Errorf("agent is not healthy")
	}

	// Preparar solicitud gRPC
	grpcReq := &pb.CommandRequest{
		ServerId: serverID,
		Command:  command,
	}

	// Timeout para la operación
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Llamar al agente vía gRPC
	resp, err := conn.Client.SendCommand(ctx, grpcReq)
	if err != nil {
		s.logger.Error("Failed to send command to server",
			zap.String("server_id", serverID),
			zap.String("command", command),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	s.logger.Info("Command sent successfully",
		zap.String("server_id", serverID),
		zap.String("output", resp.Output),
	)

	return resp.Output, nil
}

// GetAgentInfo obtiene información del agente
func (s *AgentService) GetAgentInfo(ctx context.Context, agentID uuid.UUID) (*pb.AgentInfo, error) {
	// Obtener conexión al agente
	conn, err := s.registry.GetAgent(agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// Timeout para la operación
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Llamar al agente vía gRPC
	resp, err := conn.Client.GetAgentInfo(ctx, &pb.Empty{})
	if err != nil {
		s.logger.Error("Failed to get agent info",
			zap.String("agent_id", agentID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get agent info: %w", err)
	}

	return resp, nil
}

// GetAgentMetrics obtiene métricas del sistema del agente
func (s *AgentService) GetAgentMetrics(ctx context.Context, agentID uuid.UUID) (*pb.SystemMetrics, error) {
	// Obtener conexión al agente
	conn, err := s.registry.GetAgent(agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// Timeout para la operación
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Llamar al agente vía gRPC
	resp, err := conn.Client.GetSystemMetrics(ctx, &pb.Empty{})
	if err != nil {
		s.logger.Error("Failed to get agent metrics",
		zap.String("agent_id", agentID.String()),
		zap.Error(err),
	)
	return nil, fmt.Errorf("failed to get agent metrics: %w", err)
}

return resp, nil
}

// StreamLogsCallback es una función callback para manejar logs en streaming
type StreamLogsCallback func(serverID uuid.UUID, entry *pb.LogEntry)

// StreamLogs inicia el streaming de logs de un servidor desde un agente
func (s *AgentService) StreamLogs(ctx context.Context, serverID, agentID uuid.UUID, callback StreamLogsCallback) error {
	s.logger.Info("Starting log stream",
		zap.String("server_id", serverID.String()),
		zap.String("agent_id", agentID.String()),
	)

	// Obtener conexión al agente
	conn, err := s.registry.GetAgent(agentID)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	// Verificar que el agente esté saludable
	if !conn.IsHealthy() {
		return fmt.Errorf("agent is not healthy")
	}

	// Crear contexto con timeout
	streamCtx, cancel := context.WithTimeout(ctx, 24*time.Hour) // Stream de larga duración
	defer cancel()

	// Llamar al método StreamLogs del agente
	stream, err := conn.Client.StreamLogs(streamCtx, &pb.ServerRequest{
		ServerId: serverID.String(),
	})
	if err != nil {
		s.logger.Error("Failed to start log stream",
			zap.String("server_id", serverID.String()),
			zap.String("agent_id", agentID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to start log stream: %w", err)
	}

	s.logger.Info("Log stream established",
		zap.String("server_id", serverID.String()),
		zap.String("agent_id", agentID.String()),
	)

	// Leer logs del stream
	go func() {
		for {
			logEntry, err := stream.Recv()
			if err != nil {
				s.logger.Warn("Log stream closed",
					zap.String("server_id", serverID.String()),
					zap.Error(err),
				)
				return
			}

			// Llamar al callback con el log entry
			callback(serverID, logEntry)
		}
	}()

	return nil
}

// GetRegistry retorna el registry de agentes (para uso interno)
func (s *AgentService) GetRegistry() *AgentRegistry {
	return s.registry
}

// InstallPlugin instala un plugin en un servidor
func (s *AgentService) InstallPlugin(ctx context.Context, agentID uuid.UUID, serverID uuid.UUID, pluginName string, downloadURL string, fileName string) error {
	s.logger.Info("Installing plugin",
		zap.String("agent_id", agentID.String()),
		zap.String("server_id", serverID.String()),
		zap.String("plugin_name", pluginName),
		zap.String("download_url", downloadURL),
	)

	// Obtener conexión al agente
	agent, err := s.registry.GetAgent(agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	// Verificar salud del agente
	if !agent.IsHealthy() {
		return fmt.Errorf("agent is not healthy")
	}

	// Llamar al agente para instalar el plugin
	req := &pb.InstallPluginRequest{
		ServerId:    serverID.String(),
		PluginName:  pluginName,
		DownloadUrl: downloadURL,
		FileName:    fileName,
		AutoRestart: false, // No reiniciar automáticamente por seguridad
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute) // Dar más tiempo para descarga
	defer cancel()

	resp, err := agent.Client.InstallPlugin(timeoutCtx, req)
	if err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("plugin installation failed: %s", resp.Message)
	}

	s.logger.Info("Plugin installed successfully",
		zap.String("server_id", serverID.String()),
		zap.String("plugin_name", pluginName),
	)

	return nil
}

// UninstallPlugin desinstala un plugin de un servidor
func (s *AgentService) UninstallPlugin(ctx context.Context, agentID uuid.UUID, serverID uuid.UUID, pluginName string, deleteConfig bool, deleteData bool) error {
	s.logger.Info("Uninstalling plugin",
		zap.String("agent_id", agentID.String()),
		zap.String("server_id", serverID.String()),
		zap.String("plugin_name", pluginName),
		zap.Bool("delete_config", deleteConfig),
		zap.Bool("delete_data", deleteData),
	)

	// Obtener conexión al agente
	agent, err := s.registry.GetAgent(agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	// Verificar salud del agente
	if !agent.IsHealthy() {
		return fmt.Errorf("agent is not healthy")
	}

	// Llamar al agente para desinstalar el plugin
	req := &pb.UninstallPluginRequest{
		ServerId:     serverID.String(),
		PluginName:   pluginName,
		DeleteConfig: deleteConfig,
		DeleteData:   deleteData,
		AutoRestart:  false, // No reiniciar automáticamente por seguridad
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := agent.Client.UninstallPlugin(timeoutCtx, req)
	if err != nil {
		return fmt.Errorf("failed to uninstall plugin: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("plugin uninstallation failed: %s", resp.Message)
	}

	s.logger.Info("Plugin uninstalled successfully",
		zap.String("server_id", serverID.String()),
		zap.String("plugin_name", pluginName),
	)

	return nil
}

// UpdatePlugin actualiza un plugin en un servidor
func (s *AgentService) UpdatePlugin(ctx context.Context, agentID uuid.UUID, serverID uuid.UUID, pluginName string, downloadURL string, fileName string) error {
	s.logger.Info("Updating plugin",
		zap.String("agent_id", agentID.String()),
		zap.String("server_id", serverID.String()),
		zap.String("plugin_name", pluginName),
		zap.String("download_url", downloadURL),
	)

	// Obtener conexión al agente
	agent, err := s.registry.GetAgent(agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	// Verificar salud del agente
	if !agent.IsHealthy() {
		return fmt.Errorf("agent is not healthy")
	}

	// Llamar al agente para actualizar el plugin
	req := &pb.UpdatePluginRequest{
		ServerId:    serverID.String(),
		PluginName:  pluginName,
		DownloadUrl: downloadURL,
		FileName:    fileName,
		AutoRestart: false, // No reiniciar automáticamente por seguridad
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute) // Dar más tiempo para descarga
	defer cancel()

	resp, err := agent.Client.UpdatePlugin(timeoutCtx, req)
	if err != nil {
		return fmt.Errorf("failed to update plugin: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("plugin update failed: %s", resp.Message)
	}

	s.logger.Info("Plugin updated successfully",
		zap.String("server_id", serverID.String()),
		zap.String("plugin_name", pluginName),
	)

	return nil
}
