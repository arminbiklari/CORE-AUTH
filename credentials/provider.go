package credentials

// Provider defines the interface for getting database credentials
type Provider interface {
	GetCredentials() (username string, password string)
}

// ProviderFunc is a helper type for creating providers
type ProviderFunc func() (string, string)

// GetCredentials implements Provider
func (f ProviderFunc) GetCredentials() (string, string) {
	return f()
}

var (
	// defaultProvider is the registered provider
	defaultProvider Provider

	// EnvironmentProvider creates a credential provider from environment variables
	EnvironmentProvider = ProviderFunc(func() (string, string) {
		// This is a placeholder - actual implementation
		// would read environment variables
		return "", ""
	})
)

// SetProvider registers a credentials provider
func SetProvider(provider Provider) {
	defaultProvider = provider
}

// GetProvider returns the current provider
func GetProvider() Provider {
	return defaultProvider
}

// Get retrieves credentials from the current provider
func Get() (string, string) {
	if defaultProvider == nil {
		return "", ""
	}
	return defaultProvider.GetCredentials()
} 