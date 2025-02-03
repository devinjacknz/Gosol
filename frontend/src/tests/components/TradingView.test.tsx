import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import TradingView from '@/pages/TradingView';
import tradingReducer from '@/store/trading/tradingSlice';
import { apiService } from '@/services/api';
import { Order, Position } from '@/types/trading';
import '@testing-library/jest-dom';

vi.mock('@/services/api');

describe('TradingView', () => {
  const mockOrders: Order[] = [
    {
      id: '1',
      type: 'market',
      side: 'buy',
      symbol: 'BTC-USD',
      size: 1.0,
      status: 'open',
      timestamp: Date.now()
    }
  ];

  const mockPositions: Position[] = [
    {
      symbol: 'BTC-USD',
      size: 1.0,
      entryPrice: 50000,
      markPrice: 51000,
      pnl: 1000,
      status: 'open',
      lastUpdated: Date.now()
    }
  ];

  const mockStore = configureStore({
    reducer: {
      trading: tradingReducer
    },
    preloadedState: {
      trading: {
        orders: mockOrders,
        positions: mockPositions,
        marketData: {},
        loading: false,
        error: null
      }
    }
  } as any);

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(apiService.getOrders).mockResolvedValue(mockOrders);
    vi.mocked(apiService.getPositions).mockResolvedValue(mockPositions);
    vi.mocked(apiService.placeOrder).mockResolvedValue({ success: true, order: mockOrders[0] });
    vi.mocked(apiService.cancelOrder).mockResolvedValue({ success: true });
  });

  const renderComponent = async () => {
    const result = render(
      <Provider store={mockStore}>
        <TradingView />
      </Provider>
    );
    await waitFor(() => {
      expect(screen.getByTestId('trading-panel')).toBeInTheDocument();
    });
    return result;
  };

  it('renders trading components and fetches initial data', async () => {
    await renderComponent();

    await waitFor(() => {
      expect(screen.getByTestId('order-form')).toBeInTheDocument();
      expect(screen.getByTestId('order-table')).toBeInTheDocument();
      expect(screen.getByTestId('position-table')).toBeInTheDocument();
    });

    await waitFor(() => {
      expect(apiService.getOrders).toHaveBeenCalled();
      expect(apiService.getPositions).toHaveBeenCalled();
    });
  });

  it('handles order placement', async () => {
    await renderComponent();

    await waitFor(() => {
      expect(screen.getByTestId('order-size-input')).toBeInTheDocument();
      expect(screen.getByTestId('submit-order-button')).toBeInTheDocument();
    });

    await act(async () => {
      fireEvent.change(screen.getByTestId('order-size-input'), { target: { value: '1.5' } });
      fireEvent.click(screen.getByTestId('submit-order-button'));
    });

    await waitFor(() => {
      expect(apiService.placeOrder).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'market',
          side: 'buy',
          symbol: 'BTC-USD',
          size: 1.5
        })
      );
    });
  });

  it('handles order cancellation', async () => {
    await renderComponent();

    await waitFor(() => {
      expect(screen.getByTestId('cancel-order-1')).toBeInTheDocument();
    });

    await act(async () => {
      fireEvent.click(screen.getByTestId('cancel-order-1'));
    });

    await waitFor(() => {
      expect(apiService.cancelOrder).toHaveBeenCalledWith('1');
    });
  });

  it('handles position closing', async () => {
    await renderComponent();

    await waitFor(() => {
      expect(screen.getByTestId('close-position-BTC-USD')).toBeInTheDocument();
    });

    await act(async () => {
      fireEvent.click(screen.getByTestId('close-position-BTC-USD'));
    });

    await waitFor(() => {
      expect(apiService.placeOrder).toHaveBeenCalledWith(
        expect.objectContaining({
          symbol: 'BTC-USD',
          type: 'market',
          side: 'sell',
          size: 1.0
        })
      );
    });
  });

  it('handles API errors gracefully', async () => {
    vi.mocked(apiService.placeOrder).mockRejectedValueOnce(new Error('API Error'));
    await renderComponent();

    await waitFor(() => {
      expect(screen.getByTestId('order-size-input')).toBeInTheDocument();
    });

    await act(async () => {
      fireEvent.change(screen.getByTestId('order-size-input'), { target: { value: '1.0' } });
      fireEvent.click(screen.getByTestId('submit-order-button'));
    });

    await waitFor(() => {
      expect(screen.getByText(/error/i)).toBeInTheDocument();
    });
  });
});
