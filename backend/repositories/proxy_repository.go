package repositories

import (
	"fmt"
	"waf-backend/database"
	"waf-backend/models"
	"gorm.io/gorm"
)

type ProxyRepository interface {
	Create(proxy *models.ProxyTarget) error
	GetByID(id string) (*models.ProxyTarget, error)
	GetByUserID(userID string) ([]*models.ProxyTarget, error)
	GetByProxyDomain(domain string) (*models.ProxyTarget, error)
	Update(proxy *models.ProxyTarget) error
	Delete(id string) error
	DeleteByUserIDAndID(userID, id string) error
	GetActiveByUserID(userID string) ([]*models.ProxyTarget, error)
}

type proxyRepository struct {
	db *gorm.DB
}

func NewProxyRepository() ProxyRepository {
	return &proxyRepository{
		db: nil, // Initialize as nil, will be set lazily
	}
}

func (r *proxyRepository) getDB() *gorm.DB {
	if r.db == nil {
		r.db = database.GetDB() // Lazy load DB connection
	}
	return r.db
}

func (r *proxyRepository) Create(proxy *models.ProxyTarget) error {
	db := r.getDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}
	return db.Create(proxy).Error
}

func (r *proxyRepository) GetByID(id string) (*models.ProxyTarget, error) {
	db := r.getDB()
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}
	var proxy models.ProxyTarget
	err := db.Where("id = ?", id).First(&proxy).Error
	if err != nil {
		return nil, err
	}
	return &proxy, nil
}

func (r *proxyRepository) GetByUserID(userID string) ([]*models.ProxyTarget, error) {
	db := r.getDB()
	if db == nil {
		return []*models.ProxyTarget{}, nil // Return empty slice when DB not available
	}
	var proxies []*models.ProxyTarget
	err := db.Where("user_id = ?", userID).Order("created_at DESC").Find(&proxies).Error
	return proxies, err
}

func (r *proxyRepository) GetByProxyDomain(domain string) (*models.ProxyTarget, error) {
	db := r.getDB()
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}
	var proxy models.ProxyTarget
	err := db.Where("proxy_domain = ?", domain).First(&proxy).Error
	if err != nil {
		return nil, err
	}
	return &proxy, nil
}

func (r *proxyRepository) Update(proxy *models.ProxyTarget) error {
	db := r.getDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}
	return db.Save(proxy).Error
}

func (r *proxyRepository) Delete(id string) error {
	db := r.getDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}
	return db.Where("id = ?", id).Delete(&models.ProxyTarget{}).Error
}

func (r *proxyRepository) DeleteByUserIDAndID(userID, id string) error {
	db := r.getDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}
	return db.Where("user_id = ? AND id = ?", userID, id).Delete(&models.ProxyTarget{}).Error
}

func (r *proxyRepository) GetActiveByUserID(userID string) ([]*models.ProxyTarget, error) {
	db := r.getDB()
	if db == nil {
		return []*models.ProxyTarget{}, nil
	}
	var proxies []*models.ProxyTarget
	err := db.Where("user_id = ? AND status = ?", userID, "active").Order("created_at DESC").Find(&proxies).Error
	return proxies, err
}