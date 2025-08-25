package main

import (
	"net/http"
	"waf-backend/config"
	"waf-backend/dto"
	"waf-backend/handlers"
	"waf-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Setup structured logging
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	
	// Set log level based on configuration
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)
	
	log.Info("Starting WAF SaaS Backend Server v2.0")
	
	// Initialize services
	log.Info("Initializing services...")
	authService := services.NewAuthService(cfg, log)
	wafService := services.NewWAFService(log)
	ruleService := services.NewRuleService(log)
	securityTestService := services.NewSecurityTestService(log)
	websocketService := services.NewWebSocketService(log, wafService)
	
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, log)
	wafHandler := handlers.NewWAFHandler(wafService, websocketService, log)
	ruleHandler := handlers.NewRuleHandler(ruleService, log)
	securityTestHandler := handlers.NewSecurityTestHandler(securityTestService, log)
	
	r := gin.Default()
	
	// Add structured logging middleware
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		log.WithFields(logrus.Fields{
			"status":     param.StatusCode,
			"method":     param.Method,
			"path":       param.Path,
			"ip":         param.ClientIP,
			"latency":    param.Latency,
			"user_agent": param.Request.UserAgent(),
		}).Info("API request")
		return ""
	}))
	
	// Add error handling middleware
	r.Use(errorHandler(log))
	
	// CORS middleware with configurable origin
	r.Use(corsMiddleware(cfg.Server.CORSOrigin))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		log.Debug("Health check requested")
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "waf-saas-backend",
			"version": "2.0.0",
			"features": []string{
				"OAuth2 Authentication",
				"WAF Log Monitoring",
				"Custom Rules Management",
				"Security Testing",
				"Real-time WebSocket",
			},
		})
	})

	// Public API routes (no authentication required)
	public := r.Group("/api/v1/public")
	{
		public.GET("/auth/url", authHandler.GetAuthURL)
		public.POST("/auth/callback", authHandler.HandleCallback)
	}

	// Protected API routes (authentication required)
	protected := r.Group("/api/v1")
	protected.Use(authHandler.AuthMiddleware())
	{
		// Authentication routes
		auth := protected.Group("/auth")
		{
			auth.GET("/profile", authHandler.GetUserProfile)
			auth.POST("/logout", authHandler.Logout)
		}
		
		// WAF monitoring routes
		waf := protected.Group("/waf")
		{
			waf.GET("/logs", wafHandler.GetLogs)
			waf.GET("/stats", wafHandler.GetStats)
			waf.GET("/dashboard", wafHandler.GetDashboard)
			waf.POST("/test-logs", wafHandler.GenerateTestLogs) // For testing purposes
		}
		
		// Custom rules management
		rules := protected.Group("/rules")
		{
			rules.POST("/", ruleHandler.CreateRule)
			rules.GET("/", ruleHandler.GetRules)
			rules.GET("/:id", ruleHandler.GetRule)
			rules.PUT("/:id", ruleHandler.UpdateRule)
			rules.DELETE("/:id", ruleHandler.DeleteRule)
		}
		
		// Security testing
		security := protected.Group("/security")
		{
			security.POST("/test", securityTestHandler.RunSecurityTest)
			security.GET("/test-types", securityTestHandler.GetTestTypes)
			security.GET("/quick-test", securityTestHandler.GetQuickTests)
		}
	}

	// WebSocket endpoint with custom authentication
	r.GET("/api/v1/ws", func(c *gin.Context) {
		// WebSocket 전용 토큰 인증 (쿼리 파라미터에서)
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token required",
				"code":  "ERR_NO_TOKEN",
			})
			return
		}

		// JWT 토큰 검증
		userData, err := authService.ValidateJWT(token)
		if err != nil {
			log.WithError(err).Error("WebSocket JWT validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "ERR_INVALID_TOKEN",
				"details": err.Error(),
			})
			return
		}

		// 사용자 정보를 컨텍스트에 설정
		c.Set("user_id", userData.UserID)
		c.Set("email", userData.Email)
		c.Set("name", userData.Name)

		// WebSocket 핸들러 호출
		wafHandler.HandleWebSocket(c)
	})

	// Legacy ping endpoint for backward compatibility
	r.GET("/api/v1/ping", func(c *gin.Context) {
		log.WithFields(logrus.Fields{
			"endpoint": "/api/v1/ping",
			"client_ip": c.ClientIP(),
		}).Debug("Legacy ping endpoint accessed")
		
		c.JSON(http.StatusOK, gin.H{
			"message": "WAF SaaS API is running",
			"version": "2.0.0",
			"features": []string{
				"OAuth2 Authentication",
				"Real-time Monitoring", 
				"Custom Rules",
				"Security Testing",
			},
		})
	})

	log.WithFields(logrus.Fields{
		"port": cfg.Server.Port,
		"cors_origin": cfg.Server.CORSOrigin,
		"features": []string{
			"Google OAuth2",
			"ModSecurity Log Analysis",
			"Custom Rule Management",
			"Security Testing Suite",
			"Real-time WebSocket Streaming",
		},
	}).Info("WAF SaaS Backend Server started successfully")
	
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}
}


// corsMiddleware handles CORS with configurable origin
func corsMiddleware(corsOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		
		// Allow all origins if corsOrigin is "*"
		if corsOrigin == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin == corsOrigin {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			// For development, also allow localhost variations
			allowedOrigins := []string{corsOrigin, "http://localhost", "http://localhost:3000", "http://127.0.0.1"}
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					c.Header("Access-Control-Allow-Origin", origin)
					break
				}
			}
			if c.GetHeader("Access-Control-Allow-Origin") == "" {
				c.Header("Access-Control-Allow-Origin", corsOrigin)
			}
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

// errorHandler middleware for centralized error handling
func errorHandler(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during request processing
		for _, err := range c.Errors {
			log.WithFields(logrus.Fields{
				"error":  err.Error(),
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"ip":     c.ClientIP(),
			}).Error("Request error occurred")

			// Return error response if not already sent
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
					"Internal server error",
					dto.ErrInternal,
				))
			}
		}
	}
}