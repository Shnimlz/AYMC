package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentStatus represents the status of an agent
type AgentStatus string

const (
	AgentStatusOnline  AgentStatus = "online"
	AgentStatusOffline AgentStatus = "offline"
	AgentStatusError   AgentStatus = "error"
)

// Agent represents a remote agent connected to the system
type Agent struct {
	ID                  uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AgentID             string      `gorm:"size:100;uniqueIndex;not null" json:"agent_id" validate:"required"`
	Hostname            string      `gorm:"size:255;not null" json:"hostname" validate:"required"`
	IPAddress           string      `gorm:"type:inet;not null" json:"ip_address" validate:"required,ip"`
	Port                int         `gorm:"default:50051" json:"port"`
	Status              AgentStatus `gorm:"type:varchar(20);default:offline" json:"status"`
	Version             string      `gorm:"size:20" json:"version"`
	OS                  string      `gorm:"size:50" json:"os"`
	CPUCores            int         `json:"cpu_cores"`
	MemoryTotal         int64       `json:"memory_total"`
	DiskTotal           int64       `json:"disk_total"`
	LastSeen            *time.Time  `json:"last_seen"`
	HealthCheckInterval int         `gorm:"default:30" json:"health_check_interval"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`

	// Relations
	Servers []Server `gorm:"foreignKey:AgentID" json:"servers,omitempty"`
}

// TableName specifies the table name for Agent model
func (Agent) TableName() string {
	return "agents"
}

// BeforeCreate hook for Agent
func (a *Agent) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// IsOnline checks if the agent is currently online
func (a *Agent) IsOnline() bool {
	return a.Status == AgentStatusOnline
}

// IsHealthy checks if the agent has been seen recently
func (a *Agent) IsHealthy() bool {
	if a.LastSeen == nil {
		return false
	}
	threshold := time.Duration(a.HealthCheckInterval*3) * time.Second
	return time.Since(*a.LastSeen) < threshold
}

// UpdateLastSeen updates the last seen timestamp
func (a *Agent) UpdateLastSeen() {
	now := time.Now()
	a.LastSeen = &now
}

// GetAddress returns the full gRPC address (ip:port)
func (a *Agent) GetAddress() string {
	return fmt.Sprintf("%s:%d", a.IPAddress, a.Port)
}
