import { WorkflowExecution } from '../types';

interface ExecutionCardProps {
  execution: WorkflowExecution;
  onClick: () => void;
}

export default function ExecutionCard({ execution, onClick }: ExecutionCardProps) {
  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'completed':
        return 'bg-green-100 text-green-800';
      case 'processing':
        return 'bg-blue-100 text-blue-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const formatDate = (dateString?: string): string => {
    if (!dateString) return 'N/A';
    try {
      return new Date(dateString).toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    } catch {
      return dateString;
    }
  };

  return (
    <div
      onClick={onClick}
      className="bg-white rounded-lg shadow p-4 hover:shadow-md transition-shadow cursor-pointer border border-gray-200"
    >
      <div className="flex items-start justify-between mb-2">
        <div className="flex-1 min-w-0">
          <h3 className="text-sm font-semibold text-gray-900 truncate">
            {execution.video_title || execution.video_id || 'Untitled Video'}
          </h3>
          {execution.video_id && (
            <p className="text-xs text-gray-500 font-mono mt-1">{execution.video_id}</p>
          )}
        </div>
        <span
          className={`ml-2 px-2 py-1 rounded-full text-xs font-medium capitalize ${getStatusColor(execution.status)}`}
        >
          {execution.status}
        </span>
      </div>

      <div className="mt-3 flex items-center justify-between text-xs text-gray-600">
        <span>{formatDate(execution.completed_at || execution.started_at || execution.created_at)}</span>
        {execution.source_id && (
          <span className="text-gray-400">From source</span>
        )}
      </div>

      {execution.error && (
        <div className="mt-2 p-2 bg-red-50 border border-red-200 rounded text-xs text-red-700">
          Error: {execution.error}
        </div>
      )}
    </div>
  );
}

