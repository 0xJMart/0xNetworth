"""Market analysis agent using Pydantic AI."""

import sys
import os
import logging

# Add parent directory to path for imports
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from pydantic_ai import Agent
from pydantic_ai.exceptions import AgentRunError
from openai import AsyncOpenAI, APIError
from models import MarketAnalysis, PortfolioContext

logger = logging.getLogger(__name__)

# Initialize OpenAI client
client = AsyncOpenAI()

# Initialize the agent with OpenAI model
# Fixed: Changed result_type to output_type (correct parameter name)
analysis_agent = Agent(
    'openai:gpt-5.2',
    system_prompt='You are a financial market analyst expert. Analyze video transcripts about cryptocurrency '
                  'and financial markets. Identify market conditions, trends, and risk factors. '
                  'Provide clear, structured analysis of market conditions based on the transcript content. '
                  'Return your analysis as a JSON object with fields: conditions (string), trends (list of strings), '
                  'risk_factors (list of strings), and summary (string).',
    output_type=MarketAnalysis,
)


async def analyze_market(transcript_text: str, portfolio_context: PortfolioContext = None) -> MarketAnalysis:
    """Analyze market conditions from transcript.
    
    Args:
        transcript_text: Full transcript text from video
        portfolio_context: Optional portfolio context for personalized analysis
        
    Returns:
        MarketAnalysis with conditions, trends, risk factors, and summary
        
    Raises:
        AgentRunError: If the AI agent fails to generate valid output
        APIError: If OpenAI API call fails
    """
    # Build context for the analysis
    context_prompt = 'Analyze the following video transcript for market conditions, trends, and risk factors:\n\n'
    context_prompt += f'TRANSCRIPT:\n{transcript_text}\n\n'
    
    if portfolio_context and portfolio_context.holdings:
        context_prompt += 'PORTFOLIO CONTEXT:\n'
        context_prompt += f'Total Value: ${portfolio_context.total_value or 0:,.2f}\n'
        context_prompt += 'Holdings:\n'
        for holding in portfolio_context.holdings:
            context_prompt += f'  - {holding.symbol}: {holding.quantity} (${holding.value:,.2f})\n'
        context_prompt += '\n'
    
    context_prompt += 'Provide a structured analysis including: '
    context_prompt += '1) Overall market conditions (bullish, bearish, or neutral), '
    context_prompt += '2) Key trends identified, '
    context_prompt += '3) Risk factors mentioned, '
    context_prompt += '4) A detailed summary of market conditions.'
    
    try:
        result = await analysis_agent.run(context_prompt)
        
        # Fixed: Changed from result.data to result.output (correct attribute name)
        if not result.output:
            raise AgentRunError("Agent returned empty output")
        
        return result.output
    except AgentRunError as e:
        logger.error(f"Agent error during market analysis: {str(e)}", exc_info=True)
        raise
    except APIError as e:
        logger.error(f"OpenAI API error during market analysis: {str(e)}", exc_info=True)
        raise AgentRunError(f"OpenAI API error: {str(e)}") from e
    except Exception as e:
        logger.error(f"Unexpected error during market analysis: {str(e)}", exc_info=True)
        raise AgentRunError(f"Unexpected error: {str(e)}") from e

