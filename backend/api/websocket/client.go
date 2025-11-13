package websocket

import (
	"encoding/json"
	"time"

	"github.com/aymc/backend/database/models"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// Tiempo permitido para escribir un mensaje al peer
	writeWait = 10 * time.Second

	// Tiempo permitido para leer el siguiente mensaje pong del peer
	pongWait = 60 * time.Second

	// Enviar pings al peer con este período. Debe ser menor que pongWait
	pingPeriod = (pongWait * 9) / 10

	// Tamaño máximo del mensaje permitido desde el peer
	maxMessageSize = 512 * 1024 // 512 KB
)

// Client representa un cliente WebSocket individual
type Client struct {
	// Hub al que pertenece el cliente
	hub *Hub

	// Conexión WebSocket
	conn *websocket.Conn

	// Canal buffered para enviar mensajes al cliente
	send chan []byte

	// Usuario autenticado
	user *models.User

	// Canales a los que está suscrito el cliente
	subscriptions map[string]bool

	// Logger
	logger *zap.Logger
}

// NewClient crea una nueva instancia de Cliente
func NewClient(hub *Hub, conn *websocket.Conn, user *models.User, logger *zap.Logger) *Client {
	return &Client{
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		user:          user,
		subscriptions: make(map[string]bool),
		logger:        logger,
	}
}

// ReadPump bombea mensajes desde la conexión WebSocket al hub
//
// La aplicación ejecuta ReadPump en una goroutine por conexión.
// La aplicación asegura que solo hay un lector por conexión ejecutando
// todos los reads desde esta goroutine.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("WebSocket read error",
					zap.Error(err),
					zap.String("user_id", c.user.ID.String()),
				)
			}
			break
		}

		// Parsear mensaje
		var baseMsg struct {
			Type MessageType    `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		if err := json.Unmarshal(messageBytes, &baseMsg); err != nil {
			c.logger.Warn("Failed to parse message",
				zap.Error(err),
				zap.String("user_id", c.user.ID.String()),
			)
			c.sendError("PARSE_ERROR", "Invalid message format", err.Error())
			continue
		}

		// Procesar mensaje según tipo
		switch baseMsg.Type {
		case MessageTypeSubscribe:
			c.handleSubscribe(baseMsg.Data)
		case MessageTypeUnsubscribe:
			c.handleUnsubscribe(baseMsg.Data)
		case MessageTypePing:
			c.handlePing()
		default:
			c.logger.Warn("Unknown message type",
				zap.String("type", string(baseMsg.Type)),
				zap.String("user_id", c.user.ID.String()),
			)
			c.sendError("UNKNOWN_TYPE", "Unknown message type", string(baseMsg.Type))
		}
	}
}

// WritePump bombea mensajes desde el hub a la conexión WebSocket
//
// Una goroutine ejecutando WritePump se inicia por cada conexión.
// La aplicación asegura que solo hay un escritor por conexión ejecutando
// todos los writes desde esta goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub cerró el canal
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Agregar mensajes en cola al mensaje actual
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSubscribe maneja solicitud de suscripción a canales
func (c *Client) handleSubscribe(data json.RawMessage) {
	var subMsg SubscribeMessage
	if err := json.Unmarshal(data, &subMsg); err != nil {
		c.logger.Warn("Failed to parse subscribe message",
			zap.Error(err),
			zap.String("user_id", c.user.ID.String()),
		)
		c.sendError("PARSE_ERROR", "Invalid subscribe format", err.Error())
		return
	}

	// TODO: Validar permisos del usuario para suscribirse a cada canal
	// Por ahora, permitir todas las suscripciones

	for _, channel := range subMsg.Channels {
		c.hub.subscribeToChannel(c, channel)
	}

	c.logger.Info("Client subscribed to channels",
		zap.String("user_id", c.user.ID.String()),
		zap.Strings("channels", subMsg.Channels),
		zap.Int("total_subscriptions", len(c.subscriptions)),
	)

	// Enviar confirmación
	c.sendSuccess("SUBSCRIBED", "Successfully subscribed to channels", subMsg.Channels)
}

// handleUnsubscribe maneja solicitud de cancelación de suscripción
func (c *Client) handleUnsubscribe(data json.RawMessage) {
	var unsubMsg UnsubscribeMessage
	if err := json.Unmarshal(data, &unsubMsg); err != nil {
		c.logger.Warn("Failed to parse unsubscribe message",
			zap.Error(err),
			zap.String("user_id", c.user.ID.String()),
		)
		c.sendError("PARSE_ERROR", "Invalid unsubscribe format", err.Error())
		return
	}

	for _, channel := range unsubMsg.Channels {
		c.hub.unsubscribeFromChannel(c, channel)
	}

	c.logger.Info("Client unsubscribed from channels",
		zap.String("user_id", c.user.ID.String()),
		zap.Strings("channels", unsubMsg.Channels),
		zap.Int("total_subscriptions", len(c.subscriptions)),
	)

	// Enviar confirmación
	c.sendSuccess("UNSUBSCRIBED", "Successfully unsubscribed from channels", unsubMsg.Channels)
}

// handlePing maneja ping del cliente
func (c *Client) handlePing() {
	pongMsg := NewMessage(MessageTypePong, "", map[string]interface{}{
		"timestamp": time.Now().Unix(),
	})

	messageJSON, _ := json.Marshal(pongMsg)
	select {
	case c.send <- messageJSON:
	default:
		c.logger.Warn("Failed to send pong, buffer full",
			zap.String("user_id", c.user.ID.String()),
		)
	}
}

// sendError envía un mensaje de error al cliente
func (c *Client) sendError(code, message, details string) {
	errMsg := NewErrorMessage(code, message, details)
	messageJSON, _ := json.Marshal(errMsg)

	select {
	case c.send <- messageJSON:
	default:
		c.logger.Warn("Failed to send error message, buffer full",
			zap.String("user_id", c.user.ID.String()),
		)
	}
}

// sendSuccess envía un mensaje de éxito al cliente
func (c *Client) sendSuccess(code, message string, data interface{}) {
	successMsg := NewMessage(MessageTypeNotification, "", map[string]interface{}{
		"code":    code,
		"message": message,
		"data":    data,
	})

	messageJSON, _ := json.Marshal(successMsg)
	select {
	case c.send <- messageJSON:
	default:
		c.logger.Warn("Failed to send success message, buffer full",
			zap.String("user_id", c.user.ID.String()),
		)
	}
}

// Subscribe suscribe al cliente a un canal específico
func (c *Client) Subscribe(channel string) {
	c.hub.subscribeToChannel(c, channel)
}

// Unsubscribe cancela la suscripción del cliente a un canal
func (c *Client) Unsubscribe(channel string) {
	c.hub.unsubscribeFromChannel(c, channel)
}

// GetSubscriptions retorna los canales a los que está suscrito
func (c *Client) GetSubscriptions() []string {
	channels := make([]string, 0, len(c.subscriptions))
	for channel := range c.subscriptions {
		channels = append(channels, channel)
	}
	return channels
}

// IsSubscribed verifica si el cliente está suscrito a un canal
func (c *Client) IsSubscribed(channel string) bool {
	return c.subscriptions[channel]
}
