package models

// WorkflowExecutionStatus represents the status of a workflow execution
type WorkflowExecutionStatus string

const (
	WorkflowStatusPending    WorkflowExecutionStatus = "pending"
	WorkflowStatusProcessing WorkflowExecutionStatus = "processing"
	WorkflowStatusCompleted  WorkflowExecutionStatus = "completed"
	WorkflowStatusFailed     WorkflowExecutionStatus = "failed"
)

// WorkflowExecution represents a workflow execution record
type WorkflowExecution struct {
	ID             string                  `json:"id"`
	Status         WorkflowExecutionStatus `json:"status"`
	VideoID        string                  `json:"video_id"`
	VideoURL       string                  `json:"video_url"`
	VideoTitle     string                  `json:"video_title,omitempty"`
	SourceID       string                  `json:"source_id,omitempty"` // Reference to YouTubeSource
	TranscriptID   string                  `json:"transcript_id,omitempty"`
	AnalysisID     string                  `json:"analysis_id,omitempty"`
	RecommendationID string                `json:"recommendation_id,omitempty"`
	Error          string                  `json:"error,omitempty"`
	CreatedAt      string                  `json:"created_at,omitempty"` // ISO 8601 timestamp
	StartedAt      string                  `json:"started_at,omitempty"` // ISO 8601 timestamp
	CompletedAt    string                  `json:"completed_at,omitempty"` // ISO 8601 timestamp
}


