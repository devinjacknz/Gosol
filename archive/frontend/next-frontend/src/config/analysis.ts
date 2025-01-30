export interface Indicator {
  id: string;
  name: string;
  type: string;
  params: Record<string, any>;
  color: string;
  visible: boolean;
}

export interface IndicatorType {
  MA: 'ma';      // Moving Average
  EMA: 'ema';    // Exponential Moving Average
  MACD: 'macd';  // Moving Average Convergence Divergence
  RSI: 'rsi';    // Relative Strength Index
  BB: 'bb';      // Bollinger Bands
  STOCH: 'stoch';// Stochastic Oscillator
  ATR: 'atr';    // Average True Range
  OBV: 'obv';    // On Balance Volume
  VWAP: 'vwap';  // Volume Weighted Average Price
  PIVOT: 'pivot';// Pivot Points
}

export interface BacktestConfig {
  id: string;
  name: string;
  startTime: number;
  endTime: number;
  initialCapital: string;
  leverage: number;
  pairs: string[];
  strategy: {
    id: string;
    params: Record<string, any>;
  };
  riskManagement: {
    stopLoss: number;
    takeProfit: number;
    trailingStop: boolean;
    maxDrawdown: number;
    positionSize: number;
  };
}

export interface BacktestResult {
  id: string;
  configId: string;
  startTime: number;
  endTime: number;
  trades: BacktestTrade[];
  metrics: {
    totalPnl: string;
    winRate: number;
    sharpeRatio: number;
    maxDrawdown: string;
    averageTrade: string;
    profitFactor: number;
    recoveryFactor: number;
    expectancy: number;
    trades: number;
    winningTrades: number;
    losingTrades: number;
  };
  equity: {
    time: number;
    value: string;
  }[];
  drawdown: {
    time: number;
    value: number;
  }[];
}

export interface BacktestTrade {
  id: string;
  pair: string;
  side: 'buy' | 'sell';
  entryTime: number;
  entryPrice: string;
  exitTime: number;
  exitPrice: string;
  amount: string;
  pnl: string;
  pnlPercent: number;
  fee: string;
}

export interface AnalysisTimeframe {
  value: string;
  label: string;
  seconds: number;
}

export const INDICATOR_TYPES: IndicatorType = {
  MA: 'ma',
  EMA: 'ema',
  MACD: 'macd',
  RSI: 'rsi',
  BB: 'bb',
  STOCH: 'stoch',
  ATR: 'atr',
  OBV: 'obv',
  VWAP: 'vwap',
  PIVOT: 'pivot',
};

export const ANALYSIS_TIMEFRAMES: AnalysisTimeframe[] = [
  { value: '1m', label: '1 Minute', seconds: 60 },
  { value: '5m', label: '5 Minutes', seconds: 300 },
  { value: '15m', label: '15 Minutes', seconds: 900 },
  { value: '30m', label: '30 Minutes', seconds: 1800 },
  { value: '1h', label: '1 Hour', seconds: 3600 },
  { value: '4h', label: '4 Hours', seconds: 14400 },
  { value: '1d', label: '1 Day', seconds: 86400 },
  { value: '1w', label: '1 Week', seconds: 604800 },
];

export const DEFAULT_INDICATORS: Indicator[] = [
  {
    id: 'ma_20',
    name: 'MA 20',
    type: INDICATOR_TYPES.MA,
    params: { period: 20 },
    color: '#2196f3',
    visible: true,
  },
  {
    id: 'ma_50',
    name: 'MA 50',
    type: INDICATOR_TYPES.MA,
    params: { period: 50 },
    color: '#ff9800',
    visible: true,
  },
  {
    id: 'bb',
    name: 'Bollinger Bands',
    type: INDICATOR_TYPES.BB,
    params: { period: 20, stdDev: 2 },
    color: '#4caf50',
    visible: false,
  },
  {
    id: 'rsi',
    name: 'RSI',
    type: INDICATOR_TYPES.RSI,
    params: { period: 14 },
    color: '#f44336',
    visible: false,
  },
]; 