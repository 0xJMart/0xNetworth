import { WorkflowExecutionDetails } from '../types';

interface ExecutionDetailsModalProps {
  details: WorkflowExecutionDetails | null;
  isOpen: boolean;
  onClose: () => void;
}

export default function ExecutionDetailsModal({ details, isOpen, onClose }: ExecutionDetailsModalProps) {
  if (!isOpen || !details) return null;

  const formatDuration = (seconds?: number): string => {
    if (!seconds) return 'Unknown';
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const getConditionColor = (conditions: string): string => {
    const lower = conditions.toLowerCase();
    if (lower.includes('bullish')) return 'text-green-600 bg-green-50';
    if (lower.includes('bearish')) return 'text-red-600 bg-red-50';
    return 'text-gray-600 bg-gray-50';
  };

  const getConfidenceColor = (confidence: number): string => {
    if (confidence >= 0.7) return 'text-green-600';
    if (confidence >= 0.4) return 'text-yellow-600';
    return 'text-red-600';
  };

  const handleBackdropClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 overflow-y-auto"
      onClick={handleBackdropClick}
      role="dialog"
      aria-modal="true"
      aria-labelledby="execution-modal-title"
    >
      <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full mx-4 my-8">
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <h2 id="execution-modal-title" className="text-xl font-semibold text-gray-900">
            Execution Details
          </h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
            aria-label="Close modal"
          >
            <svg
              className="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        <div className="p-6 space-y-6 max-h-[calc(100vh-200px)] overflow-y-auto">
          {/* Video Info */}
          {details.transcript && (
            <div className="border-b border-gray-200 pb-4">
              <h3 className="text-lg font-semibold text-gray-900 mb-2">
                {details.transcript.video_title}
              </h3>
              <div className="flex flex-wrap gap-4 text-sm text-gray-600">
                <span>Video ID: <span className="font-mono">{details.transcript.video_id}</span></span>
                {details.transcript.duration && (
                  <span>Duration: {formatDuration(details.transcript.duration)}</span>
                )}
              </div>
              {details.transcript.video_url && (
                <a
                  href={details.transcript.video_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-blue-600 hover:underline mt-2 inline-block"
                >
                  Watch on YouTube →
                </a>
              )}
            </div>
          )}

          {/* Market Analysis */}
          {details.market_analysis && (
            <div className="border-b border-gray-200 pb-4">
              <h3 className="text-lg font-semibold text-gray-900 mb-3">Market Analysis</h3>
              
              <div className="mb-3">
                <span className={`inline-block px-3 py-1 rounded-full text-sm font-medium ${getConditionColor(details.market_analysis.conditions)}`}>
                  {details.market_analysis.conditions}
                </span>
              </div>

              {details.market_analysis.summary && (
                <p className="text-sm text-gray-700 mb-3">{details.market_analysis.summary}</p>
              )}

              {details.market_analysis.trends && details.market_analysis.trends.length > 0 && (
                <div className="mb-3">
                  <h4 className="text-sm font-medium text-gray-900 mb-2">Trends:</h4>
                  <ul className="list-disc list-inside text-sm text-gray-700 space-y-1">
                    {details.market_analysis.trends.map((trend, idx) => (
                      <li key={idx}>{trend}</li>
                    ))}
                  </ul>
                </div>
              )}

              {details.market_analysis.risk_factors && details.market_analysis.risk_factors.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium text-gray-900 mb-2">Risk Factors:</h4>
                  <ul className="list-disc list-inside text-sm text-gray-700 space-y-1">
                    {details.market_analysis.risk_factors.map((risk, idx) => (
                      <li key={idx}>{risk}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}

          {/* Recommendations */}
          {details.recommendation && (
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-3">Recommendations</h3>
              
              <div className="mb-3">
                <div className="flex items-center gap-3 mb-2">
                  <span className="text-sm font-medium text-gray-900">
                    Action: <span className="capitalize">{details.recommendation.action}</span>
                  </span>
                  <span className={`text-sm font-medium ${getConfidenceColor(details.recommendation.confidence)}`}>
                    Confidence: {(details.recommendation.confidence * 100).toFixed(0)}%
                  </span>
                </div>
                
                {details.recommendation.summary && (
                  <p className="text-sm text-gray-700 mb-3">{details.recommendation.summary}</p>
                )}
              </div>

              {details.recommendation.suggested_actions && details.recommendation.suggested_actions.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium text-gray-900 mb-2">Suggested Actions:</h4>
                  <div className="space-y-2">
                    {details.recommendation.suggested_actions.map((action, idx) => (
                      <div key={idx} className="p-3 bg-gray-50 rounded-md border border-gray-200">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="text-sm font-medium text-gray-900 capitalize">{action.type}</span>
                          {action.symbol && (
                            <span className="text-sm font-mono text-gray-600">({action.symbol})</span>
                          )}
                        </div>
                        {action.rationale && (
                          <p className="text-xs text-gray-600">{action.rationale}</p>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Transcript Preview */}
          {details.transcript && (
            <div className="border-t border-gray-200 pt-4">
              <h3 className="text-lg font-semibold text-gray-900 mb-3">Transcript Preview</h3>
              <div className="p-3 bg-gray-50 rounded-md border border-gray-200 max-h-48 overflow-y-auto">
                <p className="text-sm text-gray-700 whitespace-pre-wrap">
                  {details.transcript.text.length > 500
                    ? `${details.transcript.text.substring(0, 500)}...`
                    : details.transcript.text}
                </p>
                {details.transcript.text.length > 500 && (
                  <p className="text-xs text-gray-500 mt-2">
                    ({details.transcript.text.length} characters total)
                  </p>
                )}
              </div>
            </div>
          )}

          <div className="flex justify-end pt-4 border-t border-gray-200">
            <button
              onClick={onClose}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

