package youtube

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
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
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
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
// - https://www.youtube.com/channel/UC...
// - https://www.youtube.com/c/ChannelName
// - https://www.youtube.com/@ChannelName
// - https://youtube.com/user/username
func ExtractChannelID(channelURL string) (string, error) {
	// For now, we'll require the full channel ID format
	// In a full implementation, we'd need to use the YouTube API to resolve
	// custom URLs and usernames to channel IDs
	// For simplicity, we'll extract from standard channel URL format
	
	// Pattern: youtube.com/channel/UC...
	// This is the most reliable format
	// TODO: Add support for resolving custom URLs via API
	
	return "", fmt.Errorf("channel ID extraction not yet implemented - please use full channel URL with channel ID")
}

