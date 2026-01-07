import { useState } from 'react';
import { executeWorkflow, getWorkflowExecution } from '../api';
import { WorkflowExecution } from '../types';

interface WorkflowTriggerProps {
  onExecutionComplete?: (execution: WorkflowExecution) => void;
}

export default function WorkflowTrigger({ onExecutionComplete }: WorkflowTriggerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [input, setInput] = useState('');
  const [isExecuting, setIsExecuting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<WorkflowExecution | null>(null);

  const validateInput = (value: string): string | null => {
    const trimmed = value.trim();
    if (!trimmed) {
      return 'Please enter a YouTube video ID or URL';
    }

    // Check if it's a URL
    if (trimmed.startsWith('http://') || trimmed.startsWith('https://')) {
      const urlPattern = /^(https?:\/\/)?(www\.)?(youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([a-zA-Z0-9_-]{11})/;
      if (!urlPattern.test(trimmed)) {
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const validationError = validateInput(input);
    if (validationError) {
      setError(validationError);
      return;
    }

    setIsExecuting(true);
    setError(null);
    setSuccess(null);

    try {
      const execution = await executeWorkflow(input);
      setSuccess(execution);
      onExecutionComplete?.(execution);
      
      // Auto-close after 3 seconds on success
      setTimeout(() => {
        setIsOpen(false);
        setInput('');
        setSuccess(null);
      }, 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to execute workflow');
    } finally {
      setIsExecuting(false);
    }
  };

  const handleClose = () => {
    if (!isExecuting) {
      setIsOpen(false);
      setInput('');
      setError(null);
      setSuccess(null);
    }
  };

  return (
    <>
      <button
        onClick={() => setIsOpen(true)}
        className="px-4 py-2 rounded-lg font-medium transition-colors bg-green-600 text-white hover:bg-green-700"
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
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
            <div className="flex items-center justify-between p-6 border-b border-gray-200">
              <h2 className="text-xl font-semibold text-gray-900">Analyze YouTube Video</h2>
              <button
                onClick={handleClose}
                disabled={isExecuting}
                className="text-gray-400 hover:text-gray-600 disabled:opacity-50"
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
                  disabled={isExecuting}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
                  autoFocus
                />
                <p className="mt-1 text-xs text-gray-500">
                  Enter a YouTube video ID (11 characters) or full URL
                </p>
              </div>

              {error && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
                  <p className="text-sm text-red-700">{error}</p>
                </div>
              )}

              {success && (
                <div className="mb-4 p-3 bg-green-50 border border-green-200 rounded-md">
                  <p className="text-sm text-green-700 font-medium">
                    Workflow execution started successfully!
                  </p>
                  <p className="text-xs text-green-600 mt-1">
                    Execution ID: {success.id}
                  </p>
                  <p className="text-xs text-green-600">
                    Status: {success.status}
                  </p>
                </div>
              )}

              <div className="flex justify-end gap-3">
                <button
                  type="button"
                  onClick={handleClose}
                  disabled={isExecuting}
                  className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isExecuting || !input.trim()}
                  className={`px-4 py-2 text-sm font-medium text-white rounded-md transition-colors ${
                    isExecuting || !input.trim()
                      ? 'bg-gray-400 cursor-not-allowed'
                      : 'bg-green-600 hover:bg-green-700'
                  }`}
                >
                  {isExecuting ? (
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
          </div>
        </div>
      )}
    </>
  );
}

