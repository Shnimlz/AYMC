package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/aymc/backend/database/models"
	"go.uber.org/zap"
)

const (
	ModrinthAPIBase = "https://api.modrinth.com/v2"
	ModrinthTimeout = 10 * time.Second
)

// ModrinthClient cliente para la API de Modrinth
type ModrinthClient struct {
	httpClient *http.Client
	logger     *zap.Logger
}

// NewModrinthClient crea un nuevo cliente de Modrinth
func NewModrinthClient(logger *zap.Logger) *ModrinthClient {
	return &ModrinthClient{
		httpClient: &http.Client{
			Timeout: ModrinthTimeout,
		},
		logger: logger.With(zap.String("client", "modrinth")),
	}
}

// modrinthSearchResponse respuesta de búsqueda de Modrinth
type modrinthSearchResponse struct {
	Hits       []modrinthProject `json:"hits"`
	Offset     int               `json:"offset"`
	Limit      int               `json:"limit"`
	TotalHits  int               `json:"total_hits"`
}

// modrinthProject proyecto de Modrinth
type modrinthProject struct {
	ProjectID    string   `json:"project_id"`
	ProjectType  string   `json:"project_type"`
	Slug         string   `json:"slug"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Categories   []string `json:"categories"`
	DisplayCategories []string `json:"display_categories"`
	ClientSide   string   `json:"client_side"`
	ServerSide   string   `json:"server_side"`
	Downloads    int64    `json:"downloads"`
	Follows      int32    `json:"follows"`
	IconURL      string   `json:"icon_url"`
	Author       string   `json:"author"`
	Versions     []string `json:"versions"` // Minecraft versions
	DateCreated  string   `json:"date_created"`
	DateModified string   `json:"date_modified"`
	LatestVersion string  `json:"latest_version"`
	License      string   `json:"license"`
	Gallery      []string `json:"gallery"`
}

// modrinthVersion versión de un proyecto
type modrinthVersion struct {
	ID              string   `json:"id"`
	ProjectID       string   `json:"project_id"`
	AuthorID        string   `json:"author_id"`
	Featured        bool     `json:"featured"`
	Name            string   `json:"name"`
	VersionNumber   string   `json:"version_number"`
	Changelog       string   `json:"changelog"`
	DatePublished   string   `json:"date_published"`
	Downloads       int32    `json:"downloads"`
	VersionType     string   `json:"version_type"` // release, beta, alpha
	Files           []modrinthFile `json:"files"`
	Dependencies    []modrinthDependency `json:"dependencies"`
	GameVersions    []string `json:"game_versions"`
	Loaders         []string `json:"loaders"` // bukkit, spigot, paper, etc.
}

// modrinthFile archivo de una versión
type modrinthFile struct {
	Hashes struct {
		SHA1   string `json:"sha1"`
		SHA512 string `json:"sha512"`
	} `json:"hashes"`
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Primary  bool   `json:"primary"`
	Size     int64  `json:"size"`
	FileType string `json:"file_type"`
}

// modrinthDependency dependencia de una versión
type modrinthDependency struct {
	VersionID      string `json:"version_id"`
	ProjectID      string `json:"project_id"`
	FileName       string `json:"file_name"`
	DependencyType string `json:"dependency_type"` // required, optional, incompatible
}

// Search busca plugins en Modrinth
func (c *ModrinthClient) Search(ctx context.Context, query string, limit int, offset int) ([]models.PluginSearchResult, int, error) {
	c.logger.Debug("Searching Modrinth",
		zap.String("query", query),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	// Construir URL con filtros
	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("offset", fmt.Sprintf("%d", offset))
	params.Set("facets", `[["project_type:plugin"]]`) // Solo plugins para servidores

	searchURL := fmt.Sprintf("%s/search?%s", ModrinthAPIBase, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "AYMC-Backend/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("modrinth API returned status %d", resp.StatusCode)
	}

	var searchResp modrinthSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convertir a PluginSearchResult
	results := make([]models.PluginSearchResult, 0, len(searchResp.Hits))
	for _, hit := range searchResp.Hits {
		category := ""
		if len(hit.Categories) > 0 {
			category = hit.Categories[0]
		}

		results = append(results, models.PluginSearchResult{
			ID:                hit.ProjectID,
			Name:              hit.Title,
			Slug:              hit.Slug,
			Description:       hit.Description,
			Author:            hit.Author,
			IconURL:           hit.IconURL,
			Source:            "modrinth",
			SourceID:          hit.ProjectID,
			Category:          category,
			Downloads:         hit.Downloads,
			LatestVersion:     hit.LatestVersion,
			MinecraftVersions: hit.Versions,
			UpdatedAt:         hit.DateModified,
		})
	}

	c.logger.Info("Modrinth search completed",
		zap.Int("results", len(results)),
		zap.Int("total", searchResp.TotalHits),
	)

	return results, searchResp.TotalHits, nil
}

// GetProject obtiene los detalles de un proyecto específico
func (c *ModrinthClient) GetProject(ctx context.Context, projectID string) (*models.PluginSearchResult, error) {
	c.logger.Debug("Getting Modrinth project", zap.String("project_id", projectID))

	projectURL := fmt.Sprintf("%s/project/%s", ModrinthAPIBase, projectID)

	req, err := http.NewRequestWithContext(ctx, "GET", projectURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "AYMC-Backend/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("project not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("modrinth API returned status %d", resp.StatusCode)
	}

	var project modrinthProject
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	category := ""
	if len(project.Categories) > 0 {
		category = project.Categories[0]
	}

	result := &models.PluginSearchResult{
		ID:                project.ProjectID,
		Name:              project.Title,
		Slug:              project.Slug,
		Description:       project.Description,
		Author:            project.Author,
		IconURL:           project.IconURL,
		Source:            "modrinth",
		SourceID:          project.ProjectID,
		Category:          category,
		Downloads:         project.Downloads,
		LatestVersion:     project.LatestVersion,
		MinecraftVersions: project.Versions,
		UpdatedAt:         project.DateModified,
	}

	c.logger.Info("Modrinth project fetched", zap.String("name", project.Title))

	return result, nil
}

// GetVersions obtiene las versiones de un proyecto
func (c *ModrinthClient) GetVersions(ctx context.Context, projectID string, minecraftVersion string) ([]models.PluginVersion, error) {
	c.logger.Debug("Getting Modrinth project versions",
		zap.String("project_id", projectID),
		zap.String("minecraft_version", minecraftVersion),
	)

	versionsURL := fmt.Sprintf("%s/project/%s/version", ModrinthAPIBase, projectID)

	// Agregar filtro de versión de Minecraft si se proporciona
	if minecraftVersion != "" {
		params := url.Values{}
		params.Set("game_versions", fmt.Sprintf(`["%s"]`, minecraftVersion))
		versionsURL = fmt.Sprintf("%s?%s", versionsURL, params.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, "GET", versionsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "AYMC-Backend/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("modrinth API returned status %d", resp.StatusCode)
	}

	var versions []modrinthVersion
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convertir a PluginVersion
	results := make([]models.PluginVersion, 0, len(versions))
	for _, ver := range versions {
		// Obtener el archivo principal
		var primaryFile modrinthFile
		for _, file := range ver.Files {
			if file.Primary {
				primaryFile = file
				break
			}
		}

		// Si no hay archivo principal, tomar el primero
		if primaryFile.URL == "" && len(ver.Files) > 0 {
			primaryFile = ver.Files[0]
		}

		// Extraer dependencias
		deps := make([]string, 0, len(ver.Dependencies))
		for _, dep := range ver.Dependencies {
			if dep.DependencyType == "required" {
				deps = append(deps, dep.ProjectID)
			}
		}

		results = append(results, models.PluginVersion{
			ID:                ver.ID,
			PluginID:          ver.ProjectID,
			VersionNumber:     ver.VersionNumber,
			VersionName:       ver.Name,
			Changelog:         ver.Changelog,
			DownloadURL:       primaryFile.URL,
			FileName:          primaryFile.Filename,
			FileSize:          primaryFile.Size,
			FileHash:          primaryFile.Hashes.SHA512,
			MinecraftVersions: ver.GameVersions,
			Dependencies:      deps,
			Downloads:         ver.Downloads,
			ReleaseDate:       ver.DatePublished,
			IsStable:          ver.VersionType == "release",
		})
	}

	c.logger.Info("Modrinth versions fetched",
		zap.Int("count", len(results)),
	)

	return results, nil
}

// GetLatestVersion obtiene la última versión compatible
func (c *ModrinthClient) GetLatestVersion(ctx context.Context, projectID string, minecraftVersion string) (*models.PluginVersion, error) {
	versions, err := c.GetVersions(ctx, projectID, minecraftVersion)
	if err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found")
	}

	// La API de Modrinth ya devuelve las versiones ordenadas por fecha de publicación
	// Buscar la primera versión estable
	for _, ver := range versions {
		if ver.IsStable {
			return &ver, nil
		}
	}

	// Si no hay versión estable, devolver la primera (más reciente)
	return &versions[0], nil
}
