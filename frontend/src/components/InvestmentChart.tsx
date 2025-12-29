import { Investment, Portfolio } from '../types';
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Legend,
  Tooltip,
} from 'recharts';

interface InvestmentChartProps {
  investments: Investment[];
  portfolios: Portfolio[];
}

const COLORS = [
  '#3b82f6', // blue
  '#10b981', // green
  '#f59e0b', // amber
  '#ef4444', // red
  '#8b5cf6', // purple
  '#ec4899', // pink
  '#06b6d4', // cyan
  '#84cc16', // lime
];

export default function InvestmentChart({ investments, portfolios }: InvestmentChartProps) {
  if (investments.length === 0) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
        <h2 className="text-lg font-semibold text-gray-700 mb-4">Investment Distribution</h2>
        <div className="text-center py-8 text-gray-500">
          No investments to display. Sync your accounts to see investments.
        </div>
      </div>
    );
  }

  // Create a mapping from portfolio ID to portfolio name
  const portfolioMap = new Map<string, string>();
  portfolios.forEach((portfolio) => {
    portfolioMap.set(portfolio.id, portfolio.name);
  });

  // Group investments by portfolio name
  const portfolioData = investments.reduce((acc, inv) => {
    const portfolioName = portfolioMap.get(inv.account_id) || `Unknown Portfolio (${inv.account_id})`;
    const existing = acc.find((item) => item.name === portfolioName);
    if (existing) {
      existing.value += inv.value;
    } else {
      acc.push({ name: portfolioName, value: inv.value });
    }
    return acc;
  }, [] as { name: string; value: number }[]);

  // Format for display
  const chartData = portfolioData.map((item) => ({
    name: item.name,
    value: item.value,
  }));

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  };

  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-white p-3 border border-gray-200 rounded-lg shadow-lg">
          <p className="font-medium text-gray-900">{payload[0].name}</p>
          <p className="text-sm text-gray-600">{formatCurrency(payload[0].value)}</p>
        </div>
      );
    }
    return null;
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
      <h2 className="text-lg font-semibold text-gray-700 mb-4">Investment Distribution</h2>
      <div className="h-80">
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={chartData}
              cx="50%"
              cy="50%"
              labelLine={false}
              label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
              outerRadius={100}
              fill="#8884d8"
              dataKey="value"
            >
              {chartData.map((_, index) => (
                <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
              ))}
            </Pie>
            <Tooltip content={<CustomTooltip />} />
            <Legend />
          </PieChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

