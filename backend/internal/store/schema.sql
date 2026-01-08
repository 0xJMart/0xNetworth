-- 0xNetworth Database Schema
-- PostgreSQL schema for persisting investment data, workflow executions, and analysis results

-- Portfolios table
CREATE TABLE IF NOT EXISTS portfolios (
    id VARCHAR(255) PRIMARY KEY,
    platform VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    last_synced TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Investments table
CREATE TABLE IF NOT EXISTS investments (
    id VARCHAR(255) PRIMARY KEY,
    account_id VARCHAR(255) NOT NULL,
    platform VARCHAR(50) NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    name VARCHAR(255),
    quantity DOUBLE PRECISION NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    asset_type VARCHAR(50),
    last_updated TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sync metadata table
CREATE TABLE IF NOT EXISTS sync_metadata (
    id VARCHAR(255) PRIMARY KEY,
    platform VARCHAR(50) NOT NULL UNIQUE,
    last_sync_time TIMESTAMP,
    sync_status VARCHAR(50),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- YouTube sources table
CREATE TABLE IF NOT EXISTS youtube_sources (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    url TEXT NOT NULL,
    name VARCHAR(255) NOT NULL,
    channel_id VARCHAR(255),
    playlist_id VARCHAR(255),
    enabled BOOLEAN DEFAULT true,
    schedule VARCHAR(255),
    last_processed TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Video transcripts table
CREATE TABLE IF NOT EXISTS video_transcripts (
    id VARCHAR(255) PRIMARY KEY,
    video_id VARCHAR(255) NOT NULL,
    video_title VARCHAR(500),
    video_url TEXT NOT NULL,
    text TEXT NOT NULL,
    duration INTEGER,
    source_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Market analyses table
CREATE TABLE IF NOT EXISTS market_analyses (
    id VARCHAR(255) PRIMARY KEY,
    transcript_id VARCHAR(255) NOT NULL,
    conditions VARCHAR(50) NOT NULL,
    trends JSONB,
    risk_factors JSONB,
    summary TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (transcript_id) REFERENCES video_transcripts(id) ON DELETE CASCADE
);

-- Recommendations table
CREATE TABLE IF NOT EXISTS recommendations (
    id VARCHAR(255) PRIMARY KEY,
    analysis_id VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    confidence DOUBLE PRECISION NOT NULL CHECK (confidence >= 0.0 AND confidence <= 1.0),
    suggested_actions JSONB,
    summary TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (analysis_id) REFERENCES market_analyses(id) ON DELETE CASCADE
);

-- Workflow executions table
CREATE TABLE IF NOT EXISTS workflow_executions (
    id VARCHAR(255) PRIMARY KEY,
    status VARCHAR(50) NOT NULL,
    video_id VARCHAR(255),
    video_url TEXT NOT NULL,
    video_title VARCHAR(500),
    source_id VARCHAR(255),
    transcript_id VARCHAR(255),
    analysis_id VARCHAR(255),
    recommendation_id VARCHAR(255),
    error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (transcript_id) REFERENCES video_transcripts(id) ON DELETE SET NULL,
    FOREIGN KEY (analysis_id) REFERENCES market_analyses(id) ON DELETE SET NULL,
    FOREIGN KEY (recommendation_id) REFERENCES recommendations(id) ON DELETE SET NULL
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_investments_account_id ON investments(account_id);
CREATE INDEX IF NOT EXISTS idx_investments_platform ON investments(platform);
CREATE INDEX IF NOT EXISTS idx_portfolios_platform ON portfolios(platform);
CREATE INDEX IF NOT EXISTS idx_workflow_executions_status ON workflow_executions(status);
CREATE INDEX IF NOT EXISTS idx_workflow_executions_source_id ON workflow_executions(source_id);
CREATE INDEX IF NOT EXISTS idx_video_transcripts_video_id ON video_transcripts(video_id);
CREATE INDEX IF NOT EXISTS idx_video_transcripts_source_id ON video_transcripts(source_id);
CREATE INDEX IF NOT EXISTS idx_market_analyses_transcript_id ON market_analyses(transcript_id);
CREATE INDEX IF NOT EXISTS idx_recommendations_analysis_id ON recommendations(analysis_id);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_portfolios_updated_at BEFORE UPDATE ON portfolios
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_investments_updated_at BEFORE UPDATE ON investments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sync_metadata_updated_at BEFORE UPDATE ON sync_metadata
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_youtube_sources_updated_at BEFORE UPDATE ON youtube_sources
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

