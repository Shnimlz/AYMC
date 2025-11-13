package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// PluginSource represents the source of a plugin
type PluginSource string

const (
	PluginSourceSpigot     PluginSource = "spigot"
	PluginSourceModrinth   PluginSource = "modrinth"
	PluginSourceCurseForge PluginSource = "curseforge"
	PluginSourceGitHub     PluginSource = "github"
	PluginSourceCustom     PluginSource = "custom"
)

// Plugin represents a Minecraft plugin/mod
type Plugin struct {
	ID                uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name              string           `gorm:"size:100;not null;index" json:"name" validate:"required"`
	Slug              string           `gorm:"size:100;uniqueIndex;not null" json:"slug" validate:"required"`
	Description       string           `gorm:"type:text" json:"description"`
	Author            string           `gorm:"size:100" json:"author"`
	Version           string           `gorm:"size:20" json:"version"`
	DownloadURL       string           `gorm:"type:text" json:"download_url"`
	IconURL           string           `gorm:"type:text" json:"icon_url"`
	Source            PluginSource     `gorm:"type:varchar(20)" json:"source"`
	SourceID          string           `gorm:"size:100;index" json:"source_id"`
	Category          string           `gorm:"size:50;index" json:"category"`
	Downloads         int64            `gorm:"default:0" json:"downloads"`
	Rating            float32          `gorm:"type:decimal(3,2)" json:"rating"`
	MinecraftVersions datatypes.JSON   `gorm:"type:jsonb" json:"minecraft_versions"`
	IsActive          bool             `gorm:"default:true" json:"is_active"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`

	// Relations
	Servers []Server `gorm:"many2many:server_plugins" json:"servers,omitempty"`
}

// TableName specifies the table name for Plugin model
func (Plugin) TableName() string {
	return "plugins"
}

// BeforeCreate hook for Plugin
func (p *Plugin) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// ServerPlugin represents the many-to-many relationship between servers and plugins
type ServerPlugin struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ServerID    uuid.UUID `gorm:"type:uuid;not null;index" json:"server_id"`
	PluginID    uuid.UUID `gorm:"type:uuid;not null;index" json:"plugin_id"`
	Version     string    `gorm:"size:20" json:"version"`
	IsEnabled   bool      `gorm:"default:true" json:"is_enabled"`
	InstalledAt time.Time `json:"installed_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	Server Server `gorm:"foreignKey:ServerID" json:"server,omitempty"`
	Plugin Plugin `gorm:"foreignKey:PluginID" json:"plugin,omitempty"`
}

// TableName specifies the table name for ServerPlugin model
func (ServerPlugin) TableName() string {
	return "server_plugins"
}

// BeforeCreate hook for ServerPlugin
func (sp *ServerPlugin) BeforeCreate(tx *gorm.DB) error {
	if sp.ID == uuid.Nil {
		sp.ID = uuid.New()
	}
	return nil
}

// PluginSearchResult representa un resultado de búsqueda de plugins desde APIs externas
type PluginSearchResult struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Slug             string   `json:"slug"`
	Description      string   `json:"description"`
	Summary          string   `json:"summary,omitempty"`
	Author           string   `json:"author"`
	IconURL          string   `json:"icon_url"`
	Source           string   `json:"source"`
	SourceID         string   `json:"source_id"`
	Category         string   `json:"category,omitempty"`
	Downloads        int64    `json:"downloads"`
	Rating           float32  `json:"rating,omitempty"`
	LatestVersion    string   `json:"latest_version,omitempty"`
	MinecraftVersions []string `json:"minecraft_versions,omitempty"`
	UpdatedAt        string   `json:"updated_at,omitempty"`
}

// PluginVersion representa una versión específica de un plugin desde APIs externas
type PluginVersion struct {
	ID               string   `json:"id"`
	PluginID         string   `json:"plugin_id"`
	VersionNumber    string   `json:"version_number"`
	VersionName      string   `json:"version_name,omitempty"`
	Changelog        string   `json:"changelog,omitempty"`
	DownloadURL      string   `json:"download_url"`
	FileName         string   `json:"file_name"`
	FileSize         int64    `json:"file_size,omitempty"`
	FileHash         string   `json:"file_hash,omitempty"`
	MinecraftVersions []string `json:"minecraft_versions"`
	Dependencies     []string `json:"dependencies,omitempty"`
	Downloads        int32    `json:"downloads"`
	ReleaseDate      string   `json:"release_date"`
	IsStable         bool     `json:"is_stable"`
}

// PluginInstallRequest representa una solicitud de instalación de plugin
type PluginInstallRequest struct {
	Source    string `json:"source" validate:"required"`
	SourceID  string `json:"source_id" validate:"required"`
	PluginName string `json:"plugin_name" validate:"required"`
	Version   string `json:"version,omitempty"`
	DownloadURL string `json:"download_url" validate:"required"`
	FileName  string `json:"file_name" validate:"required"`
	AutoUpdate bool   `json:"auto_update"`
}

// PluginUninstallRequest representa una solicitud de desinstalación
type PluginUninstallRequest struct {
	PluginName   string `json:"plugin_name" validate:"required"`
	DeleteConfig bool   `json:"delete_config"`
	DeleteData   bool   `json:"delete_data"`
}

// PluginUpdateRequest representa una solicitud de actualización
type PluginUpdateRequest struct {
	PluginName  string `json:"plugin_name" validate:"required"`
	Version     string `json:"version,omitempty"`
	DownloadURL string `json:"download_url" validate:"required"`
	FileName    string `json:"file_name" validate:"required"`
}

// PluginListResponse representa la respuesta de lista de plugins instalados
type PluginListResponse struct {
	Plugins []InstalledPlugin `json:"plugins"`
	Total   int               `json:"total"`
}

// InstalledPlugin representa un plugin instalado en un servidor
type InstalledPlugin struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	IsEnabled   bool      `json:"is_enabled"`
	Description string    `json:"description,omitempty"`
	Author      string    `json:"author,omitempty"`
	Source      string    `json:"source,omitempty"`
	InstalledAt time.Time `json:"installed_at"`
}
