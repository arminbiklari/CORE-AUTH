package oauth2

import (
	database "core-auth/db"
	"time"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	// Redis key prefixes
	redisAuthCodePrefix    = "oauth2:authcode:"
	redisAccessTokenPrefix = "oauth2:accesstoken:"
	redisRefreshTokenPrefix = "oauth2:refreshtoken:"
	redisClientPrefix      = "oauth2:client:"
)

type authorizeData struct {
	Client      oauth2.ClientInfo
	Code        string
	ExpiresAt   time.Time
	Scope       string
	RedirectURI string
	UserID      uint
}

type tokenData struct {
	Client            oauth2.ClientInfo
	UserID            uint
	AccessToken       string
	RefreshToken      string
	Scope             string
	AccessExpiresAt   time.Time
	RefreshExpiresAt  time.Time
}

// Manager handles OAuth2 token and authorization code storage with Redis and database
type Manager struct {
	storage *Storage
	queries *database.OAuth2Queries
}

// NewManager creates a new OAuth2 manager
func NewManager(rdb *redis.Client, db *gorm.DB) *Manager {
	return &Manager{
		storage: NewStorage(rdb, db),
		queries: database.NewOAuth2Queries(db),
	}
}

// GetClient retrieves client information
func (m *Manager) GetClient(clientID string) (oauth2.ClientInfo, error) {
	return m.storage.GetClient(clientID)
}

// StoreAuthorizationCode stores an authorization code
func (m *Manager) StoreAuthorizationCode(auth *database.OAuth2Authorization) error {
	client, err := m.storage.GetClient(auth.ClientID)
	if err != nil {
		return err
	}

	authorizeData := &authorizeData{
		Client:      client,
		Code:        auth.Code,
		ExpiresAt:   auth.ExpiresAt,
		Scope:       auth.Scope,
		RedirectURI: auth.RedirectURI,
		UserID:      auth.UserID,
	}

	if err := m.queries.StoreAuthorizationCode(auth); err != nil {
		return err
	}

	return m.storage.SaveAuthorize(authorizeData)
}

// GetAuthorizationCode retrieves an authorization code
func (m *Manager) GetAuthorizationCode(code string) (*database.OAuth2Authorization, error) {
	return m.queries.GetAuthorizationCode(code)
}

// DeleteAuthorizationCode deletes an authorization code
func (m *Manager) DeleteAuthorizationCode(code string) error {
	return m.queries.MarkAuthorizationCodeAsUsed(code)
}

// StoreToken stores an access token
func (m *Manager) StoreToken(token *database.OAuth2Token) error {
	client, err := m.storage.GetClient(token.ClientID)
	if err != nil {
		return err
	}

	tokenData := &tokenData{
		Client:            client,
		UserID:            token.UserID,
		AccessToken:       token.AccessToken,
		RefreshToken:      token.RefreshToken,
		Scope:             token.Scope,
		AccessExpiresAt:   token.AccessExpiresAt,
		RefreshExpiresAt:  *token.RefreshExpiresAt,
	}

	if err := m.queries.StoreToken(token); err != nil {
		return err
	}

	// Convert tokenData to OAuth2Token for storage
	storageToken := &database.OAuth2Token{
		AccessToken:      tokenData.AccessToken,
		RefreshToken:    tokenData.RefreshToken,
		ClientID:        tokenData.Client.GetID(),
		UserID:          tokenData.UserID,
		Scope:           tokenData.Scope,
		AccessExpiresAt: tokenData.AccessExpiresAt,
		RefreshExpiresAt: &tokenData.RefreshExpiresAt,
	}

	return m.storage.SaveAccess(storageToken)
}

// GetTokenByAccess retrieves an access token
func (m *Manager) GetTokenByAccess(accessToken string) (*database.OAuth2Token, error) {
	return m.queries.GetTokenByAccess(accessToken)
}

// GetTokenByRefresh retrieves a refresh token
func (m *Manager) GetTokenByRefresh(refreshToken string) (*database.OAuth2Token, error) {
	return m.queries.GetTokenByRefresh(refreshToken)
}

// DeleteToken deletes a token
func (m *Manager) DeleteToken(accessToken string) error {
	return m.queries.DeleteToken(accessToken)
} 