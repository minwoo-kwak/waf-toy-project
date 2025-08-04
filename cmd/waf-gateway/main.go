package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 로그 설정
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	// Gin 라우터 초기화
	router := gin.Default()

	// 미들웨어 설정
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 헬스체크 엔드포인트
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"service":   "waf-gateway",
		})
	})

	// WAF 상태 확인 엔드포인트
	router.GET("/waf/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"waf_enabled":     true,
			"rules_loaded":    true,
			"modsecurity":     "active",
			"owasp_crs":      "enabled",
			"last_updated":   time.Now().Unix(),
		})
	})

	// 테스트용 엔드포인트 (WAF 룰 테스트용)
	router.GET("/test", func(c *gin.Context) {
		// 쿼리 파라미터 로깅 (공격 탐지용)
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				logrus.WithFields(logrus.Fields{
					"parameter": key,
					"value":     value,
					"client_ip": c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
				}).Info("Request parameter logged")
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Test endpoint - 이 요청이 WAF를 통과했습니다",
			"client_ip": c.ClientIP(),
			"timestamp": time.Now().Unix(),
		})
	})

	// 메트릭 엔드포인트 (나중에 Prometheus 연동용)
	router.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, "# Prometheus metrics will be implemented here")
	})

	// HTTP 서버 설정
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// 서버 시작
	go func() {
		logrus.Info("WAF Gateway 서버가 포트 8080에서 시작됩니다...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("서버 시작 실패: %v", err)
		}
	}()

	// Graceful shutdown 설정
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("서버 종료 중...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatal("서버 강제 종료:", err)
	}

	logrus.Info("서버가 정상적으로 종료되었습니다")
}