package oauth

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// --- GitHub Provider ---

type githubProvider struct {
	config      *oauth2.Config
	userInfoURL string
	emailURL    string
}

func newGitHubProvider(cfg ProviderConfig) OAuthProvider {
	return &githubProvider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       cfg.Scopes,
			Endpoint:     github.Endpoint,
		},
		userInfoURL: cfg.UserInfoURL,
		emailURL:    cfg.EmailURL,
	}
}

func (p *githubProvider) GetConfig() *oauth2.Config {
	return p.config
}

func (p *githubProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	client := p.config.Client(context.Background(), token)

	// Get primary email
	resp, err := client.Get(p.emailURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return nil, err
	}

	var email string
	for _, e := range emails {
		if e.Primary && e.Verified {
			email = e.Email
			break
		}
	}

	// Get user profile
	resp, err = client.Get(p.userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var profile struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, err
	}

	name := profile.Name
	if name == "" {
		name = profile.Login
	}

	return &UserInfo{
		ID:      fmt.Sprintf("%d", profile.ID),
		Email:   email,
		Name:    name,
		Picture: profile.AvatarURL,
	}, nil
}