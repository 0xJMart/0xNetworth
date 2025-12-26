package models

// Platform represents the investment platform
type Platform string

const (
	PlatformCoinbase Platform = "coinbase"
)

// Account represents an investment account
type Account struct {
	ID          string   `json:"id"`
	Platform    Platform `json:"platform"`
	Name        string   `json:"name"`
	Balance     float64  `json:"balance"`
	Currency    string   `json:"currency"`
	AccountType string   `json:"account_type,omitempty"` // e.g., "trading", "savings", "retirement"
	LastSynced  string   `json:"last_synced,omitempty"` // ISO 8601 timestamp
}

