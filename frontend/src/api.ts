import {
  Investment,
  InvestmentsResponse,
  NetWorth,
  NetWorthBreakdown,
  Platform,
  PlatformInvestmentsResponse,
  Portfolio,
  PortfoliosResponse,
  PlatformPortfoliosResponse,
  WorkflowExecution,
  ExecuteWorkflowRequest,
  WorkflowExecutionDetails,
  YouTubeSource,
  CreateYouTubeSourceRequest,
  UpdateSourceScheduleRequest,
} from './types';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

async function fetchAPI<T>(endpoint: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`);
  if (!response.ok) {
    throw new Error(`Failed to fetch ${endpoint}: ${response.statusText}`);
  }
  return response.json();
}

async function postAPI<T>(endpoint: string, body?: any): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!response.ok) {
    const contentType = response.headers.get('content-type');
    let errorMessage = `Failed to post ${endpoint}: ${response.statusText}`;
    
    try {
      if (contentType && contentType.includes('application/json')) {
        const errorJson = await response.json();
        // Try to extract a meaningful error message
        if (errorJson.error) {
          errorMessage = errorJson.error;
        } else if (errorJson.detail) {
          errorMessage = errorJson.detail;
        } else if (errorJson.message) {
          errorMessage = errorJson.message;
        } else {
          errorMessage = `${errorMessage} - ${JSON.stringify(errorJson)}`;
        }
      } else {
        const errorText = await response.text();
        if (errorText) {
          errorMessage = `${errorMessage} - ${errorText}`;
        }
      }
    } catch {
      // If parsing fails, use the default error message
    }
    
    throw new Error(errorMessage);
  }
  return response.json();
}

// Portfolio API
export async function fetchPortfolios(): Promise<Portfolio[]> {
  const data: PortfoliosResponse = await fetchAPI('/portfolios');
  return data.portfolios || [];
}

export async function fetchPortfolio(id: string): Promise<Portfolio> {
  return fetchAPI(`/portfolios/${id}`);
}

export async function fetchPortfoliosByPlatform(platform: Platform): Promise<Portfolio[]> {
  const data: PlatformPortfoliosResponse = await fetchAPI(`/portfolios/platform/${platform}`);
  return data.portfolios || [];
}

// Investment API
export async function fetchInvestments(): Promise<Investment[]> {
  const data: InvestmentsResponse = await fetchAPI('/investments');
  return data.investments || [];
}

export async function fetchInvestmentsByPortfolio(portfolioId: string): Promise<Investment[]> {
  const data: InvestmentsResponse = await fetchAPI(`/investments/portfolio/${portfolioId}`);
  return data.investments || [];
}

export async function fetchInvestmentsByPlatform(platform: Platform): Promise<Investment[]> {
  const data: PlatformInvestmentsResponse = await fetchAPI(`/investments/platform/${platform}`);
  return data.investments || [];
}

// Net Worth API
export async function fetchNetWorth(): Promise<NetWorth> {
  return fetchAPI('/networth');
}

export async function fetchNetWorthBreakdown(): Promise<NetWorthBreakdown> {
  return fetchAPI('/networth/breakdown');
}

// Sync API
export async function syncAll(): Promise<{ message: string; last_sync: string }> {
  return postAPI('/sync');
}

export async function syncPlatform(platform: Platform): Promise<{ message: string; platform: string }> {
  return postAPI(`/sync/${platform}`);
}

// Workflow API
/**
 * Converts a YouTube video ID or URL to a full YouTube URL
 * @param input - Video ID (11 chars) or YouTube URL
 * @returns Full YouTube URL
 */
function normalizeYouTubeInput(input: string): string {
  const trimmed = input.trim();
  
  // Check if it's already a URL
  if (trimmed.startsWith('http://') || trimmed.startsWith('https://')) {
    return trimmed;
  }
  
  // Check if it's a video ID (11 alphanumeric characters, may include - and _)
  const videoIdPattern = /^[a-zA-Z0-9_-]{11}$/;
  if (videoIdPattern.test(trimmed)) {
    return `https://www.youtube.com/watch?v=${trimmed}`;
  }
  
  // If it doesn't match either pattern, return as-is (backend will validate)
  return trimmed;
}

export async function executeWorkflow(videoIdOrUrl: string, sourceId?: string): Promise<WorkflowExecution> {
  const youtubeUrl = normalizeYouTubeInput(videoIdOrUrl);
  const request: ExecuteWorkflowRequest = {
    youtube_url: youtubeUrl,
    source_id: sourceId,
  };
  return postAPI<WorkflowExecution>('/workflow/execute', request);
}

export async function getWorkflowExecution(id: string): Promise<WorkflowExecution> {
  return fetchAPI<WorkflowExecution>(`/workflow/executions/${id}`);
}

export async function getWorkflowExecutions(): Promise<WorkflowExecution[]> {
  const executions = await fetchAPI<WorkflowExecution[]>('/workflow/executions');
  return executions || [];
}

export async function getWorkflowExecutionDetails(id: string): Promise<WorkflowExecutionDetails> {
  return fetchAPI<WorkflowExecutionDetails>(`/workflow/executions/${id}/details`);
}

// Recommendations Summary API
export interface SuggestedAction {
  type: string;
  symbol: string;
  rationale: string;
}

export interface AggregatedRecommendation {
  action: string;
  confidence: number;
  suggested_actions: SuggestedAction[];
  summary: string;
  key_insights: string[];
}

export interface RecommendationsSummary {
  total_count: number;
  action_distribution: Record<string, number>;
  average_confidence: number;
  condition_distribution: Record<string, number>;
  recent_recommendations: RecommendationSummaryItem[];
  aggregated_recommendation?: AggregatedRecommendation; // AI-generated consolidated recommendation
}

export interface RecommendationSummaryItem {
  execution_id: string;
  video_title: string;
  video_id: string;
  action: string;
  confidence: number;
  condition: string;
  completed_at: string;
}

export async function getRecommendationsSummary(days: number = 7): Promise<RecommendationsSummary> {
  return fetchAPI<RecommendationsSummary>(`/workflow/recommendations/summary?days=${days}`);
}

export async function generateAggregatedRecommendation(): Promise<AggregatedRecommendation> {
  const response = await fetch(`${API_BASE_URL}/workflow/recommendations/aggregate`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
  });
  
  if (!response.ok) {
    const error = await response.text();
    throw new Error(error || 'Failed to generate aggregated recommendation');
  }
  
  return response.json();
}

// YouTube Source Management API

export async function getYouTubeSources(): Promise<YouTubeSource[]> {
  const sources = await fetchAPI<YouTubeSource[]>('/workflow/sources');
  return sources || [];
}

export async function getYouTubeSource(id: string): Promise<YouTubeSource> {
  return fetchAPI<YouTubeSource>(`/workflow/sources/${id}`);
}

export async function createYouTubeSource(source: CreateYouTubeSourceRequest): Promise<YouTubeSource> {
  return postAPI<YouTubeSource>('/workflow/sources', source);
}

export async function updateYouTubeSource(id: string, source: CreateYouTubeSourceRequest): Promise<YouTubeSource> {
  const response = await fetch(`${API_BASE_URL}/workflow/sources/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(source),
  });
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(error.error || `Failed to update source: ${response.statusText}`);
  }
  return response.json();
}

export async function deleteYouTubeSource(id: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/workflow/sources/${id}`, {
    method: 'DELETE',
  });
  if (!response.ok) {
    throw new Error(`Failed to delete source: ${response.statusText}`);
  }
}

export async function updateSourceSchedule(id: string, schedule: string): Promise<YouTubeSource> {
  const request: UpdateSourceScheduleRequest = { schedule };
  return postAPI<YouTubeSource>(`/workflow/sources/${id}/schedule`, request);
}

export interface TestYouTubeSourceResponse {
  success: boolean;
  channel_id?: string;
  message?: string;
  error?: string;
}

export async function testYouTubeSource(url: string): Promise<TestYouTubeSourceResponse> {
  const response = await fetch(`${API_BASE_URL}/workflow/sources/test`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ url }),
  });
  
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(error.error || `Failed to test source: ${response.statusText}`);
  }
  
  return response.json();
}

export interface TriggerAllSourcesResponse {
  success: boolean;
  message: string;
  triggered_sources: string[];
  count: number;
}

export async function triggerAllSources(): Promise<TriggerAllSourcesResponse> {
  const response = await fetch(`${API_BASE_URL}/workflow/sources/trigger-all`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
  });
  
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(error.error || `Failed to trigger sources: ${response.statusText}`);
  }
  
  return response.json();
}

