package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"go_example/google_login/config"
	"go_example/google_login/middleware"
	"go_example/google_login/models"
)

type AuthHandler struct {
	config       *config.Config
	db           *models.Database
	store        *sessions.CookieStore
	oauthConfigs map[string]*oauth2.Config
}

func NewAuthHandler(cfg *config.Config, db *models.Database) *AuthHandler {
	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))

	oauthConfigs := make(map[string]*oauth2.Config)
	
	// Google OAuth Config
	oauthConfigs["google"] = &oauth2.Config{
		ClientID:     cfg.GoogleOAuth.ClientID,
		ClientSecret: cfg.GoogleOAuth.ClientSecret,
		RedirectURL:  cfg.GoogleOAuth.RedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
	
	// GitHub OAuth Config
	oauthConfigs["github"] = &oauth2.Config{
		ClientID:     cfg.GitHubOAuth.ClientID,
		ClientSecret: cfg.GitHubOAuth.ClientSecret,
		RedirectURL:  cfg.GitHubOAuth.RedirectURL,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
	
	// Facebook OAuth Config
	oauthConfigs["facebook"] = &oauth2.Config{
		ClientID:     cfg.FacebookOAuth.ClientID,
		ClientSecret: cfg.FacebookOAuth.ClientSecret,
		RedirectURL:  cfg.FacebookOAuth.RedirectURL,
		Scopes:       []string{"email", "public_profile"},
		Endpoint:     facebook.Endpoint,
	}

	return &AuthHandler{
		config:       cfg,
		db:           db,
		store:        store,
		oauthConfigs: oauthConfigs,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	config, ok := h.oauthConfigs[provider]
	if !ok {
		http.Error(w, "Unsupported OAuth provider", http.StatusBadRequest)
		return
	}

	// Generate state and save to session
	state := generateRandomString(32)
	session, _ := h.store.Get(r, "oauth_state")
	session.Values["state"] = state
	session.Save(r, w)

	url := config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	config, ok := h.oauthConfigs[provider]
	if !ok {
		http.Error(w, "Unsupported OAuth provider", http.StatusBadRequest)
		return
	}

	// Verify state
	session, _ := h.store.Get(r, "oauth_state")
	state, ok := session.Values["state"].(string)
	if !ok || state != r.FormValue("state") {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := config.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user info
	userInfo, err := h.getUserInfo(provider, token)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Find or create user in database
	user := &models.User{
		Email:      userInfo.Email,
		Name:       userInfo.Name,
		Picture:    userInfo.Picture,
		Provider:   provider,
		ProviderID: userInfo.ID,
	}

	dbUser, err := h.db.FindOrCreateUser(user)
	if err != nil {
		http.Error(w, "Failed to save user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate JWT
	jwtToken, err := middleware.GenerateJWT(dbUser.ID, dbUser.Email, dbUser.Provider, h.config.JWTSecret)
	if err != nil {
		http.Error(w, "Failed to generate token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set JWT in cookie or return in response
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    jwtToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true, // Set to true in production with HTTPS
		Path:     "/",
	})

	http.Redirect(w, r, "/profile", http.StatusFound)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear JWT cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	// Clear session
	session, _ := h.store.Get(r, "oauth_state")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *AuthHandler) getUserInfo(provider string, token *oauth2.Token) (*UserInfo, error) {
	switch provider {
	case "google":
		return h.getGoogleUserInfo(token)
	case "github":
		return h.getGitHubUserInfo(token)
	case "facebook":
		return h.getFacebookUserInfo(token)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (h *AuthHandler) getGoogleUserInfo(token *oauth2.Token) (*UserInfo, error) {
	client := h.oauthConfigs["google"].Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
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

func (h *AuthHandler) getGitHubUserInfo(token *oauth2.Token) (*UserInfo, error) {
	client := h.oauthConfigs["github"].Client(context.Background(), token)
	
	// Get primary email
	resp, err := client.Get("https://api.github.com/user/emails")
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
	resp, err = client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var profile struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Login string `json:"login"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, err
	}

	name := profile.Name
	if name == "" {
		name = profile.Login
	}

	return &UserInfo{
		ID:    fmt.Sprintf("%d", profile.ID),
		Email: email,
		Name:  name,
	}, nil
}

func (h *AuthHandler) getFacebookUserInfo(token *oauth2.Token) (*UserInfo, error) {
	client := h.oauthConfigs["facebook"].Client(context.Background(), token)
	resp, err := client.Get("https://graph.facebook.com/me?fields=id,name,email,picture&access_token=" + token.AccessToken)
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

type UserInfo struct {
	ID      string
	Email   string
	Name    string
	Picture string
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}