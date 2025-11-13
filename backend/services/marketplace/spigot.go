package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aymc/backend/database/models"
	"go.uber.org/zap"
)

const (
	SpigetAPIBase = "https://api.spiget.org/v2"
	SpigetTimeout = 10 * time.Second
)

// SpigotClient cliente para la API de Spiget (Spigot)
type SpigotClient struct {
	httpClient *http.Client
	logger     *zap.Logger
}

// NewSpigotClient crea un nuevo cliente de Spigot
func NewSpigotClient(logger *zap.Logger) *SpigotClient {
	return &SpigotClient{
		httpClient: &http.Client{
			Timeout: SpigetTimeout,
		},
		logger: logger.With(zap.String("client", "spigot")),
	}
}

// spigetResource recurso de Spiget
type spigetResource struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Tag         string `json:"tag"` // Short description
	Contributors string `json:"contributors"`
	Likes       int32  `json:"likes"`
	File        struct {
		Type        string `json:"type"`
		Size        int64  `json:"size"`
		SizeUnit    string `json:"sizeUnit"`
		URL         string `json:"url"`
	} `json:"file"`
	TestedVersions []string `json:"testedVersions"`
	Rating         struct {
		Count   int32   `json:"count"`
		Average float64 `json:"average"`
	} `json:"rating"`
	Icon struct {
		URL  string `json:"url"`
		Data string `json:"data"`
	} `json:"icon"`
	Premium    bool   `json:"premium"`
	Price      float64 `json:"price"`
	Currency   string `json:"currency"`
	Downloads  int64  `json:"downloads"`
	UpdateDate int64  `json:"updateDate"` // Unix timestamp
	ReleaseDate int64 `json:"releaseDate"` // Unix timestamp
	External   bool   `json:"external"`
	Author     struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"author"`
	Category   struct {
		ID   int32  `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
	Version struct {
		ID   int64  `json:"id"`
		Name string `json:"name"` // Version number
	} `json:"version"`
}

// spigetVersion versión de un recurso
type spigetVersion struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"` // Version number
	ReleaseDate int64  `json:"releaseDate"` // Unix timestamp
	Downloads   int32  `json:"downloads"`
	Rating      struct {
		Count   int32   `json:"count"`
		Average float64 `json:"average"`
	} `json:"rating"`
	Size     int64 `json:"size"`
	SizeUnit string `json:"sizeUnit"`
}

// spigetAuthor autor de un recurso
type spigetAuthor struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Search busca recursos en Spigot
func (c *SpigotClient) Search(ctx context.Context, query string, limit int, offset int) ([]models.PluginSearchResult, int, error) {
	c.logger.Debug("Searching Spigot",
		zap.String("query", query),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	// Spiget usa paginación basada en páginas
	page := offset / limit
	if offset%limit != 0 {
		page++
	}

	// Construir URL de búsqueda
	searchURL := fmt.Sprintf("%s/search/resources/%s", SpigetAPIBase, url.QueryEscape(query))

	// Agregar parámetros de paginación
	params := url.Values{}
	params.Set("size", strconv.Itoa(limit))
	params.Set("page", strconv.Itoa(page))
	params.Set("sort", "-downloads") // Ordenar por descargas (descendente)
	params.Set("fields", "id,name,tag,author,rating,downloads,likes,testedVersions,updateDate,icon,version,category,premium,external")

	fullURL := fmt.Sprintf("%s?%s", searchURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
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
		return nil, 0, fmt.Errorf("spiget API returned status %d", resp.StatusCode)
	}

	var resources []spigetResource
	if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convertir a PluginSearchResult
	results := make([]models.PluginSearchResult, 0, len(resources))
	for _, res := range resources {
		// Saltar recursos premium o externos
		if res.Premium || res.External {
			continue
		}

		// Construir URL del icono
		iconURL := ""
		if res.Icon.URL != "" {
			iconURL = fmt.Sprintf("https://www.spigotmc.org/%s", res.Icon.URL)
		}

		// Formatear fecha de actualización
		updatedAt := time.Unix(res.UpdateDate, 0).Format(time.RFC3339)

		results = append(results, models.PluginSearchResult{
			ID:                strconv.FormatInt(res.ID, 10),
			Name:              res.Name,
			Slug:              strings.ToLower(strings.ReplaceAll(res.Name, " ", "-")),
			Description:       res.Tag,
			Summary:           res.Tag,
			Author:            res.Author.Name,
			IconURL:           iconURL,
			Source:            "spigot",
			SourceID:          strconv.FormatInt(res.ID, 10),
			Category:          res.Category.Name,
			Downloads:         res.Downloads,
			Rating:            float32(res.Rating.Average),
			LatestVersion:     res.Version.Name,
			MinecraftVersions: res.TestedVersions,
			UpdatedAt:         updatedAt,
		})
	}

	// Spiget no devuelve el total de resultados, estimamos basado en los resultados
	total := len(results)
	if len(results) == limit {
		total = offset + limit + 1 // Hay más resultados
	}

	c.logger.Info("Spigot search completed",
		zap.Int("results", len(results)),
	)

	return results, total, nil
}

// GetResource obtiene los detalles de un recurso específico
func (c *SpigotClient) GetResource(ctx context.Context, resourceID string) (*models.PluginSearchResult, error) {
	c.logger.Debug("Getting Spigot resource", zap.String("resource_id", resourceID))

	resourceURL := fmt.Sprintf("%s/resources/%s", SpigetAPIBase, resourceID)

	// Solicitar campos específicos
	params := url.Values{}
	params.Set("fields", "id,name,tag,author,rating,downloads,likes,testedVersions,updateDate,releaseDate,icon,version,category,premium,external,file")

	fullURL := fmt.Sprintf("%s?%s", resourceURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
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
		return nil, fmt.Errorf("resource not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spiget API returned status %d", resp.StatusCode)
	}

	var resource spigetResource
	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Saltar recursos premium o externos
	if resource.Premium {
		return nil, fmt.Errorf("premium resources are not supported")
	}
	if resource.External {
		return nil, fmt.Errorf("external resources are not supported")
	}

	// Construir URL del icono
	iconURL := ""
	if resource.Icon.URL != "" {
		iconURL = fmt.Sprintf("https://www.spigotmc.org/%s", resource.Icon.URL)
	}

	// Formatear fecha de actualización
	updatedAt := time.Unix(resource.UpdateDate, 0).Format(time.RFC3339)

	result := &models.PluginSearchResult{
		ID:                strconv.FormatInt(resource.ID, 10),
		Name:              resource.Name,
		Slug:              strings.ToLower(strings.ReplaceAll(resource.Name, " ", "-")),
		Description:       resource.Tag,
		Summary:           resource.Tag,
		Author:            resource.Author.Name,
		IconURL:           iconURL,
		Source:            "spigot",
		SourceID:          strconv.FormatInt(resource.ID, 10),
		Category:          resource.Category.Name,
		Downloads:         resource.Downloads,
		Rating:            float32(resource.Rating.Average),
		LatestVersion:     resource.Version.Name,
		MinecraftVersions: resource.TestedVersions,
		UpdatedAt:         updatedAt,
	}

	c.logger.Info("Spigot resource fetched", zap.String("name", resource.Name))

	return result, nil
}

// GetVersions obtiene las versiones de un recurso
func (c *SpigotClient) GetVersions(ctx context.Context, resourceID string) ([]models.PluginVersion, error) {
	c.logger.Debug("Getting Spigot resource versions", zap.String("resource_id", resourceID))

	versionsURL := fmt.Sprintf("%s/resources/%s/versions", SpigetAPIBase, resourceID)

	// Solicitar hasta 100 versiones
	params := url.Values{}
	params.Set("size", "100")
	params.Set("sort", "-releaseDate")
	params.Set("fields", "id,name,releaseDate,downloads,rating,size,sizeUnit")

	fullURL := fmt.Sprintf("%s?%s", versionsURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
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
		return nil, fmt.Errorf("spiget API returned status %d", resp.StatusCode)
	}

	var versions []spigetVersion
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Obtener información del recurso para las versiones de Minecraft soportadas
	resource, err := c.GetResource(ctx, resourceID)
	if err != nil {
		c.logger.Warn("Failed to get resource details for versions", zap.Error(err))
	}

	// Convertir a PluginVersion
	results := make([]models.PluginVersion, 0, len(versions))
	for _, ver := range versions {
		// Calcular tamaño en bytes
		fileSize := ver.Size
		switch ver.SizeUnit {
		case "KB":
			fileSize *= 1024
		case "MB":
			fileSize *= 1024 * 1024
		case "GB":
			fileSize *= 1024 * 1024 * 1024
		}

		// Construir URL de descarga
		downloadURL := fmt.Sprintf("%s/resources/%s/versions/%d/download", SpigetAPIBase, resourceID, ver.ID)

		// Usar versiones de Minecraft del recurso si están disponibles
		minecraftVersions := []string{}
		if resource != nil {
			minecraftVersions = resource.MinecraftVersions
		}

		results = append(results, models.PluginVersion{
			ID:                strconv.FormatInt(ver.ID, 10),
			PluginID:          resourceID,
			VersionNumber:     ver.Name,
			VersionName:       ver.Name,
			DownloadURL:       downloadURL,
			FileName:          fmt.Sprintf("%s-%s.jar", resourceID, ver.Name),
			FileSize:          fileSize,
			MinecraftVersions: minecraftVersions,
			Downloads:         ver.Downloads,
			ReleaseDate:       time.Unix(ver.ReleaseDate, 0).Format(time.RFC3339),
			IsStable:          true, // Spiget no indica tipo de versión, asumimos estable
		})
	}

	c.logger.Info("Spigot versions fetched", zap.Int("count", len(results)))

	return results, nil
}

// GetLatestVersion obtiene la última versión de un recurso
func (c *SpigotClient) GetLatestVersion(ctx context.Context, resourceID string) (*models.PluginVersion, error) {
	versions, err := c.GetVersions(ctx, resourceID)
	if err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found")
	}

	// La API de Spiget ya devuelve las versiones ordenadas por fecha de lanzamiento
	return &versions[0], nil
}

// GetDownloadURL obtiene la URL de descarga de una versión específica
func (c *SpigotClient) GetDownloadURL(ctx context.Context, resourceID string, versionID string) (string, error) {
	// Spiget proporciona URLs de descarga directas
	return fmt.Sprintf("%s/resources/%s/versions/%s/download", SpigetAPIBase, resourceID, versionID), nil
}
