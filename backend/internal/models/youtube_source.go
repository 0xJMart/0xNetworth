package models

// YouTubeSourceType represents the type of YouTube source
type YouTubeSourceType string

const (
	YouTubeSourceTypeChannel  YouTubeSourceType = "channel"
	YouTubeSourceTypePlaylist YouTubeSourceType = "playlist"
	YouTubeSourceTypeWebScraper YouTubeSourceType = "web_scraper"
)

// YouTubeSource represents a YouTube channel, playlist, or web scraper source to monitor
type YouTubeSource struct {
	ID          string            `json:"id"`
	Type        YouTubeSourceType `json:"type"`
	URL         string            `json:"url"`
	Name        string            `json:"name"`
	ChannelID   string            `json:"channel_id,omitempty"`
	PlaylistID  string            `json:"playlist_id,omitempty"`
	Enabled     bool              `json:"enabled"`
	Schedule    string            `json:"schedule,omitempty"` // Cron expression
	LastProcessed string          `json:"last_processed,omitempty"` // ISO 8601 timestamp
	AuthEmail   string            `json:"auth_email,omitempty"` // Email for web scraper authentication
	AuthSessionCookie string      `json:"auth_session_cookie,omitempty"` // Encrypted session cookie
	AuthLastRefreshed string       `json:"auth_last_refreshed,omitempty"` // ISO 8601 timestamp
	CreatedAt   string            `json:"created_at,omitempty"` // ISO 8601 timestamp
}


