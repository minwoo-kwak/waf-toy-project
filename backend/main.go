package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Setup structured logging
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)
	
	// Get configuration from environment variables
	port := getEnv("PORT", "8080")
	corsOrigin := getEnv("CORS_ORIGIN", "*")
	
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
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", corsOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		log.Info("Health check requested")
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "waf-backend",
			"version": "1.0.0",
		})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			log.WithFields(logrus.Fields{
				"endpoint": "/api/v1/ping",
				"client_ip": c.ClientIP(),
			}).Info("Ping endpoint accessed")
			
			c.JSON(http.StatusOK, gin.H{
				"message": "WAF API is running",
				"timestamp": gin.H{"unix": gin.H{"seconds": 0, "nanoseconds": 0}},
			})
		})
	}

	log.WithField("port", port).Info("Starting WAF Backend Server")
	if err := r.Run(":" + port); err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}
}

// getEnv returns environment variable value or default if not set
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
					"code":  "ERR_INTERNAL",
				})
			}
		}
	}
}