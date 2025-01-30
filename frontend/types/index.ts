// 市场数据类型
export interface MarketData {
  symbol: string;
  price: number;
  volume: number;
  high24h: number;
  low24h: number;
  change24h: number;
  timestamp: number;
}

// 持仓类型
export interface Position {
  symbol: string;
  direction: 'long' | 'short';
  size: number;
  entryPrice: number;
  leverage: number;
  marginType: 'isolated' | 'cross';
  liquidationPrice: number;
  unrealizedPnL: number;
  marginRatio: number;
  maintMargin: number;
  openTime: string;
}

// 资金费率信息
export interface FundingRate {
  symbol: string;
  currentRate: number;
  predictedRate: number;
  nextTime: string;
}

// 账户信息
export interface AccountInfo {
  totalEquity: number;
  availableBalance: number;
  usedMargin: number;
  marginLevel: number;
  unrealizedPnL: number;
  realizedPnL: number;
  dailyPnL: number;
  positions: Position[];
  fundingRates: FundingRate[];
}

// 交易历史
export interface Trade {
  id: string;
  symbol: string;
  direction: 'long' | 'short';
  type: 'open' | 'close';
  size: number;
  price: number;
  leverage: number;
  pnl: number;
  fee: number;
  timestamp: string;
}

// 风险警告
export interface RiskAlert {
  type: 'margin' | 'liquidation' | 'exposure';
  symbol: string;
  severity: 'low' | 'medium' | 'high';
  message: string;
  threshold: number;
  currentValue: number;
  timestamp: string;
}

// 系统状态
export interface SystemStatus {
  cpuUsage: number;
  memoryUsage: number;
  dbSize: number;
  dataDelay: number;
  errorCount: number;
  warningCount: number;
  lastUpdate: string;
}

// WebSocket消息类型
export type WSMessageType = 
  | { type: 'market'; data: MarketData }
  | { type: 'account'; data: AccountInfo }
  | { type: 'trade'; data: Trade }
  | { type: 'risk'; data: RiskAlert }
  | { type: 'system'; data: SystemStatus }; 