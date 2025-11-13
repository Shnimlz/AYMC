package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Hub mantiene el conjunto de clientes activos y broadcast de mensajes
type Hub struct {
	// Clientes registrados
	clients map[*Client]bool

	// Mensajes broadcast desde los clientes
	broadcast chan Message

	// Registrar solicitudes de los clientes
	register chan *Client

	// Cancelar registro de los clientes
	unregister chan *Client

	// Suscripciones: canal -> conjunto de clientes
	subscriptions map[string]map[*Client]bool

	// Mutex para acceso concurrente
	mu sync.RWMutex

	// Logger
	logger *zap.Logger

	// Context para shutdown graceful
	ctx    context.Context
	cancel context.CancelFunc
}

// NewHub crea una nueva instancia de Hub
func NewHub(logger *zap.Logger) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		clients:       make(map[*Client]bool),
		broadcast:     make(chan Message, 256),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		subscriptions: make(map[string]map[*Client]bool),
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Run inicia el hub principal loop
func (h *Hub) Run() {
	h.logger.Info("WebSocket Hub started")
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			h.logger.Info("WebSocket Hub shutting down")
			h.closeAllClients()
			return

		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-ticker.C:
			// Ping periódico a todos los clientes
			h.pingAllClients()
		}
	}
}

// Stop detiene el hub gracefully
func (h *Hub) Stop() {
	h.logger.Info("Stopping WebSocket Hub")
	h.cancel()
}

// registerClient registra un nuevo cliente
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true
	
	h.logger.Info("Client registered",
		zap.String("user_id", client.user.ID.String()),
		zap.String("username", client.user.Username),
		zap.Int("total_clients", len(h.clients)),
	)
}

// unregisterClient cancela el registro de un cliente
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		// Remover de todos los canales suscritos
		for channel := range client.subscriptions {
			h.unsubscribeFromChannel(client, channel)
		}

		delete(h.clients, client)
		close(client.send)

		h.logger.Info("Client unregistered",
			zap.String("user_id", client.user.ID.String()),
			zap.String("username", client.user.Username),
			zap.Int("total_clients", len(h.clients)),
		)
	}
}

// subscribeToChannel suscribe un cliente a un canal
func (h *Hub) subscribeToChannel(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.subscriptions[channel] == nil {
		h.subscriptions[channel] = make(map[*Client]bool)
	}

	h.subscriptions[channel][client] = true
	client.subscriptions[channel] = true

	h.logger.Debug("Client subscribed to channel",
		zap.String("user_id", client.user.ID.String()),
		zap.String("channel", channel),
		zap.Int("subscribers", len(h.subscriptions[channel])),
	)
}

// unsubscribeFromChannel cancela la suscripción de un cliente a un canal
func (h *Hub) unsubscribeFromChannel(client *Client, channel string) {
	if clients, ok := h.subscriptions[channel]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.subscriptions, channel)
		}
	}
	delete(client.subscriptions, channel)

	h.logger.Debug("Client unsubscribed from channel",
		zap.String("user_id", client.user.ID.String()),
		zap.String("channel", channel),
	)
}

// broadcastMessage envía un mensaje a todos los clientes suscritos al canal
func (h *Hub) broadcastMessage(message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Si el mensaje tiene un canal específico, solo enviarlo a clientes suscritos
	if message.Channel != "" {
		if clients, ok := h.subscriptions[message.Channel]; ok {
			h.sendToClients(clients, message)
		}
	} else {
		// Sin canal = broadcast a todos
		h.sendToClients(h.clients, message)
	}
}

// sendToClients envía un mensaje a un conjunto de clientes
func (h *Hub) sendToClients(clients map[*Client]bool, message Message) {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal message", zap.Error(err))
		return
	}

	count := 0
	for client := range clients {
		select {
		case client.send <- messageJSON:
			count++
		default:
			// Cliente no puede recibir, cerrar
			h.logger.Warn("Client send buffer full, closing connection",
				zap.String("user_id", client.user.ID.String()),
			)
			close(client.send)
			delete(h.clients, client)
		}
	}

	h.logger.Debug("Message broadcasted",
		zap.String("type", string(message.Type)),
		zap.String("channel", message.Channel),
		zap.Int("recipients", count),
	)
}

// BroadcastToChannel envía un mensaje a todos los clientes suscritos a un canal
func (h *Hub) BroadcastToChannel(channel string, message Message) {
	message.Channel = channel
	h.broadcast <- message
}

// BroadcastToUser envía un mensaje a un usuario específico
func (h *Hub) BroadcastToUser(userID uuid.UUID, message Message) {
	channel := BuildUserChannel(userID)
	h.BroadcastToChannel(channel, message)
}

// BroadcastToServer envía un mensaje a todos los clientes suscritos a un servidor
func (h *Hub) BroadcastToServer(serverID uuid.UUID, channelType ChannelType, message Message) {
	channel := BuildChannel(channelType, serverID)
	h.BroadcastToChannel(channel, message)
}

// BroadcastServerLogs envía logs de un servidor
func (h *Hub) BroadcastServerLogs(serverID uuid.UUID, entry LogEntry) {
	message := NewLogEntryMessage(serverID, entry)
	h.broadcast <- message
}

// BroadcastServerMetrics envía métricas de un servidor
func (h *Hub) BroadcastServerMetrics(serverID uuid.UUID, metrics ServerMetrics) {
	message := NewMetricsMessage(serverID, metrics)
	h.broadcast <- message
}

// BroadcastServerStatus envía cambio de estado de un servidor
func (h *Hub) BroadcastServerStatus(serverID uuid.UUID, status ServerStatusChange) {
	message := NewServerStatusMessage(serverID, status)
	h.broadcast <- message
}

// BroadcastNotification envía una notificación a un usuario
func (h *Hub) BroadcastNotification(userID uuid.UUID, notification Notification) {
	message := NewNotificationMessage(userID, notification)
	h.broadcast <- message
}

// BroadcastAlert envía una alerta global
func (h *Hub) BroadcastAlert(alert Alert) {
	message := NewAlertMessage(alert)
	h.broadcast <- message
}

// pingAllClients envía ping a todos los clientes conectados
func (h *Hub) pingAllClients() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	message := NewMessage(MessageTypePong, "", map[string]interface{}{
		"timestamp": time.Now().Unix(),
	})

	messageJSON, _ := json.Marshal(message)

	for client := range h.clients {
		select {
		case client.send <- messageJSON:
		default:
			// Skip si el buffer está lleno
		}
	}

	h.logger.Debug("Pinged all clients", zap.Int("count", len(h.clients)))
}

// closeAllClients cierra todas las conexiones de clientes
func (h *Hub) closeAllClients() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		close(client.send)
		delete(h.clients, client)
	}

	h.logger.Info("All clients closed", zap.Int("count", len(h.clients)))
}

// GetStats retorna estadísticas del hub
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]interface{}{
		"total_clients":      len(h.clients),
		"total_subscriptions": len(h.subscriptions),
		"timestamp":          time.Now().Unix(),
	}
}

// GetClientCount retorna el número de clientes conectados
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetSubscriptionCount retorna el número de canales con suscriptores
func (h *Hub) GetSubscriptionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subscriptions)
}
