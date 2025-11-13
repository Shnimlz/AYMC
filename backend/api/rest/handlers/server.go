package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/aymc/backend/api/rest/middleware"
	"github.com/aymc/backend/services/server"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ServerHandler handles server endpoints
type ServerHandler struct {
	serverService *server.ServerService
	validator     *validator.Validate
	logger        *zap.Logger
}

// NewServerHandler creates a new server handler
func NewServerHandler(serverService *server.ServerService, logger *zap.Logger) *ServerHandler {
	return &ServerHandler{
		serverService: serverService,
		validator:     validator.New(),
		logger:        logger,
	}
}

// Create creates a new server
// @Summary Create a new Minecraft server
// @Tags servers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body server.CreateServerRequest true "Server data"
// @Success 201 {object} server.ServerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/v1/servers [post]
func (h *ServerHandler) Create(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req server.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create server request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Details: err.Error(),
		})
		return
	}

	// Create server
	srv, err := h.serverService.Create(userID, &req)
	if err != nil {
		if errors.Is(err, server.ErrAgentNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Agent not found",
			})
			return
		}
		if errors.Is(err, server.ErrAgentOffline) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Agent is offline",
			})
			return
		}
		if errors.Is(err, server.ErrServerAlreadyExists) {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error: "Server with this name already exists",
			})
			return
		}
		h.logger.Error("Failed to create server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to create server",
		})
		return
	}

	c.JSON(http.StatusCreated, srv)
}

// Get retrieves a server by ID
// @Summary Get server by ID
// @Tags servers
// @Produce json
// @Security BearerAuth
// @Param id path string true "Server ID (UUID)"
// @Success 200 {object} server.ServerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/servers/{id} [get]
func (h *ServerHandler) Get(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	user := middleware.MustGetUser(c)
	isAdmin := user.IsAdmin()

	serverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid server ID",
		})
		return
	}

	srv, err := h.serverService.GetByID(serverID, userID, isAdmin)
	if err != nil {
		if errors.Is(err, server.ErrServerNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Server not found",
			})
			return
		}
		h.logger.Error("Failed to get server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve server",
		})
		return
	}

	c.JSON(http.StatusOK, srv)
}

// List retrieves all servers
// @Summary List servers
// @Tags servers
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Results per page" default(20)
// @Success 200 {object} server.ServerListResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/servers [get]
func (h *ServerHandler) List(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	user := middleware.MustGetUser(c)
	isAdmin := user.IsAdmin()

	// Parse pagination params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	list, err := h.serverService.List(userID, isAdmin, page, perPage)
	if err != nil {
		h.logger.Error("Failed to list servers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve servers",
		})
		return
	}

	c.JSON(http.StatusOK, list)
}

// Update updates a server
// @Summary Update server
// @Tags servers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Server ID (UUID)"
// @Param request body server.UpdateServerRequest true "Server update data"
// @Success 200 {object} server.ServerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/servers/{id} [put]
func (h *ServerHandler) Update(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	user := middleware.MustGetUser(c)
	isAdmin := user.IsAdmin()

	serverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid server ID",
		})
		return
	}

	var req server.UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update server request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Details: err.Error(),
		})
		return
	}

	srv, err := h.serverService.Update(serverID, userID, isAdmin, &req)
	if err != nil {
		if errors.Is(err, server.ErrServerNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Server not found",
			})
			return
		}
		h.logger.Error("Failed to update server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to update server",
		})
		return
	}

	c.JSON(http.StatusOK, srv)
}

// Delete deletes a server
// @Summary Delete server
// @Tags servers
// @Produce json
// @Security BearerAuth
// @Param id path string true "Server ID (UUID)"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/servers/{id} [delete]
func (h *ServerHandler) Delete(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	user := middleware.MustGetUser(c)
	isAdmin := user.IsAdmin()

	serverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid server ID",
		})
		return
	}

	err = h.serverService.Delete(serverID, userID, isAdmin)
	if err != nil {
		if errors.Is(err, server.ErrServerNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Server not found",
			})
			return
		}
		if errors.Is(err, server.ErrInvalidServerState) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Cannot delete a running server. Stop it first.",
			})
			return
		}
		h.logger.Error("Failed to delete server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to delete server",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Server deleted successfully",
	})
}

// Start starts a server
// @Summary Start server
// @Tags servers
// @Produce json
// @Security BearerAuth
// @Param id path string true "Server ID (UUID)"
// @Success 200 {object} server.ServerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/servers/{id}/start [post]
func (h *ServerHandler) Start(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	user := middleware.MustGetUser(c)
	isAdmin := user.IsAdmin()

	serverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid server ID",
		})
		return
	}

	srv, err := h.serverService.Start(serverID, userID, isAdmin)
	if err != nil {
		if errors.Is(err, server.ErrServerNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Server not found",
			})
			return
		}
		if errors.Is(err, server.ErrAgentOffline) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Agent is offline",
			})
			return
		}
		if errors.Is(err, server.ErrInvalidServerState) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Server cannot be started in current state",
			})
			return
		}
		h.logger.Error("Failed to start server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to start server",
		})
		return
	}

	c.JSON(http.StatusOK, srv)
}

// Stop stops a server
// @Summary Stop server
// @Tags servers
// @Produce json
// @Security BearerAuth
// @Param id path string true "Server ID (UUID)"
// @Success 200 {object} server.ServerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/servers/{id}/stop [post]
func (h *ServerHandler) Stop(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	user := middleware.MustGetUser(c)
	isAdmin := user.IsAdmin()

	serverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid server ID",
		})
		return
	}

	srv, err := h.serverService.Stop(serverID, userID, isAdmin)
	if err != nil {
		if errors.Is(err, server.ErrServerNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Server not found",
			})
			return
		}
		if errors.Is(err, server.ErrInvalidServerState) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Server cannot be stopped in current state",
			})
			return
		}
		h.logger.Error("Failed to stop server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to stop server",
		})
		return
	}

	c.JSON(http.StatusOK, srv)
}

// Restart restarts a server
// @Summary Restart server
// @Tags servers
// @Produce json
// @Security BearerAuth
// @Param id path string true "Server ID (UUID)"
// @Success 200 {object} server.ServerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/servers/{id}/restart [post]
func (h *ServerHandler) Restart(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	user := middleware.MustGetUser(c)
	isAdmin := user.IsAdmin()

	serverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid server ID",
		})
		return
	}

	srv, err := h.serverService.Restart(serverID, userID, isAdmin)
	if err != nil {
		if errors.Is(err, server.ErrServerNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Server not found",
			})
			return
		}
		if errors.Is(err, server.ErrAgentOffline) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Agent is offline",
			})
			return
		}
		h.logger.Error("Failed to restart server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to restart server",
		})
		return
	}

	c.JSON(http.StatusOK, srv)
}

// GetStatus retrieves server status
// @Summary Get server status
// @Tags servers
// @Produce json
// @Security BearerAuth
// @Param id path string true "Server ID (UUID)"
// @Success 200 {object} server.ServerStatusResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/servers/{id}/status [get]
func (h *ServerHandler) GetStatus(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	user := middleware.MustGetUser(c)
	isAdmin := user.IsAdmin()

	serverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid server ID",
		})
		return
	}

	status, err := h.serverService.GetStatus(serverID, userID, isAdmin)
	if err != nil {
		if errors.Is(err, server.ErrServerNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Server not found",
			})
			return
		}
		h.logger.Error("Failed to get server status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve server status",
		})
		return
	}

	c.JSON(http.StatusOK, status)
}
