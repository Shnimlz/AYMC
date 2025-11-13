package agents

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aymc/backend/database/models"
	pb "github.com/aymc/backend/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AgentStatus representa el estado de conexión del agente
type AgentStatus string

const (
	AgentStatusConnecting AgentStatus = "connecting"
	AgentStatusOnline     AgentStatus = "online"
	AgentStatusOffline    AgentStatus = "offline"
	AgentStatusError      AgentStatus = "error"
)

// AgentMetrics almacena métricas del agente
type AgentMetrics struct {
	CPUPercent    float64   `json:"cpu_percent"`
	MemoryTotal   uint64    `json:"memory_total"`
	MemoryUsed    uint64    `json:"memory_used"`
	MemoryPercent float64   `json:"memory_percent"`
	DiskTotal     uint64    `json:"disk_total"`
	DiskUsed      uint64    `json:"disk_used"`
	DiskPercent   float64   `json:"disk_percent"`
	ActiveServers int32     `json:"active_servers"`
	MaxServers    int32     `json:"max_servers"`
	Uptime        int64     `json:"uptime_seconds"`
	LastUpdated   time.Time `json:"last_updated"`
}

// AgentConnection representa una conexión activa con un agente
type AgentConnection struct {
	ID              string
	Agent           *models.Agent
	Client          pb.AgentServiceClient
	conn            *grpc.ClientConn
	lastSeen        time.Time
	status          AgentStatus
	metrics         *AgentMetrics
	consecutiveFails int
	mu              sync.RWMutex
	logger          *zap.Logger
}

// NewAgentConnection crea una nueva conexión al agente
func NewAgentConnection(agent *models.Agent, logger *zap.Logger) *AgentConnection {
	return &AgentConnection{
		ID:              agent.ID.String(),
		Agent:           agent,
		status:          AgentStatusOffline,
		consecutiveFails: 0,
		metrics:         &AgentMetrics{},
		logger:          logger.With(zap.String("agent_id", agent.ID.String())),
	}
}

// Connect establece la conexión gRPC con el agente
func (ac *AgentConnection) Connect(ctx context.Context) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.conn != nil {
		ac.logger.Warn("Agent already connected, closing existing connection")
		ac.conn.Close()
	}

	ac.status = AgentStatusConnecting
	ac.logger.Info("Connecting to agent", zap.String("address", ac.Agent.GetAddress()))

	// Crear conexión gRPC con timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// TODO: Agregar TLS en producción
	conn, err := grpc.DialContext(
		ctx,
		ac.Agent.GetAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		ac.status = AgentStatusError
		ac.logger.Error("Failed to connect to agent", zap.Error(err))
		return fmt.Errorf("failed to connect to agent: %w", err)
	}

	ac.conn = conn
	ac.Client = pb.NewAgentServiceClient(conn)
	ac.status = AgentStatusOnline
	ac.lastSeen = time.Now()
	ac.consecutiveFails = 0

	ac.logger.Info("Successfully connected to agent")

	// Obtener información inicial del agente
	if err := ac.updateAgentInfo(ctx); err != nil {
		ac.logger.Warn("Failed to get initial agent info", zap.Error(err))
		// No retornamos error, la conexión está establecida
	}

	return nil
}

// Disconnect cierra la conexión con el agente
func (ac *AgentConnection) Disconnect() error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.conn == nil {
		return nil
	}

	ac.logger.Info("Disconnecting from agent")

	err := ac.conn.Close()
	ac.conn = nil
	ac.Client = nil
	ac.status = AgentStatusOffline

	if err != nil {
		ac.logger.Error("Error closing connection", zap.Error(err))
		return fmt.Errorf("error closing connection: %w", err)
	}

	return nil
}

// Ping envía un ping al agente para verificar conectividad
func (ac *AgentConnection) Ping(ctx context.Context) error {
	ac.mu.RLock()
	if ac.Client == nil {
		ac.mu.RUnlock()
		return fmt.Errorf("agent not connected")
	}
	client := ac.Client
	ac.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	start := time.Now()
	resp, err := client.Ping(ctx, &pb.Empty{})
	latency := time.Since(start)

	if err != nil {
		ac.mu.Lock()
		ac.consecutiveFails++
		ac.mu.Unlock()
		
		ac.logger.Warn("Ping failed",
			zap.Error(err),
			zap.Int("consecutive_fails", ac.consecutiveFails),
		)
		return fmt.Errorf("ping failed: %w", err)
	}

	ac.mu.Lock()
	ac.lastSeen = time.Now()
	ac.consecutiveFails = 0
	ac.status = AgentStatusOnline
	ac.mu.Unlock()

	ac.logger.Debug("Ping successful",
		zap.Duration("latency", latency),
		zap.String("message", resp.Message),
	)

	return nil
}

// IsHealthy verifica si el agente está saludable
func (ac *AgentConnection) IsHealthy() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	// Consideramos saludable si:
	// 1. Status es online
	// 2. No hay más de 2 fallos consecutivos
	// 3. Se vio en los últimos 2 minutos
	return ac.status == AgentStatusOnline &&
		ac.consecutiveFails < 3 &&
		time.Since(ac.lastSeen) < 2*time.Minute
}

// GetStatus retorna el estado actual del agente
func (ac *AgentConnection) GetStatus() AgentStatus {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.status
}

// GetMetrics retorna las métricas actuales del agente
func (ac *AgentConnection) GetMetrics() *AgentMetrics {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	// Retornar copia para evitar race conditions
	metrics := *ac.metrics
	return &metrics
}

// GetLastSeen retorna la última vez que se vio el agente
func (ac *AgentConnection) GetLastSeen() time.Time {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.lastSeen
}

// GetConsecutiveFails retorna el número de fallos consecutivos
func (ac *AgentConnection) GetConsecutiveFails() int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.consecutiveFails
}

// UpdateMetrics obtiene y actualiza las métricas del agente
func (ac *AgentConnection) UpdateMetrics(ctx context.Context) error {
	ac.mu.RLock()
	if ac.Client == nil {
		ac.mu.RUnlock()
		return fmt.Errorf("agent not connected")
	}
	client := ac.Client
	ac.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	metrics, err := client.GetSystemMetrics(ctx, &pb.Empty{})
	if err != nil {
		ac.logger.Warn("Failed to get system metrics", zap.Error(err))
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	ac.mu.Lock()
	ac.metrics.CPUPercent = metrics.CpuPercent
	ac.metrics.MemoryTotal = metrics.MemoryTotal
	ac.metrics.MemoryUsed = metrics.MemoryUsed
	ac.metrics.MemoryPercent = metrics.MemoryPercent
	ac.metrics.DiskTotal = metrics.DiskTotal
	ac.metrics.DiskUsed = metrics.DiskUsed
	ac.metrics.DiskPercent = metrics.DiskPercent
	ac.metrics.LastUpdated = time.Now()
	ac.mu.Unlock()

	return nil
}

// updateAgentInfo obtiene información inicial del agente
func (ac *AgentConnection) updateAgentInfo(ctx context.Context) error {
	if ac.Client == nil {
		return fmt.Errorf("agent not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	info, err := ac.Client.GetAgentInfo(ctx, &pb.Empty{})
	if err != nil {
		return fmt.Errorf("failed to get agent info: %w", err)
	}

	ac.mu.Lock()
	ac.metrics.ActiveServers = info.ActiveServers
	ac.metrics.MaxServers = info.MaxServers
	ac.metrics.Uptime = info.UptimeSeconds
	ac.mu.Unlock()

	ac.logger.Info("Updated agent info",
		zap.Int32("active_servers", info.ActiveServers),
		zap.Int32("max_servers", info.MaxServers),
		zap.String("version", info.Version),
		zap.String("platform", info.Platform),
	)

	return nil
}

// MarkAsOffline marca el agente como offline
func (ac *AgentConnection) MarkAsOffline() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.status = AgentStatusOffline
	ac.logger.Warn("Agent marked as offline",
		zap.Int("consecutive_fails", ac.consecutiveFails),
	)
}

// ResetFailCount resetea el contador de fallos
func (ac *AgentConnection) ResetFailCount() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.consecutiveFails = 0
}
