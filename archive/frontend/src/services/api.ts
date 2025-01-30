import { MarketData, AnalysisResult, TradeSignal, TradingState } from '../types';

const API_BASE = '/api';

class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const error = await response.text().catch(() => 'Unknown error');
    throw new ApiError(response.status, error);
  }
  return response.json();
}

export async function getMarketData(tokenAddress: string): Promise<MarketData> {
  const response = await fetch(`${API_BASE}/market-data/${tokenAddress}`);
  return handleResponse<MarketData>(response);
}

export async function getAnalysis(tokenAddress: string): Promise<AnalysisResult> {
  const response = await fetch(`${API_BASE}/analysis/${tokenAddress}`);
  return handleResponse<AnalysisResult>(response);
}

export async function getTradingStats(
  tokenAddress: string,
  timeframe: '1d' | '7d' | '30d' = '7d'
): Promise<{
  daily_volume: number[];
  profit_loss: number[];
  win_rate: number;
  average_win: number;
  average_loss: number;
  largest_win: number;
  largest_loss: number;
  trade_distribution: Array<{ label: string; value: number }>;
  timestamps: string[];
}> {
  const response = await fetch(
    `${API_BASE}/trading-stats/${tokenAddress}?timeframe=${timeframe}`
  );
  return handleResponse(response);
}

export async function getTradingState(tokenAddress: string): Promise<TradingState> {
  const response = await fetch(`${API_BASE}/status?token=${tokenAddress}`);
  return handleResponse<TradingState>(response);
}

export async function executeTrade(
  tokenAddress: string,
  action: string,
  amount: number
): Promise<{ tx_hash: string }> {
  const response = await fetch(`${API_BASE}/trade`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      token_address: tokenAddress,
      action,
      amount,
    }),
  });
  return handleResponse(response);
}

export async function toggleTrading(enabled: boolean): Promise<{ status: string }> {
  const response = await fetch(`${API_BASE}/toggle-trading`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ enabled }),
  });
  return handleResponse(response);
}

export async function getWalletBalance(address: string): Promise<{ balance: number }> {
  const response = await fetch(`${API_BASE}/wallet-balance?address=${address}`);
  return handleResponse(response);
}

export async function transferProfit(
  amount: number,
  destinationAddress: string
): Promise<{ tx_hash: string }> {
  const response = await fetch(`${API_BASE}/transfer-profit`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      amount,
      destination_address: destinationAddress,
    }),
  });
  return handleResponse(response);
}

// WebSocket connection for real-time updates
export function subscribeToMarketData(
  tokenAddress: string,
  onData: (data: MarketData) => void,
  onError?: (error: Error) => void
): () => void {
  const ws = new WebSocket(
    `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${
      window.location.host
    }${API_BASE}/ws/market-data/${tokenAddress}`
  );

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      onData(data);
    } catch (err) {
      onError?.(err instanceof Error ? err : new Error('Failed to parse WebSocket data'));
    }
  };

  ws.onerror = (event) => {
    onError?.(new Error('WebSocket error'));
  };

  // Return cleanup function
  return () => {
    ws.close();
  };
}

// Helper function to format numbers
export function formatNumber(num: number, decimals: number = 2): string {
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(num);
}

// Helper function to format currency (SOL)
export function formatSOL(amount: number): string {
  return `${formatNumber(amount, 6)} SOL`;
}

// Helper function to format percentage
export function formatPercent(value: number): string {
  return `${formatNumber(value * 100, 2)}%`;
}

// Helper function to format date
export function formatDate(date: string | Date): string {
  return new Date(date).toLocaleString();
}
