package server

import (
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

var (
	ErrServerNotFound      = errors.New("server not found")
	ErrUnauthorizedAccess  = errors.New("unauthorized access to server")
	ErrServerAlreadyExists = errors.New("server with this name already exists")
	ErrInvalidServerState  = errors.New("invalid server state for this operation")
	ErrAgentNotFound       = errors.New("agent not found")
	ErrAgentOffline        = errors.New("agent is offline")
)

// CreateServerRequest represents the request to create a new server
type CreateServerRequest struct {
	AgentID     string            `json:"agent_id" validate:"required,uuid"`
	Name        string            `json:"name" validate:"required,min=3,max=100,alphanum"`
	DisplayName string            `json:"display_name,omitempty" validate:"omitempty,max=100"`
	ServerType  models.ServerType `json:"server_type" validate:"required,oneof=paper spigot purpur vanilla fabric forge"`
	Version     string            `json:"version" validate:"required"`
	Port        int               `json:"port" validate:"required,min=1024,max=65535"`
	MaxPlayers  int               `json:"max_players" validate:"required,min=1,max=1000"`
	WorkDir     string            `json:"work_dir,omitempty"`
	JavaArgs    string            `json:"java_args,omitempty"`
	AutoStart   bool              `json:"auto_start"`
	AutoRestart bool              `json:"auto_restart"`
	MemoryMin   int               `json:"memory_min" validate:"required,min=512"`
	MemoryMax   int               `json:"memory_max" validate:"required,min=1024"`
}

// UpdateServerRequest represents the request to update a server
type UpdateServerRequest struct {
	DisplayName *string           `json:"display_name,omitempty" validate:"omitempty,max=100"`
	ServerType  *models.ServerType `json:"server_type,omitempty" validate:"omitempty,oneof=paper spigot purpur vanilla fabric forge"`
	Version     *string           `json:"version,omitempty"`
	Port        *int              `json:"port,omitempty" validate:"omitempty,min=1024,max=65535"`
	MaxPlayers  *int              `json:"max_players,omitempty" validate:"omitempty,min=1,max=1000"`
	WorkDir     *string           `json:"work_dir,omitempty"`
	JavaArgs    *string           `json:"java_args,omitempty"`
	AutoStart   *bool             `json:"auto_start,omitempty"`
	AutoRestart *bool             `json:"auto_restart,omitempty"`
	MemoryMin   *int              `json:"memory_min,omitempty" validate:"omitempty,min=512"`
	MemoryMax   *int              `json:"memory_max,omitempty" validate:"omitempty,min=1024"`
}

// ServerResponse represents a server in API responses
type ServerResponse struct {
	ID          uuid.UUID         `json:"id"`
	AgentID     uuid.UUID         `json:"agent_id"`
	UserID      uuid.UUID         `json:"user_id"`
	Name        string            `json:"name"`
	DisplayName string            `json:"display_name"`
	ServerType  models.ServerType `json:"server_type"`
	Version     string            `json:"version"`
	Port        int               `json:"port"`
	MaxPlayers  int               `json:"max_players"`
	Status      models.ServerStatus `json:"status"`
	WorkDir     string            `json:"work_dir"`
	JavaArgs    string            `json:"java_args"`
	AutoStart   bool              `json:"auto_start"`
	AutoRestart bool              `json:"auto_restart"`
	MemoryMin   int               `json:"memory_min"`
	MemoryMax   int               `json:"memory_max"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	LastStarted *time.Time        `json:"last_started,omitempty"`
	LastStopped *time.Time        `json:"last_stopped,omitempty"`
	
	// Relational data (optional)
	Agent *AgentInfo `json:"agent,omitempty"`
	User  *UserInfo  `json:"user,omitempty"`
}

// ServerListResponse represents a paginated list of servers
type ServerListResponse struct {
	Servers []ServerResponse `json:"servers"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	PerPage int              `json:"per_page"`
}

// AgentInfo represents basic agent information
type AgentInfo struct {
	ID       uuid.UUID          `json:"id"`
	AgentID  string             `json:"agent_id"`
	Hostname string             `json:"hostname"`
	Status   models.AgentStatus `json:"status"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

// ServerService handles server business logic
type ServerService struct {
	agentService *agents.AgentService
	logger       *zap.Logger
}

// NewServerService creates a new server service
func NewServerService(agentService *agents.AgentService, logger *zap.Logger) *ServerService {
	return &ServerService{
		agentService: agentService,
		logger:       logger,
	}
}

// Create creates a new server
func (s *ServerService) Create(userID uuid.UUID, req *CreateServerRequest) (*ServerResponse, error) {
	db := database.GetDB()

	// Parse agent ID
	agentID, err := uuid.Parse(req.AgentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agent ID: %w", err)
	}

	// Verify agent exists and is online
	var agent models.Agent
	if err := db.First(&agent, "id = ?", agentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAgentNotFound
		}
		return nil, fmt.Errorf("failed to query agent: %w", err)
	}

	if !agent.IsOnline() {
		s.logger.Warn("Attempt to create server on offline agent",
			zap.String("agent_id", agentID.String()),
			zap.String("user_id", userID.String()),
		)
		return nil, ErrAgentOffline
	}

	// Check if server name already exists for this user
	var existingServer models.Server
	if err := db.Where("user_id = ? AND name = ?", userID, req.Name).First(&existingServer).Error; err == nil {
		return nil, ErrServerAlreadyExists
	}

	// Set display name to name if not provided
	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Name
	}

	// Create server
	server := &models.Server{
		AgentID:     agentID,
		UserID:      userID,
		Name:        req.Name,
		DisplayName: displayName,
		ServerType:  req.ServerType,
		Version:     req.Version,
		Port:        req.Port,
		MaxPlayers:  req.MaxPlayers,
		Status:      models.ServerStatusStopped,
		WorkDir:     req.WorkDir,
		JavaArgs:    req.JavaArgs,
		AutoStart:   req.AutoStart,
		AutoRestart: req.AutoRestart,
		MemoryMin:   req.MemoryMin,
		MemoryMax:   req.MemoryMax,
	}

	if err := db.Create(server).Error; err != nil {
		s.logger.Error("Failed to create server", zap.Error(err))
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	s.logger.Info("Server created successfully",
		zap.String("server_id", server.ID.String()),
		zap.String("name", server.Name),
		zap.String("user_id", userID.String()),
		zap.String("agent_id", agentID.String()),
	)

	return s.toServerResponse(server, &agent, nil), nil
}

// GetByID retrieves a server by ID
func (s *ServerService) GetByID(serverID, userID uuid.UUID, isAdmin bool) (*ServerResponse, error) {
	db := database.GetDB()

	var server models.Server
	query := db.Preload("Agent").Preload("User")
	
	// Non-admin users can only access their own servers
	if !isAdmin {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&server, "id = ?", serverID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrServerNotFound
		}
		return nil, fmt.Errorf("failed to query server: %w", err)
	}

	return s.toServerResponse(&server, &server.Agent, &server.User), nil
}

// List retrieves all servers for a user (or all servers if admin)
func (s *ServerService) List(userID uuid.UUID, isAdmin bool, page, perPage int) (*ServerListResponse, error) {
	db := database.GetDB()

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	query := db.Model(&models.Server{}).Preload("Agent").Preload("User")
	
	// Non-admin users can only see their own servers
	if !isAdmin {
		query = query.Where("user_id = ?", userID)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count servers: %w", err)
	}

	// Get paginated results
	var servers []models.Server
	offset := (page - 1) * perPage
	if err := query.Offset(offset).Limit(perPage).Order("created_at DESC").Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("failed to query servers: %w", err)
	}

	// Convert to response format
	serverResponses := make([]ServerResponse, len(servers))
	for i, server := range servers {
		serverResponses[i] = *s.toServerResponse(&server, &server.Agent, &server.User)
	}

	return &ServerListResponse{
		Servers: serverResponses,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

// Update updates a server
func (s *ServerService) Update(serverID, userID uuid.UUID, isAdmin bool, req *UpdateServerRequest) (*ServerResponse, error) {
	db := database.GetDB()

	// Get existing server
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

	// Update fields
	updates := make(map[string]interface{})
	
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.ServerType != nil {
		updates["server_type"] = *req.ServerType
	}
	if req.Version != nil {
		updates["version"] = *req.Version
	}
	if req.Port != nil {
		updates["port"] = *req.Port
	}
	if req.MaxPlayers != nil {
		updates["max_players"] = *req.MaxPlayers
	}
	if req.WorkDir != nil {
		updates["work_dir"] = *req.WorkDir
	}
	if req.JavaArgs != nil {
		updates["java_args"] = *req.JavaArgs
	}
	if req.AutoStart != nil {
		updates["auto_start"] = *req.AutoStart
	}
	if req.AutoRestart != nil {
		updates["auto_restart"] = *req.AutoRestart
	}
	if req.MemoryMin != nil {
		updates["memory_min"] = *req.MemoryMin
	}
	if req.MemoryMax != nil {
		updates["memory_max"] = *req.MemoryMax
	}

	if len(updates) > 0 {
		if err := db.Model(&server).Updates(updates).Error; err != nil {
			s.logger.Error("Failed to update server", zap.Error(err))
			return nil, fmt.Errorf("failed to update server: %w", err)
		}
	}

	s.logger.Info("Server updated successfully",
		zap.String("server_id", server.ID.String()),
		zap.Int("updates", len(updates)),
	)

	// Reload to get updated data
	if err := db.Preload("Agent").Preload("User").First(&server, "id = ?", serverID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload server: %w", err)
	}

	return s.toServerResponse(&server, &server.Agent, &server.User), nil
}

// Delete deletes a server
func (s *ServerService) Delete(serverID, userID uuid.UUID, isAdmin bool) error {
	db := database.GetDB()

	// Get server
	var server models.Server
	query := db
	
	if !isAdmin {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&server, "id = ?", serverID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrServerNotFound
		}
		return fmt.Errorf("failed to query server: %w", err)
	}

	// Cannot delete running server
	if server.IsRunning() {
		return ErrInvalidServerState
	}

	// Delete server
	if err := db.Delete(&server).Error; err != nil {
		s.logger.Error("Failed to delete server", zap.Error(err))
		return fmt.Errorf("failed to delete server: %w", err)
	}

	s.logger.Info("Server deleted successfully",
		zap.String("server_id", server.ID.String()),
		zap.String("name", server.Name),
	)

	return nil
}

// toServerResponse converts a model to response format
func (s *ServerService) toServerResponse(server *models.Server, agent *models.Agent, user *models.User) *ServerResponse {
	resp := &ServerResponse{
		ID:          server.ID,
		AgentID:     server.AgentID,
		UserID:      server.UserID,
		Name:        server.Name,
		DisplayName: server.DisplayName,
		ServerType:  server.ServerType,
		Version:     server.Version,
		Port:        server.Port,
		MaxPlayers:  server.MaxPlayers,
		Status:      server.Status,
		WorkDir:     server.WorkDir,
		JavaArgs:    server.JavaArgs,
		AutoStart:   server.AutoStart,
		AutoRestart: server.AutoRestart,
		MemoryMin:   server.MemoryMin,
		MemoryMax:   server.MemoryMax,
		CreatedAt:   server.CreatedAt,
		UpdatedAt:   server.UpdatedAt,
		LastStarted: server.LastStarted,
		LastStopped: server.LastStopped,
	}

	if agent != nil {
		resp.Agent = &AgentInfo{
			ID:       agent.ID,
			AgentID:  agent.AgentID,
			Hostname: agent.Hostname,
			Status:   agent.Status,
		}
	}

	if user != nil {
		resp.User = &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		}
	}

	return resp
}
