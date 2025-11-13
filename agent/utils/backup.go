package utils

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CreateTarGzBackup crea un archivo tar.gz del directorio especificado
// includePaths: si no es nil, solo incluye estos paths. Si es nil, incluye todo
// excludePaths: paths a excluir del backup
// compress: si es true, usa compresión gzip
func CreateTarGzBackup(sourceDir, destFile string, includePaths map[string]bool, excludePaths []string, compress bool) (int64, string, error) {
	// Crear archivo de destino
	file, err := os.Create(destFile)
	if err != nil {
		return 0, "", fmt.Errorf("error creando archivo: %w", err)
	}
	defer file.Close()

	// Crear hasher para checksum
	hasher := sha256.New()
	writer := io.MultiWriter(file, hasher)

	// Crear writer con o sin compresión
	var tarWriter *tar.Writer
	if compress {
		gzipWriter := gzip.NewWriter(writer)
		defer gzipWriter.Close()
		tarWriter = tar.NewWriter(gzipWriter)
	} else {
		tarWriter = tar.NewWriter(writer)
	}
	defer tarWriter.Close()

	// Convertir excludePaths a map para búsqueda rápida
	excludeMap := make(map[string]bool)
	for _, path := range excludePaths {
		excludeMap[path] = true
	}

	// Recorrer el directorio
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Obtener path relativo
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Saltar el directorio raíz
		if relPath == "." {
			return nil
		}

		// Verificar si está excluido
		if shouldExclude(relPath, excludeMap) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Si hay includePaths, verificar si está incluido
		if includePaths != nil && !shouldInclude(relPath, includePaths) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Crear header del tar
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		// Usar path relativo en el tar
		header.Name = relPath

		// Escribir header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Si es un archivo regular, escribir el contenido
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return 0, "", fmt.Errorf("error recorriendo directorio: %w", err)
	}

	// Obtener tamaño del archivo
	stat, err := os.Stat(destFile)
	if err != nil {
		return 0, "", fmt.Errorf("error obteniendo tamaño: %w", err)
	}

	// Calcular checksum
	checksum := hex.EncodeToString(hasher.Sum(nil))

	return stat.Size(), checksum, nil
}

// ExtractTarGzBackup extrae un archivo tar.gz
// restorePaths: si no es nil, solo restaura estos paths. Si es nil, restaura todo
func ExtractTarGzBackup(srcFile, destDir string, restorePaths map[string]bool) error {
	// Abrir archivo
	file, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("error abriendo archivo: %w", err)
	}
	defer file.Close()

	// Determinar si está comprimido
	var tarReader *tar.Reader
	if strings.HasSuffix(srcFile, ".gz") || strings.HasSuffix(srcFile, ".gzip") {
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("error creando gzip reader: %w", err)
		}
		defer gzipReader.Close()
		tarReader = tar.NewReader(gzipReader)
	} else {
		tarReader = tar.NewReader(file)
	}

	// Extraer archivos
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error leyendo tar: %w", err)
		}

		// Si hay restorePaths, verificar si está incluido
		if restorePaths != nil && !shouldInclude(header.Name, restorePaths) {
			continue
		}

		// Crear path completo
		targetPath := filepath.Join(destDir, header.Name)

		// Verificar que el path está dentro del destDir (seguridad)
		if !strings.HasPrefix(targetPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("intento de path traversal detectado: %s", header.Name)
		}

		// Crear directorios si es necesario
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("error creando directorio: %w", err)
			}
			continue
		}

		// Crear directorio padre si no existe
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("error creando directorio padre: %w", err)
		}

		// Crear archivo
		outFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("error creando archivo: %w", err)
		}

		// Copiar contenido
		if _, err := io.Copy(outFile, tarReader); err != nil {
			outFile.Close()
			return fmt.Errorf("error escribiendo archivo: %w", err)
		}
		outFile.Close()

		// Establecer permisos
		if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
			return fmt.Errorf("error estableciendo permisos: %w", err)
		}
	}

	return nil
}

// shouldExclude verifica si un path debe ser excluido
func shouldExclude(path string, excludeMap map[string]bool) bool {
	// Verificar exacto
	if excludeMap[path] {
		return true
	}

	// Verificar si algún padre está excluido
	parts := strings.Split(path, string(os.PathSeparator))
	for i := 1; i < len(parts); i++ {
		partial := strings.Join(parts[:i], string(os.PathSeparator))
		if excludeMap[partial] {
			return true
		}
	}

	return false
}

// shouldInclude verifica si un path debe ser incluido
func shouldInclude(path string, includeMap map[string]bool) bool {
	// Verificar exacto
	if includeMap[path] {
		return true
	}

	// Verificar si es hijo de algún incluido
	parts := strings.Split(path, string(os.PathSeparator))
	for i := 1; i <= len(parts); i++ {
		partial := strings.Join(parts[:i], string(os.PathSeparator))
		if includeMap[partial] {
			return true
		}
	}

	return false
}
