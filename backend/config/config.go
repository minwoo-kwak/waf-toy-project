package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"waf-backend/utils"
)

type Config struct {
	Server   ServerConfig
	OAuth    OAuthConfig
	Security SecurityConfig
	Logging  LoggingConfig
}

type ServerConfig struct {
	Port       string
	CORSOrigin string
}

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	RedirectURL        string
}

type SecurityConfig struct {
	JWTSecret string
}

type LoggingConfig struct {
	Level string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:       utils.GetEnv("PORT", "8080"),
			CORSOrigin: utils.GetEnv("CORS_ORIGIN", "*"),
		},
		OAuth: OAuthConfig{
			GoogleClientID:     utils.GetEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: utils.GetEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:        utils.GetEnv("GOOGLE_REDIRECT_URL", "http://localhost:3000/auth/callback"),
		},
		Security: SecurityConfig{
			JWTSecret: getJWTSecret(),
		},
		Logging: LoggingConfig{
			Level: utils.GetEnv("LOG_LEVEL", "info"),
		},
	}
}

// getJWTSecret generates a secure JWT secret if not provided via environment
func getJWTSecret() string {
	secret := utils.GetEnv("JWT_SECRET", "")
	if secret == "" {
		// Generate a random 32-byte secret
		bytes := make([]byte, 32)
		if _, err := rand.Read(bytes); err != nil {
			log.Printf("Warning: Failed to generate random JWT secret, using fallback: %v", err)
			return "waf-saas-default-jwt-secret-please-change-in-production"
		}
		secret = hex.EncodeToString(bytes)
		log.Println("Warning: JWT_SECRET not set, generated random secret. Set JWT_SECRET environment variable for production.")
	}
	return secret
}