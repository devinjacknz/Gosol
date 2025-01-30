export interface MarketData {
  token_address: string;
  price: number;
  volume_24h: number;
  market_cap: number;
  liquidity: number;
  price_impact: number;
  timestamp: string;
}

export interface TechnicalIndicators {
  rsi: number;
  macd: number;
  signal: number;
  bb_upper: number;
  bb_lower: number;
  ema20: number;
  volume: number;
  volatility: number;
}

export interface RiskAnalysis {
  manipulation_risk: string;
  liquidity_risk: string;
  volatility_risk: string;
}

export interface TradeRecommendation {
  action: string;
  entry_points: number[];
  exit_points: number[];
  stop_loss: number;
}

export interface DeepseekAnalysis {
  sentiment: string;
  confidence: number;
  key_factors: string[];
  risk_analysis: RiskAnalysis;
  recommendation: TradeRecommendation;
}

export interface AnalysisResult {
  token_address: string;
  prediction: number;
  confidence: number;
  sentiment: string;
  risk_level: number;
  technical_indicators: string; // JSON string
  deepseek_analysis: string; // JSON string
  timestamp: string;
}

export interface TradeSignal {
  token_address: string;
  action: string;
  target_price: number;
  amount: number;
  timestamp: string;
}

export interface TradingState {
  is_trading: boolean;
  last_trade: string;
  current_profit: number;
  total_trades: number;
  successful_trades: number;
  market_data: MarketData;
}
