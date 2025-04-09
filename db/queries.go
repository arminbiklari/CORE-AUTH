package database

import (
	"errors"
	"time"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateUser creates a new user in the database
func CreateUser(db *gorm.DB, user *User) error {
	// Hash the password before creating the user
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	
	return db.Create(user).Error
}

// GetUserByUsername retrieves a user by their username	
func GetUserByUsername(db *gorm.DB, username string) (*User, error) {
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by their email
func GetUserByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(db *gorm.DB, id uint) (*User, error) {
	var user User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates user information
func UpdateUser(db *gorm.DB, user *User) error {
	return db.Save(user).Error
}

// DeleteUser deletes a user by their ID
func DeleteUser(db *gorm.DB, id uint) error {
	return db.Delete(&User{}, id).Error
}

func CheckPasswordDB(db *gorm.DB, username, password string) bool {
	var user User
	// fmt.Println("password is ", password)
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return false
	}
	// fmt.Println("user.Password is ", user.Password)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// fmt.Println("Error comparing passwords:", err)
		return false
	}
	return true
}

func CheckUsernameDB(db *gorm.DB, username string) bool {
	var user User
	db.First(&user, "username = ?", username)
	return user.Username == username
}

func UpdateRefreshToken(db *gorm.DB, username, refreshToken string, tokenExpiry time.Time) error {
	return db.Model(&User{}).Where("username = ?", username).Updates(User{
		RefreshToken: refreshToken,
		RefreshTokenExpiry: &tokenExpiry,
	}).Error
}

func CheckRefreshToken(db *gorm.DB, refreshToken string) bool {
	var user User
	db.First(&user, "refresh_token = ?", refreshToken)
	return user.RefreshToken == refreshToken
}

func StoreRefreshToken(db *gorm.DB, username, refreshToken string, tokenExpiry time.Time) error {
	return db.Model(&User{}).Where("username = ?", username).Updates(User{
		RefreshToken: refreshToken,
		RefreshTokenExpiry: &tokenExpiry,
	}).Error
}

func StoreAccessToken(db *gorm.DB, username, accessToken string, tokenExpiry time.Time) error {
	return db.Model(&User{}).Where("username = ?", username).Updates(User{
		AccessToken: accessToken,
		AccessTokenExpiry: &tokenExpiry,
	}).Error
}

func ValidateAccessToken(db *gorm.DB, accessToken string) (bool, error) {
	var user User
	db.First(&user, "access_token = ?", accessToken)
	return user.AccessToken == accessToken, nil
}

func DeleteRefreshToken(db *gorm.DB, refreshToken string) error {
	return db.Model(&User{}).Where("refresh_token = ?", refreshToken).Updates(User{
		RefreshToken: "",
		RefreshTokenExpiry: nil,
	}).Error
}
