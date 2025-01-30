import React from 'react';
import { render, screen, fireEvent, waitFor } from '../../utils/test-utils';
import { mockMarketData, mockOrderBook, mockTrades } from '../../utils/test-utils';
import TradingView from '@/components/trading/TradingView';

// 模拟WebSocket连接
jest.mock('@/hooks/useWebSocket', () => ({
  useWebSocket: () => ({
    data: mockMarketData,
    isConnected: true,
    error: null,
  }),
}));

describe('TradingView Component', () => {
  beforeEach(() => {
    // 重置所有模拟函数
    jest.clearAllMocks();
  });

  it('renders trading view with market data', () => {
    render(<TradingView symbol="BTC/USDT" />);

    // 验证价格显示
    expect(screen.getByText('50000')).toBeInTheDocument();
    expect(screen.getByText('+2.5%')).toBeInTheDocument();
  });

  it('displays order book correctly', () => {
    render(<TradingView symbol="BTC/USDT" />);

    // 验证买单和卖单
    mockOrderBook.bids.forEach(([price]) => {
      expect(screen.getByText(price.toString())).toBeInTheDocument();
    });

    mockOrderBook.asks.forEach(([price]) => {
      expect(screen.getByText(price.toString())).toBeInTheDocument();
    });
  });

  it('handles order placement', async () => {
    const mockPlaceOrder = jest.fn();
    render(<TradingView symbol="BTC/USDT" onPlaceOrder={mockPlaceOrder} />);

    // 填写订单表单
    fireEvent.change(screen.getByLabelText(/price/i), {
      target: { value: '50000' },
    });
    fireEvent.change(screen.getByLabelText(/amount/i), {
      target: { value: '1' },
    });

    // 提交订单
    fireEvent.click(screen.getByRole('button', { name: /buy/i }));

    await waitFor(() => {
      expect(mockPlaceOrder).toHaveBeenCalledWith({
        symbol: 'BTC/USDT',
        side: 'buy',
        price: 50000,
        amount: 1,
      });
    });
  });

  it('validates order input', async () => {
    render(<TradingView symbol="BTC/USDT" />);

    // 提交空表单
    fireEvent.click(screen.getByRole('button', { name: /buy/i }));

    // 验证错误消息
    await waitFor(() => {
      expect(screen.getByText(/price is required/i)).toBeInTheDocument();
      expect(screen.getByText(/amount is required/i)).toBeInTheDocument();
    });
  });

  it('displays recent trades', () => {
    render(<TradingView symbol="BTC/USDT" />);

    // 验证最近交易列表
    mockTrades.forEach(trade => {
      expect(screen.getByText(trade.price.toString())).toBeInTheDocument();
      expect(screen.getByText(trade.amount.toString())).toBeInTheDocument();
    });
  });

  it('handles WebSocket connection status', () => {
    // 模拟断开连接
    jest.mock('@/hooks/useWebSocket', () => ({
      useWebSocket: () => ({
        data: null,
        isConnected: false,
        error: 'Connection lost',
      }),
    }));

    render(<TradingView symbol="BTC/USDT" />);

    // 验证断开连接提示
    expect(screen.getByText(/connection lost/i)).toBeInTheDocument();
  });

  it('updates chart data correctly', async () => {
    render(<TradingView symbol="BTC/USDT" />);

    // 模拟新的市场数据
    const newMarketData = {
      ...mockMarketData,
      price: 51000,
      change: 3.0,
    };

    // 更新数据
    jest.mock('@/hooks/useWebSocket', () => ({
      useWebSocket: () => ({
        data: newMarketData,
        isConnected: true,
        error: null,
      }),
    }));

    // 验证更新后的价格
    await waitFor(() => {
      expect(screen.getByText('51000')).toBeInTheDocument();
      expect(screen.getByText('+3.0%')).toBeInTheDocument();
    });
  });
}); 