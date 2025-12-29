package handlers

import (
	"net/http"

	"0xnetworth/backend/internal/models"
	"0xnetworth/backend/internal/store"

	"github.com/gin-gonic/gin"
)

// PortfoliosHandler handles portfolio-related HTTP requests
type PortfoliosHandler struct {
	store *store.Store
}

// NewPortfoliosHandler creates a new portfolios handler
func NewPortfoliosHandler(store *store.Store) *PortfoliosHandler {
	return &PortfoliosHandler{
		store: store,
	}
}

// GetPortfolios returns all portfolios
func (h *PortfoliosHandler) GetPortfolios(c *gin.Context) {
	portfolios := h.store.GetAllPortfolios()
	c.JSON(http.StatusOK, gin.H{
		"portfolios": portfolios,
	})
}

// GetPortfoliosByPlatform returns portfolios for a specific platform
func (h *PortfoliosHandler) GetPortfoliosByPlatform(c *gin.Context) {
	platformStr := c.Param("platform")
	platform := models.Platform(platformStr)

	// Validate platform
	if platform != models.PlatformCoinbase {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid platform. Must be 'coinbase'",
		})
		return
	}

	portfolios := h.store.GetPortfoliosByPlatform(platform)
	c.JSON(http.StatusOK, gin.H{
		"platform": platform,
		"portfolios": portfolios,
	})
}

// GetPortfolio returns a portfolio by ID
func (h *PortfoliosHandler) GetPortfolio(c *gin.Context) {
	portfolioID := c.Param("id")
	portfolio, exists := h.store.GetPortfolioByID(portfolioID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "portfolio not found",
		})
		return
	}

	c.JSON(http.StatusOK, portfolio)
}

