package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	Agent       AgentConfig
	Logging     LoggingConfig
	CORS        CORSConfig
	RateLimit   RateLimitConfig
	Upload      UploadConfig
	Marketplace MarketplaceConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port string
	Host string
	Env  string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host               string
	Port               int
	User               string
	Password           string
	Name               string
	SSLMode            string
	MaxConnections     int
	MaxIdleConnections int
	MaxLifetime        time.Duration
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// JWTConfig holds JWT token configuration
type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// AgentConfig holds agent-specific configuration
type AgentConfig struct {
	GRPCTimeout          time.Duration
	HealthCheckInterval  time.Duration
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Requests int
	Duration time.Duration
}

// UploadConfig holds file upload configuration
type UploadConfig struct {
	MaxSize int64
}

// MarketplaceConfig holds marketplace API configuration
type MarketplaceConfig struct {
	CurseForgeAPIKey string
	ModrinthAPIURL   string
	SpigotAPIURL     string
}

// Load loads configuration from environment variables and config file
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Set defaults
	setDefaults()

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Environment variables take precedence
	viper.AutomaticEnv()

	cfg := &Config{
		Server: ServerConfig{
			Port: viper.GetString("PORT"),
			Host: viper.GetString("HOST"),
			Env:  viper.GetString("ENV"),
		},
		Database: DatabaseConfig{
			Host:               viper.GetString("DB_HOST"),
			Port:               viper.GetInt("DB_PORT"),
			User:               viper.GetString("DB_USER"),
			Password:           viper.GetString("DB_PASSWORD"),
			Name:               viper.GetString("DB_NAME"),
			SSLMode:            viper.GetString("DB_SSL_MODE"),
			MaxConnections:     viper.GetInt("DB_MAX_CONNECTIONS"),
			MaxIdleConnections: viper.GetInt("DB_MAX_IDLE_CONNECTIONS"),
			MaxLifetime:        viper.GetDuration("DB_MAX_LIFETIME") * time.Second,
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetInt("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			Secret:        viper.GetString("JWT_SECRET"),
			AccessExpiry:  viper.GetDuration("JWT_ACCESS_EXPIRY"),
			RefreshExpiry: viper.GetDuration("JWT_REFRESH_EXPIRY"),
		},
		Agent: AgentConfig{
			GRPCTimeout:         viper.GetDuration("AGENT_GRPC_TIMEOUT"),
			HealthCheckInterval: viper.GetDuration("AGENT_HEALTH_CHECK_INTERVAL"),
		},
		Logging: LoggingConfig{
			Level:  viper.GetString("LOG_LEVEL"),
			Format: viper.GetString("LOG_FORMAT"),
		},
		CORS: CORSConfig{
			AllowedOrigins: viper.GetStringSlice("CORS_ALLOWED_ORIGINS"),
			AllowedMethods: viper.GetStringSlice("CORS_ALLOWED_METHODS"),
			AllowedHeaders: viper.GetStringSlice("CORS_ALLOWED_HEADERS"),
		},
		RateLimit: RateLimitConfig{
			Requests: viper.GetInt("RATE_LIMIT_REQUESTS"),
			Duration: viper.GetDuration("RATE_LIMIT_DURATION"),
		},
		Upload: UploadConfig{
			MaxSize: viper.GetInt64("MAX_UPLOAD_SIZE"),
		},
		Marketplace: MarketplaceConfig{
			CurseForgeAPIKey: viper.GetString("CURSEFORGE_API_KEY"),
			ModrinthAPIURL:   viper.GetString("MODRINTH_API_URL"),
			SpigotAPIURL:     viper.GetString("SPIGOT_API_URL"),
		},
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("PORT is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.JWT.Secret == "" || c.JWT.Secret == "your-super-secret-jwt-key-change-this-in-production" {
		return fmt.Errorf("JWT_SECRET must be set to a secure value")
	}
	return nil
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("ENV", "development")

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "aymc")
	viper.SetDefault("DB_NAME", "aymc_db")
	viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("DB_MAX_CONNECTIONS", 50)
	viper.SetDefault("DB_MAX_IDLE_CONNECTIONS", 10)
	viper.SetDefault("DB_MAX_LIFETIME", 3600)

	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("REDIS_DB", 0)

	viper.SetDefault("JWT_ACCESS_EXPIRY", "24h")
	viper.SetDefault("JWT_REFRESH_EXPIRY", "168h")

	viper.SetDefault("AGENT_GRPC_TIMEOUT", "30s")
	viper.SetDefault("AGENT_HEALTH_CHECK_INTERVAL", "30s")

	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "json")

	viper.SetDefault("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"})
	viper.SetDefault("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"})

	viper.SetDefault("RATE_LIMIT_REQUESTS", 100)
	viper.SetDefault("RATE_LIMIT_DURATION", "1m")

	viper.SetDefault("MAX_UPLOAD_SIZE", 104857600) // 100MB

	viper.SetDefault("MODRINTH_API_URL", "https://api.modrinth.com/v2")
	viper.SetDefault("SPIGOT_API_URL", "https://api.spiget.org/v2")
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}
