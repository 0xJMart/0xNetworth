import { useState, useEffect } from 'react';
import { CreateYouTubeSourceRequest, YouTubeSource } from '../types';
import { testYouTubeSource } from '../api';

interface SourceModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (source: CreateYouTubeSourceRequest) => Promise<void>;
  editingSource?: YouTubeSource | null;
}

const SCHEDULE_PRESETS = [
  { label: 'Every 6 hours', value: '0 */6 * * *' },
  { label: 'Daily at 9 AM', value: '0 9 * * *' },
  { label: 'Daily at 12 PM', value: '0 12 * * *' },
  { label: 'Twice daily (9 AM, 5 PM)', value: '0 9,17 * * *' },
  { label: 'Weekdays only (9 AM)', value: '0 9 * * 1-5' },
  { label: 'Weekly (Monday)', value: '0 0 * * 1' },
];

export default function SourceModal({ isOpen, onClose, onSubmit, editingSource }: SourceModalProps) {
  const [formData, setFormData] = useState<CreateYouTubeSourceRequest>({
    type: 'channel',
    url: '',
    name: '',
    enabled: true,
    schedule: '0 9 * * *',
  });
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testSuccess, setTestSuccess] = useState(false);

  useEffect(() => {
    if (editingSource) {
      setFormData({
        type: editingSource.type,
        url: editingSource.url,
        name: editingSource.name,
        enabled: editingSource.enabled,
        schedule: editingSource.schedule || '0 9 * * *',
      });
    } else {
      setFormData({
        type: 'channel',
        url: '',
        name: '',
        enabled: true,
        schedule: '0 9 * * *',
      });
    }
    setError(null);
    setTestSuccess(false);
  }, [editingSource, isOpen]);

  const handleTest = async () => {
    if (!formData.url.trim()) {
      setError('Please enter a YouTube URL first');
      return;
    }

    if (!formData.url.includes('youtube.com')) {
      setError('Please enter a valid YouTube URL');
      return;
    }

    setTesting(true);
    setError(null);
    setTestSuccess(false);

    try {
      await testYouTubeSource(formData.url);
      setTestSuccess(true);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to test channel');
      setTestSuccess(false);
    } finally {
      setTesting(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    // Validate URL format
    if (!formData.url.includes('youtube.com')) {
      setError('Please enter a valid YouTube URL');
      setLoading(false);
      return;
    }

    try {
      await onSubmit(formData);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save source');
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" onClick={onClose}>
      <div
        className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-6">
          <h3 className="text-xl font-bold text-gray-900 mb-4">
            {editingSource ? 'Edit YouTube Source' : 'Add YouTube Source'}
          </h3>

          {error && (
            <div className="mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Source Type</label>
              <select
                value={formData.type}
                onChange={(e) => setFormData({ ...formData, type: e.target.value as 'channel' | 'playlist' })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                required
              >
                <option value="channel">Channel</option>
                <option value="playlist">Playlist</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="My Investment Channel"
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">YouTube URL</label>
              <div className="flex gap-2">
                <input
                  type="url"
                  value={formData.url}
                  onChange={(e) => {
                    setFormData({ ...formData, url: e.target.value });
                    setTestSuccess(false);
                  }}
                  placeholder="https://www.youtube.com/@username or https://www.youtube.com/channel/UC..."
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md"
                  required
                />
                <button
                  type="button"
                  onClick={handleTest}
                  disabled={testing || !formData.url.trim()}
                  className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed transition-colors whitespace-nowrap"
                >
                  {testing ? 'Testing...' : 'Test'}
                </button>
              </div>
              {testSuccess && (
                <div className="mt-2 flex items-center text-green-600 text-sm">
                  <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                  </svg>
                  Channel found and accessible
                </div>
              )}
              <p className="mt-1 text-xs text-gray-500">
                Supports: /channel/UC..., /@username, or /c/ChannelName formats
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Schedule (Cron Expression)</label>
              <input
                type="text"
                value={formData.schedule}
                onChange={(e) => setFormData({ ...formData, schedule: e.target.value })}
                placeholder="0 9 * * *"
                className="w-full px-3 py-2 border border-gray-300 rounded-md font-mono text-sm"
                required
              />
              <select
                onChange={(e) => {
                  if (e.target.value) {
                    setFormData({ ...formData, schedule: e.target.value });
                  }
                }}
                className="mt-2 w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                defaultValue=""
              >
                <option value="">Or select a preset...</option>
                {SCHEDULE_PRESETS.map((preset) => (
                  <option key={preset.value} value={preset.value}>
                    {preset.label}
                  </option>
                ))}
              </select>
              <p className="mt-1 text-xs text-gray-500">
                Cron format: minute hour day month weekday (e.g., "0 9 * * *" = daily at 9 AM)
              </p>
            </div>

            <div className="flex items-center">
              <input
                type="checkbox"
                id="enabled"
                checked={formData.enabled}
                onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <label htmlFor="enabled" className="ml-2 block text-sm text-gray-700">
                Enable this source
              </label>
            </div>

            <div className="flex gap-3 pt-4">
              <button
                type="submit"
                disabled={loading}
                className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
              >
                {loading ? 'Saving...' : editingSource ? 'Update' : 'Create'}
              </button>
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

