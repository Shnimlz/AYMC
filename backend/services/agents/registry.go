package agents

import (
	"context"
	"fmt"
	"sync"

	"github.com/aymc/backend/database"
	"github.com/aymc/backend/database/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AgentRegistry gestiona las conexiones activas a los agentes
type AgentRegistry struct {
	agents map[uuid.UUID]*AgentConnection
	mu     sync.RWMutex
	db     *gorm.DB
	logger *zap.Logger
}

// NewAgentRegistry crea un nuevo registro de agentes
func NewAgentRegistry(logger *zap.Logger) *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[uuid.UUID]*AgentConnection),
		db:     database.GetDB(),
		logger: logger.With(zap.String("component", "agent_registry")),
	}
}

// Register registra y conecta un agente
func (r *AgentRegistry) Register(ctx context.Context, agent *models.Agent) (*AgentConnection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verificar si el agente ya está registrado
	if existing, exists := r.agents[agent.ID]; exists {
		r.logger.Info("Agent already registered",
			zap.String("agent_id", agent.ID.String()),
			zap.String("status", string(existing.GetStatus())),
		)
		return existing, nil
	}

	r.logger.Info("Registering new agent",
		zap.String("agent_id", agent.ID.String()),
		zap.String("hostname", agent.Hostname),
		zap.String("address", agent.GetAddress()),
	)

	// Crear nueva conexión
	conn := NewAgentConnection(agent, r.logger)

	// Intentar conectar
	if err := conn.Connect(ctx); err != nil {
		r.logger.Error("Failed to connect to agent",
			zap.String("agent_id", agent.ID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to connect to agent: %w", err)
	}

	// Guardar en el mapa
	r.agents[agent.ID] = conn

	// Actualizar estado en base de datos
	if err := r.db.Model(agent).Updates(map[string]interface{}{
		"status": models.AgentStatusOnline,
	}).Error; err != nil {
		r.logger.Warn("Failed to update agent status in database",
			zap.String("agent_id", agent.ID.String()),
			zap.Error(err),
		)
	}

	r.logger.Info("Agent registered successfully",
		zap.String("agent_id", agent.ID.String()),
		zap.Int("total_agents", len(r.agents)),
	)

	return conn, nil
}

// Unregister desconecta y elimina un agente del registro
func (r *AgentRegistry) Unregister(agentID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found in registry")
	}

	r.logger.Info("Unregistering agent", zap.String("agent_id", agentID.String()))

	// Desconectar
	if err := conn.Disconnect(); err != nil {
		r.logger.Warn("Error disconnecting agent",
			zap.String("agent_id", agentID.String()),
			zap.Error(err),
		)
	}

	// Eliminar del mapa
	delete(r.agents, agentID)

	// Actualizar estado en base de datos
	if err := r.db.Model(&models.Agent{}).Where("id = ?", agentID).Updates(map[string]interface{}{
		"status": models.AgentStatusOffline,
	}).Error; err != nil {
		r.logger.Warn("Failed to update agent status in database",
			zap.String("agent_id", agentID.String()),
			zap.Error(err),
		)
	}

	r.logger.Info("Agent unregistered successfully",
		zap.String("agent_id", agentID.String()),
		zap.Int("remaining_agents", len(r.agents)),
	)

	return nil
}

// GetAgent obtiene la conexión de un agente específico
func (r *AgentRegistry) GetAgent(agentID uuid.UUID) (*AgentConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conn, exists := r.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found in registry")
	}

	return conn, nil
}

// ListAgents retorna todas las conexiones de agentes activas
func (r *AgentRegistry) ListAgents() []*AgentConnection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*AgentConnection, 0, len(r.agents))
	for _, conn := range r.agents {
		agents = append(agents, conn)
	}

	return agents
}

// GetOnlineAgents retorna solo los agentes que están online y saludables
func (r *AgentRegistry) GetOnlineAgents() []*AgentConnection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*AgentConnection, 0)
	for _, conn := range r.agents {
		if conn.IsHealthy() {
			agents = append(agents, conn)
		}
	}

	return agents
}

// Count retorna el número total de agentes registrados
func (r *AgentRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}

// CountOnline retorna el número de agentes online
func (r *AgentRegistry) CountOnline() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, conn := range r.agents {
		if conn.IsHealthy() {
			count++
		}
	}

	return count
}

// LoadAgentsFromDatabase carga todos los agentes marcados como online en la BD
// e intenta reconectarlos
func (r *AgentRegistry) LoadAgentsFromDatabase(ctx context.Context) error {
	r.logger.Info("Loading agents from database")

	var agents []models.Agent
	if err := r.db.Find(&agents).Error; err != nil {
		return fmt.Errorf("failed to load agents from database: %w", err)
	}

	r.logger.Info("Found agents in database", zap.Int("count", len(agents)))

	// Intentar conectar a cada agente
	for i := range agents {
		agent := &agents[i]

		// Solo intentar conectar si estaba marcado como online o si es la primera vez
		if agent.Status == models.AgentStatusOnline || agent.LastSeen == nil {
			r.logger.Info("Attempting to connect to agent",
				zap.String("agent_id", agent.ID.String()),
				zap.String("hostname", agent.Hostname),
			)

			if _, err := r.Register(ctx, agent); err != nil {
				r.logger.Warn("Failed to connect to agent",
					zap.String("agent_id", agent.ID.String()),
					zap.Error(err),
				)
				// No retornamos error, continuamos con el siguiente
			}
		}
	}

	r.logger.Info("Finished loading agents",
		zap.Int("registered", r.Count()),
		zap.Int("online", r.CountOnline()),
	)

	return nil
}

// Shutdown cierra todas las conexiones de agentes
func (r *AgentRegistry) Shutdown() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logger.Info("Shutting down agent registry", zap.Int("agents", len(r.agents)))

	for agentID, conn := range r.agents {
		r.logger.Info("Closing connection to agent", zap.String("agent_id", agentID.String()))
		if err := conn.Disconnect(); err != nil {
			r.logger.Warn("Error disconnecting agent",
				zap.String("agent_id", agentID.String()),
				zap.Error(err),
			)
		}
	}

	// Limpiar el mapa
	r.agents = make(map[uuid.UUID]*AgentConnection)

	r.logger.Info("Agent registry shutdown complete")
}

// UpdateAgentStatus actualiza el estado de un agente en la BD
func (r *AgentRegistry) UpdateAgentStatus(agentID uuid.UUID, status models.AgentStatus) error {
	err := r.db.Model(&models.Agent{}).Where("id = ?", agentID).Updates(map[string]interface{}{
		"status": status,
	}).Error

	if err != nil {
		r.logger.Error("Failed to update agent status",
			zap.String("agent_id", agentID.String()),
			zap.String("status", string(status)),
			zap.Error(err),
		)
	}

	return err
}

// UpdateAgentLastSeen actualiza el timestamp de last_seen de un agente
func (r *AgentRegistry) UpdateAgentLastSeen(agentID uuid.UUID) error {
	err := r.db.Exec("UPDATE agents SET last_seen = NOW() WHERE id = ?", agentID).Error

	if err != nil {
		r.logger.Error("Failed to update agent last_seen",
			zap.String("agent_id", agentID.String()),
			zap.Error(err),
		)
	}

	return err
}
