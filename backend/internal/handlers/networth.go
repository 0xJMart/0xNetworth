package handlers

import (
	"net/http"

	"0xnetworth/backend/internal/store"

	"github.com/gin-gonic/gin"
)

// NetWorthHandler handles net worth-related HTTP requests
type NetWorthHandler struct {
	store *store.Store
}

// NewNetWorthHandler creates a new net worth handler
func NewNetWorthHandler(store *store.Store) *NetWorthHandler {
	return &NetWorthHandler{
		store: store,
	}
}

// GetNetWorth returns the current net worth
func (h *NetWorthHandler) GetNetWorth(c *gin.Context) {
	// Recalculate before returning to ensure accuracy
	h.store.RecalculateNetWorth()
	networth := h.store.GetNetWorth()
	c.JSON(http.StatusOK, networth)
}

// GetNetWorthBreakdown returns detailed breakdown of net worth
func (h *NetWorthHandler) GetNetWorthBreakdown(c *gin.Context) {
	// Recalculate before returning
	h.store.RecalculateNetWorth()
	networth := h.store.GetNetWorth()
	accounts := h.store.GetAllAccounts()
	investments := h.store.GetAllInvestments()

	c.JSON(http.StatusOK, gin.H{
		"networth":   networth,
		"accounts":   accounts,
		"investments": investments,
	})
}

