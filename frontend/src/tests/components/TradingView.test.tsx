import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import TradingView from '@/pages/TradingView';
import tradingReducer from '@/store/trading/tradingSlice';
import monitoringReducer from '@/store/monitoring/monitoringSlice';
import React from 'react';
import '@testing-library/jest-dom';

describe('TradingView', () => {
  const mockStore = configureStore({
    reducer: {
      trading: tradingReducer,
      monitoring: monitoringReducer,
    },
  });

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders trading components', () => {
    const mockState = {
      trading: {
        orders: [{ id: '1', type: 'market', size: '1.0', price: '50000' }],
        positions: [{ symbol: 'BTC-USD', size: '1.0', entryPrice: '50000' }],
        marketData: {},
        selectedSymbol: 'BTC-USD',
        loading: false
      },
      monitoring: {
        metrics: {
          systemStatus: 'healthy',
          cpuUsage: 25,
          memoryUsage: 50
        }
      }
    };

    const storeWithData = configureStore({
      reducer: {
        trading: tradingReducer,
        monitoring: monitoringReducer
      },
      preloadedState: mockState
    });

    render(
      <Provider store={storeWithData}>
        <TradingView />
      </Provider>
    );

    expect(screen.getByTestId('trading-panel')).toBeInTheDocument();
    expect(screen.getByTestId('order-form')).toBeInTheDocument();
    expect(screen.getByTestId('order-table')).toBeInTheDocument();
    expect(screen.getByTestId('position-table')).toBeInTheDocument();
    expect(screen.getByText('BTC-USD')).toBeInTheDocument();
    expect(screen.getByTestId('order-table')).toHaveTextContent('50000');
    expect(screen.getByTestId('position-table')).toHaveTextContent('50000');
  });
});
