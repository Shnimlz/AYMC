package core

import (
"strings"
"testing"
)

func TestNewLogParser(t *testing.T) {
parser := NewLogParser()
if parser == nil {
t.Fatal("Parser es nil")
}
}

func TestParseSimpleLog(t *testing.T) {
parser := NewLogParser()

logLine := "[10:30:45] [Server thread/INFO]: Starting minecraft server version 1.20.1"
entry := parser.ParseLog(logLine)

if entry == nil {
t.Fatal("Entry es nil")
}

if entry.Level != LogLevelInfo {
t.Errorf("Level esperado INFO, obtenido %s", entry.Level)
}
}

func TestParseException(t *testing.T) {
parser := NewLogParser()

logLine := "[10:30:45] [Server thread/ERROR]: java.lang.NullPointerException: Cannot read field"
entry := parser.ParseLog(logLine)

if !entry.IsException {
t.Error("Deberia detectar excepcion")
}

if entry.ErrorType != "java.lang.NullPointerException" {
t.Errorf("ErrorType esperado NullPointerException, obtenido %s", entry.ErrorType)
}

if entry.Level != LogLevelError {
t.Error("Level deberia ser ERROR para excepciones")
}
}

func TestParseStackTrace(t *testing.T) {
parser := NewLogParser()

logLine := "    at com.example.MyPlugin.onEnable(MyPlugin.java:123)"
entry := parser.ParseLog(logLine)

if entry.ClassName != "com.example.MyPlugin" {
t.Errorf("ClassName esperado com.example.MyPlugin, obtenido %s", entry.ClassName)
}

if entry.MethodName != "onEnable" {
t.Errorf("MethodName esperado onEnable, obtenido %s", entry.MethodName)
}

if entry.LineNumber != 123 {
t.Errorf("LineNumber esperado 123, obtenido %d", entry.LineNumber)
}
}

func TestNormalizeLevel(t *testing.T) {
parser := NewLogParser()

tests := []struct {
input    string
expected LogLevel
}{
{"INFO", LogLevelInfo},
{"INFORMATION", LogLevelInfo},
{"WARN", LogLevelWarn},
{"WARNING", LogLevelWarn},
{"ERROR", LogLevelError},
{"SEVERE", LogLevelSevere},
{"DEBUG", LogLevelDebug},
{"UNKNOWN_LEVEL", LogLevelUnknown},
}

for _, tt := range tests {
result := parser.normalizeLevel(tt.input)
if result != tt.expected {
t.Errorf("normalizeLevel(%s) esperado %s, obtenido %s", 
tt.input, tt.expected, result)
}
}
}

func TestClassifySource(t *testing.T) {
parser := NewLogParser()

tests := []struct {
logLine  string
expected string
}{
{"[Server thread/INFO]: Starting server", "SERVER"},
{"[Async Task/INFO]: Running async task", "ASYNC"},
{"[Netty IO Thread/INFO]: Connection", "NETWORK"},
{"[Chunk Generator/INFO]: Generating chunks", "WORLD"},
{"[Player Join/INFO]: Player connected", "PLAYER"},
{"[Plugin Manager/INFO]: Loading plugins", "PLUGIN"},
}

for _, tt := range tests {
result := parser.classifySource(tt.logLine)
if result != tt.expected {
t.Errorf("classifySource(%s) esperado %s, obtenido %s",
tt.logLine, tt.expected, result)
}
}
}

func TestIsError(t *testing.T) {
entry := &ParsedLogEntry{Level: LogLevelError}
if !entry.IsError() {
t.Error("Deberia ser error")
}

entry.Level = LogLevelInfo
entry.IsException = true
if !entry.IsError() {
t.Error("Excepcion deberia ser error")
}
}

func TestGetSeverity(t *testing.T) {
tests := []struct {
level    LogLevel
severity int
}{
{LogLevelDebug, 1},
{LogLevelInfo, 2},
{LogLevelWarn, 3},
{LogLevelError, 4},
{LogLevelSevere, 5},
{LogLevelUnknown, 0},
}

for _, tt := range tests {
entry := &ParsedLogEntry{Level: tt.level}
if entry.GetSeverity() != tt.severity {
t.Errorf("Severity para %s esperado %d, obtenido %d",
tt.level, tt.severity, entry.GetSeverity())
}
}
}

func TestNewErrorDetector(t *testing.T) {
detector := NewErrorDetector()
if detector == nil {
t.Fatal("Detector es nil")
}

if len(detector.patterns) == 0 {
t.Error("Detector deberia tener patrones")
}
}

func TestDetectOutOfMemoryError(t *testing.T) {
detector := NewErrorDetector()
parser := NewLogParser()

logLine := "[ERROR]: java.lang.OutOfMemoryError: Java heap space"
entry := parser.ParseLog(logLine)

pattern := detector.DetectError(entry)
if pattern == nil {
t.Fatal("Deberia detectar OutOfMemoryError")
}

if pattern.ErrorType != "OUT_OF_MEMORY" {
t.Errorf("ErrorType esperado OUT_OF_MEMORY, obtenido %s", pattern.ErrorType)
}

if !strings.Contains(pattern.Suggestion, "memoria") {
t.Error("Suggestion deberia mencionar memoria")
}
}

func TestDetectPortInUse(t *testing.T) {
detector := NewErrorDetector()
parser := NewLogParser()

logLine := "[ERROR]: Failed to bind to port 25565: Address already in use"
entry := parser.ParseLog(logLine)

pattern := detector.DetectError(entry)
if pattern == nil {
t.Fatal("Deberia detectar error de puerto")
}

if pattern.ErrorType != "PORT_IN_USE" {
t.Errorf("ErrorType esperado PORT_IN_USE, obtenido %s", pattern.ErrorType)
}
}

func TestDetectMissingDependency(t *testing.T) {
detector := NewErrorDetector()
parser := NewLogParser()

logLine := "[ERROR]: java.lang.NoClassDefFoundError: org/bukkit/plugin/Plugin"
entry := parser.ParseLog(logLine)

pattern := detector.DetectError(entry)
if pattern == nil {
t.Fatal("Deberia detectar dependencia faltante")
}

if pattern.ErrorType != "MISSING_DEPENDENCY" {
t.Errorf("ErrorType esperado MISSING_DEPENDENCY, obtenido %s", pattern.ErrorType)
}
}
