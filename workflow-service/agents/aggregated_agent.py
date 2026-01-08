"""Aggregated recommendation agent using Pydantic AI.

This agent processes multiple video analyses and recommendations to provide
an overall actionable recommendation based on the last 10 videos.
"""

import sys
import os
import logging

# Add parent directory to path for imports
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from typing import List
from pydantic_ai import Agent
from pydantic_ai.exceptions import AgentRunError
from openai import AsyncOpenAI, APIError
from models import Recommendation, MarketAnalysis, PortfolioContext, AggregatedRecommendation

logger = logging.getLogger(__name__)

# Initialize OpenAI client
client = AsyncOpenAI()

# Initialize the aggregated recommendation agent
aggregated_agent = Agent(
    'openai:gpt-5.2',
    system_prompt='You are a senior investment advisor expert. You analyze multiple market analyses and recommendations '
                  'from recent videos to provide a consolidated, actionable investment strategy. Your goal is to '
                  'identify patterns, trends, and consensus across multiple sources to provide the most reliable '
                  'overall recommendation. Consider risk management, diversification, and alignment with market conditions. '
                  'Return your consolidated recommendation as a JSON object with fields: '
                  'action (string - overall recommended action), confidence (float 0.0-1.0), '
                  'suggested_actions (list of objects with type, symbol, rationale), summary (string - detailed rationale), '
                  'and key_insights (list of strings - main insights from the aggregated analysis).',
    output_type=AggregatedRecommendation,
)


async def generate_aggregated_recommendation(
    market_analyses: List[MarketAnalysis],
    recommendations: List[Recommendation],
    portfolio_context: PortfolioContext = None
) -> AggregatedRecommendation:
    """Generate consolidated investment recommendation from multiple video analyses.
    
    Args:
        market_analyses: List of market analyses from recent videos
        recommendations: List of recommendations from recent videos
        portfolio_context: Current portfolio holdings for personalized recommendations
        
    Returns:
        AggregatedRecommendation with consolidated action, confidence, and insights
        
    Raises:
        AgentRunError: If the AI agent fails to generate valid output
        APIError: If OpenAI API call fails
    """
    if not market_analyses or not recommendations:
        raise ValueError("Both market_analyses and recommendations lists must be non-empty")
    
    if len(market_analyses) != len(recommendations):
        raise ValueError("market_analyses and recommendations must have the same length")
    
    # Build context for aggregated recommendation
    context_prompt = f'Based on analysis of {len(market_analyses)} recent videos, provide a consolidated investment recommendation:\n\n'
    
    context_prompt += 'RECENT MARKET ANALYSES:\n'
    for i, analysis in enumerate(market_analyses, 1):
        context_prompt += f'\nVideo {i}:\n'
        context_prompt += f'  Conditions: {analysis.conditions}\n'
        context_prompt += f'  Trends: {", ".join(analysis.trends) if analysis.trends else "None"}\n'
        context_prompt += f'  Risk Factors: {", ".join(analysis.risk_factors) if analysis.risk_factors else "None"}\n'
        context_prompt += f'  Summary: {analysis.summary}\n'
    
    context_prompt += '\n\nRECENT RECOMMENDATIONS:\n'
    for i, rec in enumerate(recommendations, 1):
        context_prompt += f'\nVideo {i}:\n'
        context_prompt += f'  Action: {rec.action}\n'
        context_prompt += f'  Confidence: {rec.confidence:.2f}\n'
        if rec.suggested_actions:
            context_prompt += '  Suggested Actions:\n'
            for action in rec.suggested_actions:
                context_prompt += f'    - {action.type.upper()} {action.symbol}: {action.rationale}\n'
        if rec.summary:
            context_prompt += f'  Summary: {rec.summary}\n'
    
    if portfolio_context and portfolio_context.holdings:
        context_prompt += '\n\nCURRENT PORTFOLIO:\n'
        context_prompt += f'Total Value: ${portfolio_context.total_value or 0:,.2f}\n'
        context_prompt += 'Holdings:\n'
        for holding in portfolio_context.holdings:
            context_prompt += f'  - {holding.symbol}: {holding.quantity} (${holding.value:,.2f})\n'
        context_prompt += '\n'
    
    context_prompt += '\nBased on these multiple analyses, provide a consolidated recommendation that:\n'
    context_prompt += '1) Identifies overall market consensus and patterns\n'
    context_prompt += '2) Provides a single actionable recommendation (action type)\n'
    context_prompt += '3) Assigns a confidence level based on agreement across sources\n'
    context_prompt += '4) Suggests specific actions for each relevant asset\n'
    context_prompt += '5) Summarizes key insights and rationale\n'
    context_prompt += '6) Highlights any conflicting signals or areas of uncertainty\n'
    
    try:
        result = await aggregated_agent.run(context_prompt)
        
        if not result.output:
            raise AgentRunError("Agent returned empty output")
        
        return result.output
    except AgentRunError as e:
        logger.error(f"Agent error during aggregated recommendation generation: {str(e)}", exc_info=True)
        raise
    except APIError as e:
        logger.error(f"OpenAI API error during aggregated recommendation generation: {str(e)}", exc_info=True)
        raise AgentRunError(f"OpenAI API error: {str(e)}") from e
    except Exception as e:
        logger.error(f"Unexpected error during aggregated recommendation generation: {str(e)}", exc_info=True)
        raise AgentRunError(f"Unexpected error: {str(e)}") from e

