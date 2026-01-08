"""Pydantic models for workflow request and response validation."""

from typing import List, Optional
from pydantic import BaseModel, Field, HttpUrl


class Holding(BaseModel):
    """Represents a portfolio holding."""
    symbol: str = Field(..., description="Asset symbol (e.g., BTC, ETH)")
    quantity: float = Field(..., description="Quantity held")
    value: float = Field(..., description="Current value in USD")


class PortfolioContext(BaseModel):
    """Portfolio context for market analysis."""
    holdings: List[Holding] = Field(default_factory=list, description="List of portfolio holdings")
    total_value: Optional[float] = Field(None, description="Total portfolio value in USD")


class WorkflowRequest(BaseModel):
    """Request model for workflow execution."""
    youtube_url: HttpUrl = Field(..., description="YouTube video URL to process")
    portfolio_context: Optional[PortfolioContext] = Field(None, description="Current portfolio context for analysis")


class Transcript(BaseModel):
    """Video transcript model."""
    video_id: str = Field(..., description="YouTube video ID")
    video_title: str = Field(..., description="Video title")
    text: str = Field(..., description="Full transcript text")
    duration: Optional[int] = Field(None, description="Video duration in seconds")


class MarketAnalysis(BaseModel):
    """Market condition analysis model."""
    conditions: str = Field(..., description="Overall market conditions (bullish, bearish, neutral)")
    trends: List[str] = Field(default_factory=list, description="Key market trends identified")
    risk_factors: List[str] = Field(default_factory=list, description="Risk factors identified")
    summary: str = Field(..., description="Detailed market analysis summary")


class SuggestedAction(BaseModel):
    """Individual suggested action."""
    type: str = Field(..., description="Action type (increase, decrease, hold, add, remove)")
    symbol: str = Field(..., description="Asset symbol affected")
    rationale: str = Field(..., description="Reasoning for this action")


class Recommendation(BaseModel):
    """Investment recommendation model."""
    action: str = Field(..., description="Overall recommended action (rebalance, hold, diversify, etc.)")
    confidence: float = Field(..., ge=0.0, le=1.0, description="Confidence level (0.0 to 1.0)")
    suggested_actions: List[SuggestedAction] = Field(default_factory=list, description="Specific suggested actions")
    summary: Optional[str] = Field(None, description="Recommendation summary")


class AggregatedRecommendation(BaseModel):
    """Aggregated recommendation from multiple video analyses."""
    action: str = Field(..., description="Overall consolidated recommended action")
    confidence: float = Field(..., ge=0.0, le=1.0, description="Confidence level based on consensus (0.0 to 1.0)")
    suggested_actions: List[SuggestedAction] = Field(default_factory=list, description="Specific suggested actions")
    summary: str = Field(..., description="Detailed consolidated recommendation summary")
    key_insights: List[str] = Field(default_factory=list, description="Key insights from aggregated analysis")


class AggregatedRecommendationRequest(BaseModel):
    """Request model for aggregated recommendation generation."""
    market_analyses: List[MarketAnalysis] = Field(..., description="List of market analyses from recent videos")
    recommendations: List[Recommendation] = Field(..., description="List of recommendations from recent videos")
    portfolio_context: Optional[PortfolioContext] = Field(None, description="Current portfolio context")


class WorkflowResponse(BaseModel):
    """Response model for workflow execution."""
    transcript: Transcript = Field(..., description="Video transcript data")
    market_analysis: MarketAnalysis = Field(..., description="Market condition analysis")
    recommendation: Recommendation = Field(..., description="Investment recommendations")


