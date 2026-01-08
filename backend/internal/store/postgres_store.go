package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"0xnetworth/backend/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore is a PostgreSQL-backed store implementation
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore creates a new PostgreSQL store
func NewPostgresStore(connString string) (*PostgresStore, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStore{pool: pool}, nil
}

// Close closes the database connection pool
func (s *PostgresStore) Close() {
	s.pool.Close()
}

// InitSchema executes the schema SQL to create tables
func (s *PostgresStore) InitSchema(schemaSQL string) error {
	_, err := s.pool.Exec(context.Background(), schemaSQL)
	return err
}

// Helper functions for timestamp conversion
func parseTimestamp(ts sql.NullTime) string {
	if ts.Valid {
		return ts.Time.UTC().Format(time.RFC3339)
	}
	return ""
}

func parseTimestampPtr(ts sql.NullTime) *string {
	if ts.Valid {
		formatted := ts.Time.UTC().Format(time.RFC3339)
		return &formatted
	}
	return nil
}

func parseIntPtr(i sql.NullInt64) *int {
	if i.Valid {
		val := int(i.Int64)
		return &val
	}
	return nil
}

// Portfolio operations

// GetAllPortfolios returns all portfolios
func (s *PostgresStore) GetAllPortfolios() []*models.Portfolio {
	rows, err := s.pool.Query(context.Background(),
		"SELECT id, platform, name, type, last_synced, created_at, updated_at FROM portfolios ORDER BY created_at DESC")
	if err != nil {
		return []*models.Portfolio{}
	}
	defer rows.Close()

	var portfolios []*models.Portfolio
	for rows.Next() {
		var p models.Portfolio
		var lastSynced, createdAt, updatedAt sql.NullTime
		var portfolioType sql.NullString

		err := rows.Scan(&p.ID, &p.Platform, &p.Name, &portfolioType, &lastSynced, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		if portfolioType.Valid {
			p.Type = portfolioType.String
		}
		p.LastSynced = parseTimestamp(lastSynced)

		portfolios = append(portfolios, &p)
	}

	return portfolios
}

// GetPortfoliosByPlatform returns portfolios for a specific platform
func (s *PostgresStore) GetPortfoliosByPlatform(platform models.Platform) []*models.Portfolio {
	rows, err := s.pool.Query(context.Background(),
		"SELECT id, platform, name, type, last_synced, created_at, updated_at FROM portfolios WHERE platform = $1 ORDER BY created_at DESC",
		platform)
	if err != nil {
		return []*models.Portfolio{}
	}
	defer rows.Close()

	var portfolios []*models.Portfolio
	for rows.Next() {
		var p models.Portfolio
		var lastSynced, createdAt, updatedAt sql.NullTime
		var portfolioType sql.NullString

		err := rows.Scan(&p.ID, &p.Platform, &p.Name, &portfolioType, &lastSynced, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		if portfolioType.Valid {
			p.Type = portfolioType.String
		}
		p.LastSynced = parseTimestamp(lastSynced)

		portfolios = append(portfolios, &p)
	}

	return portfolios
}

// GetPortfolioByID returns a portfolio by ID
func (s *PostgresStore) GetPortfolioByID(id string) (*models.Portfolio, bool) {
	var p models.Portfolio
	var lastSynced, createdAt, updatedAt sql.NullTime
	var portfolioType sql.NullString

	err := s.pool.QueryRow(context.Background(),
		"SELECT id, platform, name, type, last_synced, created_at, updated_at FROM portfolios WHERE id = $1",
		id).Scan(&p.ID, &p.Platform, &p.Name, &portfolioType, &lastSynced, &createdAt, &updatedAt)

	if err != nil {
		return nil, false
	}

	if portfolioType.Valid {
		p.Type = portfolioType.String
	}
	p.LastSynced = parseTimestamp(lastSynced)

	return &p, true
}

// CreateOrUpdatePortfolio creates or updates a portfolio
func (s *PostgresStore) CreateOrUpdatePortfolio(portfolio *models.Portfolio) {
	var lastSynced interface{}
	if portfolio.LastSynced != "" {
		t, err := time.Parse(time.RFC3339, portfolio.LastSynced)
		if err == nil {
			lastSynced = t
		}
	}

	_, err := s.pool.Exec(context.Background(),
		`INSERT INTO portfolios (id, platform, name, type, last_synced, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		 ON CONFLICT (id) DO UPDATE SET
		 platform = EXCLUDED.platform,
		 name = EXCLUDED.name,
		 type = EXCLUDED.type,
		 last_synced = EXCLUDED.last_synced,
		 updated_at = CURRENT_TIMESTAMP`,
		portfolio.ID, portfolio.Platform, portfolio.Name, portfolio.Type, lastSynced)

	if err != nil {
		// Log error but don't fail silently - could add logging here
		_ = err
	}
}

// DeletePortfolio deletes a portfolio by ID
func (s *PostgresStore) DeletePortfolio(id string) bool {
	result, err := s.pool.Exec(context.Background(), "DELETE FROM portfolios WHERE id = $1", id)
	if err != nil {
		return false
	}
	return result.RowsAffected() > 0
}

// Investment operations

// GetAllInvestments returns all investments
func (s *PostgresStore) GetAllInvestments() []*models.Investment {
	rows, err := s.pool.Query(context.Background(),
		"SELECT id, account_id, platform, symbol, name, quantity, value, price, currency, asset_type, last_updated, created_at, updated_at FROM investments ORDER BY created_at DESC")
	if err != nil {
		return []*models.Investment{}
	}
	defer rows.Close()

	var investments []*models.Investment
	for rows.Next() {
		var inv models.Investment
		var lastUpdated, createdAt, updatedAt sql.NullTime
		var name, assetType sql.NullString

		err := rows.Scan(&inv.ID, &inv.AccountID, &inv.Platform, &inv.Symbol, &name, &inv.Quantity, &inv.Value, &inv.Price, &inv.Currency, &assetType, &lastUpdated, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		if name.Valid {
			inv.Name = name.String
		}
		if assetType.Valid {
			inv.AssetType = assetType.String
		}
		inv.LastUpdated = parseTimestamp(lastUpdated)

		investments = append(investments, &inv)
	}

	return investments
}

// GetInvestmentsByAccount returns investments for a specific account
func (s *PostgresStore) GetInvestmentsByAccount(accountID string) []*models.Investment {
	rows, err := s.pool.Query(context.Background(),
		"SELECT id, account_id, platform, symbol, name, quantity, value, price, currency, asset_type, last_updated, created_at, updated_at FROM investments WHERE account_id = $1 ORDER BY created_at DESC",
		accountID)
	if err != nil {
		return []*models.Investment{}
	}
	defer rows.Close()

	var investments []*models.Investment
	for rows.Next() {
		var inv models.Investment
		var lastUpdated, createdAt, updatedAt sql.NullTime
		var name, assetType sql.NullString

		err := rows.Scan(&inv.ID, &inv.AccountID, &inv.Platform, &inv.Symbol, &name, &inv.Quantity, &inv.Value, &inv.Price, &inv.Currency, &assetType, &lastUpdated, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		if name.Valid {
			inv.Name = name.String
		}
		if assetType.Valid {
			inv.AssetType = assetType.String
		}
		inv.LastUpdated = parseTimestamp(lastUpdated)

		investments = append(investments, &inv)
	}

	return investments
}

// GetInvestmentsByPlatform returns investments for a specific platform
func (s *PostgresStore) GetInvestmentsByPlatform(platform models.Platform) []*models.Investment {
	rows, err := s.pool.Query(context.Background(),
		"SELECT id, account_id, platform, symbol, name, quantity, value, price, currency, asset_type, last_updated, created_at, updated_at FROM investments WHERE platform = $1 ORDER BY created_at DESC",
		platform)
	if err != nil {
		return []*models.Investment{}
	}
	defer rows.Close()

	var investments []*models.Investment
	for rows.Next() {
		var inv models.Investment
		var lastUpdated, createdAt, updatedAt sql.NullTime
		var name, assetType sql.NullString

		err := rows.Scan(&inv.ID, &inv.AccountID, &inv.Platform, &inv.Symbol, &name, &inv.Quantity, &inv.Value, &inv.Price, &inv.Currency, &assetType, &lastUpdated, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		if name.Valid {
			inv.Name = name.String
		}
		if assetType.Valid {
			inv.AssetType = assetType.String
		}
		inv.LastUpdated = parseTimestamp(lastUpdated)

		investments = append(investments, &inv)
	}

	return investments
}

// CreateOrUpdateInvestment creates or updates an investment
func (s *PostgresStore) CreateOrUpdateInvestment(investment *models.Investment) {
	var lastUpdated interface{}
	if investment.LastUpdated != "" {
		t, err := time.Parse(time.RFC3339, investment.LastUpdated)
		if err == nil {
			lastUpdated = t
		}
	}

	_, err := s.pool.Exec(context.Background(),
		`INSERT INTO investments (id, account_id, platform, symbol, name, quantity, value, price, currency, asset_type, last_updated, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		 ON CONFLICT (id) DO UPDATE SET
		 account_id = EXCLUDED.account_id,
		 platform = EXCLUDED.platform,
		 symbol = EXCLUDED.symbol,
		 name = EXCLUDED.name,
		 quantity = EXCLUDED.quantity,
		 value = EXCLUDED.value,
		 price = EXCLUDED.price,
		 currency = EXCLUDED.currency,
		 asset_type = EXCLUDED.asset_type,
		 last_updated = EXCLUDED.last_updated,
		 updated_at = CURRENT_TIMESTAMP`,
		investment.ID, investment.AccountID, investment.Platform, investment.Symbol, investment.Name,
		investment.Quantity, investment.Value, investment.Price, investment.Currency, investment.AssetType, lastUpdated)

	if err != nil {
		_ = err
	}
}

// DeleteInvestment deletes an investment by ID
func (s *PostgresStore) DeleteInvestment(id string) bool {
	result, err := s.pool.Exec(context.Background(), "DELETE FROM investments WHERE id = $1", id)
	if err != nil {
		return false
	}
	return result.RowsAffected() > 0
}

// NetWorth operations

// GetNetWorth returns the current net worth (calculated on the fly)
func (s *PostgresStore) GetNetWorth() *models.NetWorth {
	return s.RecalculateNetWorth()
}

// UpdateNetWorth updates the net worth calculation (no-op for PostgresStore, always recalculates)
func (s *PostgresStore) UpdateNetWorth(networth *models.NetWorth) {
	// No-op: NetWorth is always calculated from investments
}

// RecalculateNetWorth recalculates net worth from current accounts and investments
func (s *PostgresStore) RecalculateNetWorth() *models.NetWorth {
	networth := &models.NetWorth{
		ByPlatform:    make(map[models.Platform]float64),
		ByAssetType:    make(map[string]float64),
		Currency:       "USD",
		LastCalculated: time.Now().UTC().Format(time.RFC3339),
	}

	// Get total value and breakdowns by platform and asset type
	rows, err := s.pool.Query(context.Background(),
		`SELECT platform, asset_type, SUM(value) as total_value
		 FROM investments
		 GROUP BY platform, asset_type`)
	if err != nil {
		return networth
	}
	defer rows.Close()

	var totalValue float64
	for rows.Next() {
		var platform models.Platform
		var assetType sql.NullString
		var value float64

		err := rows.Scan(&platform, &assetType, &value)
		if err != nil {
			continue
		}

		totalValue += value
		networth.ByPlatform[platform] += value

		if assetType.Valid {
			networth.ByAssetType[assetType.String] += value
		}
	}

	networth.TotalValue = totalValue

	// Get portfolio count
	var count int
	err = s.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM portfolios").Scan(&count)
	if err == nil {
		networth.AccountCount = count
	}

	return networth
}

// Sync metadata operations

// GetLastSyncTime returns the last sync time
func (s *PostgresStore) GetLastSyncTime() time.Time {
	var lastSync sql.NullTime
	err := s.pool.QueryRow(context.Background(),
		"SELECT last_sync_time FROM sync_metadata WHERE platform = $1 ORDER BY updated_at DESC LIMIT 1",
		models.PlatformCoinbase).Scan(&lastSync)

	if err != nil || !lastSync.Valid {
		return time.Time{}
	}

	return lastSync.Time
}

// SetLastSyncTime sets the last sync time
func (s *PostgresStore) SetLastSyncTime(t time.Time) {
	_, err := s.pool.Exec(context.Background(),
		`INSERT INTO sync_metadata (id, platform, last_sync_time, sync_status, created_at, updated_at)
		 VALUES ($1, $2, $3, 'success', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		 ON CONFLICT (platform) DO UPDATE SET
		 last_sync_time = EXCLUDED.last_sync_time,
		 sync_status = 'success',
		 updated_at = CURRENT_TIMESTAMP`,
		fmt.Sprintf("sync-%s", models.PlatformCoinbase), models.PlatformCoinbase, t)

	if err != nil {
		_ = err
	}
}

// YouTube Source operations

// GetAllYouTubeSources returns all YouTube sources
func (s *PostgresStore) GetAllYouTubeSources() []*models.YouTubeSource {
	rows, err := s.pool.Query(context.Background(),
		"SELECT id, type, url, name, channel_id, playlist_id, enabled, schedule, last_processed, created_at, updated_at FROM youtube_sources ORDER BY created_at DESC")
	if err != nil {
		return []*models.YouTubeSource{}
	}
	defer rows.Close()

	var sources []*models.YouTubeSource
	for rows.Next() {
		var src models.YouTubeSource
		var channelID, playlistID, schedule sql.NullString
		var lastProcessed, createdAt, updatedAt sql.NullTime

		err := rows.Scan(&src.ID, &src.Type, &src.URL, &src.Name, &channelID, &playlistID, &src.Enabled, &schedule, &lastProcessed, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		if channelID.Valid {
			src.ChannelID = channelID.String
		}
		if playlistID.Valid {
			src.PlaylistID = playlistID.String
		}
		if schedule.Valid {
			src.Schedule = schedule.String
		}
		src.LastProcessed = parseTimestamp(lastProcessed)

		sources = append(sources, &src)
	}

	return sources
}

// GetYouTubeSourceByID returns a YouTube source by ID
func (s *PostgresStore) GetYouTubeSourceByID(id string) (*models.YouTubeSource, bool) {
	var src models.YouTubeSource
	var channelID, playlistID, schedule sql.NullString
	var lastProcessed, createdAt, updatedAt sql.NullTime

	err := s.pool.QueryRow(context.Background(),
		"SELECT id, type, url, name, channel_id, playlist_id, enabled, schedule, last_processed, created_at, updated_at FROM youtube_sources WHERE id = $1",
		id).Scan(&src.ID, &src.Type, &src.URL, &src.Name, &channelID, &playlistID, &src.Enabled, &schedule, &lastProcessed, &createdAt, &updatedAt)

	if err != nil {
		return nil, false
	}

	if channelID.Valid {
		src.ChannelID = channelID.String
	}
	if playlistID.Valid {
		src.PlaylistID = playlistID.String
	}
	if schedule.Valid {
		src.Schedule = schedule.String
	}
	src.LastProcessed = parseTimestamp(lastProcessed)

	return &src, true
}

// CreateOrUpdateYouTubeSource creates or updates a YouTube source
func (s *PostgresStore) CreateOrUpdateYouTubeSource(source *models.YouTubeSource) {
	var lastProcessed interface{}
	if source.LastProcessed != "" {
		t, err := time.Parse(time.RFC3339, source.LastProcessed)
		if err == nil {
			lastProcessed = t
		}
	}

	_, err := s.pool.Exec(context.Background(),
		`INSERT INTO youtube_sources (id, type, url, name, channel_id, playlist_id, enabled, schedule, last_processed, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		 ON CONFLICT (id) DO UPDATE SET
		 type = EXCLUDED.type,
		 url = EXCLUDED.url,
		 name = EXCLUDED.name,
		 channel_id = EXCLUDED.channel_id,
		 playlist_id = EXCLUDED.playlist_id,
		 enabled = EXCLUDED.enabled,
		 schedule = EXCLUDED.schedule,
		 last_processed = EXCLUDED.last_processed,
		 updated_at = CURRENT_TIMESTAMP`,
		source.ID, source.Type, source.URL, source.Name, source.ChannelID, source.PlaylistID, source.Enabled, source.Schedule, lastProcessed)

	if err != nil {
		_ = err
	}
}

// DeleteYouTubeSource deletes a YouTube source by ID
func (s *PostgresStore) DeleteYouTubeSource(id string) bool {
	result, err := s.pool.Exec(context.Background(), "DELETE FROM youtube_sources WHERE id = $1", id)
	if err != nil {
		return false
	}
	return result.RowsAffected() > 0
}

// Video Transcript operations

// CreateOrUpdateTranscript creates or updates a video transcript
func (s *PostgresStore) CreateOrUpdateTranscript(transcript *models.VideoTranscript) {
	var duration interface{}
	if transcript.Duration != nil {
		duration = *transcript.Duration
	}

	_, err := s.pool.Exec(context.Background(),
		`INSERT INTO video_transcripts (id, video_id, video_title, video_url, text, duration, source_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
		 ON CONFLICT (id) DO UPDATE SET
		 video_id = EXCLUDED.video_id,
		 video_title = EXCLUDED.video_title,
		 video_url = EXCLUDED.video_url,
		 text = EXCLUDED.text,
		 duration = EXCLUDED.duration,
		 source_id = EXCLUDED.source_id`,
		transcript.ID, transcript.VideoID, transcript.VideoTitle, transcript.VideoURL, transcript.Text, duration, transcript.SourceID)

	if err != nil {
		_ = err
	}
}

// GetTranscriptByID returns a transcript by ID
func (s *PostgresStore) GetTranscriptByID(id string) (*models.VideoTranscript, bool) {
	var t models.VideoTranscript
	var duration sql.NullInt64
	var sourceID sql.NullString

	err := s.pool.QueryRow(context.Background(),
		"SELECT id, video_id, video_title, video_url, text, duration, source_id, created_at FROM video_transcripts WHERE id = $1",
		id).Scan(&t.ID, &t.VideoID, &t.VideoTitle, &t.VideoURL, &t.Text, &duration, &sourceID, &t.CreatedAt)

	if err != nil {
		return nil, false
	}

	t.Duration = parseIntPtr(duration)
	if sourceID.Valid {
		t.SourceID = sourceID.String
	}

	return &t, true
}

// GetTranscriptsByVideoID returns transcripts for a specific video ID
func (s *PostgresStore) GetTranscriptsByVideoID(videoID string) []*models.VideoTranscript {
	rows, err := s.pool.Query(context.Background(),
		"SELECT id, video_id, video_title, video_url, text, duration, source_id, created_at FROM video_transcripts WHERE video_id = $1 ORDER BY created_at DESC",
		videoID)
	if err != nil {
		return []*models.VideoTranscript{}
	}
	defer rows.Close()

	var transcripts []*models.VideoTranscript
	for rows.Next() {
		var t models.VideoTranscript
		var duration sql.NullInt64
		var sourceID sql.NullString

		err := rows.Scan(&t.ID, &t.VideoID, &t.VideoTitle, &t.VideoURL, &t.Text, &duration, &sourceID, &t.CreatedAt)
		if err != nil {
			continue
		}

		t.Duration = parseIntPtr(duration)
		if sourceID.Valid {
			t.SourceID = sourceID.String
		}

		transcripts = append(transcripts, &t)
	}

	return transcripts
}

// Market Analysis operations

// CreateOrUpdateMarketAnalysis creates or updates a market analysis
func (s *PostgresStore) CreateOrUpdateMarketAnalysis(analysis *models.MarketAnalysis) {
	trendsJSON, _ := json.Marshal(analysis.Trends)
	riskFactorsJSON, _ := json.Marshal(analysis.RiskFactors)

	_, err := s.pool.Exec(context.Background(),
		`INSERT INTO market_analyses (id, transcript_id, conditions, trends, risk_factors, summary, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
		 ON CONFLICT (id) DO UPDATE SET
		 transcript_id = EXCLUDED.transcript_id,
		 conditions = EXCLUDED.conditions,
		 trends = EXCLUDED.trends,
		 risk_factors = EXCLUDED.risk_factors,
		 summary = EXCLUDED.summary`,
		analysis.ID, analysis.TranscriptID, analysis.Conditions, trendsJSON, riskFactorsJSON, analysis.Summary)

	if err != nil {
		_ = err
	}
}

// GetMarketAnalysisByID returns a market analysis by ID
func (s *PostgresStore) GetMarketAnalysisByID(id string) (*models.MarketAnalysis, bool) {
	var a models.MarketAnalysis
	var trendsJSON, riskFactorsJSON []byte

	err := s.pool.QueryRow(context.Background(),
		"SELECT id, transcript_id, conditions, trends, risk_factors, summary, created_at FROM market_analyses WHERE id = $1",
		id).Scan(&a.ID, &a.TranscriptID, &a.Conditions, &trendsJSON, &riskFactorsJSON, &a.Summary, &a.CreatedAt)

	if err != nil {
		return nil, false
	}

	json.Unmarshal(trendsJSON, &a.Trends)
	json.Unmarshal(riskFactorsJSON, &a.RiskFactors)

	return &a, true
}

// GetMarketAnalysesByTranscriptID returns market analyses for a specific transcript ID
func (s *PostgresStore) GetMarketAnalysesByTranscriptID(transcriptID string) []*models.MarketAnalysis {
	rows, err := s.pool.Query(context.Background(),
		"SELECT id, transcript_id, conditions, trends, risk_factors, summary, created_at FROM market_analyses WHERE transcript_id = $1 ORDER BY created_at DESC",
		transcriptID)
	if err != nil {
		return []*models.MarketAnalysis{}
	}
	defer rows.Close()

	var analyses []*models.MarketAnalysis
	for rows.Next() {
		var a models.MarketAnalysis
		var trendsJSON, riskFactorsJSON []byte

		err := rows.Scan(&a.ID, &a.TranscriptID, &a.Conditions, &trendsJSON, &riskFactorsJSON, &a.Summary, &a.CreatedAt)
		if err != nil {
			continue
		}

		json.Unmarshal(trendsJSON, &a.Trends)
		json.Unmarshal(riskFactorsJSON, &a.RiskFactors)

		analyses = append(analyses, &a)
	}

	return analyses
}

// Recommendation operations

// CreateOrUpdateRecommendation creates or updates a recommendation
func (s *PostgresStore) CreateOrUpdateRecommendation(recommendation *models.Recommendation) {
	suggestedActionsJSON, _ := json.Marshal(recommendation.SuggestedActions)

	_, err := s.pool.Exec(context.Background(),
		`INSERT INTO recommendations (id, analysis_id, action, confidence, suggested_actions, summary, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
		 ON CONFLICT (id) DO UPDATE SET
		 analysis_id = EXCLUDED.analysis_id,
		 action = EXCLUDED.action,
		 confidence = EXCLUDED.confidence,
		 suggested_actions = EXCLUDED.suggested_actions,
		 summary = EXCLUDED.summary`,
		recommendation.ID, recommendation.AnalysisID, recommendation.Action, recommendation.Confidence, suggestedActionsJSON, recommendation.Summary)

	if err != nil {
		_ = err
	}
}

// GetRecommendationByID returns a recommendation by ID
func (s *PostgresStore) GetRecommendationByID(id string) (*models.Recommendation, bool) {
	var r models.Recommendation
	var suggestedActionsJSON []byte
	var summary sql.NullString

	err := s.pool.QueryRow(context.Background(),
		"SELECT id, analysis_id, action, confidence, suggested_actions, summary, created_at FROM recommendations WHERE id = $1",
		id).Scan(&r.ID, &r.AnalysisID, &r.Action, &r.Confidence, &suggestedActionsJSON, &summary, &r.CreatedAt)

	if err != nil {
		return nil, false
	}

	json.Unmarshal(suggestedActionsJSON, &r.SuggestedActions)
	if summary.Valid {
		r.Summary = summary.String
	}

	return &r, true
}

// GetRecommendationsByAnalysisID returns recommendations for a specific analysis ID
func (s *PostgresStore) GetRecommendationsByAnalysisID(analysisID string) []*models.Recommendation {
	rows, err := s.pool.Query(context.Background(),
		"SELECT id, analysis_id, action, confidence, suggested_actions, summary, created_at FROM recommendations WHERE analysis_id = $1 ORDER BY created_at DESC",
		analysisID)
	if err != nil {
		return []*models.Recommendation{}
	}
	defer rows.Close()

	var recommendations []*models.Recommendation
	for rows.Next() {
		var r models.Recommendation
		var suggestedActionsJSON []byte
		var summary sql.NullString

		err := rows.Scan(&r.ID, &r.AnalysisID, &r.Action, &r.Confidence, &suggestedActionsJSON, &summary, &r.CreatedAt)
		if err != nil {
			continue
		}

		json.Unmarshal(suggestedActionsJSON, &r.SuggestedActions)
		if summary.Valid {
			r.Summary = summary.String
		}

		recommendations = append(recommendations, &r)
	}

	return recommendations
}

// Workflow Execution operations

// CreateOrUpdateWorkflowExecution creates or updates a workflow execution
func (s *PostgresStore) CreateOrUpdateWorkflowExecution(execution *models.WorkflowExecution) {
	var startedAt, completedAt interface{}
	if execution.StartedAt != "" {
		t, err := time.Parse(time.RFC3339, execution.StartedAt)
		if err == nil {
			startedAt = t
		}
	}
	if execution.CompletedAt != "" {
		t, err := time.Parse(time.RFC3339, execution.CompletedAt)
		if err == nil {
			completedAt = t
		}
	}

	_, err := s.pool.Exec(context.Background(),
		`INSERT INTO workflow_executions (id, status, video_id, video_url, video_title, source_id, transcript_id, analysis_id, recommendation_id, error, created_at, started_at, completed_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, CURRENT_TIMESTAMP, $11, $12)
		 ON CONFLICT (id) DO UPDATE SET
		 status = EXCLUDED.status,
		 video_id = EXCLUDED.video_id,
		 video_url = EXCLUDED.video_url,
		 video_title = EXCLUDED.video_title,
		 source_id = EXCLUDED.source_id,
		 transcript_id = EXCLUDED.transcript_id,
		 analysis_id = EXCLUDED.analysis_id,
		 recommendation_id = EXCLUDED.recommendation_id,
		 error = EXCLUDED.error,
		 started_at = EXCLUDED.started_at,
		 completed_at = EXCLUDED.completed_at`,
		execution.ID, execution.Status, execution.VideoID, execution.VideoURL, execution.VideoTitle,
		execution.SourceID, execution.TranscriptID, execution.AnalysisID, execution.RecommendationID,
		execution.Error, startedAt, completedAt)

	if err != nil {
		_ = err
	}
}

// GetWorkflowExecutionByID returns a workflow execution by ID
func (s *PostgresStore) GetWorkflowExecutionByID(id string) (*models.WorkflowExecution, bool) {
	var e models.WorkflowExecution
	var videoTitle, videoID, sourceID, transcriptID, analysisID, recommendationID, errorMsg sql.NullString
	var startedAt, completedAt sql.NullTime

	err := s.pool.QueryRow(context.Background(),
		"SELECT id, status, video_id, video_url, video_title, source_id, transcript_id, analysis_id, recommendation_id, error, created_at, started_at, completed_at FROM workflow_executions WHERE id = $1",
		id).Scan(&e.ID, &e.Status, &videoID, &e.VideoURL, &videoTitle, &sourceID, &transcriptID, &analysisID, &recommendationID, &errorMsg, &e.CreatedAt, &startedAt, &completedAt)

	if err != nil {
		return nil, false
	}

	if videoID.Valid {
		e.VideoID = videoID.String
	}
	if videoTitle.Valid {
		e.VideoTitle = videoTitle.String
	}
	if sourceID.Valid {
		e.SourceID = sourceID.String
	}
	if transcriptID.Valid {
		e.TranscriptID = transcriptID.String
	}
	if analysisID.Valid {
		e.AnalysisID = analysisID.String
	}
	if recommendationID.Valid {
		e.RecommendationID = recommendationID.String
	}
	if errorMsg.Valid {
		e.Error = errorMsg.String
	}
	e.StartedAt = parseTimestamp(startedAt)
	e.CompletedAt = parseTimestamp(completedAt)

	return &e, true
}

// GetAllWorkflowExecutions returns all workflow executions
func (s *PostgresStore) GetAllWorkflowExecutions() []*models.WorkflowExecution {
	rows, err := s.pool.Query(context.Background(),
		"SELECT id, status, video_id, video_url, video_title, source_id, transcript_id, analysis_id, recommendation_id, error, created_at, started_at, completed_at FROM workflow_executions ORDER BY created_at DESC")
	if err != nil {
		return []*models.WorkflowExecution{}
	}
	defer rows.Close()

	var executions []*models.WorkflowExecution
	for rows.Next() {
		var e models.WorkflowExecution
		var videoTitle, videoID, sourceID, transcriptID, analysisID, recommendationID, errorMsg sql.NullString
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(&e.ID, &e.Status, &videoID, &e.VideoURL, &videoTitle, &sourceID, &transcriptID, &analysisID, &recommendationID, &errorMsg, &e.CreatedAt, &startedAt, &completedAt)
		if err != nil {
			continue
		}

		if videoID.Valid {
			e.VideoID = videoID.String
		}
		if videoTitle.Valid {
			e.VideoTitle = videoTitle.String
		}
		if sourceID.Valid {
			e.SourceID = sourceID.String
		}
		if transcriptID.Valid {
			e.TranscriptID = transcriptID.String
		}
		if analysisID.Valid {
			e.AnalysisID = analysisID.String
		}
		if recommendationID.Valid {
			e.RecommendationID = recommendationID.String
		}
		if errorMsg.Valid {
			e.Error = errorMsg.String
		}
		e.StartedAt = parseTimestamp(startedAt)
		e.CompletedAt = parseTimestamp(completedAt)

		executions = append(executions, &e)
	}

	return executions
}

// GetWorkflowExecutionsBySourceID returns workflow executions for a specific source ID
func (s *PostgresStore) GetWorkflowExecutionsBySourceID(sourceID string) []*models.WorkflowExecution {
	rows, err := s.pool.Query(context.Background(),
		"SELECT id, status, video_id, video_url, video_title, source_id, transcript_id, analysis_id, recommendation_id, error, created_at, started_at, completed_at FROM workflow_executions WHERE source_id = $1 ORDER BY created_at DESC",
		sourceID)
	if err != nil {
		return []*models.WorkflowExecution{}
	}
	defer rows.Close()

	var executions []*models.WorkflowExecution
	for rows.Next() {
		var e models.WorkflowExecution
		var videoTitle, videoID, sourceIDVal, transcriptID, analysisID, recommendationID, errorMsg sql.NullString
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(&e.ID, &e.Status, &videoID, &e.VideoURL, &videoTitle, &sourceIDVal, &transcriptID, &analysisID, &recommendationID, &errorMsg, &e.CreatedAt, &startedAt, &completedAt)
		if err != nil {
			continue
		}

		if videoID.Valid {
			e.VideoID = videoID.String
		}
		if videoTitle.Valid {
			e.VideoTitle = videoTitle.String
		}
		if sourceIDVal.Valid {
			e.SourceID = sourceIDVal.String
		}
		if transcriptID.Valid {
			e.TranscriptID = transcriptID.String
		}
		if analysisID.Valid {
			e.AnalysisID = analysisID.String
		}
		if recommendationID.Valid {
			e.RecommendationID = recommendationID.String
		}
		if errorMsg.Valid {
			e.Error = errorMsg.String
		}
		e.StartedAt = parseTimestamp(startedAt)
		e.CompletedAt = parseTimestamp(completedAt)

		executions = append(executions, &e)
	}

	return executions
}

