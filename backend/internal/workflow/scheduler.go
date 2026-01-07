package workflow

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"0xnetworth/backend/internal/integrations/youtube"
	"0xnetworth/backend/internal/models"
	"0xnetworth/backend/internal/store"
)

// Scheduler manages scheduled workflow executions
type Scheduler struct {
	store       *store.Store
	engine      *Engine
	cron        *cron.Cron
	enabled     bool
	youtubeClient *youtube.Client
}

// NewScheduler creates a new workflow scheduler
func NewScheduler(store *store.Store, engine *Engine) *Scheduler {
	enabled := os.Getenv("WORKFLOW_SCHEDULE_ENABLED")
	if enabled == "" || enabled == "true" {
		enabled = "true"
	}
	
	// Initialize YouTube client if API key is provided
	youtubeAPIKey := os.Getenv("YOUTUBE_API_KEY")
	var youtubeClient *youtube.Client
	if youtubeAPIKey != "" {
		youtubeClient = youtube.NewClient(youtubeAPIKey)
		log.Println("YouTube API client initialized")
	} else {
		log.Println("Warning: YOUTUBE_API_KEY not set. Channel polling will be disabled.")
	}
	
	s := &Scheduler{
		store:        store,
		engine:       engine,
		cron:         cron.New(),
		enabled:      enabled == "true",
		youtubeClient: youtubeClient,
	}
	
	if s.enabled {
		s.setupSchedules()
	}
	
	return s
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	if !s.enabled {
		log.Println("Workflow scheduler is disabled")
		return
	}
	
	log.Println("Starting workflow scheduler...")
	s.cron.Start()
	log.Println("Workflow scheduler started")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	if !s.enabled {
		return
	}
	
	log.Println("Stopping workflow scheduler...")
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("Workflow scheduler stopped")
}

// setupSchedules sets up cron jobs for each enabled YouTube source
func (s *Scheduler) setupSchedules() {
	sources := s.store.GetAllYouTubeSources()
	
	for _, source := range sources {
		if !source.Enabled {
			continue
		}
		
		// Use source-specific schedule if available, otherwise use default
		schedule := source.Schedule
		if schedule == "" {
			schedule = os.Getenv("WORKFLOW_DEFAULT_SCHEDULE")
			if schedule == "" {
				schedule = "0 9 * * *" // Default: daily at 9 AM
			}
		}
		
		// Create closure to capture source
		sourceID := source.ID
		sourceURL := source.URL
		
		_, err := s.cron.AddFunc(schedule, func() {
			log.Printf("Scheduled execution triggered for source: %s (%s)", source.Name, sourceID)
			s.executeSource(sourceID, sourceURL)
		})
		
		if err != nil {
			log.Printf("Error scheduling source %s: %v", sourceID, err)
			continue
		}
		
		log.Printf("Scheduled source %s (%s) with schedule: %s", source.Name, sourceID, schedule)
	}
}

// executeSource executes workflow for a YouTube source
func (s *Scheduler) executeSource(sourceID string, sourceURL string) {
	log.Printf("Executing workflow for source %s: %s", sourceID, sourceURL)
	
	source, exists := s.store.GetYouTubeSourceByID(sourceID)
	if !exists {
		log.Printf("Source %s not found", sourceID)
		return
	}
	
	// If YouTube client is not available or source is not a channel, fall back to direct URL processing
	if s.youtubeClient == nil || source.Type != models.YouTubeSourceTypeChannel {
		log.Printf("Processing source URL directly (YouTube client not available or not a channel)")
		execution, err := s.engine.ExecuteWorkflow(sourceURL, sourceID)
		if err != nil {
			log.Printf("Error executing workflow for source %s: %v", sourceID, err)
			return
		}
		
		if execution.CompletedAt != "" {
			source.LastProcessed = execution.CompletedAt
			s.store.CreateOrUpdateYouTubeSource(source)
		}
		
		log.Printf("Workflow execution completed for source %s: %s", sourceID, execution.ID)
		return
	}
	
	// Extract channel ID from URL using YouTube client
	var channelID string
	var err error
	if s.youtubeClient != nil {
		channelID, err = s.youtubeClient.ExtractChannelID(sourceURL)
		if err != nil {
			log.Printf("Could not extract channel ID from URL %s: %v", sourceURL, err)
		}
	} else {
		// Fallback to simple extraction if client not available
		channelID = s.extractChannelIDFromURL(sourceURL)
	}
	
	if channelID == "" {
		// If we can't extract channel ID, try to use the stored channel_id
		if source.ChannelID != "" {
			channelID = source.ChannelID
		} else {
			log.Printf("Could not extract channel ID from URL %s, falling back to direct processing", sourceURL)
			execution, err := s.engine.ExecuteWorkflow(sourceURL, sourceID)
			if err != nil {
				log.Printf("Error executing workflow for source %s: %v", sourceID, err)
				return
			}
			if execution.CompletedAt != "" {
				source.LastProcessed = execution.CompletedAt
				s.store.CreateOrUpdateYouTubeSource(source)
			}
			return
		}
	}
	
	// Store the resolved channel ID for future use
	if source.ChannelID != channelID {
		source.ChannelID = channelID
		s.store.CreateOrUpdateYouTubeSource(source)
		log.Printf("Resolved channel ID for source %s: %s", sourceID, channelID)
	}
	
	// Determine publishedAfter time from last processed timestamp
	var publishedAfter *time.Time
	if source.LastProcessed != "" {
		parsed, err := time.Parse(time.RFC3339, source.LastProcessed)
		if err == nil {
			publishedAfter = &parsed
		}
	}
	
	// Fetch videos from channel
	// Add rate limiting: wait 100ms before API call to avoid quota issues
	// YouTube API allows 10,000 units/day, each search costs 100 units
	// This simple delay helps prevent rapid quota consumption
	time.Sleep(100 * time.Millisecond)
	
	videos, err := s.youtubeClient.GetChannelVideos(channelID, 50, publishedAfter)
	if err != nil {
		// Log quota-related errors specifically
		if apiErr, ok := err.(*youtube.APIError); ok && apiErr.StatusCode == http.StatusForbidden {
			log.Printf("YouTube API quota exceeded or invalid key for channel %s: %v", channelID, err)
		} else {
			log.Printf("Error fetching videos from channel %s: %v", channelID, err)
		}
		return
	}
	
	if len(videos) == 0 {
		log.Printf("No new videos found for channel %s", channelID)
		// Update last processed time even if no new videos
		now := time.Now().UTC().Format(time.RFC3339)
		source.LastProcessed = now
		s.store.CreateOrUpdateYouTubeSource(source)
		return
	}
	
	log.Printf("Found %d videos from channel %s", len(videos), channelID)
	
	// Get already processed video IDs for this source (optimized)
	processedVideoIDs := s.getProcessedVideoIDs(sourceID)
	
	// Process each new video
	processedCount := 0
	latestProcessedTime := source.LastProcessed
	
	for _, video := range videos {
		// Skip if already processed
		if processedVideoIDs[video.ID] {
			log.Printf("Skipping already processed video: %s (%s)", video.ID, video.Title)
			continue
		}
		
		// Build YouTube URL for the video
		videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.ID)
		
		log.Printf("Processing new video: %s (%s)", video.ID, video.Title)
		execution, err := s.engine.ExecuteWorkflow(videoURL, sourceID)
		if err != nil {
			log.Printf("Error executing workflow for video %s: %v", video.ID, err)
			continue
		}
		
		processedCount++
		
		// Update latest processed time
		if execution.CompletedAt != "" {
			latestProcessedTime = execution.CompletedAt
		} else if execution.StartedAt != "" {
			latestProcessedTime = execution.StartedAt
		}
		
		log.Printf("Workflow execution completed for video %s: %s", video.ID, execution.ID)
	}
	
	// Update source last processed time
	if latestProcessedTime != "" {
		source.LastProcessed = latestProcessedTime
		s.store.CreateOrUpdateYouTubeSource(source)
	}
	
	log.Printf("Processed %d new videos from source %s", processedCount, sourceID)
}

// extractChannelIDFromURL extracts channel ID from YouTube URL
func (s *Scheduler) extractChannelIDFromURL(url string) string {
	// Pattern: https://www.youtube.com/channel/UC...
	if strings.Contains(url, "/channel/") {
		parts := strings.Split(url, "/channel/")
		if len(parts) > 1 {
			channelID := strings.Split(parts[1], "/")[0]
			channelID = strings.Split(channelID, "?")[0]
			return channelID
		}
	}
	return ""
}

// getProcessedVideoIDs returns a map of already processed video IDs for a specific source
// This is optimized to only check executions from the same source
func (s *Scheduler) getProcessedVideoIDs(sourceID string) map[string]bool {
	executions := s.store.GetWorkflowExecutionsBySourceID(sourceID)
	processed := make(map[string]bool)
	
	for _, exec := range executions {
		if exec.VideoID != "" {
			processed[exec.VideoID] = true
		}
	}
	
	return processed
}

// TriggerSourceManually triggers a workflow execution for a source immediately
func (s *Scheduler) TriggerSourceManually(sourceID string) error {
	source, exists := s.store.GetYouTubeSourceByID(sourceID)
	if !exists {
		return &SourceNotFoundError{SourceID: sourceID}
	}
	
	if !source.Enabled {
		return &SourceDisabledError{SourceID: sourceID}
	}
	
	go s.executeSource(sourceID, source.URL)
	return nil
}

// SourceNotFoundError represents an error when a source is not found
type SourceNotFoundError struct {
	SourceID string
}

func (e *SourceNotFoundError) Error() string {
	return "source not found: " + e.SourceID
}

// SourceDisabledError represents an error when a source is disabled
type SourceDisabledError struct {
	SourceID string
}

func (e *SourceDisabledError) Error() string {
	return "source is disabled: " + e.SourceID
}

