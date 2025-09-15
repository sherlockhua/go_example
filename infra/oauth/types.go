package oauth

import "golang.org/x/oauth2"

// UserInfo holds the user details we get from the provider.
type UserInfo struct {
	ID      string
	Email   string
	Name    string
	Picture string
}

// OAuthProvider defines the interface for an authentication provider.
// It abstracts the logic for getting OAuth2 configuration and fetching user information.
type OAuthProvider interface {
	GetConfig() *oauth2.Config
	GetUserInfo(token *oauth2.Token) (*UserInfo, error)
}

// ProviderConfig holds the configuration for an OAuth provider.
type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	UserInfoURL  string
	EmailURL     string
	Scopes       []string
}