package tenant

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// TenantManager 멀티테넌트 관리자
type TenantManager struct {
	clientset *kubernetes.Clientset
	logger    *logrus.Logger
}

// Tenant 테넌트 정보
type Tenant struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Domain        string            `json:"domain"`
	Namespace     string            `json:"namespace"`
	APIKey        string            `json:"api_key"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Status        TenantStatus      `json:"status"`
	Configuration TenantConfig      `json:"configuration"`
	Quotas        TenantQuotas      `json:"quotas"`
	Metadata      map[string]string `json:"metadata"`
}

// TenantStatus 테넌트 상태
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusPending   TenantStatus = "pending"
	TenantStatusDeleting  TenantStatus = "deleting"
)

// TenantConfig 테넌트별 WAF 설정
type TenantConfig struct {
	RateLimiting    RateLimitConfig    `json:"rate_limiting"`
	SecurityPolicy  SecurityPolicy     `json:"security_policy"`
	LoggingConfig   LoggingConfig      `json:"logging_config"`
	CustomRules     []CustomRule       `json:"custom_rules"`
	AllowedIPs      []string           `json:"allowed_ips"`
	BlockedIPs      []string           `json:"blocked_ips"`
	GeoBlocking     GeoBlockingConfig  `json:"geo_blocking"`
}

// RateLimitConfig Rate Limiting 설정
type RateLimitConfig struct {
	RequestsPerMinute int    `json:"requests_per_minute"`
	BurstSize         int    `json:"burst_size"`
	WindowSize        string `json:"window_size"`
	Enabled           bool   `json:"enabled"`
}

// SecurityPolicy 보안 정책
type SecurityPolicy struct {
	ParanoiaLevel      int      `json:"paranoia_level"`
	SQLInjectionBlock  bool     `json:"sql_injection_block"`
	XSSBlock           bool     `json:"xss_block"`
	PathTraversalBlock bool     `json:"path_traversal_block"`
	CommandInjBlock    bool     `json:"command_injection_block"`
	AllowedMethods     []string `json:"allowed_methods"`
	MaxRequestSize     int64    `json:"max_request_size"`
	MaxHeaderSize      int      `json:"max_header_size"`
}

// LoggingConfig 로깅 설정
type LoggingConfig struct {
	Enabled       bool   `json:"enabled"`
	Level         string `json:"level"`
	SampleRate    float64 `json:"sample_rate"`
	RetentionDays int    `json:"retention_days"`
}

// CustomRule 커스텀 보안 룰
type CustomRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Pattern     string `json:"pattern"`
	Action      string `json:"action"` // block, log, allow
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// GeoBlockingConfig 지역 차단 설정
type GeoBlockingConfig struct {
	Enabled        bool     `json:"enabled"`
	BlockedCountries []string `json:"blocked_countries"`
	AllowedCountries []string `json:"allowed_countries"`
	Mode           string   `json:"mode"` // blacklist, whitelist
}

// TenantQuotas 테넌트 할당량
type TenantQuotas struct {
	MaxRequestsPerDay   int64 `json:"max_requests_per_day"`
	MaxBandwidthPerDay  int64 `json:"max_bandwidth_per_day"` // bytes
	MaxDomains          int   `json:"max_domains"`
	MaxCustomRules      int   `json:"max_custom_rules"`
	StorageQuotaGB      int   `json:"storage_quota_gb"`
}

// NewTenantManager 테넌트 매니저 생성
func NewTenantManager() (*TenantManager, error) {
	// Kubernetes 클러스터 내부에서 실행되는 경우 InClusterConfig 사용
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("Kubernetes 설정 로드 실패: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Kubernetes 클라이언트 생성 실패: %v", err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &TenantManager{
		clientset: clientset,
		logger:    logger,
	}, nil
}

// CreateTenant 새 테넌트 생성
func (tm *TenantManager) CreateTenant(ctx context.Context, tenantReq CreateTenantRequest) (*Tenant, error) {
	// 테넌트 ID 생성
	tenantID := tm.generateTenantID()
	namespace := fmt.Sprintf("tenant-%s", tenantID)
	apiKey := tm.generateAPIKey()

	tenant := &Tenant{
		ID:        tenantID,
		Name:      tenantReq.Name,
		Domain:    tenantReq.Domain,
		Namespace: namespace,
		APIKey:    apiKey,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    TenantStatusPending,
		Configuration: TenantConfig{
			RateLimiting: RateLimitConfig{
				RequestsPerMinute: tenantReq.RateLimitRPM,
				BurstSize:         tenantReq.BurstSize,
				WindowSize:        "1m",
				Enabled:          true,
			},
			SecurityPolicy: SecurityPolicy{
				ParanoiaLevel:      tenantReq.ParanoiaLevel,
				SQLInjectionBlock:  true,
				XSSBlock:           true,
				PathTraversalBlock: true,
				CommandInjBlock:    true,
				AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
				MaxRequestSize:     10 * 1024 * 1024, // 10MB
				MaxHeaderSize:      8192,              // 8KB
			},
			LoggingConfig: LoggingConfig{
				Enabled:       true,
				Level:         "info",
				SampleRate:    1.0,
				RetentionDays: 30,
			},
			GeoBlocking: GeoBlockingConfig{
				Enabled: false,
				Mode:    "blacklist",
			},
		},
		Quotas: TenantQuotas{
			MaxRequestsPerDay:   tenantReq.DailyRequestLimit,
			MaxBandwidthPerDay:  tenantReq.BandwidthLimitGB * 1024 * 1024 * 1024,
			MaxDomains:          tenantReq.MaxDomains,
			MaxCustomRules:      100,
			StorageQuotaGB:      10,
		},
		Metadata: map[string]string{
			"plan":         tenantReq.Plan,
			"contact":      tenantReq.ContactEmail,
			"created_by":   "waf-system",
		},
	}

	// Kubernetes 리소스 생성
	if err := tm.createTenantResources(ctx, tenant); err != nil {
		return nil, fmt.Errorf("테넌트 리소스 생성 실패: %v", err)
	}

	tenant.Status = TenantStatusActive

	tm.logger.WithFields(logrus.Fields{
		"tenant_id":   tenant.ID,
		"tenant_name": tenant.Name,
		"namespace":   tenant.Namespace,
		"domain":      tenant.Domain,
	}).Info("새 테넌트 생성 완료")

	return tenant, nil
}

// createTenantResources 테넌트용 Kubernetes 리소스 생성
func (tm *TenantManager) createTenantResources(ctx context.Context, tenant *Tenant) error {
	// 1. 네임스페이스 생성
	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: tenant.Namespace,
			Labels: map[string]string{
				"tenant.waf.io/id":   tenant.ID,
				"tenant.waf.io/name": tenant.Name,
				"managed-by":         "waf-tenant-manager",
			},
			Annotations: map[string]string{
				"tenant.waf.io/domain":     tenant.Domain,
				"tenant.waf.io/created-at": tenant.CreatedAt.Format(time.RFC3339),
				"tenant.waf.io/plan":       tenant.Metadata["plan"],
			},
		},
	}

	_, err := tm.clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("네임스페이스 생성 실패: %v", err)
	}

	// 2. 서비스 어카운트 생성
	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-service-account",
			Namespace: tenant.Namespace,
		},
	}

	_, err = tm.clientset.CoreV1().ServiceAccounts(tenant.Namespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("서비스 어카운트 생성 실패: %v", err)
	}

	// 3. 롤 생성 (네임스페이스 스코프)
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-role",
			Namespace: tenant.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: [""],
				Resources: ["pods", "services", "configmaps", "secrets"],
				Verbs:     ["get", "list", "watch", "create", "update", "patch", "delete"],
			},
			{
				APIGroups: ["networking.k8s.io"],
				Resources: ["ingresses"],
				Verbs:     ["get", "list", "watch", "create", "update", "patch", "delete"],
			},
		},
	}

	_, err = tm.clientset.RbacV1().Roles(tenant.Namespace).Create(ctx, role, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("롤 생성 실패: %v", err)
	}

	// 4. 롤 바인딩 생성
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-role-binding",
			Namespace: tenant.Namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "tenant-service-account",
				Namespace: tenant.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "tenant-role",
		},
	}

	_, err = tm.clientset.RbacV1().RoleBindings(tenant.Namespace).Create(ctx, roleBinding, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("롤 바인딩 생성 실패: %v", err)
	}

	// 5. 테넌트 설정용 ConfigMap 생성
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-config",
			Namespace: tenant.Namespace,
			Labels: map[string]string{
				"tenant.waf.io/id": tenant.ID,
			},
		},
		Data: map[string]string{
			"tenant_id":              tenant.ID,
			"domain":                 tenant.Domain,
			"rate_limit_rpm":         fmt.Sprintf("%d", tenant.Configuration.RateLimiting.RequestsPerMinute),
			"burst_size":             fmt.Sprintf("%d", tenant.Configuration.RateLimiting.BurstSize),
			"paranoia_level":         fmt.Sprintf("%d", tenant.Configuration.SecurityPolicy.ParanoiaLevel),
			"max_request_size":       fmt.Sprintf("%d", tenant.Configuration.SecurityPolicy.MaxRequestSize),
			"log_level":              tenant.Configuration.LoggingConfig.Level,
			"log_retention_days":     fmt.Sprintf("%d", tenant.Configuration.LoggingConfig.RetentionDays),
		},
	}

	_, err = tm.clientset.CoreV1().ConfigMaps(tenant.Namespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("ConfigMap 생성 실패: %v", err)
	}

	// 6. API 키 저장용 Secret 생성
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-api-key",
			Namespace: tenant.Namespace,
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			"api-key": []byte(tenant.APIKey),
		},
	}

	_, err = tm.clientset.CoreV1().Secrets(tenant.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("Secret 생성 실패: %v", err)
	}

	// 7. 테넌트용 Ingress 생성
	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-ingress",
			Namespace: tenant.Namespace,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":                    "nginx",
				"nginx.ingress.kubernetes.io/enable-modsecurity": "true",
				"nginx.ingress.kubernetes.io/modsecurity-transaction-id": "$request_id",
				"nginx.ingress.kubernetes.io/rate-limit":         fmt.Sprintf("%d", tenant.Configuration.RateLimiting.RequestsPerMinute),
				"nginx.ingress.kubernetes.io/rate-limit-window":  tenant.Configuration.RateLimiting.WindowSize,
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: tenant.Domain,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "waf-gateway-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 8080,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = tm.clientset.NetworkingV1().Ingresses(tenant.Namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("Ingress 생성 실패: %v", err)
	}

	return nil
}

// generateTenantID 테넌트 ID 생성
func (tm *TenantManager) generateTenantID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:12]
}

// generateAPIKey API 키 생성
func (tm *TenantManager) generateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "waf_" + hex.EncodeToString(bytes)
}

// CreateTenantRequest 테넌트 생성 요청
type CreateTenantRequest struct {
	Name              string `json:"name" binding:"required"`
	Domain            string `json:"domain" binding:"required"`
	ContactEmail      string `json:"contact_email" binding:"required,email"`
	Plan              string `json:"plan"`                                 // free, pro, enterprise
	RateLimitRPM      int    `json:"rate_limit_rpm"`                       // 분당 요청 수
	BurstSize         int    `json:"burst_size"`                           // 버스트 크기
	ParanoiaLevel     int    `json:"paranoia_level"`                       // 1-4
	DailyRequestLimit int64  `json:"daily_request_limit"`                  // 일일 요청 제한
	BandwidthLimitGB  int64  `json:"bandwidth_limit_gb"`                   // GB 단위
	MaxDomains        int    `json:"max_domains"`                          // 최대 도메인 수
}

// DeleteTenant 테넌트 삭제
func (tm *TenantManager) DeleteTenant(ctx context.Context, tenantID string) error {
	namespace := fmt.Sprintf("tenant-%s", tenantID)
	
	// 네임스페이스 삭제 (연관된 모든 리소스 함께 삭제)
	err := tm.clientset.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("테넌트 삭제 실패: %v", err)
	}

	tm.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"namespace": namespace,
	}).Info("테넌트 삭제 완료")

	return nil
}

// GetTenant 테넌트 정보 조회
func (tm *TenantManager) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	namespace := fmt.Sprintf("tenant-%s", tenantID)
	
	// 네임스페이스 조회
	ns, err := tm.clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("테넌트 조회 실패: %v", err)
	}

	// ConfigMap에서 설정 조회
	configMap, err := tm.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, "tenant-config", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("테넌트 설정 조회 실패: %v", err)
	}

	// Secret에서 API 키 조회
	secret, err := tm.clientset.CoreV1().Secrets(namespace).Get(ctx, "tenant-api-key", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("API 키 조회 실패: %v", err)
	}

	// Tenant 객체 재구성
	tenant := &Tenant{
		ID:        tenantID,
		Name:      ns.Labels["tenant.waf.io/name"],
		Domain:    ns.Annotations["tenant.waf.io/domain"],
		Namespace: namespace,
		APIKey:    string(secret.Data["api-key"]),
		Status:    TenantStatusActive,
		Metadata:  make(map[string]string),
	}

	// 생성 시간 파싱
	if createdAtStr, exists := ns.Annotations["tenant.waf.io/created-at"]; exists {
		if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			tenant.CreatedAt = createdAt
		}
	}

	// ConfigMap에서 설정 복원
	if rpm := configMap.Data["rate_limit_rpm"]; rpm != "" {
		fmt.Sscanf(rpm, "%d", &tenant.Configuration.RateLimiting.RequestsPerMinute)
	}

	return tenant, nil
}

// ListTenants 모든 테넌트 목록 조회
func (tm *TenantManager) ListTenants(ctx context.Context) ([]Tenant, error) {
	// 테넌트 네임스페이스 목록 조회
	namespaces, err := tm.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: "managed-by=waf-tenant-manager",
	})
	if err != nil {
		return nil, fmt.Errorf("테넌트 목록 조회 실패: %v", err)
	}

	var tenants []Tenant
	for _, ns := range namespaces.Items {
		if tenantID, exists := ns.Labels["tenant.waf.io/id"]; exists {
			tenant, err := tm.GetTenant(ctx, tenantID)
			if err != nil {
				tm.logger.WithError(err).WithField("tenant_id", tenantID).Error("테넌트 정보 조회 실패")
				continue
			}
			tenants = append(tenants, *tenant)
		}
	}

	return tenants, nil
}