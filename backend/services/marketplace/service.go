package marketplace

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aymc/backend/database/models"
	"github.com/aymc/backend/services/agents"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service servicio de marketplace que unifica múltiples fuentes
type Service struct {
	db            *gorm.DB
	modrinthClient *ModrinthClient
	spigotClient   *SpigotClient
	agentService   *agents.AgentService
	logger        *zap.Logger
}

// NewService crea un nuevo servicio de marketplace
func NewService(db *gorm.DB, agentService *agents.AgentService, logger *zap.Logger) *Service {
	return &Service{
		db:            db,
		modrinthClient: NewModrinthClient(logger),
		spigotClient:   NewSpigotClient(logger),
		agentService:   agentService,
		logger:        logger.With(zap.String("service", "marketplace")),
	}
}

// SearchRequest petición de búsqueda
type SearchRequest struct {
	Query   string   `json:"query"`
	Sources []string `json:"sources,omitempty"` // modrinth, spigot, curseforge
	Limit   int      `json:"limit,omitempty"`
	Offset  int      `json:"offset,omitempty"`
}

// SearchResponse respuesta de búsqueda
type SearchResponse struct {
	Results []models.PluginSearchResult `json:"results"`
	Total   int                        `json:"total"`
	Sources []string                   `json:"sources"`
}

// SearchPlugins busca plugins en múltiples fuentes
func (s *Service) SearchPlugins(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	s.logger.Info("Searching plugins",
		zap.String("query", req.Query),
		zap.Strings("sources", req.Sources),
		zap.Int("limit", req.Limit),
		zap.Int("offset", req.Offset),
	)

	// Valores por defecto
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if len(req.Sources) == 0 {
		req.Sources = []string{"modrinth", "spigot"} // Búsqueda en ambas fuentes por defecto
	}

	// Buscar en paralelo en todas las fuentes solicitadas
	var wg sync.WaitGroup
	resultsChan := make(chan []models.PluginSearchResult, len(req.Sources))
	errorsChan := make(chan error, len(req.Sources))

	for _, source := range req.Sources {
		wg.Add(1)
		go func(src string) {
			defer wg.Done()

			var results []models.PluginSearchResult
			var err error

			switch src {
			case "modrinth":
				results, _, err = s.modrinthClient.Search(ctx, req.Query, req.Limit, req.Offset)
			case "spigot":
				results, _, err = s.spigotClient.Search(ctx, req.Query, req.Limit, req.Offset)
			default:
				err = fmt.Errorf("unsupported source: %s", src)
			}

			if err != nil {
				s.logger.Warn("Search failed for source",
					zap.String("source", src),
					zap.Error(err),
				)
				errorsChan <- err
				return
			}

			resultsChan <- results
		}(source)
	}

	wg.Wait()
	close(resultsChan)
	close(errorsChan)

	// Combinar resultados
	allResults := make([]models.PluginSearchResult, 0)
	for results := range resultsChan {
		allResults = append(allResults, results...)
	}

	// Eliminar duplicados basados en nombre (case-insensitive)
	seen := make(map[string]bool)
	uniqueResults := make([]models.PluginSearchResult, 0)
	for _, result := range allResults {
		key := strings.ToLower(result.Name)
		if !seen[key] {
			seen[key] = true
			uniqueResults = append(uniqueResults, result)
		}
	}

	// Ordenar por descargas (descendente)
	// Simple bubble sort para mantener el código simple
	for i := 0; i < len(uniqueResults); i++ {
		for j := i + 1; j < len(uniqueResults); j++ {
			if uniqueResults[i].Downloads < uniqueResults[j].Downloads {
				uniqueResults[i], uniqueResults[j] = uniqueResults[j], uniqueResults[i]
			}
		}
	}

	// Aplicar límite si es necesario
	if len(uniqueResults) > req.Limit {
		uniqueResults = uniqueResults[:req.Limit]
	}

	s.logger.Info("Search completed",
		zap.Int("total_results", len(uniqueResults)),
	)

	return &SearchResponse{
		Results: uniqueResults,
		Total:   len(uniqueResults),
		Sources: req.Sources,
	}, nil
}

// GetPlugin obtiene los detalles de un plugin específico
func (s *Service) GetPlugin(ctx context.Context, source string, pluginID string) (*models.PluginSearchResult, error) {
	s.logger.Debug("Getting plugin details",
		zap.String("source", source),
		zap.String("plugin_id", pluginID),
	)

	var result *models.PluginSearchResult
	var err error

	switch source {
	case "modrinth":
		result, err = s.modrinthClient.GetProject(ctx, pluginID)
	case "spigot":
		result, err = s.spigotClient.GetResource(ctx, pluginID)
	default:
		return nil, fmt.Errorf("unsupported source: %s", source)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get plugin: %w", err)
	}

	return result, nil
}

// GetPluginVersions obtiene las versiones disponibles de un plugin
func (s *Service) GetPluginVersions(ctx context.Context, source string, pluginID string, minecraftVersion string) ([]models.PluginVersion, error) {
	s.logger.Debug("Getting plugin versions",
		zap.String("source", source),
		zap.String("plugin_id", pluginID),
		zap.String("minecraft_version", minecraftVersion),
	)

	var versions []models.PluginVersion
	var err error

	switch source {
	case "modrinth":
		versions, err = s.modrinthClient.GetVersions(ctx, pluginID, minecraftVersion)
	case "spigot":
		versions, err = s.spigotClient.GetVersions(ctx, pluginID)
		// Filtrar por versión de Minecraft si se especifica
		if minecraftVersion != "" && err == nil {
			filtered := make([]models.PluginVersion, 0)
			for _, v := range versions {
				for _, mcVer := range v.MinecraftVersions {
					if mcVer == minecraftVersion {
						filtered = append(filtered, v)
						break
					}
				}
			}
			versions = filtered
		}
	default:
		return nil, fmt.Errorf("unsupported source: %s", source)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}

	return versions, nil
}

// InstallPlugin instala un plugin en un servidor
func (s *Service) InstallPlugin(ctx context.Context, serverID uuid.UUID, req models.PluginInstallRequest) error {
	s.logger.Info("Installing plugin",
		zap.String("server_id", serverID.String()),
		zap.String("plugin_name", req.PluginName),
		zap.String("version", req.Version),
		zap.String("source", req.Source),
	)

	// Verificar que el servidor existe
	var server models.Server
	if err := s.db.WithContext(ctx).First(&server, "id = ?", serverID).Error; err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Si no se proporciona download URL, obtenerla de la fuente
	downloadURL := req.DownloadURL
	fileName := req.FileName
	if downloadURL == "" {
		// Obtener la versión específica o la última
		var version *models.PluginVersion
		var err error

		if req.Version != "" {
			// Buscar versión específica
			versions, err := s.GetPluginVersions(ctx, req.Source, req.SourceID, "")
			if err != nil {
				return fmt.Errorf("failed to get versions: %w", err)
			}

			for _, v := range versions {
				if v.VersionNumber == req.Version {
					version = &v
					break
				}
			}

			if version == nil {
				return fmt.Errorf("version %s not found", req.Version)
			}
		} else {
			// Obtener última versión
			switch req.Source {
			case "modrinth":
				version, err = s.modrinthClient.GetLatestVersion(ctx, req.SourceID, "")
			case "spigot":
				version, err = s.spigotClient.GetLatestVersion(ctx, req.SourceID)
			default:
				return fmt.Errorf("unsupported source: %s", req.Source)
			}

			if err != nil {
				return fmt.Errorf("failed to get latest version: %w", err)
			}
		}

		downloadURL = version.DownloadURL
		fileName = version.FileName
	}

	// Instalar plugin a través del agente
	if err := s.agentService.InstallPlugin(ctx, server.AgentID, serverID, req.PluginName, downloadURL, fileName); err != nil {
		return fmt.Errorf("failed to install plugin via agent: %w", err)
	}

	// Registrar plugin en la base de datos
	plugin := models.Plugin{
		Name:        req.PluginName,
		Version:     req.Version,
		DownloadURL: downloadURL,
		Source:      models.PluginSource(req.Source),
		SourceID:    req.SourceID,
		IsActive:    true,
	}

	// Buscar o crear el plugin
	if err := s.db.WithContext(ctx).Where("source = ? AND source_id = ?", req.Source, req.SourceID).FirstOrCreate(&plugin).Error; err != nil {
		s.logger.Warn("Failed to save plugin to database", zap.Error(err))
	}

	// Crear relación server-plugin
	serverPlugin := models.ServerPlugin{
		ServerID:    serverID,
		PluginID:    plugin.ID,
		Version:     req.Version,
		IsEnabled:   true,
		InstalledAt: time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(&serverPlugin).Error; err != nil {
		s.logger.Warn("Failed to save server-plugin relationship", zap.Error(err))
	}

	s.logger.Info("Plugin installed successfully",
		zap.String("plugin_name", req.PluginName),
		zap.String("server_id", serverID.String()),
	)

	return nil
}

// UninstallPlugin desinstala un plugin de un servidor
func (s *Service) UninstallPlugin(ctx context.Context, serverID uuid.UUID, req models.PluginUninstallRequest) error {
	s.logger.Info("Uninstalling plugin",
		zap.String("server_id", serverID.String()),
		zap.String("plugin_name", req.PluginName),
		zap.Bool("delete_config", req.DeleteConfig),
		zap.Bool("delete_data", req.DeleteData),
	)

	// Verificar que el servidor existe
	var server models.Server
	if err := s.db.WithContext(ctx).First(&server, "id = ?", serverID).Error; err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Desinstalar plugin a través del agente
	if err := s.agentService.UninstallPlugin(ctx, server.AgentID, serverID, req.PluginName, req.DeleteConfig, req.DeleteData); err != nil {
		return fmt.Errorf("failed to uninstall plugin via agent: %w", err)
	}

	// Actualizar base de datos
	if err := s.db.WithContext(ctx).
		Model(&models.ServerPlugin{}).
		Where("server_id = ? AND plugin_id IN (SELECT id FROM plugins WHERE name = ?)", serverID, req.PluginName).
		Update("is_enabled", false).Error; err != nil {
		s.logger.Warn("Failed to update server-plugin relationship", zap.Error(err))
	}

	s.logger.Info("Plugin uninstalled successfully",
		zap.String("plugin_name", req.PluginName),
		zap.String("server_id", serverID.String()),
	)

	return nil
}

// UpdatePlugin actualiza un plugin en un servidor
func (s *Service) UpdatePlugin(ctx context.Context, serverID uuid.UUID, req models.PluginUpdateRequest) error {
	s.logger.Info("Updating plugin",
		zap.String("server_id", serverID.String()),
		zap.String("plugin_name", req.PluginName),
		zap.String("version", req.Version),
	)

	// Verificar que el servidor existe
	var server models.Server
	if err := s.db.WithContext(ctx).First(&server, "id = ?", serverID).Error; err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Actualizar plugin a través del agente
	if err := s.agentService.UpdatePlugin(ctx, server.AgentID, serverID, req.PluginName, req.DownloadURL, req.FileName); err != nil {
		return fmt.Errorf("failed to update plugin via agent: %w", err)
	}

	// Actualizar base de datos
	if err := s.db.WithContext(ctx).
		Model(&models.ServerPlugin{}).
		Where("server_id = ? AND plugin_id IN (SELECT id FROM plugins WHERE name = ?)", serverID, req.PluginName).
		Update("version", req.Version).Error; err != nil {
		s.logger.Warn("Failed to update server-plugin version", zap.Error(err))
	}

	s.logger.Info("Plugin updated successfully",
		zap.String("plugin_name", req.PluginName),
		zap.String("version", req.Version),
		zap.String("server_id", serverID.String()),
	)

	return nil
}

// ListInstalledPlugins lista los plugins instalados en un servidor
func (s *Service) ListInstalledPlugins(ctx context.Context, serverID uuid.UUID) (*models.PluginListResponse, error) {
	s.logger.Debug("Listing installed plugins", zap.String("server_id", serverID.String()))

	// Verificar que el servidor existe
	var server models.Server
	if err := s.db.WithContext(ctx).First(&server, "id = ?", serverID).Error; err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	// Obtener plugins del servidor
	var serverPlugins []models.ServerPlugin
	if err := s.db.WithContext(ctx).
		Preload("Plugin").
		Where("server_id = ?", serverID).
		Find(&serverPlugins).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch plugins: %w", err)
	}

	// Convertir a InstalledPlugin
	plugins := make([]models.InstalledPlugin, 0, len(serverPlugins))
	for _, sp := range serverPlugins {
		plugins = append(plugins, models.InstalledPlugin{
			Name:        sp.Plugin.Name,
			Version:     sp.Version,
			IsEnabled:   sp.IsEnabled,
			Description: sp.Plugin.Description,
			Author:      sp.Plugin.Author,
			Source:      string(sp.Plugin.Source),
			InstalledAt: sp.InstalledAt,
		})
	}

	return &models.PluginListResponse{
		Plugins: plugins,
		Total:   len(plugins),
	}, nil
}
