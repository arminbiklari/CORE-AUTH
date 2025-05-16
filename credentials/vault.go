package credentials

import (
	"context"
	"core-auth/config"
	"core-auth/vault"
	"log"
	"sync"
	"time"
)

// VaultProvider manages retrieving and refreshing credentials from Vault
type VaultProvider struct {
	mu          sync.RWMutex
	username    string
	password    string
	initialized bool
	readyChan   chan struct{}
	config      *config.Config
}

// NewVaultProvider creates a new Vault credential provider
func NewVaultProvider(cfg *config.Config) *VaultProvider {
	return &VaultProvider{
		readyChan: make(chan struct{}),
		config:    cfg,
	}
}

// GetCredentials implements the Provider interface
func (v *VaultProvider) GetCredentials() (string, string) {
	// Block until credentials are ready
	if !v.initialized {
		<-v.readyChan
	}
	
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.username, v.password
}

// Initialize fetches initial credentials from Vault
func (v *VaultProvider) Initialize() error {
	// Create a Vault client
	vaultClient, err := vault.NewVaultClient(v.config)
	if err != nil {
		return err
	}
	
	// Get credentials
	err = v.refreshCredentials(vaultClient)
	if err != nil {
		return err
	}
	
	// Mark as initialized and close ready channel
	v.initialized = true
	close(v.readyChan)
	
	// Register as the global provider
	SetProvider(v)
	
	return nil
}

// StartRotation begins periodic credential rotation
func (v *VaultProvider) StartRotation() {
	if !v.initialized {
		if err := v.Initialize(); err != nil {
			log.Printf("Failed to initialize Vault credentials: %v", err)
			return
		}
	}
	
	// Calculate rotation interval (half the lease duration)
	leaseDuration := time.Duration(v.config.Vault.LeaseDuration) * time.Second
	renewInterval := leaseDuration / 2
	
	log.Printf("Starting credential rotation with lease duration %v, renewal every %v",
		leaseDuration, renewInterval)
	
	// Create Vault client for rotations
	vaultClient, err := vault.NewVaultClient(v.config)
	if err != nil {
		log.Printf("Failed to create Vault client for rotation: %v", err)
		return
	}
	
	// Wait for first interval
	time.Sleep(renewInterval)
	
	// Start rotation loop
	for {
		if err := v.refreshCredentials(vaultClient); err != nil {
			log.Printf("Failed to refresh credentials: %v", err)
		}
		time.Sleep(renewInterval)
	}
}

// refreshCredentials fetches new credentials from Vault
func (v *VaultProvider) refreshCredentials(client *vault.VaultClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	username, password, err := client.GetDatabaseCredentials(ctx)
	if err != nil {
		return err
	}
	
	v.mu.Lock()
	v.username = username
	v.password = password
	v.mu.Unlock()
	
	log.Println("Refreshed database credentials from Vault")
	return nil
}

// RegisterEnvironmentProvider creates and registers an environment-based provider
func RegisterEnvironmentProvider(cfg *config.Config) {
	SetProvider(ProviderFunc(func() (string, string) {
		return cfg.Database.User, cfg.Database.Password
	}))
} 