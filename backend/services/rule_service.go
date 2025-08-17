package services

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"
	"waf-backend/dto"
	"waf-backend/utils"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type RuleService struct {
	log          *logrus.Logger
	rules        map[string]*dto.CustomRule
	mutex        sync.RWMutex
	k8sClient    kubernetes.Interface
	configMapName string
	namespace    string
}

func NewRuleService(log *logrus.Logger) *RuleService {
	service := &RuleService{
		log:           log,
		rules:         make(map[string]*dto.CustomRule),
		configMapName: utils.GetEnv("MODSECURITY_CONFIGMAP", "modsecurity-config"),
		namespace:     utils.GetEnv("KUBERNETES_NAMESPACE", "default"),
	}
	
	// Kubernetes 클라이언트 초기화
	if k8sClient, err := service.initK8sClient(); err == nil {
		service.k8sClient = k8sClient
		log.Info("Kubernetes client initialized successfully")
	} else {
		log.WithError(err).Warn("Failed to initialize Kubernetes client, rules will be stored in memory only")
	}
	
	// 기존 룰들을 로드
	service.loadExistingRules()
	
	return service
}

func (s *RuleService) initK8sClient() (kubernetes.Interface, error) {
	// 클러스터 내부에서 실행되는 경우 ServiceAccount 사용
	config, err := rest.InClusterConfig()
	if err != nil {
		s.log.WithError(err).Debug("Not running in cluster, this is normal for development")
		return nil, err
	}
	
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}
	
	return clientset, nil
}

func (s *RuleService) CreateRule(userID string, req *dto.CustomRuleRequest) (*dto.CustomRuleResponse, error) {
	// 룰 유효성 검증
	if err := s.validateRule(req.RuleText); err != nil {
		return nil, fmt.Errorf("invalid rule syntax: %w", err)
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	rule := &dto.CustomRule{
		ID:          generateRuleID(),
		Name:        req.Name,
		Description: req.Description,
		RuleText:    req.RuleText,
		Enabled:     req.Enabled,
		Severity:    req.Severity,
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	s.rules[rule.ID] = rule
	
	// Kubernetes ConfigMap 업데이트
	if err := s.updateConfigMap(); err != nil {
		s.log.WithError(err).Error("Failed to update ConfigMap")
		// ConfigMap 업데이트 실패해도 메모리에는 저장
	}
	
	s.log.WithFields(logrus.Fields{
		"rule_id": rule.ID,
		"user_id": userID,
		"name":    rule.Name,
	}).Info("Custom rule created")
	
	return s.ruleToResponse(rule), nil
}

func (s *RuleService) GetRules(userID string) ([]*dto.CustomRuleResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	var result []*dto.CustomRuleResponse
	
	for _, rule := range s.rules {
		if rule.UserID == userID {
			result = append(result, s.ruleToResponse(rule))
		}
	}
	
	return result, nil
}

func (s *RuleService) GetRule(userID, ruleID string) (*dto.CustomRuleResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	rule, exists := s.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule not found")
	}
	
	if rule.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}
	
	return s.ruleToResponse(rule), nil
}

func (s *RuleService) UpdateRule(userID, ruleID string, req *dto.CustomRuleRequest) (*dto.CustomRuleResponse, error) {
	// 룰 유효성 검증
	if err := s.validateRule(req.RuleText); err != nil {
		return nil, fmt.Errorf("invalid rule syntax: %w", err)
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	rule, exists := s.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule not found")
	}
	
	if rule.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}
	
	// 룰 업데이트
	rule.Name = req.Name
	rule.Description = req.Description
	rule.RuleText = req.RuleText
	rule.Enabled = req.Enabled
	rule.Severity = req.Severity
	rule.UpdatedAt = time.Now()
	
	// Kubernetes ConfigMap 업데이트
	if err := s.updateConfigMap(); err != nil {
		s.log.WithError(err).Error("Failed to update ConfigMap")
	}
	
	s.log.WithFields(logrus.Fields{
		"rule_id": rule.ID,
		"user_id": userID,
		"name":    rule.Name,
	}).Info("Custom rule updated")
	
	return s.ruleToResponse(rule), nil
}

func (s *RuleService) DeleteRule(userID, ruleID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	rule, exists := s.rules[ruleID]
	if !exists {
		return fmt.Errorf("rule not found")
	}
	
	if rule.UserID != userID {
		return fmt.Errorf("access denied")
	}
	
	delete(s.rules, ruleID)
	
	// Kubernetes ConfigMap 업데이트
	if err := s.updateConfigMap(); err != nil {
		s.log.WithError(err).Error("Failed to update ConfigMap")
	}
	
	s.log.WithFields(logrus.Fields{
		"rule_id": rule.ID,
		"user_id": userID,
		"name":    rule.Name,
	}).Info("Custom rule deleted")
	
	return nil
}

func (s *RuleService) validateRule(ruleText string) error {
	// 기본적인 ModSecurity 룰 문법 검증
	if len(ruleText) == 0 {
		return fmt.Errorf("rule text cannot be empty")
	}
	
	// SecRule로 시작하는지 확인
	if !regexp.MustCompile(`^SecRule\s+`).MatchString(ruleText) {
		return fmt.Errorf("rule must start with 'SecRule'")
	}
	
	// 기본적인 문법 구조 검증
	if !regexp.MustCompile(`SecRule\s+\S+\s+"[^"]*"\s+"[^"]*"`).MatchString(ruleText) {
		return fmt.Errorf("invalid ModSecurity rule syntax")
	}
	
	// 위험한 키워드 검증 (보안을 위해)
	dangerousKeywords := []string{"exec", "system", "eval", "cmd"}
	for _, keyword := range dangerousKeywords {
		if regexp.MustCompile(`(?i)`+keyword).MatchString(ruleText) {
			return fmt.Errorf("dangerous keyword '%s' not allowed", keyword)
		}
	}
	
	return nil
}

func (s *RuleService) updateConfigMap() error {
	if s.k8sClient == nil {
		s.log.Debug("No Kubernetes client available, skipping ConfigMap update")
		return nil
	}
	
	ctx := context.Background()
	configMapsClient := s.k8sClient.CoreV1().ConfigMaps(s.namespace)
	
	// 현재 ConfigMap 가져오기
	configMap, err := configMapsClient.Get(ctx, s.configMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ConfigMap: %w", err)
	}
	
	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}
	
	// 커스텀 룰들을 생성
	var customRules string
	for _, rule := range s.rules {
		if rule.Enabled {
			customRules += fmt.Sprintf("# %s\n# %s\n%s\n\n", rule.Name, rule.Description, rule.RuleText)
		}
	}
	
	// ConfigMap 업데이트
	configMap.Data["custom-rules.conf"] = customRules
	
	_, err = configMapsClient.Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ConfigMap: %w", err)
	}
	
	s.log.Info("ConfigMap updated successfully")
	return nil
}

func (s *RuleService) loadExistingRules() {
	s.log.Info("Loading existing custom rules")
	
	// 실제 환경에서는 데이터베이스에서 로드하거나 ConfigMap에서 파싱
	// 여기서는 메모리에 샘플 룰들을 추가
	sampleRules := []*dto.CustomRule{
		{
			ID:          "rule_001",
			Name:        "Block Common SQL Injection Patterns",
			Description: "Blocks common SQL injection attack patterns",
			RuleText:    `SecRule ARGS "@detectSQLi" "id:1001,phase:2,block,msg:'SQL Injection Attack',severity:HIGH"`,
			Enabled:     true,
			Severity:    "HIGH",
			UserID:      "system",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rule_002",
			Name:        "Block XSS Attempts",
			Description: "Blocks cross-site scripting attacks",
			RuleText:    `SecRule ARGS "@detectXSS" "id:1002,phase:2,block,msg:'XSS Attack',severity:HIGH"`,
			Enabled:     true,
			Severity:    "HIGH",
			UserID:      "system",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
	
	for _, rule := range sampleRules {
		s.rules[rule.ID] = rule
	}
	
	s.log.WithField("count", len(sampleRules)).Info("Loaded existing rules")
}

func (s *RuleService) ruleToResponse(rule *dto.CustomRule) *dto.CustomRuleResponse {
	return &dto.CustomRuleResponse{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		RuleText:    rule.RuleText,
		Enabled:     rule.Enabled,
		Severity:    rule.Severity,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}
}

func generateRuleID() string {
	return fmt.Sprintf("rule_%d", time.Now().UnixNano())
}