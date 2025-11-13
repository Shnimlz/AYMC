package websocket

import (
	"time"

	"github.com/google/uuid"
)

// MessageType representa el tipo de mensaje WebSocket
type MessageType string

const (
	// Tipos de mensajes de servidor a cliente
	MessageTypeLogEntry     MessageType = "log_entry"
	MessageTypeMetrics      MessageType = "metrics"
	MessageTypeServerStatus MessageType = "server_status"
	MessageTypeAlert        MessageType = "alert"
	MessageTypeNotification MessageType = "notification"
	MessageTypeError        MessageType = "error"
	MessageTypePong         MessageType = "pong"

	// Tipos de mensajes de cliente a servidor
	MessageTypeSubscribe   MessageType = "subscribe"
	MessageTypeUnsubscribe MessageType = "unsubscribe"
	MessageTypePing        MessageType = "ping"
)

// Message es la estructura base de todos los mensajes WebSocket
type Message struct {
	Type      MessageType `json:"type"`
	Channel   string      `json:"channel,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// SubscribeMessage representa una solicitud de suscripción
type SubscribeMessage struct {
	Channels []string `json:"channels"`
}

// UnsubscribeMessage representa una solicitud de cancelación de suscripción
type UnsubscribeMessage struct {
	Channels []string `json:"channels"`
}

// LogEntry representa una entrada de log de un servidor
type LogEntry struct {
	ServerID  uuid.UUID `json:"server_id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`     // INFO, WARN, ERROR, DEBUG
	Source    string    `json:"source"`    // server, plugin, etc.
	Message   string    `json:"message"`
	Exception string    `json:"exception,omitempty"`
}

// ServerMetrics representa métricas en tiempo real de un servidor
type ServerMetrics struct {
	ServerID      uuid.UUID `json:"server_id"`
	Timestamp     time.Time `json:"timestamp"`
	CPUPercent    float64   `json:"cpu_percent"`
	MemoryUsed    uint64    `json:"memory_used"`
	MemoryTotal   uint64    `json:"memory_total"`
	MemoryPercent float64   `json:"memory_percent"`
	PlayersOnline int32     `json:"players_online"`
	MaxPlayers    int32     `json:"max_players"`
	TPS           float64   `json:"tps,omitempty"` // Ticks per second
	UptimeSeconds int64     `json:"uptime_seconds"`
}

// ServerStatusChange representa un cambio de estado de un servidor
type ServerStatusChange struct {
	ServerID   uuid.UUID `json:"server_id"`
	ServerName string    `json:"server_name"`
	OldStatus  string    `json:"old_status"`
	NewStatus  string    `json:"new_status"`
	Timestamp  time.Time `json:"timestamp"`
	Reason     string    `json:"reason,omitempty"`
}

// Alert representa una alerta del sistema
type Alert struct {
	ID        uuid.UUID `json:"id"`
	Severity  string    `json:"severity"` // info, warning, error, critical
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Source    string    `json:"source"` // server, agent, system
	SourceID  uuid.UUID `json:"source_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Notification representa una notificación general para el usuario
type Notification struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Type      string    `json:"type"` // info, success, warning, error
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Link      string    `json:"link,omitempty"`
	Read      bool      `json:"read"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorMessage representa un mensaje de error
type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ChannelType representa los tipos de canales disponibles
type ChannelType string

const (
	ChannelTypeLogs         ChannelType = "logs"
	ChannelTypeMetrics      ChannelType = "metrics"
	ChannelTypeStatus       ChannelType = "status"
	ChannelTypeNotification ChannelType = "notifications"
)

// BuildChannel construye el nombre de un canal basado en tipo y ID
func BuildChannel(channelType ChannelType, resourceID uuid.UUID) string {
	switch channelType {
	case ChannelTypeLogs:
		return "server:" + resourceID.String() + ":logs"
	case ChannelTypeMetrics:
		return "server:" + resourceID.String() + ":metrics"
	case ChannelTypeStatus:
		return "server:" + resourceID.String() + ":status"
	case ChannelTypeNotification:
		return "user:" + resourceID.String() + ":notifications"
	default:
		return ""
	}
}

// BuildUserChannel construye el canal de notificaciones de un usuario
func BuildUserChannel(userID uuid.UUID) string {
	return "user:" + userID.String() + ":notifications"
}

// BuildServerLogsChannel construye el canal de logs de un servidor
func BuildServerLogsChannel(serverID uuid.UUID) string {
	return "server:" + serverID.String() + ":logs"
}

// BuildServerMetricsChannel construye el canal de métricas de un servidor
func BuildServerMetricsChannel(serverID uuid.UUID) string {
	return "server:" + serverID.String() + ":metrics"
}

// BuildServerStatusChannel construye el canal de estado de un servidor
func BuildServerStatusChannel(serverID uuid.UUID) string {
	return "server:" + serverID.String() + ":status"
}

// NewMessage crea un nuevo mensaje con timestamp actual
func NewMessage(msgType MessageType, channel string, data interface{}) Message {
	return Message{
		Type:      msgType,
		Channel:   channel,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewLogEntryMessage crea un mensaje de entrada de log
func NewLogEntryMessage(serverID uuid.UUID, entry LogEntry) Message {
	channel := BuildServerLogsChannel(serverID)
	return NewMessage(MessageTypeLogEntry, channel, entry)
}

// NewMetricsMessage crea un mensaje de métricas
func NewMetricsMessage(serverID uuid.UUID, metrics ServerMetrics) Message {
	channel := BuildServerMetricsChannel(serverID)
	return NewMessage(MessageTypeMetrics, channel, metrics)
}

// NewServerStatusMessage crea un mensaje de cambio de estado
func NewServerStatusMessage(serverID uuid.UUID, status ServerStatusChange) Message {
	channel := BuildServerStatusChannel(serverID)
	return NewMessage(MessageTypeServerStatus, channel, status)
}

// NewNotificationMessage crea un mensaje de notificación
func NewNotificationMessage(userID uuid.UUID, notification Notification) Message {
	channel := BuildUserChannel(userID)
	return NewMessage(MessageTypeNotification, channel, notification)
}

// NewAlertMessage crea un mensaje de alerta
func NewAlertMessage(alert Alert) Message {
	// Las alertas se envían a todos los usuarios suscritos al servidor/agente
	return NewMessage(MessageTypeAlert, "", alert)
}

// NewErrorMessage crea un mensaje de error
func NewErrorMessage(code, message, details string) Message {
	return NewMessage(MessageTypeError, "", ErrorMessage{
		Code:    code,
		Message: message,
		Details: details,
	})
}
