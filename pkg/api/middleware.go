package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"waf-k8s-project/pkg/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// WAFMiddleware WAF 관련 미들웨어들
type WAFMiddleware struct {
	rateLimiter *ratelimit.RedisRateLimiter
	logger      *logrus.Logger
}

// NewWAFMiddleware WAF 미들웨어 생성
func NewWAFMiddleware(rateLimiter *ratelimit.RedisRateLimiter) *WAFMiddleware {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	
	return &WAFMiddleware{
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// RateLimitMiddleware Rate Limiting 미들웨어
func (w *WAFMiddleware) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := w.getClientIP(c)
		
		// IP 차단 상태 확인
		blocked, reason, err := w.rateLimiter.IsBlocked(c.Request.Context(), clientIP)
		if err != nil {
			w.logger.WithError(err).Error("IP 차단 상태 확인 실패")
		}
		
		if blocked {
			w.logger.WithFields(logrus.Fields{
				"client_ip": clientIP,
				"reason":    reason,
				"path":      c.Request.URL.Path,
			}).Warn("차단된 IP에서 요청 시도")
			
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "IP가 차단되었습니다",
				"reason":  reason,
				"message": "관리자에게 문의하세요",
			})
			c.Abort()
			return
		}
		
		// 일반 Rate Limiting (분당 100회)
		normalConfig := ratelimit.LimitConfig{
			MaxRequests: 100,
			Window:      time.Minute,
			BurstSize:   10,
		}
		
		result, err := w.rateLimiter.CheckLimit(c.Request.Context(), clientIP, normalConfig)
		if err != nil {
			w.logger.WithError(err).Error("Rate limit 체크 실패")
			// 에러 발생 시 요청 허용 (Fail Open)
			c.Next()
			return
		}
		
		// Rate Limit 헤더 추가
		c.Header("X-RateLimit-Limit", strconv.Itoa(normalConfig.MaxRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))
		
		if !result.Allowed {
			// Burst limit도 체크
			burstResult, burstErr := w.rateLimiter.CheckBurstLimit(c.Request.Context(), clientIP, normalConfig)
			if burstErr == nil && !burstResult.Allowed {
				// 연속 초과 시 IP 차단 (1시간)
				blockErr := w.rateLimiter.BlockIP(c.Request.Context(), clientIP, time.Hour, "Rate limit 연속 초과")
				if blockErr != nil {
					w.logger.WithError(blockErr).Error("IP 차단 실패")
				}
			}
			
			c.Header("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
			
			w.logger.WithFields(logrus.Fields{
				"client_ip":    clientIP,
				"path":         c.Request.URL.Path,
				"remaining":    result.Remaining,
				"retry_after":  result.RetryAfter,
			}).Warn("Rate limit 초과")
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "요청 횟수가 제한을 초과했습니다",
				"retry_after": int(result.RetryAfter.Seconds()),
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// SecurityHeadersMiddleware 보안 헤더 추가
func (w *WAFMiddleware) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// XSS 보호
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Content Type Sniffing 방지
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Clickjacking 방지
		c.Header("X-Frame-Options", "DENY")
		
		// HSTS (HTTPS 강제)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		// Content Security Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"connect-src 'self'; " +
			"font-src 'self'; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"
		c.Header("Content-Security-Policy", csp)
		
		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// 파워드 바이 헤더 제거 (정보 누출 방지)
		c.Header("Server", "WAF-Gateway/1.0")
		
		c.Next()
	}
}

// LoggingMiddleware 고급 로깅 미들웨어
func (w *WAFMiddleware) LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logEntry := logrus.WithFields(logrus.Fields{
			"timestamp":    param.TimeStamp.Format(time.RFC3339),
			"status":       param.StatusCode,
			"method":       param.Method,
			"path":         param.Path,
			"client_ip":    param.ClientIP,
			"user_agent":   param.Request.UserAgent(),
			"latency_ms":   param.Latency.Milliseconds(),
			"body_size":    param.BodySize,
			"referer":      param.Request.Referer(),
		})
		
		// 의심스러운 요청 탐지
		if w.isSuspiciousRequest(param) {
			logEntry = logEntry.WithField("suspicious", true)
			logEntry.Warn("의심스러운 요청 탐지")
		}
		
		// 상태 코드별 로그 레벨 설정
		switch {
		case param.StatusCode >= 500:
			logEntry.Error("서버 에러")
		case param.StatusCode >= 400:
			logEntry.Warn("클라이언트 에러")
		default:
			logEntry.Info("정상 요청")
		}
		
		return ""
	})
}

// ThreatDetectionMiddleware 위협 탐지 미들웨어
func (w *WAFMiddleware) ThreatDetectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := w.getClientIP(c)
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		userAgent := c.Request.UserAgent()
		
		// 의심스러운 패턴 탐지
		threats := w.detectThreats(path, query, userAgent)
		
		if len(threats) > 0 {
			// 위협 점수 계산
			threatScore := w.calculateThreatScore(threats)
			
			logFields := logrus.Fields{
				"client_ip":    clientIP,
				"path":         path,
				"query":        query,
				"user_agent":   userAgent,
				"threats":      threats,
				"threat_score": threatScore,
			}
			
			if threatScore >= 8 { // 높은 위험도
				// IP 즉시 차단
				blockErr := w.rateLimiter.BlockIP(c.Request.Context(), clientIP, time.Hour*24, "고위험 위협 탐지")
				if blockErr != nil {
					w.logger.WithError(blockErr).Error("IP 차단 실패")
				}
				
				w.logger.WithFields(logFields).Error("고위험 위협 탐지 - IP 차단")
				
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Security threat detected",
					"message": "보안 위협이 탐지되어 접근이 차단되었습니다",
				})
				c.Abort()
				return
				
			} else if threatScore >= 5 { // 중간 위험도
				w.logger.WithFields(logFields).Warn("중간 위험도 위협 탐지")
				
				// Rate limit 강화 (임시)
				strictConfig := ratelimit.LimitConfig{
					MaxRequests: 10, // 10회로 제한
					Window:      time.Minute,
					BurstSize:   2,
				}
				
				result, err := w.rateLimiter.CheckLimit(c.Request.Context(), clientIP+":strict", strictConfig)
				if err == nil && !result.Allowed {
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error":   "Enhanced rate limit",
						"message": "보안상의 이유로 요청이 제한됩니다",
					})
					c.Abort()
					return
				}
			}
		}
		
		c.Next()
	}
}

// getClientIP 클라이언트 IP 추출 (프록시 고려)
func (w *WAFMiddleware) getClientIP(c *gin.Context) string {
	// X-Forwarded-For 헤더 확인 (로드밸런서/프록시 환경)
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// X-Real-IP 헤더 확인
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	
	// 기본 클라이언트 IP
	return c.ClientIP()
}

// isSuspiciousRequest 의심스러운 요청 판단
func (w *WAFMiddleware) isSuspiciousRequest(param gin.LogFormatterParams) bool {
	// SQL Injection 패턴
	sqlPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_", 
		"union", "select", "insert", "delete", "update", "drop",
	}
	
	path := strings.ToLower(param.Path)
	query := strings.ToLower(param.Request.URL.RawQuery)
	
	for _, pattern := range sqlPatterns {
		if strings.Contains(path, pattern) || strings.Contains(query, pattern) {
			return true
		}
	}
	
	// XSS 패턴
	xssPatterns := []string{
		"<script", "</script>", "javascript:", "vbscript:", "onload=", "onerror=",
	}
	
	for _, pattern := range xssPatterns {
		if strings.Contains(path, pattern) || strings.Contains(query, pattern) {
			return true
		}
	}
	
	// Path Traversal 패턴
	if strings.Contains(path, "..") || strings.Contains(query, "..") {
		return true
	}
	
	return false
}

// detectThreats 위협 탐지
func (w *WAFMiddleware) detectThreats(path, query, userAgent string) []string {
	var threats []string
	
	path = strings.ToLower(path)
	query = strings.ToLower(query)
	userAgent = strings.ToLower(userAgent)
	
	// SQL Injection 탐지
	sqlPatterns := []string{"union select", "' or '1'='1", "drop table", "exec xp_"}
	for _, pattern := range sqlPatterns {
		if strings.Contains(path, pattern) || strings.Contains(query, pattern) {
			threats = append(threats, "SQL_INJECTION")
			break
		}
	}
	
	// XSS 탐지
	xssPatterns := []string{"<script>alert", "javascript:alert", "<img src=x onerror="}
	for _, pattern := range xssPatterns {
		if strings.Contains(path, pattern) || strings.Contains(query, pattern) {
			threats = append(threats, "XSS")
			break
		}
	}
	
	// Command Injection 탐지
	cmdPatterns := []string{"; cat ", "| nc ", "&& whoami", "$(curl"}
	for _, pattern := range cmdPatterns {
		if strings.Contains(query, pattern) {
			threats = append(threats, "COMMAND_INJECTION")
			break
		}
	}
	
	// 악성 User-Agent 탐지
	maliciousUA := []string{"sqlmap", "nikto", "burpsuite", "nessus", "acunetix"}
	for _, ua := range maliciousUA {
		if strings.Contains(userAgent, ua) {
			threats = append(threats, "MALICIOUS_USER_AGENT")
			break
		}
	}
	
	// Path Traversal 탐지
	if strings.Contains(path, "../") || strings.Contains(query, "../") {
		threats = append(threats, "PATH_TRAVERSAL")
	}
	
	return threats
}

// calculateThreatScore 위협 점수 계산
func (w *WAFMiddleware) calculateThreatScore(threats []string) int {
	scoreMap := map[string]int{
		"SQL_INJECTION":        9,
		"COMMAND_INJECTION":    10,
		"XSS":                  7,
		"PATH_TRAVERSAL":       6,
		"MALICIOUS_USER_AGENT": 8,
	}
	
	maxScore := 0
	for _, threat := range threats {
		if score, exists := scoreMap[threat]; exists && score > maxScore {
			maxScore = score
		}
	}
	
	return maxScore
}