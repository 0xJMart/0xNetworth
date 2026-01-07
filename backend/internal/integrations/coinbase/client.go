package coinbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/coinbase/cdp-sdk/go/auth"
	"0xnetworth/backend/internal/models"
)

const (
	coinbaseAPIBaseURL = "https://api.coinbase.com/api/v3"
)

// APIError represents an error from the Coinbase API with status code
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("coinbase API error: %d - %s", e.StatusCode, e.Message)
}

// Client handles Coinbase API interactions
// Coinbase Advanced Trade API uses CDP API Keys for authentication:
// - apiKeyName: The CDP API Key ID (UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
//   or full path format: organizations/{org_id}/apiKeys/{key_id})
// - apiKeySecret: The Private Key (PEM format or base64-encoded DER) associated with that API key
// See: https://docs.cdp.coinbase.com/api-reference/v2/authentication
type Client struct {
	apiKeyName   string // CDP API Key ID (UUID or full path format)
	apiKeySecret string // API Key Secret (PEM or base64-encoded DER format)
	httpClient   *http.Client
}

// NewClient creates a new Coinbase API client using CDP API v2 authentication
// apiKeyName: The CDP API Key ID (UUID or full path format: organizations/{org_id}/apiKeys/{key_id})
// apiKeySecret: The Private Key - can be in PEM format or base64-encoded DER (as provided in JSON file from CDP Portal)
// The CDP SDK handles parsing of the private key in various formats (ES256 or Ed25519)
// See: https://docs.cdp.coinbase.com/api-reference/v2/authentication#creating-secret-api-keys
func NewClient(apiKeyName, apiKeySecret string) (*Client, error) {
	if apiKeyName == "" {
		return nil, fmt.Errorf("apiKeyName cannot be empty")
	}
	if apiKeySecret == "" {
		return nil, fmt.Errorf("apiKeySecret cannot be empty")
	}

	return &Client{
		apiKeyName:   apiKeyName,
		apiKeySecret: apiKeySecret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Coinbase API Response Types
type coinbasePortfolio struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	UserID   string `json:"user_id"`
}

type coinbasePortfoliosResponse struct {
	Portfolios []coinbasePortfolio `json:"portfolios"`
	// Some API versions might return data directly
	Data []coinbasePortfolio `json:"data"`
}

type coinbaseProduct struct {
	ProductID   string `json:"product_id"`
	Price       string `json:"price"`
	Price24h    string `json:"price_percentage_change_24h"`
	Volume24h   string `json:"volume_24h"`
	VolumePercentageChange24h string `json:"volume_percentage_change_24h"`
	BaseIncrement string `json:"base_increment"`
	QuoteIncrement string `json:"quote_increment"`
	BaseMinSize    string `json:"base_min_size"`
	BaseMaxSize    string `json:"base_max_size"`
	QuoteMinSize   string `json:"quote_min_size"`
	QuoteMaxSize   string `json:"quote_max_size"`
	BaseName       string `json:"base_name"`
	QuoteName      string `json:"quote_name"`
	Watched        bool   `json:"watched"`
	IsDisabled     bool   `json:"is_disabled"`
	New            bool   `json:"new"`
	Status         string `json:"status"`
	CancelOnly     bool   `json:"cancel_only"`
	LimitOnly      bool   `json:"limit_only"`
	PostOnly       bool   `json:"post_only"`
	TradingDisabled bool   `json:"trading_disabled"`
}

type coinbasePortfolioHoldings struct {
	PortfolioID string `json:"portfolio_id"`
	ProductID   string `json:"product_id"`
	Quantity    string `json:"quantity"`
	Available   string `json:"available"`
	Hold        string `json:"hold"`
}

type coinbasePortfolioHoldingsResponse struct {
	Holdings []coinbasePortfolioHoldings `json:"holdings"`
	// Some API versions might return data directly
	Data []coinbasePortfolioHoldings `json:"data"`
}

// Portfolio Breakdown Response Types (using the correct API endpoint)
type coinbaseSpotPosition struct {
	Asset                    string  `json:"asset"`
	AccountUUID              string  `json:"account_uuid"`
	TotalBalanceFiat         float64 `json:"total_balance_fiat"`
	TotalBalanceCrypto       float64 `json:"total_balance_crypto"`
	AvailableToTradeFiat     float64 `json:"available_to_trade_fiat"`
	AvailableToTradeCrypto   float64 `json:"available_to_trade_crypto"`
	AverageEntryPrice        struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"average_entry_price"`
	AssetUUID string `json:"asset_uuid"`
	IsCash    bool   `json:"is_cash"`
}

type coinbasePortfolioBreakdown struct {
	Portfolio struct {
		Name string `json:"name"`
		UUID string `json:"uuid"`
		Type string `json:"type"`
	} `json:"portfolio"`
	SpotPositions []coinbaseSpotPosition `json:"spot_positions"`
	PerpPositions []interface{}          `json:"perp_positions"`
	FuturesPositions []interface{}       `json:"futures_positions"`
}

type coinbasePortfolioBreakdownResponse struct {
	Breakdown coinbasePortfolioBreakdown `json:"breakdown"`
}

// GenerateJWT creates a JWT token for Coinbase Advanced Trade API authentication
// Uses the official CDP SDK for JWT generation, which handles all the complexity
// of CDP API v2 authentication format automatically
// This method is exported for testing purposes
func (c *Client) GenerateJWT(method, path string) (string, error) {
	return c.generateJWT(method, path)
}

// generateJWT creates a JWT token for Coinbase Advanced Trade API authentication
// Uses the official CDP SDK which handles:
// - Proper JWT claim structure (uris array, sub, iss, nbf, exp)
// - Support for both ES256 and Ed25519 keys
// - Correct header format (kid, no nonce for REST API)
// See: https://docs.cdp.coinbase.com/api-reference/v2/authentication
func (c *Client) generateJWT(method, path string) (string, error) {
	// Use CDP SDK to generate JWT
	// For Advanced Trade API, host is api.coinbase.com
	jwt, err := auth.GenerateJWT(auth.JwtOptions{
		KeyID:         c.apiKeyName,
		KeySecret:     c.apiKeySecret,
		RequestMethod: method,
		RequestHost:   "api.coinbase.com",
		RequestPath:   path,
		ExpiresIn:     120, // 2 minutes, default
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT using CDP SDK: %w", err)
	}

	return jwt, nil
}

// makeRequest makes an authenticated request to Coinbase API using JWT
func (c *Client) makeRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := coinbaseAPIBaseURL + path
	
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Generate JWT token for this request
	// JWT path must include /api/v3 to match the actual request URL
	fullPath := "/api/v3" + path
	jwtToken, err := c.generateJWT(method, fullPath)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}
	log.Printf("Generated JWT for [%s %s], token length: %d", method, fullPath, len(jwtToken))

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	// Log non-2xx responses for debugging
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		// Create a new reader for the body since we consumed it
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		log.Printf("Coinbase API error [%s %s]: %d - %s", method, path, resp.StatusCode, string(bodyBytes))
		log.Printf("Request URL: %s", url)
		log.Printf("API Key Name: %s", c.apiKeyName)
	}

	return resp, nil
}

// GetPortfolios fetches portfolios from Coinbase
// IMPORTANT: This endpoint only returns portfolios that the API key has access to.
// If your API key is scoped to a specific portfolio, only that portfolio will be returned.
// To see all portfolios, ensure your API key has "Portfolio primary view access" 
// or is not scoped to a specific portfolio in Coinbase Developer Platform.
func (c *Client) GetPortfolios() ([]coinbasePortfolio, error) {
	resp, err := c.makeRequest("GET", "/brokerage/portfolios", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch portfolios: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}

	// Read the full response body for logging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	// Log the raw response for debugging
	log.Printf("GetPortfolios: Raw API response: %s", string(bodyBytes))

	var apiResp coinbasePortfoliosResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Log what we found
	log.Printf("GetPortfolios: Found %d portfolios in 'Portfolios' field, %d in 'Data' field", 
		len(apiResp.Portfolios), len(apiResp.Data))
	
	// Log portfolio details
	if len(apiResp.Portfolios) > 0 {
		for i, p := range apiResp.Portfolios {
			log.Printf("GetPortfolios: Portfolio[%d]: UUID=%s, Name=%s, Type=%s", i, p.UUID, p.Name, p.Type)
		}
	}
	if len(apiResp.Data) > 0 {
		for i, p := range apiResp.Data {
			log.Printf("GetPortfolios: Data[%d]: UUID=%s, Name=%s, Type=%s", i, p.UUID, p.Name, p.Type)
		}
	}

	// Handle different response formats
	if len(apiResp.Portfolios) > 0 {
		return apiResp.Portfolios, nil
	}
	return apiResp.Data, nil
}

// GetPortfolioHoldings fetches holdings for a specific portfolio using the Portfolio Breakdown endpoint
// This endpoint returns spot_positions which contain the actual holdings/assets in the portfolio
func (c *Client) GetPortfolioHoldings(portfolioID string) ([]coinbaseSpotPosition, error) {
	// Use the correct endpoint: GET /api/v3/brokerage/portfolios/{portfolio_uuid}
	path := fmt.Sprintf("/brokerage/portfolios/%s", portfolioID)
	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch portfolio breakdown: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}

	var apiResp coinbasePortfolioBreakdownResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode portfolio breakdown response: %w", err)
	}

	// Return spot positions (filter out cash positions)
	positions := []coinbaseSpotPosition{}
	for _, pos := range apiResp.Breakdown.SpotPositions {
		if !pos.IsCash {
			positions = append(positions, pos)
		}
	}

	return positions, nil
}

// GetProductPrice fetches current price for a product
func (c *Client) GetProductPrice(productID string) (float64, error) {
	path := fmt.Sprintf("/brokerage/products/%s", productID)
	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch product: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return 0, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}

	var product coinbaseProduct
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	price, err := strconv.ParseFloat(product.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}

	return price, nil
}

// GetInvestments fetches investment holdings from Coinbase
func (c *Client) GetInvestments(accountID string) ([]*models.Investment, error) {
	// First, get all portfolios
	portfolios, err := c.GetPortfolios()
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolios: %w", err)
	}

	investments := make([]*models.Investment, 0)

	// For each portfolio, get holdings
	for _, portfolio := range portfolios {
		holdings, err := c.GetPortfolioHoldings(portfolio.UUID)
		if err != nil {
			// Log error but continue with other portfolios
			continue
		}

		for _, position := range holdings {
			// Use the asset symbol (e.g., "BTC", "ETH")
			symbol := position.Asset
			
			// Calculate price from average entry price if available
			var price float64
			if position.AverageEntryPrice.Value != "" {
				price, _ = strconv.ParseFloat(position.AverageEntryPrice.Value, 64)
			} else if position.TotalBalanceCrypto > 0 {
				price = position.TotalBalanceFiat / position.TotalBalanceCrypto
			} else {
				continue
			}

			quantity := position.TotalBalanceCrypto
			value := position.TotalBalanceFiat

			investment := &models.Investment{
				ID:          fmt.Sprintf("%s-%s", portfolio.UUID, position.AssetUUID),
				AccountID:   portfolio.UUID,
				Platform:    models.PlatformCoinbase,
				Symbol:      symbol,
				Name:        symbol,
				Quantity:    quantity,
				Value:       value,
				Price:       price,
				Currency:    "USD",
				AssetType:    "crypto",
				LastUpdated: time.Now().UTC().Format(time.RFC3339),
			}
			investments = append(investments, investment)
		}
	}

	return investments, nil
}

// SyncAll syncs all portfolios and investments from Coinbase
// Uses Portfolio primary view access which is the standard for Coinbase Advanced Trade
func (c *Client) SyncAll() ([]*models.Portfolio, []*models.Investment, error) {
	log.Printf("SyncAll: Starting sync with API key: %s", c.apiKeyName)

	// Get portfolios and investments
	// This works with "Portfolio primary view access"
	investments := make([]*models.Investment, 0)
	log.Printf("SyncAll: Attempting to fetch portfolios...")
	portfolios, err := c.GetPortfolios()
	if err != nil {
		// If we can't get portfolios either, return what we have
		if apiErr, ok := err.(*APIError); ok {
			log.Printf("Error: Failed to get portfolios: %d - %s", apiErr.StatusCode, apiErr.Message)
		} else {
			log.Printf("Error: Failed to get portfolios: %v", err)
		}
		return nil, investments, fmt.Errorf("failed to get portfolios: %w", err)
	}

	log.Printf("Info: Found %d portfolios", len(portfolios))

	// Convert portfolios to models
	portfolioModels := make([]*models.Portfolio, 0, len(portfolios))
	for _, p := range portfolios {
		portfolioModels = append(portfolioModels, &models.Portfolio{
			ID:         p.UUID,
			Platform:   models.PlatformCoinbase,
			Name:       p.Name,
			Type:       p.Type,
			LastSynced: time.Now().UTC().Format(time.RFC3339),
		})
	}

	// For each portfolio, get holdings directly
	for _, portfolio := range portfolios {
		log.Printf("Info: Fetching holdings for portfolio %s (%s)", portfolio.UUID, portfolio.Name)
		holdings, err := c.GetPortfolioHoldings(portfolio.UUID)
		if err != nil {
			// Log but continue with other portfolios
			if apiErr, ok := err.(*APIError); ok {
				log.Printf("Warning: Failed to get holdings for portfolio %s: %d - %s", portfolio.UUID, apiErr.StatusCode, apiErr.Message)
			} else {
				log.Printf("Warning: Failed to get holdings for portfolio %s: %v", portfolio.UUID, err)
			}
			continue
		}

		log.Printf("Info: Found %d spot positions in portfolio %s", len(holdings), portfolio.UUID)

		// Convert spot positions to investments
		for _, position := range holdings {
			// Use the asset symbol as the symbol (e.g., "BTC", "ETH")
			symbol := position.Asset
			
			// Calculate price from average entry price if available, otherwise use current balance
			var price float64
			if position.AverageEntryPrice.Value != "" {
				price, _ = strconv.ParseFloat(position.AverageEntryPrice.Value, 64)
			} else if position.TotalBalanceCrypto > 0 {
				// Fallback: calculate price from fiat balance / crypto balance
				price = position.TotalBalanceFiat / position.TotalBalanceCrypto
			} else {
				// If no price available, skip this position
				log.Printf("Warning: No price available for asset %s, skipping", symbol)
				continue
			}

			// Use total balance in crypto as quantity
			quantity := position.TotalBalanceCrypto
			value := position.TotalBalanceFiat

			investment := &models.Investment{
				ID:          fmt.Sprintf("%s-%s", portfolio.UUID, position.AssetUUID),
				AccountID:   portfolio.UUID,
				Platform:    models.PlatformCoinbase,
				Symbol:      symbol,
				Name:        symbol,
				Quantity:    quantity,
				Value:       value,
				Price:       price,
				Currency:    "USD", // Portfolio breakdown returns values in USD
				AssetType:    "crypto",
				LastUpdated: time.Now().UTC().Format(time.RFC3339),
			}
			investments = append(investments, investment)
			log.Printf("Info: Added investment: %s - Quantity: %f, Value: $%.2f", symbol, quantity, value)
		}

		log.Printf("Info: Converted %d spot positions to investments from portfolio %s", len(holdings), portfolio.UUID)
	}

	log.Printf("Info: SyncAll completed - %d portfolios, %d investments", len(portfolioModels), len(investments))
	return portfolioModels, investments, nil
}
