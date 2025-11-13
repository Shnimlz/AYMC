package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aymc/backend/api/rest"
	"github.com/aymc/backend/api/websocket"
	"github.com/aymc/backend/config"
	"github.com/aymc/backend/database"
	"github.com/aymc/backend/database/migrations"
	"github.com/aymc/backend/pkg/logger"
	"github.com/aymc/backend/services/agents"
	"github.com/aymc/backend/services/auth"
	"github.com/aymc/backend/services/backup"
	"github.com/aymc/backend/services/marketplace"
	"github.com/aymc/backend/services/server"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.Logging.Level, cfg.Logging.Format); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting AYMC Backend Server",
		zap.String("version", "0.1.0"),
		zap.String("env", cfg.Server.Env),
		zap.String("port", cfg.Server.Port),
	)

	// Initialize database connection
	if err := database.Connect(&cfg.Database, logger.GetLogger()); err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	// Run migrations
	if err := migrations.RunMigrations(database.GetDB(), logger.GetLogger()); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Initialize JWT service
	jwtService := auth.NewJWTService(cfg.JWT.Secret, logger.GetLogger())
	logger.Info("JWT service initialized")

	// Initialize auth service
	authService := auth.NewAuthService(jwtService, logger.GetLogger())
	logger.Info("Auth service initialized")

	// Initialize agent registry
	agentRegistry := agents.NewAgentRegistry(logger.GetLogger())
	logger.Info("Agent registry initialized")

	// Load agents from database and connect
	ctx := context.Background()
	if err := agentRegistry.LoadAgentsFromDatabase(ctx); err != nil {
		logger.Warn("Failed to load agents from database", zap.Error(err))
	}

	// Initialize health monitor
	healthMonitor := agents.NewHealthMonitor(agentRegistry, 30*time.Second, logger.GetLogger())
	if err := healthMonitor.Start(); err != nil {
		logger.Fatal("Failed to start health monitor", zap.Error(err))
	}
	logger.Info("Health monitor started")

	// Initialize agent service
	agentService := agents.NewAgentService(agentRegistry, logger.GetLogger())
	logger.Info("Agent service initialized")

	// Initialize server service
	serverService := server.NewServerService(agentService, logger.GetLogger())
	logger.Info("Server service initialized")

	// Initialize marketplace service
	marketplaceService := marketplace.NewService(database.GetDB(), agentService, logger.GetLogger())
	logger.Info("Marketplace service initialized")

	// Initialize backup service
	backupDir := cfg.Server.Host + "/backups" // TODO: hacer esto configurable
	backupService := backup.NewService(database.GetDB(), agentService, logger.GetLogger(), backupDir)
	logger.Info("Backup service initialized")

	// Initialize backup scheduler
	backupScheduler := backup.NewScheduler(database.GetDB(), backupService, logger.GetLogger())
	if err := backupScheduler.Start(); err != nil {
		logger.Fatal("Failed to start backup scheduler", zap.Error(err))
	}
	logger.Info("Backup scheduler started")

	// Initialize WebSocket hub
	wsHub := websocket.NewHub(logger.GetLogger())
	logger.Info("WebSocket hub initialized")

	// Start WebSocket hub in a goroutine
	go wsHub.Run()

	// Initialize REST API server
	apiServer := rest.NewServer(cfg, jwtService, authService, serverService, agentService, marketplaceService, backupService, backupScheduler, wsHub, logger.GetLogger())
	logger.Info("REST API server initialized")

	// Start server in a goroutine
	go func() {
		if err := apiServer.Start(); err != nil {
			logger.Fatal("Failed to start REST API server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Stop backup scheduler
	backupScheduler.Stop()
	logger.Info("Backup scheduler stopped")

	// Stop WebSocket hub
	wsHub.Stop()
	logger.Info("WebSocket hub stopped")

	// Stop health monitor
	healthMonitor.Stop()
	logger.Info("Health monitor stopped")

	// Shutdown agent registry (close all connections)
	agentRegistry.Shutdown()
	logger.Info("Agent registry shutdown complete")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := apiServer.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown server gracefully", zap.Error(err))
	}

	logger.Info("Server exited successfully")
}
