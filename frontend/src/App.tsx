import { useEffect, useState } from 'react';
import {
  Account,
  Investment,
  NetWorth,
  Platform,
  fetchAccounts,
  fetchInvestments,
  fetchNetWorth,
  fetchNetWorthBreakdown,
} from './api';
import NetWorthCard from './components/NetWorthCard';
import AccountList from './components/AccountList';
import InvestmentChart from './components/InvestmentChart';
import PlatformCard from './components/PlatformCard';
import SyncButton from './components/SyncButton';

function App() {
  const [networth, setNetworth] = useState<NetWorth | null>(null);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [investments, setInvestments] = useState<Investment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedPlatform, setSelectedPlatform] = useState<Platform | null>(null);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);

      const [networthData, accountsData, investmentsData] = await Promise.all([
        fetchNetWorth(),
        fetchAccounts(),
        fetchInvestments(),
      ]);

      setNetworth(networthData);
      setAccounts(accountsData);
      setInvestments(investmentsData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleSyncComplete = () => {
    loadData();
  };

  const filteredAccounts = selectedPlatform
    ? accounts.filter((acc) => acc.platform === selectedPlatform)
    : accounts;

  const filteredInvestments = selectedPlatform
    ? investments.filter((inv) => inv.platform === selectedPlatform)
    : investments;

  const coinbaseAccounts = accounts.filter((acc) => acc.platform === 'coinbase');
  const coinbaseInvestments = investments.filter((inv) => inv.platform === 'coinbase');
  const coinbaseValue =
    coinbaseAccounts.reduce((sum, acc) => sum + acc.balance, 0) +
    coinbaseInvestments.reduce((sum, inv) => sum + inv.value, 0);

  const platforms: Platform[] = ['coinbase'];

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">0xNetworth</h1>
              <p className="mt-2 text-gray-600">Investment Tracking Dashboard</p>
            </div>
            <div className="flex items-center gap-4">
              <SyncButton onSyncComplete={handleSyncComplete} />
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {loading && (
          <div className="flex justify-center items-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
          </div>
        )}

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-6">
            <p className="font-medium">Error loading data</p>
            <p className="text-sm mt-1">{error}</p>
            <button
              onClick={loadData}
              className="mt-3 text-sm underline hover:no-underline"
            >
              Try again
            </button>
          </div>
        )}

        {!loading && !error && (
          <>
            {/* Platform Filter */}
            <div className="mb-6">
              <div className="flex flex-wrap gap-2">
                <button
                  onClick={() => setSelectedPlatform(null)}
                  className={`px-4 py-2 rounded-full text-sm font-medium transition-colors ${
                    selectedPlatform === null
                      ? 'bg-blue-600 text-white'
                      : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                  }`}
                >
                  All Platforms
                </button>
                {platforms.map((platform) => (
                  <button
                    key={platform}
                    onClick={() => setSelectedPlatform(platform)}
                    className={`px-4 py-2 rounded-full text-sm font-medium transition-colors ${
                      selectedPlatform === platform
                        ? 'bg-blue-600 text-white'
                        : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                    }`}
                  >
                    Coinbase
                  </button>
                ))}
              </div>
            </div>

            {/* Net Worth Card */}
            {networth && (
              <div className="mb-6">
                <NetWorthCard networth={networth} />
              </div>
            )}

            {/* Platform Card */}
            <div className="mb-6">
              <PlatformCard
                platform="coinbase"
                accounts={coinbaseAccounts}
                investments={coinbaseInvestments}
                totalValue={coinbaseValue}
                currency={networth?.currency || 'USD'}
              />
            </div>

            {/* Main Content Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Accounts */}
              <div>
                <AccountList accounts={filteredAccounts} />
              </div>

              {/* Investment Chart */}
              <div>
                <InvestmentChart investments={filteredInvestments} />
              </div>
            </div>
          </>
        )}
      </main>
    </div>
  );
}

export default App;

