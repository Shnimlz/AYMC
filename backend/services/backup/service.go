package backup

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aymc/backend/database/models"
	"github.com/aymc/backend/services/agents"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service maneja la lógica de backups
type Service struct {
	db           *gorm.DB
	agentService *agents.AgentService
	logger       *zap.Logger
	backupDir    string // Directorio base para almacenar backups
}

// NewService crea una nueva instancia del servicio de backups
func NewService(db *gorm.DB, agentService *agents.AgentService, logger *zap.Logger, backupDir string) *Service {
	return &Service{
		db:           db,
		agentService: agentService,
		logger:       logger,
		backupDir:    backupDir,
	}
}

// CreateBackup crea un nuevo backup de un servidor
func (s *Service) CreateBackup(ctx context.Context, req *models.CreateBackupRequest, userID uuid.UUID) (*models.Backup, error) {
	s.logger.Info("Creating backup",
		zap.String("server_id", req.ServerID.String()),
		zap.String("filename", req.Filename),
		zap.String("type", string(req.BackupType)),
	)

	// Verificar que el servidor existe
	var server models.Server
	if err := s.db.First(&server, "id = ?", req.ServerID).Error; err != nil {
		return nil, fmt.Errorf("servidor no encontrado: %w", err)
	}

	// Crear registro de backup en DB
	backup := &models.Backup{
		ID:          uuid.New(),
		ServerID:    req.ServerID,
		Filename:    req.Filename,
		Path:        filepath.Join(s.backupDir, server.ID.String(), req.Filename),
		BackupType:  req.BackupType,
		Status:      models.BackupStatusPending,
		Compression: req.Compression,
		CreatedBy:   &userID,
		CreatedAt:   time.Now(),
	}

	if err := s.db.Create(backup).Error; err != nil {
		return nil, fmt.Errorf("error creando registro de backup: %w", err)
	}

	// Actualizar estado a "in progress"
	backup.Status = models.BackupStatusInProgress
	s.db.Save(backup)

	// Llamar al agente para crear el backup físico
	go s.executeBackup(ctx, backup, &server)

	return backup, nil
}

// executeBackup ejecuta el backup en segundo plano
func (s *Service) executeBackup(ctx context.Context, backup *models.Backup, server *models.Server) {
	s.logger.Info("Executing backup",
		zap.String("backup_id", backup.ID.String()),
		zap.String("server_id", server.ID.String()),
	)

	// Llamar al agente para crear el backup
	// TODO: Implementar llamada gRPC cuando se agreguen los métodos
	s.logger.Warn("Backup creation not yet implemented in agent - marking as completed")

	// Simular creación de backup por ahora
	time.Sleep(2 * time.Second)

	// Marcar como completado
	now := time.Now()
	backup.Status = models.BackupStatusCompleted
	backup.CompletedAt = &now
	backup.SizeBytes = 1024 * 1024 * 100 // 100MB placeholder

	if err := s.db.Save(backup).Error; err != nil {
		s.logger.Error("Error updating backup status", zap.Error(err))
		return
	}

	// Actualizar last_backup_at en config si existe
	var config models.BackupConfig
	if err := s.db.First(&config, "server_id = ?", server.ID).Error; err == nil {
		config.LastBackupAt = &now
		s.db.Save(&config)
	}

	// Limpiar backups antiguos según retention policy
	go s.cleanupOldBackups(server.ID)

	s.logger.Info("Backup completed successfully",
		zap.String("backup_id", backup.ID.String()),
		zap.Int64("size_bytes", backup.SizeBytes),
	)
}

// RestoreBackup restaura un backup en un servidor
func (s *Service) RestoreBackup(ctx context.Context, req *models.RestoreBackupRequest) error {
	s.logger.Info("Restoring backup",
		zap.String("backup_id", req.BackupID.String()),
		zap.String("server_id", req.ServerID.String()),
	)

	// Obtener backup
	var backup models.Backup
	if err := s.db.Preload("Server").First(&backup, "id = ?", req.BackupID).Error; err != nil {
		return fmt.Errorf("backup no encontrado: %w", err)
	}

	// Verificar que el backup pertenece al servidor
	if backup.ServerID != req.ServerID {
		return fmt.Errorf("backup no pertenece al servidor especificado")
	}

	// Verificar que el backup está completado
	if backup.Status != models.BackupStatusCompleted {
		return fmt.Errorf("backup no está completado (estado: %s)", backup.Status)
	}

	// Verificar que el servidor existe
	var server models.Server
	if err := s.db.First(&server, "id = ?", req.ServerID).Error; err != nil {
		return fmt.Errorf("servidor no encontrado: %w", err)
	}

	// Crear backup antes de restaurar si se solicita
	if req.BackupBeforeRestore {
		s.logger.Info("Creating safety backup before restore")
		backupReq := &models.CreateBackupRequest{
			ServerID:    req.ServerID,
			Filename:    fmt.Sprintf("pre-restore-%d.tar.gz", time.Now().Unix()),
			BackupType:  models.BackupTypeFull,
			Compression: "gzip",
		}
		if _, err := s.CreateBackup(ctx, backupReq, uuid.Nil); err != nil {
			s.logger.Warn("Failed to create safety backup", zap.Error(err))
		}
	}

	// Detener servidor si se solicita
	if req.StopServer {
		s.logger.Info("Stopping server before restore")
		// TODO: Implementar detención del servidor
	}

	// Llamar al agente para restaurar el backup
	// TODO: Implementar llamada gRPC cuando se agreguen los métodos
	s.logger.Warn("Backup restore not yet implemented in agent")

	s.logger.Info("Backup restore initiated",
		zap.String("backup_id", req.BackupID.String()),
	)

	return nil
}

// ListBackups lista los backups de un servidor
func (s *Service) ListBackups(ctx context.Context, serverID uuid.UUID, limit, offset int) (*models.BackupListResponse, error) {
	var backups []models.Backup
	var total int64

	query := s.db.Model(&models.Backup{}).Where("server_id = ?", serverID)

	// Contar total
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("error contando backups: %w", err)
	}

	// Obtener backups con paginación
	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&backups).Error; err != nil {
		return nil, fmt.Errorf("error obteniendo backups: %w", err)
	}

	// Calcular estadísticas
	var totalSize int64
	var oldestDate, newestDate *time.Time

	for i := range backups {
		totalSize += backups[i].SizeBytes

		if oldestDate == nil || backups[i].CreatedAt.Before(*oldestDate) {
			oldestDate = &backups[i].CreatedAt
		}
		if newestDate == nil || backups[i].CreatedAt.After(*newestDate) {
			newestDate = &backups[i].CreatedAt
		}
	}

	return &models.BackupListResponse{
		Backups:    backups,
		Total:      int(total),
		TotalSize:  totalSize,
		OldestDate: oldestDate,
		NewestDate: newestDate,
	}, nil
}

// GetBackup obtiene un backup por ID
func (s *Service) GetBackup(ctx context.Context, backupID uuid.UUID) (*models.Backup, error) {
	var backup models.Backup
	if err := s.db.Preload("Server").First(&backup, "id = ?", backupID).Error; err != nil {
		return nil, fmt.Errorf("backup no encontrado: %w", err)
	}
	return &backup, nil
}

// DeleteBackup elimina un backup
func (s *Service) DeleteBackup(ctx context.Context, backupID uuid.UUID) error {
	s.logger.Info("Deleting backup", zap.String("backup_id", backupID.String()))

	var backup models.Backup
	if err := s.db.First(&backup, "id = ?", backupID).Error; err != nil {
		return fmt.Errorf("backup no encontrado: %w", err)
	}

	// Eliminar archivo físico
	// TODO: Implementar eliminación del archivo en el agente

	// Eliminar registro de DB
	if err := s.db.Delete(&backup).Error; err != nil {
		return fmt.Errorf("error eliminando backup: %w", err)
	}

	s.logger.Info("Backup deleted successfully", zap.String("backup_id", backupID.String()))
	return nil
}

// GetBackupConfig obtiene la configuración de backups de un servidor
func (s *Service) GetBackupConfig(ctx context.Context, serverID uuid.UUID) (*models.BackupConfig, error) {
	var config models.BackupConfig
	if err := s.db.First(&config, "server_id = ?", serverID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Crear configuración por defecto
			config = models.BackupConfig{
				ID:               uuid.New(),
				ServerID:         serverID,
				Enabled:          false,
				AutoBackup:       false,
				BackupType:       models.BackupTypeFull,
				Schedule:         "0 2 * * *", // 2 AM diario
				MaxBackups:       10,
				RetentionDays:    30,
				CompressBackups:  true,
				IncludeWorld:     true,
				IncludePlugins:   true,
				IncludeConfig:    true,
				IncludeLogs:      false,
				NotifyOnComplete: true,
				NotifyOnFailure:  true,
				StorageType:      "local",
				StoragePath:      filepath.Join(s.backupDir, serverID.String()),
			}
			if err := s.db.Create(&config).Error; err != nil {
				return nil, fmt.Errorf("error creando configuración por defecto: %w", err)
			}
		} else {
			return nil, fmt.Errorf("error obteniendo configuración: %w", err)
		}
	}
	return &config, nil
}

// UpdateBackupConfig actualiza la configuración de backups
func (s *Service) UpdateBackupConfig(ctx context.Context, serverID uuid.UUID, req *models.UpdateBackupConfigRequest) (*models.BackupConfig, error) {
	s.logger.Info("Updating backup config", zap.String("server_id", serverID.String()))

	config, err := s.GetBackupConfig(ctx, serverID)
	if err != nil {
		return nil, err
	}

	// Actualizar campos si se proporcionan
	if req.Enabled != nil {
		config.Enabled = *req.Enabled
	}
	if req.AutoBackup != nil {
		config.AutoBackup = *req.AutoBackup
	}
	if req.BackupType != "" {
		config.BackupType = req.BackupType
	}
	if req.Schedule != "" {
		config.Schedule = req.Schedule
	}
	if req.MaxBackups != nil {
		config.MaxBackups = *req.MaxBackups
	}
	if req.RetentionDays != nil {
		config.RetentionDays = *req.RetentionDays
	}
	if req.CompressBackups != nil {
		config.CompressBackups = *req.CompressBackups
	}
	if req.IncludeWorld != nil {
		config.IncludeWorld = *req.IncludeWorld
	}
	if req.IncludePlugins != nil {
		config.IncludePlugins = *req.IncludePlugins
	}
	if req.IncludeConfig != nil {
		config.IncludeConfig = *req.IncludeConfig
	}
	if req.IncludeLogs != nil {
		config.IncludeLogs = *req.IncludeLogs
	}
	if req.ExcludePaths != nil {
		config.ExcludePaths = req.ExcludePaths
	}
	if req.NotifyOnComplete != nil {
		config.NotifyOnComplete = *req.NotifyOnComplete
	}
	if req.NotifyOnFailure != nil {
		config.NotifyOnFailure = *req.NotifyOnFailure
	}
	if req.StorageType != "" {
		config.StorageType = req.StorageType
	}
	if req.StoragePath != "" {
		config.StoragePath = req.StoragePath
	}

	if err := s.db.Save(config).Error; err != nil {
		return nil, fmt.Errorf("error actualizando configuración: %w", err)
	}

	s.logger.Info("Backup config updated successfully", zap.String("server_id", serverID.String()))
	return config, nil
}

// GetBackupStats obtiene estadísticas de backups de un servidor
func (s *Service) GetBackupStats(ctx context.Context, serverID uuid.UUID) (*models.BackupStats, error) {
	var stats models.BackupStats

	// Contar total de backups
	var totalBackups int64
	if err := s.db.Model(&models.Backup{}).
		Where("server_id = ?", serverID).
		Count(&totalBackups).Error; err != nil {
		return nil, fmt.Errorf("error contando backups: %w", err)
	}
	stats.TotalBackups = int(totalBackups)

	// Sumar tamaño total
	s.db.Model(&models.Backup{}).
		Where("server_id = ?", serverID).
		Select("COALESCE(SUM(size_bytes), 0)").
		Scan(&stats.TotalSize)

	stats.TotalSizeGB = float64(stats.TotalSize) / (1024 * 1024 * 1024)

	// Contar exitosos
	var successfulBackups int64
	s.db.Model(&models.Backup{}).
		Where("server_id = ? AND status = ?", serverID, models.BackupStatusCompleted).
		Count(&successfulBackups)
	stats.SuccessfulBackups = int(successfulBackups)

	// Contar fallidos
	var failedBackups int64
	s.db.Model(&models.Backup{}).
		Where("server_id = ? AND status = ?", serverID, models.BackupStatusFailed).
		Count(&failedBackups)
	stats.FailedBackups = int(failedBackups)

	// Obtener fechas
	var oldestBackup, latestBackup models.Backup
	if err := s.db.Where("server_id = ?", serverID).
		Order("created_at ASC").
		First(&oldestBackup).Error; err == nil {
		stats.OldestBackup = &oldestBackup.CreatedAt
	}

	if err := s.db.Where("server_id = ?", serverID).
		Order("created_at DESC").
		First(&latestBackup).Error; err == nil {
		stats.LatestBackup = &latestBackup.CreatedAt
	}

	// Calcular promedio
	if stats.TotalBackups > 0 {
		stats.AvgBackupSize = stats.TotalSizeGB / float64(stats.TotalBackups)
	}

	return &stats, nil
}

// cleanupOldBackups elimina backups antiguos según la política de retención
func (s *Service) cleanupOldBackups(serverID uuid.UUID) {
	s.logger.Info("Cleaning up old backups", zap.String("server_id", serverID.String()))

	// Obtener configuración
	var config models.BackupConfig
	if err := s.db.First(&config, "server_id = ?", serverID).Error; err != nil {
		s.logger.Error("Error getting backup config", zap.Error(err))
		return
	}

	// Eliminar por cantidad máxima
	var backups []models.Backup
	if err := s.db.Where("server_id = ? AND status = ?", serverID, models.BackupStatusCompleted).
		Order("created_at DESC").
		Find(&backups).Error; err != nil {
		s.logger.Error("Error listing backups", zap.Error(err))
		return
	}

	if len(backups) > config.MaxBackups {
		// Eliminar los más antiguos
		toDelete := backups[config.MaxBackups:]
		for _, backup := range toDelete {
			s.logger.Info("Deleting old backup (max limit)",
				zap.String("backup_id", backup.ID.String()),
				zap.Time("created_at", backup.CreatedAt),
			)
			s.db.Delete(&backup)
		}
	}

	// Eliminar por días de retención
	cutoffDate := time.Now().AddDate(0, 0, -config.RetentionDays)
	var oldBackups []models.Backup
	if err := s.db.Where("server_id = ? AND created_at < ?", serverID, cutoffDate).
		Find(&oldBackups).Error; err != nil {
		s.logger.Error("Error finding old backups", zap.Error(err))
		return
	}

	for _, backup := range oldBackups {
		s.logger.Info("Deleting old backup (retention)",
			zap.String("backup_id", backup.ID.String()),
			zap.Time("created_at", backup.CreatedAt),
		)
		s.db.Delete(&backup)
	}

	s.logger.Info("Cleanup completed",
		zap.String("server_id", serverID.String()),
		zap.Int("deleted_by_max", len(backups)-config.MaxBackups),
		zap.Int("deleted_by_retention", len(oldBackups)),
	)
}
