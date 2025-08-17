package dto

import "time"

type WAFLog struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	ClientIP    string    `json:"client_ip"`
	Method      string    `json:"method"`
	URL         string    `json:"url"`
	UserAgent   string    `json:"user_agent"`
	AttackType  string    `json:"attack_type"`
	RuleID      string    `json:"rule_id"`
	Message     string    `json:"message"`
	Blocked     bool      `json:"blocked"`
	Severity    string    `json:"severity"`
	RawLog      string    `json:"raw_log"`
}

type WAFStats struct {
	TotalRequests   int64            `json:"total_requests"`
	BlockedRequests int64            `json:"blocked_requests"`
	AttacksByType   map[string]int64 `json:"attacks_by_type"`
	TopIPs          []IPStat         `json:"top_ips"`
	RecentLogs      []WAFLog         `json:"recent_logs"`
	Timestamp       time.Time        `json:"timestamp"`
}

type IPStat struct {
	IP       string `json:"ip"`
	Requests int64  `json:"requests"`
	Blocked  int64  `json:"blocked"`
}

type CustomRule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	RuleText    string    `json:"rule_text" binding:"required"`
	Enabled     bool      `json:"enabled"`
	Severity    string    `json:"severity" binding:"required"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CustomRuleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	RuleText    string `json:"rule_text" binding:"required"`
	Enabled     bool   `json:"enabled"`
	Severity    string `json:"severity" binding:"required,oneof=LOW MEDIUM HIGH CRITICAL"`
}

type CustomRuleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	RuleText    string    `json:"rule_text"`
	Enabled     bool      `json:"enabled"`
	Severity    string    `json:"severity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SecurityTest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	TestType    string            `json:"test_type"`
	Payloads    []string          `json:"payloads"`
	Results     []SecurityResult  `json:"results"`
	CreatedAt   time.Time         `json:"created_at"`
}

type SecurityResult struct {
	Payload    string `json:"payload"`
	Blocked    bool   `json:"blocked"`
	StatusCode int    `json:"status_code"`
	Response   string `json:"response"`
}

type SecurityTestRequest struct {
	TestType string   `json:"test_type" binding:"required,oneof=sql_injection xss path_traversal command_injection"`
	Payloads []string `json:"payloads"`
}