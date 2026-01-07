import { useEffect, useState, useMemo } from 'react';
import { getWorkflowExecutions, getWorkflowExecutionDetails } from '../api';
import { WorkflowExecution, WorkflowExecutionDetails } from '../types';
import ExecutionCard from '../components/ExecutionCard';
import ExecutionDetailsModal from '../components/ExecutionDetailsModal';
import { parseDate } from '../utils/date';

type StatusFilter = 'all' | 'completed' | 'processing' | 'failed';
type DateFilter = '7' | '30' | 'all';
type SortBy = 'date-desc' | 'date-asc' | 'status';

export default function WorkflowReviewPage() {
  const [executions, setExecutions] = useState<WorkflowExecution[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedExecution, setSelectedExecution] = useState<WorkflowExecution | null>(null);
  const [executionDetails, setExecutionDetails] = useState<WorkflowExecutionDetails | null>(null);
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [dateFilter, setDateFilter] = useState<DateFilter>('all');
  const [sortBy, setSortBy] = useState<SortBy>('date-desc');

  const loadExecutions = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getWorkflowExecutions();
      setExecutions(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load workflow executions');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadExecutions();
  }, []);

  const handleExecutionClick = async (execution: WorkflowExecution) => {
    setSelectedExecution(execution);
    try {
      const details = await getWorkflowExecutionDetails(execution.id);
      setExecutionDetails(details);
    } catch (err) {
      console.error('Failed to load execution details:', err);
      setExecutionDetails(null);
    }
  };

  const handleCloseModal = () => {
    setSelectedExecution(null);
    setExecutionDetails(null);
  };

  // Filter and sort executions
  const filteredAndSortedExecutions = useMemo(() => {
    let filtered = [...executions];

    // Apply status filter
    if (statusFilter !== 'all') {
      filtered = filtered.filter((e) => e.status === statusFilter);
    }

    // Apply date filter
    if (dateFilter !== 'all') {
      const days = parseInt(dateFilter);
      const cutoffDate = new Date();
      cutoffDate.setDate(cutoffDate.getDate() - days);

      filtered = filtered.filter((e) => {
        const dateStr = e.completed_at || e.started_at || e.created_at;
        if (!dateStr) return false;
        const date = parseDate(dateStr);
        if (!date) return false;
        return date >= cutoffDate;
      });
    }

    // Apply sorting
    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'date-desc': {
          const dateA = a.completed_at || a.started_at || a.created_at || '';
          const dateB = b.completed_at || b.started_at || b.created_at || '';
          return dateB.localeCompare(dateA);
        }
        case 'date-asc': {
          const dateA = a.completed_at || a.started_at || a.created_at || '';
          const dateB = b.completed_at || b.started_at || b.created_at || '';
          return dateA.localeCompare(dateB);
        }
        case 'status':
          return a.status.localeCompare(b.status);
        default:
          return 0;
      }
    });

    return filtered;
  }, [executions, statusFilter, dateFilter, sortBy]);

  if (loading) {
    return (
      <div className="flex justify-center items-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <>
      <div className="mb-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-2xl font-bold text-gray-900">Workflow Executions</h2>
          <button
            onClick={loadExecutions}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
          >
            Refresh
          </button>
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4">
            <p className="font-medium">Error loading executions</p>
            <p className="text-sm mt-1">{error}</p>
            <button
              onClick={loadExecutions}
              className="mt-3 text-sm underline hover:no-underline"
            >
              Try again
            </button>
          </div>
        )}

        {/* Filters */}
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* Status Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Status
              </label>
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value as StatusFilter)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="all">All Statuses</option>
                <option value="completed">Completed</option>
                <option value="processing">Processing</option>
                <option value="failed">Failed</option>
              </select>
            </div>

            {/* Date Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Date Range
              </label>
              <select
                value={dateFilter}
                onChange={(e) => setDateFilter(e.target.value as DateFilter)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="all">All Time</option>
                <option value="7">Last 7 Days</option>
                <option value="30">Last 30 Days</option>
              </select>
            </div>

            {/* Sort By */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Sort By
              </label>
              <select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value as SortBy)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="date-desc">Newest First</option>
                <option value="date-asc">Oldest First</option>
                <option value="status">Status</option>
              </select>
            </div>
          </div>
        </div>

        {/* Results Count */}
        <div className="mb-4 text-sm text-gray-600">
          Showing {filteredAndSortedExecutions.length} of {executions.length} executions
        </div>

        {/* Executions Grid */}
        {filteredAndSortedExecutions.length === 0 ? (
          <div className="bg-white rounded-lg shadow p-12 text-center">
            <p className="text-gray-600">No workflow executions found matching your filters.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredAndSortedExecutions.map((execution) => (
              <ExecutionCard
                key={execution.id}
                execution={execution}
                onClick={() => handleExecutionClick(execution)}
              />
            ))}
          </div>
        )}
      </div>

      {/* Execution Details Modal */}
      <ExecutionDetailsModal
        details={executionDetails}
        isOpen={selectedExecution !== null}
        onClose={handleCloseModal}
      />
    </>
  );
}
