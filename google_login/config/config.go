package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type OAuthProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type Config struct {
	DatabaseURL       string
	SessionSecret     string
	JWTSecret         string
	GoogleOAuth       OAuthProvider
	GitHubOAuth       OAuthProvider
	FacebookOAuth     OAuthProvider
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found")
	}

	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://user:password@localhost/dbname?sslmode=disable"),
		SessionSecret: getEnv("SESSION_SECRET", "super-secret-key"),
		JWTSecret:     getEnv("JWT_SECRET", "jwt-super-secret-key"),
		GoogleOAuth: OAuthProvider{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		},
		GitHubOAuth: OAuthProvider{
			ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/auth/github/callback"),
		},
		FacebookOAuth: OAuthProvider{
			ClientID:     getEnv("FACEBOOK_CLIENT_ID", ""),
			ClientSecret: getEnv("FACEBOOK_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("FACEBOOK_REDIRECT_URL", "http://localhost:8080/auth/facebook/callback"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}