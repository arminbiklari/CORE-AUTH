package database

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// BeforeCreate is a GORM hook that hashes the password before creating the user
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

// CheckPassword verifies if the provided password matches the hashed password
func (u *User) CheckPassword(username, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// UpdatePassword updates user's password with a new hashed password
func UpdatePassword(db *gorm.DB, userID uint, newPassword string) error {
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	
	return db.Model(&User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

func (u *User) CheckActive(db *gorm.DB) bool {
	var user User
	if err := db.First(&user, u.ID).Error; err != nil {
		return false
	}
	return user.IsActive
}