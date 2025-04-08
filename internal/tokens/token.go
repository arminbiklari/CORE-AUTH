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
