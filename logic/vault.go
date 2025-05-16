package logic

import (
	"context"
	"core-auth/config"
	"core-auth/vault"
	"log"
	"sync"
	"time"
)

var (
	dbCredentialsMutex sync.RWMutex
	dbUsername         string
	dbPassword         string
	credentialsReady   = make(chan struct{})
	initialized        bool
	initOnce           sync.Once
	// Function to register credential provider, set by the database package
	registerProvider   func(provider interface{}) = nil
)

// VaultCredentialProvider implements a credential provider interface
type VaultCredentialProvider struct{}

// GetCredentials implements the credential provider interface
func (p *VaultCredentialProvider) GetCredentials() (string, string) {
	return GetDatabaseCredentials()
}

// SetRegisterProviderFunc allows the database package to set the registration function
// This avoids import cycles
func SetRegisterProviderFunc(fn func(provider interface{})) {
	registerProvider = fn
}

// GetDatabaseCredentials returns the current database credentials
// If credentials aren't ready, it blocks until they are
func GetDatabaseCredentials() (string, string) {
	// If credentials aren't initialized yet, wait for them
	if !initialized {
		<-credentialsReady
	}
	
	dbCredentialsMutex.RLock()
	defer dbCredentialsMutex.RUnlock()
	return dbUsername, dbPassword
}

// InitializeVaultCredentials fetches credentials from Vault once
// Returns true if credentials were successfully initialized
func InitializeVaultCredentials(cfg *config.Config) bool {
	var success bool
	
	initOnce.Do(func() {
		if cfg.Vault.Token == "" {
			log.Println("Vault token not provided, skipping credential initialization")
			close(credentialsReady) // Unblock any waiters
			return
		}

		// Create Vault client
		vaultClient, err := vault.NewVaultClient(cfg)
		if err != nil {
			log.Printf("Failed to create Vault client: %v", err)
			close(credentialsReady) // Unblock any waiters
			return
		}

		// Initial credentials fetch
		if err := renewDatabaseCredentials(vaultClient); err != nil {
			log.Printf("Initial database credential fetch failed: %v", err)
			close(credentialsReady) // Unblock any waiters
			return
		}

		initialized = true
		success = true
		
		// Register ourselves as the credential provider if possible
		if registerProvider != nil {
			registerProvider(&VaultCredentialProvider{})
			log.Println("Registered Vault credential provider with database package")
		}
		
		close(credentialsReady) // Unblock any waiters
		log.Println("Database credentials successfully initialized from Vault")
	})
	
	return success
}

// WatchDatabaseCredentials manages database credentials from Vault
// It periodically renews credentials before the lease expires
func WatchDatabaseCredentials(cfg *config.Config) {
	if cfg.Vault.Token == "" {
		log.Println("Vault token not provided, skipping credential rotation")
		return
	}

	// Initialize if not already done
	if !initialized {
		if !InitializeVaultCredentials(cfg) {
			return
		}
	}

	leaseDuration := time.Duration(cfg.Vault.LeaseDuration) * time.Second
	renewalInterval := leaseDuration / 2

	log.Printf("Starting database credential rotation with lease duration %v, renewal every %v", 
		leaseDuration, renewalInterval)

	// Create Vault client
	vaultClient, err := vault.NewVaultClient(cfg)
	if err != nil {
		log.Printf("Failed to create Vault client: %v", err)
		return
	}

	// Start periodic renewal (skip initial fetch since we already have credentials)
	time.Sleep(renewalInterval)
	
	for {
		if err := renewDatabaseCredentials(vaultClient); err != nil {
			log.Printf("Failed to renew database credentials: %v", err)
		}
		time.Sleep(renewalInterval)
	}
}

// renewDatabaseCredentials gets new credentials from Vault and updates the database connection
func renewDatabaseCredentials(vaultClient *vault.VaultClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get new credentials from Vault
	username, password, err := vaultClient.GetDatabaseCredentials(ctx)
	if err != nil {
		return err
	}

	// Update global credentials
	dbCredentialsMutex.Lock()
	dbUsername = username
	dbPassword = password
	dbCredentialsMutex.Unlock()

	log.Println("Database credentials successfully renewed")
	
	// We can't directly call RenewConnection because of import cycles
	// This is handled by the database package watching for credential changes
	
	return nil
}