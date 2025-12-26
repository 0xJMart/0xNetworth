package coinbase

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"0xnetworth/backend/internal/models"
)

const (
	coinbaseAPIBaseURL = "https://api.coinbase.com/api/v3"
)

// Client handles Coinbase API interactions
// Coinbase Advanced Trade API uses:
// - apiKeyName: The API Key Name (ID) you created in Coinbase
// - privateKey: The Private Key associated with that API key
type Client struct {
	apiKeyName string // API Key Name/ID from Coinbase
	privateKey string // Private Key from Coinbase
	httpClient *http.Client
}

// NewClient creates a new Coinbase API client
// apiKeyName: The API Key Name (ID) from Coinbase Advanced Trade API
// privateKey: The Private Key from Coinbase Advanced Trade API
func NewClient(apiKeyName, privateKey string) *Client {
	return &Client{
		apiKeyName: apiKeyName,
		privateKey: privateKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Coinbase API Response Types
type coinbaseAccount struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Currency    string `json:"currency"`
	Available   string `json:"available_balance"`
	Hold        string `json:"hold_balance"`
	Type        string `json:"type"`
	Active      bool   `json:"active"`
}

type coinbaseAccountsResponse struct {
	Accounts []coinbaseAccount `json:"accounts"`
	// Some API versions might return data directly
	Data []coinbaseAccount `json:"data"`
}

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

// createSignature creates HMAC signature for Coinbase API authentication
func (c *Client) createSignature(timestamp, method, path, body string) string {
	message := timestamp + method + path + body
	mac := hmac.New(sha256.New, []byte(c.privateKey))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// makeRequest makes an authenticated request to Coinbase API
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

	// Coinbase API authentication
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := c.createSignature(timestamp, method, path, string(bodyBytes))

	req.Header.Set("CB-ACCESS-KEY", c.apiKeyName)
	req.Header.Set("CB-ACCESS-SIGN", signature)
	req.Header.Set("CB-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

// GetAccounts fetches accounts from Coinbase
func (c *Client) GetAccounts() ([]*models.Account, error) {
	resp, err := c.makeRequest("GET", "/brokerage/accounts", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("coinbase API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp coinbaseAccountsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Handle different response formats
	accountList := apiResp.Accounts
	if len(accountList) == 0 {
		accountList = apiResp.Data
	}

	accounts := make([]*models.Account, 0, len(accountList))
	for _, acc := range accountList {
		if !acc.Active {
			continue
		}

		available, _ := strconv.ParseFloat(acc.Available, 64)
		hold, _ := strconv.ParseFloat(acc.Hold, 64)
		balance := available + hold

		account := &models.Account{
			ID:          acc.UUID,
			Platform:    models.PlatformCoinbase,
			Name:        acc.Name,
			Balance:     balance,
			Currency:    acc.Currency,
			AccountType: acc.Type,
			LastSynced:  time.Now().UTC().Format(time.RFC3339),
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// GetPortfolios fetches portfolios from Coinbase
func (c *Client) GetPortfolios() ([]coinbasePortfolio, error) {
	resp, err := c.makeRequest("GET", "/brokerage/portfolios", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch portfolios: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("coinbase API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp coinbasePortfoliosResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Handle different response formats
	if len(apiResp.Portfolios) > 0 {
		return apiResp.Portfolios, nil
	}
	return apiResp.Data, nil
}

// GetPortfolioHoldings fetches holdings for a specific portfolio
func (c *Client) GetPortfolioHoldings(portfolioID string) ([]coinbasePortfolioHoldings, error) {
	path := fmt.Sprintf("/brokerage/portfolios/%s/holdings", portfolioID)
	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch portfolio holdings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("coinbase API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp coinbasePortfolioHoldingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Handle different response formats
	if len(apiResp.Holdings) > 0 {
		return apiResp.Holdings, nil
	}
	return apiResp.Data, nil
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
		return 0, fmt.Errorf("coinbase API error: %d - %s", resp.StatusCode, string(bodyBytes))
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

		for _, holding := range holdings {
			// Get current price for the product
			price, err := c.GetProductPrice(holding.ProductID)
			if err != nil {
				// If we can't get price, skip this holding
				continue
			}

			quantity, _ := strconv.ParseFloat(holding.Quantity, 64)
			value := quantity * price

			// Parse product ID (format: BTC-USD, ETH-USD, etc.)
			baseCurrency := holding.ProductID
			if len(holding.ProductID) > 4 {
				baseCurrency = holding.ProductID[:len(holding.ProductID)-4]
			}

			investment := &models.Investment{
				ID:          fmt.Sprintf("%s-%s", portfolio.UUID, holding.ProductID),
				AccountID:   portfolio.UUID,
				Platform:    models.PlatformCoinbase,
				Symbol:      baseCurrency,
				Name:        baseCurrency,
				Quantity:    quantity,
				Value:       value,
				Price:       price,
				Currency:    "USD", // Coinbase typically uses USD as quote currency
				AssetType:    "crypto",
				LastUpdated: time.Now().UTC().Format(time.RFC3339),
			}
			investments = append(investments, investment)
		}
	}

	return investments, nil
}

// SyncAccount syncs a specific account from Coinbase
func (c *Client) SyncAccount(accountID string) error {
	// This is a placeholder - in a full implementation, we might want to
	// sync a specific account. For now, we'll sync all accounts.
	_, err := c.GetAccounts()
	return err
}

// SyncAll syncs all accounts and investments from Coinbase
func (c *Client) SyncAll() ([]*models.Account, []*models.Investment, error) {
	accounts, err := c.GetAccounts()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	// Get investments from all portfolios
	investments := make([]*models.Investment, 0)
	portfolios, err := c.GetPortfolios()
	if err == nil {
		for _, portfolio := range portfolios {
			portfolioInvestments, err := c.GetInvestments(portfolio.UUID)
			if err == nil {
				investments = append(investments, portfolioInvestments...)
			}
		}
	}

	return accounts, investments, nil
}
