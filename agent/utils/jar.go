package utils

import (
	"archive/zip"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PluginMetadata contiene información del plugin.yml
type PluginMetadata struct {
	Name         string   `yaml:"name"`
	Version      string   `yaml:"version"`
	Main         string   `yaml:"main"`
	Description  string   `yaml:"description"`
	Author       string   `yaml:"author"`
	Authors      []string `yaml:"authors"`
	Website      string   `yaml:"website"`
	APIVersion   string   `yaml:"api-version"`
	Depend       []string `yaml:"depend"`
	SoftDepend   []string `yaml:"softdepend"`
	LoadBefore   []string `yaml:"loadbefore"`
	Prefix       string   `yaml:"prefix"`
	Commands     map[string]interface{} `yaml:"commands"`
	Permissions  map[string]interface{} `yaml:"permissions"`
}

// ReadPluginYml lee el plugin.yml de un archivo JAR
func ReadPluginYml(jarPath string) (*PluginMetadata, error) {
	// Abrir el archivo ZIP/JAR
	reader, err := zip.OpenReader(jarPath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo JAR: %w", err)
	}
	defer reader.Close()

	// Buscar plugin.yml
	for _, file := range reader.File {
		if file.Name == "plugin.yml" || file.Name == "paper-plugin.yml" || file.Name == "bungee.yml" {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("error abriendo %s: %w", file.Name, err)
			}
			defer rc.Close()

			// Parsear YAML
			var metadata PluginMetadata
			decoder := yaml.NewDecoder(rc)
			if err := decoder.Decode(&metadata); err != nil {
				return nil, fmt.Errorf("error parseando %s: %w", file.Name, err)
			}

			return &metadata, nil
		}
	}

	return nil, fmt.Errorf("plugin.yml no encontrado en JAR")
}

// ValidateSHA512 valida el hash SHA512 de un archivo
func ValidateSHA512(filePath string, expectedHash string) (bool, error) {
	if expectedHash == "" {
		return true, nil // No hay hash para validar
	}

	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("error abriendo archivo: %w", err)
	}
	defer file.Close()

	hasher := sha512.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return false, fmt.Errorf("error calculando hash: %w", err)
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	return strings.EqualFold(actualHash, expectedHash), nil
}

// GetFileSize retorna el tamaño de un archivo en bytes
func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// CopyFile copia un archivo de src a dst
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error abriendo origen: %w", err)
	}
	defer sourceFile.Close()

	// Crear directorios si no existen
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("error creando directorios: %w", err)
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creando destino: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("error copiando archivo: %w", err)
	}

	return destFile.Sync()
}

// BackupFile crea un backup de un archivo agregando .backup al nombre
func BackupFile(filePath string) (string, error) {
	backupPath := filePath + ".backup"
	
	if err := CopyFile(filePath, backupPath); err != nil {
		return "", fmt.Errorf("error creando backup: %w", err)
	}
	
	return backupPath, nil
}

// DownloadFile descarga un archivo desde una URL
func DownloadFile(url, destPath string) error {
	// Esta función se implementaría con http.Client
	// Por ahora dejamos que sea el agente quien maneje la descarga
	return fmt.Errorf("DownloadFile: not implemented - use agent's download system")
}

// RemoveDirectory elimina un directorio recursivamente
func RemoveDirectory(path string) error {
	return os.RemoveAll(path)
}

// IsJarFile verifica si un archivo es un JAR válido
func IsJarFile(filePath string) bool {
	if !strings.HasSuffix(strings.ToLower(filePath), ".jar") {
		return false
	}

	// Intentar abrir como ZIP
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return false
	}
	defer reader.Close()

	return true
}

// GetPluginFileName normaliza el nombre de archivo de un plugin
func GetPluginFileName(pluginName string, version string) string {
	// Normalizar nombre (remover espacios, caracteres especiales)
	name := strings.ReplaceAll(pluginName, " ", "-")
	name = strings.ToLower(name)
	
	if version != "" {
		return fmt.Sprintf("%s-%s.jar", name, version)
	}
	return fmt.Sprintf("%s.jar", name)
}

// ListJarFiles lista todos los archivos .jar en un directorio
func ListJarFiles(dir string) ([]string, error) {
	var jars []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error leyendo directorio: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(strings.ToLower(entry.Name()), ".jar") {
			jars = append(jars, filepath.Join(dir, entry.Name()))
		}
	}

	return jars, nil
}
