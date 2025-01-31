export interface Position {
  symbol: string;
  size: number;
  entryPrice: number;
  direction: 'long' | 'short';
  leverage: number;
  liquidationPrice: number;
  marginType: 'isolated' | 'cross';
  marginRatio: number;
  unrealizedPnL: number;
}

export interface MarketData {
  [symbol: string]: {
    symbol: string;
    price: number;
    change24h: number;
    high24h: number;
    low24h: number;
    volume: number;
    fundingRate?: number;
    nextFundingTime?: string;
    volume24h?: number;
    openInterest?: number;
    orderBook?: {
      asks: [number, number][];
      bids: [number, number][];
    };
    trades?: Array<{
      id: string;
      price: number;
      amount: number;
      side: 'buy' | 'sell';
      timestamp: string;
    }>;
  };
}

export interface AccountInfo {
  positions: Position[];
  totalEquity: number;
  availableBalance: number;
  usedMargin: number;
  marginLevel: number;
  unrealizedPnL: number;
  realizedPnL: number;
  dailyPnL: number;
  balance: number;
  margin: number;
  freeMargin: number;
  marginRatio: number;
}

export interface SystemStatus {
  isConnected: boolean;
  lastUpdate: string;
  status: 'online' | 'offline' | 'maintenance';
  message?: string;
  dataDelay: number;
}

export interface Trade {
  id: string;
  symbol: string;
  direction: 'buy' | 'sell';
  size: number;
  price: number;
  timestamp: string;
  type: 'open' | 'close';
  status: 'pending' | 'filled' | 'cancelled' | 'failed';
  pnl?: number;
}

export interface RiskAlert {
  id: string;
  type: 'margin' | 'liquidation' | 'funding' | 'system';
  severity: 'low' | 'medium' | 'high' | 'critical';
  message: string;
  timestamp: string;
  symbol?: string;
}

export interface TradingStore {
  marketData: MarketData;
  setMarketData: (symbol: string, data: Partial<MarketData[string]>) => void;
  accountInfo: AccountInfo | null;
  setAccountInfo: (info: AccountInfo) => void;
  systemStatus: SystemStatus | null;
  setSystemStatus: (status: SystemStatus) => void;
  recentTrades: Trade[];
  addTrade: (trade: Trade) => void;
  riskAlerts: RiskAlert[];
  addRiskAlert: (alert: RiskAlert) => void;
  removeRiskAlert: (timestamp: string) => void;
  selectedSymbol: string;
  setSelectedSymbol: (symbol: string) => void;
}
