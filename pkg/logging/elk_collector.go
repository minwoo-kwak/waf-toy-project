package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ELKCollector ELK Stack으로 로그 전송하는 컬렉터
type ELKCollector struct {
	elasticsearchURL string
	indexPattern     string
	client          *http.Client
	logger          *logrus.Logger
	batchSize       int
	flushInterval   time.Duration
	logBuffer       []LogEntry
	bufferChan      chan LogEntry
}

// LogEntry 로그 엔트리 구조
type LogEntry struct {
	Timestamp    time.Time         `json:"@timestamp"`
	Level        string           `json:"level"`
	Message      string           `json:"message"`
	Service      string           `json:"service"`
	TenantID     string           `json:"tenant_id,omitempty"`
	ClientIP     string           `json:"client_ip,omitempty"`
	UserAgent    string           `json:"user_agent,omitempty"`
	Method       string           `json:"method,omitempty"`
	Path         string           `json:"path,omitempty"`
	StatusCode   int              `json:"status_code,omitempty"`
	ResponseTime float64          `json:"response_time_ms,omitempty"`
	RequestSize  int64            `json:"request_size,omitempty"`
	ResponseSize int64            `json:"response_size,omitempty"`
	Referer      string           `json:"referer,omitempty"`
	
	// 보안 관련 필드
	ThreatType    string   `json:"threat_type,omitempty"`
	ThreatScore   int      `json:"threat_score,omitempty"`
	AttackVector  string   `json:"attack_vector,omitempty"`
	Blocked       bool     `json:"blocked"`
	BlockReason   string   `json:"block_reason,omitempty"`
	RateLimited   bool     `json:"rate_limited"`
	
	// 추가 컨텍스트
	Fields        map[string]interface{} `json:"fields,omitempty"`
	Tags          []string              `json:"tags,omitempty"`
	
	// Geo 정보 (나중에 추가 가능)
	GeoIP         *GeoIPInfo            `json:"geoip,omitempty"`
}

// GeoIPInfo 지리적 위치 정보
type GeoIPInfo struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ASN         string  `json:"asn"`
	ISP         string  `json:"isp"`
}

// NewELKCollector ELK 컬렉터 생성
func NewELKCollector(elasticsearchURL, indexPattern string) *ELKCollector {
	collector := &ELKCollector{
		elasticsearchURL: elasticsearchURL,
		indexPattern:     indexPattern, // 예: "waf-logs-*"
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:        logrus.New(),
		batchSize:     100,
		flushInterval: 5 * time.Second,
		logBuffer:     make([]LogEntry, 0),
		bufferChan:    make(chan LogEntry, 1000),
	}
	
	// 백그라운드에서 배치 처리 시작
	go collector.startBatchProcessor()
	
	return collector
}

// LogRequest WAF 요청 로그 기록
func (e *ELKCollector) LogRequest(entry LogEntry) {
	// 기본 필드 설정
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	if entry.Service == "" {
		entry.Service = "waf-gateway"
	}
	
	// 태그 추가
	entry.Tags = append(entry.Tags, "waf", "request")
	
	if entry.Blocked {
		entry.Tags = append(entry.Tags, "blocked", "security")
	}
	
	if entry.RateLimited {
		entry.Tags = append(entry.Tags, "rate-limited")
	}
	
	if entry.ThreatType != "" {
		entry.Tags = append(entry.Tags, "threat:"+entry.ThreatType)
	}
	
	// 버퍼에 추가 (논블로킹)
	select {
	case e.bufferChan <- entry:
	default:
		e.logger.Warn("로그 버퍼가 가득참, 로그 손실 발생")
	}
}

// LogSecurityEvent 보안 이벤트 로깅
func (e *ELKCollector) LogSecurityEvent(event SecurityEvent) {
	entry := LogEntry{
		Timestamp:    time.Now(),
		Level:        event.Severity,
		Message:      event.Description,
		Service:      "waf-gateway",
		TenantID:     event.TenantID,
		ClientIP:     event.ClientIP,
		UserAgent:    event.UserAgent,
		Method:       event.Method,
		Path:         event.Path,
		ThreatType:   event.ThreatType,
		ThreatScore:  event.ThreatScore,
		AttackVector: event.AttackVector,
		Blocked:      event.Action == "block",
		BlockReason:  event.BlockReason,
		Fields:       event.AdditionalData,
		Tags:         []string{"waf", "security-event", event.ThreatType},
	}
	
	e.LogRequest(entry)
}

// SecurityEvent 보안 이벤트 구조
type SecurityEvent struct {
	TenantID       string                 `json:"tenant_id"`
	ClientIP       string                 `json:"client_ip"`
	UserAgent      string                 `json:"user_agent"`
	Method         string                 `json:"method"`
	Path           string                 `json:"path"`
	ThreatType     string                 `json:"threat_type"`     // SQL_INJECTION, XSS, etc.
	ThreatScore    int                    `json:"threat_score"`    // 1-10
	AttackVector   string                 `json:"attack_vector"`   // payload
	Severity       string                 `json:"severity"`        // low, medium, high, critical
	Action         string                 `json:"action"`          // block, log, allow
	BlockReason    string                 `json:"block_reason"`
	Description    string                 `json:"description"`
	AdditionalData map[string]interface{} `json:"additional_data"`
}

// startBatchProcessor 배치 처리 시작
func (e *ELKCollector) startBatchProcessor() {
	ticker := time.NewTicker(e.flushInterval)
	defer ticker.Stop()
	
	for {
		select {
		case entry := <-e.bufferChan:
			e.logBuffer = append(e.logBuffer, entry)
			
			// 배치 크기에 도달하면 즉시 전송
			if len(e.logBuffer) >= e.batchSize {
				e.flushBatch()
			}
			
		case <-ticker.C:
			// 주기적으로 배치 전송
			if len(e.logBuffer) > 0 {
				e.flushBatch()
			}
		}
	}
}

// flushBatch 배치 전송
func (e *ELKCollector) flushBatch() {
	if len(e.logBuffer) == 0 {
		return
	}
	
	// Elasticsearch Bulk API 형식으로 변환
	var bulkData bytes.Buffer
	
	for _, entry := range e.logBuffer {
		// 인덱스 이름 생성 (날짜 기반)
		indexName := e.generateIndexName(entry.Timestamp)
		
		// Bulk API 헤더
		indexAction := map[string]interface{}{
			"index": map[string]string{
				"_index": indexName,
			},
		}
		
		headerBytes, _ := json.Marshal(indexAction)
		bulkData.Write(headerBytes)
		bulkData.WriteByte('\n')
		
		// 실제 데이터
		entryBytes, err := json.Marshal(entry)
		if err != nil {
			e.logger.WithError(err).Error("로그 엔트리 JSON 변환 실패")
			continue
		}
		
		bulkData.Write(entryBytes)
		bulkData.WriteByte('\n')
	}
	
	// Elasticsearch로 전송
	err := e.sendToElasticsearch(bulkData.Bytes())
	if err != nil {
		e.logger.WithError(err).WithField("batch_size", len(e.logBuffer)).Error("Elasticsearch 전송 실패")
		
		// 재전송 로직 (간단한 예제)
		time.Sleep(time.Second * 2)
		if retryErr := e.sendToElasticsearch(bulkData.Bytes()); retryErr != nil {
			e.logger.WithError(retryErr).Error("Elasticsearch 재전송 실패")
		}
	} else {
		e.logger.WithField("batch_size", len(e.logBuffer)).Debug("로그 배치 전송 완료")
	}
	
	// 버퍼 클리어
	e.logBuffer = e.logBuffer[:0]
}

// sendToElasticsearch Elasticsearch로 데이터 전송
func (e *ELKCollector) sendToElasticsearch(data []byte) error {
	url := fmt.Sprintf("%s/_bulk", e.elasticsearchURL)
	
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/x-ndjson")
	
	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("Elasticsearch 응답 에러: %d", resp.StatusCode)
	}
	
	return nil
}

// generateIndexName 날짜 기반 인덱스 이름 생성
func (e *ELKCollector) generateIndexName(timestamp time.Time) string {
	// waf-logs-2024.01.15 형식
	dateStr := timestamp.Format("2006.01.02")
	return strings.Replace(e.indexPattern, "*", dateStr, 1)
}

// Close 컬렉터 종료
func (e *ELKCollector) Close() error {
	// 남은 로그 전송
	if len(e.logBuffer) > 0 {
		e.flushBatch()
	}
	
	close(e.bufferChan)
	return nil
}

// CreateIndexTemplate Elasticsearch 인덱스 템플릿 생성
func (e *ELKCollector) CreateIndexTemplate() error {
	template := map[string]interface{}{
		"index_patterns": []string{e.indexPattern},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
			"index": map[string]interface{}{
				"lifecycle": map[string]interface{}{
					"name": "waf-logs-policy",
				},
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"@timestamp": map[string]string{
					"type": "date",
				},
				"client_ip": map[string]interface{}{
					"type": "ip",
				},
				"geoip": map[string]interface{}{
					"properties": map[string]interface{}{
						"location": map[string]string{
							"type": "geo_point",
						},
						"country": map[string]string{
							"type": "keyword",
						},
						"city": map[string]string{
							"type": "keyword",
						},
					},
				},
				"threat_type": map[string]string{
					"type": "keyword",
				},
				"threat_score": map[string]string{
					"type": "integer",
				},
				"blocked": map[string]string{
					"type": "boolean",
				},
				"rate_limited": map[string]string{
					"type": "boolean",
				},
				"response_time_ms": map[string]string{
					"type": "float",
				},
				"status_code": map[string]string{
					"type": "integer",
				},
				"tags": map[string]string{
					"type": "keyword",
				},
				"method": map[string]string{
					"type": "keyword",
				},
				"path": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]string{
							"type": "keyword",
						},
					},
				},
				"user_agent": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]string{
							"type": "keyword",
						},
					},
				},
			},
		},
	}
	
	templateBytes, err := json.Marshal(template)
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("%s/_index_template/waf-logs-template", e.elasticsearchURL)
	
	req, err := http.NewRequest("PUT", url, bytes.NewReader(templateBytes))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("인덱스 템플릿 생성 실패: %d", resp.StatusCode)
	}
	
	e.logger.Info("Elasticsearch 인덱스 템플릿 생성 완료")
	return nil
}

// SearchLogs 로그 검색 (간단한 예제)
func (e *ELKCollector) SearchLogs(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	searchBody := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{},
				"filter": []interface{}{
					map[string]interface{}{
						"range": map[string]interface{}{
							"@timestamp": map[string]interface{}{
								"gte": query.StartTime.Format(time.RFC3339),
								"lte": query.EndTime.Format(time.RFC3339),
							},
						},
					},
				},
			},
		},
		"size": query.Size,
		"sort": []interface{}{
			map[string]interface{}{
				"@timestamp": map[string]string{
					"order": "desc",
				},
			},
		},
	}
	
	// 필터 조건 추가
	filters := searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]interface{})
	
	if query.TenantID != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]string{
				"tenant_id": query.TenantID,
			},
		})
	}
	
	if query.ClientIP != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]string{
				"client_ip": query.ClientIP,
			},
		})
	}
	
	if query.ThreatType != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]string{
				"threat_type": query.ThreatType,
			},
		})
	}
	
	if query.BlockedOnly {
		filters = append(filters, map[string]interface{}{
			"term": map[string]bool{
				"blocked": true,
			},
		})
	}
	
	searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = filters
	
	searchBytes, err := json.Marshal(searchBody)
	if err != nil {
		return nil, err
	}
	
	url := fmt.Sprintf("%s/%s/_search", e.elasticsearchURL, e.indexPattern)
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(searchBytes))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

// SearchQuery 검색 쿼리 구조
type SearchQuery struct {
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	TenantID    string    `json:"tenant_id,omitempty"`
	ClientIP    string    `json:"client_ip,omitempty"`
	ThreatType  string    `json:"threat_type,omitempty"`
	BlockedOnly bool      `json:"blocked_only"`
	Size        int       `json:"size"`
}

// SearchResult 검색 결과 구조
type SearchResult struct {
	Took int `json:"took"`
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source LogEntry `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}