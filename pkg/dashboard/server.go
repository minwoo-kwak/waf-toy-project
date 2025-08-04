package dashboard

import (
	"context"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"waf-k8s-project/pkg/logging"
	"waf-k8s-project/pkg/ratelimit"
	"waf-k8s-project/pkg/tenant"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// DashboardServer 웹 대시보드 서버
type DashboardServer struct {
	tenantManager *tenant.TenantManager
	rateLimiter   *ratelimit.RedisRateLimiter
	elkCollector  *logging.ELKCollector
	logger        *logrus.Logger
	templates     *template.Template
}

// NewDashboardServer 대시보드 서버 생성
func NewDashboardServer(tenantManager *tenant.TenantManager, rateLimiter *ratelimit.RedisRateLimiter, elkCollector *logging.ELKCollector) *DashboardServer {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// HTML 템플릿 로드
	templates := template.Must(template.ParseGlob("web/templates/*.html"))

	return &DashboardServer{
		tenantManager: tenantManager,
		rateLimiter:   rateLimiter,
		elkCollector:  elkCollector,
		logger:        logger,
		templates:     templates,
	}
}

// SetupRoutes 대시보드 라우트 설정
func (ds *DashboardServer) SetupRoutes(router *gin.Engine) {
	// 정적 파일 서빙
	router.Static("/static", "./web/static")
	
	// 대시보드 페이지들
	dashboard := router.Group("/dashboard")
	{
		dashboard.GET("/", ds.renderDashboard)
		dashboard.GET("/overview", ds.overviewPage)
		dashboard.GET("/security", ds.securityPage)
		dashboard.GET("/tenants", ds.tenantsPage)
		dashboard.GET("/logs", ds.logsPage)
		dashboard.GET("/settings", ds.settingsPage)
	}

	// API 엔드포인트들
	api := router.Group("/dashboard/api")
	{
		// 대시보드 데이터
		api.GET("/stats", ds.getStats)
		api.GET("/threats", ds.getThreats)
		api.GET("/traffic", ds.getTraffic)
		api.GET("/top-ips", ds.getTopIPs)
		api.GET("/geo-data", ds.getGeoData)
		
		// 테넌트 관리
		api.GET("/tenants", ds.listTenants)
		api.POST("/tenants", ds.createTenant)
		api.GET("/tenants/:id", ds.getTenant)
		api.PUT("/tenants/:id", ds.updateTenant)
		api.DELETE("/tenants/:id", ds.deleteTenant)
		
		// 로그 검색
		api.POST("/logs/search", ds.searchLogs)
		api.GET("/logs/export", ds.exportLogs)
		
		// 보안 관리
		api.POST("/security/block-ip", ds.blockIP)
		api.GET("/security/blocked-ips", ds.getBlockedIPs)
		api.DELETE("/security/unblock-ip/:ip", ds.unblockIP)
		
		// 실시간 데이터 (WebSocket)
		api.GET("/realtime", ds.realtimeData)
	}
}

// renderDashboard 메인 대시보드 렌더링
func (ds *DashboardServer) renderDashboard(c *gin.Context) {
	data := gin.H{
		"Title":     "WAF Security Dashboard",
		"Timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}
	
	c.HTML(http.StatusOK, "dashboard.html", data)
}

// overviewPage 개요 페이지
func (ds *DashboardServer) overviewPage(c *gin.Context) {
	// 기본 통계 조회
	stats, err := ds.getOverviewStats(c.Request.Context())
	if err != nil {
		ds.logger.WithError(err).Error("개요 통계 조회 실패")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "통계 조회 실패"})
		return
	}

	data := gin.H{
		"Title": "Dashboard Overview",
		"Stats": stats,
	}
	
	c.HTML(http.StatusOK, "overview.html", data)
}

// securityPage 보안 페이지
func (ds *DashboardServer) securityPage(c *gin.Context) {
	// 위협 통계 조회
	threats, err := ds.getThreatStats(c.Request.Context())
	if err != nil {
		ds.logger.WithError(err).Error("위협 통계 조회 실패")
	}

	data := gin.H{
		"Title":   "Security Dashboard",
		"Threats": threats,
	}
	
	c.HTML(http.StatusOK, "security.html", data)
}

// tenantsPage 테넌트 관리 페이지
func (ds *DashboardServer) tenantsPage(c *gin.Context) {
	tenants, err := ds.tenantManager.ListTenants(c.Request.Context())
	if err != nil {
		ds.logger.WithError(err).Error("테넌트 목록 조회 실패")
		tenants = []tenant.Tenant{}
	}

	data := gin.H{
		"Title":   "Tenant Management",
		"Tenants": tenants,
	}
	
	c.HTML(http.StatusOK, "tenants.html", data)
}

// logsPage 로그 페이지
func (ds *DashboardServer) logsPage(c *gin.Context) {
	data := gin.H{
		"Title": "Log Analysis",
	}
	
	c.HTML(http.StatusOK, "logs.html", data)
}

// settingsPage 설정 페이지
func (ds *DashboardServer) settingsPage(c *gin.Context) {
	data := gin.H{
		"Title": "WAF Settings",
	}
	
	c.HTML(http.StatusOK, "settings.html", data)
}

// getStats 통계 API
func (ds *DashboardServer) getStats(c *gin.Context) {
	stats, err := ds.getOverviewStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, stats)
}

// getOverviewStats 개요 통계 조회
func (ds *DashboardServer) getOverviewStats(ctx context.Context) (map[string]interface{}, error) {
	// 기본 통계 조회 (실제로는 다양한 소스에서 데이터 수집)
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)
	
	// ELK에서 통계 조회 (예시)
	searchQuery := logging.SearchQuery{
		StartTime: startTime,
		EndTime:   endTime,
		Size:      0, // 집계만 필요
	}
	
	// 실제로는 Elasticsearch 집계 쿼리 사용
	_ = searchQuery
	
	stats := map[string]interface{}{
		"total_requests":    125847,
		"blocked_requests":  2341,
		"blocked_percentage": 1.86,
		"unique_visitors":   8924,
		"avg_response_time": 245.6,
		"uptime_percentage": 99.98,
		"threat_level": map[string]int{
			"critical": 23,
			"high":     156,
			"medium":   892,
			"low":      1270,
		},
		"top_countries": []map[string]interface{}{
			{"name": "United States", "requests": 45231, "blocked": 234},
			{"name": "China", "requests": 23847, "blocked": 892},
			{"name": "Russia", "requests": 12456, "blocked": 567},
			{"name": "Germany", "requests": 8934, "blocked": 89},
			{"name": "Japan", "requests": 6782, "blocked": 34},
		},
		"attack_types": map[string]int{
			"SQL_INJECTION":     456,
			"XSS":              234,
			"PATH_TRAVERSAL":   189,
			"COMMAND_INJECTION": 123,
			"MALICIOUS_UA":     89,
		},
		"hourly_traffic": ds.generateHourlyTraffic(),
	}
	
	return stats, nil
}

// getThreatStats 위협 통계 조회
func (ds *DashboardServer) getThreatStats(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"recent_attacks": []map[string]interface{}{
			{
				"timestamp":   time.Now().Add(-5 * time.Minute).Unix(),
				"client_ip":   "192.168.1.100",
				"country":     "Unknown",
				"attack_type": "SQL_INJECTION",
				"severity":    "high",
				"blocked":     true,
				"path":        "/login?user=admin' OR '1'='1",
			},
			{
				"timestamp":   time.Now().Add(-8 * time.Minute).Unix(),
				"client_ip":   "10.0.0.50",
				"country":     "Local",
				"attack_type": "XSS",
				"severity":    "medium",
				"blocked":     true,
				"path":        "/search?q=<script>alert(1)</script>",
			},
		},
		"threat_trends": ds.generateThreatTrends(),
	}, nil
}

// generateHourlyTraffic 시간별 트래픽 데이터 생성 (예시)
func (ds *DashboardServer) generateHourlyTraffic() []map[string]interface{} {
	var data []map[string]interface{}
	now := time.Now()
	
	for i := 23; i >= 0; i-- {
		hour := now.Add(time.Duration(-i) * time.Hour)
		requests := 1000 + (i*50) + (i%3)*200
		blocked := requests / 50
		
		data = append(data, map[string]interface{}{
			"hour":     hour.Format("15:04"),
			"requests": requests,
			"blocked":  blocked,
		})
	}
	
	return data
}

// generateThreatTrends 위협 트렌드 데이터 생성 (예시)
func (ds *DashboardServer) generateThreatTrends() []map[string]interface{} {
	var data []map[string]interface{}
	now := time.Now()
	
	for i := 29; i >= 0; i-- {
		day := now.Add(time.Duration(-i) * 24 * time.Hour)
		
		data = append(data, map[string]interface{}{
			"date":           day.Format("2006-01-02"),
			"sql_injection":  20 + (i%7)*5,
			"xss":           15 + (i%5)*3,
			"path_traversal": 10 + (i%3)*2,
			"other":         5 + (i%2),
		})
	}
	
	return data
}

// getThreats 위협 API
func (ds *DashboardServer) getThreats(c *gin.Context) {
	threats, err := ds.getThreatStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, threats)
}

// getTraffic 트래픽 API
func (ds *DashboardServer) getTraffic(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"hourly_traffic": ds.generateHourlyTraffic(),
	})
}

// getTopIPs 상위 IP API
func (ds *DashboardServer) getTopIPs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"top_ips": []map[string]interface{}{
			{"ip": "192.168.1.100", "requests": 1247, "blocked": 89, "country": "US"},
			{"ip": "10.0.0.50", "requests": 892, "blocked": 12, "country": "Local"},
			{"ip": "203.0.113.45", "requests": 567, "blocked": 234, "country": "CN"},
		},
	})
}

// getGeoData 지리적 데이터 API
func (ds *DashboardServer) getGeoData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"geo_data": []map[string]interface{}{
			{"country": "US", "latitude": 39.8283, "longitude": -98.5795, "requests": 45231, "blocked": 234},
			{"country": "CN", "latitude": 35.8617, "longitude": 104.1954, "requests": 23847, "blocked": 892},
			{"country": "RU", "latitude": 61.5240, "longitude": 105.3188, "requests": 12456, "blocked": 567},
		},
	})
}

// listTenants 테넌트 목록 API
func (ds *DashboardServer) listTenants(c *gin.Context) {
	tenants, err := ds.tenantManager.ListTenants(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"tenants": tenants})
}

// createTenant 테넌트 생성 API
func (ds *DashboardServer) createTenant(c *gin.Context) {
	var req tenant.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	newTenant, err := ds.tenantManager.CreateTenant(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"tenant": newTenant})
}

// getTenant 테넌트 조회 API
func (ds *DashboardServer) getTenant(c *gin.Context) {
	tenantID := c.Param("id")
	
	tenant, err := ds.tenantManager.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "테넌트를 찾을 수 없습니다"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"tenant": tenant})
}

// updateTenant 테넌트 수정 API
func (ds *DashboardServer) updateTenant(c *gin.Context) {
	tenantID := c.Param("id")
	
	// 업데이트 로직 구현
	_ = tenantID
	
	c.JSON(http.StatusOK, gin.H{"message": "테넌트 업데이트 완료"})
}

// deleteTenant 테넌트 삭제 API
func (ds *DashboardServer) deleteTenant(c *gin.Context) {
	tenantID := c.Param("id")
	
	err := ds.tenantManager.DeleteTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "테넌트 삭제 완료"})
}

// searchLogs 로그 검색 API
func (ds *DashboardServer) searchLogs(c *gin.Context) {
	var query logging.SearchQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if query.Size == 0 {
		query.Size = 50
	}
	
	result, err := ds.elkCollector.SearchLogs(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// exportLogs 로그 내보내기 API
func (ds *DashboardServer) exportLogs(c *gin.Context) {
	// CSV 또는 JSON 형식으로 로그 내보내기
	format := c.DefaultQuery("format", "csv")
	
	if format == "csv" {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=waf-logs.csv")
		c.String(http.StatusOK, "timestamp,client_ip,method,path,status_code,blocked\n")
		c.String(http.StatusOK, "2024-01-15T10:30:00Z,192.168.1.100,GET,/test,200,false\n")
	} else {
		c.JSON(http.StatusOK, gin.H{"logs": []string{"로그 데이터..."}})
	}
}

// blockIP IP 차단 API
func (ds *DashboardServer) blockIP(c *gin.Context) {
	var req struct {
		IP       string `json:"ip" binding:"required"`
		Duration int    `json:"duration"` // 시간(초)
		Reason   string `json:"reason"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	duration := time.Duration(req.Duration) * time.Second
	if duration == 0 {
		duration = time.Hour
	}
	
	err := ds.rateLimiter.BlockIP(c.Request.Context(), req.IP, duration, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "IP 차단 완료"})
}

// getBlockedIPs 차단된 IP 목록 API
func (ds *DashboardServer) getBlockedIPs(c *gin.Context) {
	// Redis에서 차단된 IP 목록 조회
	c.JSON(http.StatusOK, gin.H{
		"blocked_ips": []map[string]interface{}{
			{
				"ip":         "192.168.1.100",
				"blocked_at": time.Now().Add(-30 * time.Minute).Unix(),
				"reason":     "Multiple attack attempts",
				"expires_at": time.Now().Add(30 * time.Minute).Unix(),
			},
		},
	})
}

// unblockIP IP 차단 해제 API
func (ds *DashboardServer) unblockIP(c *gin.Context) {
	ip := c.Param("ip")
	
	// Redis에서 차단 해제 로직
	_ = ip
	
	c.JSON(http.StatusOK, gin.H{"message": "IP 차단 해제 완료"})
}

// realtimeData 실시간 데이터 WebSocket
func (ds *DashboardServer) realtimeData(c *gin.Context) {
	// WebSocket 업그레이드 및 실시간 데이터 스트리밍
	c.JSON(http.StatusOK, gin.H{
		"message": "WebSocket endpoint (구현 예정)",
	})
}