package main

import (
	"log"
	"os"
	"strings"

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
	// Coinbase Advanced Trade API uses CDP API Keys for authentication
	// API Key Name can be UUID format (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
	// or full path format (organizations/{org_id}/apiKeys/{key_id})
	// See: https://docs.cdp.coinbase.com/api-reference/v2/authentication
	var coinbaseClient *coinbase.Client
	coinbaseAPIKeyName := os.Getenv("COINBASE_API_KEY_NAME")
	coinbaseAPIPrivateKey := os.Getenv("COINBASE_API_PRIVATE_KEY")
	// Support legacy environment variable names for backward compatibility
	if coinbaseAPIKeyName == "" {
		coinbaseAPIKeyName = os.Getenv("COINBASE_API_KEY")
	}
	if coinbaseAPIPrivateKey == "" {
		coinbaseAPIPrivateKey = os.Getenv("COINBASE_API_SECRET")
	}
	if coinbaseAPIKeyName != "" && coinbaseAPIPrivateKey != "" {
		var err error
		coinbaseClient, err = coinbase.NewClient(coinbaseAPIKeyName, coinbaseAPIPrivateKey)
		if err != nil {
			log.Fatalf("Failed to initialize Coinbase client: %v", err)
		}
		log.Println("Coinbase client initialized")
	} else {
		log.Println("Warning: Coinbase API keys not configured. Sync functionality will be limited.")
	}

	// Initialize handlers
	portfoliosHandler := handlers.NewPortfoliosHandler(store)
	investmentsHandler := handlers.NewInvestmentsHandler(store)
	networthHandler := handlers.NewNetWorthHandler(store)
	syncHandler := handlers.NewSyncHandler(store, coinbaseClient)

	// Setup router
	router := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	// Read allowed origins from environment variable, with fallback to defaults
	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsOrigins == "" {
		// Default: allow all localhost origins (any port) for development/port-forwarding
		// This makes it work regardless of which port you use for port-forwarding
		config.AllowOriginFunc = func(origin string) bool {
			// Allow all localhost origins (any port)
			if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "https://localhost:") {
				return true
			}
			// Also allow the standard development ports explicitly
			return origin == "http://localhost:5173" || origin == "http://localhost:3000"
		}
	} else {
		// Parse comma-separated origins
		origins := []string{}
		for _, origin := range strings.Split(corsOrigins, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				origins = append(origins, origin)
			}
		}
		if len(origins) > 0 {
			config.AllowOrigins = origins
		} else {
			// If empty after parsing, allow all origins (for development)
			config.AllowAllOrigins = true
		}
	}
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
		// Portfolio routes
		api.GET("/portfolios", portfoliosHandler.GetPortfolios)
		api.GET("/portfolios/platform/:platform", portfoliosHandler.GetPortfoliosByPlatform)
		api.GET("/portfolios/:id", portfoliosHandler.GetPortfolio)

		// Investment routes
		api.GET("/investments", investmentsHandler.GetInvestments)
		api.GET("/investments/portfolio/:portfolioId", investmentsHandler.GetInvestmentsByPortfolio)
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

