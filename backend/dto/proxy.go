package dto

import "time"

type CreateProxyTargetRequest struct {
	Name        string `json:"name" binding:"required"`
	OriginURL   string `json:"origin_url" binding:"required,url"`
	ProxyDomain string `json:"proxy_domain" binding:"required"`
}

type UpdateProxyTargetRequest struct {
	Name        string `json:"name"`
	OriginURL   string `json:"origin_url" binding:"omitempty,url"`
	ProxyDomain string `json:"proxy_domain"`
	Status      string `json:"status"`
}

type ProxyTargetResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	OriginURL    string    `json:"origin_url"`
	ProxyDomain  string    `json:"proxy_domain"`
	Status       string    `json:"status"`
	UserID       string    `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ProxyTargetListResponse struct {
	ProxyTargets []ProxyTargetResponse `json:"proxy_targets"`
	Total        int64                 `json:"total"`
}