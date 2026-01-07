package models

// YouTubeSourceType represents the type of YouTube source
type YouTubeSourceType string

const (
	YouTubeSourceTypeChannel  YouTubeSourceType = "channel"
	YouTubeSourceTypePlaylist YouTubeSourceType = "playlist"
)

// YouTubeSource represents a YouTube channel or playlist to monitor
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
	CreatedAt   string            `json:"created_at,omitempty"` // ISO 8601 timestamp
}


