import { useEffect, useState } from 'react';
import { getYouTubeSources, createYouTubeSource, updateYouTubeSource, deleteYouTubeSource, updateSourceSchedule, triggerAllSources } from '../api';
import { YouTubeSource, CreateYouTubeSourceRequest } from '../types';
import SourceCard from '../components/SourceCard';
import SourceModal from '../components/SourceModal';

export default function YouTubeSourcesPage() {
  const [sources, setSources] = useState<YouTubeSource[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingSource, setEditingSource] = useState<YouTubeSource | null>(null);
  const [triggering, setTriggering] = useState(false);
  const [triggerSuccess, setTriggerSuccess] = useState<string | null>(null);

  const loadSources = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getYouTubeSources();
      setSources(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load YouTube sources');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadSources();
  }, []);

  const handleCreate = async (sourceData: CreateYouTubeSourceRequest) => {
    try {
      await createYouTubeSource(sourceData);
      await loadSources();
      setIsModalOpen(false);
    } catch (err) {
      throw err; // Let the modal handle the error
    }
  };

  const handleUpdate = async (sourceData: CreateYouTubeSourceRequest) => {
    if (!editingSource) return;
    try {
      await updateYouTubeSource(editingSource.id, sourceData);
      await loadSources();
      setIsModalOpen(false);
      setEditingSource(null);
    } catch (err) {
      throw err; // Let the modal handle the error
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this source? This will stop monitoring this channel.')) {
      return;
    }

    try {
      await deleteYouTubeSource(id);
      await loadSources();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete source');
    }
  };

  const handleUpdateSchedule = async (id: string, schedule: string) => {
    try {
      await updateSourceSchedule(id, schedule);
      await loadSources();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to update schedule');
    }
  };

  const handleEdit = (source: YouTubeSource) => {
    setEditingSource(source);
    setIsModalOpen(true);
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
    setEditingSource(null);
  };

  const handleTriggerAll = async () => {
    const enabledSources = sources.filter(s => s.enabled);
    if (enabledSources.length === 0) {
      alert('No enabled sources to trigger');
      return;
    }

    if (!confirm(`Trigger workflow for all ${enabledSources.length} enabled source(s)?`)) {
      return;
    }

    setTriggering(true);
    setTriggerSuccess(null);
    setError(null);

    try {
      const result = await triggerAllSources();
      setTriggerSuccess(`Successfully triggered ${result.count} source(s)`);
      // Clear success message after 5 seconds
      setTimeout(() => setTriggerSuccess(null), 5000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to trigger sources');
    } finally {
      setTriggering(false);
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">YouTube Sources</h2>
          <p className="mt-1 text-sm text-gray-600">
            Manage channels and playlists to automatically monitor for new videos
          </p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={handleTriggerAll}
            disabled={triggering || sources.filter(s => s.enabled).length === 0}
            className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {triggering ? 'Triggering...' : 'Trigger All Sources'}
          </button>
          <button
            onClick={() => setIsModalOpen(true)}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Add Source
          </button>
        </div>
      </div>

      {triggerSuccess && (
        <div className="mb-4 bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded flex items-center">
          <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
          </svg>
          {triggerSuccess}
        </div>
      )}

      {loading && <p className="text-gray-600">Loading sources...</p>}
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4">
          Error: {error}
          <button onClick={loadSources} className="ml-4 text-red-800 underline">
            Retry
          </button>
        </div>
      )}

      {!loading && !error && sources.length === 0 && (
        <div className="bg-white rounded-lg shadow p-12 text-center">
          <p className="text-gray-600 mb-4">No YouTube sources configured yet.</p>
          <button
            onClick={() => setIsModalOpen(true)}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Add Your First Source
          </button>
        </div>
      )}

      {!loading && !error && sources.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {sources.map((source) => (
            <SourceCard
              key={source.id}
              source={source}
              onEdit={handleEdit}
              onDelete={handleDelete}
              onUpdateSchedule={handleUpdateSchedule}
            />
          ))}
        </div>
      )}

      <SourceModal
        isOpen={isModalOpen}
        onClose={handleCloseModal}
        onSubmit={editingSource ? handleUpdate : handleCreate}
        editingSource={editingSource}
      />
    </div>
  );
}

