package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ServerDownloader maneja la descarga de JARs de servidores
type ServerDownloader struct {
	serverType string
	version    string
	outputDir  string
	client     *http.Client
}

// DownloadProgress representa el progreso de una descarga
type DownloadProgress struct {
	Downloaded int64
	Total      int64
	Percent    float64
	Speed      int64 // bytes por segundo
	Message    string
}

// ProgressCallback es llamado durante la descarga
type ProgressCallback func(progress DownloadProgress)

// NewServerDownloader crea un nuevo descargador de servidores
func NewServerDownloader(serverType, version, outputDir string) *ServerDownloader {
	return &ServerDownloader{
		serverType: serverType,
		version:    version,
		outputDir:  outputDir,
		client: &http.Client{
			Timeout: 30 * time.Minute,
		},
	}
}

// PaperMCVersion representa una versión de PaperMC
type PaperMCVersion struct {
	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	Version     string `json:"version"`
	Builds      []int  `json:"builds"`
}

// PaperMCBuilds representa los builds de una versión
type PaperMCBuilds struct {
	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	Version     string `json:"version"`
	Builds      []struct {
		Build     int       `json:"build"`
		Time      time.Time `json:"time"`
		Downloads struct {
			Application struct {
				Name   string `json:"name"`
				SHA256 string `json:"sha256"`
			} `json:"application"`
		} `json:"downloads"`
	} `json:"builds"`
}

// PurpurBuilds representa los builds de Purpur
type PurpurBuilds struct {
	Builds struct {
		All    []string `json:"all"`
		Latest string   `json:"latest"`
	} `json:"builds"`
}

// GetDownloadURL obtiene la URL de descarga según el tipo de servidor
func (sd *ServerDownloader) GetDownloadURL() (string, string, error) {
	switch sd.serverType {
	case "paper":
		return sd.getPaperURL()
	case "purpur":
		return sd.getPurpurURL()
	case "spigot":
		return "", "", fmt.Errorf("spigot requiere compilación con BuildTools")
	case "vanilla":
		return sd.getVanillaURL()
	default:
		return "", "", fmt.Errorf("tipo de servidor no soportado: %s", sd.serverType)
	}
}

// getPaperURL obtiene la URL de descarga de PaperMC
func (sd *ServerDownloader) getPaperURL() (string, string, error) {
	// API: https://api.papermc.io/v2/projects/paper/versions/{version}/builds
	apiURL := fmt.Sprintf("https://api.papermc.io/v2/projects/paper/versions/%s/builds", sd.version)

	resp, err := sd.client.Get(apiURL)
	if err != nil {
		return "", "", fmt.Errorf("error consultando API de PaperMC: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("API de PaperMC retornó código %d", resp.StatusCode)
	}

	var builds PaperMCBuilds
	if err := json.NewDecoder(resp.Body).Decode(&builds); err != nil {
		return "", "", fmt.Errorf("error decodificando respuesta de PaperMC: %w", err)
	}

	if len(builds.Builds) == 0 {
		return "", "", fmt.Errorf("no se encontraron builds para la versión %s", sd.version)
	}

	// Obtener el build más reciente
	latestBuild := builds.Builds[len(builds.Builds)-1]
	jarName := latestBuild.Downloads.Application.Name
	sha256Hash := latestBuild.Downloads.Application.SHA256

	downloadURL := fmt.Sprintf(
		"https://api.papermc.io/v2/projects/paper/versions/%s/builds/%d/downloads/%s",
		sd.version,
		latestBuild.Build,
		jarName,
	)

	return downloadURL, sha256Hash, nil
}

// getPurpurURL obtiene la URL de descarga de Purpur
func (sd *ServerDownloader) getPurpurURL() (string, string, error) {
	// API: https://api.purpurmc.org/v2/purpur/{version}
	apiURL := fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s", sd.version)

	resp, err := sd.client.Get(apiURL)
	if err != nil {
		return "", "", fmt.Errorf("error consultando API de Purpur: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("API de Purpur retornó código %d", resp.StatusCode)
	}

	var builds PurpurBuilds
	if err := json.NewDecoder(resp.Body).Decode(&builds); err != nil {
		return "", "", fmt.Errorf("error decodificando respuesta de Purpur: %w", err)
	}

	latestBuild := builds.Builds.Latest
	downloadURL := fmt.Sprintf(
		"https://api.purpurmc.org/v2/purpur/%s/%s/download",
		sd.version,
		latestBuild,
	)

	// Purpur no proporciona SHA256 en la API, lo calcularemos después de descargar
	return downloadURL, "", nil
}

// getVanillaURL obtiene la URL de descarga de Minecraft Vanilla
func (sd *ServerDownloader) getVanillaURL() (string, string, error) {
	// Para vanilla necesitaríamos parsear el manifest de Mojang
	// API: https://launchermeta.mojang.com/mc/game/version_manifest.json
	return "", "", fmt.Errorf("descarga de vanilla no implementada aún")
}

// Download descarga el JAR del servidor
func (sd *ServerDownloader) Download(callback ProgressCallback) (string, error) {
	// Obtener URL de descarga
	downloadURL, expectedSHA256, err := sd.GetDownloadURL()
	if err != nil {
		return "", err
	}

	callback(DownloadProgress{
		Message: fmt.Sprintf("Descargando %s %s...", sd.serverType, sd.version),
	})

	// Crear directorio de salida si no existe
	if err := os.MkdirAll(sd.outputDir, 0755); err != nil {
		return "", fmt.Errorf("error creando directorio: %w", err)
	}

	// Nombre del archivo
	outputFile := filepath.Join(sd.outputDir, fmt.Sprintf("%s-%s.jar", sd.serverType, sd.version))

	// Iniciar descarga
	resp, err := sd.client.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("error iniciando descarga: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("servidor retornó código %d", resp.StatusCode)
	}

	// Crear archivo de salida
	out, err := os.Create(outputFile)
	if err != nil {
		return "", fmt.Errorf("error creando archivo: %w", err)
	}
	defer out.Close()

	// Descargar con progreso
	totalSize := resp.ContentLength
	downloaded := int64(0)
	startTime := time.Now()
	lastUpdate := time.Now()

	buffer := make([]byte, 32*1024) // 32KB buffer
	hasher := sha256.New()

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			// Escribir al archivo
			if _, writeErr := out.Write(buffer[:n]); writeErr != nil {
				return "", fmt.Errorf("error escribiendo archivo: %w", writeErr)
			}

			// Actualizar hash
			hasher.Write(buffer[:n])

			downloaded += int64(n)

			// Actualizar progreso cada 100ms
			if time.Since(lastUpdate) > 100*time.Millisecond {
				elapsed := time.Since(startTime).Seconds()
				speed := int64(float64(downloaded) / elapsed)
				percent := float64(downloaded) / float64(totalSize) * 100

				callback(DownloadProgress{
					Downloaded: downloaded,
					Total:      totalSize,
					Percent:    percent,
					Speed:      speed,
					Message:    fmt.Sprintf("Descargando... %.1f%% (%s/s)", percent, formatBytes(speed)),
				})

				lastUpdate = time.Now()
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error durante descarga: %w", err)
		}
	}

	// Descarga completada
	callback(DownloadProgress{
		Downloaded: downloaded,
		Total:      totalSize,
		Percent:    100,
		Message:    "Descarga completada",
	})

	// Verificar SHA256 si está disponible
	if expectedSHA256 != "" {
		actualSHA256 := hex.EncodeToString(hasher.Sum(nil))
		if actualSHA256 != expectedSHA256 {
			os.Remove(outputFile)
			return "", fmt.Errorf("checksum SHA256 no coincide. Esperado: %s, Obtenido: %s", expectedSHA256, actualSHA256)
		}
		callback(DownloadProgress{
			Message: "✅ Checksum SHA256 verificado",
		})
	}

	callback(DownloadProgress{
		Message: fmt.Sprintf("✅ Servidor descargado: %s", outputFile),
	})

	return outputFile, nil
}

// formatBytes formatea bytes en formato legible
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// DownloadWithRetry descarga con reintentos automáticos
func (sd *ServerDownloader) DownloadWithRetry(maxRetries int, callback ProgressCallback) (string, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			callback(DownloadProgress{
				Message: fmt.Sprintf("Reintentando... (intento %d/%d)", attempt, maxRetries),
			})
			time.Sleep(time.Duration(attempt) * time.Second) // Backoff exponencial simple
		}

		filePath, err := sd.Download(callback)
		if err == nil {
			return filePath, nil
		}

		lastErr = err
		callback(DownloadProgress{
			Message: fmt.Sprintf("Error en intento %d: %v", attempt, err),
		})
	}

	return "", fmt.Errorf("descarga falló después de %d intentos: %w", maxRetries, lastErr)
}
