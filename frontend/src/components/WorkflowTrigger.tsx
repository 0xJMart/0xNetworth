import { useState, useRef, useEffect, useCallback } from 'react';
import { executeWorkflow, getWorkflowExecution, getWorkflowExecutionDetails } from '../api';
import { WorkflowExecution, WorkflowExecutionDetails } from '../types';

interface WorkflowTriggerProps {
  onExecutionComplete?: (execution: WorkflowExecution) => void;
}

export default function WorkflowTrigger({ onExecutionComplete }: WorkflowTriggerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [input, setInput] = useState('');
  const [isExecuting, setIsExecuting] = useState(false);
  const [isPolling, setIsPolling] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [execution, setExecution] = useState<WorkflowExecution | null>(null);
  const [details, setDetails] = useState<WorkflowExecutionDetails | null>(null);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pollingIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Cleanup timeouts/intervals on unmount
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current);
      }
    };
  }, []);

  const validateInput = (value: string): string | null => {
    const trimmed = value.trim();
    if (!trimmed) {
      return 'Please enter a YouTube video ID or URL';
    }

    // Check if it's a URL - more permissive to handle query parameters
    if (trimmed.startsWith('http://') || trimmed.startsWith('https://')) {
      // Extract video ID from various YouTube URL formats, allowing query parameters
      const urlPatterns = [
        /(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([a-zA-Z0-9_-]{11})/,
        /youtube\.com\/watch\?.*v=([a-zA-Z0-9_-]{11})/,
      ];
      
      const hasValidVideoId = urlPatterns.some(pattern => pattern.test(trimmed));
      if (!hasValidVideoId) {
        return 'Invalid YouTube URL format';
      }
      return null;
    }

    // Check if it's a video ID (11 alphanumeric characters, may include - and _)
    const videoIdPattern = /^[a-zA-Z0-9_-]{11}$/;
    if (!videoIdPattern.test(trimmed)) {
      return 'Invalid video ID format (must be 11 characters)';
    }

    return null;
  };

  const pollExecutionStatus = useCallback(async (executionId: string) => {
    try {
      const currentExecution = await getWorkflowExecution(executionId);
      setExecution(currentExecution);

      if (currentExecution.status === 'completed') {
        setIsPolling(false);
        if (pollingIntervalRef.current) {
          clearInterval(pollingIntervalRef.current);
          pollingIntervalRef.current = null;
        }
        
        // Fetch full details
        const fullDetails = await getWorkflowExecutionDetails(executionId);
        setDetails(fullDetails);
        setIsExecuting(false);
      } else if (currentExecution.status === 'failed') {
        setIsPolling(false);
        if (pollingIntervalRef.current) {
          clearInterval(pollingIntervalRef.current);
          pollingIntervalRef.current = null;
        }
        setError(currentExecution.error || 'Workflow execution failed');
        setIsExecuting(false);
      }
      // If still processing, continue polling
    } catch (err) {
      setIsPolling(false);
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current);
        pollingIntervalRef.current = null;
      }
      setError(err instanceof Error ? err.message : 'Failed to check execution status');
      setIsExecuting(false);
    }
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const validationError = validateInput(input);
    if (validationError) {
      setError(validationError);
      return;
    }

    setIsExecuting(true);
    setIsPolling(false);
    setError(null);
    setExecution(null);
    setDetails(null);

    try {
      const newExecution = await executeWorkflow(input);
      setExecution(newExecution);
      onExecutionComplete?.(newExecution);
      
      // Start polling for completion
      setIsPolling(true);
      pollingIntervalRef.current = setInterval(() => {
        pollExecutionStatus(newExecution.id);
      }, 2000); // Poll every 2 seconds
      
      // Also poll immediately
      pollExecutionStatus(newExecution.id);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to execute workflow');
      setIsExecuting(false);
    }
  };

  const handleClose = useCallback(() => {
    if (!isExecuting && !isPolling) {
      // Clear timeouts/intervals
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
        timeoutRef.current = null;
      }
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current);
        pollingIntervalRef.current = null;
      }
      setIsOpen(false);
      setInput('');
      setError(null);
      setExecution(null);
      setDetails(null);
      setIsExecuting(false);
      setIsPolling(false);
    }
  }, [isExecuting, isPolling]);

  // Handle Escape key to close modal
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen && !isExecuting && !isPolling) {
        handleClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      return () => {
        document.removeEventListener('keydown', handleEscape);
      };
    }
  }, [isOpen, isExecuting, isPolling, handleClose]);

  const handleBackdropClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget && !isExecuting && !isPolling) {
      handleClose();
    }
  };

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

  return (
    <>
      <button
        onClick={() => setIsOpen(true)}
        className="px-4 py-2 rounded-lg font-medium transition-colors bg-green-600 text-white hover:bg-green-700"
        aria-label="Open workflow trigger modal to analyze YouTube video"
      >
        <span className="flex items-center gap-2">
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
              d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
            />
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          Analyze Video
        </span>
      </button>

      {isOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 overflow-y-auto"
          onClick={handleBackdropClick}
          role="dialog"
          aria-modal="true"
          aria-labelledby="workflow-modal-title"
        >
          <div className={`bg-white rounded-lg shadow-xl w-full mx-4 my-8 ${details ? 'max-w-4xl' : 'max-w-md'}`}>
            <div className="flex items-center justify-between p-6 border-b border-gray-200">
              <h2 id="workflow-modal-title" className="text-xl font-semibold text-gray-900">
                {details ? 'Workflow Results' : 'Analyze YouTube Video'}
              </h2>
              <button
                onClick={handleClose}
                disabled={isExecuting || isPolling}
                className="text-gray-400 hover:text-gray-600 disabled:opacity-50"
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

            {!details ? (
              <form onSubmit={handleSubmit} className="p-6">
                <div className="mb-4">
                  <label htmlFor="video-input" className="block text-sm font-medium text-gray-700 mb-2">
                    YouTube Video ID or URL
                  </label>
                  <input
                    id="video-input"
                    type="text"
                    value={input}
                    onChange={(e) => {
                      setInput(e.target.value);
                      setError(null);
                    }}
                    placeholder="dQw4w9WgXcQ or https://www.youtube.com/watch?v=..."
                    disabled={isExecuting || isPolling}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
                    autoFocus
                    aria-describedby="video-input-help"
                    aria-invalid={error ? 'true' : 'false'}
                    aria-required="true"
                  />
                  <p id="video-input-help" className="mt-1 text-xs text-gray-500">
                    Enter a YouTube video ID (11 characters) or full URL
                  </p>
                </div>

                {error && (
                  <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
                    <p className="text-sm text-red-700">{error}</p>
                  </div>
                )}

                {(execution || isPolling) && (
                  <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-md">
                    <p className="text-sm text-blue-700 font-medium">
                      {isPolling ? 'Processing workflow...' : 'Workflow execution started!'}
                    </p>
                    {execution && (
                      <>
                        <p className="text-xs text-blue-600 mt-1">
                          Execution ID: <span className="font-mono">{execution.id}</span>
                        </p>
                        <p className="text-xs text-blue-600">
                          Status: <span className="font-medium capitalize">{execution.status}</span>
                        </p>
                        {execution.video_title && (
                          <p className="text-xs text-blue-600 mt-1">
                            Video: {execution.video_title}
                          </p>
                        )}
                      </>
                    )}
                    {isPolling && (
                      <div className="mt-2">
                        <div className="animate-pulse flex items-center gap-2">
                          <div className="h-2 w-2 bg-blue-600 rounded-full"></div>
                          <span className="text-xs text-blue-600">Waiting for completion...</span>
                        </div>
                      </div>
                    )}
                  </div>
                )}

                <div className="flex justify-end gap-3">
                  <button
                    type="button"
                    onClick={handleClose}
                    disabled={isExecuting || isPolling}
                    className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={isExecuting || isPolling || !input.trim()}
                    className={`px-4 py-2 text-sm font-medium text-white rounded-md transition-colors ${
                      isExecuting || isPolling || !input.trim()
                        ? 'bg-gray-400 cursor-not-allowed'
                        : 'bg-green-600 hover:bg-green-700'
                    }`}
                  >
                    {isExecuting || isPolling ? (
                      <span className="flex items-center gap-2">
                        <svg
                          className="animate-spin h-4 w-4"
                          xmlns="http://www.w3.org/2000/svg"
                          fill="none"
                          viewBox="0 0 24 24"
                        >
                          <circle
                            className="opacity-25"
                            cx="12"
                            cy="12"
                            r="10"
                            stroke="currentColor"
                            strokeWidth="4"
                          ></circle>
                          <path
                            className="opacity-75"
                            fill="currentColor"
                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                          ></path>
                        </svg>
                        Processing...
                      </span>
                    ) : (
                      'Execute Workflow'
                    )}
                  </button>
                </div>
              </form>
            ) : (
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
                    onClick={handleClose}
                    className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
                  >
                    Close
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </>
  );
}
