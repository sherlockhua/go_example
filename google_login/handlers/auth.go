package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/sessions"

	"go_example/google_login/config"
	"go_example/google_login/middleware"
	"go_example/google_login/models"
	"go_example/infra/oauth"
)

// AuthHandler handles authentication requests.
type AuthHandler struct {
	config *config.Config
	db     *models.Database
	store  *sessions.CookieStore
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(cfg *config.Config, db *models.Database) *AuthHandler {
	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	return &AuthHandler{
		config: cfg,
		db:     db,
		store:  store,
	}
}

func (h *AuthHandler) getProviderConfig(providerName string) (oauth.ProviderConfig, error) {
	switch providerName {
	case "google":
		return oauth.ProviderConfig{
			ClientID:     h.config.GoogleOAuth.ClientID,
			ClientSecret: h.config.GoogleOAuth.ClientSecret,
			RedirectURL:  h.config.GoogleOAuth.RedirectURL,
			UserInfoURL:  h.config.GoogleOAuth.UserInfoURL,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		}, nil
	case "github":
		return oauth.ProviderConfig{
			ClientID:     h.config.GitHubOAuth.ClientID,
			ClientSecret: h.config.GitHubOAuth.ClientSecret,
			RedirectURL:  h.config.GitHubOAuth.RedirectURL,
			UserInfoURL:  h.config.GitHubOAuth.UserInfoURL,
			EmailURL:     h.config.GitHubOAuth.EmailURL,
			Scopes:       []string{"user:email"},
		}, nil
	case "facebook":
		return oauth.ProviderConfig{
			ClientID:     h.config.FacebookOAuth.ClientID,
			ClientSecret: h.config.FacebookOAuth.ClientSecret,
			RedirectURL:  h.config.FacebookOAuth.RedirectURL,
			UserInfoURL:  h.config.FacebookOAuth.UserInfoURL,
			Scopes:       []string{"email", "public_profile"},
		}, nil
	default:
		return oauth.ProviderConfig{}, fmt.Errorf("provider %s not supported", providerName)
	}
}

// Login handles the login request for a given provider.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	providerName := r.URL.Query().Get("provider")
	providerConfig, err := h.getProviderConfig(providerName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	provider, err := oauth.GetProvider(providerName, providerConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate state and save to session
	state := generateRandomString(32)
	session, _ := h.store.Get(r, "oauth_state")
	session.Values["state"] = state
	session.Save(r, w)

	url := provider.GetConfig().AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Callback handles the callback from the OAuth provider.
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	providerName := r.URL.Query().Get("provider")
	providerConfig, err := h.getProviderConfig(providerName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	provider, err := oauth.GetProvider(providerName, providerConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	token, err := provider.GetConfig().Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user info
	userInfo, err := provider.GetUserInfo(token)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Find or create user in database
	user := &models.User{
		Email:      userInfo.Email,
		Name:       userInfo.Name,
		Picture:    userInfo.Picture,
		Provider:   providerName,
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

	// Set JWT in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    jwtToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true, // Set to true in production with HTTPS
		Path:     "/",
	})

	http.Redirect(w, r, "/", http.StatusFound) // Redirect to home/profile page
}

// Logout handles the logout request.
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

	// Clear session state
	session, _ := h.store.Get(r, "oauth_state")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}