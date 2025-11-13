package handlers

import (
	"net/http"

	"github.com/aymc/backend/services/agents"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AgentHandler handles agent-related HTTP requests
type AgentHandler struct {
	agentService *agents.AgentService
	logger       *zap.Logger
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(agentService *agents.AgentService, logger *zap.Logger) *AgentHandler {
	return &AgentHandler{
		agentService: agentService,
		logger:       logger.With(zap.String("handler", "agent")),
	}
}

// AgentResponse representa un agente en las respuestas
type AgentResponse struct {
	ID              string              `json:"id"`
	AgentID         string              `json:"agent_id"`
	Hostname        string              `json:"hostname"`
	IPAddress       string              `json:"ip_address"`
	Port            int                 `json:"port"`
	Status          string              `json:"status"`
	Version         string              `json:"version"`
	OS              string              `json:"os"`
	IsHealthy       bool                `json:"is_healthy"`
	LastSeen        string              `json:"last_seen,omitempty"`
	ConsecutiveFails int                `json:"consecutive_fails"`
	Metrics         *agents.AgentMetrics `json:"metrics,omitempty"`
}

// AgentListResponse representa una lista de agentes
type AgentListResponse struct {
	Agents      []AgentResponse `json:"agents"`
	Total       int             `json:"total"`
	Online      int             `json:"online"`
	Offline     int             `json:"offline"`
}

// ListAgents lista todos los agentes
// @Summary List all agents
// @Description Get a list of all registered agents with their status
// @Tags agents
// @Accept json
// @Produce json
// @Success 200 {object} AgentListResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/agents [get]
// @Security BearerAuth
func (h *AgentHandler) ListAgents(c *gin.Context) {
	registry := h.agentService.GetRegistry()
	connections := registry.ListAgents()

	agents := make([]AgentResponse, 0, len(connections))
	online := 0
	offline := 0

	for _, conn := range connections {
		agent := conn.Agent
		isHealthy := conn.IsHealthy()
		
		if isHealthy {
			online++
		} else {
			offline++
		}

		lastSeen := ""
		if agent.LastSeen != nil {
			lastSeen = agent.LastSeen.Format("2006-01-02T15:04:05Z07:00")
		}

		agentResp := AgentResponse{
			ID:               agent.ID.String(),
			AgentID:          agent.AgentID,
			Hostname:         agent.Hostname,
			IPAddress:        agent.IPAddress,
			Port:             agent.Port,
			Status:           string(conn.GetStatus()),
			Version:          agent.Version,
			OS:               agent.OS,
			IsHealthy:        isHealthy,
			LastSeen:         lastSeen,
			ConsecutiveFails: conn.GetConsecutiveFails(),
		}

		agents = append(agents, agentResp)
	}

	h.logger.Debug("Listed agents",
		zap.Int("total", len(agents)),
		zap.Int("online", online),
		zap.Int("offline", offline),
	)

	c.JSON(http.StatusOK, AgentListResponse{
		Agents:  agents,
		Total:   len(agents),
		Online:  online,
		Offline: offline,
	})
}

// GetAgent obtiene un agente específico
// @Summary Get agent details
// @Description Get detailed information about a specific agent
// @Tags agents
// @Accept json
// @Produce json
// @Param id path string true "Agent ID (UUID)"
// @Success 200 {object} AgentResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/agents/{id} [get]
// @Security BearerAuth
func (h *AgentHandler) GetAgent(c *gin.Context) {
	agentIDStr := c.Param("id")

	// Validar UUID
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		h.logger.Warn("Invalid agent ID", zap.String("id", agentIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}

	// Obtener conexión del registry
	registry := h.agentService.GetRegistry()
	conn, err := registry.GetAgent(agentID)
	if err != nil {
		h.logger.Warn("Agent not found", zap.String("id", agentIDStr), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	agent := conn.Agent
	lastSeen := ""
	if agent.LastSeen != nil {
		lastSeen = agent.LastSeen.Format("2006-01-02T15:04:05Z07:00")
	}

	agentResp := AgentResponse{
		ID:               agent.ID.String(),
		AgentID:          agent.AgentID,
		Hostname:         agent.Hostname,
		IPAddress:        agent.IPAddress,
		Port:             agent.Port,
		Status:           string(conn.GetStatus()),
		Version:          agent.Version,
		OS:               agent.OS,
		IsHealthy:        conn.IsHealthy(),
		LastSeen:         lastSeen,
		ConsecutiveFails: conn.GetConsecutiveFails(),
		Metrics:          conn.GetMetrics(),
	}

	c.JSON(http.StatusOK, agentResp)
}

// GetAgentHealth verifica el estado de salud de un agente
// @Summary Check agent health
// @Description Perform a health check on a specific agent
// @Tags agents
// @Accept json
// @Produce json
// @Param id path string true "Agent ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/agents/{id}/health [get]
// @Security BearerAuth
func (h *AgentHandler) GetAgentHealth(c *gin.Context) {
	agentIDStr := c.Param("id")

	// Validar UUID
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		h.logger.Warn("Invalid agent ID", zap.String("id", agentIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}

	// Obtener información del agente vía gRPC
	ctx := c.Request.Context()
	agentInfo, err := h.agentService.GetAgentInfo(ctx, agentID)
	if err != nil {
		h.logger.Error("Failed to get agent info",
			zap.String("agent_id", agentIDStr),
			zap.Error(err),
		)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to communicate with agent",
			"details": err.Error(),
		})
		return
	}

	// Obtener conexión para estado adicional
	registry := h.agentService.GetRegistry()
	conn, err := registry.GetAgent(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found in registry"})
		return
	}

	health := map[string]interface{}{
		"agent_id":        agentInfo.AgentId,
		"version":         agentInfo.Version,
		"platform":        agentInfo.Platform,
		"platform_version": agentInfo.PlatformVersion,
		"uptime_seconds":  agentInfo.UptimeSeconds,
		"active_servers":  agentInfo.ActiveServers,
		"max_servers":     agentInfo.MaxServers,
		"is_healthy":      conn.IsHealthy(),
		"status":          string(conn.GetStatus()),
		"last_seen":       conn.GetLastSeen().Format("2006-01-02T15:04:05Z07:00"),
		"consecutive_fails": conn.GetConsecutiveFails(),
	}

	h.logger.Debug("Agent health check completed", zap.String("agent_id", agentIDStr))

	c.JSON(http.StatusOK, health)
}

// GetAgentMetrics obtiene las métricas del sistema del agente
// @Summary Get agent metrics
// @Description Get system metrics (CPU, memory, disk) from a specific agent
// @Tags agents
// @Accept json
// @Produce json
// @Param id path string true "Agent ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/agents/{id}/metrics [get]
// @Security BearerAuth
func (h *AgentHandler) GetAgentMetrics(c *gin.Context) {
	agentIDStr := c.Param("id")

	// Validar UUID
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		h.logger.Warn("Invalid agent ID", zap.String("id", agentIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}

	// Obtener métricas del agente vía gRPC
	ctx := c.Request.Context()
	metrics, err := h.agentService.GetAgentMetrics(ctx, agentID)
	if err != nil {
		h.logger.Error("Failed to get agent metrics",
			zap.String("agent_id", agentIDStr),
			zap.Error(err),
		)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to get metrics from agent",
			"details": err.Error(),
		})
		return
	}

	metricsResp := map[string]interface{}{
		"timestamp":      metrics.Timestamp,
		"cpu_percent":    metrics.CpuPercent,
		"memory_total":   metrics.MemoryTotal,
		"memory_used":    metrics.MemoryUsed,
		"memory_percent": metrics.MemoryPercent,
		"disk_total":     metrics.DiskTotal,
		"disk_used":      metrics.DiskUsed,
		"disk_percent":   metrics.DiskPercent,
	}

	h.logger.Debug("Agent metrics retrieved", zap.String("agent_id", agentIDStr))

	c.JSON(http.StatusOK, metricsResp)
}

// GetAgentStats obtiene estadísticas del health monitor
// @Summary Get agent statistics
// @Description Get statistics about all agents and the health monitor
// @Tags agents
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/agents/stats [get]
// @Security BearerAuth
func (h *AgentHandler) GetAgentStats(c *gin.Context) {
	registry := h.agentService.GetRegistry()
	
	stats := map[string]interface{}{
		"total_agents":   registry.Count(),
		"online_agents":  registry.CountOnline(),
		"offline_agents": registry.Count() - registry.CountOnline(),
	}

	c.JSON(http.StatusOK, stats)
}
