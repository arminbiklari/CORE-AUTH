package config

import (
	"time"
)

// set to 5 days
var RefreshTokenExpiry = 5 * 24 * time.Hour
var AccessTokenExpiry = 1 * time.Hour
