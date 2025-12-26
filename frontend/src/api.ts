import {
  Account,
  AccountsResponse,
  Investment,
  InvestmentsResponse,
  NetWorth,
  NetWorthBreakdown,
  Platform,
  PlatformAccountsResponse,
  PlatformInvestmentsResponse,
} from './types';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

async function fetchAPI<T>(endpoint: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`);
  if (!response.ok) {
    throw new Error(`Failed to fetch ${endpoint}: ${response.statusText}`);
  }
  return response.json();
}

async function postAPI<T>(endpoint: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
  });
  if (!response.ok) {
    throw new Error(`Failed to post ${endpoint}: ${response.statusText}`);
  }
  return response.json();
}

// Account API
export async function fetchAccounts(): Promise<Account[]> {
  const data: AccountsResponse = await fetchAPI('/accounts');
  return data.accounts;
}

export async function fetchAccount(id: string): Promise<Account> {
  return fetchAPI(`/accounts/${id}`);
}

export async function fetchAccountsByPlatform(platform: Platform): Promise<Account[]> {
  const data: PlatformAccountsResponse = await fetchAPI(`/accounts/platform/${platform}`);
  return data.accounts;
}

// Investment API
export async function fetchInvestments(): Promise<Investment[]> {
  const data: InvestmentsResponse = await fetchAPI('/investments');
  return data.investments;
}

export async function fetchInvestmentsByAccount(accountId: string): Promise<Investment[]> {
  const data: InvestmentsResponse = await fetchAPI(`/investments/account/${accountId}`);
  return data.investments;
}

export async function fetchInvestmentsByPlatform(platform: Platform): Promise<Investment[]> {
  const data: PlatformInvestmentsResponse = await fetchAPI(`/investments/platform/${platform}`);
  return data.investments;
}

// Net Worth API
export async function fetchNetWorth(): Promise<NetWorth> {
  return fetchAPI('/networth');
}

export async function fetchNetWorthBreakdown(): Promise<NetWorthBreakdown> {
  return fetchAPI('/networth/breakdown');
}

// Sync API
export async function syncAll(): Promise<{ message: string; last_sync: string }> {
  return postAPI('/sync');
}

export async function syncPlatform(platform: Platform): Promise<{ message: string; platform: string }> {
  return postAPI(`/sync/${platform}`);
}

