package core

import (
	"testing"
)

func TestNewServerDownloader(t *testing.T) {
	downloader := NewServerDownloader("paper", "1.20.1", "/tmp")
	
	if downloader == nil {
		t.Fatal("ServerDownloader es nil")
	}
	
	if downloader.serverType != "paper" {
		t.Errorf("serverType esperado 'paper', obtenido '%s'", downloader.serverType)
	}
	
	if downloader.version != "1.20.1" {
		t.Errorf("version esperada '1.20.1', obtenida '%s'", downloader.version)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}
	
	for _, tt := range tests {
		result := formatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %s, esperado %s", tt.bytes, result, tt.expected)
		}
	}
}

func TestGetDownloadURL_UnsupportedType(t *testing.T) {
	downloader := NewServerDownloader("unknown", "1.20.1", "/tmp")
	
	_, _, err := downloader.GetDownloadURL()
	if err == nil {
		t.Error("Debería retornar error para tipo de servidor no soportado")
	}
}

func TestGetDownloadURL_Spigot(t *testing.T) {
	downloader := NewServerDownloader("spigot", "1.20.1", "/tmp")
	
	_, _, err := downloader.GetDownloadURL()
	if err == nil {
		t.Error("Spigot debería retornar error indicando que requiere BuildTools")
	}
}

func TestDownload_WithMockServer(t *testing.T) {
	t.Skip("Test requiere refactoring para soportar dependency injection")
	
	// TODO: Refactorizar ServerDownloader para permitir inyección de cliente HTTP
	// y función GetDownloadURL para facilitar testing con mocks
}

func TestDownload_ServerError(t *testing.T) {
	t.Skip("Test requiere refactoring para soportar dependency injection")
}

func TestDownloadWithRetry_Success(t *testing.T) {
	t.Skip("Test requiere refactoring para soportar dependency injection")
}

func TestDownloadWithRetry_AllFail(t *testing.T) {
	t.Skip("Test requiere refactoring para soportar dependency injection")
}

// Test de integración con APIs reales (se salta en modo corto)
func TestGetPaperURL_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	downloader := NewServerDownloader("paper", "1.20.1", "/tmp")
	
	url, sha256, err := downloader.getPaperURL()
	if err != nil {
		t.Fatalf("Error obteniendo URL de Paper: %v", err)
	}
	
	if url == "" {
		t.Error("URL de descarga vacía")
	}
	
	if sha256 == "" {
		t.Error("SHA256 vacío")
	}
	
	t.Logf("URL: %s", url)
	t.Logf("SHA256: %s", sha256)
}

func TestGetPurpurURL_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	downloader := NewServerDownloader("purpur", "1.20.1", "/tmp")
	
	url, _, err := downloader.getPurpurURL()
	if err != nil {
		t.Fatalf("Error obteniendo URL de Purpur: %v", err)
	}
	
	if url == "" {
		t.Error("URL de descarga vacía")
	}
	
	t.Logf("URL: %s", url)
}

func TestDownload_RealPaper_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Solo ejecutar si la variable de entorno está configurada
	// Ejemplo: TEST_REAL_DOWNLOAD=1 go test -run TestDownload_RealPaper_Integration
	// if os.Getenv("TEST_REAL_DOWNLOAD") != "1" {
	// 	t.Skip("TEST_REAL_DOWNLOAD no está configurado")
	// }
	
	t.Skip("Test de descarga real requiere conexión a internet y tiempo")
	
	// Descomentar para test manual:
	// tmpDir := t.TempDir()
	// downloader := NewServerDownloader("paper", "1.20.1", tmpDir)
	//
	// callback := func(progress DownloadProgress) {
	// 	if progress.Percent > 0 {
	// 		t.Logf("%.1f%% - %s", progress.Percent, progress.Message)
	// 	} else {
	// 		t.Log(progress.Message)
	// 	}
	// }
	//
	// filePath, err := downloader.Download(callback)
	// if err != nil {
	// 	t.Fatalf("Error descargando: %v", err)
	// }
	//
	// // Verificar que el archivo existe y tiene tamaño razonable
	// info, err := os.Stat(filePath)
	// if err != nil {
	// 	t.Fatalf("Error verificando archivo: %v", err)
	// }
	//
	// if info.Size() < 10*1024*1024 { // Menos de 10MB sería sospechoso
	// 	t.Errorf("Tamaño de archivo sospechosamente pequeño: %d bytes", info.Size())
	// }
	//
	// t.Logf("Descarga exitosa: %s (%s)", filePath, formatBytes(info.Size()))
}
