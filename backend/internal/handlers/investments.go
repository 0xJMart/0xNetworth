package handlers

import (
	"net/http"

	"0xnetworth/backend/internal/models"
	"0xnetworth/backend/internal/store"

	"github.com/gin-gonic/gin"
)

// InvestmentsHandler handles investment-related HTTP requests
type InvestmentsHandler struct {
	store *store.Store
}

// NewInvestmentsHandler creates a new investments handler
func NewInvestmentsHandler(store *store.Store) *InvestmentsHandler {
	return &InvestmentsHandler{
		store: store,
	}
}

// GetInvestments returns all investments
func (h *InvestmentsHandler) GetInvestments(c *gin.Context) {
	investments := h.store.GetAllInvestments()
	c.JSON(http.StatusOK, gin.H{
		"investments": investments,
	})
}

// GetInvestmentsByAccount returns investments for a specific account
func (h *InvestmentsHandler) GetInvestmentsByAccount(c *gin.Context) {
	accountID := c.Param("accountId")
	investments := h.store.GetInvestmentsByAccount(accountID)
	c.JSON(http.StatusOK, gin.H{
		"account_id": accountID,
		"investments": investments,
	})
}

// GetInvestmentsByPlatform returns investments for a specific platform
func (h *InvestmentsHandler) GetInvestmentsByPlatform(c *gin.Context) {
	platformStr := c.Param("platform")
	platform := models.Platform(platformStr)

	// Validate platform
	if platform != models.PlatformCoinbase {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid platform. Must be 'coinbase'",
		})
		return
	}

	investments := h.store.GetInvestmentsByPlatform(platform)
	c.JSON(http.StatusOK, gin.H{
		"platform": platform,
		"investments": investments,
	})
}

