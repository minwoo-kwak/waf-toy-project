package dto

import "time"

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type LoginRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

type LoginResponse struct {
	Token       string    `json:"token"`
	User        User      `json:"user"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	Verified bool   `json:"verified"`
}

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

type AuthMiddlewareData struct {
	UserID string
	Email  string
	Name   string
}