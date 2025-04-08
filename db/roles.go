package database

import (
	"errors"
	"gorm.io/gorm"
)

const (
	RoleAdmin     uint = 1
	RoleUser      uint = 2
	RoleModerator uint = 3
)

// DefaultRoles defines the default roles in the system
var DefaultRoles = []Role{
	{
		Model: gorm.Model{ID: RoleAdmin},
		Name:  "admin",
	},
	{
		Model: gorm.Model{ID: RoleUser},
		Name:  "user",
	},
	{
		Model: gorm.Model{ID: RoleModerator},
		Name:  "moderator",
	},
}

// InitializeRoles creates default roles if they don't exist
func InitializeRoles(db *gorm.DB) error {
	for _, role := range DefaultRoles {
		var existingRole Role
		if err := db.Where("id = ?", role.ID).First(&existingRole).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&role).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	return nil
}

// GetDefaultUserRole returns the default role ID for new users
func GetDefaultUserRole() uint {
	return RoleUser
}

// ValidateRoleID checks if a role ID is valid
func ValidateRoleID(db *gorm.DB, roleID uint) error {
	var role Role
	if err := db.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid role ID")
		}
		return err
	}
	return nil
} 