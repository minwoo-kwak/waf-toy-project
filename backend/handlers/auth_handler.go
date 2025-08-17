package handlers

import (
	"net/http"
	"strings"
	"waf-backend/dto"
	"waf-backend/services"
	"waf-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	authService *services.AuthService
	log         *logrus.Logger
}

func NewAuthHandler(authService *services.AuthService, log *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		log:         log,
	}
}

func (h *AuthHandler) GetAuthURL(c *gin.Context) {
	state := c.Query("state")
	if state == "" {
		state = "random-state-string"
	}
	
	authURL := h.authService.GetAuthURL(state)
	
	h.log.WithFields(logrus.Fields{
		"state":    state,
		"auth_url": authURL,
	}).Info("Generated OAuth URL")
	
	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

func (h *AuthHandler) HandleCallback(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Error("Invalid login request")
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			"Invalid request format",
			dto.ErrInvalidRequest,
			err.Error(),
		))
		return
	}
	
	h.log.WithFields(logrus.Fields{
		"code":  req.Code[:utils.Min(10, len(req.Code))] + "...",
		"state": req.State,
	}).Info("Processing OAuth callback")
	
	response, err := h.authService.HandleCallback(c.Request.Context(), req.Code, req.State)
	if err != nil {
		h.log.WithError(err).Error("OAuth callback failed")
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
			"Authentication failed",
			dto.ErrAuthFailed,
			err.Error(),
		))
		return
	}
	
	h.log.WithFields(logrus.Fields{
		"user_id": response.User.ID,
		"email":   response.User.Email,
	}).Info("User successfully authenticated via OAuth")
	
	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) GetUserProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")
	name, _ := c.Get("name")
	
	h.log.WithFields(logrus.Fields{
		"user_id": userID,
		"email":   email,
	}).Debug("User profile requested")
	
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    userID,
			"email": email,
			"name":  name,
		},
		"message": "Profile retrieved successfully",
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	h.log.WithField("user_id", userID).Info("User logged out")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// AuthMiddleware JWT 토큰 검증 미들웨어
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
				"Authorization header required",
				dto.ErrNoAuthHeader,
			))
			c.Abort()
			return
		}
		
		// Bearer 토큰 형식 확인
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
				"Invalid authorization header format",
				dto.ErrInvalidAuthFormat,
			))
			c.Abort()
			return
		}
		
		token := tokenParts[1]
		
		// JWT 토큰 검증
		userData, err := h.authService.ValidateJWT(token)
		if err != nil {
			h.log.WithError(err).Debug("JWT validation failed")
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
				"Invalid or expired token",
				dto.ErrInvalidToken,
				err.Error(),
			))
			c.Abort()
			return
		}
		
		// 사용자 정보를 컨텍스트에 저장
		c.Set("user_id", userData.UserID)
		c.Set("email", userData.Email)
		c.Set("name", userData.Name)
		
		c.Next()
	}
}

// OptionalAuthMiddleware 선택적 인증 미들웨어 (토큰이 있으면 검증, 없어도 통과)
func (h *AuthHandler) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}
		
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}
		
		token := tokenParts[1]
		userData, err := h.authService.ValidateJWT(token)
		if err != nil {
			c.Next()
			return
		}
		
		c.Set("user_id", userData.UserID)
		c.Set("email", userData.Email)
		c.Set("name", userData.Name)
		
		c.Next()
	}
}

