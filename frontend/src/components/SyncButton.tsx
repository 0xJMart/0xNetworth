import { useState } from 'react';
import { Platform } from '../types';
import { syncAll, syncPlatform } from '../api';

interface SyncButtonProps {
  platform?: Platform;
  onSyncComplete?: () => void;
}

export default function SyncButton({ platform, onSyncComplete }: SyncButtonProps) {
  const [isSyncing, setIsSyncing] = useState(false);
  const [lastSync, setLastSync] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleSync = async () => {
    setIsSyncing(true);
    setError(null);

    try {
      if (platform) {
        await syncPlatform(platform);
      } else {
        const result = await syncAll();
        setLastSync(result.last_sync);
      }
      onSyncComplete?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to sync');
    } finally {
      setIsSyncing(false);
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <button
        onClick={handleSync}
        disabled={isSyncing}
        className={`px-4 py-2 rounded-lg font-medium transition-colors ${
          isSyncing
            ? 'bg-gray-400 text-white cursor-not-allowed'
            : 'bg-blue-600 text-white hover:bg-blue-700'
        }`}
      >
        {isSyncing ? (
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
            Syncing...
          </span>
        ) : (
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
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
            {platform ? `Sync ${platform}` : 'Sync All'}
          </span>
        )}
      </button>
      {error && (
        <div className="text-sm text-red-600 bg-red-50 p-2 rounded">{error}</div>
      )}
      {lastSync && !error && (
        <div className="text-xs text-gray-500">
          Last synced: {new Date(lastSync).toLocaleString()}
        </div>
      )}
    </div>
  );
}

