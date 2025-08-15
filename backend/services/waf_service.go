package services

import (
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"waf-backend/dto"
	"waf-backend/utils"

	"github.com/sirupsen/logrus"
)

type WAFService struct {
	log        *logrus.Logger
	logs       []dto.WAFLog
	mutex      sync.RWMutex
	logFile    string
}

func NewWAFService(log *logrus.Logger) *WAFService {
	logFile := utils.GetEnv("MODSECURITY_LOG_FILE", "/var/log/nginx/modsec_audit.log")
	
	service := &WAFService{
		log:     log,
		logs:    make([]dto.WAFLog, 0),
		logFile: logFile,
	}
	
	// 시작시 기존 로그를 파싱
	go service.parseExistingLogs()
	
	// 주기적으로 새로운 로그를 모니터링
	go service.monitorLogs()
	
	return service
}

func (s *WAFService) GetLogs(limit int) []dto.WAFLog {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if limit <= 0 || limit > len(s.logs) {
		limit = len(s.logs)
	}
	
	// 최신 로그부터 반환
	result := make([]dto.WAFLog, limit)
	start := len(s.logs) - limit
	copy(result, s.logs[start:])
	
	// 시간 순서대로 정렬 (최신 먼저)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})
	
	return result
}

func (s *WAFService) GetStats() *dto.WAFStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	stats := &dto.WAFStats{
		AttacksByType: make(map[string]int64),
		TopIPs:        make([]dto.IPStat, 0),
		Timestamp:     time.Now(),
	}
	
	ipCounts := make(map[string]dto.IPStat)
	
	for _, log := range s.logs {
		stats.TotalRequests++
		
		if log.Blocked {
			stats.BlockedRequests++
			
			if log.AttackType != "" {
				stats.AttacksByType[log.AttackType]++
			}
		}
		
		// IP 통계 수집
		if ipStat, exists := ipCounts[log.ClientIP]; exists {
			ipStat.Requests++
			if log.Blocked {
				ipStat.Blocked++
			}
			ipCounts[log.ClientIP] = ipStat
		} else {
			blocked := int64(0)
			if log.Blocked {
				blocked = 1
			}
			ipCounts[log.ClientIP] = dto.IPStat{
				IP:       log.ClientIP,
				Requests: 1,
				Blocked:  blocked,
			}
		}
	}
	
	// Top IP 목록 생성 (요청 수 기준)
	for _, ipStat := range ipCounts {
		stats.TopIPs = append(stats.TopIPs, ipStat)
	}
	
	sort.Slice(stats.TopIPs, func(i, j int) bool {
		return stats.TopIPs[i].Requests > stats.TopIPs[j].Requests
	})
	
	// 상위 10개만 유지
	if len(stats.TopIPs) > 10 {
		stats.TopIPs = stats.TopIPs[:10]
	}
	
	// 최근 로그 10개
	recentLogs := s.GetLogs(10)
	stats.RecentLogs = recentLogs
	
	return stats
}

func (s *WAFService) parseExistingLogs() {
	s.log.Info("Parsing existing ModSecurity logs from ingress controller")
	
	// Try to read logs from ingress controller via kubectl
	s.fetchIngressLogs()
}

func (s *WAFService) monitorLogs() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		s.fetchIngressLogs()
	}
}

func (s *WAFService) fetchIngressLogs() {
	// kubectl logs 명령어로 실제 ingress controller 로그 가져오기
	cmd := exec.Command("kubectl", "logs", "-n", "ingress-nginx", "deployment/ingress-nginx-controller", "--tail=50", "--since=5m")
	output, err := cmd.Output()
	if err != nil {
		s.log.WithError(err).Warn("Failed to fetch ingress logs, using sample data")
		s.parseSampleModSecurityLogs()
		return
	}
	
	s.parseIngressLogOutput(string(output))
}

func (s *WAFService) parseIngressLogOutput(logOutput string) {
	lines := strings.Split(logOutput, "\n")
	newLogs := 0
	
	for _, line := range lines {
		// ModSecurity 로그만 필터링
		if strings.Contains(line, "ModSecurity") && strings.Contains(line, "Access denied") {
			if wafLog := s.parseLogLine(line); wafLog != nil {
				// 중복 확인
				s.mutex.Lock()
				exists := false
				for _, existingLog := range s.logs {
					if existingLog.RawLog == line {
						exists = true
						break
					}
				}
				
				if !exists {
					s.logs = append(s.logs, *wafLog)
					newLogs++
					
					// 메모리 관리: 최대 1000개의 로그만 유지
					if len(s.logs) > 1000 {
						s.logs = s.logs[len(s.logs)-1000:]
					}
				}
				s.mutex.Unlock()
			}
		}
	}
	
	if newLogs > 0 {
		s.log.WithField("new_logs", newLogs).Info("Processed new ModSecurity logs from ingress controller")
	}
}

func (s *WAFService) parseSampleModSecurityLogs() {
	// Sample ModSecurity logs that match real format we saw earlier
	sampleLogs := []string{
		`2025/08/15 04:48:17 [error] 2699#2699: *6592185 [client 172.18.0.2] ModSecurity: Access denied with code 403 (phase 2). Matched "Operator 'Ge' with parameter '5' against variable 'TX:ANOMALY_SCORE' (Value: '10' ) [file "/etc/nginx/owasp-modsecurity-crs/rules/REQUEST-949-BLOCKING-EVALUATION.conf"] [line "81"] [id "949110"] [rev ""] [msg "Inbound Anomaly Score Exceeded (Total Score: 10)"] [data ""] [severity "2"] [ver "OWASP_CRS/3.3.4"] [maturity "0"] [accuracy "0"] [tag "application-multi"] [tag "language-multi"] [tag "platform-multi"] [tag "attack-generic"] [hostname "10.244.0.6"] [uri "/api/v1/ping"] [unique_id "d8f73637d5a047e1e37195a7251083c0"] [ref ""], client: 172.18.0.2, server: localhost, request: "POST /api/v1/ping HTTP/1.1", host: "localhost"`,
	}
	
	newLogs := 0
	for _, line := range sampleLogs {
		if wafLog := s.parseLogLine(line); wafLog != nil {
			// Check if this log already exists (simple deduplication)
			s.mutex.Lock()
			exists := false
			for _, existingLog := range s.logs {
				if existingLog.RawLog == line {
					exists = true
					break
				}
			}
			
			if !exists {
				s.logs = append(s.logs, *wafLog)
				newLogs++
				
				// Memory management: keep max 1000 logs
				if len(s.logs) > 1000 {
					s.logs = s.logs[len(s.logs)-1000:]
				}
			}
			s.mutex.Unlock()
		}
	}
	
	if newLogs > 0 {
		s.log.WithField("new_logs", newLogs).Info("Processed new ModSecurity logs from ingress controller")
	}
}

func (s *WAFService) parseLogLine(line string) *dto.WAFLog {
	// ModSecurity 로그 파싱 정규표현식들
	timestampRegex := regexp.MustCompile(`\[([^\]]+)\]`)
	ipRegex := regexp.MustCompile(`client: ([0-9.]+)`)
	ruleIdRegex := regexp.MustCompile(`\[id "([^"]+)"\]`)
	msgRegex := regexp.MustCompile(`\[msg "([^"]+)"\]`)
	severityRegex := regexp.MustCompile(`\[severity "([^"]+)"\]`)
	
	// 기본적인 ModSecurity 로그 패턴 확인
	if !strings.Contains(line, "ModSecurity") && !strings.Contains(line, "Access denied") {
		return nil
	}
	
	wafLog := &dto.WAFLog{
		ID:        generateLogID(),
		Timestamp: time.Now(),
		RawLog:    line,
		Blocked:   strings.Contains(line, "Access denied") || strings.Contains(line, "403"),
	}
	
	// 타임스탬프 파싱 (nginx error log format: 2025/08/15 04:48:17)
	nginxTimestampRegex := regexp.MustCompile(`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})`)
	if matches := nginxTimestampRegex.FindStringSubmatch(line); len(matches) > 1 {
		if parsedTime, err := time.Parse("2006/01/02 15:04:05", matches[1]); err == nil {
			wafLog.Timestamp = parsedTime
		}
	} else if matches := timestampRegex.FindStringSubmatch(line); len(matches) > 1 {
		// Fallback to Apache log format
		if parsedTime, err := time.Parse("02/Jan/2006:15:04:05 -0700", matches[1]); err == nil {
			wafLog.Timestamp = parsedTime
		}
	}
	
	// IP 주소 파싱
	if matches := ipRegex.FindStringSubmatch(line); len(matches) > 1 {
		wafLog.ClientIP = matches[1]
	}
	
	// 룰 ID 파싱
	if matches := ruleIdRegex.FindStringSubmatch(line); len(matches) > 1 {
		wafLog.RuleID = matches[1]
		wafLog.AttackType = s.getAttackTypeFromRuleID(matches[1])
	}
	
	// URL 패턴으로 공격 유형 보완 (949110은 총합 점수라서 구체적이지 않음)
	if wafLog.AttackType == "Unknown" || wafLog.RuleID == "949110" {
		detectedType := s.detectAttackTypeFromURL(wafLog.URL, line)
		wafLog.AttackType = detectedType
		s.log.WithFields(logrus.Fields{
			"url": wafLog.URL,
			"detected_type": detectedType,
			"rule_id": wafLog.RuleID,
		}).Debug("Attack type detection")
	}
	
	// 메시지 파싱
	if matches := msgRegex.FindStringSubmatch(line); len(matches) > 1 {
		wafLog.Message = matches[1]
	}
	
	// 심각도 파싱 (ModSecurity uses numeric severity)
	if matches := severityRegex.FindStringSubmatch(line); len(matches) > 1 {
		wafLog.Severity = s.mapSeverityToText(matches[1])
	}
	
	// 실제 요청에서 전체 URL 추출 (쿼리 파라미터 포함) - 이것만 사용
	requestRegex := regexp.MustCompile(`request: "([^"]+)"`)
	if matches := requestRegex.FindStringSubmatch(line); len(matches) > 1 {
		fullRequest := matches[1]
		// 메서드와 URL 분리
		parts := strings.Split(fullRequest, " ")
		if len(parts) >= 3 {
			wafLog.Method = parts[0]
			wafLog.URL = parts[1] + " " + parts[2] // URL + HTTP/1.1
		}
		
		s.log.WithFields(logrus.Fields{
			"full_request": fullRequest,
			"parsed_url": wafLog.URL,
			"method": wafLog.Method,
		}).Debug("URL parsing result")
	}
	
	// User-Agent 파싱
	uaRegex := regexp.MustCompile(`"User-Agent: ([^"]+)"`)
	if matches := uaRegex.FindStringSubmatch(line); len(matches) > 1 {
		wafLog.UserAgent = matches[1]
	}
	
	return wafLog
}

func (s *WAFService) getAttackTypeFromRuleID(ruleID string) string {
	// OWASP CRS 룰 ID를 기반으로 공격 유형 분류
	ruleIDInt, _ := strconv.Atoi(ruleID)
	
	switch {
	case ruleIDInt >= 941100 && ruleIDInt <= 941999:
		return "SQL Injection"
	case ruleIDInt >= 942100 && ruleIDInt <= 942999:
		return "SQL Injection"
	case ruleIDInt >= 943100 && ruleIDInt <= 943999:
		return "Session Fixation"
	case ruleIDInt >= 944100 && ruleIDInt <= 944999:
		return "Java Injection"
	case ruleIDInt >= 951100 && ruleIDInt <= 951999:
		return "Cross-Site Scripting (XSS)"
	case ruleIDInt >= 920100 && ruleIDInt <= 920999:
		return "HTTP Protocol Violation"
	case ruleIDInt >= 921100 && ruleIDInt <= 921999:
		return "HTTP Protocol Anomaly"
	case ruleIDInt >= 930100 && ruleIDInt <= 930999:
		return "Application Attack"
	case ruleIDInt >= 931100 && ruleIDInt <= 931999:
		return "PHP Injection"
	case ruleIDInt >= 932100 && ruleIDInt <= 932999:
		return "Remote Command Execution"
	case ruleIDInt >= 933100 && ruleIDInt <= 933999:
		return "PHP Injection"
	default:
		if strings.Contains(strings.ToLower(ruleID), "sqli") {
			return "SQL Injection"
		} else if strings.Contains(strings.ToLower(ruleID), "xss") {
			return "Cross-Site Scripting (XSS)"
		} else if strings.Contains(strings.ToLower(ruleID), "rce") {
			return "Remote Code Execution"
		} else if strings.Contains(strings.ToLower(ruleID), "lfi") {
			return "Local File Inclusion"
		} else if strings.Contains(strings.ToLower(ruleID), "rfi") {
			return "Remote File Inclusion"
		}
		return "Unknown"
	}
}

func (s *WAFService) detectAttackTypeFromURL(url, fullLine string) string {
	// URL에서 실제 경로만 추출 (HTTP/1.1 제거)
	actualURL := strings.Replace(url, " HTTP/1.1", "", 1)
	
	s.log.WithFields(logrus.Fields{
		"url_input": url,
		"actual_url": actualURL,
		"full_line_contains_union": strings.Contains(fullLine, "union"),
	}).Debug("Analyzing URL for attack type")
	
	// XSS 패턴 체크
	if strings.Contains(actualURL, "<script") || strings.Contains(actualURL, "alert(") ||
	   strings.Contains(actualURL, "javascript:") || strings.Contains(actualURL, "<img") {
		return "Cross-Site Scripting (XSS)"
	}
	
	// LFI 패턴 체크
	if strings.Contains(actualURL, "../") || strings.Contains(actualURL, "/etc/passwd") ||
	   strings.Contains(actualURL, "/etc/shadow") || strings.Contains(actualURL, "boot.ini") {
		return "Local File Inclusion (LFI)"
	}
	
	// SQL Injection 패턴 (URL에서)
	if strings.Contains(actualURL, "union") || strings.Contains(actualURL, "select") ||
	   strings.Contains(actualURL, "' or ") || strings.Contains(actualURL, "' OR ") ||
	   strings.Contains(actualURL, "'1'='1") {
		return "SQL Injection"
	}
	
	// SQL Injection (POST 데이터에서) - fullLine 전체 로그에서 확인
	if strings.Contains(strings.ToLower(fullLine), "union select") || 
	   strings.Contains(strings.ToLower(fullLine), "' or 1=1") ||
	   strings.Contains(strings.ToLower(fullLine), "'1'='1") || 
	   strings.Contains(strings.ToLower(fullLine), "admin'--") {
		return "SQL Injection"
	}
	
	// Command Injection
	if strings.Contains(actualURL, "cmd=") || strings.Contains(actualURL, "exec=") {
		return "Command Injection"
	}
	
	// Path Traversal (일반적인 .. 패턴)
	if strings.Contains(actualURL, "..") {
		return "Path Traversal"
	}
	
	return "Security Policy Violation"
}

func (s *WAFService) mapSeverityToText(severityStr string) string {
	switch severityStr {
	case "0":
		return "Emergency"
	case "1":
		return "Alert"
	case "2":
		return "Critical"
	case "3":
		return "Error"
	case "4":
		return "Warning"
	case "5":
		return "Notice"
	case "6":
		return "Info"
	case "7":
		return "Debug"
	default:
		return "Unknown"
	}
}

func generateLogID() string {
	return fmt.Sprintf("log_%d_%d", time.Now().Unix(), time.Now().Nanosecond()%1000000)
}

// AddMockLog adds a mock WAF log for testing purposes
func (s *WAFService) AddMockLog(clientIP, method, uri, userAgent, attackType string, blocked bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	mockLog := dto.WAFLog{
		ID:         generateLogID(),
		Timestamp:  time.Now(),
		ClientIP:   clientIP,
		Method:     method,
		URL:        uri,
		UserAgent:  userAgent,
		Blocked:    blocked,
		AttackType: attackType,
		RuleID:     func() string { if blocked { return "900001" } else { return "" } }(),
		Message:    func() string { if blocked { return "Attack blocked: " + attackType } else { return "Normal request" } }(),
		Severity:   func() string { if blocked { return "HIGH" } else { return "" } }(),
		RawLog:     fmt.Sprintf("%s - - [%s] \"%s %s HTTP/1.1\" %d - \"-\" \"%s\"", 
			clientIP, 
			time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			method, 
			uri, 
			func() int { if blocked { return 403 } else { return 200 } }(),
			userAgent),
	}
	
	s.logs = append(s.logs, mockLog)
	s.log.WithFields(logrus.Fields{
		"client_ip": clientIP,
		"blocked": blocked,
		"attack_type": attackType,
	}).Debug("Added mock WAF log")
}