package main

import (
	"log"
	"os"

	"0xnetworth/backend/internal/handlers"
	"0xnetworth/backend/internal/integrations/coinbase"
	"0xnetworth/backend/internal/store"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize store
	store := store.NewStore()

	// Initialize Coinbase client if API keys are provided
	var coinbaseClient *coinbase.Client
	coinbaseAPIKey := os.Getenv("COINBASE_API_KEY")
	coinbaseAPISecret := os.Getenv("COINBASE_API_SECRET")
	if coinbaseAPIKey != "" && coinbaseAPISecret != "" {
		coinbaseClient = coinbase.NewClient(coinbaseAPIKey, coinbaseAPISecret)
		log.Println("Coinbase client initialized")
	} else {
		log.Println("Warning: Coinbase API keys not configured. Sync functionality will be limited.")
	}

	// Initialize handlers
	accountsHandler := handlers.NewAccountsHandler(store)
	investmentsHandler := handlers.NewInvestmentsHandler(store)
	networthHandler := handlers.NewNetWorthHandler(store)
	syncHandler := handlers.NewSyncHandler(store, coinbaseClient)

	// Setup router
	router := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(config))

	// Health check endpoint
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "0xnetworth-backend",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Account routes
		api.GET("/accounts", accountsHandler.GetAccounts)
		api.GET("/accounts/platform/:platform", accountsHandler.GetAccountsByPlatform)
		api.GET("/accounts/:id", accountsHandler.GetAccount)

		// Investment routes
		api.GET("/investments", investmentsHandler.GetInvestments)
		api.GET("/investments/account/:accountId", investmentsHandler.GetInvestmentsByAccount)
		api.GET("/investments/platform/:platform", investmentsHandler.GetInvestmentsByPlatform)

		// Net worth routes
		api.GET("/networth", networthHandler.GetNetWorth)
		api.GET("/networth/breakdown", networthHandler.GetNetWorthBreakdown)

		// Sync routes
		api.POST("/sync", syncHandler.SyncAll)
		api.POST("/sync/:platform", syncHandler.SyncPlatform)
	}

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server starting on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

