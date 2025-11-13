package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/aymc/backend/database"
	"github.com/aymc/backend/database/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrWeakPassword       = errors.New("password does not meet requirements")
)

// RegisterRequest represents user registration data
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=100"`
	Role     string `json:"role,omitempty" validate:"omitempty,oneof=admin user viewer"`
}

// LoginRequest represents user login data
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the login response with tokens and user info
type LoginResponse struct {
	User   UserResponse `json:"user"`
	Tokens TokenPair    `json:"tokens"`
}

// UserResponse represents user data in responses
type UserResponse struct {
	ID        uuid.UUID  `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	IsActive  bool       `json:"is_active"`
	LastLogin *time.Time `json:"last_login,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// ChangePasswordRequest represents password change data
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=100"`
}

// AuthService handles authentication business logic
type AuthService struct {
	jwtService *JWTService
	logger     *zap.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(jwtService *JWTService, logger *zap.Logger) *AuthService {
	return &AuthService{
		jwtService: jwtService,
		logger:     logger,
	}
}

// Register creates a new user account
func (s *AuthService) Register(req *RegisterRequest) (*UserResponse, error) {
	db := database.GetDB()

	// Check if user already exists
	var existingUser models.User
	if err := db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		s.logger.Warn("User registration failed - user already exists",
			zap.String("email", req.Email),
			zap.String("username", req.Username),
		)
		return nil, ErrUserExists
	}

	// Validate password strength
	if len(req.Password) < 8 {
		return nil, ErrWeakPassword
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set default role if not provided
	role := models.RoleUser
	if req.Role != "" {
		role = models.UserRole(req.Role)
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		IsActive:     true,
	}

	if err := db.Create(user).Error; err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("User registered successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
		zap.String("role", string(user.Role)),
	)

	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
	}, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	db := database.GetDB()

	// Find user by email
	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Login failed - user not found", zap.String("email", req.Email))
			return nil, ErrInvalidCredentials
		}
		s.logger.Error("Failed to query user", zap.Error(err))
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		s.logger.Warn("Login failed - user inactive",
			zap.String("user_id", user.ID.String()),
			zap.String("email", user.Email),
		)
		return nil, ErrUserInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.logger.Warn("Login failed - invalid password", zap.String("email", req.Email))
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Username, user.Email, string(user.Role))
	if err != nil {
		return nil, err
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	if err := db.Model(&user).Update("last_login", now).Error; err != nil {
		s.logger.Warn("Failed to update last login", zap.Error(err))
		// Don't fail the login for this
	}

	s.logger.Info("User logged in successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
	)

	return &LoginResponse{
		User: UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      string(user.Role),
			IsActive:  user.IsActive,
			LastLogin: user.LastLogin,
			CreatedAt: user.CreatedAt,
		},
		Tokens: *tokens,
	}, nil
}

// RefreshToken generates new tokens from a refresh token
func (s *AuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	tokens, err := s.jwtService.RefreshAccessToken(refreshToken)
	if err != nil {
		s.logger.Warn("Token refresh failed", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("Tokens refreshed successfully")
	return tokens, nil
}

// GetProfile retrieves user profile by ID
func (s *AuthService) GetProfile(userID uuid.UUID) (*UserResponse, error) {
	db := database.GetDB()

	var user models.User
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		s.logger.Error("Failed to query user", zap.Error(err))
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		LastLogin: user.LastLogin,
		CreatedAt: user.CreatedAt,
	}, nil
}

// ChangePassword changes user password
func (s *AuthService) ChangePassword(userID uuid.UUID, req *ChangePasswordRequest) error {
	db := database.GetDB()

	// Get user
	var user models.User
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to query user: %w", err)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		s.logger.Warn("Change password failed - invalid old password",
			zap.String("user_id", userID.String()),
		)
		return ErrInvalidCredentials
	}

	// Validate new password
	if len(req.NewPassword) < 8 {
		return ErrWeakPassword
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash new password", zap.Error(err))
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := db.Model(&user).Update("password_hash", string(hashedPassword)).Error; err != nil {
		s.logger.Error("Failed to update password", zap.Error(err))
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.logger.Info("Password changed successfully", zap.String("user_id", userID.String()))
	return nil
}

// Logout performs logout (token invalidation would be implemented with Redis)
func (s *AuthService) Logout(userID uuid.UUID) error {
	// In a production system, you would:
	// 1. Add the token to a blacklist in Redis
	// 2. Set expiry to match token expiry
	// For now, we just log the event
	s.logger.Info("User logged out", zap.String("user_id", userID.String()))
	return nil
}
