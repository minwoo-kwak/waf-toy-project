package services

import (
	"bufio"
	"fmt"
	"os"
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
	s.log.Info("Parsing existing ModSecurity logs")
	
	file, err := os.Open(s.logFile)
	if err != nil {
		s.log.WithError(err).Warn("Could not open ModSecurity log file, starting with empty logs")
		return
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	count := 0
	
	for scanner.Scan() {
		line := scanner.Text()
		if wafLog := s.parseLogLine(line); wafLog != nil {
			s.mutex.Lock()
			s.logs = append(s.logs, *wafLog)
			s.mutex.Unlock()
			count++
		}
	}
	
	s.log.WithField("count", count).Info("Finished parsing existing logs")
}

func (s *WAFService) monitorLogs() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	lastPosition := int64(0)
	
	for range ticker.C {
		file, err := os.Open(s.logFile)
		if err != nil {
			continue
		}
		
		stat, err := file.Stat()
		if err != nil {
			file.Close()
			continue
		}
		
		// 파일이 작아졌으면 로그가 로테이션된 것으로 간주
		if stat.Size() < lastPosition {
			lastPosition = 0
		}
		
		// 마지막 위치부터 읽기
		file.Seek(lastPosition, 0)
		scanner := bufio.NewScanner(file)
		
		newLogs := 0
		for scanner.Scan() {
			line := scanner.Text()
			if wafLog := s.parseLogLine(line); wafLog != nil {
				s.mutex.Lock()
				s.logs = append(s.logs, *wafLog)
				
				// 메모리 관리: 최대 1000개의 로그만 유지
				if len(s.logs) > 1000 {
					s.logs = s.logs[len(s.logs)-1000:]
				}
				s.mutex.Unlock()
				newLogs++
			}
		}
		
		lastPosition, _ = file.Seek(0, 1)
		file.Close()
		
		if newLogs > 0 {
			s.log.WithField("new_logs", newLogs).Debug("Processed new WAF logs")
		}
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
	
	// 타임스탬프 파싱
	if matches := timestampRegex.FindStringSubmatch(line); len(matches) > 1 {
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
	
	// 메시지 파싱
	if matches := msgRegex.FindStringSubmatch(line); len(matches) > 1 {
		wafLog.Message = matches[1]
	}
	
	// 심각도 파싱
	if matches := severityRegex.FindStringSubmatch(line); len(matches) > 1 {
		wafLog.Severity = matches[1]
	}
	
	// HTTP 메서드와 URL 파싱
	methodUrlRegex := regexp.MustCompile(`"(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS) ([^"]+)"`)
	if matches := methodUrlRegex.FindStringSubmatch(line); len(matches) > 2 {
		wafLog.Method = matches[1]
		wafLog.URL = matches[2]
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

func generateLogID() string {
	return fmt.Sprintf("log_%d_%d", time.Now().Unix(), time.Now().Nanosecond()%1000000)
}