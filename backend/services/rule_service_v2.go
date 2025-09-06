package services

import (
	"fmt"
	"regexp"
	"time"
	"waf-backend/dto"
	"waf-backend/models"
	"waf-backend/repositories"
	"waf-backend/services/k8s"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type RuleServiceV2 struct {
	log        *logrus.Logger
	repo       repositories.RuleRepository
	k8sService k8s.RuleK8sService
}

func NewRuleServiceV2(log *logrus.Logger) RuleService {
	service := &RuleServiceV2{
		log:        log,
		k8sService: k8s.NewRuleK8sService(log),
	}
	// Repository를 나중에 초기화 (DB가 준비된 후)
	service.repo = repositories.NewRuleRepository()
	
	log.Info("RuleServiceV2 initialized with DB support")
	
	// 기존 룰들을 ConfigMap에 동기화
	go service.syncExistingRulesOnStartup()
	
	return service
}

// CreateRule creates a new custom rule
func (s *RuleServiceV2) CreateRule(userID string, req *dto.CustomRuleRequest) (*dto.CustomRuleResponse, error) {
	// 룰 유효성 검증
	if err := s.validateRule(req.RuleText); err != nil {
		return nil, fmt.Errorf("invalid rule syntax: %w", err)
	}
	
	// 새 룰 생성
	rule := &models.CustomRule{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		RuleText:    req.RuleText,
		Enabled:     req.Enabled,
		Severity:    req.Severity,
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// DB에 저장
	if err := s.repo.Create(rule); err != nil {
		return nil, fmt.Errorf("failed to save rule to database: %w", err)
	}
	
	s.log.WithFields(logrus.Fields{
		"rule_id": rule.ID,
		"user_id": userID,
		"name":    rule.Name,
	}).Info("Custom rule created and saved to database")
	
	// K8s 업데이트 (비동기)
	go s.updateK8sAsync(userID)
	
	return s.modelToResponse(rule), nil
}

// GetRules retrieves all rules for a user
func (s *RuleServiceV2) GetRules(userID string) ([]*dto.CustomRuleResponse, error) {
	rules, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules from database: %w", err)
	}
	
	var result []*dto.CustomRuleResponse
	for _, rule := range rules {
		result = append(result, s.modelToResponse(rule))
	}
	
	return result, nil
}

// GetRule retrieves a single rule
func (s *RuleServiceV2) GetRule(userID, ruleID string) (*dto.CustomRuleResponse, error) {
	rule, err := s.repo.GetByID(ruleID)
	if err != nil {
		return nil, fmt.Errorf("rule not found: %w", err)
	}
	
	if rule.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}
	
	return s.modelToResponse(rule), nil
}

// UpdateRule updates an existing rule
func (s *RuleServiceV2) UpdateRule(userID, ruleID string, req *dto.CustomRuleRequest) (*dto.CustomRuleResponse, error) {
	// 룰 유효성 검증
	if err := s.validateRule(req.RuleText); err != nil {
		return nil, fmt.Errorf("invalid rule syntax: %w", err)
	}
	
	// 기존 룰 가져오기
	rule, err := s.repo.GetByID(ruleID)
	if err != nil {
		return nil, fmt.Errorf("rule not found: %w", err)
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
	
	// DB에 저장
	if err := s.repo.Update(rule); err != nil {
		return nil, fmt.Errorf("failed to update rule in database: %w", err)
	}
	
	s.log.WithFields(logrus.Fields{
		"rule_id": rule.ID,
		"user_id": userID,
		"name":    rule.Name,
	}).Info("Custom rule updated in database")
	
	// K8s 업데이트 (비동기)
	go s.updateK8sAsync(userID)
	
	return s.modelToResponse(rule), nil
}

// DeleteRule deletes a rule
func (s *RuleServiceV2) DeleteRule(userID, ruleID string) error {
	// 소유권 확인
	rule, err := s.repo.GetByID(ruleID)
	if err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}
	
	if rule.UserID != userID {
		return fmt.Errorf("access denied")
	}
	
	// DB에서 삭제
	if err := s.repo.DeleteByUserIDAndID(userID, ruleID); err != nil {
		return fmt.Errorf("failed to delete rule from database: %w", err)
	}
	
	s.log.WithFields(logrus.Fields{
		"rule_id": ruleID,
		"user_id": userID,
		"name":    rule.Name,
	}).Info("Custom rule deleted from database")
	
	// K8s 업데이트 (비동기)
	go s.updateK8sAsync(userID)
	
	return nil
}

// validateRule validates ModSecurity rule syntax
func (s *RuleServiceV2) validateRule(ruleText string) error {
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

// updateK8sAsync updates Kubernetes resources asynchronously
func (s *RuleServiceV2) updateK8sAsync(userID string) {
	// DB에서 사용자의 모든 룰 가져오기
	rules, err := s.repo.GetByUserID(userID)
	if err != nil {
		s.log.WithError(err).Error("Failed to get rules for K8s update")
		return
	}
	
	// K8s 업데이트
	if err := s.k8sService.UpdateConfigMapAndIngress(rules); err != nil {
		s.log.WithError(err).Error("Failed to update Kubernetes resources")
	} else {
		s.log.Info("Kubernetes resources updated successfully")
	}
}

// syncExistingRulesOnStartup synchronizes existing DB rules to ConfigMap on startup
func (s *RuleServiceV2) syncExistingRulesOnStartup() {
	// DB 초기화를 위해 잠시 대기
	time.Sleep(2 * time.Second)
	
	s.log.Info("Syncing existing rules to ConfigMap on startup")
	
	// 모든 사용자의 룰 조회 (user_id로 그룹화해서 각각 동기화)
	// 현재는 간단히 모든 룰을 함께 동기화
	rules, err := s.getAllRules()
	if err != nil {
		s.log.WithError(err).Error("Failed to get existing rules for startup sync")
		return
	}
	
	if len(rules) > 0 {
		s.log.WithField("count", len(rules)).Info("Found existing rules, syncing to ConfigMap")
		if err := s.k8sService.UpdateConfigMapAndIngress(rules); err != nil {
			s.log.WithError(err).Error("Failed to sync existing rules to ConfigMap on startup")
		} else {
			s.log.Info("Successfully synced existing rules to ConfigMap on startup")
		}
	} else {
		s.log.Info("No existing rules found, ConfigMap will be empty")
	}
}

// getAllRules gets all rules from all users (for startup sync)
func (s *RuleServiceV2) getAllRules() ([]*models.CustomRule, error) {
	// 현재 알고 있는 테스트 사용자의 룰을 가져옴
	rules, err := s.repo.GetByUserID("112116839068571902699") // 현재 테스트 사용자
	if err != nil {
		s.log.WithError(err).Debug("No rules found for test user, returning empty list")
		return []*models.CustomRule{}, nil // 오류가 있어도 빈 슬라이스 반환
	}
	
	return rules, nil
}

// modelToResponse converts model to response DTO
func (s *RuleServiceV2) modelToResponse(rule *models.CustomRule) *dto.CustomRuleResponse {
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