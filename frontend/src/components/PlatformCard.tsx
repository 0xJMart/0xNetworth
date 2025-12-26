import { Account, Investment, Platform } from '../types';

interface PlatformCardProps {
  platform: Platform;
  accounts: Account[];
  investments: Investment[];
  totalValue: number;
  currency: string;
}

export default function PlatformCard({
  platform,
  accounts,
  investments,
  totalValue,
  currency,
}: PlatformCardProps) {
  const formatCurrency = (value: number, currency: string = 'USD') => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(value);
  };

  const getPlatformName = (platform: Platform) => {
    switch (platform) {
      case 'coinbase':
        return 'Coinbase';
      case 'm1_finance':
        return 'M1 Finance';
      default:
        return platform;
    }
  };

  const getPlatformColor = (platform: Platform) => {
    switch (platform) {
      case 'coinbase':
        return 'border-blue-500 bg-blue-50';
      case 'm1_finance':
        return 'border-green-500 bg-green-50';
      default:
        return 'border-gray-500 bg-gray-50';
    }
  };

  const accountBalance = accounts.reduce((sum, acc) => sum + acc.balance, 0);
  const investmentValue = investments.reduce((sum, inv) => sum + inv.value, 0);

  return (
    <div className={`bg-white rounded-lg shadow-md p-6 border-2 ${getPlatformColor(platform)}`}>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-bold text-gray-900">{getPlatformName(platform)}</h2>
        <div className="text-2xl font-bold text-gray-900">{formatCurrency(totalValue, currency)}</div>
      </div>

      <div className="space-y-4">
        <div>
          <div className="text-sm text-gray-600 mb-2">Summary</div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <div className="text-xs text-gray-500">Accounts</div>
              <div className="text-lg font-semibold text-gray-900">{accounts.length}</div>
            </div>
            <div>
              <div className="text-xs text-gray-500">Holdings</div>
              <div className="text-lg font-semibold text-gray-900">{investments.length}</div>
            </div>
          </div>
        </div>

        <div className="border-t border-gray-200 pt-4">
          <div className="text-sm text-gray-600 mb-2">Breakdown</div>
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-gray-700">Account Balances</span>
              <span className="font-medium text-gray-900">{formatCurrency(accountBalance, currency)}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-700">Investment Value</span>
              <span className="font-medium text-gray-900">{formatCurrency(investmentValue, currency)}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

