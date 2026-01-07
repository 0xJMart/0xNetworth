import { useState } from 'react';
import { YouTubeSource } from '../api';
import { formatDate } from '../utils/date';

interface SourceCardProps {
  source: YouTubeSource;
  onEdit: (source: YouTubeSource) => void;
  onDelete: (id: string) => void;
  onUpdateSchedule: (id: string, schedule: string) => Promise<void>;
}

const SCHEDULE_PRESETS = [
  { label: 'Every 6 hours', value: '0 */6 * * *' },
  { label: 'Daily at 9 AM', value: '0 9 * * *' },
  { label: 'Daily at 12 PM', value: '0 12 * * *' },
  { label: 'Twice daily (9 AM, 5 PM)', value: '0 9,17 * * *' },
  { label: 'Weekdays only (9 AM)', value: '0 9 * * 1-5' },
  { label: 'Weekly (Monday)', value: '0 0 * * 1' },
];

export default function SourceCard({ source, onEdit, onDelete, onUpdateSchedule }: SourceCardProps) {
  const [scheduleInput, setScheduleInput] = useState(source.schedule || '');
  const [isEditingSchedule, setIsEditingSchedule] = useState(false);
  const [scheduleLoading, setScheduleLoading] = useState(false);

  const handleScheduleUpdate = async () => {
    if (!scheduleInput.trim()) {
      alert('Schedule cannot be empty');
      return;
    }

    setScheduleLoading(true);
    try {
      await onUpdateSchedule(source.id, scheduleInput);
      setIsEditingSchedule(false);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to update schedule');
    } finally {
      setScheduleLoading(false);
    }
  };

  const handlePresetSelect = (preset: string) => {
    setScheduleInput(preset);
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow">
      <div className="flex items-start justify-between mb-4">
        <div className="flex-1">
          <h3 className="text-lg font-semibold text-gray-900 mb-1">{source.name}</h3>
          <p className="text-sm text-gray-600 mb-2">{source.type === 'channel' ? 'Channel' : 'Playlist'}</p>
          <a
            href={source.url}
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm text-blue-600 hover:underline break-all"
          >
            {source.url}
          </a>
        </div>
        <div className="flex items-center gap-2">
          <span
            className={`px-2 py-1 text-xs font-medium rounded ${
              source.enabled
                ? 'bg-green-100 text-green-800'
                : 'bg-gray-100 text-gray-800'
            }`}
          >
            {source.enabled ? 'Enabled' : 'Disabled'}
          </span>
        </div>
      </div>

      {source.channel_id && (
        <div className="mb-3">
          <p className="text-xs text-gray-500">Channel ID:</p>
          <p className="text-sm font-mono text-gray-700">{source.channel_id}</p>
        </div>
      )}

      <div className="mb-4">
        <label className="block text-sm font-medium text-gray-700 mb-1">Schedule (Cron)</label>
        {isEditingSchedule ? (
          <div className="space-y-2">
            <input
              type="text"
              value={scheduleInput}
              onChange={(e) => setScheduleInput(e.target.value)}
              placeholder="0 9 * * *"
              className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
            />
            <div className="flex gap-2">
              <select
                onChange={(e) => handlePresetSelect(e.target.value)}
                className="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm"
                defaultValue=""
              >
                <option value="">Quick presets...</option>
                {SCHEDULE_PRESETS.map((preset) => (
                  <option key={preset.value} value={preset.value}>
                    {preset.label}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleScheduleUpdate}
                disabled={scheduleLoading}
                className="flex-1 px-3 py-1 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:opacity-50"
              >
                {scheduleLoading ? 'Saving...' : 'Save'}
              </button>
              <button
                onClick={() => {
                  setScheduleInput(source.schedule || '');
                  setIsEditingSchedule(false);
                }}
                className="px-3 py-1 bg-gray-200 text-gray-700 text-sm rounded hover:bg-gray-300"
              >
                Cancel
              </button>
            </div>
          </div>
        ) : (
          <div className="flex items-center justify-between">
            <p className="text-sm font-mono text-gray-700">{source.schedule || 'Not set'}</p>
            <button
              onClick={() => setIsEditingSchedule(true)}
              className="text-xs text-blue-600 hover:underline"
            >
              Edit
            </button>
          </div>
        )}
      </div>

      {source.last_processed && (
        <div className="mb-4">
          <p className="text-xs text-gray-500">Last processed:</p>
          <p className="text-sm text-gray-700">{formatDate(source.last_processed)}</p>
        </div>
      )}

      <div className="flex gap-2 pt-4 border-t border-gray-200">
        <button
          onClick={() => onEdit(source)}
          className="flex-1 px-3 py-2 bg-gray-100 text-gray-700 rounded hover:bg-gray-200 text-sm font-medium transition-colors"
        >
          Edit
        </button>
        <button
          onClick={() => onDelete(source.id)}
          className="px-3 py-2 bg-red-100 text-red-700 rounded hover:bg-red-200 text-sm font-medium transition-colors"
        >
          Delete
        </button>
      </div>
    </div>
  );
}

