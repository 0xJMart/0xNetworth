package handlers

import (
	"net/http"

	"0xnetworth/backend/internal/models"
	"0xnetworth/backend/internal/store"

	"github.com/gin-gonic/gin"
)

// AccountsHandler handles account-related HTTP requests
type AccountsHandler struct {
	store *store.Store
}

// NewAccountsHandler creates a new accounts handler
func NewAccountsHandler(store *store.Store) *AccountsHandler {
	return &AccountsHandler{
		store: store,
	}
}

// GetAccounts returns all accounts
func (h *AccountsHandler) GetAccounts(c *gin.Context) {
	accounts := h.store.GetAllAccounts()
	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
	})
}

// GetAccountsByPlatform returns accounts for a specific platform
func (h *AccountsHandler) GetAccountsByPlatform(c *gin.Context) {
	platformStr := c.Param("platform")
	platform := models.Platform(platformStr)

	// Validate platform
	if platform != models.PlatformCoinbase && platform != models.PlatformM1Finance {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid platform. Must be 'coinbase' or 'm1_finance'",
		})
		return
	}

	accounts := h.store.GetAccountsByPlatform(platform)
	c.JSON(http.StatusOK, gin.H{
		"platform": platform,
		"accounts": accounts,
	})
}

// GetAccount returns a single account by ID
func (h *AccountsHandler) GetAccount(c *gin.Context) {
	id := c.Param("id")
	account, exists := h.store.GetAccountByID(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "account not found",
		})
		return
	}
	c.JSON(http.StatusOK, account)
}

