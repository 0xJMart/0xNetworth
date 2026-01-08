package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"0xnetworth/backend/internal/handlers"
	"0xnetworth/backend/internal/integrations/coinbase"
	workflowclient "0xnetworth/backend/internal/integrations/workflow"
	"0xnetworth/backend/internal/store"
	"0xnetworth/backend/internal/workflow"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize store - use PostgreSQL if DATABASE_URL is set, otherwise fall back to in-memory
	var storeInstance store.Store
	databaseURL := os.Getenv("DATABASE_URL")
	
	// Build DATABASE_URL from individual components if not provided
	if databaseURL == "" {
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbUser := os.Getenv("DB_USER")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")
		
		if dbHost != "" && dbUser != "" && dbPassword != "" && dbName != "" {
			if dbPort == "" {
				dbPort = "5432"
			}
			databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
		}
	}

	if databaseURL != "" {
		log.Println("Initializing PostgreSQL store...")
		postgresStore, err := store.NewPostgresStore(databaseURL)
		if err != nil {
			log.Fatalf("Failed to initialize PostgreSQL store: %v", err)
		}
		defer postgresStore.Close()

		// Read and execute schema
		// Get the path to schema.sql relative to the executable
		execPath, err := os.Executable()
		if err != nil {
			// Fallback: try relative path from current working directory
			execPath = "."
		}
		schemaPath := filepath.Join(filepath.Dir(execPath), "..", "internal", "store", "schema.sql")
		// Also try relative to current working directory (for development)
		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			schemaPath = filepath.Join("internal", "store", "schema.sql")
		}
		schemaSQL, err := os.ReadFile(schemaPath)
		if err != nil {
			log.Printf("Warning: Failed to read schema file at %s: %v. Schema may need to be initialized manually.", schemaPath, err)
		} else {

			if err := postgresStore.InitSchema(string(schemaSQL)); err != nil {
				log.Printf("Warning: Failed to initialize schema (may already exist): %v", err)
			} else {
				log.Println("Database schema initialized successfully")
			}
		}

		storeInstance = postgresStore
		log.Println("PostgreSQL store initialized successfully")
	} else {
		log.Println("Warning: DATABASE_URL not set, using in-memory store (data will not persist)")
		storeInstance = store.NewStore()
	}

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

	// Initialize workflow service client
	workflowServiceURL := os.Getenv("WORKFLOW_SERVICE_URL")
	if workflowServiceURL == "" {
		workflowServiceURL = "http://localhost:8000"
	}
	workflowClient := workflowclient.NewClient(workflowServiceURL)

	// Initialize workflow engine and scheduler
	workflowEngine := workflow.NewEngine(storeInstance, workflowClient)
	workflowScheduler := workflow.NewScheduler(storeInstance, workflowEngine)

	// Initialize handlers
	portfoliosHandler := handlers.NewPortfoliosHandler(storeInstance)
	investmentsHandler := handlers.NewInvestmentsHandler(storeInstance)
	networthHandler := handlers.NewNetWorthHandler(storeInstance)
	syncHandler := handlers.NewSyncHandler(storeInstance, coinbaseClient)
	workflowHandler := handlers.NewWorkflowHandler(storeInstance, workflowEngine, workflowScheduler)

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

		// Workflow routes
		api.POST("/workflow/execute", workflowHandler.ExecuteWorkflow)
		api.GET("/workflow/executions", workflowHandler.GetWorkflowExecutions)
		api.GET("/workflow/executions/:id", workflowHandler.GetWorkflowExecution)
		api.GET("/workflow/executions/:id/details", workflowHandler.GetWorkflowExecutionDetails)
		api.GET("/workflow/transcripts/:id", workflowHandler.GetTranscript)
		api.GET("/workflow/analyses/:id", workflowHandler.GetMarketAnalysis)
		api.GET("/workflow/recommendations/:id", workflowHandler.GetRecommendation)
		api.GET("/workflow/recommendations/summary", workflowHandler.GetRecommendationsSummary)
		api.POST("/workflow/sources", workflowHandler.CreateYouTubeSource)
		api.GET("/workflow/sources", workflowHandler.GetYouTubeSources)
		api.GET("/workflow/sources/:id", workflowHandler.GetYouTubeSource)
		api.PUT("/workflow/sources/:id", workflowHandler.UpdateYouTubeSource)
		api.DELETE("/workflow/sources/:id", workflowHandler.DeleteYouTubeSource)
		api.POST("/workflow/sources/:id/schedule", workflowHandler.UpdateSourceSchedule)
		api.POST("/workflow/sources/test", workflowHandler.TestYouTubeSource)
		api.POST("/workflow/sources/trigger-all", workflowHandler.TriggerAllSources)
	}

	// Start workflow scheduler
	workflowScheduler.Start()
	defer workflowScheduler.Stop()

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

