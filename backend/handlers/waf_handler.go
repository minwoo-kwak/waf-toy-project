package handlers

import (
	"net/http"
	"strconv"
	"waf-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type WAFHandler struct {
	wafService       *services.WAFService
	websocketService *services.WebSocketService
	log              *logrus.Logger
}

func NewWAFHandler(wafService *services.WAFService, websocketService *services.WebSocketService, log *logrus.Logger) *WAFHandler {
	return &WAFHandler{
		wafService:       wafService,
		websocketService: websocketService,
		log:              log,
	}
}

func (h *WAFHandler) GetLogs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	
	if limit > 500 {
		limit = 500 // 최대 500개로 제한
	}
	
	userID, _ := c.Get("user_id")
	h.log.WithFields(logrus.Fields{
		"user_id": userID,
		"limit":   limit,
	}).Debug("WAF logs requested")
	
	logs := h.wafService.GetLogs(limit)
	
	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"count": len(logs),
		"limit": limit,
	})
}

func (h *WAFHandler) GetStats(c *gin.Context) {
	userID, _ := c.Get("user_id")
	h.log.WithField("user_id", userID).Debug("WAF stats requested")
	
	stats := h.wafService.GetStats()
	
	// WebSocket 연결 수 추가
	stats.Timestamp = stats.Timestamp
	
	c.JSON(http.StatusOK, gin.H{
		"stats":             stats,
		"websocket_clients": h.websocketService.GetConnectedClients(),
	})
}

func (h *WAFHandler) GetDashboard(c *gin.Context) {
	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")
	
	h.log.WithFields(logrus.Fields{
		"user_id": userID,
		"email":   email,
	}).Debug("Dashboard data requested")
	
	// 통계 정보 가져오기
	stats := h.wafService.GetStats()
	
	// 최근 로그 10개
	recentLogs := h.wafService.GetLogs(10)
	
	// 대시보드 응답 구성
	dashboard := gin.H{
		"user": gin.H{
			"id":    userID,
			"email": email,
		},
		"stats": stats,
		"recent_logs": recentLogs,
		"system_info": gin.H{
			"waf_engine":        "ModSecurity 3.x",
			"rule_set":          "OWASP CRS 4.x",
			"websocket_clients": h.websocketService.GetConnectedClients(),
			"uptime":            "Running", // 실제로는 시스템 업타임 계산
		},
	}
	
	c.JSON(http.StatusOK, dashboard)
}

func (h *WAFHandler) HandleWebSocket(c *gin.Context) {
	h.log.WithFields(logrus.Fields{
		"user_id": c.GetString("user_id"),
		"email":   c.GetString("email"),
	}).Info("WebSocket connection requested")
	
	h.websocketService.HandleWebSocket(c)
}

func (h *WAFHandler) GenerateTestLogs(c *gin.Context) {
	userID, _ := c.Get("user_id")
	h.log.WithField("user_id", userID).Info("Test log generation requested")
	
	// Generate some sample WAF logs for testing
	mockLogs := []struct {
		ClientIP   string
		Method     string
		URI        string
		UserAgent  string
		AttackType string
		Blocked    bool
	}{
		{"192.168.1.100", "GET", "/", "Mozilla/5.0", "", false},
		{"10.0.0.50", "POST", "/login", "curl/7.68.0", "", false},
		{"203.0.113.5", "GET", "/admin", "BadBot/1.0", "Directory Traversal", true},
		{"198.51.100.10", "POST", "/search", "Mozilla/5.0", "SQL Injection", true},
		{"192.0.2.1", "GET", "/profile", "Chrome/91.0", "", false},
		{"203.0.113.15", "GET", "/api/users", "AttackScript", "XSS", true},
		{"10.0.0.25", "GET", "/dashboard", "Firefox/89.0", "", false},
		{"198.51.100.20", "POST", "/upload", "Exploit/2.0", "File Upload Attack", true},
	}
	
	for _, mockLog := range mockLogs {
		h.wafService.AddMockLog(
			mockLog.ClientIP,
			mockLog.Method,
			mockLog.URI,
			mockLog.UserAgent,
			mockLog.AttackType,
			mockLog.Blocked,
		)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Test logs generated successfully",
		"count":   len(mockLogs),
	})
}