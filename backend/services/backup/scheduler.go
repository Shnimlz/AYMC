package backup

import (
	"context"
	"fmt"
	"time"

	"github.com/aymc/backend/database/models"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Scheduler maneja los backups automáticos programados
type Scheduler struct {
	db             *gorm.DB
	backupService  *Service
	logger         *zap.Logger
	cron           *cron.Cron
	scheduledJobs  map[uuid.UUID]cron.EntryID // server_id -> entry_id
}

// NewScheduler crea un nuevo scheduler de backups
func NewScheduler(db *gorm.DB, backupService *Service, logger *zap.Logger) *Scheduler {
	return &Scheduler{
		db:            db,
		backupService: backupService,
		logger:        logger,
		cron:          cron.New(cron.WithSeconds()), // Soporte para segundos
		scheduledJobs: make(map[uuid.UUID]cron.EntryID),
	}
}

// Start inicia el scheduler y carga todas las configuraciones
func (s *Scheduler) Start() error {
	s.logger.Info("Starting backup scheduler")

	// Cargar todas las configuraciones activas
	var configs []models.BackupConfig
	if err := s.db.Where("enabled = ? AND auto_backup = ?", true, true).Find(&configs).Error; err != nil {
		return fmt.Errorf("error loading backup configs: %w", err)
	}

	s.logger.Info("Loading backup configurations", zap.Int("count", len(configs)))

	// Programar cada servidor
	for _, config := range configs {
		if err := s.ScheduleServer(config.ServerID, config.Schedule); err != nil {
			s.logger.Error("Error scheduling server",
				zap.String("server_id", config.ServerID.String()),
				zap.Error(err),
			)
			continue
		}
	}

	// Iniciar cron
	s.cron.Start()

	s.logger.Info("Backup scheduler started", zap.Int("scheduled_servers", len(s.scheduledJobs)))
	return nil
}

// Stop detiene el scheduler
func (s *Scheduler) Stop() {
	s.logger.Info("Stopping backup scheduler")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("Backup scheduler stopped")
}

// ScheduleServer programa backups para un servidor específico
func (s *Scheduler) ScheduleServer(serverID uuid.UUID, schedule string) error {
	s.logger.Info("Scheduling server backups",
		zap.String("server_id", serverID.String()),
		zap.String("schedule", schedule),
	)

	// Eliminar schedule anterior si existe
	if entryID, exists := s.scheduledJobs[serverID]; exists {
		s.cron.Remove(entryID)
		delete(s.scheduledJobs, serverID)
	}

	// Validar expresión cron
	if schedule == "" {
		schedule = "0 2 * * *" // Default: 2 AM diario
	}

	// Agregar nuevo job
	entryID, err := s.cron.AddFunc(schedule, func() {
		s.executeScheduledBackup(serverID)
	})

	if err != nil {
		return fmt.Errorf("error adding cron job: %w", err)
	}

	s.scheduledJobs[serverID] = entryID

	// Actualizar next_backup_at en la configuración
	go s.updateNextBackupTime(serverID, schedule)

	s.logger.Info("Server backup scheduled successfully",
		zap.String("server_id", serverID.String()),
		zap.Int("entry_id", int(entryID)),
	)

	return nil
}

// UnscheduleServer elimina la programación de backups de un servidor
func (s *Scheduler) UnscheduleServer(serverID uuid.UUID) error {
	s.logger.Info("Unscheduling server backups", zap.String("server_id", serverID.String()))

	if entryID, exists := s.scheduledJobs[serverID]; exists {
		s.cron.Remove(entryID)
		delete(s.scheduledJobs, serverID)
		s.logger.Info("Server unscheduled successfully", zap.String("server_id", serverID.String()))
		return nil
	}

	return fmt.Errorf("server not scheduled")
}

// RescheduleServer actualiza la programación de un servidor
func (s *Scheduler) RescheduleServer(serverID uuid.UUID, newSchedule string) error {
	s.logger.Info("Rescheduling server backups",
		zap.String("server_id", serverID.String()),
		zap.String("new_schedule", newSchedule),
	)

	return s.ScheduleServer(serverID, newSchedule)
}

// executeScheduledBackup ejecuta un backup programado
func (s *Scheduler) executeScheduledBackup(serverID uuid.UUID) {
	s.logger.Info("Executing scheduled backup", zap.String("server_id", serverID.String()))

	ctx := context.Background()

	// Obtener configuración
	config, err := s.backupService.GetBackupConfig(ctx, serverID)
	if err != nil {
		s.logger.Error("Error getting backup config",
			zap.String("server_id", serverID.String()),
			zap.Error(err),
		)
		return
	}

	// Verificar que está habilitado
	if !config.Enabled || !config.AutoBackup {
		s.logger.Warn("Auto backup disabled, skipping",
			zap.String("server_id", serverID.String()),
		)
		return
	}

	// Crear nombre de archivo con timestamp
	filename := fmt.Sprintf("auto-backup-%s.tar.gz", time.Now().Format("2006-01-02-15-04-05"))

	// Determinar compresión
	compression := "gzip"
	if !config.CompressBackups {
		compression = "none"
		filename = fmt.Sprintf("auto-backup-%s.tar", time.Now().Format("2006-01-02-15-04-05"))
	}

	// Crear request
	req := &models.CreateBackupRequest{
		ServerID:    serverID,
		Filename:    filename,
		BackupType:  config.BackupType,
		Compression: compression,
	}

	// Ejecutar backup
	backup, err := s.backupService.CreateBackup(ctx, req, uuid.Nil)
	if err != nil {
		s.logger.Error("Error creating scheduled backup",
			zap.String("server_id", serverID.String()),
			zap.Error(err),
		)
		return
	}

	s.logger.Info("Scheduled backup created successfully",
		zap.String("backup_id", backup.ID.String()),
		zap.String("server_id", serverID.String()),
	)

	// Actualizar next_backup_at
	go s.updateNextBackupTime(serverID, config.Schedule)
}

// updateNextBackupTime calcula y actualiza el próximo tiempo de backup
func (s *Scheduler) updateNextBackupTime(serverID uuid.UUID, schedule string) {
	// Parsear schedule
	cronSchedule, err := cron.ParseStandard(schedule)
	if err != nil {
		s.logger.Error("Error parsing cron schedule",
			zap.String("server_id", serverID.String()),
			zap.Error(err),
		)
		return
	}

	// Calcular próximo run
	nextTime := cronSchedule.Next(time.Now())

	// Actualizar en DB
	if err := s.db.Model(&models.BackupConfig{}).
		Where("server_id = ?", serverID).
		Update("next_backup_at", nextTime).Error; err != nil {
		s.logger.Error("Error updating next_backup_at",
			zap.String("server_id", serverID.String()),
			zap.Error(err),
		)
	}
}

// GetScheduledServers retorna la lista de servidores programados
func (s *Scheduler) GetScheduledServers() []uuid.UUID {
	servers := make([]uuid.UUID, 0, len(s.scheduledJobs))
	for serverID := range s.scheduledJobs {
		servers = append(servers, serverID)
	}
	return servers
}

// IsServerScheduled verifica si un servidor está programado
func (s *Scheduler) IsServerScheduled(serverID uuid.UUID) bool {
	_, exists := s.scheduledJobs[serverID]
	return exists
}

// GetJobCount retorna el número de jobs programados
func (s *Scheduler) GetJobCount() int {
	return len(s.scheduledJobs)
}

// RefreshSchedules recarga todas las configuraciones desde la DB
func (s *Scheduler) RefreshSchedules() error {
	s.logger.Info("Refreshing backup schedules")

	// Eliminar todos los schedules actuales
	for serverID := range s.scheduledJobs {
		s.UnscheduleServer(serverID)
	}

	// Recargar desde DB
	var configs []models.BackupConfig
	if err := s.db.Where("enabled = ? AND auto_backup = ?", true, true).Find(&configs).Error; err != nil {
		return fmt.Errorf("error loading backup configs: %w", err)
	}

	// Re-programar
	for _, config := range configs {
		if err := s.ScheduleServer(config.ServerID, config.Schedule); err != nil {
			s.logger.Error("Error scheduling server",
				zap.String("server_id", config.ServerID.String()),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("Schedules refreshed", zap.Int("scheduled_servers", len(s.scheduledJobs)))
	return nil
}

// RunManualBackup ejecuta un backup manual inmediatamente
func (s *Scheduler) RunManualBackup(serverID uuid.UUID, userID uuid.UUID) (*models.Backup, error) {
	s.logger.Info("Running manual backup",
		zap.String("server_id", serverID.String()),
		zap.String("user_id", userID.String()),
	)

	ctx := context.Background()

	// Obtener configuración
	config, err := s.backupService.GetBackupConfig(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("error getting backup config: %w", err)
	}

	// Crear nombre de archivo
	filename := fmt.Sprintf("manual-backup-%s.tar.gz", time.Now().Format("2006-01-02-15-04-05"))

	// Crear request
	req := &models.CreateBackupRequest{
		ServerID:    serverID,
		Filename:    filename,
		BackupType:  config.BackupType,
		Compression: "gzip",
	}

	// Ejecutar backup
	return s.backupService.CreateBackup(ctx, req, userID)
}
