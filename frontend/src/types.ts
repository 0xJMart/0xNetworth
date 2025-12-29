export type Platform = 'coinbase';

export interface Portfolio {
  id: string;
  platform: Platform;
  name: string;
  type?: string;
  last_synced?: string;
}

export interface Investment {
  id: string;
  account_id: string;
  platform: Platform;
  symbol: string;
  name: string;
  quantity: number;
  value: number;
  price: number;
  currency: string;
  asset_type: string;
  last_updated?: string;
}

export interface NetWorth {
  total_value: number;
  currency: string;
  by_platform: Record<Platform, number>;
  by_asset_type: Record<string, number>;
  account_count: number;
  last_calculated: string;
}

export interface NetWorthBreakdown {
  networth: NetWorth;
  portfolios: Portfolio[];
  investments: Investment[];
}

export interface InvestmentsResponse {
  investments: Investment[];
}

export interface PlatformInvestmentsResponse {
  platform: Platform;
  investments: Investment[];
}

export interface PortfoliosResponse {
  portfolios: Portfolio[];
}

export interface PlatformPortfoliosResponse {
  platform: Platform;
  portfolios: Portfolio[];
}

