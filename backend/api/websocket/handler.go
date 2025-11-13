package websocket

import (
	"net/http"
	"strings"

	"github.com/aymc/backend/database/models"
	"github.com/aymc/backend/services/auth"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Permitir todas las orígenes en desarrollo
	// TODO: Configurar CORS apropiadamente en producción
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handler maneja las solicitudes WebSocket
type Handler struct {
	hub        *Hub
	jwtService *auth.JWTService
	logger     *zap.Logger
}

// NewHandler crea un nuevo handler de WebSocket
func NewHandler(hub *Hub, jwtService *auth.JWTService, logger *zap.Logger) *Handler {
	return &Handler{
		hub:        hub,
		jwtService: jwtService,
		logger:     logger,
	}
}

// HandleWebSocket maneja el upgrade de HTTP a WebSocket
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Autenticar usuario desde JWT token
	user, err := h.authenticateUser(c)
	if err != nil {
		h.logger.Warn("WebSocket authentication failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication failed",
		})
		return
	}

	// Upgrade HTTP connection a WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
		)
		return
	}

	// Crear cliente
	client := NewClient(h.hub, conn, user, h.logger)

	// Registrar cliente en el hub
	h.hub.register <- client

	h.logger.Info("WebSocket connection established",
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username),
		zap.String("remote_addr", c.Request.RemoteAddr),
	)

	// Iniciar goroutines para leer y escribir
	go client.WritePump()
	go client.ReadPump()
}

// authenticateUser extrae y valida el JWT token
func (h *Handler) authenticateUser(c *gin.Context) (*models.User, error) {
	// Intentar obtener token de múltiples fuentes
	token := h.extractToken(c)
	if token == "" {
		return nil, http.ErrNoCookie
	}

	// Validar token
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// Crear objeto User desde claims
	user := &models.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Role:     models.UserRole(claims.Role),
	}

	return user, nil
}

// extractToken extrae el token JWT de la solicitud
func (h *Handler) extractToken(c *gin.Context) string {
	// 1. Buscar en query parameter (para compatibilidad WebSocket)
	token := c.Query("token")
	if token != "" {
		return token
	}

	// 2. Buscar en header Authorization
	bearerToken := c.GetHeader("Authorization")
	if bearerToken != "" {
		parts := strings.Split(bearerToken, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// 3. Buscar en cookie
	cookie, err := c.Cookie("token")
	if err == nil && cookie != "" {
		return cookie
	}

	return ""
}

// GetHub retorna el hub
func (h *Handler) GetHub() *Hub {
	return h.hub
}
