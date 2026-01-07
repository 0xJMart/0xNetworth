"""Investment recommendation agent using Pydantic AI."""

import sys
import os
import logging

# Add parent directory to path for imports
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from pydantic_ai import Agent
from pydantic_ai.exceptions import AgentError
from openai import AsyncOpenAI, APIError
from models import Recommendation, MarketAnalysis, PortfolioContext

logger = logging.getLogger(__name__)

# Initialize OpenAI client
client = AsyncOpenAI()

# Initialize the agent with OpenAI model
# Fixed: Changed result_type to output_type (correct parameter name)
recommendation_agent = Agent(
    'openai:gpt-4o',
    system_prompt='You are an investment advisor expert. Based on market analysis and portfolio context, '
                  'provide actionable investment recommendations. Consider risk management, diversification, '
                  'and alignment with market conditions. Provide specific, actionable recommendations '
                  'with confidence levels. Return your recommendation as a JSON object with fields: '
                  'action (string), confidence (float 0.0-1.0), suggested_actions (list of objects with '
                  'type, symbol, rationale), and summary (string).',
    output_type=Recommendation,
)


async def generate_recommendation(
    market_analysis: MarketAnalysis,
    portfolio_context: PortfolioContext = None
) -> Recommendation:
    """Generate investment recommendations based on market analysis.
    
    Args:
        market_analysis: Market condition analysis
        portfolio_context: Current portfolio holdings for personalized recommendations
        
    Returns:
        Recommendation with action type, confidence, and suggested actions
        
    Raises:
        AgentError: If the AI agent fails to generate valid output
        APIError: If OpenAI API call fails
    """
    # Build context for recommendation
    context_prompt = 'Based on the following market analysis, provide investment recommendations:\n\n'
    context_prompt += 'MARKET ANALYSIS:\n'
    context_prompt += f'Conditions: {market_analysis.conditions}\n'
    context_prompt += f'Trends: {", ".join(market_analysis.trends)}\n'
    context_prompt += f'Risk Factors: {", ".join(market_analysis.risk_factors)}\n'
    context_prompt += f'Summary: {market_analysis.summary}\n\n'
    
    if portfolio_context and portfolio_context.holdings:
        context_prompt += 'CURRENT PORTFOLIO:\n'
        context_prompt += f'Total Value: ${portfolio_context.total_value or 0:,.2f}\n'
        context_prompt += 'Holdings:\n'
        for holding in portfolio_context.holdings:
            context_prompt += f'  - {holding.symbol}: {holding.quantity} (${holding.value:,.2f})\n'
        context_prompt += '\n'
    
    context_prompt += 'Provide specific recommendations including: '
    context_prompt += '1) Overall action type (rebalance, hold, diversify, increase allocation, etc.), '
    context_prompt += '2) Confidence level (0.0 to 1.0), '
    context_prompt += '3) Specific suggested actions for each asset (type: increase/decrease/hold/add/remove, symbol, rationale), '
    context_prompt += '4) A summary of the recommendation rationale.'
    
    try:
        result = await recommendation_agent.run(context_prompt)
        
        # Fixed: Changed from result.data to result.output (correct attribute name)
        if not result.output:
            raise AgentError("Agent returned empty output")
        
        return result.output
    except AgentError as e:
        logger.error(f"Agent error during recommendation generation: {str(e)}", exc_info=True)
        raise
    except APIError as e:
        logger.error(f"OpenAI API error during recommendation generation: {str(e)}", exc_info=True)
        raise AgentError(f"OpenAI API error: {str(e)}") from e
    except Exception as e:
        logger.error(f"Unexpected error during recommendation generation: {str(e)}", exc_info=True)
        raise AgentError(f"Unexpected error: {str(e)}") from e

