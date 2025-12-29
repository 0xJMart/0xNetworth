package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"0xnetworth/backend/internal/integrations/coinbase"
	"0xnetworth/backend/internal/models"
	"0xnetworth/backend/internal/store"

	"github.com/gin-gonic/gin"
)

// SyncHandler handles data synchronization requests
type SyncHandler struct {
	store         *store.Store
	coinbaseClient *coinbase.Client
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(store *store.Store, coinbaseClient *coinbase.Client) *SyncHandler {
	return &SyncHandler{
		store:          store,
		coinbaseClient: coinbaseClient,
	}
}

// SyncAll triggers synchronization from all platforms
func (h *SyncHandler) SyncAll(c *gin.Context) {
	if h.coinbaseClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Coinbase client not configured",
		})
		return
	}

	// Sync from Coinbase
	portfolios, investments, err := h.coinbaseClient.SyncAll()
	if err != nil {
		log.Printf("Error syncing from Coinbase: %v", err)
		// Check if it's a 403 error from Coinbase API
		errMsg := err.Error()
		if strings.Contains(errMsg, "403") || strings.Contains(errMsg, "forbidden") {
			log.Printf("Coinbase API returned 403 Forbidden: %s", errMsg)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Coinbase API access forbidden: " + errMsg,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to sync from Coinbase: " + err.Error(),
		})
		return
	}

	// Store portfolios
	for _, portfolio := range portfolios {
		h.store.CreateOrUpdatePortfolio(portfolio)
	}

	// Store investments
	for _, investment := range investments {
		h.store.CreateOrUpdateInvestment(investment)
	}

	// Recalculate net worth
	h.store.RecalculateNetWorth()
	h.store.SetLastSyncTime(time.Now())

	c.JSON(http.StatusOK, gin.H{
		"message":   "sync completed successfully",
		"last_sync": h.store.GetLastSyncTime().Format(time.RFC3339),
		"portfolios_synced": len(portfolios),
		"investments_synced": len(investments),
	})
}

// SyncPlatform triggers synchronization for a specific platform
func (h *SyncHandler) SyncPlatform(c *gin.Context) {
	platformStr := c.Param("platform")
	platform := models.Platform(platformStr)

	if platform != models.PlatformCoinbase {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid platform. Only 'coinbase' is supported",
		})
		return
	}

	if h.coinbaseClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Coinbase client not configured",
		})
		return
	}

	// Sync from Coinbase
	portfolios, investments, err := h.coinbaseClient.SyncAll()
	if err != nil {
		log.Printf("Error syncing from Coinbase: %v", err)
		// Check if it's a 403 error from Coinbase API
		errMsg := err.Error()
		if strings.Contains(errMsg, "403") || strings.Contains(errMsg, "forbidden") {
			log.Printf("Coinbase API returned 403 Forbidden: %s", errMsg)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Coinbase API access forbidden: " + errMsg,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to sync from Coinbase: " + err.Error(),
		})
		return
	}

	// Store portfolios
	for _, portfolio := range portfolios {
		h.store.CreateOrUpdatePortfolio(portfolio)
	}

	// Store investments
	for _, investment := range investments {
		h.store.CreateOrUpdateInvestment(investment)
	}

	// Recalculate net worth
	h.store.RecalculateNetWorth()
	h.store.SetLastSyncTime(time.Now())

	c.JSON(http.StatusOK, gin.H{
		"message":   "sync completed successfully for " + platformStr,
		"platform":  platformStr,
		"last_sync": h.store.GetLastSyncTime().Format(time.RFC3339),
		"portfolios_synced": len(portfolios),
		"investments_synced": len(investments),
	})
}

