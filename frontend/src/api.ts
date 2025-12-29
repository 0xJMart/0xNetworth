import {
  Investment,
  InvestmentsResponse,
  NetWorth,
  NetWorthBreakdown,
  Platform,
  PlatformInvestmentsResponse,
  Portfolio,
  PortfoliosResponse,
  PlatformPortfoliosResponse,
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

// Portfolio API
export async function fetchPortfolios(): Promise<Portfolio[]> {
  const data: PortfoliosResponse = await fetchAPI('/portfolios');
  return data.portfolios;
}

export async function fetchPortfolio(id: string): Promise<Portfolio> {
  return fetchAPI(`/portfolios/${id}`);
}

export async function fetchPortfoliosByPlatform(platform: Platform): Promise<Portfolio[]> {
  const data: PlatformPortfoliosResponse = await fetchAPI(`/portfolios/platform/${platform}`);
  return data.portfolios;
}

// Investment API
export async function fetchInvestments(): Promise<Investment[]> {
  const data: InvestmentsResponse = await fetchAPI('/investments');
  return data.investments;
}

export async function fetchInvestmentsByPortfolio(portfolioId: string): Promise<Investment[]> {
  const data: InvestmentsResponse = await fetchAPI(`/investments/portfolio/${portfolioId}`);
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

