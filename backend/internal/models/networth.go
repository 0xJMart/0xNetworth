package models

// NetWorth represents aggregated net worth information
type NetWorth struct {
	TotalValue    float64            `json:"total_value"`
	Currency      string             `json:"currency"`
	ByPlatform    map[Platform]float64 `json:"by_platform"`    // Value per platform
	ByAssetType   map[string]float64   `json:"by_asset_type"`  // Value per asset type
	AccountCount  int                `json:"account_count"`
	LastCalculated string            `json:"last_calculated"`  // ISO 8601 timestamp
}

// NetWorthBreakdown provides detailed breakdown of net worth
type NetWorthBreakdown struct {
	NetWorth
	Portfolios  []*Portfolio  `json:"portfolios"`
	Investments []*Investment `json:"investments"`
}

