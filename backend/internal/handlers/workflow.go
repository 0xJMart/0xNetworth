package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"

	"0xnetworth/backend/internal/integrations/scraper"
	"0xnetworth/backend/internal/integrations/youtube"
	"0xnetworth/backend/internal/models"
	"0xnetworth/backend/internal/store"
	"0xnetworth/backend/internal/workflow"
)

const (
	// RecentRecommendationsLimit is the maximum number of recent recommendations to return
	RecentRecommendationsLimit = 10
)

// WorkflowHandler handles workflow-related HTTP requests
type WorkflowHandler struct {
	store    store.Store
	engine   *workflow.Engine
	scheduler *workflow.Scheduler
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(store store.Store, engine *workflow.Engine, scheduler *workflow.Scheduler) *WorkflowHandler {
	return &WorkflowHandler{
		store:     store,
		engine:    engine,
		scheduler: scheduler,
	}
}

// ExecuteWorkflowRequest represents the request body for executing a workflow
type ExecuteWorkflowRequest struct {
	YouTubeURL string `json:"youtube_url" binding:"required"`
	SourceID   string `json:"source_id,omitempty"`
}

// ExecuteWorkflow handles POST /api/workflow/execute
func (h *WorkflowHandler) ExecuteWorkflow(c *gin.Context) {
	var req ExecuteWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	execution, err := h.engine.ExecuteWorkflow(req.YouTubeURL, req.SourceID)
	if err != nil {
		log.Printf("Error executing workflow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to execute workflow: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// GetWorkflowExecutions handles GET /api/workflow/executions
func (h *WorkflowHandler) GetWorkflowExecutions(c *gin.Context) {
	executions := h.store.GetAllWorkflowExecutions()
	c.JSON(http.StatusOK, executions)
}

// GetWorkflowExecution handles GET /api/workflow/executions/:id
func (h *WorkflowHandler) GetWorkflowExecution(c *gin.Context) {
	id := c.Param("id")
	
	execution, exists := h.store.GetWorkflowExecutionByID(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// CreateYouTubeSourceRequest represents the request body for creating a YouTube source
type CreateYouTubeSourceRequest struct {
	Type     models.YouTubeSourceType `json:"type" binding:"required"`
	URL      string                   `json:"url" binding:"required"`
	Name     string                   `json:"name" binding:"required"`
	Enabled  bool                     `json:"enabled"`
	Schedule string                   `json:"schedule,omitempty"`
	AuthEmail string                  `json:"auth_email,omitempty"` // For web scraper sources
}

// CreateYouTubeSource handles POST /api/workflow/sources
func (h *WorkflowHandler) CreateYouTubeSource(c *gin.Context) {
	var req CreateYouTubeSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	source := &models.YouTubeSource{
		ID:        uuid.New().String(),
		Type:      req.Type,
		URL:       req.URL,
		Name:      req.Name,
		Enabled:   req.Enabled,
		Schedule:  req.Schedule,
		AuthEmail: req.AuthEmail,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	h.store.CreateOrUpdateYouTubeSource(source)
	
	// Schedule the source if it's enabled
	if h.scheduler != nil && source.Enabled {
		if err := h.scheduler.ReloadSourceSchedule(source.ID); err != nil {
			log.Printf("Failed to schedule newly created source %s: %v", source.ID, err)
			// Don't fail the request, just log the error
		}
	}
	
	c.JSON(http.StatusCreated, source)
}

// GetYouTubeSources handles GET /api/workflow/sources
func (h *WorkflowHandler) GetYouTubeSources(c *gin.Context) {
	sources := h.store.GetAllYouTubeSources()
	c.JSON(http.StatusOK, sources)
}

// GetYouTubeSource handles GET /api/workflow/sources/:id
func (h *WorkflowHandler) GetYouTubeSource(c *gin.Context) {
	id := c.Param("id")
	
	source, exists := h.store.GetYouTubeSourceByID(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
		return
	}

	c.JSON(http.StatusOK, source)
}

// DeleteYouTubeSource handles DELETE /api/workflow/sources/:id
func (h *WorkflowHandler) DeleteYouTubeSource(c *gin.Context) {
	id := c.Param("id")
	
	if !h.store.DeleteYouTubeSource(id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// UpdateSourceScheduleRequest represents the request body for updating a source schedule
type UpdateSourceScheduleRequest struct {
	Schedule string `json:"schedule" binding:"required"`
}

// UpdateSourceSchedule handles POST /api/workflow/sources/:id/schedule
func (h *WorkflowHandler) UpdateSourceSchedule(c *gin.Context) {
	id := c.Param("id")
	
	source, exists := h.store.GetYouTubeSourceByID(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
		return
	}

	var req UpdateSourceScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	source.Schedule = req.Schedule
	h.store.CreateOrUpdateYouTubeSource(source)

	// Reload the schedule in the scheduler
	if h.scheduler != nil {
		if err := h.scheduler.ReloadSourceSchedule(id); err != nil {
			log.Printf("Failed to reload schedule for source %s: %v", id, err)
			// Don't fail the request, just log the error
		}
	}

	c.JSON(http.StatusOK, source)
}

// UpdateYouTubeSource handles PUT /api/workflow/sources/:id
func (h *WorkflowHandler) UpdateYouTubeSource(c *gin.Context) {
	id := c.Param("id")
	
	source, exists := h.store.GetYouTubeSourceByID(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
		return
	}

	var req CreateYouTubeSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update source fields
	source.Type = req.Type
	source.URL = req.URL
	source.Name = req.Name
	source.Enabled = req.Enabled
	if req.Schedule != "" {
		source.Schedule = req.Schedule
	}
	if req.AuthEmail != "" {
		source.AuthEmail = req.AuthEmail
	}

	h.store.CreateOrUpdateYouTubeSource(source)
	
	// Reload the schedule in the scheduler if schedule or enabled status changed
	if h.scheduler != nil && (req.Schedule != "" || req.Enabled != source.Enabled) {
		if err := h.scheduler.ReloadSourceSchedule(id); err != nil {
			log.Printf("Failed to reload schedule for source %s: %v", id, err)
			// Don't fail the request, just log the error
		}
	}
	
	c.JSON(http.StatusOK, source)
}

// TestYouTubeSourceRequest represents the request body for testing a YouTube source
type TestYouTubeSourceRequest struct {
	URL string `json:"url" binding:"required"`
}

// TestYouTubeSource handles POST /api/workflow/sources/test
func (h *WorkflowHandler) TestYouTubeSource(c *gin.Context) {
	var req TestYouTubeSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get YouTube client from scheduler (if available)
	// We need to access the scheduler's youtubeClient
	// For now, we'll create a temporary client
	youtubeAPIKey := os.Getenv("YOUTUBE_API_KEY")
	if youtubeAPIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "YouTube API key not configured"})
		return
	}

	// Import youtube client
	youtubeClient := youtube.NewClient(youtubeAPIKey)
	if youtubeClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize YouTube client"})
		return
	}

	// Try to extract/resolve channel ID
	channelID, err := youtubeClient.ExtractChannelID(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the channel exists by trying to fetch videos
	_, err = youtubeClient.GetChannelVideos(channelID, 1, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Channel not found or inaccessible: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"channel_id": channelID,
		"message": "Channel found and accessible",
	})
}

// TestWebScraperSourceRequest represents the request body for testing a web scraper source
type TestWebScraperSourceRequest struct {
	URL      string `json:"url" binding:"required"`
	Email    string `json:"email,omitempty"`
	OTPCode  string `json:"otp_code,omitempty"`
}

// TestWebScraperSource handles POST /api/workflow/sources/:id/test-scraper
func (h *WorkflowHandler) TestWebScraperSource(c *gin.Context) {
	id := c.Param("id")
	
	source, exists := h.store.GetYouTubeSourceByID(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
		return
	}
	
	if source.Type != models.YouTubeSourceTypeWebScraper {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source is not a web scraper source"})
		return
	}
	
	var req TestWebScraperSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Use source URL if not provided in request
	if req.URL == "" {
		req.URL = source.URL
	}
	
	// Use source email if not provided in request
	if req.Email == "" {
		req.Email = source.AuthEmail
	}
	
	// Initialize scraper client
	headless := os.Getenv("SCRAPER_HEADLESS")
	if headless == "" {
		headless = "true"
	}
	headlessMode := headless == "true"
	
	timeoutStr := os.Getenv("SCRAPER_TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "30s"
	}
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		timeout = 30 * time.Second
	}
	
	scraperClient := scraper.NewClient(headlessMode, timeout)
	ctx := context.Background()
	
	// Try to load existing session if we have cookies and no OTP provided
	if source.AuthSessionCookie != "" && req.OTPCode == "" {
		cookies, err := scraper.DeserializeCookies(source.AuthSessionCookie)
		if err == nil {
			err = scraperClient.LoadSession(ctx, cookies, req.URL)
			if err == nil {
				// Test URL extraction
				result, err := scraperClient.ExtractYouTubeURLs(ctx, req.URL)
				if err == nil && result.Error == "" {
					c.JSON(http.StatusOK, gin.H{
						"success": true,
						"message": "Session valid and URLs extracted successfully",
						"urls_found": len(result.URLs),
						"urls": result.URLs,
					})
					return
				}
			}
		}
	}
	
	// If OTP is provided, try authentication
	if req.OTPCode != "" && req.Email != "" {
		loginURL := "https://app.intothecryptoverse.com/login" // Default login URL
		authResult, err := scraperClient.Authenticate(ctx, req.Email, req.OTPCode, loginURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": err.Error(),
			})
			return
		}
		
		if !authResult.Success {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": authResult.Error,
			})
			return
		}
		
		// Save session cookies
		cookieData, err := scraper.SerializeCookies(authResult.SessionCookies)
		if err == nil {
			source.AuthSessionCookie = cookieData
			source.AuthEmail = req.Email
			source.AuthLastRefreshed = time.Now().UTC().Format(time.RFC3339)
			h.store.CreateOrUpdateYouTubeSource(source)
		}
		
		// Test URL extraction
		result, err := scraperClient.ExtractYouTubeURLs(ctx, req.URL)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Authentication successful, but URL extraction failed",
				"error": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Authentication successful and URLs extracted",
			"urls_found": len(result.URLs),
			"urls": result.URLs,
		})
		return
	}
	
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"error": "No valid session found and no OTP code provided. Please provide OTP code for authentication.",
	})
}

// RefreshWebScraperAuthRequest represents the request body for refreshing web scraper authentication
type RefreshWebScraperAuthRequest struct {
	Email   string `json:"email" binding:"required"`
	OTPCode string `json:"otp_code" binding:"required"`
}

// RefreshWebScraperAuth handles POST /api/workflow/sources/:id/refresh-auth
func (h *WorkflowHandler) RefreshWebScraperAuth(c *gin.Context) {
	id := c.Param("id")
	
	source, exists := h.store.GetYouTubeSourceByID(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
		return
	}
	
	if source.Type != models.YouTubeSourceTypeWebScraper {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source is not a web scraper source"})
		return
	}
	
	var req RefreshWebScraperAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Initialize scraper client
	headless := os.Getenv("SCRAPER_HEADLESS")
	if headless == "" {
		headless = "true"
	}
	headlessMode := headless == "true"
	
	timeoutStr := os.Getenv("SCRAPER_TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "30s"
	}
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		timeout = 30 * time.Second
	}
	
	scraperClient := scraper.NewClient(headlessMode, timeout)
	ctx := context.Background()
	
	// Authenticate
	loginURL := "https://app.intothecryptoverse.com/login" // Default login URL
	authResult, err := scraperClient.Authenticate(ctx, req.Email, req.OTPCode, loginURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": err.Error(),
		})
		return
	}
	
	if !authResult.Success {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": authResult.Error,
		})
		return
	}
	
	// Save session cookies
	cookieData, err := scraper.SerializeCookies(authResult.SessionCookies)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": fmt.Sprintf("Failed to serialize cookies: %v", err),
		})
		return
	}
	
	source.AuthSessionCookie = cookieData
	source.AuthEmail = req.Email
	source.AuthLastRefreshed = time.Now().UTC().Format(time.RFC3339)
	h.store.CreateOrUpdateYouTubeSource(source)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Authentication refreshed successfully",
		"cookies_count": len(authResult.SessionCookies),
	})
}

// GetWebScraperAuthStatus handles GET /api/workflow/sources/:id/auth-status
func (h *WorkflowHandler) GetWebScraperAuthStatus(c *gin.Context) {
	id := c.Param("id")
	
	source, exists := h.store.GetYouTubeSourceByID(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
		return
	}
	
	if source.Type != models.YouTubeSourceTypeWebScraper {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source is not a web scraper source"})
		return
	}
	
	hasSession := source.AuthSessionCookie != ""
	hasEmail := source.AuthEmail != ""
	
	c.JSON(http.StatusOK, gin.H{
		"has_session": hasSession,
		"has_email": hasEmail,
		"email": source.AuthEmail,
		"last_refreshed": source.AuthLastRefreshed,
		"needs_auth": !hasSession,
	})
}

// TriggerAllSources handles POST /api/workflow/sources/trigger-all
func (h *WorkflowHandler) TriggerAllSources(c *gin.Context) {
	if h.scheduler == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Scheduler not available"})
		return
	}
	
	triggered := h.scheduler.TriggerAllSources()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Triggered %d source(s)", len(triggered)),
		"triggered_sources": triggered,
		"count": len(triggered),
	})
}

// GetTranscript handles GET /api/workflow/transcripts/:id
func (h *WorkflowHandler) GetTranscript(c *gin.Context) {
	id := c.Param("id")
	
	transcript, exists := h.store.GetTranscriptByID(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "transcript not found"})
		return
	}

	c.JSON(http.StatusOK, transcript)
}

// GetMarketAnalysis handles GET /api/workflow/analyses/:id
func (h *WorkflowHandler) GetMarketAnalysis(c *gin.Context) {
	id := c.Param("id")
	
	analysis, exists := h.store.GetMarketAnalysisByID(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "analysis not found"})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetRecommendation handles GET /api/workflow/recommendations/:id
func (h *WorkflowHandler) GetRecommendation(c *gin.Context) {
	id := c.Param("id")
	
	recommendation, exists := h.store.GetRecommendationByID(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "recommendation not found"})
		return
	}

	c.JSON(http.StatusOK, recommendation)
}

// GetWorkflowExecutionDetails handles GET /api/workflow/executions/:id/details
// Returns the full execution with all related data (transcript, analysis, recommendation)
func (h *WorkflowHandler) GetWorkflowExecutionDetails(c *gin.Context) {
	id := c.Param("id")
	
	execution, exists := h.store.GetWorkflowExecutionByID(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	// Build response with all related data
	response := gin.H{
		"execution": execution,
	}

	// Add transcript if available
	if execution.TranscriptID != "" {
		transcript, exists := h.store.GetTranscriptByID(execution.TranscriptID)
		if exists {
			response["transcript"] = transcript
		}
	}

	// Add market analysis if available
	if execution.AnalysisID != "" {
		analysis, exists := h.store.GetMarketAnalysisByID(execution.AnalysisID)
		if exists {
			response["market_analysis"] = analysis
		}
	}

	// Add recommendation if available
	if execution.RecommendationID != "" {
		recommendation, exists := h.store.GetRecommendationByID(execution.RecommendationID)
		if exists {
			response["recommendation"] = recommendation
		}
	}

	c.JSON(http.StatusOK, response)
}

// RecommendationsSummary represents aggregated recommendation data
type RecommendationsSummary struct {
	TotalCount           int                `json:"total_count"`
	ActionDistribution   map[string]int     `json:"action_distribution"`
	AverageConfidence    float64            `json:"average_confidence"`
	ConditionDistribution map[string]int    `json:"condition_distribution"`
	RecentRecommendations []RecommendationSummaryItem `json:"recent_recommendations"`
	AggregatedRecommendation *AggregatedRecommendationResponse `json:"aggregated_recommendation,omitempty"` // AI-generated consolidated recommendation
}

// AggregatedRecommendationResponse represents the response from the aggregated recommendation agent
type AggregatedRecommendationResponse struct {
	Action           string           `json:"action"`
	Confidence       float64          `json:"confidence"`
	SuggestedActions []SuggestedActionResponse `json:"suggested_actions"`
	Summary          string           `json:"summary"`
	KeyInsights      []string         `json:"key_insights"`
}

// SuggestedActionResponse represents a suggested action in the aggregated recommendation
type SuggestedActionResponse struct {
	Type      string `json:"type"`
	Symbol    string `json:"symbol"`
	Rationale string `json:"rationale"`
}

// RecommendationSummaryItem represents a single recommendation in the summary
type RecommendationSummaryItem struct {
	ExecutionID   string  `json:"execution_id"`
	VideoTitle    string  `json:"video_title"`
	VideoID       string  `json:"video_id"`
	Action        string  `json:"action"`
	Confidence    float64 `json:"confidence"`
	Condition     string  `json:"condition"`
	CompletedAt   string  `json:"completed_at"`
}

// GetRecommendationsSummary handles GET /api/workflow/recommendations/summary
func (h *WorkflowHandler) GetRecommendationsSummary(c *gin.Context) {
	// Get days parameter (default 7)
	daysStr := c.DefaultQuery("days", "7")
	days := 7
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
		days = d
	}
	
	// Calculate cutoff time
	cutoffTime := time.Now().UTC().AddDate(0, 0, -days)
	
	// Get all executions
	allExecutions := h.store.GetAllWorkflowExecutions()
	
	// Filter executions from the past N days with recommendations
	recentExecutions := make([]*models.WorkflowExecution, 0)
	// Also collect all completed executions for aggregated summary (regardless of days)
	allCompletedExecutions := make([]*models.WorkflowExecution, 0)
	
	for _, exec := range allExecutions {
		if exec.Status != models.WorkflowStatusCompleted {
			continue
		}
		if exec.RecommendationID == "" {
			continue
		}
		if exec.CompletedAt == "" {
			continue
		}
		
		completedAt, err := time.Parse(time.RFC3339, exec.CompletedAt)
		if err != nil {
			log.Printf("Warning: Invalid CompletedAt timestamp for execution %s: %v", exec.ID, err)
			continue
		}
		
		// Add to all completed for aggregated summary
		allCompletedExecutions = append(allCompletedExecutions, exec)
		
		if completedAt.After(cutoffTime) {
			recentExecutions = append(recentExecutions, exec)
		}
	}
	
	// Build summary
	summary := RecommendationsSummary{
		TotalCount:          len(recentExecutions),
		ActionDistribution:  make(map[string]int),
		ConditionDistribution: make(map[string]int),
		RecentRecommendations: make([]RecommendationSummaryItem, 0, len(recentExecutions)),
	}
	
	totalConfidence := 0.0
	validConfidenceCount := 0
	
	// Collect all recommendation items first
	allRecommendationItems := make([]RecommendationSummaryItem, 0, len(recentExecutions))
	
	// Process each execution
	for _, exec := range recentExecutions {
		// Get recommendation
		rec, exists := h.store.GetRecommendationByID(exec.RecommendationID)
		if !exists {
			continue
		}
		
		// Get market analysis for condition
		condition := "unknown"
		if exec.AnalysisID != "" {
			analysis, exists := h.store.GetMarketAnalysisByID(exec.AnalysisID)
			if exists {
				condition = analysis.Conditions
				summary.ConditionDistribution[condition]++
			}
		}
		
		// Track action distribution
		summary.ActionDistribution[rec.Action]++
		
		// Track confidence
		if rec.Confidence > 0 {
			totalConfidence += rec.Confidence
			validConfidenceCount++
		}
		
		// Collect all items for sorting
		item := RecommendationSummaryItem{
			ExecutionID: exec.ID,
			VideoTitle:   exec.VideoTitle,
			VideoID:      exec.VideoID,
			Action:       rec.Action,
			Confidence:   rec.Confidence,
			Condition:   condition,
			CompletedAt:  exec.CompletedAt,
		}
		allRecommendationItems = append(allRecommendationItems, item)
	}
	
	// Sort by completed_at (newest first), then take top N
	sort.Slice(allRecommendationItems, func(i, j int) bool {
		return allRecommendationItems[i].CompletedAt > allRecommendationItems[j].CompletedAt
	})
	
	// Take top N most recent
	limit := RecentRecommendationsLimit
	if len(allRecommendationItems) < limit {
		limit = len(allRecommendationItems)
	}
	summary.RecentRecommendations = allRecommendationItems[:limit]
	
	// Calculate average confidence
	if validConfidenceCount > 0 {
		summary.AverageConfidence = totalConfidence / float64(validConfidenceCount)
	}
	
	// Get cached aggregated recommendation if it exists (don't auto-generate)
	cachedRec, exists := h.store.GetLatestAggregatedRecommendation()
	if exists {
		// Convert to response format
		suggestedActions := make([]SuggestedActionResponse, len(cachedRec.SuggestedActions))
		for i, sa := range cachedRec.SuggestedActions {
			suggestedActions[i] = SuggestedActionResponse{
				Type:      sa.Type,
				Symbol:    sa.Symbol,
				Rationale: sa.Rationale,
			}
		}
		
		summary.AggregatedRecommendation = &AggregatedRecommendationResponse{
			Action:           cachedRec.Action,
			Confidence:       cachedRec.Confidence,
			SuggestedActions: suggestedActions,
			Summary:          cachedRec.Summary,
			KeyInsights:      cachedRec.KeyInsights,
		}
	}
	
	c.JSON(http.StatusOK, summary)
}

// GenerateAggregatedRecommendation handles POST /api/workflow/recommendations/aggregate
// Manually triggers generation of aggregated recommendation from the last 10 videos
func (h *WorkflowHandler) GenerateAggregatedRecommendation(c *gin.Context) {
	// Get all completed workflow executions
	allExecutions := h.store.GetAllWorkflowExecutions()
	
	// Filter to only completed executions
	allCompletedExecutions := make([]*models.WorkflowExecution, 0)
	for _, exec := range allExecutions {
		if exec.Status == models.WorkflowStatusCompleted {
			allCompletedExecutions = append(allCompletedExecutions, exec)
		}
	}
	
	if len(allCompletedExecutions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No completed workflow executions found. Process some videos first.",
		})
		return
	}
	
	// Generate aggregated recommendation
	aggregatedRec, err := h.generateAggregatedRecommendation(allCompletedExecutions)
	if err != nil {
		log.Printf("Failed to generate aggregated recommendation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to generate aggregated recommendation: %v", err),
		})
		return
	}
	
	// Return the generated recommendation
	c.JSON(http.StatusOK, aggregatedRec)
}

// generateAggregatedRecommendation creates an AI-powered consolidated recommendation from the most recent 10 completed workflow executions
func (h *WorkflowHandler) generateAggregatedRecommendation(executions []*models.WorkflowExecution) (*AggregatedRecommendationResponse, error) {
	if len(executions) == 0 {
		return nil, fmt.Errorf("no workflow executions provided")
	}
	
	// Sort by completed_at (newest first)
	sort.Slice(executions, func(i, j int) bool {
		if executions[i].CompletedAt == "" || executions[j].CompletedAt == "" {
			return false
		}
		return executions[i].CompletedAt > executions[j].CompletedAt
	})
	
	// Take the most recent 10
	limit := 10
	if len(executions) < limit {
		limit = len(executions)
	}
	recentExecutions := executions[:limit]
	
	// Build portfolio context
	portfolioContext := h.engine.BuildPortfolioContext()
	
	// Call engine to generate aggregated recommendation
	aggregatedRec, err := h.engine.GenerateAggregatedRecommendation(recentExecutions, portfolioContext)
	if err != nil {
		return nil, fmt.Errorf("failed to generate aggregated recommendation: %w", err)
	}
	
	// Get execution IDs for storage
	currentExecutionIDs := make([]string, len(recentExecutions))
	for i, exec := range recentExecutions {
		currentExecutionIDs[i] = exec.ID
	}
	
	// Store the aggregated recommendation with execution IDs
	recID := "latest" // Use a fixed ID so we always update the same record
	storedRec := &models.AggregatedRecommendation{
		ID:              recID,
		Action:          aggregatedRec.Action,
		Confidence:      aggregatedRec.Confidence,
		SuggestedActions: make([]models.SuggestedAction, len(aggregatedRec.SuggestedActions)),
		Summary:         aggregatedRec.Summary,
		KeyInsights:     aggregatedRec.KeyInsights,
		ExecutionIDs:   currentExecutionIDs,
	}
	
	for i, sa := range aggregatedRec.SuggestedActions {
		storedRec.SuggestedActions[i] = models.SuggestedAction{
			Type:      sa.Type,
			Symbol:    sa.Symbol,
			Rationale: sa.Rationale,
		}
	}
	
	if err := h.store.CreateOrUpdateAggregatedRecommendation(storedRec); err != nil {
		log.Printf("Failed to store aggregated recommendation: %v", err)
		// Continue anyway, we still return the recommendation
	}
	
	// Convert to response format
	suggestedActions := make([]SuggestedActionResponse, len(aggregatedRec.SuggestedActions))
	for i, sa := range aggregatedRec.SuggestedActions {
		suggestedActions[i] = SuggestedActionResponse{
			Type:      sa.Type,
			Symbol:    sa.Symbol,
			Rationale: sa.Rationale,
		}
	}
	
	return &AggregatedRecommendationResponse{
		Action:           aggregatedRec.Action,
		Confidence:       aggregatedRec.Confidence,
		SuggestedActions: suggestedActions,
		Summary:          aggregatedRec.Summary,
		KeyInsights:      aggregatedRec.KeyInsights,
	}, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}



