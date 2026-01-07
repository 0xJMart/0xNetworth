package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"

	"0xnetworth/backend/internal/models"
	"0xnetworth/backend/internal/store"
	"0xnetworth/backend/internal/workflow"
)

// WorkflowHandler handles workflow-related HTTP requests
type WorkflowHandler struct {
	store    *store.Store
	engine   *workflow.Engine
	scheduler *workflow.Scheduler
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(store *store.Store, engine *workflow.Engine, scheduler *workflow.Scheduler) *WorkflowHandler {
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

	c.JSON(http.StatusOK, source)
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


