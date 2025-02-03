export interface TradingState {
  orders: Order[];
  positions: Position[];
  marketData: Record<string, {
    price: number;
    volume: number;
    timestamp: number;
    high24h?: number;
    low24h?: number;
    change24h?: number;
  }>;
  loading: boolean;
  error: null | string;
  selectedSymbol: string;
}

export interface Order {
  id: string;
  type: 'market' | 'limit';
  side: 'buy' | 'sell';
  symbol: string;
  size: number;
  price?: number;
  status: 'open' | 'filled' | 'cancelled';
  timestamp: number;
}

export interface Position {
  symbol: string;
  size: number;
  entryPrice: number;
  markPrice?: number;
  pnl?: number;
  status?: 'open' | 'closed';
  lastUpdated?: number;
}
