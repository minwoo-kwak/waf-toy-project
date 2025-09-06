package repositories

import (
	"fmt"
	"waf-backend/database"
	"waf-backend/models"
	"gorm.io/gorm"
)

type RuleRepository interface {
	Create(rule *models.CustomRule) error
	GetByID(id string) (*models.CustomRule, error)
	GetByUserID(userID string) ([]*models.CustomRule, error)
	Update(rule *models.CustomRule) error
	Delete(id string) error
	DeleteByUserIDAndID(userID, id string) error
}

type ruleRepository struct {
	db *gorm.DB
}

func NewRuleRepository() RuleRepository {
	return &ruleRepository{
		db: nil, // Initialize as nil, will be set lazily
	}
}

func (r *ruleRepository) getDB() *gorm.DB {
	if r.db == nil {
		r.db = database.GetDB() // Lazy load DB connection
	}
	return r.db
}

func (r *ruleRepository) Create(rule *models.CustomRule) error {
	db := r.getDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}
	return db.Create(rule).Error
}

func (r *ruleRepository) GetByID(id string) (*models.CustomRule, error) {
	db := r.getDB()
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}
	var rule models.CustomRule
	err := db.Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *ruleRepository) GetByUserID(userID string) ([]*models.CustomRule, error) {
	db := r.getDB()
	if db == nil {
		return []*models.CustomRule{}, nil // Return empty slice when DB not available
	}
	var rules []*models.CustomRule
	err := db.Where("user_id = ?", userID).Find(&rules).Error
	return rules, err
}

func (r *ruleRepository) Update(rule *models.CustomRule) error {
	db := r.getDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}
	return db.Save(rule).Error
}

func (r *ruleRepository) Delete(id string) error {
	db := r.getDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}
	return db.Where("id = ?", id).Delete(&models.CustomRule{}).Error
}

func (r *ruleRepository) DeleteByUserIDAndID(userID, id string) error {
	db := r.getDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}
	return db.Where("user_id = ? AND id = ?", userID, id).Delete(&models.CustomRule{}).Error
}