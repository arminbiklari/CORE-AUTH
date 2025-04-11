package token

import (
	"time"
	config "core-auth/config"
	"github.com/google/uuid"
)

func GenerateRefreshToken() (string, time.Time, error) {
	token_expiry := config.RefreshTokenExpiry
	tokenExpiry := time.Now().Add(token_expiry)
	token := uuid.New().String()
	return token, tokenExpiry, nil
}

func GenerateAccessToken() (string, time.Time, error) {
	token_expiry := config.AccessTokenExpiry
	tokenExpiry := time.Now().Add(token_expiry)
	token := uuid.New().String()
	return token, tokenExpiry, nil
}

func GenerateRandomString(length int) (string, error) {
	token := uuid.New().String()
	return token[:length], nil
}


// Helper functions
func generateAuthCode() string {
	// TODO: Implement secure random code generation
	return "auth_" + time.Now().Format("20060102150405")
}

func generateAccessToken() string {
	// TODO: Implement secure random token generation
	return "access_" + time.Now().Format("20060102150405")
}

func generateRefreshToken() string {
	// TODO: Implement secure random token generation
	return "refresh_" + time.Now().Format("20060102150405")
}
