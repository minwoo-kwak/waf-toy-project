package services

import "waf-backend/dto"

// RuleService interface defines methods for rule management
type RuleService interface {
	CreateRule(userID string, req *dto.CustomRuleRequest) (*dto.CustomRuleResponse, error)
	GetRules(userID string) ([]*dto.CustomRuleResponse, error)
	GetRule(userID, ruleID string) (*dto.CustomRuleResponse, error)
	UpdateRule(userID, ruleID string, req *dto.CustomRuleRequest) (*dto.CustomRuleResponse, error)
	DeleteRule(userID, ruleID string) error
}