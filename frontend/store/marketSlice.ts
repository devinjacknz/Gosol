import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export interface MarketData {
  symbol: string;
  price: number;
  volume: number;
  change: number;
  high: number;
  low: number;
}

export interface OrderBook {
  bids: [number, number][];
  asks: [number, number][];
}

export interface Trade {
  id: string;
  symbol: string;
  side: 'buy' | 'sell';
  price: number;
  amount: number;
  timestamp: string;
}

export interface MarketState {
  marketData: Record<string, MarketData>;
  marketDataHistory: Record<string, MarketData[]>;
  orderBook: Record<string, OrderBook>;
  trades: Record<string, Trade[]>;
  selectedSymbol: string;
  errors: Record<string, string>;
}

const initialState: MarketState = {
  marketData: {},
  marketDataHistory: {},
  orderBook: {},
  trades: {},
  selectedSymbol: 'BTC/USDT',
  errors: {},
};

const marketSlice = createSlice({
  name: 'market',
  initialState,
  reducers: {
    updateMarketData(
      state,
      action: PayloadAction<{
        symbol: string;
        data?: MarketData;
        error?: string;
      }>
    ) {
      const { symbol, data, error } = action.payload;

      if (error) {
        state.errors[symbol] = error;
        return;
      }

      if (data) {
        // 保存历史数据
        if (!state.marketDataHistory[symbol]) {
          state.marketDataHistory[symbol] = [];
        }
        if (state.marketData[symbol]) {
          state.marketDataHistory[symbol].push(state.marketData[symbol]);
        }
        // 限制历史数据大小
        if (state.marketDataHistory[symbol].length > 1000) {
          state.marketDataHistory[symbol].shift();
        }

        state.marketData[symbol] = data;
        delete state.errors[symbol];
      }
    },

    updateOrderBook(
      state,
      action: PayloadAction<{
        symbol: string;
        data: OrderBook;
        isSnapshot?: boolean;
      }>
    ) {
      const { symbol, data, isSnapshot = true } = action.payload;

      if (isSnapshot) {
        state.orderBook[symbol] = data;
      } else {
        // 增量更新
        if (!state.orderBook[symbol]) {
          state.orderBook[symbol] = { bids: [], asks: [] };
        }

        const updateLevels = (
          existing: [number, number][],
          updates: [number, number][]
        ) => {
          const result = [...existing];
          for (const [price, amount] of updates) {
            const index = result.findIndex(([p]) => p === price);
            if (index >= 0) {
              if (amount === 0) {
                result.splice(index, 1);
              } else {
                result[index] = [price, amount];
              }
            } else if (amount > 0) {
              result.push([price, amount]);
            }
          }
          return result.sort((a, b) => b[0] - a[0]);
        };

        state.orderBook[symbol] = {
          bids: updateLevels(state.orderBook[symbol].bids, data.bids),
          asks: updateLevels(state.orderBook[symbol].asks, data.asks),
        };
      }
    },

    updateTrades(
      state,
      action: PayloadAction<{
        symbol: string;
        data: Trade[];
      }>
    ) {
      const { symbol, data } = action.payload;

      if (!state.trades[symbol]) {
        state.trades[symbol] = [];
      }

      state.trades[symbol] = [...data, ...state.trades[symbol]];

      // 限制交易历史大小
      if (state.trades[symbol].length > 500) {
        state.trades[symbol] = state.trades[symbol].slice(0, 500);
      }
    },

    setSelectedSymbol(state, action: PayloadAction<string>) {
      state.selectedSymbol = action.payload;
    },
  },
});

export const {
  updateMarketData,
  updateOrderBook,
  updateTrades,
  setSelectedSymbol,
} = marketSlice.actions;

export default marketSlice.reducer; 