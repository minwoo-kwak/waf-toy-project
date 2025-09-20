package handlers

import (
	"net/http"
	"waf-backend/dto"
	"waf-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ProxyHandler struct {
	proxyService services.ProxyService
	log          *logrus.Logger
}

func NewProxyHandler(proxyService services.ProxyService, log *logrus.Logger) *ProxyHandler {
	return &ProxyHandler{
		proxyService: proxyService,
		log:          log,
	}
}

func (h *ProxyHandler) CreateProxyTarget(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
			"code":  "ERR_NO_USER_ID",
		})
		return
	}

	var req dto.CreateProxyTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Error("Invalid proxy target creation request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"code":    "ERR_INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	h.log.WithFields(logrus.Fields{
		"user_id":      userID,
		"name":         req.Name,
		"origin_url":   req.OriginURL,
		"proxy_domain": req.ProxyDomain,
	}).Info("Creating proxy target")

	proxyTarget, err := h.proxyService.CreateProxyTarget(userID, &req)
	if err != nil {
		h.log.WithError(err).Error("Failed to create proxy target")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create proxy target",
			"code":  "ERR_CREATE_PROXY_FAILED",
		})
		return
	}

	h.log.WithFields(logrus.Fields{
		"user_id":  userID,
		"proxy_id": proxyTarget.ID,
	}).Info("Proxy target created successfully")

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Proxy target created successfully",
		"proxy_target": proxyTarget,
	})
}

func (h *ProxyHandler) GetProxyTargets(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
			"code":  "ERR_NO_USER_ID",
		})
		return
	}

	h.log.WithField("user_id", userID).Info("Getting proxy targets")

	response, err := h.proxyService.GetProxyTargets(userID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get proxy targets")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get proxy targets",
			"code":  "ERR_GET_PROXIES_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProxyHandler) GetProxyTarget(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
			"code":  "ERR_NO_USER_ID",
		})
		return
	}

	proxyID := c.Param("id")
	if proxyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Proxy ID is required",
			"code":  "ERR_MISSING_PROXY_ID",
		})
		return
	}

	h.log.WithFields(logrus.Fields{
		"user_id":  userID,
		"proxy_id": proxyID,
	}).Info("Getting proxy target by ID")

	proxyTarget, err := h.proxyService.GetProxyTargetByID(userID, proxyID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get proxy target")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Proxy target not found",
			"code":  "ERR_PROXY_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"proxy_target": proxyTarget,
	})
}

func (h *ProxyHandler) UpdateProxyTarget(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
			"code":  "ERR_NO_USER_ID",
		})
		return
	}

	proxyID := c.Param("id")
	if proxyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Proxy ID is required",
			"code":  "ERR_MISSING_PROXY_ID",
		})
		return
	}

	var req dto.UpdateProxyTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Error("Invalid proxy target update request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"code":    "ERR_INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	h.log.WithFields(logrus.Fields{
		"user_id":  userID,
		"proxy_id": proxyID,
	}).Info("Updating proxy target")

	proxyTarget, err := h.proxyService.UpdateProxyTarget(userID, proxyID, &req)
	if err != nil {
		h.log.WithError(err).Error("Failed to update proxy target")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update proxy target",
			"code":  "ERR_UPDATE_PROXY_FAILED",
		})
		return
	}

	h.log.WithFields(logrus.Fields{
		"user_id":  userID,
		"proxy_id": proxyID,
	}).Info("Proxy target updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"message":      "Proxy target updated successfully",
		"proxy_target": proxyTarget,
	})
}

func (h *ProxyHandler) DeleteProxyTarget(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
			"code":  "ERR_NO_USER_ID",
		})
		return
	}

	proxyID := c.Param("id")
	if proxyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Proxy ID is required",
			"code":  "ERR_MISSING_PROXY_ID",
		})
		return
	}

	h.log.WithFields(logrus.Fields{
		"user_id":  userID,
		"proxy_id": proxyID,
	}).Info("Deleting proxy target")

	err := h.proxyService.DeleteProxyTarget(userID, proxyID)
	if err != nil {
		h.log.WithError(err).Error("Failed to delete proxy target")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete proxy target",
			"code":  "ERR_DELETE_PROXY_FAILED",
		})
		return
	}

	h.log.WithFields(logrus.Fields{
		"user_id":  userID,
		"proxy_id": proxyID,
	}).Info("Proxy target deleted successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Proxy target deleted successfully",
	})
}

func (h *ProxyHandler) GetActiveProxyTargets(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
			"code":  "ERR_NO_USER_ID",
		})
		return
	}

	h.log.WithField("user_id", userID).Info("Getting active proxy targets")

	proxyTargets, err := h.proxyService.GetActiveProxyTargets(userID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get active proxy targets")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get active proxy targets",
			"code":  "ERR_GET_ACTIVE_PROXIES_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"proxy_targets": proxyTargets,
		"total":         len(proxyTargets),
	})
}