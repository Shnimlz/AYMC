package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ServerStatus represents the status of a Minecraft server
type ServerStatus string

const (
	ServerStatusRunning  ServerStatus = "running"
	ServerStatusStopped  ServerStatus = "stopped"
	ServerStatusStarting ServerStatus = "starting"
	ServerStatusStopping ServerStatus = "stopping"
	ServerStatusError    ServerStatus = "error"
)

// ServerType represents the type of Minecraft server
type ServerType string

const (
	ServerTypePaper   ServerType = "paper"
	ServerTypeSpigot  ServerType = "spigot"
	ServerTypePurpur  ServerType = "purpur"
	ServerTypeVanilla ServerType = "vanilla"
	ServerTypeFabric  ServerType = "fabric"
	ServerTypeForge   ServerType = "forge"
)

// Server represents a Minecraft server instance
type Server struct {
	ID          uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AgentID     uuid.UUID    `gorm:"type:uuid;not null;index" json:"agent_id" validate:"required"`
	UserID      uuid.UUID    `gorm:"type:uuid;not null;index" json:"user_id" validate:"required"`
	Name        string       `gorm:"size:100;not null;index" json:"name" validate:"required,min=3,max=100"`
	DisplayName string       `gorm:"size:100" json:"display_name"`
	ServerType  ServerType   `gorm:"type:varchar(50)" json:"server_type"`
	Version     string       `gorm:"size:20" json:"version"`
	Port        int          `json:"port" validate:"min=1024,max=65535"`
	MaxPlayers  int          `gorm:"default:20" json:"max_players" validate:"min=1,max=1000"`
	Status      ServerStatus `gorm:"type:varchar(20);default:stopped" json:"status"`
	WorkDir     string       `gorm:"type:text" json:"work_dir"`
	JavaArgs    string       `gorm:"type:text" json:"java_args"`
	AutoStart   bool         `gorm:"default:false" json:"auto_start"`
	AutoRestart bool         `gorm:"default:true" json:"auto_restart"`
	MemoryMin   int          `gorm:"default:1024" json:"memory_min" validate:"min=512"`
	MemoryMax   int          `gorm:"default:2048" json:"memory_max" validate:"min=1024"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	LastStarted *time.Time   `json:"last_started,omitempty"`
	LastStopped *time.Time   `json:"last_stopped,omitempty"`

	// Relations
	Agent   Agent          `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
	User    User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Plugins []Plugin       `gorm:"many2many:server_plugins" json:"plugins,omitempty"`
	Backups []Backup       `gorm:"foreignKey:ServerID" json:"backups,omitempty"`
	Metrics []ServerMetric `gorm:"foreignKey:ServerID" json:"metrics,omitempty"`
}

// TableName specifies the table name for Server model
func (Server) TableName() string {
	return "servers"
}

// BeforeCreate hook for Server
func (s *Server) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.DisplayName == "" {
		s.DisplayName = s.Name
	}
	return nil
}

// IsRunning checks if the server is currently running
func (s *Server) IsRunning() bool {
	return s.Status == ServerStatusRunning
}

// CanStart checks if the server can be started
func (s *Server) CanStart() bool {
	return s.Status == ServerStatusStopped || s.Status == ServerStatusError
}

// CanStop checks if the server can be stopped
func (s *Server) CanStop() bool {
	return s.Status == ServerStatusRunning || s.Status == ServerStatusStarting
}

// UpdateStatus updates the server status and timestamps
func (s *Server) UpdateStatus(status ServerStatus) {
	now := time.Now()
	s.Status = status

	if status == ServerStatusRunning {
		s.LastStarted = &now
	} else if status == ServerStatusStopped {
		s.LastStopped = &now
	}
}
