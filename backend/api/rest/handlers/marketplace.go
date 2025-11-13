package handlers

import (
	"fmt"
	"net/http"

	"github.com/aymc/backend/api/rest/middleware"
	"github.com/aymc/backend/database/models"
	"github.com/aymc/backend/services/marketplace"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MarketplaceHandler handles marketplace endpoints
type MarketplaceHandler struct {
	marketplaceService *marketplace.Service
	validator          *validator.Validate
	logger             *zap.Logger
}

// NewMarketplaceHandler creates a new marketplace handler
func NewMarketplaceHandler(marketplaceService *marketplace.Service, logger *zap.Logger) *MarketplaceHandler {
	return &MarketplaceHandler{
		marketplaceService: marketplaceService,
		validator:          validator.New(),
		logger:             logger,
	}
}

// SearchPlugins searches for plugins across multiple sources
// @Summary Search for plugins
// @Tags marketplace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param query query string true "Search query"
// @Param sources query []string false "Sources to search (modrinth, spigot)" collectionFormat(csv)
// @Param limit query int false "Limit results (default: 20, max: 100)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Success 200 {object} marketplace.SearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/marketplace/search [get]
func (h *MarketplaceHandler) SearchPlugins(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Query parameter is required",
			Details: "Please provide a search query",
		})
		return
	}

	// Parse sources
	sources := c.QueryArray("sources")
	if len(sources) == 0 {
		sources = []string{"modrinth", "spigot"} // Default sources
	}

	// Parse limit and offset
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := parseInt(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := parseInt(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	req := marketplace.SearchRequest{
		Query:   query,
		Sources: sources,
		Limit:   limit,
		Offset:  offset,
	}

	result, err := h.marketplaceService.SearchPlugins(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to search plugins", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to search plugins",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPluginDetails gets details of a specific plugin
// @Summary Get plugin details
// @Tags marketplace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param source path string true "Plugin source (modrinth, spigot)"
// @Param id path string true "Plugin ID"
// @Success 200 {object} models.PluginSearchResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/marketplace/{source}/{id} [get]
func (h *MarketplaceHandler) GetPluginDetails(c *gin.Context) {
	source := c.Param("source")
	pluginID := c.Param("id")

	if source == "" || pluginID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid parameters",
			Details: "Source and plugin ID are required",
		})
		return
	}

	result, err := h.marketplaceService.GetPlugin(c.Request.Context(), source, pluginID)
	if err != nil {
		h.logger.Error("Failed to get plugin details",
			zap.String("source", source),
			zap.String("plugin_id", pluginID),
			zap.Error(err),
		)
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Plugin not found",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPluginVersions gets available versions of a plugin
// @Summary Get plugin versions
// @Tags marketplace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param source path string true "Plugin source (modrinth, spigot)"
// @Param id path string true "Plugin ID"
// @Param minecraft_version query string false "Filter by Minecraft version"
// @Success 200 {array} models.PluginVersion
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/marketplace/{source}/{id}/versions [get]
func (h *MarketplaceHandler) GetPluginVersions(c *gin.Context) {
	source := c.Param("source")
	pluginID := c.Param("id")
	minecraftVersion := c.Query("minecraft_version")

	if source == "" || pluginID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid parameters",
			Details: "Source and plugin ID are required",
		})
		return
	}

	versions, err := h.marketplaceService.GetPluginVersions(c.Request.Context(), source, pluginID, minecraftVersion)
	if err != nil {
		h.logger.Error("Failed to get plugin versions",
			zap.String("source", source),
			zap.String("plugin_id", pluginID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get plugin versions",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, versions)
}

// InstallPlugin installs a plugin on a server
// @Summary Install plugin on server
// @Tags marketplace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param server_id path string true "Server ID" format(uuid)
// @Param request body models.PluginInstallRequest true "Install request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/marketplace/servers/{server_id}/plugins/install [post]
func (h *MarketplaceHandler) InstallPlugin(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	serverIDStr := c.Param("server_id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid server ID",
			Details: err.Error(),
		})
		return
	}

	var req models.PluginInstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid install plugin request", zap.Error(err))
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

	// TODO: Verify user owns the server
	h.logger.Debug("Installing plugin",
		zap.String("user_id", userID.String()),
		zap.String("server_id", serverID.String()),
		zap.String("plugin_name", req.PluginName),
	)

	if err := h.marketplaceService.InstallPlugin(c.Request.Context(), serverID, req); err != nil {
		h.logger.Error("Failed to install plugin",
			zap.String("server_id", serverID.String()),
			zap.String("plugin_name", req.PluginName),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to install plugin",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Plugin installed successfully",
	})
}

// UninstallPlugin uninstalls a plugin from a server
// @Summary Uninstall plugin from server
// @Tags marketplace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param server_id path string true "Server ID" format(uuid)
// @Param request body models.PluginUninstallRequest true "Uninstall request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/marketplace/servers/{server_id}/plugins/uninstall [post]
func (h *MarketplaceHandler) UninstallPlugin(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	serverIDStr := c.Param("server_id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid server ID",
			Details: err.Error(),
		})
		return
	}

	var req models.PluginUninstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid uninstall plugin request", zap.Error(err))
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

	// TODO: Verify user owns the server
	h.logger.Debug("Uninstalling plugin",
		zap.String("user_id", userID.String()),
		zap.String("server_id", serverID.String()),
		zap.String("plugin_name", req.PluginName),
	)

	if err := h.marketplaceService.UninstallPlugin(c.Request.Context(), serverID, req); err != nil {
		h.logger.Error("Failed to uninstall plugin",
			zap.String("server_id", serverID.String()),
			zap.String("plugin_name", req.PluginName),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to uninstall plugin",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Plugin uninstalled successfully",
	})
}

// UpdatePlugin updates a plugin on a server
// @Summary Update plugin on server
// @Tags marketplace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param server_id path string true "Server ID" format(uuid)
// @Param request body models.PluginUpdateRequest true "Update request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/marketplace/servers/{server_id}/plugins/update [post]
func (h *MarketplaceHandler) UpdatePlugin(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	serverIDStr := c.Param("server_id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid server ID",
			Details: err.Error(),
		})
		return
	}

	var req models.PluginUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update plugin request", zap.Error(err))
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

	// TODO: Verify user owns the server
	h.logger.Debug("Updating plugin",
		zap.String("user_id", userID.String()),
		zap.String("server_id", serverID.String()),
		zap.String("plugin_name", req.PluginName),
	)

	if err := h.marketplaceService.UpdatePlugin(c.Request.Context(), serverID, req); err != nil {
		h.logger.Error("Failed to update plugin",
			zap.String("server_id", serverID.String()),
			zap.String("plugin_name", req.PluginName),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update plugin",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Plugin updated successfully",
	})
}

// ListInstalledPlugins lists plugins installed on a server
// @Summary List installed plugins
// @Tags marketplace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param server_id path string true "Server ID" format(uuid)
// @Success 200 {object} models.PluginListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/marketplace/servers/{server_id}/plugins [get]
func (h *MarketplaceHandler) ListInstalledPlugins(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	serverIDStr := c.Param("server_id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid server ID",
			Details: err.Error(),
		})
		return
	}

	// TODO: Verify user owns the server
	h.logger.Debug("Listing installed plugins",
		zap.String("user_id", userID.String()),
		zap.String("server_id", serverID.String()),
	)

	result, err := h.marketplaceService.ListInstalledPlugins(c.Request.Context(), serverID)
	if err != nil {
		h.logger.Error("Failed to list installed plugins",
			zap.String("server_id", serverID.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to list installed plugins",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// parseInt is a helper to parse integers
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
