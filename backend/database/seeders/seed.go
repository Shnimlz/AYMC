package seeders

import (
	"time"

	"github.com/aymc/backend/database/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedAll seeds all data
func SeedAll(db *gorm.DB, log *zap.Logger) error {
	log.Info("Seeding database...")

	// Check if data already exists
	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return err
	}

	if userCount > 0 {
		log.Info("Database already seeded, skipping...")
		return nil
	}

	// Seed users
	if err := seedUsers(db, log); err != nil {
		return err
	}

	// Seed agents
	if err := seedAgents(db, log); err != nil {
		return err
	}

	// Seed servers
	if err := seedServers(db, log); err != nil {
		return err
	}

	// Seed plugins
	if err := seedPlugins(db, log); err != nil {
		return err
	}

	log.Info("Database seeding completed successfully")
	return nil
}

// seedUsers creates default users
func seedUsers(db *gorm.DB, log *zap.Logger) error {
	log.Info("Seeding users...")

	// Hash passwords
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	demoPassword, _ := bcrypt.GenerateFromPassword([]byte("demo123"), bcrypt.DefaultCost)

	users := []models.User{
		{
			ID:           uuid.New(),
			Username:     "admin",
			Email:        "admin@aymc.local",
			PasswordHash: string(adminPassword),
			Role:         models.RoleAdmin,
			IsActive:     true,
		},
		{
			ID:           uuid.New(),
			Username:     "demo",
			Email:        "demo@aymc.local",
			PasswordHash: string(demoPassword),
			Role:         models.RoleUser,
			IsActive:     true,
		},
		{
			ID:           uuid.New(),
			Username:     "viewer",
			Email:        "viewer@aymc.local",
			PasswordHash: string(demoPassword),
			Role:         models.RoleViewer,
			IsActive:     true,
		},
	}

	if err := db.Create(&users).Error; err != nil {
		return err
	}

	log.Info("Users seeded", zap.Int("count", len(users)))
	return nil
}

// seedAgents creates example agents
func seedAgents(db *gorm.DB, log *zap.Logger) error {
	log.Info("Seeding agents...")

	now := time.Now()
	agents := []models.Agent{
		{
			ID:                  uuid.New(),
			AgentID:             "agent-local-001",
			Hostname:            "localhost",
			IPAddress:           "127.0.0.1",
			Port:                50051,
			Status:              models.AgentStatusOffline,
			Version:             "1.0.0",
			OS:                  "Linux",
			CPUCores:            8,
			MemoryTotal:         16 * 1024 * 1024 * 1024, // 16GB
			DiskTotal:           500 * 1024 * 1024 * 1024, // 500GB
			LastSeen:            &now,
			HealthCheckInterval: 30,
		},
		{
			ID:                  uuid.New(),
			AgentID:             "agent-prod-001",
			Hostname:            "mc-server-01",
			IPAddress:           "10.0.1.100",
			Port:                50051,
			Status:              models.AgentStatusOffline,
			Version:             "1.0.0",
			OS:                  "Linux",
			CPUCores:            16,
			MemoryTotal:         32 * 1024 * 1024 * 1024, // 32GB
			DiskTotal:           1024 * 1024 * 1024 * 1024, // 1TB
			LastSeen:            &now,
			HealthCheckInterval: 30,
		},
	}

	if err := db.Create(&agents).Error; err != nil {
		return err
	}

	log.Info("Agents seeded", zap.Int("count", len(agents)))
	return nil
}

// seedServers creates example servers
func seedServers(db *gorm.DB, log *zap.Logger) error {
	log.Info("Seeding servers...")

	// Get demo user
	var demoUser models.User
	if err := db.Where("username = ?", "demo").First(&demoUser).Error; err != nil {
		return err
	}

	// Get local agent
	var localAgent models.Agent
	if err := db.Where("agent_id = ?", "agent-local-001").First(&localAgent).Error; err != nil {
		return err
	}

	servers := []models.Server{
		{
			ID:          uuid.New(),
			AgentID:     localAgent.ID,
			UserID:      demoUser.ID,
			Name:        "survival-server",
			DisplayName: "Survival Server",
			ServerType:  models.ServerTypePaper,
			Version:     "1.20.1",
			Port:        25565,
			MaxPlayers:  50,
			Status:      models.ServerStatusStopped,
			WorkDir:     "/opt/minecraft/survival",
			JavaArgs:    "-Xmx4G -Xms2G",
			AutoStart:   false,
			AutoRestart: true,
			MemoryMin:   2048,
			MemoryMax:   4096,
		},
		{
			ID:          uuid.New(),
			AgentID:     localAgent.ID,
			UserID:      demoUser.ID,
			Name:        "creative-server",
			DisplayName: "Creative Server",
			ServerType:  models.ServerTypePaper,
			Version:     "1.20.1",
			Port:        25566,
			MaxPlayers:  20,
			Status:      models.ServerStatusStopped,
			WorkDir:     "/opt/minecraft/creative",
			JavaArgs:    "-Xmx2G -Xms1G",
			AutoStart:   false,
			AutoRestart: true,
			MemoryMin:   1024,
			MemoryMax:   2048,
		},
		{
			ID:          uuid.New(),
			AgentID:     localAgent.ID,
			UserID:      demoUser.ID,
			Name:        "modded-server",
			DisplayName: "Modded Server (Fabric)",
			ServerType:  models.ServerTypeFabric,
			Version:     "1.20.1",
			Port:        25567,
			MaxPlayers:  30,
			Status:      models.ServerStatusStopped,
			WorkDir:     "/opt/minecraft/modded",
			JavaArgs:    "-Xmx6G -Xms4G",
			AutoStart:   false,
			AutoRestart: true,
			MemoryMin:   4096,
			MemoryMax:   6144,
		},
	}

	if err := db.Create(&servers).Error; err != nil {
		return err
	}

	log.Info("Servers seeded", zap.Int("count", len(servers)))
	return nil
}

// seedPlugins creates example plugins
func seedPlugins(db *gorm.DB, log *zap.Logger) error {
	log.Info("Seeding plugins...")

	plugins := []models.Plugin{
		{
			ID:          uuid.New(),
			Name:        "EssentialsX",
			Slug:        "essentialsx",
			Description: "The essential plugin suite for Minecraft servers",
			Author:      "EssentialsX Team",
			Version:     "2.20.1",
			DownloadURL: "https://example.com/essentialsx.jar",
			Source:      models.PluginSourceSpigot,
			SourceID:    "9089",
			Category:    "Admin Tools",
			Downloads:   5000000,
			Rating:      4.8,
			IsActive:    true,
		},
		{
			ID:          uuid.New(),
			Name:        "WorldEdit",
			Slug:        "worldedit",
			Description: "Fast world editing for builders, terraformers and more",
			Author:      "sk89q",
			Version:     "7.2.15",
			DownloadURL: "https://example.com/worldedit.jar",
			Source:      models.PluginSourceSpigot,
			SourceID:    "13932",
			Category:    "World Editing",
			Downloads:   10000000,
			Rating:      4.9,
			IsActive:    true,
		},
		{
			ID:          uuid.New(),
			Name:        "Vault",
			Slug:        "vault",
			Description: "Economy and permissions API for Bukkit/Spigot",
			Author:      "MilkBowl",
			Version:     "1.7.3",
			DownloadURL: "https://example.com/vault.jar",
			Source:      models.PluginSourceSpigot,
			SourceID:    "34315",
			Category:    "Developer Tools",
			Downloads:   15000000,
			Rating:      4.7,
			IsActive:    true,
		},
		{
			ID:          uuid.New(),
			Name:        "LuckPerms",
			Slug:        "luckperms",
			Description: "A permissions plugin for Minecraft servers",
			Author:      "Luck",
			Version:     "5.4.102",
			DownloadURL: "https://example.com/luckperms.jar",
			Source:      models.PluginSourceSpigot,
			SourceID:    "28140",
			Category:    "Admin Tools",
			Downloads:   8000000,
			Rating:      5.0,
			IsActive:    true,
		},
		{
			ID:          uuid.New(),
			Name:        "CoreProtect",
			Slug:        "coreprotect",
			Description: "Fast, efficient block logging and rollback tool",
			Author:      "Intelli",
			Version:     "21.2",
			DownloadURL: "https://example.com/coreprotect.jar",
			Source:      models.PluginSourceSpigot,
			SourceID:    "8631",
			Category:    "Admin Tools",
			Downloads:   3000000,
			Rating:      4.9,
			IsActive:    true,
		},
	}

	if err := db.Create(&plugins).Error; err != nil {
		return err
	}

	log.Info("Plugins seeded", zap.Int("count", len(plugins)))
	return nil
}
