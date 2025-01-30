import { create } from 'zustand'
import { AccountInfo, MarketData, SystemStatus, Trade, RiskAlert } from '../types'

interface TradingStore {
  // 市场数据
  marketData: Record<string, MarketData>;
  setMarketData: (symbol: string, data: MarketData) => void;
  
  // 账户信息
  accountInfo: AccountInfo | null;
  setAccountInfo: (info: AccountInfo) => void;
  
  // 系统状态
  systemStatus: SystemStatus | null;
  setSystemStatus: (status: SystemStatus) => void;
  
  // 最近交易
  recentTrades: Trade[];
  addTrade: (trade: Trade) => void;
  
  // 风险警告
  riskAlerts: RiskAlert[];
  addRiskAlert: (alert: RiskAlert) => void;
  removeRiskAlert: (timestamp: string) => void;
  
  // 选中的交易对
  selectedSymbol: string;
  setSelectedSymbol: (symbol: string) => void;
}

export const useTradingStore = create<TradingStore>((set) => ({
  // 市场数据
  marketData: {},
  setMarketData: (symbol, data) => 
    set((state) => ({
      marketData: { ...state.marketData, [symbol]: data }
    })),
  
  // 账户信息
  accountInfo: null,
  setAccountInfo: (info) => set({ accountInfo: info }),
  
  // 系统状态
  systemStatus: null,
  setSystemStatus: (status) => set({ systemStatus: status }),
  
  // 最近交易
  recentTrades: [],
  addTrade: (trade) => 
    set((state) => ({
      recentTrades: [trade, ...state.recentTrades].slice(0, 100)
    })),
  
  // 风险警告
  riskAlerts: [],
  addRiskAlert: (alert) =>
    set((state) => ({
      riskAlerts: [alert, ...state.riskAlerts]
    })),
  removeRiskAlert: (timestamp) =>
    set((state) => ({
      riskAlerts: state.riskAlerts.filter(
        (alert) => alert.timestamp !== timestamp
      )
    })),
  
  // 选中的交易对
  selectedSymbol: 'BTC/USDT',
  setSelectedSymbol: (symbol) => set({ selectedSymbol: symbol }),
})); 