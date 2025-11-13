package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aymc/backend/api/rest/handlers"
	"github.com/aymc/backend/api/rest/middleware"
	"github.com/aymc/backend/api/websocket"
	"github.com/aymc/backend/config"
	"github.com/aymc/backend/services/agents"
	"github.com/aymc/backend/services/auth"
	"github.com/aymc/backend/services/backup"
	"github.com/aymc/backend/services/marketplace"
	"github.com/aymc/backend/services/server"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server represents the REST API server
type Server struct {
	router            *gin.Engine
	httpServer        *http.Server
	config            *config.Config
	authHandler       *handlers.AuthHandler
	serverHandler     *handlers.ServerHandler
	agentHandler      *handlers.AgentHandler
	marketplaceHandler *handlers.MarketplaceHandler
	backupHandler     *handlers.BackupHandler
	wsHandler         *websocket.Handler
	jwtService        *auth.JWTService
	logger            *zap.Logger
}

// NewServer creates a new REST API server
func NewServer(cfg *config.Config, jwtService *auth.JWTService, authService *auth.AuthService, serverService *server.ServerService, agentService *agents.AgentService, marketplaceService *marketplace.Service, backupService *backup.Service, backupScheduler *backup.Scheduler, wsHub *websocket.Hub, logger *zap.Logger) *Server {
	// Set Gin mode based on environment
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	// Create handlers
	authHandler := handlers.NewAuthHandler(authService, logger)
	serverHandler := handlers.NewServerHandler(serverService, logger)
	agentHandler := handlers.NewAgentHandler(agentService, logger)
	marketplaceHandler := handlers.NewMarketplaceHandler(marketplaceService, logger)
	backupHandler := handlers.NewBackupHandler(backupService, backupScheduler, logger)
	wsHandler := websocket.NewHandler(wsHub, jwtService, logger)

	server := &Server{
		router:            router,
		config:            cfg,
		authHandler:       authHandler,
		serverHandler:     serverHandler,
		agentHandler:      agentHandler,
		marketplaceHandler: marketplaceHandler,
		backupHandler:     backupHandler,
		wsHandler:         wsHandler,
		jwtService:        jwtService,
		logger:            logger,
	}

	// Setup middleware and routes
	server.setupMiddleware()
	server.setupRoutes()

	return server
}

// setupMiddleware configures global middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Custom logger middleware
	s.router.Use(s.loggerMiddleware())

	// CORS middleware
	s.router.Use(s.corsMiddleware())

	// Request ID middleware
	s.router.Use(s.requestIDMiddleware())
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check endpoint (no auth required)
	s.router.GET("/health", s.healthCheck)
	s.router.GET("/", s.welcome)

	// WebSocket endpoint (authentication handled in handler)
	s.router.GET("/api/v1/ws", s.wsHandler.HandleWebSocket)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Public auth routes
		authPublic := v1.Group("/auth")
		{
			authPublic.POST("/register", s.authHandler.Register)
			authPublic.POST("/login", s.authHandler.Login)
			authPublic.POST("/refresh", s.authHandler.RefreshToken)
		}

		// Protected auth routes (require authentication)
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.AuthMiddleware(s.jwtService, s.logger))
		{
			authProtected.GET("/me", s.authHandler.GetProfile)
			authProtected.POST("/logout", s.authHandler.Logout)
			authProtected.POST("/change-password", s.authHandler.ChangePassword)
		}

		// Protected API routes (require authentication)
		api := v1.Group("")
		api.Use(middleware.AuthMiddleware(s.jwtService, s.logger))
		{
			// Server management routes
			servers := api.Group("/servers")
			{
				servers.GET("", s.serverHandler.List)
				servers.POST("", s.serverHandler.Create)
				servers.GET("/:id", s.serverHandler.Get)
				servers.PUT("/:id", s.serverHandler.Update)
				servers.DELETE("/:id", s.serverHandler.Delete)
				
				// Server control routes
				servers.POST("/:id/start", s.serverHandler.Start)
				servers.POST("/:id/stop", s.serverHandler.Stop)
				servers.POST("/:id/restart", s.serverHandler.Restart)
				servers.GET("/:id/status", s.serverHandler.GetStatus)
			}

			// Agent management routes
			agents := api.Group("/agents")
			{
				agents.GET("", s.agentHandler.ListAgents)
				agents.GET("/stats", s.agentHandler.GetAgentStats)
				agents.GET("/:id", s.agentHandler.GetAgent)
				agents.GET("/:id/health", s.agentHandler.GetAgentHealth)
				agents.GET("/:id/metrics", s.agentHandler.GetAgentMetrics)
			}

			// Marketplace routes
			marketplace := api.Group("/marketplace")
			{
				// Search plugins
				marketplace.GET("/search", s.marketplaceHandler.SearchPlugins)
				
				// Plugin details
				marketplace.GET("/:source/:id", s.marketplaceHandler.GetPluginDetails)
				marketplace.GET("/:source/:id/versions", s.marketplaceHandler.GetPluginVersions)
				
				// Server plugin management
				marketplace.GET("/servers/:server_id/plugins", s.marketplaceHandler.ListInstalledPlugins)
				marketplace.POST("/servers/:server_id/plugins/install", s.marketplaceHandler.InstallPlugin)
				marketplace.POST("/servers/:server_id/plugins/uninstall", s.marketplaceHandler.UninstallPlugin)
				marketplace.POST("/servers/:server_id/plugins/update", s.marketplaceHandler.UpdatePlugin)
			}

			// Backup routes
			backups := api.Group("/backups")
			{
				// Backup details
				backups.GET("/:backup_id", s.backupHandler.GetBackup)
				backups.DELETE("/:backup_id", s.backupHandler.DeleteBackup)
				backups.POST("/:backup_id/restore", s.backupHandler.RestoreBackup)
			}

			// Server backup management
			servers.GET("/:id/backups", s.backupHandler.ListBackups)
			servers.POST("/:id/backups", s.backupHandler.CreateBackup)
			servers.POST("/:id/backups/manual", s.backupHandler.RunManualBackup)
			servers.GET("/:id/backup-config", s.backupHandler.GetBackupConfig)
			servers.PUT("/:id/backup-config", s.backupHandler.UpdateBackupConfig)
			servers.GET("/:id/backup-stats", s.backupHandler.GetBackupStats)

			// Protected example endpoint
			api.GET("/protected", func(c *gin.Context) {
				user := middleware.MustGetUser(c)
				c.JSON(http.StatusOK, gin.H{
					"message": "This is a protected endpoint",
					"user": gin.H{
						"id":       user.ID,
						"username": user.Username,
						"role":     user.Role,
					},
				})
			})
		}

		// Admin-only routes
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(s.jwtService, s.logger))
		admin.Use(middleware.RequireAdmin())
		{
			// Future admin endpoints
			admin.GET("/stats", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Admin stats endpoint",
				})
			})
		}
	}
}

// healthCheck returns server health status
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "healthy",
		"timestamp":   time.Now().UTC(),
		"environment": s.config.Server.Env,
	})
}

// welcome returns welcome message
func (s *Server) welcome(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to AYMC Backend API",
		"docs":    "/api/v1/docs",
	})
}

// loggerMiddleware creates a custom logger middleware
func (s *Server) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		logFields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
		}

		if len(c.Errors) > 0 {
			s.logger.Error("Request completed with errors", logFields...)
		} else if statusCode >= 500 {
			s.logger.Error("Request failed", logFields...)
		} else if statusCode >= 400 {
			s.logger.Warn("Request failed", logFields...)
		} else {
			s.logger.Info("Request completed", logFields...)
		}
	}
}

// corsMiddleware configures CORS headers
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// En desarrollo, permitir cualquier origen
		// En producción, deberías especificar orígenes específicos en la configuración
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		
		// Para desarrollo local
		allowedOrigins := []string{
			"http://localhost:5173",  // Vite dev server
			"http://localhost:3000",  // Alternativo
			"http://localhost:8080",  // Mismo puerto (por si acaso)
			"tauri://localhost",      // Tauri app
			"https://tauri.localhost", // Tauri app (HTTPS)
		}
		
		// Verificar si el origen está en la lista permitida o si es "*" en desarrollo
		allowed := s.config.Server.Env != "production"
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
			c.Writer.Header().Set("Access-Control-Max-Age", "3600") // Cache preflight por 1 hora
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// requestIDMiddleware adds a unique request ID to each request
func (s *Server) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.config.Server.Host, s.config.Server.Port)

	s.httpServer = &http.Server{
		Addr:           addr,
		Handler:        s.router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	s.logger.Info("Starting HTTP server",
		zap.String("addr", addr),
		zap.String("environment", s.config.Server.Env),
	)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to shutdown server gracefully", zap.Error(err))
		return err
	}

	s.logger.Info("HTTP server stopped")
	return nil
}

// GetRouter returns the Gin router (useful for testing)
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
