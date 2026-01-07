import { useEffect, useState } from 'react';
import { Portfolio, Investment, NetWorth, Platform } from '../types';
import { fetchPortfolios, fetchInvestments, fetchNetWorth } from '../api';
import NetWorthCard from '../components/NetWorthCard';
import PortfolioList from '../components/PortfolioList';
import InvestmentChart from '../components/InvestmentChart';
import PlatformCard from '../components/PlatformCard';
import RecommendationsCard from '../components/RecommendationsCard';

export default function Dashboard() {
  const [networth, setNetworth] = useState<NetWorth | null>(null);
  const [portfolios, setPortfolios] = useState<Portfolio[]>([]);
  const [investments, setInvestments] = useState<Investment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedPlatform, setSelectedPlatform] = useState<Platform | null>(null);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);

      const [networthData, portfoliosData, investmentsData] = await Promise.all([
        fetchNetWorth(),
        fetchPortfolios(),
        fetchInvestments(),
      ]);

      setNetworth(networthData);
      setPortfolios(portfoliosData);
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

  const filteredPortfolios = selectedPlatform
    ? portfolios.filter((port) => port.platform === selectedPlatform)
    : portfolios;

  const filteredInvestments = selectedPlatform
    ? investments.filter((inv) => inv.platform === selectedPlatform)
    : investments;

  const coinbasePortfolios = portfolios.filter((port) => port.platform === 'coinbase');
  const coinbaseInvestments = investments.filter((inv) => inv.platform === 'coinbase');
  const coinbaseValue = coinbaseInvestments.reduce((sum, inv) => sum + inv.value, 0);

  const platforms: Platform[] = ['coinbase'];

  return (
    <>
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

          {/* Recommendations Summary Card */}
          <div className="mb-6">
            <RecommendationsCard />
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
              portfolios={coinbasePortfolios}
              investments={coinbaseInvestments}
              totalValue={coinbaseValue}
              currency={networth?.currency || 'USD'}
            />
          </div>

          {/* Investment Chart */}
          <div className="mb-6">
            <InvestmentChart investments={filteredInvestments} portfolios={filteredPortfolios} />
          </div>

          {/* Portfolios with Holdings */}
          <div className="mb-6">
            <PortfolioList portfolios={filteredPortfolios} investments={filteredInvestments} />
          </div>
        </>
      )}
    </>
  );
}

