# YouTube Channel Configuration for Cron Jobs

This guide explains how to configure YouTube channels for automatic video polling via cron jobs.

## Prerequisites

1. **YouTube Data API v3 Key**: You need a valid `YOUTUBE_API_KEY` environment variable set in your backend deployment
2. **Scheduler Enabled**: The workflow scheduler must be enabled (set `WORKFLOW_SCHEDULE_ENABLED=true`)

## API Endpoints

The backend provides REST API endpoints to manage YouTube sources:

- `POST /api/workflow/sources` - Create a new YouTube source
- `GET /api/workflow/sources` - List all YouTube sources
- `GET /api/workflow/sources/:id` - Get a specific source
- `DELETE /api/workflow/sources/:id` - Delete a source
- `POST /api/workflow/sources/:id/schedule` - Update a source's schedule

## Creating a YouTube Channel Source

### Method 1: Using cURL

```bash
curl -X POST http://localhost:8080/api/workflow/sources \
  -H "Content-Type: application/json" \
  -d '{
    "type": "channel",
    "url": "https://www.youtube.com/channel/UCxxxxxxxxxxxxxxxxxxxxxx",
    "name": "My Investment Channel",
    "enabled": true,
    "schedule": "0 9 * * *"
  }'
```

### Method 2: Using the Frontend (if UI is added)

Currently, the frontend doesn't have a UI for managing sources, but you can add one or use the API directly.

### Method 3: Direct API Call from Browser Console

```javascript
fetch('/api/workflow/sources', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    type: 'channel',
    url: 'https://www.youtube.com/channel/UCxxxxxxxxxxxxxxxxxxxxxx',
    name: 'My Investment Channel',
    enabled: true,
    schedule: '0 9 * * *'  // Daily at 9 AM
  })
})
.then(res => res.json())
.then(data => console.log('Source created:', data));
```

## Finding Your Channel ID

### Standard Channel URLs

YouTube channel URLs come in different formats:

1. **Channel ID format** (recommended):
   ```
   https://www.youtube.com/channel/UCxxxxxxxxxxxxxxxxxxxxxx
   ```
   The channel ID starts with `UC` and is 24 characters long.

2. **Custom URL format** (not directly supported):
   ```
   https://www.youtube.com/c/ChannelName
   https://www.youtube.com/@ChannelName
   ```
   These require API calls to resolve to channel IDs.

### How to Find Your Channel ID

1. **From YouTube Studio**:
   - Go to https://studio.youtube.com
   - Click on "Customization" → "Basic info"
   - Your Channel ID is shown at the bottom

2. **From Channel Page**:
   - Go to your channel's "About" page
   - Scroll down to find the Channel ID

3. **From Video Page**:
   - Click on the channel name under any video
   - Look at the URL - it will contain the channel ID

## Schedule Configuration

The `schedule` field uses **cron expression** format:

| Expression | Description | Example |
|------------|-------------|---------|
| `0 9 * * *` | Daily at 9:00 AM | Every day at 9 AM |
| `0 */6 * * *` | Every 6 hours | 12 AM, 6 AM, 12 PM, 6 PM |
| `0 0 * * 1` | Weekly on Monday | Every Monday at midnight |
| `0 9,17 * * *` | Twice daily | 9 AM and 5 PM every day |
| `0 9 * * 1-5` | Weekdays only | 9 AM Monday-Friday |

### Default Schedule

If you don't specify a schedule, the system uses:
- `WORKFLOW_DEFAULT_SCHEDULE` environment variable, or
- `0 9 * * *` (daily at 9 AM) if not set

## Example: Complete Setup

### 1. Create a Channel Source

```bash
curl -X POST http://localhost:8080/api/workflow/sources \
  -H "Content-Type: application/json" \
  -d '{
    "type": "channel",
    "url": "https://www.youtube.com/channel/UCxxxxxxxxxxxxxxxxxxxxxx",
    "name": "Investment Analysis Channel",
    "enabled": true,
    "schedule": "0 9 * * *"
  }'
```

**Response:**
```json
{
  "id": "abc123-def456-ghi789",
  "type": "channel",
  "url": "https://www.youtube.com/channel/UCxxxxxxxxxxxxxxxxxxxxxx",
  "name": "Investment Analysis Channel",
  "channel_id": "UCxxxxxxxxxxxxxxxxxxxxxx",
  "enabled": true,
  "schedule": "0 9 * * *",
  "created_at": "2026-01-07T22:00:00Z"
}
```

### 2. Verify the Source

```bash
curl http://localhost:8080/api/workflow/sources
```

### 3. Check Scheduler Status

The scheduler will automatically:
- Extract the channel ID from the URL
- Set up a cron job for the specified schedule
- Poll the channel for new videos
- Process only videos published after the last processed timestamp

## How It Works

1. **Scheduler Initialization**: When the backend starts, it reads all enabled YouTube sources and sets up cron jobs
2. **Scheduled Execution**: At the scheduled time, the scheduler:
   - Fetches recent videos from the channel (last 10 videos)
   - Filters out videos that were already processed
   - Executes the workflow for each new video
   - Updates the `last_processed` timestamp
3. **Duplicate Prevention**: The system tracks processed video IDs to avoid reprocessing

## Troubleshooting

### Channel Not Being Polled

1. **Check if source is enabled**:
   ```bash
   curl http://localhost:8080/api/workflow/sources/:id
   ```
   Ensure `"enabled": true`

2. **Check scheduler is running**:
   - Look for log messages: `"Starting workflow scheduler..."`
   - Ensure `WORKFLOW_SCHEDULE_ENABLED=true`

3. **Check YouTube API Key**:
   - Verify `YOUTUBE_API_KEY` is set
   - Check logs for: `"YouTube API client initialized"`

### Channel ID Extraction Failed

If the channel ID can't be extracted from the URL:
- Use the standard channel URL format: `https://www.youtube.com/channel/UC...`
- Or manually set the `channel_id` field when creating the source (requires backend modification)

### No New Videos Found

- The system only processes videos published after the `last_processed` timestamp
- Check the `last_processed` field to see when the channel was last checked
- New videos must be published after this timestamp to be processed

## Environment Variables

Make sure these are set in your deployment:

```bash
YOUTUBE_API_KEY=your-api-key-here
WORKFLOW_SCHEDULE_ENABLED=true
WORKFLOW_DEFAULT_SCHEDULE=0 9 * * *  # Optional: default schedule
```

## Notes

- The system processes up to 10 recent videos per scheduled run
- Videos are processed sequentially to avoid overwhelming the workflow service
- Rate limiting (100ms delay) is built-in to prevent YouTube API quota issues
- The scheduler must be restarted if you add/modify sources (or implement hot-reload)

