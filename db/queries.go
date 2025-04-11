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

func CreateSession(db *gorm.DB, session *Session) error {
	return db.Create(session).Error
}

func GetActiveSession(db *gorm.DB, sessionID string) (*Session, error) {
	var session Session
	err := db.Where("session_id = ? AND is_active = ? AND expires_at > ?", 
		sessionID, true, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func UpdateSessionActivity(db *gorm.DB, sessionID string) error {
	return db.Model(&Session{}).
		Where("session_id = ?", sessionID).
		Update("last_activity", time.Now()).Error
}

func InvalidateSession(db *gorm.DB, sessionID string) error {
	return db.Model(&Session{}).
		Where("session_id = ?", sessionID).
		Update("is_active", false).Error
}

// OAuth2 Functions
func CreateAuthorizationCode(db *gorm.DB, auth *OAuth2Authorization) error {
	return db.Create(auth).Error
}

func UpsertAuthorizationCode(db *gorm.DB, auth *OAuth2Authorization) error {
	return db.Save(auth).Error
}

func GetValidAuthorizationCode(db *gorm.DB, code string) (*OAuth2Authorization, error) {
	var auth OAuth2Authorization
	err := db.Where("code = ? AND used = ? AND expires_at > ?", 
		code, false, time.Now()).First(&auth).Error
	if err != nil {
		return nil, err
	}
	return &auth, nil
}

func DeleteAuthorizationCode(db *gorm.DB, code string) error {
	return db.Where("code = ?", code).Delete(&OAuth2Authorization{}).Error
}

func MarkAuthorizationCodeUsed(db *gorm.DB, code string) error {
	return db.Model(&OAuth2Authorization{}).
		Where("code = ?", code).
		Update("used", true).Error
}

func UpsertToken(db *gorm.DB, token *OAuth2Token) error {
	return db.Save(token).Error
}

func GetValidTokenByAccess(db *gorm.DB, accessToken string) (*OAuth2Token, error) {
	var token OAuth2Token
	err := db.Where("access_token = ? AND access_expires_at > ?", 
		accessToken, time.Now()).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func GetValidTokenByRefresh(db *gorm.DB, refreshToken string) (*OAuth2Token, error) {
	var token OAuth2Token
	err := db.Where("refresh_token = ? AND refresh_expires_at > ?", 
		refreshToken, time.Now()).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func DeleteTokenByAccess(db *gorm.DB, accessToken string) error {
	return db.Where("access_token = ?", accessToken).Delete(&OAuth2Token{}).Error
}

func GetClientByID(db *gorm.DB, clientID string) (*OAuth2Client, error) {
	var client OAuth2Client
	err := db.Where("client_id = ?", clientID).First(&client).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

// OAuth2Queries contains all OAuth2 related database queries
type OAuth2Queries struct {
	db *gorm.DB
}

// NewOAuth2Queries creates a new OAuth2Queries instance
func NewOAuth2Queries(db *gorm.DB) *OAuth2Queries {
	return &OAuth2Queries{db: db}
}

// StoreAuthorizationCode stores an authorization code
func (q *OAuth2Queries) StoreAuthorizationCode(auth *OAuth2Authorization) error {
	return q.db.Create(auth).Error
}

// GetAuthorizationCode retrieves an authorization code
func (q *OAuth2Queries) GetAuthorizationCode(code string) (*OAuth2Authorization, error) {
	var auth OAuth2Authorization
	err := q.db.Where("code = ? AND used = ?", code, false).First(&auth).Error
	if err != nil {
		return nil, err
	}
	return &auth, nil
}

// MarkAuthorizationCodeAsUsed marks an authorization code as used
func (q *OAuth2Queries) MarkAuthorizationCodeAsUsed(code string) error {
	return q.db.Model(&OAuth2Authorization{}).
		Where("code = ?", code).
		Update("used", true).
		Error
}

// StoreToken stores an OAuth2 token
func (q *OAuth2Queries) StoreToken(token *OAuth2Token) error {
	return q.db.Create(token).Error
}

// GetTokenByAccess retrieves a token by access token
func (q *OAuth2Queries) GetTokenByAccess(accessToken string) (*OAuth2Token, error) {
	var token OAuth2Token
	err := q.db.Where("access_token = ?", accessToken).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetTokenByRefresh retrieves a token by refresh token
func (q *OAuth2Queries) GetTokenByRefresh(refreshToken string) (*OAuth2Token, error) {
	var token OAuth2Token
	err := q.db.Where("refresh_token = ?", refreshToken).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteToken deletes a token by access token
func (q *OAuth2Queries) DeleteToken(accessToken string) error {
	return q.db.Where("access_token = ?", accessToken).Delete(&OAuth2Token{}).Error
}

// GetClient retrieves client information
func (q *OAuth2Queries) GetClient(clientID string) (*OAuth2Client, error) {
	var client OAuth2Client
	err := q.db.Where("client_id = ? AND is_active = ?", clientID, true).First(&client).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

// CleanupExpiredTokens removes expired tokens
func (q *OAuth2Queries) CleanupExpiredTokens() error {
	now := time.Now()
	return q.db.Where("access_expires_at < ?", now).Delete(&OAuth2Token{}).Error
}

// CleanupExpiredAuthorizationCodes removes expired authorization codes
func (q *OAuth2Queries) CleanupExpiredAuthorizationCodes() error {
	now := time.Now()
	return q.db.Where("expires_at < ?", now).Delete(&OAuth2Authorization{}).Error
}