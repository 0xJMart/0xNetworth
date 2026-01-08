package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles communication with the Python workflow service
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// WorkflowRequest represents the request payload for the workflow service
type WorkflowRequest struct {
	YoutubeURL       string                 `json:"youtube_url"`
	PortfolioContext *PortfolioContext      `json:"portfolio_context,omitempty"`
}

// PortfolioContext represents portfolio holdings for context
type PortfolioContext struct {
	Holdings   []Holding `json:"holdings"`
	TotalValue float64   `json:"total_value,omitempty"`
}

// Holding represents a portfolio holding
type Holding struct {
	Symbol   string  `json:"symbol"`
	Quantity float64 `json:"quantity"`
	Value    float64 `json:"value"`
}

// WorkflowResponse represents the response from the workflow service
type WorkflowResponse struct {
	Transcript      Transcript      `json:"transcript"`
	MarketAnalysis  MarketAnalysis  `json:"market_analysis"`
	Recommendation  Recommendation  `json:"recommendation"`
}

// Transcript represents video transcript data
type Transcript struct {
	VideoID    string `json:"video_id"`
	VideoTitle string `json:"video_title"`
	Text       string `json:"text"`
	Duration   *int   `json:"duration,omitempty"`
}

// MarketAnalysis represents market condition analysis
type MarketAnalysis struct {
	Conditions  string   `json:"conditions"`
	Trends      []string `json:"trends"`
	RiskFactors []string `json:"risk_factors"`
	Summary     string   `json:"summary"`
}

// SuggestedAction represents a suggested action
type SuggestedAction struct {
	Type      string `json:"type"`
	Symbol    string `json:"symbol"`
	Rationale string `json:"rationale"`
}

// Recommendation represents investment recommendations
type Recommendation struct {
	Action           string           `json:"action"`
	Confidence       float64          `json:"confidence"`
	SuggestedActions []SuggestedAction `json:"suggested_actions"`
	Summary          string           `json:"summary,omitempty"`
}

// AggregatedRecommendationRequest represents the request for aggregated recommendations
type AggregatedRecommendationRequest struct {
	MarketAnalyses  []MarketAnalysis `json:"market_analyses"`
	Recommendations []Recommendation  `json:"recommendations"`
	PortfolioContext *PortfolioContext `json:"portfolio_context,omitempty"`
}

// AggregatedRecommendation represents consolidated recommendation from multiple videos
type AggregatedRecommendation struct {
	Action           string           `json:"action"`
	Confidence       float64          `json:"confidence"`
	SuggestedActions []SuggestedAction `json:"suggested_actions"`
	Summary          string           `json:"summary"`
	KeyInsights      []string         `json:"key_insights"`
}

// APIError represents an error from the workflow service
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("workflow service error: %d - %s", e.StatusCode, e.Message)
}

// NewClient creates a new workflow service client
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Workflow processing can take time
		},
	}
}

// ProcessVideo processes a YouTube video through the workflow
func (c *Client) ProcessVideo(request WorkflowRequest) (*WorkflowResponse, error) {
	url := c.baseURL + "/process"
	
	// Serialize request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}
	
	// Parse response
	var response WorkflowResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return &response, nil
}

// HealthCheck checks if the workflow service is healthy
func (c *Client) HealthCheck() error {
	url := c.baseURL + "/health"
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    "health check failed",
		}
	}
	
	return nil
}

// GenerateAggregatedRecommendation generates a consolidated recommendation from multiple video analyses
func (c *Client) GenerateAggregatedRecommendation(request AggregatedRecommendationRequest) (*AggregatedRecommendation, error) {
	url := c.baseURL + "/aggregate"
	
	// Serialize request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}
	
	// Parse response
	var response AggregatedRecommendation
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return &response, nil
}


