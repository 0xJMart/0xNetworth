package models

// Investment represents an investment holding
type Investment struct {
	ID          string   `json:"id"`
	AccountID   string   `json:"account_id"`
	Platform    Platform `json:"platform"`
	Symbol      string   `json:"symbol"`       // e.g., "BTC", "ETH", "AAPL", "VTI"
	Name        string   `json:"name"`          // Full name of the asset
	Quantity    float64  `json:"quantity"`     // Number of shares/coins
	Value       float64  `json:"value"`        // Current value in account currency
	Price       float64  `json:"price"`        // Current price per unit
	Currency    string   `json:"currency"`     // Currency of the investment
	AssetType   string   `json:"asset_type"`  // e.g., "crypto", "stock", "etf", "bond"
	LastUpdated string   `json:"last_updated,omitempty"` // ISO 8601 timestamp
}

