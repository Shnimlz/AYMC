package agents

import (
	"context"
	"sync"
	"time"

	"github.com/aymc/backend/database/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// DefaultHealthCheckInterval es el intervalo por defecto para health checks (30 segundos)
	DefaultHealthCheckInterval = 30 * time.Second

	// MaxConsecutiveFailures es el número máximo de fallos consecutivos antes de marcar como offline
	MaxConsecutiveFailures = 3

	// HealthCheckTimeout es el timeout para cada health check individual
	HealthCheckTimeout = 5 * time.Second
)

// HealthMonitor monitorea el estado de salud de los agentes
type HealthMonitor struct {
	registry *AgentRegistry
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	logger   *zap.Logger
}

// NewHealthMonitor crea un nuevo monitor de salud
func NewHealthMonitor(registry *AgentRegistry, interval time.Duration, logger *zap.Logger) *HealthMonitor {
	if interval == 0 {
		interval = DefaultHealthCheckInterval
	}

	return &HealthMonitor{
		registry: registry,
		interval: interval,
		logger:   logger.With(zap.String("component", "health_monitor")),
	}
}

// Start inicia el monitoreo de salud de agentes
func (hm *HealthMonitor) Start() error {
	hm.ctx, hm.cancel = context.WithCancel(context.Background())

	hm.logger.Info("Starting health monitor",
		zap.Duration("interval", hm.interval),
		zap.Int("max_failures", MaxConsecutiveFailures),
	)

	hm.wg.Add(1)
	go hm.monitorLoop()

	return nil
}

// Stop detiene el monitoreo de salud
func (hm *HealthMonitor) Stop() {
	hm.logger.Info("Stopping health monitor")

	if hm.cancel != nil {
		hm.cancel()
	}

	hm.wg.Wait()

	hm.logger.Info("Health monitor stopped")
}

// monitorLoop es el bucle principal de monitoreo
func (hm *HealthMonitor) monitorLoop() {
	defer hm.wg.Done()

	ticker := time.NewTicker(hm.interval)
	defer ticker.Stop()

	// Realizar check inicial inmediatamente
	hm.checkAllAgents()

	for {
		select {
		case <-hm.ctx.Done():
			hm.logger.Info("Monitor loop stopped")
			return
		case <-ticker.C:
			hm.checkAllAgents()
		}
	}
}

// checkAllAgents verifica el estado de todos los agentes registrados
func (hm *HealthMonitor) checkAllAgents() {
	agents := hm.registry.ListAgents()

	if len(agents) == 0 {
		hm.logger.Debug("No agents to check")
		return
	}

	hm.logger.Debug("Checking agent health", zap.Int("agents", len(agents)))

	// Verificar cada agente en paralelo
	var wg sync.WaitGroup
	for _, agent := range agents {
		wg.Add(1)
		go func(conn *AgentConnection) {
			defer wg.Done()
			hm.checkAgent(conn)
		}(agent)
	}

	wg.Wait()

	// Log resumen
	online := hm.registry.CountOnline()
	total := hm.registry.Count()
	hm.logger.Debug("Health check complete",
		zap.Int("online", online),
		zap.Int("total", total),
	)
}

// checkAgent verifica el estado de un agente específico
func (hm *HealthMonitor) checkAgent(conn *AgentConnection) {
	ctx, cancel := context.WithTimeout(hm.ctx, HealthCheckTimeout)
	defer cancel()

	agentID := conn.ID

	// Realizar ping
	err := conn.Ping(ctx)
	if err != nil {
		hm.logger.Warn("Agent ping failed",
			zap.String("agent_id", agentID),
			zap.Int("consecutive_fails", conn.GetConsecutiveFails()),
			zap.Error(err),
		)

		// Si supera el máximo de fallos, marcar como offline
		if conn.GetConsecutiveFails() >= MaxConsecutiveFailures {
			hm.handleAgentFailure(conn)
		}

		return
	}

	// Ping exitoso, actualizar métricas
	if err := conn.UpdateMetrics(ctx); err != nil {
		hm.logger.Warn("Failed to update agent metrics",
			zap.String("agent_id", agentID),
			zap.Error(err),
		)
	}

	// Actualizar last_seen en base de datos
	agentUUID, err := parseUUID(agentID)
	if err != nil {
		hm.logger.Error("Invalid agent UUID", zap.String("agent_id", agentID), zap.Error(err))
		return
	}

	if err := hm.registry.UpdateAgentLastSeen(agentUUID); err != nil {
		hm.logger.Warn("Failed to update agent last_seen",
			zap.String("agent_id", agentID),
			zap.Error(err),
		)
	}

	// Si el agente estaba offline, marcarlo como online
	if conn.GetStatus() != AgentStatusOnline {
		hm.logger.Info("Agent recovered",
			zap.String("agent_id", agentID),
		)

		if err := hm.registry.UpdateAgentStatus(agentUUID, models.AgentStatusOnline); err != nil {
			hm.logger.Warn("Failed to update agent status to online",
				zap.String("agent_id", agentID),
				zap.Error(err),
			)
		}
	}
}

// handleAgentFailure maneja un agente que ha fallado múltiples veces
func (hm *HealthMonitor) handleAgentFailure(conn *AgentConnection) {
	agentID := conn.ID

	hm.logger.Error("Agent marked as offline after multiple failures",
		zap.String("agent_id", agentID),
		zap.Int("consecutive_fails", conn.GetConsecutiveFails()),
	)

	// Marcar conexión como offline
	conn.MarkAsOffline()

	// Actualizar estado en base de datos
	agentUUID, err := parseUUID(agentID)
	if err != nil {
		hm.logger.Error("Invalid agent UUID", zap.String("agent_id", agentID), zap.Error(err))
		return
	}

	if err := hm.registry.UpdateAgentStatus(agentUUID, models.AgentStatusOffline); err != nil {
		hm.logger.Error("Failed to update agent status to offline",
			zap.String("agent_id", agentID),
			zap.Error(err),
		)
	}

	// TODO: Implementar failover de servidores
	// 1. Obtener todos los servidores asignados a este agente
	// 2. Si hay servidores running, buscar agentes alternativos
	// 3. Migrar o detener los servidores afectados
	// 4. Notificar a los usuarios vía WebSocket

	hm.logger.Warn("Failover not implemented yet for agent",
		zap.String("agent_id", agentID),
	)
}

// CheckAgent ejecuta un health check manual en un agente específico
func (hm *HealthMonitor) CheckAgent(agentID string) error {
	agentUUID, err := parseUUID(agentID)
	if err != nil {
		return err
	}

	conn, err := hm.registry.GetAgent(agentUUID)
	if err != nil {
		return err
	}

	hm.checkAgent(conn)
	return nil
}

// GetStats retorna estadísticas del health monitor
func (hm *HealthMonitor) GetStats() map[string]interface{} {
	total := hm.registry.Count()
	online := hm.registry.CountOnline()

	return map[string]interface{}{
		"total_agents":    total,
		"online_agents":   online,
		"offline_agents":  total - online,
		"check_interval":  hm.interval.String(),
		"max_failures":    MaxConsecutiveFailures,
		"check_timeout":   HealthCheckTimeout.String(),
	}
}

// parseUUID es una función helper para parsear UUIDs
func parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}
