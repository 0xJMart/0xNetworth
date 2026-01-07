"""Transcript extraction agent - direct tool usage."""

import sys
import os

# Add parent directory to path for imports
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from models import Transcript
from tools.youtube_tool import get_youtube_transcript


async def extract_transcript(youtube_url: str) -> Transcript:
    """Extract transcript from YouTube video.
    
    This function directly uses the YouTube tool to fetch transcripts.
    In a more complex setup, this could be wrapped in a Pydantic AI agent.
    
    Args:
        youtube_url: YouTube video URL
        
    Returns:
        Transcript model with video metadata and transcript text
    """
    transcript_data = get_youtube_transcript(youtube_url)
    
    return Transcript(
        video_id=transcript_data['video_id'],
        video_title=transcript_data['video_title'],
        text=transcript_data['text'],
        duration=transcript_data.get('duration'),
    )
