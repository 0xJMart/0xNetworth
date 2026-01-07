package youtube

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// MaxResultsDefault is the default number of videos to fetch
	MaxResultsDefault = 50
	// MaxResultsMax is the maximum number of videos per request
	MaxResultsMax = 50
	// MaxErrorMessageSize limits error message size to prevent memory issues
	MaxErrorMessageSize = 500
)

// Client handles communication with YouTube Data API v3
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// Video represents a YouTube video from the API
type Video struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PublishedAt time.Time `json:"publishedAt"`
	ChannelID   string    `json:"channelId"`
	ChannelTitle string   `json:"channelTitle"`
}

// SearchResponse represents the response from YouTube Data API search endpoint
type SearchResponse struct {
	Items []SearchItem `json:"items"`
}

// SearchItem represents a single item in the search response
type SearchItem struct {
	ID      VideoID      `json:"id"`
	Snippet VideoSnippet `json:"snippet"`
}

// VideoID represents the video ID structure
type VideoID struct {
	VideoID string `json:"videoId"`
}

// VideoSnippet represents video metadata
type VideoSnippet struct {
	PublishedAt  string `json:"publishedAt"`
	ChannelID    string `json:"channelId"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ChannelTitle string `json:"channelTitle"`
}

// APIError represents an error from the YouTube API
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("YouTube API error: %d - %s", e.StatusCode, e.Message)
}

// NewClient creates a new YouTube Data API client
func NewClient(apiKey string) *Client {
	if apiKey == "" {
		return nil
	}

	return &Client{
		apiKey: apiKey,
		baseURL: "https://www.googleapis.com/youtube/v3",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetChannelVideos fetches recent videos from a YouTube channel
// channelID: The YouTube channel ID (not the custom URL)
// maxResults: Maximum number of videos to return (1-50)
// publishedAfter: Only return videos published after this time (optional)
func (c *Client) GetChannelVideos(channelID string, maxResults int, publishedAfter *time.Time) ([]Video, error) {
	if c == nil {
		return nil, fmt.Errorf("YouTube client not initialized (API key not set)")
	}

	if maxResults < 1 {
		maxResults = 10
	}
	if maxResults > 50 {
		maxResults = 50
	}

	// Build request URL
	reqURL := fmt.Sprintf("%s/search", c.baseURL)
	params := url.Values{}
	params.Set("key", c.apiKey)
	params.Set("channelId", channelID)
	params.Set("type", "video")
	params.Set("order", "date")
	params.Set("part", "snippet")
	params.Set("maxResults", fmt.Sprintf("%d", maxResults))

	if publishedAfter != nil {
		params.Set("publishedAfter", publishedAfter.Format(time.RFC3339))
	}

	reqURL += "?" + params.Encode()

	// Make request
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		errorMsg := string(bodyBytes)
		// Limit error message size
		if len(errorMsg) > MaxErrorMessageSize {
			errorMsg = errorMsg[:MaxErrorMessageSize] + "..."
		}
		
		// Provide user-friendly error messages for common cases
		switch resp.StatusCode {
		case http.StatusForbidden:
			errorMsg = "YouTube API quota exceeded or API key invalid"
		case http.StatusBadRequest:
			errorMsg = "Invalid YouTube API request: " + errorMsg
		case http.StatusUnauthorized:
			errorMsg = "YouTube API key is invalid or missing"
		}
		
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    errorMsg,
		}
	}

	// Parse response
	var searchResp SearchResponse
	if err := json.Unmarshal(bodyBytes, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to Video structs
	videos := make([]Video, 0, len(searchResp.Items))
	for _, item := range searchResp.Items {
		publishedAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			// Skip videos with invalid timestamps
			continue
		}

		videos = append(videos, Video{
			ID:           item.ID.VideoID,
			Title:        item.Snippet.Title,
			Description:  item.Snippet.Description,
			PublishedAt:  publishedAt,
			ChannelID:    item.Snippet.ChannelID,
			ChannelTitle: item.Snippet.ChannelTitle,
		})
	}

	return videos, nil
}

// ExtractChannelID extracts channel ID from various YouTube URL formats
// Supports:
// - https://www.youtube.com/channel/UC... (standard channel ID format)
// - https://www.youtube.com/@username (custom handle format)
// - https://www.youtube.com/c/ChannelName (custom URL format)
func (c *Client) ExtractChannelID(channelURL string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("YouTube client not initialized (API key not set)")
	}

	// Handle standard channel URL: youtube.com/channel/UC...
	if strings.Contains(channelURL, "/channel/") {
		parts := strings.Split(channelURL, "/channel/")
		if len(parts) > 1 {
			channelID := strings.Split(parts[1], "/")[0]
			channelID = strings.Split(channelID, "?")[0]
			// Channel IDs typically start with UC and are 24 characters
			if strings.HasPrefix(channelID, "UC") && len(channelID) >= 24 {
				return channelID, nil
			}
		}
	}
	
	// Handle @username format: youtube.com/@username
	if strings.Contains(channelURL, "/@") {
		parts := strings.Split(channelURL, "/@")
		if len(parts) > 1 {
			handle := strings.Split(parts[1], "/")[0]
			handle = strings.Split(handle, "?")[0]
			if handle != "" {
				// Use YouTube API to resolve handle to channel ID
				return c.resolveHandleToChannelID(handle)
			}
		}
	}
	
	// Handle /c/ChannelName format: youtube.com/c/ChannelName
	if strings.Contains(channelURL, "/c/") {
		parts := strings.Split(channelURL, "/c/")
		if len(parts) > 1 {
			username := strings.Split(parts[1], "/")[0]
			username = strings.Split(username, "?")[0]
			if username != "" {
				// Use YouTube API to resolve username to channel ID
				return c.resolveUsernameToChannelID(username)
			}
		}
	}
	
	return "", fmt.Errorf("unable to extract channel ID from URL: %s (unsupported format)", channelURL)
}

// resolveHandleToChannelID resolves a YouTube handle (@username) to a channel ID
func (c *Client) resolveHandleToChannelID(handle string) (string, error) {
	reqURL := fmt.Sprintf("%s/channels", c.baseURL)
	params := url.Values{}
	params.Set("key", c.apiKey)
	params.Set("part", "id")
	params.Set("forHandle", handle)
	
	reqURL += "?" + params.Encode()
	
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("failed to resolve handle: %w", err)
	}
	defer resp.Body.Close()
	
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to resolve handle: %s", string(bodyBytes))
	}
	
	var channelResp struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	
	if err := json.Unmarshal(bodyBytes, &channelResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	if len(channelResp.Items) == 0 {
		return "", fmt.Errorf("channel not found for handle: %s", handle)
	}
	
	return channelResp.Items[0].ID, nil
}

// resolveUsernameToChannelID resolves a YouTube username (/c/ChannelName) to a channel ID
func (c *Client) resolveUsernameToChannelID(username string) (string, error) {
	reqURL := fmt.Sprintf("%s/channels", c.baseURL)
	params := url.Values{}
	params.Set("key", c.apiKey)
	params.Set("part", "id")
	params.Set("forUsername", username)
	
	reqURL += "?" + params.Encode()
	
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("failed to resolve username: %w", err)
	}
	defer resp.Body.Close()
	
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to resolve username: %s", string(bodyBytes))
	}
	
	var channelResp struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	
	if err := json.Unmarshal(bodyBytes, &channelResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	if len(channelResp.Items) == 0 {
		return "", fmt.Errorf("channel not found for username: %s", username)
	}
	
	return channelResp.Items[0].ID, nil
}

