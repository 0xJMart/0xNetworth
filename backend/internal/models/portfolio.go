package models

// Platform represents the investment platform
type Platform string

const (
	PlatformCoinbase Platform = "coinbase"
)

// Portfolio represents a portfolio/account from an investment platform
type Portfolio struct {
	ID          string   `json:"id"`
	Platform    Platform `json:"platform"`
	Name        string   `json:"name"`
	Type        string   `json:"type,omitempty"` // e.g., "default", "main"
	LastSynced  string   `json:"last_synced,omitempty"` // ISO 8601 timestamp
}

