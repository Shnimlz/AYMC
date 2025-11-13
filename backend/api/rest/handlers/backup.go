package handlers

import (
	"net/http"
	"strconv"

	"github.com/aymc/backend/database/models"
	"github.com/aymc/backend/services/backup"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BackupHandler maneja las peticiones HTTP relacionadas con backups
type BackupHandler struct {
	backupService *backup.Service
	scheduler     *backup.Scheduler
	validator     *validator.Validate
	logger        *zap.Logger
}

// NewBackupHandler crea un nuevo handler de backups
func NewBackupHandler(backupService *backup.Service, scheduler *backup.Scheduler, logger *zap.Logger) *BackupHandler {
	return &BackupHandler{
		backupService: backupService,
		scheduler:     scheduler,
		validator:     validator.New(),
		logger:        logger,
	}
}

// CreateBackup crea un nuevo backup
// POST /api/v1/servers/:server_id/backups
func (h *BackupHandler) CreateBackup(c *gin.Context) {
	serverIDStr := c.Param("server_id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de servidor inválido"})
		return
	}

	var req models.CreateBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ServerID = serverID

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Obtener usuario del contexto
	userID := getUserIDFromContext(c)

	backup, err := h.backupService.CreateBackup(c.Request.Context(), &req, userID)
	if err != nil {
		h.logger.Error("Error creating backup", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, backup)
}

// ListBackups lista los backups de un servidor
// GET /api/v1/servers/:server_id/backups
func (h *BackupHandler) ListBackups(c *gin.Context) {
	serverIDStr := c.Param("server_id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de servidor inválido"})
		return
	}

	// Paginación
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	response, err := h.backupService.ListBackups(c.Request.Context(), serverID, limit, offset)
	if err != nil {
		h.logger.Error("Error listing backups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetBackup obtiene un backup específico
// GET /api/v1/backups/:backup_id
func (h *BackupHandler) GetBackup(c *gin.Context) {
	backupIDStr := c.Param("backup_id")
	backupID, err := uuid.Parse(backupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de backup inválido"})
		return
	}

	backup, err := h.backupService.GetBackup(c.Request.Context(), backupID)
	if err != nil {
		h.logger.Error("Error getting backup", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup no encontrado"})
		return
	}

	c.JSON(http.StatusOK, backup)
}

// DeleteBackup elimina un backup
// DELETE /api/v1/backups/:backup_id
func (h *BackupHandler) DeleteBackup(c *gin.Context) {
	backupIDStr := c.Param("backup_id")
	backupID, err := uuid.Parse(backupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de backup inválido"})
		return
	}

	if err := h.backupService.DeleteBackup(c.Request.Context(), backupID); err != nil {
		h.logger.Error("Error deleting backup", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Backup eliminado exitosamente"})
}

// RestoreBackup restaura un backup
// POST /api/v1/backups/:backup_id/restore
func (h *BackupHandler) RestoreBackup(c *gin.Context) {
	backupIDStr := c.Param("backup_id")
	backupID, err := uuid.Parse(backupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de backup inválido"})
		return
	}

	var req models.RestoreBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.BackupID = backupID

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.backupService.RestoreBackup(c.Request.Context(), &req); err != nil {
		h.logger.Error("Error restoring backup", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Restauración de backup iniciada"})
}

// GetBackupConfig obtiene la configuración de backups de un servidor
// GET /api/v1/servers/:server_id/backup-config
func (h *BackupHandler) GetBackupConfig(c *gin.Context) {
	serverIDStr := c.Param("server_id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de servidor inválido"})
		return
	}

	config, err := h.backupService.GetBackupConfig(c.Request.Context(), serverID)
	if err != nil {
		h.logger.Error("Error getting backup config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateBackupConfig actualiza la configuración de backups
// PUT /api/v1/servers/:server_id/backup-config
func (h *BackupHandler) UpdateBackupConfig(c *gin.Context) {
	serverIDStr := c.Param("server_id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de servidor inválido"})
		return
	}

	var req models.UpdateBackupConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.backupService.UpdateBackupConfig(c.Request.Context(), serverID, &req)
	if err != nil {
		h.logger.Error("Error updating backup config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Si se habilitó/deshabilitó auto backup, actualizar scheduler
	if req.AutoBackup != nil {
		if *req.AutoBackup && config.Enabled {
			h.scheduler.ScheduleServer(serverID, config.Schedule)
		} else {
			h.scheduler.UnscheduleServer(serverID)
		}
	}

	// Si cambió el schedule, reprogramar
	if req.Schedule != "" && config.AutoBackup && config.Enabled {
		h.scheduler.RescheduleServer(serverID, req.Schedule)
	}

	c.JSON(http.StatusOK, config)
}

// GetBackupStats obtiene estadísticas de backups
// GET /api/v1/servers/:server_id/backup-stats
func (h *BackupHandler) GetBackupStats(c *gin.Context) {
	serverIDStr := c.Param("server_id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de servidor inválido"})
		return
	}

	stats, err := h.backupService.GetBackupStats(c.Request.Context(), serverID)
	if err != nil {
		h.logger.Error("Error getting backup stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// RunManualBackup ejecuta un backup manual inmediatamente
// POST /api/v1/servers/:server_id/backups/manual
func (h *BackupHandler) RunManualBackup(c *gin.Context) {
	serverIDStr := c.Param("server_id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de servidor inválido"})
		return
	}

	userID := getUserIDFromContext(c)

	backup, err := h.scheduler.RunManualBackup(serverID, userID)
	if err != nil {
		h.logger.Error("Error running manual backup", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, backup)
}

// getUserIDFromContext extrae el ID del usuario del contexto
func getUserIDFromContext(c *gin.Context) uuid.UUID {
	// TODO: Implementar extracción real del JWT
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil
	}

	if userID, ok := userIDStr.(uuid.UUID); ok {
		return userID
	}

	if userIDStr, ok := userIDStr.(string); ok {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			return userID
		}
	}

	return uuid.Nil
}
