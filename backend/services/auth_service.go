package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"waf-backend/config"
	"waf-backend/dto"
	"waf-backend/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthService struct {
	oauthConfig *oauth2.Config
	jwtSecret   string
	log         *logrus.Logger
}

func NewAuthService(cfg *config.Config, log *logrus.Logger) *AuthService {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.OAuth.GoogleClientID,
		ClientSecret: cfg.OAuth.GoogleClientSecret,
		RedirectURL:  cfg.OAuth.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	return &AuthService{
		oauthConfig: oauthConfig,
		jwtSecret:   cfg.Security.JWTSecret,
		log:         log,
	}
}

func (s *AuthService) GetAuthURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *AuthService) HandleCallback(ctx context.Context, code, state string) (*dto.LoginResponse, error) {
	s.log.WithFields(logrus.Fields{
		"code":  code[:utils.Min(10, len(code))] + "...",
		"state": state,
	}).Info("Processing OAuth callback")

	// Exchange code for token
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		s.log.WithError(err).Error("Failed to exchange OAuth code")
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from Google
	userInfo, err := s.getUserInfo(ctx, token.AccessToken)
	if err != nil {
		s.log.WithError(err).Error("Failed to get user info")
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Generate JWT token
	jwtToken, expiresAt, err := s.generateJWT(userInfo)
	if err != nil {
		s.log.WithError(err).Error("Failed to generate JWT")
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"user_id": userInfo.ID,
		"email":   userInfo.Email,
	}).Info("User successfully authenticated")

	return &dto.LoginResponse{
		Token:     jwtToken,
		User:      *userInfo,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *AuthService) ValidateJWT(tokenString string) (*dto.AuthMiddlewareData, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, _ := claims["user_id"].(string)
		email, _ := claims["email"].(string)
		name, _ := claims["name"].(string)

		return &dto.AuthMiddlewareData{
			UserID: userID,
			Email:  email,
			Name:   name,
		}, nil
	}

	return nil, errors.New("invalid JWT token")
}

func (s *AuthService) getUserInfo(ctx context.Context, accessToken string) (*dto.User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+accessToken)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		VerifiedEmail bool   `json:"verified_email"`
	}

	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, err
	}

	return &dto.User{
		ID:       googleUser.ID,
		Email:    googleUser.Email,
		Name:     googleUser.Name,
		Picture:  googleUser.Picture,
		Verified: googleUser.VerifiedEmail,
	}, nil
}

func (s *AuthService) generateJWT(user *dto.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour)
	
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"name":    user.Name,
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

