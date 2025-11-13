package models

import (
	"time"

	"github.com/google/uuid"
)

// ServerMetric represents performance metrics for a server
type ServerMetric struct {
	ID            uint      `gorm:"primary_key;autoIncrement" json:"id"`
	ServerID      uuid.UUID `gorm:"type:uuid;not null;index:idx_metrics_server_timestamp" json:"server_id" validate:"required"`
	Timestamp     time.Time `gorm:"default:CURRENT_TIMESTAMP;index:idx_metrics_server_timestamp" json:"timestamp"`
	CPUPercent    float64   `json:"cpu_percent"`
	MemoryUsed    int64     `json:"memory_used"`
	PlayersOnline int       `json:"players_online"`
	TPS           float64   `json:"tps"`
	UptimeSeconds int64     `json:"uptime_seconds"`

	// Relations
	Server Server `gorm:"foreignKey:ServerID" json:"server,omitempty"`
}

// TableName specifies the table name for ServerMetric model
func (ServerMetric) TableName() string {
	return "server_metrics"
}
