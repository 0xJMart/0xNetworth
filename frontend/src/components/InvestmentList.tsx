import { Investment } from '../types';

interface InvestmentListProps {
  investments: Investment[];
  onInvestmentClick?: (investment: Investment) => void;
}

export default function InvestmentList({ investments, onInvestmentClick }: InvestmentListProps) {
  const formatCurrency = (value: number, currency: string = 'USD') => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(value);
  };

  const formatNumber = (value: number, decimals: number = 8) => {
    return new Intl.NumberFormat('en-US', {
      minimumFractionDigits: 0,
      maximumFractionDigits: decimals,
    }).format(value);
  };

  const getPlatformColor = (platform: string) => {
    switch (platform) {
      case 'coinbase':
        return 'bg-blue-100 text-blue-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  if (investments.length === 0) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
        <h2 className="text-lg font-semibold text-gray-700 mb-4">Holdings</h2>
        <div className="text-center py-8 text-gray-500">
          No holdings found. Sync your accounts to see investments.
        </div>
      </div>
    );
  }

  // Sort investments by value (highest first)
  const sortedInvestments = [...investments].sort((a, b) => b.value - a.value);

  return (
    <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
      <h2 className="text-lg font-semibold text-gray-700 mb-4">Holdings ({investments.length})</h2>
      <div className="space-y-3">
        {sortedInvestments.map((investment) => (
          <div
            key={investment.id}
            onClick={() => onInvestmentClick?.(investment)}
            className={`p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors ${
              onInvestmentClick ? 'cursor-pointer' : ''
            }`}
          >
            <div className="flex justify-between items-start mb-3">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <h3 className="font-semibold text-gray-900 text-lg">{investment.symbol}</h3>
                  <span className="text-sm text-gray-500">{investment.name}</span>
                  <span
                    className={`px-2 py-1 text-xs font-medium rounded ${getPlatformColor(
                      investment.platform
                    )}`}
                  >
                    {investment.platform.replace('_', ' ').toUpperCase()}
                  </span>
                </div>
                {investment.asset_type && (
                  <div className="text-xs text-gray-500 uppercase tracking-wide">
                    {investment.asset_type}
                  </div>
                )}
              </div>
              <div className="text-right">
                <div className="text-lg font-semibold text-gray-900">
                  {formatCurrency(investment.value, investment.currency)}
                </div>
                <div className="text-sm text-gray-500 mt-1">
                  {formatCurrency(investment.price, investment.currency)} per unit
                </div>
              </div>
            </div>
            
            <div className="grid grid-cols-2 gap-4 mt-3 pt-3 border-t border-gray-100">
              <div>
                <div className="text-xs text-gray-500 mb-1">Quantity</div>
                <div className="text-sm font-medium text-gray-900">
                  {formatNumber(investment.quantity)} {investment.symbol}
                </div>
              </div>
              <div>
                <div className="text-xs text-gray-500 mb-1">Price</div>
                <div className="text-sm font-medium text-gray-900">
                  {formatCurrency(investment.price, investment.currency)}
                </div>
              </div>
            </div>

            {investment.last_updated && (
              <div className="text-xs text-gray-500 mt-3 pt-3 border-t border-gray-100">
                Last updated: {new Date(investment.last_updated).toLocaleString()}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}


