package oauth

import (
	"fmt"
)

// ProviderFactory is a factory for creating OAuthProvider instances.
// It centralizes the provider creation logic.
var providers = make(map[string]func(ProviderConfig) OAuthProvider)

func init() {
	providers["google"] = newGoogleProvider
	providers["github"] = newGitHubProvider
	providers["facebook"] = newFacebookProvider
}

// GetProvider returns an OAuthProvider for the given provider name.
// It returns an error if the provider is not supported.
func GetProvider(providerName string, cfg ProviderConfig) (OAuthProvider, error) {
	factory, ok := providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider %s not supported", providerName)
	}
	return factory(cfg), nil
}