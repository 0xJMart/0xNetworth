"""YouTube transcript extraction tool."""

import re
from typing import Optional
from youtube_transcript_api import YouTubeTranscriptApi
from youtube_transcript_api._errors import TranscriptsDisabled, NoTranscriptFound, VideoUnavailable


def extract_video_id(url: str) -> str:
    """Extract video ID from YouTube URL."""
    patterns = [
        r'(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([a-zA-Z0-9_-]{11})',
        r'youtube\.com\/watch\?.*v=([a-zA-Z0-9_-]{11})',
    ]
    
    for pattern in patterns:
        match = re.search(pattern, str(url))
        if match:
            return match.group(1)
    
    # If no pattern matches, assume the input is already a video ID
    if len(url) == 11 and url.replace('-', '').replace('_', '').isalnum():
        return url
    
    raise ValueError(f"Invalid YouTube URL or video ID: {url}")


def get_video_metadata(video_id: str) -> dict:
    """Get basic video metadata (title, channel, etc.) from video ID.
    
    Note: This is a placeholder. In production, you might want to use
    the YouTube Data API v3 to get full metadata.
    """
    # For now, return basic info. Can be enhanced with YouTube Data API
    return {
        "video_id": video_id,
        "video_title": f"Video {video_id}",  # Placeholder - use YouTube API for real title
    }


def fetch_transcript(video_id: str) -> tuple[str, Optional[int]]:
    """Fetch transcript for a YouTube video.
    
    Args:
        video_id: YouTube video ID
        
    Returns:
        Tuple of (transcript_text, duration_in_seconds)
        
    Raises:
        ValueError: If video ID is invalid
        TranscriptsDisabled: If transcripts are disabled for the video
        NoTranscriptFound: If no transcript is available
        VideoUnavailable: If video is unavailable
    """
    try:
        # Create instance of YouTubeTranscriptApi (it's an instance method, not class method)
        # Fixed: Changed from YouTubeTranscriptApi.list_transcripts() to YouTubeTranscriptApi().list()
        yt_api = YouTubeTranscriptApi()
        
        # Try to get transcript (prefer English, but fallback to any available)
        transcript_list = yt_api.list(video_id)
        
        # Try to get English transcript first
        try:
            transcript = transcript_list.find_transcript(['en'])
        except NoTranscriptFound:
            # Fallback to any available transcript
            transcript = transcript_list.find_generated_transcript(['en'])
        
        # Fetch the actual transcript data
        transcript_data = transcript.fetch()
        
        # Combine all text segments
        # Fixed: FetchedTranscriptSnippet objects use attributes, not dictionary keys
        # Changed from item['text'] to item.text
        full_text = ' '.join([item.text for item in transcript_data])
        
        # Calculate duration (last item's start + duration)
        # Fixed: Changed from last_item['start'] to last_item.start
        duration = None
        if transcript_data:
            last_item = transcript_data[-1]
            duration = int(last_item.start + last_item.duration)
        
        return full_text, duration
        
    except TranscriptsDisabled:
        raise TranscriptsDisabled(f"Transcripts are disabled for video {video_id}")
    except NoTranscriptFound:
        raise NoTranscriptFound(f"No transcript found for video {video_id}")
    except VideoUnavailable:
        raise VideoUnavailable(f"Video {video_id} is unavailable")


def get_youtube_transcript(url: str) -> dict:
    """Main function to get YouTube video transcript with metadata.
    
    Args:
        url: YouTube URL or video ID
        
    Returns:
        Dictionary with video_id, video_title, text, and duration
    """
    video_id = extract_video_id(url)
    transcript_text, duration = fetch_transcript(video_id)
    metadata = get_video_metadata(video_id)
    
    return {
        "video_id": video_id,
        "video_title": metadata["video_title"],
        "text": transcript_text,
        "duration": duration,
    }


