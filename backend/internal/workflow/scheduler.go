package workflow

import (
	"log"
	"os"

	"github.com/robfig/cron/v3"
	"0xnetworth/backend/internal/store"
)

// Scheduler manages scheduled workflow executions
type Scheduler struct {
	store   *store.Store
	engine  *Engine
	cron    *cron.Cron
	enabled bool
}

// NewScheduler creates a new workflow scheduler
func NewScheduler(store *store.Store, engine *Engine) *Scheduler {
	enabled := os.Getenv("WORKFLOW_SCHEDULE_ENABLED")
	if enabled == "" || enabled == "true" {
		enabled = "true"
	}
	
	s := &Scheduler{
		store:   store,
		engine:  engine,
		cron:    cron.New(),
		enabled: enabled == "true",
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
	// For now, we'll process the source URL directly
	// In a full implementation, this would:
	// 1. Fetch videos from the channel/playlist
	// 2. Check which ones haven't been processed
	// 3. Process each new video
	
	log.Printf("Executing workflow for source %s: %s", sourceID, sourceURL)
	
	// Execute workflow for the source URL
	// Note: This is a simplified implementation
	// Full implementation would fetch channel/playlist videos first
	execution, err := s.engine.ExecuteWorkflow(sourceURL, sourceID)
	if err != nil {
		log.Printf("Error executing workflow for source %s: %v", sourceID, err)
		return
	}
	
	// Update source last processed time
	source, exists := s.store.GetYouTubeSourceByID(sourceID)
	if exists {
		source.LastProcessed = execution.CompletedAt
		s.store.CreateOrUpdateYouTubeSource(source)
	}
	
	log.Printf("Workflow execution completed for source %s: %s", sourceID, execution.ID)
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

