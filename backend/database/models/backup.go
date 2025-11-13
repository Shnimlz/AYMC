package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BackupType represents the type of backup
type BackupType string

const (
	BackupTypeFull    BackupType = "full"
	BackupTypeWorld   BackupType = "world"
	BackupTypePlugins BackupType = "plugins"
	BackupTypeConfig  BackupType = "config"
)

// BackupStatus represents the status of a backup
type BackupStatus string

const (
	BackupStatusPending    BackupStatus = "pending"
	BackupStatusInProgress BackupStatus = "in_progress"
	BackupStatusCompleted  BackupStatus = "completed"
	BackupStatusFailed     BackupStatus = "failed"
)

// Backup represents a server backup
type Backup struct {
	ID          uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ServerID    uuid.UUID    `gorm:"type:uuid;not null;index" json:"server_id" validate:"required"`
	Filename    string       `gorm:"size:255;not null" json:"filename" validate:"required"`
	Path        string       `gorm:"type:text;not null" json:"path" validate:"required"`
	SizeBytes   int64        `json:"size_bytes"`
	BackupType  BackupType   `gorm:"type:varchar(20)" json:"backup_type"`
	Status      BackupStatus `gorm:"type:varchar(20);default:pending" json:"status"`
	Compression string       `gorm:"size:10;default:gzip" json:"compression"`
	CreatedBy   *uuid.UUID   `gorm:"type:uuid" json:"created_by"`
	CreatedAt   time.Time    `json:"created_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`

	// Relations
	Server Server `gorm:"foreignKey:ServerID" json:"server,omitempty"`
	User   *User  `gorm:"foreignKey:CreatedBy" json:"user,omitempty"`
}

// TableName specifies the table name for Backup model
func (Backup) TableName() string {
	return "backups"
}

// BeforeCreate hook for Backup
func (b *Backup) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// IsCompleted checks if the backup is completed
func (b *Backup) IsCompleted() bool {
	return b.Status == BackupStatusCompleted
}

// MarkCompleted marks the backup as completed
func (b *Backup) MarkCompleted() {
	now := time.Now()
	b.Status = BackupStatusCompleted
	b.CompletedAt = &now
}

// MarkFailed marks the backup as failed
func (b *Backup) MarkFailed() {
	b.Status = BackupStatusFailed
}

// FileSizeMB retorna el tamaño del backup en MB
func (b *Backup) FileSizeMB() float64 {
	return float64(b.SizeBytes) / (1024 * 1024)
}

// FileSizeGB retorna el tamaño del backup en GB
func (b *Backup) FileSizeGB() float64 {
	return float64(b.SizeBytes) / (1024 * 1024 * 1024)
}

// BackupConfig representa la configuración de backups de un servidor
type BackupConfig struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ServerID          uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"server_id"`
	Server            *Server        `gorm:"foreignKey:ServerID" json:"server,omitempty"`
	Enabled           bool           `gorm:"default:false" json:"enabled"`
	AutoBackup        bool           `gorm:"default:false" json:"auto_backup"`
	BackupType        BackupType     `gorm:"size:50;default:'full'" json:"backup_type"`
	Schedule          string         `gorm:"size:100" json:"schedule"` // Cron expression
	MaxBackups        int            `gorm:"default:10" json:"max_backups"`
	RetentionDays     int            `gorm:"default:30" json:"retention_days"`
	CompressBackups   bool           `gorm:"default:true" json:"compress_backups"`
	IncludeWorld      bool           `gorm:"default:true" json:"include_world"`
	IncludePlugins    bool           `gorm:"default:true" json:"include_plugins"`
	IncludeConfig     bool           `gorm:"default:true" json:"include_config"`
	IncludeLogs       bool           `gorm:"default:false" json:"include_logs"`
	ExcludePaths      []string       `gorm:"type:jsonb" json:"exclude_paths"`
	NotifyOnComplete  bool           `gorm:"default:true" json:"notify_on_complete"`
	NotifyOnFailure   bool           `gorm:"default:true" json:"notify_on_failure"`
	StorageType       string         `gorm:"size:50;default:'local'" json:"storage_type"`
	StoragePath       string         `gorm:"size:512" json:"storage_path"`
	LastBackupAt      *time.Time     `json:"last_backup_at,omitempty"`
	NextBackupAt      *time.Time     `json:"next_backup_at,omitempty"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName especifica el nombre de la tabla
func (BackupConfig) TableName() string {
	return "backup_configs"
}

// BeforeCreate hook para BackupConfig
func (bc *BackupConfig) BeforeCreate(tx *gorm.DB) error {
	if bc.ID == uuid.Nil {
		bc.ID = uuid.New()
	}
	return nil
}

// --- DTOs para API ---

// CreateBackupRequest representa una solicitud para crear un backup
type CreateBackupRequest struct {
	ServerID    uuid.UUID  `json:"server_id" validate:"required"`
	Filename    string     `json:"filename" validate:"required,min=3,max=255"`
	BackupType  BackupType `json:"backup_type" validate:"required,oneof=full world plugins config"`
	Compression string     `json:"compression" validate:"omitempty,oneof=gzip bzip2 none"`
}

// RestoreBackupRequest representa una solicitud para restaurar un backup
type RestoreBackupRequest struct {
	BackupID           uuid.UUID `json:"backup_id" validate:"required"`
	ServerID           uuid.UUID `json:"server_id" validate:"required"`
	StopServer         bool      `json:"stop_server"`
	BackupBeforeRestore bool     `json:"backup_before_restore"`
	RestoreWorld       bool      `json:"restore_world"`
	RestorePlugins     bool      `json:"restore_plugins"`
	RestoreConfig      bool      `json:"restore_config"`
}

// BackupListResponse representa la respuesta de listar backups
type BackupListResponse struct {
	Backups    []Backup   `json:"backups"`
	Total      int        `json:"total"`
	TotalSize  int64      `json:"total_size"`
	OldestDate *time.Time `json:"oldest_date,omitempty"`
	NewestDate *time.Time `json:"newest_date,omitempty"`
}

// UpdateBackupConfigRequest representa una solicitud para actualizar config de backups
type UpdateBackupConfigRequest struct {
	Enabled          *bool      `json:"enabled"`
	AutoBackup       *bool      `json:"auto_backup"`
	BackupType       BackupType `json:"backup_type" validate:"omitempty,oneof=full world plugins config"`
	Schedule         string     `json:"schedule"`
	MaxBackups       *int       `json:"max_backups" validate:"omitempty,min=1,max=100"`
	RetentionDays    *int       `json:"retention_days" validate:"omitempty,min=1,max=365"`
	CompressBackups  *bool      `json:"compress_backups"`
	IncludeWorld     *bool      `json:"include_world"`
	IncludePlugins   *bool      `json:"include_plugins"`
	IncludeConfig    *bool      `json:"include_config"`
	IncludeLogs      *bool      `json:"include_logs"`
	ExcludePaths     []string   `json:"exclude_paths"`
	NotifyOnComplete *bool      `json:"notify_on_complete"`
	NotifyOnFailure  *bool      `json:"notify_on_failure"`
	StorageType      string     `json:"storage_type" validate:"omitempty,oneof=local s3"`
	StoragePath      string     `json:"storage_path"`
}

// BackupStats representa estadísticas de backups
type BackupStats struct {
	TotalBackups      int        `json:"total_backups"`
	TotalSize         int64      `json:"total_size"`
	TotalSizeGB       float64    `json:"total_size_gb"`
	SuccessfulBackups int        `json:"successful_backups"`
	FailedBackups     int        `json:"failed_backups"`
	OldestBackup      *time.Time `json:"oldest_backup,omitempty"`
	LatestBackup      *time.Time `json:"latest_backup,omitempty"`
	AvgBackupSize     float64    `json:"avg_backup_size_gb"`
}

