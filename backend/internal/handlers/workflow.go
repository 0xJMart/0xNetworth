package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"

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
	AggregatedSummary    string             `json:"aggregated_summary"` // Summary from most recent 10 videos
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
	
	// Generate aggregated summary from most recent 10 videos
	summary.AggregatedSummary = h.generateAggregatedSummary(allCompletedExecutions)
	
	c.JSON(http.StatusOK, summary)
}

// generateAggregatedSummary creates a single consolidated summary from the most recent 10 completed workflow executions
func (h *WorkflowHandler) generateAggregatedSummary(executions []*models.WorkflowExecution) string {
	if len(executions) == 0 {
		return "No workflow executions available for analysis."
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
	
	// Collect data from market analyses and recommendations
	conditions := make(map[string]int)
	allTrends := make([]string, 0)
	allRiskFactors := make([]string, 0)
	allSuggestedActions := make(map[string][]string) // action type -> list of symbols/rationales
	videoCount := 0
	
	for _, exec := range recentExecutions {
		if exec.AnalysisID == "" {
			continue
		}
		
		analysis, exists := h.store.GetMarketAnalysisByID(exec.AnalysisID)
		if !exists {
			continue
		}
		
		videoCount++
		
		// Track conditions
		if analysis.Conditions != "" {
			conditions[strings.ToLower(analysis.Conditions)]++
		}
		
		// Collect trends
		allTrends = append(allTrends, analysis.Trends...)
		
		// Collect risk factors
		allRiskFactors = append(allRiskFactors, analysis.RiskFactors...)
		
		// Get recommendations and suggested actions
		if exec.RecommendationID != "" {
			rec, exists := h.store.GetRecommendationByID(exec.RecommendationID)
			if exists {
				for _, action := range rec.SuggestedActions {
					key := fmt.Sprintf("%s:%s", action.Type, action.Symbol)
					if _, ok := allSuggestedActions[key]; !ok {
						allSuggestedActions[key] = make([]string, 0)
					}
					if action.Rationale != "" {
						allSuggestedActions[key] = append(allSuggestedActions[key], action.Rationale)
					}
				}
			}
		}
	}
	
	if videoCount == 0 {
		return fmt.Sprintf("Based on the most recent %d video(s) analyzed, no detailed market analysis data is available yet.", limit)
	}
	
	// Build consolidated summary
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Based on analysis of the most recent %d video(s), here's a consolidated market outlook:\n\n", videoCount))
	
	// Determine overall market condition
	overallCondition := "neutral"
	maxConditionCount := 0
	for cond, count := range conditions {
		if count > maxConditionCount {
			maxConditionCount = count
			overallCondition = cond
		}
	}
	
	summary.WriteString(fmt.Sprintf("**Market Sentiment:** %s (based on %d of %d analyses)\n\n", 
		strings.Title(overallCondition), maxConditionCount, videoCount))
	
	// Aggregate trends (remove duplicates, show most common)
	trendCounts := make(map[string]int)
	for _, trend := range allTrends {
		if trend != "" {
			trendCounts[trend]++
		}
	}
	if len(trendCounts) > 0 {
		summary.WriteString("**Key Trends:**\n")
		// Sort by frequency and take top 5
		type trendItem struct {
			trend string
			count int
		}
		trendList := make([]trendItem, 0, len(trendCounts))
		for trend, count := range trendCounts {
			trendList = append(trendList, trendItem{trend, count})
		}
		sort.Slice(trendList, func(i, j int) bool {
			return trendList[i].count > trendList[j].count
		})
		for i, item := range trendList {
			if i >= 5 {
				break
			}
			summary.WriteString(fmt.Sprintf("• %s\n", item.trend))
		}
		summary.WriteString("\n")
	}
	
	// Aggregate risk factors (remove duplicates, show most common)
	riskCounts := make(map[string]int)
	for _, risk := range allRiskFactors {
		if risk != "" {
			riskCounts[risk]++
		}
	}
	if len(riskCounts) > 0 {
		summary.WriteString("**Risk Factors:**\n")
		// Sort by frequency and take top 5
		type riskItem struct {
			risk  string
			count int
		}
		riskList := make([]riskItem, 0, len(riskCounts))
		for risk, count := range riskCounts {
			riskList = append(riskList, riskItem{risk, count})
		}
		sort.Slice(riskList, func(i, j int) bool {
			return riskList[i].count > riskList[j].count
		})
		for i, item := range riskList {
			if i >= 5 {
				break
			}
			summary.WriteString(fmt.Sprintf("• %s\n", item.risk))
		}
		summary.WriteString("\n")
	}
	
	// Aggregate actionable recommendations
	if len(allSuggestedActions) > 0 {
		summary.WriteString("**Actionable Recommendations:**\n")
		actionCount := 0
		for key, rationales := range allSuggestedActions {
			if actionCount >= 5 {
				break
			}
			parts := strings.Split(key, ":")
			if len(parts) == 2 {
				actionType := strings.Title(parts[0])
				symbol := strings.ToUpper(parts[1])
				summary.WriteString(fmt.Sprintf("• **%s %s**", actionType, symbol))
				if len(rationales) > 0 && rationales[0] != "" {
					summary.WriteString(fmt.Sprintf(": %s", rationales[0]))
				}
				summary.WriteString("\n")
				actionCount++
			}
		}
	}
	
	return summary.String()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


