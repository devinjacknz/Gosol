import { configureStore } from '@reduxjs/toolkit';
import marketReducer, {
  updateMarketData,
  updateOrderBook,
  updateTrades,
  setSelectedSymbol,
  MarketState,
} from '@/store/marketSlice';
import { mockMarketData, mockOrderBook, mockTrades } from '../utils/test-utils';

describe('Market Slice', () => {
  let store: ReturnType<typeof configureStore>;

  beforeEach(() => {
    store = configureStore({
      reducer: {
        market: marketReducer,
      },
    });
  });

  it('should handle initial state', () => {
    const state = store.getState().market;
    expect(state.marketData).toEqual({});
    expect(state.orderBook).toEqual({});
    expect(state.trades).toEqual({});
    expect(state.selectedSymbol).toBe('BTC/USDT');
  });

  it('should handle updating market data', () => {
    store.dispatch(updateMarketData({
      symbol: 'BTC/USDT',
      data: mockMarketData,
    }));

    const state = store.getState().market;
    expect(state.marketData['BTC/USDT']).toEqual(mockMarketData);
  });

  it('should handle updating order book', () => {
    store.dispatch(updateOrderBook({
      symbol: 'BTC/USDT',
      data: mockOrderBook,
    }));

    const state = store.getState().market;
    expect(state.orderBook['BTC/USDT']).toEqual(mockOrderBook);
  });

  it('should handle updating trades', () => {
    store.dispatch(updateTrades({
      symbol: 'BTC/USDT',
      data: mockTrades,
    }));

    const state = store.getState().market;
    expect(state.trades['BTC/USDT']).toEqual(mockTrades);
  });

  it('should handle setting selected symbol', () => {
    store.dispatch(setSelectedSymbol('ETH/USDT'));

    const state = store.getState().market;
    expect(state.selectedSymbol).toBe('ETH/USDT');
  });

  it('should maintain historical market data', () => {
    // 添加初始数据
    store.dispatch(updateMarketData({
      symbol: 'BTC/USDT',
      data: mockMarketData,
    }));

    // 添加新数据
    const newMarketData = {
      ...mockMarketData,
      price: 51000,
    };

    store.dispatch(updateMarketData({
      symbol: 'BTC/USDT',
      data: newMarketData,
    }));

    const state = store.getState().market;
    expect(state.marketData['BTC/USDT']).toEqual(newMarketData);
    expect(state.marketDataHistory['BTC/USDT']).toContainEqual(mockMarketData);
  });

  it('should handle multiple symbols', () => {
    // BTC数据
    store.dispatch(updateMarketData({
      symbol: 'BTC/USDT',
      data: mockMarketData,
    }));

    // ETH数据
    const ethMarketData = {
      ...mockMarketData,
      symbol: 'ETH/USDT',
      price: 3000,
    };

    store.dispatch(updateMarketData({
      symbol: 'ETH/USDT',
      data: ethMarketData,
    }));

    const state = store.getState().market;
    expect(state.marketData['BTC/USDT']).toEqual(mockMarketData);
    expect(state.marketData['ETH/USDT']).toEqual(ethMarketData);
  });

  it('should handle order book updates efficiently', () => {
    // 初始订单簿
    store.dispatch(updateOrderBook({
      symbol: 'BTC/USDT',
      data: mockOrderBook,
    }));

    // 增量更新
    const deltaOrderBook = {
      bids: [[49950, 1.0]],
      asks: [[50050, 1.0]],
    };

    store.dispatch(updateOrderBook({
      symbol: 'BTC/USDT',
      data: deltaOrderBook,
      isSnapshot: false,
    }));

    const state = store.getState().market;
    expect(state.orderBook['BTC/USDT'].bids).toContainEqual([49950, 1.0]);
    expect(state.orderBook['BTC/USDT'].asks).toContainEqual([50050, 1.0]);
  });

  it('should limit trade history size', () => {
    // 添加大量交易
    const largeTrades = Array(1000).fill(null).map((_, i) => ({
      id: i.toString(),
      price: 50000 + i,
      amount: 1.0,
      side: i % 2 === 0 ? 'buy' : 'sell',
      timestamp: new Date().toISOString(),
    }));

    store.dispatch(updateTrades({
      symbol: 'BTC/USDT',
      data: largeTrades,
    }));

    const state = store.getState().market;
    expect(state.trades['BTC/USDT'].length).toBeLessThanOrEqual(500); // 假设限制为500
  });

  it('should handle error states', () => {
    store.dispatch(updateMarketData({
      symbol: 'BTC/USDT',
      error: 'Failed to fetch market data',
    }));

    const state = store.getState().market;
    expect(state.errors['BTC/USDT']).toBe('Failed to fetch market data');
  });
}); 