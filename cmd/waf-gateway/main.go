package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"waf-k8s-project/pkg/api"
	"waf-k8s-project/pkg/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus 메트릭 정의
var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "waf_requests_total",
			Help: "Total number of requests processed by WAF",
		},
		[]string{"method", "status", "path"},
	)
	
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "waf_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	
	blockedRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "waf_blocked_requests_total",
			Help: "Total number of blocked requests",
		},
		[]string{"reason", "client_ip"},
	)
	
	activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "waf_active_connections",
			Help: "Number of active connections",
		},
	)
	
	rateLimitHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "waf_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"client_ip"},
	)
	
	threatDetections = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "waf_threat_detections_total",
			Help: "Total number of threat detections",
		},
		[]string{"threat_type", "severity"},
	)
)

func init() {
	// Prometheus 메트릭 등록
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(blockedRequests)
	prometheus.MustRegister(activeConnections)
	prometheus.MustRegister(rateLimitHits)
	prometheus.MustRegister(threatDetections)
}

func main() {
	// 로그 설정
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	// 환경변수에서 설정 읽기
	redisURL := getEnv("REDIS_URL", "redis://redis-service.waf-system.svc.cluster.local:6379")
	port := getEnv("PORT", "8080")

	// Redis Rate Limiter 초기화
	rateLimiter, err := ratelimit.NewRedisRateLimiter(redisURL, "waf")
	if err != nil {
		logrus.WithError(err).Fatal("Redis Rate Limiter 초기화 실패")
	}
	defer rateLimiter.Close()

	// WAF 미들웨어 초기화
	wafMiddleware := api.NewWAFMiddleware(rateLimiter)

	// Gin 라우터 초기화 (운영 모드)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// 미들웨어 적용
	router.Use(prometheusMiddleware())          // Prometheus 메트릭
	router.Use(wafMiddleware.LoggingMiddleware()) // 고급 로깅
	router.Use(gin.Recovery())                    // 패닉 복구
	router.Use(wafMiddleware.SecurityHeadersMiddleware()) // 보안 헤더
	router.Use(wafMiddleware.ThreatDetectionMiddleware()) // 위협 탐지
	router.Use(wafMiddleware.RateLimitMiddleware())       // Rate Limiting

	// 기본 라우트들
	setupRoutes(router, rateLimiter)

	// HTTP 서버 설정
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 서버 시작
	go func() {
		logrus.WithField("port", port).Info("🛡️ WAF Gateway 서버 시작")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("서버 시작 실패")
		}
	}()

	// Graceful shutdown 설정
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("🛑 서버 종료 중...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logrus.WithError(err).Fatal("서버 강제 종료")
	}

	logrus.Info("✅ 서버가 정상적으로 종료되었습니다")
}

// setupRoutes 라우트 설정
func setupRoutes(router *gin.Engine, rateLimiter *ratelimit.RedisRateLimiter) {
	// 헬스체크 엔드포인트
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"service":   "waf-gateway",
			"version":   "1.0.0",
		})
	})

	// 상세 헬스체크 (의존성 포함)
	router.GET("/health/detailed", func(ctx *gin.Context) {
		health := gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"service":   "waf-gateway",
			"version":   "1.0.0",
			"checks":    gin.H{},
		}

		// Redis 연결 상태 확인
		pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		
		if stats, err := rateLimiter.GetStats(pingCtx, "health-check"); err != nil {
			health["checks"].(gin.H)["redis"] = gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			health["status"] = "degraded"
		} else {
			health["checks"].(gin.H)["redis"] = gin.H{
				"status": "healthy",
				"stats":  stats,
			}
		}

		ctx.JSON(http.StatusOK, health)
	})

	// WAF 상태 확인 엔드포인트
	router.GET("/waf/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"waf_enabled":        true,
			"rules_loaded":       true,
			"modsecurity":        "active",
			"owasp_crs":         "v4.0.0",
			"rate_limiting":     "enabled",
			"threat_detection":  "active",
			"last_updated":      time.Now().Unix(),
			"features": gin.H{
				"sql_injection_protection": true,
				"xss_protection":           true,
				"path_traversal_protection": true,
				"command_injection_protection": true,
				"malicious_ua_detection":   true,
				"rate_limiting":            true,
				"ip_blocking":              true,
				"burst_protection":         true,
			},
		})
	})

	// WAF 통계 엔드포인트
	router.GET("/waf/stats", func(c *gin.Context) {
		clientIP := c.ClientIP()
		stats, err := rateLimiter.GetStats(c.Request.Context(), clientIP)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "통계 조회 실패",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"timestamp": time.Now().Unix(),
			"client_ip": clientIP,
			"rate_limit_stats": stats,
		})
	})

	// 테스트용 엔드포인트 (WAF 룰 테스트용)
	router.GET("/test", func(c *gin.Context) {
		// 쿼리 파라미터 로깅 (공격 탐지용)
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				logrus.WithFields(logrus.Fields{
					"parameter":  key,
					"value":      value,
					"client_ip":  c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
				}).Info("테스트 요청 파라미터")
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "🧪 테스트 엔드포인트 - WAF를 통과한 정상 요청",
			"client_ip": c.ClientIP(),
			"timestamp": time.Now().Unix(),
			"headers": gin.H{
				"user_agent": c.Request.UserAgent(),
				"referer":    c.Request.Referer(),
			},
		})
	})

	// API 그룹
	api := router.Group("/api/v1")
	{
		// 사용자 관리 시뮬레이션
		api.GET("/users", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"users": []gin.H{
					{"id": 1, "name": "admin", "role": "administrator"},
					{"id": 2, "name": "user", "role": "user"},
				},
			})
		})

		// 로그인 시뮬레이션 (SQL Injection 테스트 대상)
		api.POST("/login", func(c *gin.Context) {
			var loginData struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}

			if err := c.ShouldBindJSON(&loginData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "잘못된 요청 형식",
				})
				return
			}

			// 입력값 로깅 (보안 분석용)
			logrus.WithFields(logrus.Fields{
				"username":   loginData.Username,
				"client_ip":  c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			}).Info("로그인 시도")

			c.JSON(http.StatusOK, gin.H{
				"message": "로그인 시뮬레이션",
				"status":  "success",
			})
		})

		// 검색 기능 시뮬레이션 (XSS 테스트 대상)
		api.GET("/search", func(c *gin.Context) {
			query := c.Query("q")
			
			logrus.WithFields(logrus.Fields{
				"search_query": query,
				"client_ip":    c.ClientIP(),
			}).Info("검색 요청")

			c.JSON(http.StatusOK, gin.H{
				"query":   query,
				"results": []string{"결과1", "결과2", "결과3"},
			})
		})

		// 파일 업로드 시뮬레이션 (Path Traversal 테스트 대상)
		api.POST("/upload", func(c *gin.Context) {
			filename := c.PostForm("filename")
			
			logrus.WithFields(logrus.Fields{
				"filename":  filename,
				"client_ip": c.ClientIP(),
			}).Info("파일 업로드 시도")

			c.JSON(http.StatusOK, gin.H{
				"message":  "파일 업로드 시뮬레이션",
				"filename": filename,
				"status":   "uploaded",
			})
		})
	}

	// 관리자 API (더 엄격한 보안)
	admin := router.Group("/admin")
	{
		// IP 차단 관리
		admin.POST("/block-ip", func(c *gin.Context) {
			var blockData struct {
				IP       string `json:"ip" binding:"required"`
				Duration int    `json:"duration"` // 시간(초)
				Reason   string `json:"reason"`
			}

			if err := c.ShouldBindJSON(&blockData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "잘못된 요청 형식",
				})
				return
			}

			duration := time.Duration(blockData.Duration) * time.Second
			if duration == 0 {
				duration = time.Hour // 기본 1시간
			}

			err := rateLimiter.BlockIP(c.Request.Context(), blockData.IP, duration, blockData.Reason)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "IP 차단 실패",
				})
				return
			}

			logrus.WithFields(logrus.Fields{
				"blocked_ip": blockData.IP,
				"duration":   duration,
				"reason":     blockData.Reason,
				"admin_ip":   c.ClientIP(),
			}).Warn("관리자에 의한 IP 차단")

			c.JSON(http.StatusOK, gin.H{
				"message":  "IP가 성공적으로 차단되었습니다",
				"ip":       blockData.IP,
				"duration": duration.String(),
			})
		})

		// 차단된 IP 목록 조회
		admin.GET("/blocked-ips", func(c *gin.Context) {
			// 실제로는 Redis에서 blocked: 패턴으로 모든 키를 조회해야 함
			c.JSON(http.StatusOK, gin.H{
				"message": "차단된 IP 목록 조회 기능 (구현 예정)",
			})
		})
	}

	// Prometheus 메트릭 엔드포인트
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

// prometheusMiddleware Prometheus 메트릭 수집 미들웨어
func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 활성 연결 수 증가
		activeConnections.Inc()
		defer activeConnections.Dec()

		c.Next()

		// 요청 처리 시간 기록
		duration := time.Since(start)
		requestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration.Seconds())

		// 요청 카운터 증가
		requestsTotal.WithLabelValues(c.Request.Method, fmt.Sprintf("%d", c.Writer.Status()), c.FullPath()).Inc()

		// 차단된 요청 기록
		if c.Writer.Status() == http.StatusForbidden || c.Writer.Status() == http.StatusTooManyRequests {
			reason := "unknown"
			if c.Writer.Status() == http.StatusTooManyRequests {
				reason = "rate_limit"
				rateLimitHits.WithLabelValues(c.ClientIP()).Inc()
			} else {
				reason = "security_block"
			}
			blockedRequests.WithLabelValues(reason, c.ClientIP()).Inc()
		}
	}
}

// getEnv 환경변수 읽기 (기본값 지원)
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}