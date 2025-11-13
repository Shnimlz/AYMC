package middleware

import (
	"net/http"
	"strings"

	"github.com/aymc/backend/database"
	"github.com/aymc/backend/database/models"
	"github.com/aymc/backend/services/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// Context keys for user data
	UserContextKey   = "user"
	UserIDContextKey = "user_id"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware(jwtService *auth.JWTService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			logger.Warn("Invalid token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Verify it's an access token
		if claims.Type != auth.AccessToken {
			logger.Warn("Invalid token type", zap.String("type", string(claims.Type)))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token type",
			})
			c.Abort()
			return
		}

		// Get user from database to verify it still exists and is active
		db := database.GetDB()
		var user models.User
		if err := db.First(&user, "id = ?", claims.UserID).Error; err != nil {
			logger.Warn("User not found", zap.String("user_id", claims.UserID.String()))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found",
			})
			c.Abort()
			return
		}

		// Check if user is active
		if !user.IsActive {
			logger.Warn("User account inactive", zap.String("user_id", user.ID.String()))
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Account is inactive",
			})
			c.Abort()
			return
		}

		// Store user data in context
		c.Set(UserContextKey, &user)
		c.Set(UserIDContextKey, user.ID)

		logger.Debug("User authenticated",
			zap.String("user_id", user.ID.String()),
			zap.String("username", user.Username),
			zap.String("role", string(user.Role)),
		)

		c.Next()
	}
}

// RequireRole creates a middleware that checks if user has required role
func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by AuthMiddleware)
		userInterface, exists := c.Get(UserContextKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user context",
			})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, role := range roles {
			if user.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Insufficient permissions",
				"required_role": roles,
				"user_role":     user.Role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin is a convenience middleware for admin-only routes
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

// GetUserFromContext retrieves the authenticated user from context
func GetUserFromContext(c *gin.Context) (*models.User, bool) {
	userInterface, exists := c.Get(UserContextKey)
	if !exists {
		return nil, false
	}

	user, ok := userInterface.(*models.User)
	return user, ok
}

// GetUserID retrieves the authenticated user ID from context
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDInterface, exists := c.Get(UserIDContextKey)
	if !exists {
		return uuid.Nil, false
	}

	userID, ok := userIDInterface.(uuid.UUID)
	return userID, ok
}

// MustGetUser retrieves user from context or panics (use only in authenticated routes)
func MustGetUser(c *gin.Context) *models.User {
	user, ok := GetUserFromContext(c)
	if !ok {
		panic("user not found in context")
	}
	return user
}

// MustGetUserID retrieves user ID from context or panics (use only in authenticated routes)
func MustGetUserID(c *gin.Context) uuid.UUID {
	userID, ok := GetUserID(c)
	if !ok {
		panic("user ID not found in context")
	}
	return userID
}
