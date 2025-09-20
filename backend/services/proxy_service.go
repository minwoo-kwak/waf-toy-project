package services

import (
	"fmt"
	"time"
	"waf-backend/dto"
	"waf-backend/models"
	"waf-backend/repositories"
	"waf-backend/services/k8s"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ProxyService interface {
	CreateProxyTarget(userID string, req *dto.CreateProxyTargetRequest) (*dto.ProxyTargetResponse, error)
	GetProxyTargets(userID string) (*dto.ProxyTargetListResponse, error)
	GetProxyTargetByID(userID, id string) (*dto.ProxyTargetResponse, error)
	UpdateProxyTarget(userID, id string, req *dto.UpdateProxyTargetRequest) (*dto.ProxyTargetResponse, error)
	DeleteProxyTarget(userID, id string) error
	GetActiveProxyTargets(userID string) ([]*dto.ProxyTargetResponse, error)
}

type proxyService struct {
	log        *logrus.Logger
	repo       repositories.ProxyRepository
	k8sService k8s.ProxyK8sService
}

func NewProxyService(log *logrus.Logger) ProxyService {
	service := &proxyService{
		log:        log,
		k8sService: k8s.NewProxyK8sService(log),
	}
	// Repository를 나중에 초기화 (DB가 준비된 후)
	service.repo = repositories.NewProxyRepository()

	log.Info("ProxyService initialized with DB support")

	return service
}

func (s *proxyService) CreateProxyTarget(userID string, req *dto.CreateProxyTargetRequest) (*dto.ProxyTargetResponse, error) {
	// 프록시 도메인 중복 검사
	existing, err := s.repo.GetByProxyDomain(req.ProxyDomain)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("proxy domain '%s' already exists", req.ProxyDomain)
	}

	// 새 프록시 타겟 생성
	proxy := &models.ProxyTarget{
		ID:          uuid.New().String(),
		Name:        req.Name,
		OriginURL:   req.OriginURL,
		ProxyDomain: req.ProxyDomain,
		Status:      "active",
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// DB에 저장
	if err := s.repo.Create(proxy); err != nil {
		s.log.WithError(err).Error("Failed to create proxy target in database")
		return nil, fmt.Errorf("failed to create proxy target: %w", err)
	}

	// K8s Ingress 생성
	if err := s.k8sService.CreateProxyIngress(proxy); err != nil {
		s.log.WithError(err).Error("Failed to create proxy ingress")
		// Ingress 생성 실패 시 DB에서 삭제
		s.repo.Delete(proxy.ID)
		return nil, fmt.Errorf("failed to create proxy ingress: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"proxy_id":     proxy.ID,
		"proxy_domain": proxy.ProxyDomain,
		"origin_url":   proxy.OriginURL,
		"user_id":      userID,
	}).Info("Proxy target created successfully")

	return s.convertToResponse(proxy), nil
}

func (s *proxyService) GetProxyTargets(userID string) (*dto.ProxyTargetListResponse, error) {
	proxies, err := s.repo.GetByUserID(userID)
	if err != nil {
		s.log.WithError(err).Error("Failed to get proxy targets from database")
		return nil, fmt.Errorf("failed to get proxy targets: %w", err)
	}

	responses := make([]dto.ProxyTargetResponse, len(proxies))
	for i, proxy := range proxies {
		responses[i] = *s.convertToResponse(proxy)
	}

	return &dto.ProxyTargetListResponse{
		ProxyTargets: responses,
		Total:        int64(len(responses)),
	}, nil
}

func (s *proxyService) GetProxyTargetByID(userID, id string) (*dto.ProxyTargetResponse, error) {
	proxy, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("proxy target not found: %w", err)
	}

	// 사용자 권한 확인
	if proxy.UserID != userID {
		return nil, fmt.Errorf("access denied: proxy target belongs to different user")
	}

	return s.convertToResponse(proxy), nil
}

func (s *proxyService) UpdateProxyTarget(userID, id string, req *dto.UpdateProxyTargetRequest) (*dto.ProxyTargetResponse, error) {
	proxy, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("proxy target not found: %w", err)
	}

	// 사용자 권한 확인
	if proxy.UserID != userID {
		return nil, fmt.Errorf("access denied: proxy target belongs to different user")
	}

	// 업데이트할 필드 적용
	updated := false
	if req.Name != "" && req.Name != proxy.Name {
		proxy.Name = req.Name
		updated = true
	}
	if req.Status != "" && req.Status != proxy.Status {
		proxy.Status = req.Status
		updated = true
	}
	if req.OriginURL != "" && req.OriginURL != proxy.OriginURL {
		proxy.OriginURL = req.OriginURL
		updated = true
	}
	if req.ProxyDomain != "" && req.ProxyDomain != proxy.ProxyDomain {
		// 프록시 도메인 중복 검사
		existing, err := s.repo.GetByProxyDomain(req.ProxyDomain)
		if err == nil && existing != nil && existing.ID != proxy.ID {
			return nil, fmt.Errorf("proxy domain '%s' already exists", req.ProxyDomain)
		}
		proxy.ProxyDomain = req.ProxyDomain
		updated = true
	}

	if !updated {
		return s.convertToResponse(proxy), nil
	}

	proxy.UpdatedAt = time.Now()

	// DB 업데이트
	if err := s.repo.Update(proxy); err != nil {
		s.log.WithError(err).Error("Failed to update proxy target in database")
		return nil, fmt.Errorf("failed to update proxy target: %w", err)
	}

	// K8s Ingress 업데이트
	if err := s.k8sService.UpdateProxyIngress(proxy); err != nil {
		s.log.WithError(err).Error("Failed to update proxy ingress")
		return nil, fmt.Errorf("failed to update proxy ingress: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"proxy_id": proxy.ID,
		"user_id":  userID,
	}).Info("Proxy target updated successfully")

	return s.convertToResponse(proxy), nil
}

func (s *proxyService) DeleteProxyTarget(userID, id string) error {
	proxy, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("proxy target not found: %w", err)
	}

	// 사용자 권한 확인
	if proxy.UserID != userID {
		return fmt.Errorf("access denied: proxy target belongs to different user")
	}

	// K8s Ingress 삭제
	if err := s.k8sService.DeleteProxyIngress(proxy); err != nil {
		s.log.WithError(err).Error("Failed to delete proxy ingress")
		return fmt.Errorf("failed to delete proxy ingress: %w", err)
	}

	// DB에서 삭제
	if err := s.repo.DeleteByUserIDAndID(userID, id); err != nil {
		s.log.WithError(err).Error("Failed to delete proxy target from database")
		return fmt.Errorf("failed to delete proxy target: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"proxy_id": id,
		"user_id":  userID,
	}).Info("Proxy target deleted successfully")

	return nil
}

func (s *proxyService) GetActiveProxyTargets(userID string) ([]*dto.ProxyTargetResponse, error) {
	proxies, err := s.repo.GetActiveByUserID(userID)
	if err != nil {
		s.log.WithError(err).Error("Failed to get active proxy targets from database")
		return nil, fmt.Errorf("failed to get active proxy targets: %w", err)
	}

	responses := make([]*dto.ProxyTargetResponse, len(proxies))
	for i, proxy := range proxies {
		responses[i] = s.convertToResponse(proxy)
	}

	return responses, nil
}

func (s *proxyService) convertToResponse(proxy *models.ProxyTarget) *dto.ProxyTargetResponse {
	return &dto.ProxyTargetResponse{
		ID:          proxy.ID,
		Name:        proxy.Name,
		OriginURL:   proxy.OriginURL,
		ProxyDomain: proxy.ProxyDomain,
		Status:      proxy.Status,
		UserID:      proxy.UserID,
		CreatedAt:   proxy.CreatedAt,
		UpdatedAt:   proxy.UpdatedAt,
	}
}