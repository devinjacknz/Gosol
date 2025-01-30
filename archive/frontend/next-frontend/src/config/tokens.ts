export interface TokenInfo {
  symbol: string;
  name: string;
  decimals: number;
  address: string;
  icon?: string;
  price?: number;
}

export interface TokenBalance {
  token: TokenInfo;
  balance: string;
  value: number;
}

export const DEFAULT_TOKENS: Record<string, TokenInfo> = {
  SUI: {
    symbol: 'SUI',
    name: 'Sui',
    decimals: 9,
    address: '0x2::sui::SUI',
    icon: '/images/tokens/sui.svg',
  },
  // Add more default tokens here
};

export interface Transaction {
  id: string;
  type: 'send' | 'receive' | 'swap';
  status: 'pending' | 'success' | 'failed';
  timestamp: number;
  from: string;
  to: string;
  amount: string;
  token: TokenInfo;
  fee?: string;
  error?: string;
}

export interface TransactionInput {
  recipient: string;
  amount: string;
  token: TokenInfo;
}

export interface SwapInput {
  fromToken: TokenInfo;
  toToken: TokenInfo;
  amount: string;
  slippage: number;
}

export const TRANSACTION_STATUS = {
  PENDING: 'pending',
  SUCCESS: 'success',
  FAILED: 'failed',
} as const;

export const DEFAULT_SLIPPAGE = 0.5; // 0.5%
export const GAS_BUFFER = 1.2; // 20% buffer for gas estimation 