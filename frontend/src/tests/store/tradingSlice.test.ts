import { describe, it, expect } from 'vitest';
import tradingReducer, { updateOrders, updatePositions, updateMarketData } from '@/store/trading/tradingSlice';

describe('tradingSlice', () => {
  const initialState = {
    orders: [],
    positions: [],
    marketData: {},
  };

  it('should handle initial state', () => {
    expect(tradingReducer(undefined, { type: 'unknown' })).toEqual(initialState);
  });

  it('should handle updateOrders', () => {
    const mockOrders = [{
      id: '1',
      symbol: 'BTC/USDT',
      side: 'buy',
      quantity: 1,
      price: 50000,
      status: 'open'
    }];
    const state = tradingReducer(initialState, updateOrders(mockOrders));
    expect(state.orders).toEqual(mockOrders);
  });

  it('should handle updatePositions', () => {
    const mockPositions = [{
      symbol: 'BTC/USDT',
      size: 1,
      entryPrice: 50000,
      markPrice: 51000,
      pnl: 1000
    }];
    const state = tradingReducer(initialState, updatePositions(mockPositions));
    expect(state.positions).toEqual(mockPositions);
  });

  it('should handle updateMarketData', () => {
    const mockMarketData = { 'BTC/USDT': { price: 50000 } };
    const state = tradingReducer(initialState, updateMarketData(mockMarketData));
    expect(state.marketData).toEqual(mockMarketData);
  });
});
