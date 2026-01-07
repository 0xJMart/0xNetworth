export type Platform = 'coinbase';

export interface Portfolio {
  id: string;
  platform: Platform;
  name: string;
  type?: string;
  last_synced?: string;
}

export interface Investment {
  id: string;
  account_id: string;
  platform: Platform;
  symbol: string;
  name: string;
  quantity: number;
  value: number;
  price: number;
  currency: string;
  asset_type: string;
  last_updated?: string;
}

export interface NetWorth {
  total_value: number;
  currency: string;
  by_platform: Record<Platform, number>;
  by_asset_type: Record<string, number>;
  account_count: number;
  last_calculated: string;
}

export interface NetWorthBreakdown {
  networth: NetWorth;
  portfolios: Portfolio[];
  investments: Investment[];
}

export interface InvestmentsResponse {
  investments: Investment[];
}

export interface PlatformInvestmentsResponse {
  platform: Platform;
  investments: Investment[];
}

export interface PortfoliosResponse {
  portfolios: Portfolio[];
}

export interface PlatformPortfoliosResponse {
  platform: Platform;
  portfolios: Portfolio[];
}

// Workflow Types
export type WorkflowExecutionStatus = 'pending' | 'processing' | 'completed' | 'failed';

export interface WorkflowExecution {
  id: string;
  status: WorkflowExecutionStatus;
  video_id: string;
  video_url: string;
  video_title?: string;
  source_id?: string;
  transcript_id?: string;
  analysis_id?: string;
  recommendation_id?: string;
  error?: string;
  created_at?: string;
  started_at?: string;
  completed_at?: string;
}

export interface ExecuteWorkflowRequest {
  youtube_url: string;
  source_id?: string;
}

export interface VideoTranscript {
  id: string;
  video_id: string;
  video_title: string;
  video_url: string;
  text: string;
  duration?: number;
  source_id?: string;
  created_at?: string;
}

export interface MarketAnalysis {
  id: string;
  transcript_id: string;
  conditions: string;
  trends: string[];
  risk_factors: string[];
  summary: string;
  created_at?: string;
}

export interface SuggestedAction {
  type: string;
  symbol: string;
  rationale: string;
}

export interface Recommendation {
  id: string;
  analysis_id: string;
  action: string;
  confidence: number;
  suggested_actions: SuggestedAction[];
  summary?: string;
  created_at?: string;
}

export interface WorkflowExecutionDetails {
  execution: WorkflowExecution;
  transcript?: VideoTranscript;
  market_analysis?: MarketAnalysis;
  recommendation?: Recommendation;
}

