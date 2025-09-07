package k8s

import (
	"context"
	"fmt"
	"time"
	"waf-backend/models"
	"waf-backend/utils"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type RuleK8sService interface {
	UpdateConfigMapAndIngress(rules []*models.CustomRule) error
	SyncConfigMap() error
	InitializeK8sClient() error
}

type ruleK8sService struct {
	log           *logrus.Logger
	k8sClient     kubernetes.Interface
	configMapName string
	namespace     string
}

func NewRuleK8sService(log *logrus.Logger) RuleK8sService {
	service := &ruleK8sService{
		log:           log,
		configMapName: utils.GetEnv("MODSECURITY_CONFIGMAP", "modsecurity-config"),
		namespace:     utils.GetEnv("KUBERNETES_NAMESPACE", "ingress-nginx"),
	}
	
	// Kubernetes 클라이언트 초기화 시도
	if err := service.InitializeK8sClient(); err != nil {
		log.WithError(err).Warn("Failed to initialize Kubernetes client, K8s features will be disabled")
	} else {
		log.Info("Kubernetes client initialized successfully")
	}
	
	return service
}

func (s *ruleK8sService) InitializeK8sClient() error {
	// 클러스터 내부에서 실행되는 경우 ServiceAccount 사용
	config, err := rest.InClusterConfig()
	if err != nil {
		s.log.WithError(err).Debug("Not running in cluster, this is normal for development")
		return err
	}
	
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}
	
	s.k8sClient = clientset
	return nil
}

func (s *ruleK8sService) UpdateConfigMapAndIngress(rules []*models.CustomRule) error {
	if s.k8sClient == nil {
		s.log.Debug("No Kubernetes client available, skipping K8s updates")
		return nil
	}

	// ConfigMap 업데이트
	if err := s.updateConfigMap(rules); err != nil {
		return fmt.Errorf("failed to update ConfigMap: %w", err)
	}

	// Ingress annotation 업데이트
	if err := s.updateIngressAnnotation(rules); err != nil {
		return fmt.Errorf("failed to update Ingress annotation: %w", err)
	}

	return nil
}

func (s *ruleK8sService) updateConfigMap(rules []*models.CustomRule) error {
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
	for _, rule := range rules {
		if rule.Enabled {
			customRulesContent += fmt.Sprintf("# %s\n# %s\n%s\n\n", rule.Name, rule.Description, rule.RuleText)
		}
	}
	
	configMap.Data["custom-rules.conf"] = customRulesContent
	
	s.log.WithField("rules_count", len(rules)).Info("Updating ConfigMap with custom rules")
	
	// ConfigMap 업데이트
	_, err = configMapClient.Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ConfigMap: %w", err)
	}
	
	s.log.Info("ConfigMap updated successfully")
	return nil
}

func (s *ruleK8sService) updateIngressAnnotation(rules []*models.CustomRule) error {
	ctx := context.Background()
	// Ingress는 default namespace에 있음 (ConfigMap과 다름)
	ingressClient := s.k8sClient.NetworkingV1().Ingresses("default")
	
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
	for _, rule := range rules {
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
	
	// NGINX 재시작
	if err := s.restartNginxIngressController(); err != nil {
		s.log.WithError(err).Warn("Failed to restart NGINX Ingress Controller")
	}
	
	return nil
}

func (s *ruleK8sService) restartNginxIngressController() error {
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

func (s *ruleK8sService) SyncConfigMap() error {
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

	// ConfigMap에서 custom-rules.conf 확인
	customRulesContent, exists := configMap.Data["custom-rules.conf"]
	if !exists || customRulesContent == "" {
		s.log.Info("ConfigMap is empty or has no custom rules")
		return nil
	}

	s.log.WithField("content_length", len(customRulesContent)).Debug("ConfigMap sync completed")
	return nil
}