import { Portfolio, Investment } from '../types';

interface PortfolioListProps {
  portfolios: Portfolio[];
  investments: Investment[];
  onPortfolioClick?: (portfolio: Portfolio) => void;
}

export default function PortfolioList({ portfolios, investments, onPortfolioClick }: PortfolioListProps) {
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

  // Group investments by portfolio (account_id is the portfolio ID)
  const getPortfolioInvestments = (portfolioId: string): Investment[] => {
    return investments.filter(inv => inv.account_id === portfolioId);
  };

  // Calculate total value for a portfolio
  const getPortfolioValue = (portfolioId: string): number => {
    return getPortfolioInvestments(portfolioId).reduce((sum, inv) => sum + inv.value, 0);
  };

  if (portfolios.length === 0) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
        <h2 className="text-lg font-semibold text-gray-700 mb-4">Portfolios</h2>
        <div className="text-center py-8 text-gray-500">
          No portfolios found. Sync your accounts to get started.
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
      <h2 className="text-lg font-semibold text-gray-700 mb-4">Portfolios ({portfolios.length})</h2>
      <div className="space-y-6">
        {portfolios.map((portfolio) => {
          const portfolioInvestments = getPortfolioInvestments(portfolio.id);
          const portfolioValue = getPortfolioValue(portfolio.id);

          return (
            <div
              key={portfolio.id}
              className="border border-gray-200 rounded-lg p-4 hover:bg-gray-50 transition-colors"
            >
              {/* Portfolio Header */}
              <div
                onClick={() => onPortfolioClick?.(portfolio)}
                className={`cursor-pointer ${onPortfolioClick ? '' : 'cursor-default'}`}
              >
                <div className="flex justify-between items-start mb-3">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="font-semibold text-gray-900 text-lg">{portfolio.name}</h3>
                      <span
                        className={`px-2 py-1 text-xs font-medium rounded ${getPlatformColor(
                          portfolio.platform
                        )}`}
                      >
                        {portfolio.platform.replace('_', ' ').toUpperCase()}
                      </span>
                    </div>
                    {portfolio.type && (
                      <div className="text-xs text-gray-500 uppercase tracking-wide">{portfolio.type}</div>
                    )}
                  </div>
                  <div className="text-right">
                    <div className="text-lg font-semibold text-gray-900">
                      {formatCurrency(portfolioValue)}
                    </div>
                    <div className="text-sm text-gray-500 mt-1">
                      {portfolioInvestments.length} holding{portfolioInvestments.length !== 1 ? 's' : ''}
                    </div>
                  </div>
                </div>
              </div>

              {/* Portfolio Holdings */}
              {portfolioInvestments.length > 0 ? (
                <div className="mt-4 pt-4 border-t border-gray-100">
                  <div className="text-xs font-semibold text-gray-600 uppercase tracking-wide mb-2">
                    Holdings
                  </div>
                  <div className="space-y-2">
                    {portfolioInvestments
                      .sort((a, b) => b.value - a.value)
                      .map((investment) => (
                        <div
                          key={investment.id}
                          className="flex justify-between items-center p-2 bg-gray-50 rounded"
                        >
                          <div className="flex-1">
                            <div className="flex items-center gap-2">
                              <span className="font-medium text-gray-900">{investment.symbol}</span>
                              <span className="text-xs text-gray-500">
                                {formatNumber(investment.quantity)} {investment.symbol}
                              </span>
                            </div>
                            <div className="text-xs text-gray-500 mt-0.5">
                              {formatCurrency(investment.price)} per unit
                            </div>
                          </div>
                          <div className="text-right">
                            <div className="font-semibold text-gray-900">
                              {formatCurrency(investment.value, investment.currency)}
                            </div>
                          </div>
                        </div>
                      ))}
                  </div>
                </div>
              ) : (
                <div className="mt-4 pt-4 border-t border-gray-100 text-center text-sm text-gray-500">
                  No holdings in this portfolio
                </div>
              )}

              {portfolio.last_synced && (
                <div className="text-xs text-gray-500 mt-4 pt-4 border-t border-gray-100">
                  Last synced: {new Date(portfolio.last_synced).toLocaleString()}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}


