import { NetWorth } from '../types';

interface NetWorthCardProps {
  networth: NetWorth;
}

export default function NetWorthCard({ networth }: NetWorthCardProps) {
  const formatCurrency = (value: number, currency: string = 'USD') => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(value);
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
      <h2 className="text-lg font-semibold text-gray-700 mb-4">Total Net Worth</h2>
      <div className="mb-4">
        <div className="text-4xl font-bold text-gray-900">
          {formatCurrency(networth.total_value, networth.currency)}
        </div>
        <div className="text-sm text-gray-500 mt-2">
          Across {networth.account_count} account{networth.account_count !== 1 ? 's' : ''}
        </div>
      </div>
      
      <div className="border-t border-gray-200 pt-4 mt-4">
        <div className="text-xs text-gray-500 mb-2">By Platform</div>
        <div className="space-y-2">
          {Object.entries(networth.by_platform).map(([platform, value]) => (
            <div key={platform} className="flex justify-between items-center">
              <span className="text-sm font-medium text-gray-700 capitalize">
                {platform.replace('_', ' ')}
              </span>
              <span className="text-sm text-gray-900 font-semibold">
                {formatCurrency(value, networth.currency)}
              </span>
            </div>
          ))}
        </div>
      </div>

      {networth.last_calculated && (
        <div className="mt-4 pt-4 border-t border-gray-200">
          <div className="text-xs text-gray-500">
            Last calculated: {formatDate(networth.last_calculated)}
          </div>
        </div>
      )}
    </div>
  );
}

