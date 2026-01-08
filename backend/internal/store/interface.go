package store

import (
	"time"

	"0xnetworth/backend/internal/models"
)

// Store defines the interface for data storage operations
type Store interface {
	// Portfolio operations
	GetAllPortfolios() []*models.Portfolio
	GetPortfoliosByPlatform(platform models.Platform) []*models.Portfolio
	GetPortfolioByID(id string) (*models.Portfolio, bool)
	CreateOrUpdatePortfolio(portfolio *models.Portfolio)
	DeletePortfolio(id string) bool

	// Investment operations
	GetAllInvestments() []*models.Investment
	GetInvestmentsByAccount(accountID string) []*models.Investment
	GetInvestmentsByPlatform(platform models.Platform) []*models.Investment
	CreateOrUpdateInvestment(investment *models.Investment)
	DeleteInvestment(id string) bool

	// NetWorth operations
	GetNetWorth() *models.NetWorth
	UpdateNetWorth(networth *models.NetWorth)
	RecalculateNetWorth() *models.NetWorth

	// Sync metadata operations
	GetLastSyncTime() time.Time
	SetLastSyncTime(t time.Time)

	// YouTube Source operations
	GetAllYouTubeSources() []*models.YouTubeSource
	GetYouTubeSourceByID(id string) (*models.YouTubeSource, bool)
	CreateOrUpdateYouTubeSource(source *models.YouTubeSource)
	DeleteYouTubeSource(id string) bool

	// Video Transcript operations
	CreateOrUpdateTranscript(transcript *models.VideoTranscript)
	GetTranscriptByID(id string) (*models.VideoTranscript, bool)
	GetTranscriptsByVideoID(videoID string) []*models.VideoTranscript

	// Market Analysis operations
	CreateOrUpdateMarketAnalysis(analysis *models.MarketAnalysis)
	GetMarketAnalysisByID(id string) (*models.MarketAnalysis, bool)
	GetMarketAnalysesByTranscriptID(transcriptID string) []*models.MarketAnalysis

	// Recommendation operations
	CreateOrUpdateRecommendation(recommendation *models.Recommendation)
	GetRecommendationByID(id string) (*models.Recommendation, bool)
	GetRecommendationsByAnalysisID(analysisID string) []*models.Recommendation

	// Workflow Execution operations
	CreateOrUpdateWorkflowExecution(execution *models.WorkflowExecution)
	GetWorkflowExecutionByID(id string) (*models.WorkflowExecution, bool)
	GetAllWorkflowExecutions() []*models.WorkflowExecution
	GetWorkflowExecutionsBySourceID(sourceID string) []*models.WorkflowExecution
	GetWorkflowExecutionsByVideoID(videoID string) []*models.WorkflowExecution
}

