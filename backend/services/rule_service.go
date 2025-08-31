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
	
	// ConfigMap과 Ingress annotation 업데이트 (즉시 적용)
	if err := s.updateConfigMap(); err != nil {
		s.log.WithError(err).Error("Failed to update ConfigMap")
	}
	if err := s.updateIngressAnnotation(); err != nil {
		s.log.WithError(err).Error("Failed to update Ingress annotation")
	}
	
	s.log.WithFields(logrus.Fields{
		"rule_id": rule.ID,
		"user_id": userID,
		"name":    rule.Name,
	}).Info("Custom rule created")
	
	return s.ruleToResponse(rule), nil
}

func (s *RuleService) GetRules(userID string) ([]*dto.CustomRuleResponse, error) {
	// ConfigMap에서 최신 상태 동기화
	if err := s.syncFromConfigMap(); err != nil {
		s.log.WithError(err).Warn("Failed to sync from ConfigMap, using cached data")
	}
	
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
	
	// ConfigMap과 Ingress annotation 업데이트 (즉시 적용)
	if err := s.updateConfigMap(); err != nil {
		s.log.WithError(err).Error("Failed to update ConfigMap")
	}
	if err := s.updateIngressAnnotation(); err != nil {
		s.log.WithError(err).Error("Failed to update Ingress annotation")
	}
	
	s.log.WithFields(logrus.Fields{
		"rule_id": rule.ID,
		"user_id": userID,
		"name":    rule.Name,
	}).Info("Custom rule updated")
	
	return s.ruleToResponse(rule), nil
}

func (s *RuleService) DeleteRule(userID, ruleID string) error {
	// ConfigMap에서 최신 상태 동기화
	if err := s.syncFromConfigMap(); err != nil {
		s.log.WithError(err).Warn("Failed to sync from ConfigMap before delete")
	}
	
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
	
	// ConfigMap과 Ingress annotation 업데이트 (즉시 적용)
	if err := s.updateConfigMap(); err != nil {
		s.log.WithError(err).Error("Failed to update ConfigMap")
	}
	if err := s.updateIngressAnnotation(); err != nil {
		s.log.WithError(err).Error("Failed to update Ingress annotation")
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



func (s *RuleService) updateIngressAnnotation() error {
	if s.k8sClient == nil {
		s.log.Debug("No Kubernetes client available, skipping Ingress annotation update")
		return nil
	}

	ctx := context.Background()
	ingressClient := s.k8sClient.NetworkingV1().Ingresses(s.namespace)
	
	// Ingress 가져오기
	ingress, err := ingressClient.Get(ctx, "waf-ingress", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Ingress: %w", err)
	}
	
	if ingress.Annotations == nil {
		ingress.Annotations = make(map[string]string)
	}
	
	// 기본 ModSecurity 설정
	baseConfig := `SecRuleEngine On
SecAuditEngine On  
SecAuditLogParts ABIJDEFHZ
SecAuditLogType Serial
SecAuditLog /var/log/nginx/modsec_audit.log

# Allow OAuth callback
SecRule REQUEST_URI "^/auth/callback" "id:9999,phase:1,pass,nolog,ctl:ruleEngine=Off"

# Allow API requests (including DELETE)
SecRule REQUEST_URI "^/api/" "id:9998,phase:1,pass,nolog,ctl:ruleEngine=Off"`
	
	// 활성화된 커스텀 룰들 추가
	var customRulesSnippet string
	for _, rule := range s.rules {
		if rule.Enabled {
			customRulesSnippet += fmt.Sprintf("\n\n# %s\n# %s\n%s", rule.Name, rule.Description, rule.RuleText)
		}
	}
	
	// ModSecurity snippet 업데이트
	fullConfig := baseConfig + customRulesSnippet
	ingress.Annotations["nginx.ingress.kubernetes.io/modsecurity-snippet"] = fullConfig
	
	s.log.WithField("config_length", len(fullConfig)).Info("Updating Ingress ModSecurity annotation")
	
	// Ingress 업데이트
	_, err = ingressClient.Update(ctx, ingress, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update Ingress: %w", err)
	}
	
	s.log.Info("Ingress ModSecurity annotation updated successfully")
	
	// Force NGINX Ingress Controller reload by restarting the pod
	// This is required because ModSecurity rules don't always apply immediately
	s.log.Info("Forcing NGINX Ingress Controller reload...")
	if err := s.forceNginxReload(); err != nil {
		s.log.WithError(err).Warn("Failed to force NGINX reload, rules may not apply immediately")
	} else {
		s.log.Info("NGINX Ingress Controller reload initiated successfully")
	}
	
	return nil
}

func (s *RuleService) forceNginxReload() error {
	if s.k8sClient == nil {
		return fmt.Errorf("Kubernetes client not available")
	}
	
	ctx := context.Background()
	podsClient := s.k8sClient.CoreV1().Pods("ingress-nginx")
	
	// Delete NGINX Ingress Controller pods to force reload
	labelSelector := "app.kubernetes.io/name=ingress-nginx,app.kubernetes.io/component=controller"
	pods, err := podsClient.List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("failed to list NGINX pods: %w", err)
	}
	
	for _, pod := range pods.Items {
		s.log.WithField("pod", pod.Name).Info("Restarting NGINX Ingress Controller pod")
		err = podsClient.Delete(ctx, pod.Name, metav1.DeleteOptions{})
		if err != nil {
			s.log.WithError(err).Warn("Failed to delete NGINX pod, continuing...")
		}
	}
	
	return nil
}

func (s *RuleService) loadExistingRules() {
	s.log.Info("Loading existing custom rules")
	
	// ConfigMap에서 기존 룰들을 로드
	if err := s.syncFromConfigMap(); err != nil {
		s.log.WithError(err).Warn("Failed to load rules from ConfigMap, starting with empty rules")
	}
	
	s.log.WithField("count", len(s.rules)).Info("Loaded existing rules")
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

func (s *RuleService) updateConfigMap() error {
	if s.k8sClient == nil {
		s.log.Debug("No Kubernetes client available, skipping ConfigMap update")
		return nil
	}

	ctx := context.Background()
	configMapClient := s.k8sClient.CoreV1().ConfigMaps(s.namespace)
	
	// ConfigMap 가져오기
	configMap, err := configMapClient.Get(ctx, s.configMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ConfigMap: %w", err)
	}
	
	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}
	
	// 활성화된 룰들을 custom-rules.conf에 추가
	var customRulesContent string
	for _, rule := range s.rules {
		if rule.Enabled {
			customRulesContent += fmt.Sprintf("# %s\n# %s\n%s\n\n", rule.Name, rule.Description, rule.RuleText)
		}
	}
	
	configMap.Data["custom-rules.conf"] = customRulesContent
	
	s.log.WithField("rules_count", len(s.rules)).Info("Updating ConfigMap with custom rules")
	
	// ConfigMap 업데이트
	_, err = configMapClient.Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ConfigMap: %w", err)
	}
	
	s.log.Info("ConfigMap updated successfully")
	
	// ConfigMap 업데이트 후 NGINX Ingress Controller ConfigMap의 modsecurity-snippet도 업데이트
	s.log.Info("Updating NGINX Ingress Controller ConfigMap...")
	if err := s.updateNginxConfigMapSnippet(); err != nil {
		s.log.WithError(err).Warn("Failed to update NGINX ConfigMap snippet")
		return err
	}
	
	return nil
}

func (s *RuleService) updateNginxConfigMapSnippet() error {
	if s.k8sClient == nil {
		return fmt.Errorf("Kubernetes client not available")
	}

	ctx := context.Background()
	configMapClient := s.k8sClient.CoreV1().ConfigMaps("ingress-nginx")
	
	// NGINX ConfigMap 가져오기
	nginxConfigMap, err := configMapClient.Get(ctx, "ingress-nginx-controller", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get NGINX ConfigMap: %w", err)
	}
	
	if nginxConfigMap.Data == nil {
		nginxConfigMap.Data = make(map[string]string)
	}
	
	// 기본 ModSecurity 설정
	baseConfig := `SecRuleEngine On
SecAuditEngine On
SecAuditLogParts ABIJDEFHZ
SecAuditLogType Serial
SecAuditLog /var/log/nginx/modsec_audit.log
Include /etc/nginx/owasp-modsecurity-crs/nginx-modsecurity.conf

# Custom rules from ConfigMap  
Include /etc/nginx/modsecurity/custom-rules/custom-rules.conf`

	nginxConfigMap.Data["modsecurity-snippet"] = baseConfig
	nginxConfigMap.Data["enable-modsecurity"] = "true"
	nginxConfigMap.Data["enable-owasp-modsecurity-crs"] = "false"
	
	s.log.Info("Updating NGINX ConfigMap modsecurity-snippet")
	
	// NGINX ConfigMap 업데이트
	_, err = configMapClient.Update(ctx, nginxConfigMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update NGINX ConfigMap: %w", err)
	}
	
	s.log.Info("NGINX ConfigMap updated successfully")
	
	// NGINX Ingress Controller 재시작
	s.log.Info("Restarting NGINX Ingress Controller...")
	if err := s.restartNginxIngressController(); err != nil {
		s.log.WithError(err).Warn("Failed to restart NGINX Ingress Controller")
		return err
	}
	
	return nil
}

func (s *RuleService) restartNginxIngressController() error {
	if s.k8sClient == nil {
		return fmt.Errorf("Kubernetes client not available")
	}
	
	ctx := context.Background()
	deploymentClient := s.k8sClient.AppsV1().Deployments("ingress-nginx")
	
	// NGINX Ingress Controller 재시작
	deployment, err := deploymentClient.Get(ctx, "ingress-nginx-controller", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get NGINX deployment: %w", err)
	}
	
	// 재시작을 위해 annotation 추가
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
	
	_, err = deploymentClient.Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to restart NGINX deployment: %w", err)
	}
	
	s.log.Info("NGINX Ingress Controller restart initiated")
	return nil
}

// syncFromConfigMap ConfigMap에서 룰을 읽어와서 메모리 상태 동기화
func (s *RuleService) syncFromConfigMap() error {
	if s.k8sClient == nil {
		s.log.Debug("No Kubernetes client available, skipping ConfigMap sync")
		return nil
	}

	ctx := context.Background()
	configMapClient := s.k8sClient.CoreV1().ConfigMaps(s.namespace)
	
	// ConfigMap 가져오기
	configMap, err := configMapClient.Get(ctx, s.configMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ConfigMap: %w", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// ConfigMap에서 custom-rules.conf 확인
	customRulesContent, exists := configMap.Data["custom-rules.conf"]
	if !exists || customRulesContent == "" {
		// ConfigMap이 비어있으면 메모리도 비움
		if len(s.rules) > 0 {
			s.log.Info("ConfigMap is empty, clearing memory rules")
			s.rules = make(map[string]*dto.CustomRule)
		}
		return nil
	}

	// ConfigMap에 내용이 있지만 메모리가 비어있는 경우
	// (예: 다른 인스턴스에서 룰을 생성했거나, 직접 ConfigMap을 수정한 경우)
	if len(s.rules) == 0 {
		s.log.WithField("content_length", len(customRulesContent)).Warn("ConfigMap has content but memory is empty - this may indicate external changes")
		// 현재로서는 ConfigMap의 ModSecurity 룰을 파싱해서 객체로 복원하는 것은 복잡하므로
		// 경고만 남기고 메모리 상태를 유지
		return nil
	}

	s.log.WithField("content_length", len(customRulesContent)).Debug("ConfigMap sync completed")
	return nil
}