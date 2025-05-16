package oauth2

import (
	"context"
	database "core-auth/db"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// Storage implements the oauth2.Server.Storage interface
type Storage struct {
	rdb    *redis.Client
	db     *gorm.DB
	ctx    context.Context
}

// NewStorage creates a new OAuth2 storage implementation
func NewStorage(rdb *redis.Client, db *gorm.DB) *Storage {
	return &Storage{
		rdb: rdb,
		db:  db,
		ctx: context.Background(),
	}
}

// GetClient implements oauth2.Server.Storage interface
func (s *Storage) GetClient(clientID string) (oauth2.ClientInfo, error) {
	// Try Redis first if available
	if s.rdb != nil {
		key := redisClientPrefix + clientID
		data, err := s.rdb.Get(s.ctx, key).Bytes()
		if err == nil {
			var client models.Client
			if err := json.Unmarshal(data, &client); err == nil {
				return &client, nil
			}
		}
	}

	// Try database if Redis is down or not found
	var client database.OAuth2Client
	if err := s.db.Where("client_id = ? AND is_active = ?", clientID, true).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("client not found")
		}
		return nil, err
	}

	// Convert to oauth2.ClientInfo
	clientInfo := &models.Client{
		ID:     client.ClientID,
		Secret: client.ClientSecret,
		Domain: client.RedirectURIs,
		UserID: "",
	}

	// Cache in Redis if available
	if s.rdb != nil {
		data, err := json.Marshal(clientInfo)
		if err == nil {
			key := redisClientPrefix + clientID
			s.rdb.Set(s.ctx, key, data, 24*time.Hour)
		}
	}

	return clientInfo, nil
}

// SaveAuthorize implements oauth2.Server.Storage interface
func (s *Storage) SaveAuthorize(data *authorizeData) error {
	auth := &database.OAuth2Authorization{
		Code:        data.Code,
		ClientID:    data.Client.GetID(),
		UserID:      data.UserID,
		RedirectURI: data.RedirectURI,
		Scope:       data.Scope,
		ExpiresAt:   data.ExpiresAt,
		Used:        false,
	}

	// Store in Redis if available
	if s.rdb != nil {
		key := redisAuthCodePrefix + data.Code
		authData, err := json.Marshal(auth)
		if err == nil {
			ttl := time.Until(data.ExpiresAt)
			if ttl > 0 {
				s.rdb.SetEX(s.ctx, key, authData, ttl)
			}
		}
	}

	// Store in database
	return s.db.Create(auth).Error
}

// GetAuthorize implements oauth2.Server.Storage interface
func (s *Storage) GetAuthorize(code string) (*authorizeData, error) {
	// Try Redis first
	if s.rdb != nil {
		key := redisAuthCodePrefix + code
		data, err := s.rdb.Get(s.ctx, key).Bytes()
		if err == nil {
			var auth database.OAuth2Authorization
			if err := json.Unmarshal(data, &auth); err == nil {
				if time.Now().UTC().Before(auth.ExpiresAt) && !auth.Used {
					return s.convertToAuthorizeData(&auth)
				}
				s.rdb.Del(s.ctx, key)
			}
		}
	}

	// Try database if Redis is down or not found
	var auth database.OAuth2Authorization
	if err := s.db.Where("code = ? AND used = ? AND expires_at > ?", code, false, time.Now().UTC()).First(&auth).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid authorize code")
		}
		return nil, err
	}

	return s.convertToAuthorizeData(&auth)
}

// SaveAccess implements oauth2.Server.Storage interface
func (s *Storage) SaveAccess(data *database.OAuth2Token) error {
	token := &database.OAuth2Token{
		AccessToken:      data.AccessToken,
		RefreshToken:    data.RefreshToken,
		ClientID:        data.ClientID,
		UserID:          data.UserID,
		Scope:           data.Scope,
		AccessExpiresAt: data.AccessExpiresAt,
		RefreshExpiresAt: data.RefreshExpiresAt,
	}

	// Store in Redis if available
	if s.rdb != nil {
		// Store access token
		accessKey := redisAccessTokenPrefix + data.AccessToken
		accessData, err := json.Marshal(token)
		if err == nil {
			ttl := time.Until(data.AccessExpiresAt)
			if ttl > 0 {
				s.rdb.SetEX(s.ctx, accessKey, accessData, ttl)
			}
		}

		// Store refresh token if exists
		if data.RefreshToken != "" {
			refreshKey := redisRefreshTokenPrefix + data.RefreshToken
			refreshData, err := json.Marshal(token)
			if err == nil {
				ttl := time.Until(*data.RefreshExpiresAt)
				if ttl > 0 {
					s.rdb.SetEX(s.ctx, refreshKey, refreshData, ttl)
				}
			}
		}
	}

	// Store in database
	return s.db.Create(token).Error
}

// GetAccess implements oauth2.Server.Storage interface
func (s *Storage) GetAccess(token string) (*database.OAuth2Token, error) {
	// Try Redis first
	if s.rdb != nil {
		key := redisAccessTokenPrefix + token
		data, err := s.rdb.Get(s.ctx, key).Bytes()
		if err == nil {
			var oauthToken database.OAuth2Token
			if err := json.Unmarshal(data, &oauthToken); err == nil {
				if time.Now().UTC().Before(oauthToken.AccessExpiresAt) {
					return s.convertToTokenData(&oauthToken)
				}
				s.rdb.Del(s.ctx, key)
			}
		}
	}

	// Try database if Redis is down or not found
	var oauthToken database.OAuth2Token
	if err := s.db.Where("access_token = ? AND access_expires_at > ?", token, time.Now().UTC()).First(&oauthToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid access token")
		}
		return nil, err
	}

	return s.convertToTokenData(&oauthToken)
}

// GetRefresh implements oauth2.Server.Storage interface
func (s *Storage) GetRefresh(token string) (*database.OAuth2Token, error) {
	// Try Redis first
	if s.rdb != nil {
		key := redisRefreshTokenPrefix + token
		data, err := s.rdb.Get(s.ctx, key).Bytes()
		if err == nil {
			var oauthToken database.OAuth2Token
			if err := json.Unmarshal(data, &oauthToken); err == nil {
				if oauthToken.RefreshExpiresAt != nil && time.Now().UTC().Before(*oauthToken.RefreshExpiresAt) {
					return s.convertToTokenData(&oauthToken)
				}
				s.rdb.Del(s.ctx, key)
			}
		}
	}

	// Try database if Redis is down or not found
	var oauthToken database.OAuth2Token
	if err := s.db.Where("refresh_token = ? AND refresh_expires_at > ?", token, time.Now().UTC()).First(&oauthToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid refresh token")
		}
		return nil, err
	}

	return s.convertToTokenData(&oauthToken)
}

// helper functions
func (s *Storage) convertToAuthorizeData(auth *database.OAuth2Authorization) (*authorizeData, error) {
	client, err := s.GetClient(auth.ClientID)
	if err != nil {
		return nil, err
	}

	return &authorizeData{
		Client:      client,
		Code:        auth.Code,
		ExpiresAt:   auth.ExpiresAt,
		Scope:       auth.Scope,
		RedirectURI: auth.RedirectURI,
		UserID:      auth.UserID,
	}, nil
}

func (s *Storage) convertToTokenData(token *database.OAuth2Token) (*database.OAuth2Token, error) {
	return &database.OAuth2Token{
		ClientID:        token.ClientID,
		UserID:          token.UserID,
		AccessToken:     token.AccessToken,
		RefreshToken:    token.RefreshToken,
		Scope:           token.Scope,
		AccessExpiresAt: token.AccessExpiresAt,
		RefreshExpiresAt: token.RefreshExpiresAt,
	}, nil
} 