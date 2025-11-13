package handlers

import (
	"errors"
	"net/http"

	"github.com/aymc/backend/api/rest/middleware"
	"github.com/aymc/backend/services/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *auth.AuthService
	validator   *validator.Validate
	logger      *zap.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
		logger:      logger,
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Register handles user registration
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.RegisterRequest true "Registration data"
// @Success 201 {object} auth.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid registration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Details: err.Error(),
		})
		return
	}

	// Register user
	user, err := h.authService.Register(&req)
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error: "User already exists",
			})
			return
		}
		if errors.Is(err, auth.ErrWeakPassword) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Password does not meet requirements (min 8 characters)",
			})
			return
		}
		h.logger.Error("Registration failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Registration failed",
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login handles user authentication
// @Summary Authenticate user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.LoginRequest true "Login credentials"
// @Success 200 {object} auth.LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Details: err.Error(),
		})
		return
	}

	// Login user
	response, err := h.authService.Login(&req)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "Invalid email or password",
			})
			return
		}
		if errors.Is(err, auth.ErrUserInactive) {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error: "Account is inactive",
			})
			return
		}
		h.logger.Error("Login failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Login failed",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{refresh_token=string} true "Refresh token"
// @Success 200 {object} auth.TokenPair
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid refresh token request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Refresh tokens
	tokens, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.Warn("Token refresh failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Invalid or expired refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// GetProfile returns the current user's profile
// @Summary Get current user profile
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} auth.UserResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	profile, err := h.authService.GetProfile(userID)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "User not found",
			})
			return
		}
		h.logger.Error("Failed to get profile", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve profile",
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// Logout handles user logout
// @Summary Logout user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	if err := h.authService.Logout(userID); err != nil {
		h.logger.Error("Logout failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Logout failed",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Logged out successfully",
	})
}

// ChangePassword handles password change
// @Summary Change user password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body auth.ChangePasswordRequest true "Password change data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req auth.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid change password request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Details: err.Error(),
		})
		return
	}

	// Change password
	if err := h.authService.ChangePassword(userID, &req); err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "Invalid old password",
			})
			return
		}
		if errors.Is(err, auth.ErrWeakPassword) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "New password does not meet requirements (min 8 characters)",
			})
			return
		}
		h.logger.Error("Password change failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Password change failed",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Password changed successfully",
	})
}
