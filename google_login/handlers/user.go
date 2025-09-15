package handlers

import (
	"encoding/json"
	"net/http"

	"go_example/google_login/config"
	"go_example/google_login/middleware"
	"go_example/google_login/models"
)

// UserHandler handles user-related requests.
type UserHandler struct {
	config *config.Config
	db     *models.Database
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(cfg *config.Config, db *models.Database) *UserHandler {
	return &UserHandler{
		config: cfg,
		db:     db,
	}
}

// Profile returns the user's profile information.
func (h *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	// Get claims from context
	claims, ok := r.Context().Value("userClaims").(*middleware.JWTClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	// Get user from database
	user, err := h.db.GetUserByID(claims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Return user as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}