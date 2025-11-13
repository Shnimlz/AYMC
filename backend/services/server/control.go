package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aymc/backend/database"
	"github.com/aymc/backend/database/models"
	"github.com/aymc/backend/services/agents"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Start starts a server
func (s *ServerService) Start(serverID, userID uuid.UUID, isAdmin bool) (*ServerResponse, error) {
	db := database.GetDB()

	// Get server
	var server models.Server
	query := db.Preload("Agent").Preload("User")
	
	if !isAdmin {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&server, "id = ?", serverID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrServerNotFound
		}
		return nil, fmt.Errorf("failed to query server: %w", err)
	}

	// Verify agent is online
	if !server.Agent.IsOnline() {
		s.logger.Warn("Attempt to start server on offline agent",
			zap.String("server_id", serverID.String()),
			zap.String("agent_id", server.AgentID.String()),
		)
		return nil, ErrAgentOffline
	}

	// Check if server can be started
	if !server.CanStart() {
		s.logger.Warn("Server cannot be started",
			zap.String("server_id", serverID.String()),
			zap.String("current_status", string(server.Status)),
		)
		return nil, ErrInvalidServerState
	}

	// Update server status to starting
	now := time.Now()
	updates := map[string]interface{}{
		"status":       models.ServerStatusStarting,
		"last_started": now,
	}

	if err := db.Model(&server).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update server status", zap.Error(err))
		return nil, fmt.Errorf("failed to update server status: %w", err)
	}

	s.logger.Info("Server start initiated",
		zap.String("server_id", serverID.String()),
		zap.String("name", server.Name),
	)

	// Send start command to agent via gRPC
	ctx := context.Background()
	req := &agents.ServerOperationRequest{
		ServerID:   serverID.String(),
		ServerName: server.Name,
		AgentID:    server.AgentID,
		Config:     &server,
	}

	resp, err := s.agentService.StartServer(ctx, req)
	if err != nil {
		// Si falla el gRPC, revertir el estado
		db.Model(&server).Update("status", models.ServerStatusStopped)
		s.logger.Error("Failed to start server on agent",
			zap.String("server_id", serverID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to start server on agent: %w", err)
	}

	// Actualizar estado según respuesta del agente
	if resp.Success {
		db.Model(&server).Update("status", models.ServerStatusRunning)
		s.logger.Info("Server started successfully on agent",
			zap.String("server_id", serverID.String()),
			zap.String("message", resp.Message),
		)
	} else {
		db.Model(&server).Update("status", models.ServerStatusError)
		s.logger.Warn("Server start failed on agent",
			zap.String("server_id", serverID.String()),
			zap.String("message", resp.Message),
		)
	}

	// Reload server
	if err := db.Preload("Agent").Preload("User").First(&server, "id = ?", serverID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload server: %w", err)
	}

	return s.toServerResponse(&server, &server.Agent, &server.User), nil
}

// Stop stops a server
func (s *ServerService) Stop(serverID, userID uuid.UUID, isAdmin bool) (*ServerResponse, error) {
	db := database.GetDB()

	// Get server
	var server models.Server
	query := db.Preload("Agent").Preload("User")
	
	if !isAdmin {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&server, "id = ?", serverID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrServerNotFound
		}
		return nil, fmt.Errorf("failed to query server: %w", err)
	}

	// Check if server can be stopped
	if !server.CanStop() {
		s.logger.Warn("Server cannot be stopped",
			zap.String("server_id", serverID.String()),
			zap.String("current_status", string(server.Status)),
		)
		return nil, ErrInvalidServerState
	}

	// Update server status to stopping
	now := time.Now()
	updates := map[string]interface{}{
		"status":       models.ServerStatusStopping,
		"last_stopped": now,
	}

	if err := db.Model(&server).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update server status", zap.Error(err))
		return nil, fmt.Errorf("failed to update server status: %w", err)
	}

	s.logger.Info("Server stop initiated",
		zap.String("server_id", serverID.String()),
		zap.String("name", server.Name),
	)

	// Send stop command to agent via gRPC
	ctx := context.Background()
	req := &agents.ServerOperationRequest{
		ServerID:   serverID.String(),
		ServerName: server.Name,
		AgentID:    server.AgentID,
		Config:     &server,
	}

	resp, err := s.agentService.StopServer(ctx, req)
	if err != nil {
		// Si falla el gRPC, marcar como error
		db.Model(&server).Update("status", models.ServerStatusError)
		s.logger.Error("Failed to stop server on agent",
			zap.String("server_id", serverID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to stop server on agent: %w", err)
	}

	// Actualizar estado según respuesta del agente
	if resp.Success {
		db.Model(&server).Update("status", models.ServerStatusStopped)
		s.logger.Info("Server stopped successfully on agent",
			zap.String("server_id", serverID.String()),
			zap.String("message", resp.Message),
		)
	} else {
		db.Model(&server).Update("status", models.ServerStatusError)
		s.logger.Warn("Server stop failed on agent",
			zap.String("server_id", serverID.String()),
			zap.String("message", resp.Message),
		)
	}

	// Reload server
	if err := db.Preload("Agent").Preload("User").First(&server, "id = ?", serverID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload server: %w", err)
	}

	return s.toServerResponse(&server, &server.Agent, &server.User), nil
}

// Restart restarts a server
func (s *ServerService) Restart(serverID, userID uuid.UUID, isAdmin bool) (*ServerResponse, error) {
	// Stop the server first
	_, err := s.Stop(serverID, userID, isAdmin)
	if err != nil && err != ErrInvalidServerState {
		return nil, err
	}

	// If server was already stopped, just start it
	if err == ErrInvalidServerState {
		return s.Start(serverID, userID, isAdmin)
	}

	// TODO: Wait for server to stop before starting
	// For now, we'll immediately start it
	// In production, this should wait for the stop confirmation from the agent

	s.logger.Info("Server restart initiated",
		zap.String("server_id", serverID.String()),
	)

	// Start the server
	return s.Start(serverID, userID, isAdmin)
}

// GetStatus retrieves the current status of a server
func (s *ServerService) GetStatus(serverID, userID uuid.UUID, isAdmin bool) (*ServerStatusResponse, error) {
	db := database.GetDB()

	// Get server
	var server models.Server
	query := db.Preload("Agent")
	
	if !isAdmin {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&server, "id = ?", serverID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrServerNotFound
		}
		return nil, fmt.Errorf("failed to query server: %w", err)
	}

	// TODO: Query actual status from agent via gRPC
	// For now, return the database status

	return &ServerStatusResponse{
		ServerID:    server.ID,
		Status:      server.Status,
		IsRunning:   server.IsRunning(),
		LastStarted: server.LastStarted,
		LastStopped: server.LastStopped,
		AgentOnline: server.Agent.IsOnline(),
	}, nil
}

// ServerStatusResponse represents server status information
type ServerStatusResponse struct {
	ServerID    uuid.UUID           `json:"server_id"`
	Status      models.ServerStatus `json:"status"`
	IsRunning   bool                `json:"is_running"`
	LastStarted *time.Time          `json:"last_started,omitempty"`
	LastStopped *time.Time          `json:"last_stopped,omitempty"`
	AgentOnline bool                `json:"agent_online"`
}
