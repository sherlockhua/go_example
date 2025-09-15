package main

import (
	"log"
	"net/http"

	"go_example/google_login/config"
	"go_example/google_login/handlers"
	"go_example/google_login/middleware"
	"go_example/google_login/models"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := models.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize database schema
	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create handlers
	authHandler := handlers.NewAuthHandler(cfg, db)
	userHandler := handlers.NewUserHandler(cfg, db)

	// Setup routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	http.HandleFunc("/login", authHandler.Login)
	http.HandleFunc("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		r.URL.RawQuery = r.URL.RawQuery + "&provider=google"
		authHandler.Callback(w, r)
	})
	http.HandleFunc("/auth/github/callback", func(w http.ResponseWriter, r *http.Request) {
		r.URL.RawQuery = r.URL.RawQuery + "&provider=github"
		authHandler.Callback(w, r)
	})
	http.HandleFunc("/auth/facebook/callback", func(w http.ResponseWriter, r *http.Request) {
		r.URL.RawQuery = r.URL.RawQuery + "&provider=facebook"
		authHandler.Callback(w, r)
	})
	http.HandleFunc("/logout", authHandler.Logout)
	http.HandleFunc("/profile", middleware.JWTMiddleware(cfg.JWTSecret, userHandler.Profile))

	// Start server
	log.Printf("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}