package models

// MarketAnalysis represents market condition analysis results
type MarketAnalysis struct {
	ID          string   `json:"id"`
	TranscriptID string  `json:"transcript_id"`
	Conditions  string   `json:"conditions"` // bullish, bearish, neutral
	Trends      []string `json:"trends"`
	RiskFactors []string `json:"risk_factors"`
	Summary     string   `json:"summary"`
	CreatedAt   string   `json:"created_at,omitempty"` // ISO 8601 timestamp
}

// SuggestedAction represents an individual suggested action
type SuggestedAction struct {
	Type      string `json:"type"`      // increase, decrease, hold, add, remove
	Symbol    string `json:"symbol"`
	Rationale string `json:"rationale"`
}

// Recommendation represents investment recommendations
type Recommendation struct {
	ID             string           `json:"id"`
	AnalysisID     string           `json:"analysis_id"`
	Action         string           `json:"action"` // rebalance, hold, diversify, etc.
	Confidence     float64          `json:"confidence"` // 0.0 to 1.0
	SuggestedActions []SuggestedAction `json:"suggested_actions"`
	Summary        string           `json:"summary,omitempty"`
	CreatedAt      string           `json:"created_at,omitempty"` // ISO 8601 timestamp
}


