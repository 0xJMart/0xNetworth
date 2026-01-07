package models

// VideoTranscript represents a YouTube video transcript
type VideoTranscript struct {
	ID          string `json:"id"`
	VideoID     string `json:"video_id"`
	VideoTitle  string `json:"video_title"`
	VideoURL    string `json:"video_url"`
	Text        string `json:"text"`
	Duration    *int   `json:"duration,omitempty"` // Duration in seconds
	SourceID    string `json:"source_id,omitempty"` // Reference to YouTubeSource
	CreatedAt   string `json:"created_at,omitempty"` // ISO 8601 timestamp
}


