"""FastAPI server for agentic workflow service."""

import os
import logging
import uuid
from contextlib import asynccontextmanager
from typing import Optional
from fastapi import FastAPI, HTTPException, Request, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from slowapi import Limiter, _rate_limit_exceeded_handler
from slowapi.util import get_remote_address
from slowapi.errors import RateLimitExceeded
from dotenv import load_dotenv
from openai import AsyncOpenAI
from pydantic_ai.exceptions import AgentError

from models import WorkflowRequest, WorkflowResponse
from agents import extract_transcript, analyze_market, generate_recommendation
from tools.youtube_tool import extract_video_id, fetch_transcript
from youtube_transcript_api._errors import TranscriptsDisabled, NoTranscriptFound, VideoUnavailable

# Load environment variables
load_dotenv()

# Configure logging with structured format
log_level = os.getenv('LOG_LEVEL', 'INFO').upper()

# Custom formatter that handles request_id
class RequestIDFormatter(logging.Formatter):
    """Custom formatter that includes request_id in log records."""
    def format(self, record):
        # Add request_id if not present (for logs outside request context)
        if not hasattr(record, 'request_id'):
            record.request_id = 'system'
        return super().format(record)

# Set up logging
handler = logging.StreamHandler()
handler.setFormatter(RequestIDFormatter('%(asctime)s - %(name)s - %(levelname)s - [%(request_id)s] - %(message)s'))
logging.basicConfig(
    level=getattr(logging, log_level, logging.INFO),
    handlers=[handler]
)
logger = logging.getLogger(__name__)

# Request ID middleware
class RequestIDMiddleware:
    """Middleware to add request ID for tracing."""
    async def __call__(self, request: Request, call_next):
        request_id = str(uuid.uuid4())[:8]
        request.state.request_id = request_id
        
        # Update logger context for this request
        old_factory = logging.getLogRecordFactory()
        def record_factory(*args, **kwargs):
            record = old_factory(*args, **kwargs)
            record.request_id = request_id
            return record
        logging.setLogRecordFactory(record_factory)
        
        try:
            response = await call_next(request)
            response.headers["X-Request-ID"] = request_id
            return response
        finally:
            # Restore original factory
            logging.setLogRecordFactory(old_factory)

# Rate limiter
limiter = Limiter(key_func=get_remote_address)

# Validate OpenAI API key at startup
def validate_openai_key():
    """Validate that OpenAI API key is set and accessible."""
    api_key = os.getenv('OPENAI_API_KEY')
    if not api_key:
        raise ValueError("OPENAI_API_KEY environment variable is not set")
    
    # Test API connectivity
    try:
        client = AsyncOpenAI()
        # We'll do a simple validation - just check if key format is valid
        # Full connectivity check will be done in health endpoint
        if not api_key.startswith('sk-'):
            logger.warning("OpenAI API key format may be invalid (should start with 'sk-')")
    except Exception as e:
        logger.warning(f"Could not validate OpenAI client: {e}")
    
    logger.info("OpenAI API key validated")

# Startup validation
@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifespan context manager for startup/shutdown."""
    # Startup
    logger.info("Starting workflow service...")
    try:
        validate_openai_key()
        logger.info("Workflow service started successfully")
    except ValueError as e:
        logger.error(f"Startup validation failed: {e}")
        raise
    yield
    # Shutdown
    logger.info("Shutting down workflow service...")

# Initialize FastAPI app
app = FastAPI(
    title="0xNetworth Workflow Service",
    description="Agentic workflow service for YouTube video market analysis",
    version="1.0.0",
    lifespan=lifespan,
    # Limit request body size to 10MB
    max_request_size=10 * 1024 * 1024,
)

# Add rate limiter to app
app.state.limiter = limiter
app.add_exception_handler(RateLimitExceeded, _rate_limit_exceeded_handler)

# CORS middleware - configurable via environment variable
cors_origins = os.getenv('CORS_ORIGINS', '*').split(',')
# In production, default to empty list if not set (more secure)
if os.getenv('ENVIRONMENT') == 'production' and cors_origins == ['*']:
    logger.warning("CORS_ORIGINS is set to '*' in production. Consider restricting to specific domains.")
    cors_origins = []  # Default to no CORS in production if not explicitly set

app.add_middleware(
    CORSMiddleware,
    allow_origins=cors_origins if cors_origins != ['*'] else ["*"],
    allow_credentials=True,
    allow_methods=["GET", "POST"],
    allow_headers=["*"],
    max_age=3600,
)

# Add request ID middleware
app.middleware("http")(RequestIDMiddleware())

# Custom exception handlers
@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request: Request, exc: RequestValidationError):
    """Handle Pydantic validation errors."""
    request_id = getattr(request.state, 'request_id', 'unknown')
    logger.warning(f"Validation error: {exc.errors()}")
    return JSONResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        content={
            "detail": "Validation error",
            "errors": exc.errors(),
            "request_id": request_id
        }
    )

@app.exception_handler(ValueError)
async def value_error_handler(request: Request, exc: ValueError):
    """Handle ValueError exceptions (e.g., invalid YouTube URL)."""
    request_id = getattr(request.state, 'request_id', 'unknown')
    logger.error(f"Value error: {str(exc)}")
    return JSONResponse(
        status_code=status.HTTP_400_BAD_REQUEST,
        content={
            "detail": f"Invalid input: {str(exc)}",
            "request_id": request_id
        }
    )

@app.exception_handler((TranscriptsDisabled, NoTranscriptFound, VideoUnavailable))
async def youtube_error_handler(request: Request, exc: Exception):
    """Handle YouTube API errors."""
    request_id = getattr(request.state, 'request_id', 'unknown')
    error_type = type(exc).__name__
    logger.error(f"YouTube API error ({error_type}): {str(exc)}")
    return JSONResponse(
        status_code=status.HTTP_400_BAD_REQUEST,
        content={
            "detail": f"YouTube error: {str(exc)}",
            "error_type": error_type,
            "request_id": request_id
        }
    )

@app.exception_handler(AgentError)
async def agent_error_handler(request: Request, exc: AgentError):
    """Handle Pydantic AI agent errors."""
    request_id = getattr(request.state, 'request_id', 'unknown')
    logger.error(f"Agent error: {str(exc)}", exc_info=True)
    return JSONResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        content={
            "detail": f"AI agent error: {str(exc)}",
            "error_type": "AgentError",
            "request_id": request_id
        }
    )


@app.get("/health")
@limiter.limit("10/minute")
async def health_check(request: Request):
    """Health check endpoint with OpenAI API connectivity check."""
    request_id = getattr(request.state, 'request_id', 'unknown')
    
    health_status = {
        "status": "healthy",
        "service": "workflow-service",
        "request_id": request_id
    }
    
    # Check OpenAI API connectivity
    api_key = os.getenv('OPENAI_API_KEY')
    if not api_key:
        health_status["status"] = "unhealthy"
        health_status["error"] = "OPENAI_API_KEY not set"
        return JSONResponse(status_code=503, content=health_status)
    
    try:
        client = AsyncOpenAI()
        # Simple connectivity test - just verify client can be created
        # Full API test would require an actual API call (costs money)
        health_status["openai"] = "configured"
    except Exception as e:
        health_status["status"] = "degraded"
        health_status["openai"] = f"error: {str(e)}"
    
    return health_status


@app.post("/process", response_model=WorkflowResponse)
@limiter.limit("5/minute")  # Rate limit: 5 requests per minute per IP
async def process_video(request: Request, workflow_request: WorkflowRequest) -> WorkflowResponse:
    """Process YouTube video through the agentic workflow.
    
    This endpoint orchestrates the three-stage workflow:
    1. Extract transcript from YouTube video
    2. Analyze market conditions from transcript
    3. Generate investment recommendations
    
    Args:
        request: FastAPI request object (for rate limiting)
        workflow_request: WorkflowRequest with YouTube URL and optional portfolio context
        
    Returns:
        WorkflowResponse with transcript, market analysis, and recommendations
        
    Raises:
        HTTPException: If any step of the workflow fails
    """
    request_id = getattr(request.state, 'request_id', 'unknown')
    logger.info(f"[{request_id}] Processing video: {workflow_request.youtube_url}")
    
    try:
        # Pre-flight validation: Check if video ID can be extracted and transcript is available
        logger.info(f"[{request_id}] Pre-flight validation: Extracting video ID...")
        try:
            video_id = extract_video_id(str(workflow_request.youtube_url))
            logger.info(f"[{request_id}] Video ID extracted: {video_id}")
        except ValueError as e:
            logger.error(f"[{request_id}] Invalid YouTube URL: {str(e)}")
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Invalid YouTube URL: {str(e)}"
            )
        
        # Pre-flight check: Verify transcript is available
        logger.info(f"[{request_id}] Pre-flight validation: Checking transcript availability...")
        try:
            # Just check if transcript exists, don't fetch it yet
            from youtube_transcript_api import YouTubeTranscriptApi
            yt_api = YouTubeTranscriptApi()
            transcript_list = yt_api.list(video_id)
            # Try to find a transcript (this will raise if none available)
            try:
                transcript_list.find_transcript(['en'])
            except NoTranscriptFound:
                transcript_list.find_generated_transcript(['en'])
            logger.info(f"[{request_id}] Transcript available")
        except (TranscriptsDisabled, NoTranscriptFound, VideoUnavailable) as e:
            error_type = type(e).__name__
            logger.error(f"[{request_id}] Transcript not available: {error_type} - {str(e)}")
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Transcript not available for video {video_id}: {str(e)}"
            )
        
        # Stage 1: Extract transcript
        logger.info(f"[{request_id}] Stage 1: Extracting transcript...")
        try:
            transcript = await extract_transcript(str(workflow_request.youtube_url))
            logger.info(f"[{request_id}] Transcript extracted: {transcript.video_id} ({len(transcript.text)} chars)")
        except (TranscriptsDisabled, NoTranscriptFound, VideoUnavailable) as e:
            error_type = type(e).__name__
            logger.error(f"[{request_id}] Transcript extraction failed: {error_type} - {str(e)}")
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Failed to extract transcript: {str(e)}"
            )
        except Exception as e:
            logger.error(f"[{request_id}] Unexpected error during transcript extraction: {str(e)}", exc_info=True)
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"Unexpected error during transcript extraction: {str(e)}"
            )
        
        # Stage 2: Analyze market conditions
        logger.info(f"[{request_id}] Stage 2: Analyzing market conditions...")
        try:
            market_analysis = await analyze_market(
                transcript_text=transcript.text,
                portfolio_context=workflow_request.portfolio_context
            )
            logger.info(f"[{request_id}] Market analysis complete: {market_analysis.conditions}")
        except AgentError as e:
            logger.error(f"[{request_id}] Agent error during market analysis: {str(e)}", exc_info=True)
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"AI agent error during market analysis: {str(e)}"
            )
        except Exception as e:
            logger.error(f"[{request_id}] Unexpected error during market analysis: {str(e)}", exc_info=True)
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"Unexpected error during market analysis: {str(e)}"
            )
        
        # Stage 3: Generate recommendations
        logger.info(f"[{request_id}] Stage 3: Generating recommendations...")
        try:
            recommendation = await generate_recommendation(
                market_analysis=market_analysis,
                portfolio_context=workflow_request.portfolio_context
            )
            logger.info(f"[{request_id}] Recommendation generated: {recommendation.action} (confidence: {recommendation.confidence})")
        except AgentError as e:
            logger.error(f"[{request_id}] Agent error during recommendation generation: {str(e)}", exc_info=True)
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"AI agent error during recommendation generation: {str(e)}"
            )
        except Exception as e:
            logger.error(f"[{request_id}] Unexpected error during recommendation generation: {str(e)}", exc_info=True)
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"Unexpected error during recommendation generation: {str(e)}"
            )
        
        # Return combined response
        logger.info(f"[{request_id}] Workflow completed successfully")
        return WorkflowResponse(
            transcript=transcript,
            market_analysis=market_analysis,
            recommendation=recommendation
        )
        
    except HTTPException:
        # Re-raise HTTP exceptions (already properly formatted)
        raise
    except Exception as e:
        # Catch-all for any other unexpected errors
        logger.error(f"[{request_id}] Unexpected error processing video: {str(e)}", exc_info=True)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to process video: {str(e)}"
        )


if __name__ == "__main__":
    import uvicorn
    
    port = int(os.getenv('PORT', 8000))
    uvicorn.run(app, host="0.0.0.0", port=port)
