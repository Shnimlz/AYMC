package migrations

import (
	"github.com/aymc/backend/database/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RunMigrations runs all database migrations
func RunMigrations(db *gorm.DB, log *zap.Logger) error {
	log.Info("Running database migrations...")

	// Auto migrate all models
	err := db.AutoMigrate(
		&models.User{},
		&models.Agent{},
		&models.Server{},
		&models.Plugin{},
		&models.ServerPlugin{},
		&models.Backup{},
		&models.ServerMetric{},
	)

	if err != nil {
		log.Error("Failed to run migrations", zap.Error(err))
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
