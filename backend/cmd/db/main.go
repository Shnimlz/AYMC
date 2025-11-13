package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aymc/backend/config"
	"github.com/aymc/backend/database"
	"github.com/aymc/backend/database/migrations"
	"github.com/aymc/backend/database/seeders"
	"github.com/aymc/backend/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Parse command line flags
	migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)
	migrateUp := migrateCmd.Bool("up", false, "Run migrations")
	migrateDown := migrateCmd.Bool("down", false, "Rollback migrations")

	seedCmd := flag.NewFlagSet("seed", flag.ExitOnError)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

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

	// Connect to database
	if err := database.Connect(&cfg.Database, logger.GetLogger()); err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	// Handle commands
	switch os.Args[1] {
	case "migrate":
		migrateCmd.Parse(os.Args[2:])
		if *migrateUp {
			if err := migrations.RunMigrations(database.GetDB(), logger.GetLogger()); err != nil {
				logger.Fatal("Migration failed", zap.Error(err))
			}
			logger.Info("Migrations completed successfully")
		} else if *migrateDown {
			if err := migrations.DropAllTables(database.GetDB(), logger.GetLogger()); err != nil {
				logger.Fatal("Rollback failed", zap.Error(err))
			}
			logger.Info("Rollback completed successfully")
		} else {
			fmt.Println("Use -up to run migrations or -down to rollback")
			os.Exit(1)
		}

	case "seed":
		seedCmd.Parse(os.Args[2:])
		if err := seeders.SeedAll(database.GetDB(), logger.GetLogger()); err != nil {
			logger.Fatal("Seeding failed", zap.Error(err))
		}
		logger.Info("Seeding completed successfully")

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Database CLI Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  db migrate -up      Run database migrations")
	fmt.Println("  db migrate -down    Rollback all migrations")
	fmt.Println("  db seed             Seed database with test data")
}
