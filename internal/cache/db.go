package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type DB struct {
	AuthorizationCodeKey string
	AccessTokenKey      string
	RefreshTokenKey     string
	SessionIDKey        string
	ClientDetailsKey    string
	StateKey           string
	client             *redis.Client
}

// AuthorizationCode represents the data stored for an authorization code
type AuthorizationCode struct {
	Code        string
	ClientID    string
	RedirectURI string
	UserID      string
	Scopes      []string
	ExpiresAt   time.Time
}

// AccessToken represents the data stored for an access token
type AccessToken struct {
	Token     string
	UserID    string
	ClientID  string
	Scopes    []string
	ExpiresAt time.Time
}

// RefreshToken represents the data stored for a refresh token
type RefreshToken struct {
	Token         string
	UserID        string
	ClientID      string
	Scopes        []string
	ExpiresAt     time.Time
	AccessTokenID string
}

// ClientDetails represents the data stored for client information
type ClientDetails struct {
	ID           string
	Secret       string
	RedirectURIs []string
	GrantTypes   []string
}

// Session represents the data stored for a user session
type Session struct {
	ID            string
	UserID        string
	Authenticated time.Time
	Data          map[string]interface{}
}

// State represents the data stored for CSRF prevention
type State struct {
	Value     string
	Timestamp time.Time
	UserID    string
}

// StoreAuthorizationCode stores an authorization code in Redis
func (db *DB) StoreAuthorizationCode(ctx context.Context, code *AuthorizationCode, expiration time.Duration) error {
	key := db.AuthorizationCodeKey + code.Code
	data, err := json.Marshal(code)
	if err != nil {
		return err
	}
	return db.client.Set(ctx, key, data, expiration).Err()
}

// GetAuthorizationCode retrieves an authorization code from Redis
func (db *DB) GetAuthorizationCode(ctx context.Context, code string) (*AuthorizationCode, error) {
	key := db.AuthorizationCodeKey + code
	data, err := db.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var authCode AuthorizationCode
	err = json.Unmarshal(data, &authCode)
	return &authCode, err
}

// DeleteAuthorizationCode removes an authorization code from Redis
func (db *DB) DeleteAuthorizationCode(ctx context.Context, code string) error {
	key := db.AuthorizationCodeKey + code
	return db.client.Del(ctx, key).Err()
}

// StoreAccessToken stores an access token in Redis
func (db *DB) StoreAccessToken(ctx context.Context, token *AccessToken, expiration time.Duration) error {
	key := db.AccessTokenKey + token.Token
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return db.client.Set(ctx, key, data, expiration).Err()
}

// GetAccessToken retrieves an access token from Redis
func (db *DB) GetAccessToken(ctx context.Context, token string) (*AccessToken, error) {
	key := db.AccessTokenKey + token
	data, err := db.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var accessToken AccessToken
	err = json.Unmarshal(data, &accessToken)
	return &accessToken, err
}

// StoreRefreshToken stores a refresh token in Redis
func (db *DB) StoreRefreshToken(ctx context.Context, token *RefreshToken, expiration time.Duration) error {
	key := db.RefreshTokenKey + token.Token
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return db.client.Set(ctx, key, data, expiration).Err()
}

// GetRefreshToken retrieves a refresh token from Redis
func (db *DB) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	key := db.RefreshTokenKey + token
	data, err := db.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var refreshToken RefreshToken
	err = json.Unmarshal(data, &refreshToken)
	return &refreshToken, err
}

// StoreClientDetails stores client details in Redis
func (db *DB) StoreClientDetails(ctx context.Context, client *ClientDetails) error {
	key := db.ClientDetailsKey + client.ID
	data, err := json.Marshal(client)
	if err != nil {
		return err
	}
	return db.client.Set(ctx, key, data, 0).Err() // No expiration for client details
}

// GetClientDetails retrieves client details from Redis
func (db *DB) GetClientDetails(ctx context.Context, clientID string) (*ClientDetails, error) {
	key := db.ClientDetailsKey + clientID
	data, err := db.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var client ClientDetails
	err = json.Unmarshal(data, &client)
	return &client, err
}

// StoreSession stores a session in Redis
func (db *DB) StoreSession(ctx context.Context, session *Session, expiration time.Duration) error {
	key := db.SessionIDKey + session.ID
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return db.client.Set(ctx, key, data, expiration).Err()
}

// GetSession retrieves a session from Redis
func (db *DB) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	key := db.SessionIDKey + sessionID
	data, err := db.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var session Session
	err = json.Unmarshal(data, &session)
	return &session, err
}

// StoreState stores a state parameter in Redis
func (db *DB) StoreState(ctx context.Context, state *State, expiration time.Duration) error {
	key := db.StateKey + state.Value
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return db.client.Set(ctx, key, data, expiration).Err()
}

// GetState retrieves a state parameter from Redis
func (db *DB) GetState(ctx context.Context, stateValue string) (*State, error) {
	key := db.StateKey + stateValue
	data, err := db.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var state State
	err = json.Unmarshal(data, &state)
	return &state, err
}

// DeleteState removes a state parameter from Redis
func (db *DB) DeleteState(ctx context.Context, stateValue string) error {
	key := db.StateKey + stateValue
	return db.client.Del(ctx, key).Err()
}
