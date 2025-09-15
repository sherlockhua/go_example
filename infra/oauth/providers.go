package oauth

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// --- Google Provider ---

type googleProvider struct {
	config      *oauth2.Config
	userInfoURL string
}

func newGoogleProvider(cfg ProviderConfig) OAuthProvider {
	return &googleProvider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       cfg.Scopes,
			Endpoint:     google.Endpoint,
		},
		userInfoURL: cfg.UserInfoURL,
	}
}

func (p *googleProvider) GetConfig() *oauth2.Config {
	return p.config
}

func (p *googleProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	client := p.config.Client(context.Background(), token)
	resp, err := client.Get(p.userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:      userInfo.ID,
		Email:   userInfo.Email,
		Name:    userInfo.Name,
		Picture: userInfo.Picture,
	}, nil
}

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

// --- Facebook Provider ---

type facebookProvider struct {
	config      *oauth2.Config
	userInfoURL string
}

func newFacebookProvider(cfg ProviderConfig) OAuthProvider {
	return &facebookProvider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       cfg.Scopes,
			Endpoint:     facebook.Endpoint,
		},
		userInfoURL: cfg.UserInfoURL,
	}
}

func (p *facebookProvider) GetConfig() *oauth2.Config {
	return p.config
}

func (p *facebookProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	client := p.config.Client(context.Background(), token)
	resp, err := client.Get(p.userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:      userInfo.ID,
		Email:   userInfo.Email,
		Name:    userInfo.Name,
		Picture: userInfo.Picture.Data.URL,
	}, nil
}