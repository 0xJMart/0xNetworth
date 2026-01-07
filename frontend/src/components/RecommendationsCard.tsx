import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getRecommendationsSummary, RecommendationsSummary } from '../api';

export default function RecommendationsCard() {
  const [summary, setSummary] = useState<RecommendationsSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadSummary = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getRecommendationsSummary(7);
      setSummary(data);
    } catch (err) {
      // Provide user-friendly error message
      const errorMessage = err instanceof Error ? err.message : 'Failed to load recommendations';
      if (errorMessage.includes('Failed to fetch') || errorMessage.includes('network')) {
        setError('Unable to load recommendations. Please check your connection and try again.');
      } else {
        setError('Unable to load recommendations. Please try again later.');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadSummary();
  }, []);

  const getConditionColor = (condition: string): string => {
    const lower = condition.toLowerCase();
    if (lower.includes('bullish')) return 'text-green-600 bg-green-50';
    if (lower.includes('bearish')) return 'text-red-600 bg-red-50';
    return 'text-gray-600 bg-gray-50';
  };

  const getConfidenceColor = (confidence: number): string => {
    if (confidence >= 0.7) return 'text-green-600';
    if (confidence >= 0.4) return 'text-yellow-600';
    return 'text-red-600';
  };

  const getMostCommonAction = (): string => {
    if (!summary || Object.keys(summary.action_distribution).length === 0) {
      return 'N/A';
    }
    
    let maxCount = 0;
    let mostCommon = '';
    for (const [action, count] of Object.entries(summary.action_distribution)) {
      if (count > maxCount) {
        maxCount = count;
        mostCommon = action;
      }
    }
    return mostCommon;
  };

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="animate-pulse">
          <div className="h-6 bg-gray-200 rounded w-1/3 mb-4"></div>
          <div className="h-4 bg-gray-200 rounded w-1/2"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold text-gray-900">Recommendations Summary</h2>
          <button
            onClick={loadSummary}
            disabled={loading}
            className="text-sm text-blue-600 hover:text-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? 'Loading...' : 'Retry'}
          </button>
        </div>
        <p className="text-sm text-red-600">{error}</p>
      </div>
    );
  }

  if (!summary || summary.total_count === 0) {
    return (
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold text-gray-900">Recommendations Summary</h2>
          <Link
            to="/workflows"
            className="text-sm text-blue-600 hover:text-blue-700"
          >
            View All →
          </Link>
        </div>
        <p className="text-gray-600">No recommendations in the past week.</p>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow p-6 mb-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold text-gray-900">Recommendations Summary</h2>
        <div className="flex items-center gap-3">
          <button
            onClick={loadSummary}
            className="text-sm text-gray-600 hover:text-gray-700"
            title="Refresh"
          >
            <svg
              className="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
          </button>
          <Link
            to="/workflows"
            className="text-sm text-blue-600 hover:text-blue-700 font-medium"
          >
            View All →
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
        <div className="bg-gray-50 rounded-lg p-4">
          <p className="text-sm text-gray-600 mb-1">Total Recommendations</p>
          <p className="text-2xl font-bold text-gray-900">{summary.total_count}</p>
        </div>
        
        <div className="bg-gray-50 rounded-lg p-4">
          <p className="text-sm text-gray-600 mb-1">Most Common Action</p>
          <p className="text-2xl font-bold text-gray-900 capitalize">{getMostCommonAction()}</p>
        </div>
        
        <div className="bg-gray-50 rounded-lg p-4">
          <p className="text-sm text-gray-600 mb-1">Average Confidence</p>
          <p className={`text-2xl font-bold ${getConfidenceColor(summary.average_confidence)}`}>
            {(summary.average_confidence * 100).toFixed(0)}%
          </p>
        </div>
      </div>

      {Object.keys(summary.condition_distribution).length > 0 && (
        <div className="mb-4">
          <p className="text-sm font-medium text-gray-700 mb-2">Market Conditions</p>
          <div className="flex flex-wrap gap-2">
            {Object.entries(summary.condition_distribution).map(([condition, count]) => (
              <span
                key={condition}
                className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${getConditionColor(condition)}`}
              >
                {condition} ({count})
              </span>
            ))}
          </div>
        </div>
      )}

      {summary.recent_recommendations && summary.recent_recommendations.length > 0 && (
        <div>
          <p className="text-sm font-medium text-gray-700 mb-2">Recent Recommendations</p>
          <div className="space-y-2">
            {summary.recent_recommendations.slice(0, 3).map((rec) => (
              <div
                key={rec.execution_id}
                className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
              >
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-gray-900 truncate">
                    {rec.video_title || rec.video_id}
                  </p>
                  <div className="flex items-center gap-2 mt-1">
                    <span className="text-xs text-gray-600 capitalize">{rec.action}</span>
                    <span className="text-xs text-gray-400">•</span>
                    <span className={`text-xs font-medium ${getConfidenceColor(rec.confidence)}`}>
                      {(rec.confidence * 100).toFixed(0)}% confidence
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

