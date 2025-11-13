package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole represents user roles in the system
type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleUser   UserRole = "user"
	RoleViewer UserRole = "viewer"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username     string     `gorm:"size:50;uniqueIndex;not null" json:"username" validate:"required,min=3,max=50"`
	Email        string     `gorm:"size:100;uniqueIndex;not null" json:"email" validate:"required,email"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Role         UserRole   `gorm:"type:varchar(20);default:user" json:"role"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Relations
	Servers []Server `gorm:"foreignKey:UserID" json:"servers,omitempty"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook for User
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// IsAdmin checks if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// CanManageServer checks if user can manage a specific server
func (u *User) CanManageServer(serverID uuid.UUID) bool {
	if u.IsAdmin() {
		return true
	}
	for _, server := range u.Servers {
		if server.ID == serverID {
			return true
		}
	}
	return false
}
