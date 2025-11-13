package migrations

import (
	"github.com/aymc/backend/database/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RunMigrations runs all database migrations
func RunMigrations(db *gorm.DB, log *zap.Logger) error {
	log.Info("Running database migrations...")

	// Migrate models ONE BY ONE to ensure proper order for foreign keys
	log.Info("Migrating users table...")
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Error("Failed to migrate users", zap.Error(err))
		return err
	}

	log.Info("Migrating agents table...")
	if err := db.AutoMigrate(&models.Agent{}); err != nil {
		log.Error("Failed to migrate agents", zap.Error(err))
		return err
	}

	log.Info("Migrating plugins table...")
	if err := db.AutoMigrate(&models.Plugin{}); err != nil {
		log.Error("Failed to migrate plugins", zap.Error(err))
		return err
	}

	log.Info("Migrating servers table...")
	if err := db.AutoMigrate(&models.Server{}); err != nil {
		log.Error("Failed to migrate servers", zap.Error(err))
		return err
	}

	log.Info("Migrating server_plugins table...")
	if err := db.AutoMigrate(&models.ServerPlugin{}); err != nil {
		log.Error("Failed to migrate server_plugins", zap.Error(err))
		return err
	}

	log.Info("Migrating backups table...")
	if err := db.AutoMigrate(&models.Backup{}); err != nil {
		log.Error("Failed to migrate backups", zap.Error(err))
		return err
	}

	log.Info("Migrating backup_configs table...")
	if err := db.AutoMigrate(&models.BackupConfig{}); err != nil {
		log.Error("Failed to migrate backup_configs", zap.Error(err))
		return err
	}

	log.Info("Migrating server_metrics table...")
	if err := db.AutoMigrate(&models.ServerMetric{}); err != nil {
		log.Error("Failed to migrate server_metrics", zap.Error(err))
		return err
	}

	// Create indexes
	if err := createIndexes(db); err != nil {
		log.Error("Failed to create indexes", zap.Error(err))
		return err
	}

	log.Info("Database migrations completed successfully")
	return nil
}

// createIndexes creates additional indexes not handled by AutoMigrate
func createIndexes(db *gorm.DB) error {
	// Add composite unique index for server_plugins
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_server_plugins_unique 
		ON server_plugins(server_id, plugin_id)
	`).Error; err != nil {
		return err
	}

	// Add index for metrics timestamp ordering
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_metrics_timestamp_desc 
		ON server_metrics(timestamp DESC)
	`).Error; err != nil {
		return err
	}

	// Add full-text search index for plugins (PostgreSQL specific)
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_plugins_name_search 
		ON plugins USING gin(to_tsvector('english', name))
	`).Error; err != nil {
		return err
	}

	// Add index for backup creation date
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_backups_created_desc 
		ON backups(created_at DESC)
	`).Error; err != nil {
		return err
	}

	return nil
}

// DropAllTables drops all tables (use with caution!)
func DropAllTables(db *gorm.DB, log *zap.Logger) error {
	log.Warn("Dropping all tables...")

	err := db.Migrator().DropTable(
		&models.ServerMetric{},
		&models.Backup{},
		&models.ServerPlugin{},
		&models.Plugin{},
		&models.Server{},
		&models.Agent{},
		&models.User{},
	)

	if err != nil {
		log.Error("Failed to drop tables", zap.Error(err))
		return err
	}

	log.Info("All tables dropped successfully")
	return nil
}
