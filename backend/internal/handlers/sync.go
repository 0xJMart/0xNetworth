package handlers

import (
	"net/http"
	"time"

	"0xnetworth/backend/internal/store"

	"github.com/gin-gonic/gin"
)

// SyncHandler handles data synchronization requests
type SyncHandler struct {
	store *store.Store
	// Integration clients will be added in later phases
	// coinbaseClient *coinbase.Client
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(store *store.Store) *SyncHandler {
	return &SyncHandler{
		store: store,
	}
}

// SyncAll triggers synchronization from all platforms
func (h *SyncHandler) SyncAll(c *gin.Context) {
	// TODO: Implement actual sync logic in Phase 4 (Coinbase)
	// For now, just recalculate net worth and return success

	h.store.RecalculateNetWorth()
	h.store.SetLastSyncTime(time.Now())

	c.JSON(http.StatusOK, gin.H{
		"message": "sync triggered (placeholder - integrations not yet implemented)",
		"last_sync": h.store.GetLastSyncTime().Format(time.RFC3339),
	})
}

// SyncPlatform triggers synchronization for a specific platform
func (h *SyncHandler) SyncPlatform(c *gin.Context) {
	platform := c.Param("platform")

	// TODO: Implement platform-specific sync logic
	// For now, just return a placeholder response

	c.JSON(http.StatusOK, gin.H{
		"message": "sync triggered for platform: " + platform + " (placeholder - integrations not yet implemented)",
		"platform": platform,
	})
}

