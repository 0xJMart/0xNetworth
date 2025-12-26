export type Platform = 'coinbase' | 'm1_finance';

export interface Account {
  id: string;
  platform: Platform;
  name: string;
  balance: number;
  currency: string;
  account_type?: string;
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
  accounts: Account[];
  investments: Investment[];
}

export interface AccountsResponse {
  accounts: Account[];
}

export interface InvestmentsResponse {
  investments: Investment[];
}

export interface PlatformAccountsResponse {
  platform: Platform;
  accounts: Account[];
}

export interface PlatformInvestmentsResponse {
  platform: Platform;
  investments: Investment[];
}

