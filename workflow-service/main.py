"""FastAPI server for agentic workflow service."""

import os
import logging
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from dotenv import load_dotenv

from models import WorkflowRequest, WorkflowResponse
from agents import extract_transcript, analyze_market, generate_recommendation

# Load environment variables
load_dotenv()

# Configure logging
log_level = os.getenv('LOG_LEVEL', 'INFO').upper()
logging.basicConfig(
    level=getattr(logging, log_level, logging.INFO),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Initialize FastAPI app
app = FastAPI(
    title="0xNetworth Workflow Service",
    description="Agentic workflow service for YouTube video market analysis",
    version="1.0.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure appropriately for production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/health")
async def health_check():
    """Health check endpoint."""
    return {"status": "healthy", "service": "workflow-service"}


@app.post("/process", response_model=WorkflowResponse)
async def process_video(request: WorkflowRequest) -> WorkflowResponse:
    """Process YouTube video through the agentic workflow.
    
    This endpoint orchestrates the three-stage workflow:
    1. Extract transcript from YouTube video
    2. Analyze market conditions from transcript
    3. Generate investment recommendations
    
    Args:
        request: WorkflowRequest with YouTube URL and optional portfolio context
        
    Returns:
        WorkflowResponse with transcript, market analysis, and recommendations
        
    Raises:
        HTTPException: If any step of the workflow fails
    """
    try:
        logger.info(f"Processing video: {request.youtube_url}")
        
        # Stage 1: Extract transcript
        logger.info("Stage 1: Extracting transcript...")
        transcript = await extract_transcript(str(request.youtube_url))
        logger.info(f"Transcript extracted: {transcript.video_id}")
        
        # Stage 2: Analyze market conditions
        logger.info("Stage 2: Analyzing market conditions...")
        market_analysis = await analyze_market(
            transcript_text=transcript.text,
            portfolio_context=request.portfolio_context
        )
        logger.info(f"Market analysis complete: {market_analysis.conditions}")
        
        # Stage 3: Generate recommendations
        logger.info("Stage 3: Generating recommendations...")
        recommendation = await generate_recommendation(
            market_analysis=market_analysis,
            portfolio_context=request.portfolio_context
        )
        logger.info(f"Recommendation generated: {recommendation.action} (confidence: {recommendation.confidence})")
        
        # Return combined response
        return WorkflowResponse(
            transcript=transcript,
            market_analysis=market_analysis,
            recommendation=recommendation
        )
        
    except Exception as e:
        logger.error(f"Error processing video: {str(e)}", exc_info=True)
        raise HTTPException(
            status_code=500,
            detail=f"Failed to process video: {str(e)}"
        )


if __name__ == "__main__":
    import uvicorn
    
    port = int(os.getenv('PORT', 8000))
    uvicorn.run(app, host="0.0.0.0", port=port)


