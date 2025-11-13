package core

import (
	"regexp"
	"strings"
	"time"
)

// LogLevel representa el nivel de severidad del log
type LogLevel string

const (
	LogLevelDebug   LogLevel = "DEBUG"
	LogLevelInfo    LogLevel = "INFO"
	LogLevelWarn    LogLevel = "WARN"
	LogLevelError   LogLevel = "ERROR"
	LogLevelSevere  LogLevel = "SEVERE"
	LogLevelUnknown LogLevel = "UNKNOWN"
)

// ParsedLogEntry representa un log parseado con metadata
type ParsedLogEntry struct {
	Timestamp   time.Time
	Level       LogLevel
	Source      string
	Message     string
	Plugin      string
	ClassName   string
	MethodName  string
	LineNumber  int
	ErrorType   string
	StackTrace  []string
	IsException bool
}

// LogParser parsea logs de servidores de Minecraft
type LogParser struct {
	// Patrones de regex compilados
	timestampPattern *regexp.Regexp
	levelPattern     *regexp.Regexp
	pluginPattern    *regexp.Regexp
	exceptionPattern *regexp.Regexp
	atLinePattern    *regexp.Regexp
	causedByPattern  *regexp.Regexp
}

// NewLogParser crea un nuevo parser de logs
func NewLogParser() *LogParser {
	return &LogParser{
		// [HH:MM:SS] [Thread/LEVEL]: message
		timestampPattern: regexp.MustCompile(`^\[(\d{2}:\d{2}:\d{2})\]`),
		
		// [LEVEL] o [Thread/LEVEL]
		levelPattern: regexp.MustCompile(`\[(?:[^/]+/)?(\w+)\]:`),
		
		// [PluginName] en el mensaje
		pluginPattern: regexp.MustCompile(`\[([A-Za-z0-9_-]+)\]`),
		
		// Detectar excepciones Java
		exceptionPattern: regexp.MustCompile(`([A-Za-z0-9.]+(?:Exception|Error))(?::\s*(.+))?`),
		
		// at package.class.method(File.java:123)
		atLinePattern: regexp.MustCompile(`at\s+([A-Za-z0-9.$_]+)\.([A-Za-z0-9_<>]+)\(([A-Za-z0-9.]+):(\d+)\)`),
		
		// Caused by: Exception
		causedByPattern: regexp.MustCompile(`Caused by:\s+([A-Za-z0-9.]+(?:Exception|Error))`),
	}
}

// ParseLog parsea una línea de log
func (lp *LogParser) ParseLog(logLine string) *ParsedLogEntry {
	entry := &ParsedLogEntry{
		Timestamp:  time.Now(),
		Level:      LogLevelUnknown,
		Source:     "UNKNOWN",
		Message:    logLine,
		StackTrace: []string{},
	}

	// Parsear timestamp
	if matches := lp.timestampPattern.FindStringSubmatch(logLine); len(matches) > 1 {
		// Intentar parsear la hora (formato HH:MM:SS)
		timeStr := matches[1]
		if t, err := time.Parse("15:04:05", timeStr); err == nil {
			// Usar la fecha actual con la hora del log
			now := time.Now()
			entry.Timestamp = time.Date(now.Year(), now.Month(), now.Day(), 
				t.Hour(), t.Minute(), t.Second(), 0, time.Local)
		}
	}

	// Parsear nivel
	if matches := lp.levelPattern.FindStringSubmatch(logLine); len(matches) > 1 {
		entry.Level = lp.normalizeLevel(matches[1])
	}

	// Detectar plugin
	if matches := lp.pluginPattern.FindAllStringSubmatch(logLine, -1); len(matches) > 0 {
		// El primer match suele ser el plugin
		entry.Plugin = matches[0][1]
	}

	// Detectar excepciones
	if matches := lp.exceptionPattern.FindStringSubmatch(logLine); len(matches) > 1 {
		entry.IsException = true
		entry.ErrorType = matches[1]
		if len(matches) > 2 && matches[2] != "" {
			entry.Message = matches[2]
		}
		entry.Level = LogLevelError
	}

	// Parsear stack trace (líneas "at ...")
	if matches := lp.atLinePattern.FindStringSubmatch(logLine); len(matches) > 4 {
		entry.ClassName = matches[1]
		entry.MethodName = matches[2]
		// matches[3] es el nombre del archivo
		// matches[4] es el número de línea
		if lineNum := parseInt(matches[4]); lineNum > 0 {
			entry.LineNumber = lineNum
		}
		entry.StackTrace = append(entry.StackTrace, logLine)
	}

	// Detectar "Caused by"
	if matches := lp.causedByPattern.FindStringSubmatch(logLine); len(matches) > 1 {
		entry.IsException = true
		entry.ErrorType = matches[1]
		entry.Level = LogLevelError
	}

	// Clasificar el mensaje
	entry.Source = lp.classifySource(logLine)

	return entry
}

// normalizeLevel normaliza el nivel de log
func (lp *LogParser) normalizeLevel(level string) LogLevel {
	level = strings.ToUpper(level)

	switch level {
	case "DEBUG", "FINE", "FINER", "FINEST":
		return LogLevelDebug
	case "INFO", "INFORMATION":
		return LogLevelInfo
	case "WARN", "WARNING":
		return LogLevelWarn
	case "ERROR":
		return LogLevelError
	case "SEVERE", "FATAL", "CRITICAL":
		return LogLevelSevere
	default:
		return LogLevelUnknown
	}
}

// classifySource clasifica el origen del log
func (lp *LogParser) classifySource(logLine string) string {
	lower := strings.ToLower(logLine)

	if strings.Contains(lower, "server thread") {
		return "SERVER"
	}
	if strings.Contains(lower, "async") {
		return "ASYNC"
	}
	if strings.Contains(lower, "netty") || strings.Contains(lower, "network") {
		return "NETWORK"
	}
	if strings.Contains(lower, "chunk") {
		return "WORLD"
	}
	if strings.Contains(lower, "player") || strings.Contains(lower, "user") {
		return "PLAYER"
	}
	if strings.Contains(lower, "plugin") {
		return "PLUGIN"
	}

	return "SERVER"
}

// IsError verifica si el log es un error
func (entry *ParsedLogEntry) IsError() bool {
	return entry.Level == LogLevelError || entry.Level == LogLevelSevere || entry.IsException
}

// IsWarning verifica si el log es una advertencia
func (entry *ParsedLogEntry) IsWarning() bool {
	return entry.Level == LogLevelWarn
}

// GetSeverity retorna la severidad numérica (mayor = más grave)
func (entry *ParsedLogEntry) GetSeverity() int {
	switch entry.Level {
	case LogLevelDebug:
		return 1
	case LogLevelInfo:
		return 2
	case LogLevelWarn:
		return 3
	case LogLevelError:
		return 4
	case LogLevelSevere:
		return 5
	default:
		return 0
	}
}

// ErrorPattern representa un patrón de error conocido
type ErrorPattern struct {
	Pattern     *regexp.Regexp
	ErrorType   string
	Severity    int
	Suggestion  string
	Plugin      string
	IsCommon    bool
}

// ErrorDetector detecta errores conocidos y sugiere soluciones
type ErrorDetector struct {
	patterns []ErrorPattern
}

// NewErrorDetector crea un nuevo detector de errores
func NewErrorDetector() *ErrorDetector {
	detector := &ErrorDetector{
		patterns: []ErrorPattern{},
	}

	// Agregar patrones comunes
	detector.addCommonPatterns()

	return detector
}

// addCommonPatterns agrega patrones de errores comunes
func (ed *ErrorDetector) addCommonPatterns() {
	// Error de memoria
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`OutOfMemoryError|java\.lang\.OutOfMemoryError`),
		ErrorType:  "OUT_OF_MEMORY",
		Severity:   5,
		Suggestion: "Aumentar la memoria asignada al servidor (flags -Xms y -Xmx)",
		IsCommon:   true,
	})

	// Error de puerto ocupado
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`Address already in use|Failed to bind to port`),
		ErrorType:  "PORT_IN_USE",
		Severity:   5,
		Suggestion: "El puerto está ocupado. Cambiar el puerto en server.properties o detener el proceso que lo usa",
		IsCommon:   true,
	})

	// Plugin faltante
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`Could not load '([^']+)' in folder 'plugins'`),
		ErrorType:  "PLUGIN_LOAD_ERROR",
		Severity:   4,
		Suggestion: "Verificar que el plugin sea compatible con la versión del servidor",
		IsCommon:   true,
	})

	// Dependencia faltante
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`NoClassDefFoundError|ClassNotFoundException`),
		ErrorType:  "MISSING_DEPENDENCY",
		Severity:   4,
		Suggestion: "Faltan dependencias. Verificar que todos los plugins requeridos estén instalados",
		IsCommon:   true,
	})

	// Error de permisos
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`Permission denied|Access denied`),
		ErrorType:  "PERMISSION_ERROR",
		Severity:   3,
		Suggestion: "Verificar permisos de archivos y carpetas del servidor",
		IsCommon:   true,
	})

	// Mundo corrupto
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`Failed to load chunk|Chunk file .* is missing|Corrupted chunk`),
		ErrorType:  "WORLD_CORRUPTION",
		Severity:   4,
		Suggestion: "El mundo puede estar corrupto. Restaurar desde backup o usar herramientas de reparación",
		IsCommon:   true,
	})

	// Timeout de conexión
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`Connection timed out|Read timed out`),
		ErrorType:  "CONNECTION_TIMEOUT",
		Severity:   2,
		Suggestion: "Problemas de red o firewall. Verificar conectividad",
		IsCommon:   true,
	})

	// Error de versión
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`Unsupported API version|requires server version`),
		ErrorType:  "VERSION_MISMATCH",
		Severity:   4,
		Suggestion: "Plugin incompatible con la versión del servidor. Actualizar plugin o servidor",
		IsCommon:   true,
	})

	// Crash del servidor
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`A fatal error has been detected|#\s+SIGSEGV|Server crashed`),
		ErrorType:  "SERVER_CRASH",
		Severity:   5,
		Suggestion: "Crash crítico. Revisar logs completos y reportar a los desarrolladores del servidor/plugins",
		IsCommon:   true,
	})

	// Error de base de datos
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`SQLException|Could not connect to database|Database connection failed`),
		ErrorType:  "DATABASE_ERROR",
		Severity:   4,
		Suggestion: "Error de base de datos. Verificar credenciales y conectividad en la configuración",
		IsCommon:   true,
	})

	// TPS bajo
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`Can't keep up!|running (\d+)ms behind`),
		ErrorType:  "PERFORMANCE_LAG",
		Severity:   3,
		Suggestion: "El servidor no puede mantener el TPS. Reducir entities, optimizar chunks o aumentar recursos",
		IsCommon:   true,
	})

	// Plugin específicos - WorldEdit
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`\[WorldEdit\].*error|WorldEditException`),
		ErrorType:  "WORLDEDIT_ERROR",
		Severity:   3,
		Suggestion: "Error en WorldEdit. Verificar sintaxis de comandos y permisos",
		Plugin:     "WorldEdit",
		IsCommon:   true,
	})

	// Plugin específicos - EssentialsX
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`\[Essentials\].*(?:ERROR|error)|EssentialsException`),
		ErrorType:  "ESSENTIALS_ERROR",
		Severity:   3,
		Suggestion: "Error en EssentialsX. Verificar configuración en plugins/Essentials/config.yml",
		Plugin:     "Essentials",
		IsCommon:   true,
	})

	// Plugin específicos - Vault
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`\[Vault\].*not found|Vault.*dependency`),
		ErrorType:  "VAULT_DEPENDENCY",
		Severity:   4,
		Suggestion: "Vault no encontrado o mal configurado. Muchos plugins requieren Vault para economía/permisos",
		Plugin:     "Vault",
		IsCommon:   true,
	})

	// NullPointerException
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`NullPointerException`),
		ErrorType:  "NULL_POINTER",
		Severity:   3,
		Suggestion: "Error de programación en plugin. Reportar al desarrollador con el stack trace completo",
		IsCommon:   true,
	})

	// Error de configuración
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`YAMLException|Could not parse|Invalid configuration`),
		ErrorType:  "CONFIG_ERROR",
		Severity:   4,
		Suggestion: "Archivo de configuración inválido. Verificar sintaxis YAML (indentación, comillas, etc)",
		IsCommon:   true,
	})

	// Disco lleno
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`No space left on device|Disk is full`),
		ErrorType:  "DISK_FULL",
		Severity:   5,
		Suggestion: "Disco lleno. Liberar espacio eliminando backups antiguos o archivos innecesarios",
		IsCommon:   true,
	})

	// Error de Java
	ed.patterns = append(ed.patterns, ErrorPattern{
		Pattern:    regexp.MustCompile(`Unsupported major\.minor version|java\.lang\.UnsupportedClassVersionError`),
		ErrorType:  "JAVA_VERSION_ERROR",
		Severity:   5,
		Suggestion: "Plugin compilado para versión más nueva de Java. Actualizar Java o downgrade del plugin",
		IsCommon:   true,
	})
}

// DetectError detecta y clasifica un error
func (ed *ErrorDetector) DetectError(entry *ParsedLogEntry) *ErrorPattern {
	// Buscar en el mensaje completo
	message := entry.Message
	if entry.ErrorType != "" {
		message += " " + entry.ErrorType
	}

	// Buscar coincidencias con patrones conocidos
	for i := range ed.patterns {
		if ed.patterns[i].Pattern.MatchString(message) {
			return &ed.patterns[i]
		}
	}

	return nil
}

// ErrorReport representa un reporte de error detallado
type ErrorReport struct {
	ErrorType   string
	Severity    int
	Message     string
	Suggestion  string
	Plugin      string
	StackTrace  []string
	FirstSeen   string
	Occurrences int
	RelatedLogs []string
}

// AnalyzeLogs analiza múltiples líneas de log y genera un reporte
func (lp *LogParser) AnalyzeLogs(logLines []string) []ErrorReport {
	detector := NewErrorDetector()
	reports := make(map[string]*ErrorReport)

	for _, line := range logLines {
		entry := lp.ParseLog(line)
		
		// Detectar errores
		if errorPattern := detector.DetectError(entry); errorPattern != nil {
			key := errorPattern.ErrorType
			
			if report, exists := reports[key]; exists {
				// Incrementar ocurrencias
				report.Occurrences++
				report.RelatedLogs = append(report.RelatedLogs, line)
				if len(entry.StackTrace) > 0 {
					report.StackTrace = append(report.StackTrace, entry.StackTrace...)
				}
			} else {
				// Crear nuevo reporte
				reports[key] = &ErrorReport{
					ErrorType:   errorPattern.ErrorType,
					Severity:    errorPattern.Severity,
					Message:     entry.Message,
					Suggestion:  errorPattern.Suggestion,
					Plugin:      errorPattern.Plugin,
					StackTrace:  entry.StackTrace,
					FirstSeen:   entry.Timestamp.Format("2006-01-02 15:04:05"),
					Occurrences: 1,
					RelatedLogs: []string{line},
				}
			}
		}
	}

	// Convertir map a slice
	result := make([]ErrorReport, 0, len(reports))
	for _, report := range reports {
		result = append(result, *report)
	}

	return result
}

// ExtractStackTrace extrae el stack trace completo de un error
func (lp *LogParser) ExtractStackTrace(logLines []string, startIndex int) []string {
	stackTrace := []string{}
	
	for i := startIndex; i < len(logLines); i++ {
		line := logLines[i]
		trimmed := strings.TrimSpace(line)
		
		// Si es una línea de stack trace (empieza con "at ")
		if strings.HasPrefix(trimmed, "at ") {
			stackTrace = append(stackTrace, trimmed)
		} else if strings.Contains(line, "Caused by:") {
			// Incluir línea "Caused by"
			stackTrace = append(stackTrace, trimmed)
		} else if len(stackTrace) > 0 {
			// Si ya empezamos a capturar y encontramos una línea que no es stack trace, terminar
			break
		}
	}
	
	return stackTrace
}

// GetPluginList extrae lista de plugins mencionados en los logs
func (lp *LogParser) GetPluginList(logLines []string) []string {
	pluginSet := make(map[string]bool)
	pluginPattern := regexp.MustCompile(`\[([A-Z][a-zA-Z0-9_-]+)\]`)
	
	for _, line := range logLines {
		matches := pluginPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				plugin := match[1]
				// Filtrar keywords comunes que no son plugins
				if plugin != "INFO" && plugin != "WARN" && plugin != "ERROR" && 
				   plugin != "DEBUG" && plugin != "Server" && plugin != "Minecraft" {
					pluginSet[plugin] = true
				}
			}
		}
	}
	
	// Convertir set a slice
	plugins := make([]string, 0, len(pluginSet))
	for plugin := range pluginSet {
		plugins = append(plugins, plugin)
	}
	
	return plugins
}

// parseInt convierte string a int de forma segura
func parseInt(s string) int {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		result = result*10 + int(c-'0')
	}
	return result
}
