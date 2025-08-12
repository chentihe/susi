package model

import (
	"time"

	"gorm.io/gorm"
)

type Admin struct {
	ID           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name         string    `json:"username" gorm:"unique;not null"`
	PasswordHash string    `json:"password_hash"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	TOTPSecret   string    `json:"totp_secret" gorm:"column:totp_secret"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserRole string
type UserStatus string

const (
	RoleUser       UserRole = "user"
	RoleAdmin      UserRole = "admin"
	RoleSuperAdmin UserRole = "super_admin"
)

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
)

type User struct {
	ID        int        `json:"id" gorm:"primaryKey;autoIncrement"`
	Email     string     `json:"email" gorm:"uniqueIndex;not null;type:varchar(100)"`
	Password  string     `json:"-" gorm:"not null;type:varchar(255)"`
	Name      string     `json:"name" gorm:"not null;type:varchar(100)"`
	Phone     string     `json:"phone" gorm:"type:varchar(20)"`
	Role      UserRole   `json:"role" gorm:"type:varchar(20);default:'user'"`
	Status    UserStatus `json:"status" gorm:"type:varchar(20);default:'active'"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	CreatedBy int32 `json:"created_by"`
	UpdatedBy int32 `json:"updated_by"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Role == "" {
		u.Role = RoleUser
	}

	if u.Status == "" {
		u.Status = StatusActive
	}
	return nil
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}

func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

func (u *User) GetPermissions() []string {
	switch u.Role {
	case RoleSuperAdmin:
		return []string{
			"user:create", "user:read", "user:update", "user:delete",
			"admin:create", "admin:read", "admin:update", "admin:delete",
			"system:manage",
		}
	case RoleAdmin:
		return []string{
			"user:create", "user:read", "user:update", "user:delete",
		}
	case RoleUser:
		return []string{
			"profile:read", "profile:update",
		}
	default:
		return []string{}
	}
}

func (u *User) CanAccessAdminPanel() bool {
	return u.IsAdmin() && u.IsActive()
}
