package store

import (
	"sync"
	"time"

	"0xnetworth/backend/internal/models"
)

// MemoryStore is an in-memory store for investment data
type MemoryStore struct {
	mu              sync.RWMutex
	portfolios      map[string]*models.Portfolio
	investments     map[string]*models.Investment
	networth        *models.NetWorth
	lastSync        time.Time
	youtubeSources  map[string]*models.YouTubeSource
	transcripts     map[string]*models.VideoTranscript
	marketAnalyses  map[string]*models.MarketAnalysis
	recommendations map[string]*models.Recommendation
	executions      map[string]*models.WorkflowExecution
}

// NewStore creates a new in-memory store
func NewStore() Store {
	return &MemoryStore{
		portfolios:      make(map[string]*models.Portfolio),
		investments:     make(map[string]*models.Investment),
		networth:        &models.NetWorth{},
		youtubeSources:  make(map[string]*models.YouTubeSource),
		transcripts:     make(map[string]*models.VideoTranscript),
		marketAnalyses:  make(map[string]*models.MarketAnalysis),
		recommendations: make(map[string]*models.Recommendation),
		executions:      make(map[string]*models.WorkflowExecution),
	}
}

// Portfolio operations

// GetAllPortfolios returns all portfolios
func (s *MemoryStore) GetAllPortfolios() []*models.Portfolio {
	s.mu.RLock()
	defer s.mu.RUnlock()

	portfolios := make([]*models.Portfolio, 0, len(s.portfolios))
	for _, p := range s.portfolios {
		portfolios = append(portfolios, p)
	}
	return portfolios
}

// GetPortfoliosByPlatform returns portfolios for a specific platform
func (s *MemoryStore) GetPortfoliosByPlatform(platform models.Platform) []*models.Portfolio {
	s.mu.RLock()
	defer s.mu.RUnlock()

	portfolios := make([]*models.Portfolio, 0)
	for _, p := range s.portfolios {
		if p.Platform == platform {
			portfolios = append(portfolios, p)
		}
	}
	return portfolios
}

// GetPortfolioByID returns a portfolio by ID
func (s *MemoryStore) GetPortfolioByID(id string) (*models.Portfolio, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	portfolio, exists := s.portfolios[id]
	return portfolio, exists
}

// CreateOrUpdatePortfolio creates or updates a portfolio
func (s *MemoryStore) CreateOrUpdatePortfolio(portfolio *models.Portfolio) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.portfolios[portfolio.ID] = portfolio
}

// DeletePortfolio deletes a portfolio by ID
func (s *MemoryStore) DeletePortfolio(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.portfolios[id]; !exists {
		return false
	}
	delete(s.portfolios, id)
	return true
}

// Investment operations

// GetAllInvestments returns all investments
func (s *MemoryStore) GetAllInvestments() []*models.Investment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	investments := make([]*models.Investment, 0, len(s.investments))
	for _, inv := range s.investments {
		investments = append(investments, inv)
	}
	return investments
}

// GetInvestmentsByAccount returns investments for a specific account
func (s *MemoryStore) GetInvestmentsByAccount(accountID string) []*models.Investment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	investments := make([]*models.Investment, 0)
	for _, inv := range s.investments {
		if inv.AccountID == accountID {
			investments = append(investments, inv)
		}
	}
	return investments
}

// GetInvestmentsByPlatform returns investments for a specific platform
func (s *MemoryStore) GetInvestmentsByPlatform(platform models.Platform) []*models.Investment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	investments := make([]*models.Investment, 0)
	for _, inv := range s.investments {
		if inv.Platform == platform {
			investments = append(investments, inv)
		}
	}
	return investments
}

// CreateOrUpdateInvestment creates or updates an investment
func (s *MemoryStore) CreateOrUpdateInvestment(investment *models.Investment) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.investments[investment.ID] = investment
}

// DeleteInvestment deletes an investment by ID
func (s *MemoryStore) DeleteInvestment(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.investments[id]; !exists {
		return false
	}
	delete(s.investments, id)
	return true
}

// NetWorth operations

// GetNetWorth returns the current net worth
func (s *MemoryStore) GetNetWorth() *models.NetWorth {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid race conditions
	networth := *s.networth
	return &networth
}

// UpdateNetWorth updates the net worth calculation
func (s *MemoryStore) UpdateNetWorth(networth *models.NetWorth) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.networth = networth
}

// RecalculateNetWorth recalculates net worth from current accounts and investments
func (s *MemoryStore) RecalculateNetWorth() *models.NetWorth {
	s.mu.Lock()
	defer s.mu.Unlock()

	networth := &models.NetWorth{
		ByPlatform:   make(map[models.Platform]float64),
		ByAssetType:  make(map[string]float64),
		Currency:     "USD", // Default currency
		LastCalculated: time.Now().UTC().Format(time.RFC3339),
	}

	// Calculate total from investments (portfolios don't have balances, only holdings)
	totalValue := 0.0
	for _, investment := range s.investments {
		totalValue += investment.Value
		networth.ByPlatform[investment.Platform] += investment.Value
		networth.ByAssetType[investment.AssetType] += investment.Value
	}

	networth.TotalValue = totalValue
	networth.AccountCount = len(s.portfolios) // Use portfolio count instead of account count
	s.networth = networth
	return networth
}

// GetLastSyncTime returns the last sync time
func (s *MemoryStore) GetLastSyncTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.lastSync
}

// SetLastSyncTime sets the last sync time
func (s *MemoryStore) SetLastSyncTime(t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastSync = t
}

// YouTube Source operations

// GetAllYouTubeSources returns all YouTube sources
func (s *MemoryStore) GetAllYouTubeSources() []*models.YouTubeSource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sources := make([]*models.YouTubeSource, 0, len(s.youtubeSources))
	for _, src := range s.youtubeSources {
		sources = append(sources, src)
	}
	return sources
}

// GetYouTubeSourceByID returns a YouTube source by ID
func (s *MemoryStore) GetYouTubeSourceByID(id string) (*models.YouTubeSource, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	source, exists := s.youtubeSources[id]
	return source, exists
}

// CreateOrUpdateYouTubeSource creates or updates a YouTube source
func (s *MemoryStore) CreateOrUpdateYouTubeSource(source *models.YouTubeSource) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.youtubeSources[source.ID] = source
}

// DeleteYouTubeSource deletes a YouTube source by ID
func (s *MemoryStore) DeleteYouTubeSource(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.youtubeSources[id]; !exists {
		return false
	}
	delete(s.youtubeSources, id)
	return true
}

// Video Transcript operations

// CreateOrUpdateTranscript creates or updates a video transcript
func (s *MemoryStore) CreateOrUpdateTranscript(transcript *models.VideoTranscript) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.transcripts[transcript.ID] = transcript
}

// GetTranscriptByID returns a transcript by ID
func (s *MemoryStore) GetTranscriptByID(id string) (*models.VideoTranscript, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	transcript, exists := s.transcripts[id]
	return transcript, exists
}

// GetTranscriptsByVideoID returns transcripts for a specific video ID
func (s *MemoryStore) GetTranscriptsByVideoID(videoID string) []*models.VideoTranscript {
	s.mu.RLock()
	defer s.mu.RUnlock()

	transcripts := make([]*models.VideoTranscript, 0)
	for _, t := range s.transcripts {
		if t.VideoID == videoID {
			transcripts = append(transcripts, t)
		}
	}
	return transcripts
}

// Market Analysis operations

// CreateOrUpdateMarketAnalysis creates or updates a market analysis
func (s *MemoryStore) CreateOrUpdateMarketAnalysis(analysis *models.MarketAnalysis) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.marketAnalyses[analysis.ID] = analysis
}

// GetMarketAnalysisByID returns a market analysis by ID
func (s *MemoryStore) GetMarketAnalysisByID(id string) (*models.MarketAnalysis, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	analysis, exists := s.marketAnalyses[id]
	return analysis, exists
}

// GetMarketAnalysesByTranscriptID returns market analyses for a specific transcript ID
func (s *MemoryStore) GetMarketAnalysesByTranscriptID(transcriptID string) []*models.MarketAnalysis {
	s.mu.RLock()
	defer s.mu.RUnlock()

	analyses := make([]*models.MarketAnalysis, 0)
	for _, a := range s.marketAnalyses {
		if a.TranscriptID == transcriptID {
			analyses = append(analyses, a)
		}
	}
	return analyses
}

// Recommendation operations

// CreateOrUpdateRecommendation creates or updates a recommendation
func (s *MemoryStore) CreateOrUpdateRecommendation(recommendation *models.Recommendation) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.recommendations[recommendation.ID] = recommendation
}

// GetRecommendationByID returns a recommendation by ID
func (s *MemoryStore) GetRecommendationByID(id string) (*models.Recommendation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	recommendation, exists := s.recommendations[id]
	return recommendation, exists
}

// GetRecommendationsByAnalysisID returns recommendations for a specific analysis ID
func (s *MemoryStore) GetRecommendationsByAnalysisID(analysisID string) []*models.Recommendation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	recommendations := make([]*models.Recommendation, 0)
	for _, r := range s.recommendations {
		if r.AnalysisID == analysisID {
			recommendations = append(recommendations, r)
		}
	}
	return recommendations
}

// Workflow Execution operations

// CreateOrUpdateWorkflowExecution creates or updates a workflow execution
func (s *MemoryStore) CreateOrUpdateWorkflowExecution(execution *models.WorkflowExecution) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.executions[execution.ID] = execution
}

// GetWorkflowExecutionByID returns a workflow execution by ID
func (s *MemoryStore) GetWorkflowExecutionByID(id string) (*models.WorkflowExecution, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	execution, exists := s.executions[id]
	return execution, exists
}

// GetAllWorkflowExecutions returns all workflow executions
func (s *MemoryStore) GetAllWorkflowExecutions() []*models.WorkflowExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()

	executions := make([]*models.WorkflowExecution, 0, len(s.executions))
	for _, e := range s.executions {
		executions = append(executions, e)
	}
	return executions
}

// GetWorkflowExecutionsBySourceID returns workflow executions for a specific source ID
func (s *MemoryStore) GetWorkflowExecutionsBySourceID(sourceID string) []*models.WorkflowExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()

	executions := make([]*models.WorkflowExecution, 0)
	for _, e := range s.executions {
		if e.SourceID == sourceID {
			executions = append(executions, e)
		}
	}
	return executions
}

// GetWorkflowExecutionsByVideoID returns workflow executions for a specific video ID
func (s *MemoryStore) GetWorkflowExecutionsByVideoID(videoID string) []*models.WorkflowExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()

	executions := make([]*models.WorkflowExecution, 0)
	for _, e := range s.executions {
		if e.VideoID == videoID {
			executions = append(executions, e)
		}
	}
	return executions
}

// GetLatestAggregatedRecommendation returns the most recent aggregated recommendation
func (s *MemoryStore) GetLatestAggregatedRecommendation() (*models.AggregatedRecommendation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Memory store doesn't persist aggregated recommendations
	// Return false to indicate not found
	return nil, false
}

// CreateOrUpdateAggregatedRecommendation creates or updates an aggregated recommendation
func (s *MemoryStore) CreateOrUpdateAggregatedRecommendation(rec *models.AggregatedRecommendation) error {
	// Memory store doesn't persist aggregated recommendations
	// This is a no-op for in-memory store
	return nil
}

