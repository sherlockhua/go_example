package handlers

import (
	"encoding/json"
	"net/http"

	"go_example/google_login/middleware"
)

func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	// Get JWT from cookie
	cookie, err := r.Cookie("jwt")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Parse JWT
	claims, err := middleware.ParseJWT(cookie.Value, h.config.JWTSecret)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Get user from database
	user, err := h.db.GetUserByID(claims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Return user info
	json.NewEncoder(w).Encode(user)
}