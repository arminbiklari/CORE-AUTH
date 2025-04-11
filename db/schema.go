package database

import (
	"time"

	"gorm.io/gorm"
)

// User represents the user model in the database
type User struct {
	gorm.Model
	Username     string     `gorm:"type:varchar(100);unique;not null"`
	Email        string     `gorm:"type:varchar(100);unique;not null"`
	Password     string     `gorm:"not null"`
	RoleID       uint       `gorm:"not null"`
	Role         Role       `gorm:"foreignKey:RoleID"`
	LastLogin    *time.Time
	IsActive     bool       `gorm:"default:true"`
	RefreshToken string     `gorm:"type:varchar(255)"`
	RefreshTokenExpiry  *time.Time `gorm:"default:null"`
	AccessToken  string     `gorm:"type:varchar(255)"`
	AccessTokenExpiry  *time.Time `gorm:"default:null"`
}

// Role represents user roles in the system
type Role struct {
	gorm.Model
	Name        string       `gorm:"type:varchar(50);unique;not null"`
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

// Permission represents available permissions in the system
type Permission struct {
	gorm.Model
	Name        string `gorm:"type:varchar(100);unique;not null"`
	Description string `gorm:"type:varchar(255)"`
}

// Session represents user sessions
type Session struct {
	gorm.Model
	UserID    uint      `gorm:"not null"`
	Token     string    `gorm:"type:varchar(255);not null"`
	ExpiresAt time.Time `gorm:"not null"`
	IP        string    `gorm:"type:varchar(45)"`
	UserAgent string    `gorm:"type:varchar(255)"`
}
// OAuth2Client represents registered applications
type OAuth2Client struct {
	gorm.Model
	ClientID     string `gorm:"type:varchar(100);unique;not null"`
	ClientSecret string `gorm:"type:varchar(100);not null"`
	Name         string `gorm:"type:varchar(200);not null"`
	RedirectURIs string `gorm:"type:text;not null"` // JSON array of allowed redirect URIs
	GrantTypes   string `gorm:"type:text;not null"` // JSON array of allowed grant types
	Scopes       string `gorm:"type:text;not null"` // JSON array of allowed scopes
	IsActive     bool   `gorm:"default:true"`
}

// OAuth2ation represents authorization codes
type OAuth2Authorization struct {
	gorm.Model
	Code        string    `gorm:"type:varchar(100);unique;not null"`
	ClientID    string    `gorm:"type:varchar(100);not null"`
	UserID      uint      `gorm:"not null"`
	RedirectURI string    `gorm:"type:varchar(500);not null"`
	Scope       string    `gorm:"type:varchar(500)"`
	ExpiresAt   time.Time `gorm:"not null"`
	Used        bool      `gorm:"default:false"`
}

// OAuth2Token represents access and refresh tokens
type OAuth2Token struct {
	gorm.Model
	AccessToken      string     `gorm:"type:varchar(100);unique;not null"`
	RefreshToken    string     `gorm:"type:varchar(100);unique"`
	ClientID        string     `gorm:"type:varchar(100);not null"`
	UserID          uint       `gorm:"not null"`
	Scope           string     `gorm:"type:varchar(500)"`
	AccessExpiresAt time.Time  `gorm:"not null"`
	RefreshExpiresAt *time.Time
} 

// AutoMigrate performs database auto migration for the schema
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Role{},
		&Permission{},
		&Session{},
		&OAuth2Client{},
		&OAuth2Authorization{},
		&OAuth2Token{},
	)
}
