package k8s

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"waf-backend/models"
	"waf-backend/utils"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ProxyK8sService interface {
	CreateProxyIngress(proxy *models.ProxyTarget) error
	UpdateProxyIngress(proxy *models.ProxyTarget) error
	DeleteProxyIngress(proxy *models.ProxyTarget) error
	GetProxyIngress(proxy *models.ProxyTarget) (*networkingv1.Ingress, error)
	createExternalService(proxy *models.ProxyTarget) error
	deleteExternalService(proxy *models.ProxyTarget) error
}

type proxyK8sService struct {
	log       *logrus.Logger
	clientset *kubernetes.Clientset
	namespace string
}

func NewProxyK8sService(log *logrus.Logger) ProxyK8sService {
	service := &proxyK8sService{
		log:       log,
		namespace: utils.GetEnv("K8S_NAMESPACE", "default"),
	}

	// K8s 클라이언트 초기화
	if err := service.initKubernetesClient(); err != nil {
		log.WithError(err).Error("Failed to initialize Kubernetes client for proxy service")
	}

	return service
}

func (s *proxyK8sService) initKubernetesClient() error {
	var config *rest.Config
	var err error

	// Try in-cluster config first (for running inside a pod)
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig file (for local development)
		kubeconfig := utils.GetEnv("KUBECONFIG", "C:\\workspace\\kubeconfig")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return fmt.Errorf("failed to build kubeconfig: %w", err)
		}
		s.log.WithField("kubeconfig", kubeconfig).Info("Using kubeconfig file for proxy service")
	} else {
		s.log.Info("Using in-cluster config for proxy service")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	s.clientset = clientset
	s.log.Info("Kubernetes client initialized for proxy service")
	return nil
}

func (s *proxyK8sService) CreateProxyIngress(proxy *models.ProxyTarget) error {
	if s.clientset == nil {
		return fmt.Errorf("kubernetes client not initialized")
	}

	// Create external service for non-internal targets
	if !s.isInternalService(proxy.OriginURL) {
		if err := s.createExternalService(proxy); err != nil {
			s.log.WithError(err).Error("Failed to create external service")
			return fmt.Errorf("failed to create external service: %w", err)
		}
	}

	ingress := s.buildIngressFromProxy(proxy)

	_, err := s.clientset.NetworkingV1().Ingresses(s.namespace).Create(
		context.TODO(),
		ingress,
		metav1.CreateOptions{},
	)

	if err != nil {
		// Clean up external service if ingress creation fails
		if !s.isInternalService(proxy.OriginURL) {
			s.deleteExternalService(proxy)
		}
		s.log.WithError(err).WithFields(logrus.Fields{
			"proxy_id":     proxy.ID,
			"proxy_domain": proxy.ProxyDomain,
			"origin_url":   proxy.OriginURL,
		}).Error("Failed to create proxy ingress")
		return fmt.Errorf("failed to create ingress: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"proxy_id":     proxy.ID,
		"proxy_domain": proxy.ProxyDomain,
		"origin_url":   proxy.OriginURL,
		"ingress_name": s.getIngressName(proxy),
	}).Info("Proxy ingress created successfully")

	return nil
}

func (s *proxyK8sService) UpdateProxyIngress(proxy *models.ProxyTarget) error {
	if s.clientset == nil {
		return fmt.Errorf("kubernetes client not initialized")
	}

	// 기존 Ingress 삭제 후 새로 생성
	if err := s.DeleteProxyIngress(proxy); err != nil {
		s.log.WithError(err).Warn("Failed to delete existing ingress during update")
	}

	return s.CreateProxyIngress(proxy)
}

func (s *proxyK8sService) DeleteProxyIngress(proxy *models.ProxyTarget) error {
	if s.clientset == nil {
		return fmt.Errorf("kubernetes client not initialized")
	}

	ingressName := s.getIngressName(proxy)

	// Delete ingress
	err := s.clientset.NetworkingV1().Ingresses(s.namespace).Delete(
		context.TODO(),
		ingressName,
		metav1.DeleteOptions{},
	)

	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"proxy_id":     proxy.ID,
			"ingress_name": ingressName,
		}).Error("Failed to delete proxy ingress")
		return fmt.Errorf("failed to delete ingress: %w", err)
	}

	// Delete external service if it exists
	if !s.isInternalService(proxy.OriginURL) {
		if err := s.deleteExternalService(proxy); err != nil {
			s.log.WithError(err).Warn("Failed to delete external service, but ingress was deleted")
		}
	}

	s.log.WithFields(logrus.Fields{
		"proxy_id":     proxy.ID,
		"ingress_name": ingressName,
	}).Info("Proxy ingress deleted successfully")

	return nil
}

func (s *proxyK8sService) GetProxyIngress(proxy *models.ProxyTarget) (*networkingv1.Ingress, error) {
	if s.clientset == nil {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	ingressName := s.getIngressName(proxy)

	ingress, err := s.clientset.NetworkingV1().Ingresses(s.namespace).Get(
		context.TODO(),
		ingressName,
		metav1.GetOptions{},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get ingress: %w", err)
	}

	return ingress, nil
}

func (s *proxyK8sService) getIngressName(proxy *models.ProxyTarget) string {
	// 프록시 도메인을 기반으로 고유한 Ingress 이름 생성
	domain := proxy.ProxyDomain

	// 포트 번호 제거 (예: example.com:8080 -> example.com)
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	safeName := strings.ReplaceAll(domain, ".", "-")
	safeName = strings.ReplaceAll(safeName, "_", "-")
	return fmt.Sprintf("proxy-%s", safeName)
}

func (s *proxyK8sService) buildIngressFromProxy(proxy *models.ProxyTarget) *networkingv1.Ingress {
	ingressName := s.getIngressName(proxy)

	// ModSecurity 설정
	annotations := map[string]string{
		"kubernetes.io/ingress.class":                        "nginx",
		"nginx.ingress.kubernetes.io/enable-modsecurity":     "true",
		"nginx.ingress.kubernetes.io/enable-owasp-modsecurity-crs": "true",
		"nginx.ingress.kubernetes.io/modsecurity-transaction-id": "$request_id",
		"nginx.ingress.kubernetes.io/modsecurity-snippet": `
			SecRuleEngine On
			SecAuditEngine On
			SecAuditLogParts ABIJDEFHZ
			SecAuditLogType Serial
			SecAuditLog /var/log/nginx/modsec_audit.log

			# Enhanced logging
			SecAuditLogRelevantStatus ".*"

			# CRS configuration
			SecAction "id:900200,phase:1,nolog,pass,setvar:tx.inbound_anomaly_score_threshold=5"
			SecAction "id:900201,phase:1,nolog,pass,setvar:tx.outbound_anomaly_score_threshold=4"

			# Include custom rules
			Include /etc/nginx/modsecurity/custom/custom-rules.conf
		`,
		"nginx.ingress.kubernetes.io/configuration-snippet": `
			more_set_headers "X-WAF-Protected: true";
			more_set_headers "X-ModSecurity-Transaction-Id: $request_id";
			more_set_headers "X-Proxy-Target: ` + proxy.OriginURL + `";
		`,
		"nginx.ingress.kubernetes.io/upstream-vhost":        s.extractHost(proxy.OriginURL),
		"nginx.ingress.kubernetes.io/proxy-ssl-verify":      "false",
		"nginx.ingress.kubernetes.io/force-ssl-redirect":    "false",
		"nginx.ingress.kubernetes.io/proxy-body-size":       "10m",
		"nginx.ingress.kubernetes.io/proxy-connect-timeout": "60",
		"nginx.ingress.kubernetes.io/proxy-send-timeout":    "60",
		"nginx.ingress.kubernetes.io/proxy-read-timeout":    "60",
	}

	pathType := networkingv1.PathTypePrefix

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        ingressName,
			Namespace:   s.namespace,
			Annotations: annotations,
			Labels: map[string]string{
				"app":        "waf-proxy",
				"proxy-id":   proxy.ID,
				"managed-by": "waf-backend",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: s.extractHostWithoutPort(proxy.ProxyDomain),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend:  s.buildBackendFromOriginURL(proxy.OriginURL),
								},
							},
						},
					},
				},
			},
		},
	}

	return ingress
}

func (s *proxyK8sService) extractHost(originURL string) string {
	// URL에서 호스트 추출 (http:// 또는 https:// 제거)
	url := strings.TrimPrefix(originURL, "http://")
	url = strings.TrimPrefix(url, "https://")

	// 포트나 경로가 있으면 제거
	if idx := strings.Index(url, ":"); idx != -1 {
		url = url[:idx]
	}
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}

	return url
}

func (s *proxyK8sService) extractHostWithoutPort(domain string) string {
	// 도메인에서 포트 번호 제거
	if idx := strings.Index(domain, ":"); idx != -1 {
		return domain[:idx]
	}
	return domain
}

func (s *proxyK8sService) buildBackendFromOriginURL(originURL string) networkingv1.IngressBackend {
	// DVWA 서비스인 경우 내부 서비스로 처리
	if s.isInternalService(originURL) {
		return networkingv1.IngressBackend{
			Service: &networkingv1.IngressServiceBackend{
				Name: "dvwa-service",
				Port: networkingv1.ServiceBackendPort{
					Number: 80,
				},
			},
		}
	}

	// 외부 URL의 경우 ExternalName service 사용
	serviceName := s.getExternalServiceName(originURL)
	port := s.getPortFromURL(originURL)

	return networkingv1.IngressBackend{
		Service: &networkingv1.IngressServiceBackend{
			Name: serviceName,
			Port: networkingv1.ServiceBackendPort{
				Number: int32(port),
			},
		},
	}
}

func (s *proxyK8sService) isInternalService(originURL string) bool {
	// DVWA 서비스나 클러스터 내부 서비스 확인
	return strings.Contains(originURL, ":30081") ||
		   strings.Contains(originURL, "dvwa") ||
		   strings.Contains(originURL, "localhost") ||
		   strings.Contains(originURL, "127.0.0.1")
}

func (s *proxyK8sService) getExternalServiceName(originURL string) string {
	parsedURL, err := url.Parse(originURL)
	if err != nil {
		return "external-service"
	}

	// 호스트명을 기반으로 서비스명 생성
	hostname := strings.ReplaceAll(parsedURL.Hostname(), ".", "-")
	hostname = strings.ReplaceAll(hostname, "_", "-")
	return fmt.Sprintf("external-%s", hostname)
}

func (s *proxyK8sService) getPortFromURL(originURL string) int {
	parsedURL, err := url.Parse(originURL)
	if err != nil {
		return 80
	}

	port := parsedURL.Port()
	if port == "" {
		if parsedURL.Scheme == "https" {
			return 443
		}
		return 80
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return 80
	}
	return portNum
}

func (s *proxyK8sService) createExternalService(proxy *models.ProxyTarget) error {
	parsedURL, err := url.Parse(proxy.OriginURL)
	if err != nil {
		return fmt.Errorf("invalid origin URL: %w", err)
	}

	serviceName := s.getExternalServiceName(proxy.OriginURL)
	port := s.getPortFromURL(proxy.OriginURL)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: s.namespace,
			Labels: map[string]string{
				"app":        "waf-proxy",
				"proxy-id":   proxy.ID,
				"managed-by": "waf-backend",
				"type":       "external",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:         corev1.ServiceTypeExternalName,
			ExternalName: parsedURL.Hostname(),
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       int32(port),
					TargetPort: intstr.FromInt(port),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	_, err = s.clientset.CoreV1().Services(s.namespace).Create(
		context.TODO(),
		service,
		metav1.CreateOptions{},
	)

	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"service_name": serviceName,
			"external_name": parsedURL.Hostname(),
			"port": port,
		}).Error("Failed to create external service")
		return fmt.Errorf("failed to create external service: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"service_name": serviceName,
		"external_name": parsedURL.Hostname(),
		"port": port,
	}).Info("External service created successfully")

	return nil
}

func (s *proxyK8sService) deleteExternalService(proxy *models.ProxyTarget) error {
	serviceName := s.getExternalServiceName(proxy.OriginURL)

	err := s.clientset.CoreV1().Services(s.namespace).Delete(
		context.TODO(),
		serviceName,
		metav1.DeleteOptions{},
	)

	if err != nil {
		s.log.WithError(err).WithField("service_name", serviceName).Error("Failed to delete external service")
		return fmt.Errorf("failed to delete external service: %w", err)
	}

	s.log.WithField("service_name", serviceName).Info("External service deleted successfully")
	return nil
}