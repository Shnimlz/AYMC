package core

import (
	"testing"
)

// BenchmarkParseLog prueba el rendimiento del parser de logs
func BenchmarkParseLog(b *testing.B) {
	parser := NewLogParser()
	logLine := "[10:30:45] [Server thread/INFO]: [Essentials] Loading Essentials v2.19.0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.ParseLog(logLine)
	}
}

// BenchmarkParseLogWithException prueba el parser con una excepción
func BenchmarkParseLogWithException(b *testing.B) {
	parser := NewLogParser()
	logLine := "[10:30:45] [Server thread/ERROR]: java.lang.OutOfMemoryError: Java heap space"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.ParseLog(logLine)
	}
}

// BenchmarkDetectError prueba la detección de errores
func BenchmarkDetectError(b *testing.B) {
	parser := NewLogParser()
	detector := NewErrorDetector()
	entry := parser.ParseLog("[10:30:45] [Server thread/ERROR]: java.lang.OutOfMemoryError: Java heap space")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.DetectError(entry)
	}
}

// BenchmarkAnalyzeLogs prueba el análisis de múltiples líneas
func BenchmarkAnalyzeLogs(b *testing.B) {
	parser := NewLogParser()
	logs := []string{
		"[10:30:45] [Server thread/INFO]: Starting server",
		"[10:30:46] [Server thread/WARN]: Can't keep up! Is the server overloaded?",
		"[10:30:47] [Server thread/ERROR]: java.lang.OutOfMemoryError: Java heap space",
		"[10:30:48] [Server thread/INFO]: [WorldEdit] Enabling WorldEdit",
		"[10:30:49] [Server thread/ERROR]: java.lang.NullPointerException",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.AnalyzeLogs(logs)
	}
}

// BenchmarkGetPluginList prueba la extracción de lista de plugins
func BenchmarkGetPluginList(b *testing.B) {
	parser := NewLogParser()
	logs := []string{
		"[10:30:45] [Server thread/INFO]: [Essentials] Loading config",
		"[10:30:46] [Server thread/INFO]: [WorldEdit] Initializing",
		"[10:30:47] [Server thread/INFO]: [Vault] Enabled",
		"[10:30:48] [Server thread/INFO]: [LuckPerms] Starting",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.GetPluginList(logs)
	}
}

// BenchmarkSystemMonitorGetMetrics prueba el rendimiento del monitoreo
func BenchmarkSystemMonitorGetMetrics(b *testing.B) {
	monitor := NewSystemMonitor()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = monitor.GetMetrics()
	}
}

// BenchmarkSystemMonitorGetOpenPorts prueba la obtención de puertos abiertos
func BenchmarkSystemMonitorGetOpenPorts(b *testing.B) {
	monitor := NewSystemMonitor()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = monitor.GetOpenPorts()
	}
}

// BenchmarkNormalizeLevel prueba la normalización de niveles de log
func BenchmarkNormalizeLevel(b *testing.B) {
	parser := NewLogParser()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.normalizeLevel("ERROR")
	}
}
