package handlers

import (
	"net/http"
	"waf-backend/dto"
	"waf-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RuleHandler struct {
	ruleService services.RuleService
	log         *logrus.Logger
}

func NewRuleHandler(ruleService services.RuleService, log *logrus.Logger) *RuleHandler {
	return &RuleHandler{
		ruleService: ruleService,
		log:         log,
	}
}

func (h *RuleHandler) CreateRule(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
			"code":  "ERR_NO_USER_ID",
		})
		return
	}
	
	var req dto.CustomRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Error("Invalid rule creation request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "ERR_INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}
	
	h.log.WithFields(logrus.Fields{
		"user_id":   userID,
		"rule_name": req.Name,
		"severity":  req.Severity,
	}).Info("Creating custom rule")
	
	rule, err := h.ruleService.CreateRule(userID, &req)
	if err != nil {
		h.log.WithError(err).Error("Failed to create rule")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "ERR_RULE_CREATION_FAILED",
		})
		return
	}
	
	h.log.WithFields(logrus.Fields{
		"rule_id":   rule.ID,
		"user_id":   userID,
		"rule_name": rule.Name,
	}).Info("Custom rule created successfully")
	
	c.JSON(http.StatusCreated, gin.H{
		"rule":    rule,
		"message": "Rule created successfully",
	})
}

func (h *RuleHandler) GetRules(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
			"code":  "ERR_NO_USER_ID",
		})
		return
	}
	
	h.log.WithField("user_id", userID).Debug("Fetching user rules")
	
	rules, err := h.ruleService.GetRules(userID)
	if err != nil {
		h.log.WithError(err).Error("Failed to fetch rules")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch rules",
			"code":  "ERR_FETCH_RULES_FAILED",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"count": len(rules),
	})
}

func (h *RuleHandler) GetRule(c *gin.Context) {
	userID := c.GetString("user_id")
	ruleID := c.Param("id")
	
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
			"code":  "ERR_NO_USER_ID",
		})
		return
	}
	
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Rule ID is required",
			"code":  "ERR_NO_RULE_ID",
		})
		return
	}
	
	h.log.WithFields(logrus.Fields{
		"user_id": userID,
		"rule_id": ruleID,
	}).Debug("Fetching specific rule")
	
	rule, err := h.ruleService.GetRule(userID, ruleID)
	if err != nil {
		h.log.WithError(err).Error("Failed to fetch rule")
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "ERR_RULE_NOT_FOUND",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"rule": rule,
	})
}

func (h *RuleHandler) UpdateRule(c *gin.Context) {
	userID := c.GetString("user_id")
	ruleID := c.Param("id")
	
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
			"code":  "ERR_NO_USER_ID",
		})
		return
	}
	
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Rule ID is required",
			"code":  "ERR_NO_RULE_ID",
		})
		return
	}
	
	var req dto.CustomRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Error("Invalid rule update request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "ERR_INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}
	
	h.log.WithFields(logrus.Fields{
		"user_id":   userID,
		"rule_id":   ruleID,
		"rule_name": req.Name,
	}).Info("Updating custom rule")
	
	rule, err := h.ruleService.UpdateRule(userID, ruleID, &req)
	if err != nil {
		h.log.WithError(err).Error("Failed to update rule")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "ERR_RULE_UPDATE_FAILED",
		})
		return
	}
	
	h.log.WithFields(logrus.Fields{
		"rule_id":   rule.ID,
		"user_id":   userID,
		"rule_name": rule.Name,
	}).Info("Custom rule updated successfully")
	
	c.JSON(http.StatusOK, gin.H{
		"rule":    rule,
		"message": "Rule updated successfully",
	})
}

func (h *RuleHandler) DeleteRule(c *gin.Context) {
	userID := c.GetString("user_id")
	ruleID := c.Param("id")
	
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
			"code":  "ERR_NO_USER_ID",
		})
		return
	}
	
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Rule ID is required",
			"code":  "ERR_NO_RULE_ID",
		})
		return
	}
	
	h.log.WithFields(logrus.Fields{
		"user_id": userID,
		"rule_id": ruleID,
	}).Info("Deleting custom rule")
	
	err := h.ruleService.DeleteRule(userID, ruleID)
	if err != nil {
		h.log.WithError(err).Error("Failed to delete rule")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "ERR_RULE_DELETE_FAILED",
		})
		return
	}
	
	h.log.WithFields(logrus.Fields{
		"rule_id": ruleID,
		"user_id": userID,
	}).Info("Custom rule deleted successfully")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Rule deleted successfully",
	})
}