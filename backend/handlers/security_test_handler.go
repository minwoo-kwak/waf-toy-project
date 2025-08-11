package handlers

import (
	"net/http"
	"time"
	"waf-backend/dto"
	"waf-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SecurityTestHandler struct {
	securityTestService *services.SecurityTestService
	log                 *logrus.Logger
}

func NewSecurityTestHandler(securityTestService *services.SecurityTestService, log *logrus.Logger) *SecurityTestHandler {
	return &SecurityTestHandler{
		securityTestService: securityTestService,
		log:                 log,
	}
}

func (h *SecurityTestHandler) RunSecurityTest(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
			"code":  "ERR_NO_USER_ID",
		})
		return
	}
	
	var req dto.SecurityTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Error("Invalid security test request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "ERR_INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}
	
	h.log.WithFields(logrus.Fields{
		"user_id":   userID,
		"test_type": req.TestType,
		"payloads":  len(req.Payloads),
	}).Info("Running security test")
	
	test, err := h.securityTestService.RunSecurityTest(req.TestType, req.Payloads)
	if err != nil {
		h.log.WithError(err).Error("Failed to run security test")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to run security test",
			"code":  "ERR_TEST_FAILED",
			"details": err.Error(),
		})
		return
	}
	
	// 테스트 결과 통계 계산
	totalTests := len(test.Results)
	blockedTests := 0
	passedTests := 0
	
	for _, result := range test.Results {
		if result.Blocked {
			blockedTests++
		} else {
			passedTests++
		}
	}
	
	h.log.WithFields(logrus.Fields{
		"user_id":      userID,
		"test_id":      test.ID,
		"test_type":    test.TestType,
		"total_tests":  totalTests,
		"blocked_tests": blockedTests,
		"passed_tests": passedTests,
	}).Info("Security test completed")
	
	c.JSON(http.StatusOK, gin.H{
		"test": test,
		"summary": gin.H{
			"total_tests":    totalTests,
			"blocked_tests":  blockedTests,
			"passed_tests":   passedTests,
			"block_rate":     float64(blockedTests) / float64(totalTests) * 100,
			"effectiveness":  getEffectivenessRating(blockedTests, totalTests),
		},
		"message": "Security test completed successfully",
	})
}

func (h *SecurityTestHandler) GetTestTypes(c *gin.Context) {
	h.log.Debug("Security test types requested")
	
	testTypes := []gin.H{
		{
			"id":          "sql_injection",
			"name":        "SQL Injection",
			"description": "Tests for SQL injection vulnerabilities using common attack patterns",
			"severity":    "HIGH",
		},
		{
			"id":          "xss",
			"name":        "Cross-Site Scripting (XSS)",
			"description": "Tests for XSS vulnerabilities using various script injection techniques",
			"severity":    "HIGH",
		},
		{
			"id":          "path_traversal",
			"name":        "Path Traversal",
			"description": "Tests for directory traversal vulnerabilities",
			"severity":    "MEDIUM",
		},
		{
			"id":          "command_injection",
			"name":        "Command Injection",
			"description": "Tests for OS command injection vulnerabilities",
			"severity":    "CRITICAL",
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"test_types": testTypes,
		"count":      len(testTypes),
	})
}

func (h *SecurityTestHandler) GetQuickTests(c *gin.Context) {
	userID := c.GetString("user_id")
	
	h.log.WithField("user_id", userID).Info("Running quick security tests")
	
	// 빠른 테스트를 위한 간단한 페이로드들
	quickTests := []struct {
		TestType string
		Payloads []string
	}{
		{
			TestType: "sql_injection",
			Payloads: []string{"' OR '1'='1", "1; DROP TABLE users--"},
		},
		{
			TestType: "xss",
			Payloads: []string{"<script>alert('XSS')</script>", "<img src=x onerror=alert('XSS')>"},
		},
		{
			TestType: "path_traversal",
			Payloads: []string{"../../../etc/passwd", "..\\..\\..\\boot.ini"},
		},
	}
	
	results := make([]gin.H, 0)
	
	for _, quickTest := range quickTests {
		test, err := h.securityTestService.RunSecurityTest(quickTest.TestType, quickTest.Payloads)
		if err != nil {
			h.log.WithError(err).Error("Quick test failed")
			continue
		}
		
		blockedCount := 0
		for _, result := range test.Results {
			if result.Blocked {
				blockedCount++
			}
		}
		
		results = append(results, gin.H{
			"test_type":     quickTest.TestType,
			"total_tests":   len(test.Results),
			"blocked_tests": blockedCount,
			"effectiveness": getEffectivenessRating(blockedCount, len(test.Results)),
			"results":       test.Results,
		})
	}
	
	h.log.WithFields(logrus.Fields{
		"user_id":      userID,
		"quick_tests":  len(results),
	}).Info("Quick security tests completed")
	
	c.JSON(http.StatusOK, gin.H{
		"quick_tests": results,
		"summary": gin.H{
			"timestamp":   time.Now(),
			"total_types": len(results),
		},
	})
}

func getEffectivenessRating(blocked, total int) string {
	if total == 0 {
		return "No Data"
	}
	
	percentage := float64(blocked) / float64(total) * 100
	
	switch {
	case percentage >= 90:
		return "Excellent"
	case percentage >= 70:
		return "Good"
	case percentage >= 50:
		return "Fair"
	case percentage >= 30:
		return "Poor"
	default:
		return "Critical"
	}
}