package coinbase

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
// Coinbase Advanced Trade API uses:
// - apiKeyName: The API Key Name (ID) you created in Coinbase
// - privateKey: The Private Key (PEM format) associated with that API key
type Client struct {
	apiKeyName string        // API Key Name/ID from Coinbase
	privateKey *ecdsa.PrivateKey // Parsed ECDSA private key
	httpClient *http.Client
}

// NewClient creates a new Coinbase API client
// apiKeyName: The API Key Name (ID) from Coinbase Advanced Trade API
// privateKeyData: The Private Key - can be in PEM format or base64-encoded DER (as provided in JSON file)
func NewClient(apiKeyName, privateKeyData string) (*Client, error) {
	var keyBytes []byte
	var err error
	
	// Coinbase provides the private key in two possible formats:
	// 1. PEM format: "-----BEGIN EC PRIVATE KEY-----\n...\n-----END EC PRIVATE KEY-----\n"
	// 2. Base64-encoded DER (from JSON file): just the base64 string without PEM headers
	
	// First, try to decode as PEM format
	block, _ := pem.Decode([]byte(privateKeyData))
	if block != nil {
		// Successfully decoded PEM block - use the DER bytes directly
		keyBytes = block.Bytes
	} else {
		// Not in PEM format, assume it's base64-encoded DER (from JSON file)
		// Remove any whitespace/newlines that might be present
		cleanedKey := strings.TrimSpace(privateKeyData)
		cleanedKey = strings.ReplaceAll(cleanedKey, "\n", "")
		cleanedKey = strings.ReplaceAll(cleanedKey, " ", "")
		
		// Decode base64 to get DER bytes
		keyBytes, err = base64.StdEncoding.DecodeString(cleanedKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode private key from base64: %w", err)
		}
	}

	// Parse the DER bytes as an EC private key
	// Try SEC1 format first (EC private key), then PKCS8 if that fails
	var privateKey *ecdsa.PrivateKey
	ecKey, err := x509.ParseECPrivateKey(keyBytes)
	if err != nil {
		// If SEC1 format fails, try PKCS8 format (used in some Coinbase API key formats)
		pkcs8Key, pkcs8Err := x509.ParsePKCS8PrivateKey(keyBytes)
		if pkcs8Err != nil {
			return nil, fmt.Errorf("failed to parse private key (tried both SEC1 and PKCS8): SEC1 error: %w, PKCS8 error: %v", err, pkcs8Err)
		}
		// Convert PKCS8 to ECDSA
		var ok bool
		privateKey, ok = pkcs8Key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not an ECDSA key")
		}
	} else {
		privateKey = ecKey
	}

	return &Client{
		apiKeyName: apiKeyName,
		privateKey: privateKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
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

// generateJWT creates a JWT token for Coinbase Advanced Trade API authentication
// The JWT must include the request URI in the payload for REST API requests
func (c *Client) generateJWT(method, path string) (string, error) {
	now := time.Now()
	
	// Create URI claim: "{method} {host}{path}"
	uri := fmt.Sprintf("%s api.coinbase.com%s", method, path)
	
	// Generate a random nonce
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonce := fmt.Sprintf("%x", nonceBytes)
	
	// Create JWT claims
	claims := jwt.MapClaims{
		"sub": c.apiKeyName,                    // Subject: API key ID
		"iss": "cdp",                           // Issuer: Coinbase Developer Platform
		"nbf": now.Unix(),                      // Not before: current time
		"exp": now.Unix() + 120,                // Expiration: 2 minutes from now
		"uri": uri,                             // URI claim for REST API
	}
	
	// Create token with headers
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = c.apiKeyName
	token.Header["nonce"] = nonce
	
	// Sign the token
	tokenString, err := token.SignedString(c.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}
	
	return tokenString, nil
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
	jwtToken, err := c.generateJWT(method, path)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

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
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
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
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
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
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
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
// Handles cases where account access is restricted (e.g., Portfolio primary view access only)
func (c *Client) SyncAll() ([]*models.Account, []*models.Investment, error) {
	log.Printf("SyncAll: Starting sync with API key: %s", c.apiKeyName)
	
	// Try to get accounts, but don't fail if we get 403 (forbidden)
	// This handles cases where the API key only has "Portfolio primary view access"
	accounts := make([]*models.Account, 0)
	log.Printf("SyncAll: Attempting to fetch accounts...")
	accountsResp, err := c.GetAccounts()
	if err != nil {
		// Check if it's a 403/forbidden error
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == http.StatusForbidden {
			// This is expected for "Portfolio primary view access" - continue with portfolios
			log.Printf("Info: Account access forbidden (expected with Portfolio primary view access), continuing with portfolios. Error: %s", apiErr.Message)
		} else {
			// For other errors, log but continue
			log.Printf("Warning: Failed to get accounts (may be expected with limited permissions): %v", err)
		}
	} else {
		accounts = accountsResp
		log.Printf("SyncAll: Successfully fetched %d accounts", len(accounts))
	}

	// Get investments from all portfolios
	// This should work with "Portfolio primary view access"
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
		return accounts, investments, fmt.Errorf("failed to get portfolios: %w", err)
	}

	log.Printf("Info: Found %d portfolios", len(portfolios))

	for _, portfolio := range portfolios {
		portfolioInvestments, err := c.GetInvestments(portfolio.UUID)
		if err == nil {
			investments = append(investments, portfolioInvestments...)
			log.Printf("Info: Got %d investments from portfolio %s", len(portfolioInvestments), portfolio.UUID)
		} else {
			// Log but continue with other portfolios
			if apiErr, ok := err.(*APIError); ok {
				log.Printf("Warning: Failed to get investments for portfolio %s: %d - %s", portfolio.UUID, apiErr.StatusCode, apiErr.Message)
			} else {
				log.Printf("Warning: Failed to get investments for portfolio %s: %v", portfolio.UUID, err)
			}
		}
	}

	log.Printf("Info: SyncAll completed - %d accounts, %d investments", len(accounts), len(investments))
	return accounts, investments, nil
}
