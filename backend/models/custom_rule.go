package models

import "time"

type CustomRule struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	RuleText    string    `gorm:"type:text;not null" json:"rule_text"`
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	Severity    string    `gorm:"default:MEDIUM" json:"severity"`
	UserID      string    `gorm:"not null;index" json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type User struct {
	ID       string `gorm:"primaryKey" json:"id"`
	Email    string `gorm:"unique;not null" json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Relationships
	CustomRules []CustomRule `gorm:"foreignKey:UserID" json:"custom_rules,omitempty"`
}