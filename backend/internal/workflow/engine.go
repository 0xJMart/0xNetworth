package workflow

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	workflowclient "0xnetworth/backend/internal/integrations/workflow"
	"0xnetworth/backend/internal/models"
	"0xnetworth/backend/internal/store"
)

// Engine orchestrates workflow executions
type Engine struct {
	store         store.Store
	workflowClient *workflowclient.Client
}

// NewEngine creates a new workflow engine
func NewEngine(store store.Store, workflowClient *workflowclient.Client) *Engine {
	return &Engine{
		store:          store,
		workflowClient: workflowClient,
	}
}

// ExecuteWorkflow processes a YouTube video through the agentic workflow
func (e *Engine) ExecuteWorkflow(videoURL string, sourceID string) (*models.WorkflowExecution, error) {
	// Extract video ID from URL to check for duplicates
	videoID := extractVideoIDFromURL(videoURL)
	
	// Check if this video has already been processed (globally, not just per-source)
	if videoID != "" {
		existingExecutions := e.store.GetWorkflowExecutionsByVideoID(videoID)
		// Check if any existing execution completed successfully
		for _, existing := range existingExecutions {
			if existing.Status == models.WorkflowStatusCompleted {
				log.Printf("Video %s has already been processed (execution %s). Skipping duplicate.", videoID, existing.ID)
				return existing, fmt.Errorf("video %s has already been processed", videoID)
			}
		}
	}
	
	// Create execution record
	executionID := uuid.New().String()
	execution := &models.WorkflowExecution{
		ID:        executionID,
		Status:    models.WorkflowStatusProcessing,
		VideoURL:  videoURL,
		SourceID:  sourceID,
		StartedAt: time.Now().UTC().Format(time.RFC3339),
	}
	e.store.CreateOrUpdateWorkflowExecution(execution)

	log.Printf("Starting workflow execution %s for video: %s", executionID, videoURL)

	// Build portfolio context from current investments
	portfolioContext := e.buildPortfolioContext()

	// Call Python workflow service
	request := workflowclient.WorkflowRequest{
		YoutubeURL:       videoURL,
		PortfolioContext: portfolioContext,
	}

	response, err := e.workflowClient.ProcessVideo(request)
	if err != nil {
		execution.Status = models.WorkflowStatusFailed
		execution.Error = err.Error()
		execution.CompletedAt = time.Now().UTC().Format(time.RFC3339)
		e.store.CreateOrUpdateWorkflowExecution(execution)
		return execution, fmt.Errorf("workflow service error: %w", err)
	}

	// Store transcript
	transcriptID := uuid.New().String()
	transcript := &models.VideoTranscript{
		ID:          transcriptID,
		VideoID:     response.Transcript.VideoID,
		VideoTitle:  response.Transcript.VideoTitle,
		VideoURL:    videoURL,
		Text:        response.Transcript.Text,
		Duration:    response.Transcript.Duration,
		SourceID:    sourceID,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	e.store.CreateOrUpdateTranscript(transcript)
	execution.TranscriptID = transcriptID
	execution.VideoID = response.Transcript.VideoID
	execution.VideoTitle = response.Transcript.VideoTitle

	// Store market analysis
	analysisID := uuid.New().String()
	analysis := &models.MarketAnalysis{
		ID:           analysisID,
		TranscriptID: transcriptID,
		Conditions:   response.MarketAnalysis.Conditions,
		Trends:       response.MarketAnalysis.Trends,
		RiskFactors:  response.MarketAnalysis.RiskFactors,
		Summary:      response.MarketAnalysis.Summary,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}
	e.store.CreateOrUpdateMarketAnalysis(analysis)
	execution.AnalysisID = analysisID

	// Store recommendation
	recommendationID := uuid.New().String()
	suggestedActions := make([]models.SuggestedAction, len(response.Recommendation.SuggestedActions))
	for i, sa := range response.Recommendation.SuggestedActions {
		suggestedActions[i] = models.SuggestedAction{
			Type:      sa.Type,
			Symbol:    sa.Symbol,
			Rationale: sa.Rationale,
		}
	}
	recommendation := &models.Recommendation{
		ID:              recommendationID,
		AnalysisID:      analysisID,
		Action:          response.Recommendation.Action,
		Confidence:      response.Recommendation.Confidence,
		SuggestedActions: suggestedActions,
		Summary:        response.Recommendation.Summary,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	}
	e.store.CreateOrUpdateRecommendation(recommendation)
	execution.RecommendationID = recommendationID

	// Mark execution as completed
	execution.Status = models.WorkflowStatusCompleted
	execution.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	e.store.CreateOrUpdateWorkflowExecution(execution)

	log.Printf("Workflow execution %s completed successfully", executionID)

	return execution, nil
}

// buildPortfolioContext builds portfolio context from current investments
func (e *Engine) buildPortfolioContext() *workflowclient.PortfolioContext {
	investments := e.store.GetAllInvestments()
	
	if len(investments) == 0 {
		return nil
	}

	holdings := make([]workflowclient.Holding, 0, len(investments))
	totalValue := 0.0

	for _, inv := range investments {
		holdings = append(holdings, workflowclient.Holding{
			Symbol:   inv.Symbol,
			Quantity: inv.Quantity,
			Value:    inv.Value,
		})
		totalValue += inv.Value
	}

	return &workflowclient.PortfolioContext{
		Holdings:   holdings,
		TotalValue: totalValue,
	}
}

// extractVideoIDFromURL extracts the video ID from various YouTube URL formats
func extractVideoIDFromURL(url string) string {
	if url == "" {
		return ""
	}
	
	// Patterns for YouTube URLs
	patterns := []string{
		`(?:youtube\.com/watch\?v=|youtu\.be/|youtube\.com/embed/)([a-zA-Z0-9_-]{11})`,
		`youtube\.com/watch\?.*v=([a-zA-Z0-9_-]{11})`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	
	// If no pattern matches, check if the input is already a video ID (11 characters)
	url = strings.TrimSpace(url)
	if len(url) == 11 && strings.ReplaceAll(strings.ReplaceAll(url, "-", ""), "_", "") != "" {
		// Check if it's alphanumeric (allowing - and _)
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{11}$`, url); matched {
			return url
		}
	}
	
	return ""
}

