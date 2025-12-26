import { Account } from '../types';

interface AccountListProps {
  accounts: Account[];
  onAccountClick?: (account: Account) => void;
}

export default function AccountList({ accounts, onAccountClick }: AccountListProps) {
  const formatCurrency = (value: number, currency: string = 'USD') => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(value);
  };

  const getPlatformColor = (platform: string) => {
    switch (platform) {
      case 'coinbase':
        return 'bg-blue-100 text-blue-800';
      case 'm1_finance':
        return 'bg-green-100 text-green-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  if (accounts.length === 0) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
        <h2 className="text-lg font-semibold text-gray-700 mb-4">Accounts</h2>
        <div className="text-center py-8 text-gray-500">
          No accounts found. Sync your accounts to get started.
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
      <h2 className="text-lg font-semibold text-gray-700 mb-4">Accounts ({accounts.length})</h2>
      <div className="space-y-3">
        {accounts.map((account) => (
          <div
            key={account.id}
            onClick={() => onAccountClick?.(account)}
            className={`p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors ${
              onAccountClick ? 'cursor-pointer' : ''
            }`}
          >
            <div className="flex justify-between items-start mb-2">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <h3 className="font-medium text-gray-900">{account.name}</h3>
                  <span
                    className={`px-2 py-1 text-xs font-medium rounded ${getPlatformColor(
                      account.platform
                    )}`}
                  >
                    {account.platform.replace('_', ' ').toUpperCase()}
                  </span>
                </div>
                {account.account_type && (
                  <div className="text-sm text-gray-500">{account.account_type}</div>
                )}
              </div>
              <div className="text-right">
                <div className="text-lg font-semibold text-gray-900">
                  {formatCurrency(account.balance, account.currency)}
                </div>
              </div>
            </div>
            {account.last_synced && (
              <div className="text-xs text-gray-500 mt-2">
                Last synced: {new Date(account.last_synced).toLocaleString()}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

