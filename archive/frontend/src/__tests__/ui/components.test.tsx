import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { TradingControls } from '../../components/TradingControls';
import { MarketAnalysis } from '../../components/MarketAnalysis';
import { TradingStats } from '../../components/TradingStats';
import { RiskSettings } from '../../components/RiskSettings';
import { TradeHistory } from '../../components/TradeHistory';
import { ErrorBoundary } from '../../components/ErrorBoundary';

describe('TradingControls Component', () => {
  const mockProps = {
    tokenAddress: 'test-token',
    onTrade: jest.fn(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders trading controls correctly', () => {
    render(<TradingControls {...mockProps} />);
    
    expect(screen.getByText(/trading controls/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/amount/i)).toBeInTheDocument();
    expect(screen.getByText(/buy/i)).toBeInTheDocument();
    expect(screen.getByText(/sell/i)).toBeInTheDocument();
  });

  test('handles trade execution', () => {
    render(<TradingControls {...mockProps} />);
    
    const amountInput = screen.getByLabelText(/amount/i);
    fireEvent.change(amountInput, { target: { value: '1.0' } });
    
    const buyButton = screen.getByText(/buy/i);
    fireEvent.click(buyButton);
    
    expect(mockProps.onTrade).toHaveBeenCalledWith('buy', 1.0);
  });

  test('validates input amount', () => {
    render(<TradingControls {...mockProps} />);
    
    const amountInput = screen.getByLabelText(/amount/i);
    fireEvent.change(amountInput, { target: { value: '-1.0' } });
    
    const buyButton = screen.getByText(/buy/i);
    fireEvent.click(buyButton);
    
    expect(screen.getByText(/please enter a valid amount/i)).toBeInTheDocument();
    expect(mockProps.onTrade).not.toHaveBeenCalled();
  });
});

describe('MarketAnalysis Component', () => {
  const mockProps = {
    tokenAddress: 'test-token',
  };

  const mockMarketData = {
    price: 1.5,
    volume_24h: 1000000,
    market_cap: 5000000,
    liquidity: 200000,
  };

  beforeEach(() => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve(mockMarketData),
      })
    ) as jest.Mock;
  });

  test('renders market data correctly', async () => {
    render(<MarketAnalysis {...mockProps} />);
    
    expect(await screen.findByText(/1.5 SOL/)).toBeInTheDocument();
    expect(await screen.findByText(/1,000,000/)).toBeInTheDocument();
    expect(await screen.findByText(/5,000,000/)).toBeInTheDocument();
  });

  test('handles loading state', () => {
    render(<MarketAnalysis {...mockProps} />);
    
    expect(screen.getByText(/loading market data/i)).toBeInTheDocument();
  });

  test('handles error state', async () => {
    global.fetch = jest.fn(() => Promise.reject('API Error')) as jest.Mock;
    
    render(<MarketAnalysis {...mockProps} />);
    
    expect(await screen.findByText(/error/i)).toBeInTheDocument();
  });
});

describe('TradingStats Component', () => {
  const mockProps = {
    tokenAddress: 'test-token',
  };

  const mockStats = {
    daily_volume: [1000, 1200, 1100],
    profit_loss: [0.1, -0.05, 0.15],
    win_rate: 0.67,
    average_win: 0.12,
    average_loss: -0.05,
    largest_win: 0.2,
    largest_loss: -0.1,
    trade_distribution: [
      { label: 'Wins', value: 67 },
      { label: 'Losses', value: 33 },
    ],
    timestamps: ['2024-01-01', '2024-01-02', '2024-01-03'],
  };

  beforeEach(() => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve(mockStats),
      })
    ) as jest.Mock;
  });

  test('renders trading statistics correctly', async () => {
    render(<TradingStats {...mockProps} />);
    
    expect(await screen.findByText(/67.0%/)).toBeInTheDocument();
    expect(await screen.findByText(/0.12 SOL/)).toBeInTheDocument();
    expect(await screen.findByText(/-0.05 SOL/)).toBeInTheDocument();
  });

  test('handles timeframe selection', async () => {
    render(<TradingStats {...mockProps} />);
    
    const weekButton = screen.getByText(/7D/);
    fireEvent.click(weekButton);
    
    expect(global.fetch).toHaveBeenCalledWith(
      expect.stringContaining('timeframe=7d')
    );
  });

  test('renders charts correctly', async () => {
    const { container } = render(<TradingStats {...mockProps} />);
    
    // Wait for charts to render
    await screen.findByText(/67.0%/);
    
    // Check if charts are rendered
    expect(container.querySelector('.recharts-surface')).toBeInTheDocument();
  });

  test('handles data updates', async () => {
    const { rerender } = render(<TradingStats {...mockProps} />);
    
    // Initial render
    expect(await screen.findByText(/67.0%/)).toBeInTheDocument();
    
    // Update mock data
    const updatedStats = {
      ...mockStats,
      win_rate: 0.75,
    };
    
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve(updatedStats),
      })
    ) as jest.Mock;
    
    // Rerender with new data
    rerender(<TradingStats {...mockProps} />);
    
    expect(await screen.findByText(/75.0%/)).toBeInTheDocument();
  });
});

describe('RiskSettings Component', () => {
  const mockProps = {
    tokenAddress: 'test-token',
    onUpdate: jest.fn(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders risk settings form correctly', () => {
    render(<RiskSettings {...mockProps} />);
    
    expect(screen.getByLabelText(/stop loss/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/take profit/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/max position size/i)).toBeInTheDocument();
    expect(screen.getByText(/save settings/i)).toBeInTheDocument();
  });

  test('handles settings update', () => {
    render(<RiskSettings {...mockProps} />);
    
    const stopLossInput = screen.getByLabelText(/stop loss/i);
    const takeProfitInput = screen.getByLabelText(/take profit/i);
    const maxSizeInput = screen.getByLabelText(/max position size/i);
    
    fireEvent.change(stopLossInput, { target: { value: '5' } });
    fireEvent.change(takeProfitInput, { target: { value: '10' } });
    fireEvent.change(maxSizeInput, { target: { value: '1000' } });
    
    const saveButton = screen.getByText(/save settings/i);
    fireEvent.click(saveButton);
    
    expect(mockProps.onUpdate).toHaveBeenCalledWith({
      stopLoss: 5,
      takeProfit: 10,
      maxPositionSize: 1000,
    });
  });

  test('validates input values', () => {
    render(<RiskSettings {...mockProps} />);
    
    const stopLossInput = screen.getByLabelText(/stop loss/i);
    fireEvent.change(stopLossInput, { target: { value: '-5' } });
    
    const saveButton = screen.getByText(/save settings/i);
    fireEvent.click(saveButton);
    
    expect(screen.getByText(/stop loss must be positive/i)).toBeInTheDocument();
    expect(mockProps.onUpdate).not.toHaveBeenCalled();
  });
});

describe('TradeHistory Component', () => {
  const mockProps = {
    tokenAddress: 'test-token',
  };

  const mockTrades = [
    {
      id: 1,
      type: 'buy',
      amount: 1.0,
      price: 100.0,
      timestamp: '2024-01-01T10:00:00Z',
      profit: 0.1,
    },
    {
      id: 2,
      type: 'sell',
      amount: 0.5,
      price: 110.0,
      timestamp: '2024-01-01T11:00:00Z',
      profit: 0.05,
    },
  ];

  beforeEach(() => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve(mockTrades),
      })
    ) as jest.Mock;
  });

  test('renders trade history correctly', async () => {
    render(<TradeHistory {...mockProps} />);
    
    expect(await screen.findByText(/buy 1.0/i)).toBeInTheDocument();
    expect(await screen.findByText(/sell 0.5/i)).toBeInTheDocument();
    expect(await screen.findByText(/\+0.1 SOL/i)).toBeInTheDocument();
  });

  test('handles empty trade history', async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve([]),
      })
    ) as jest.Mock;
    
    render(<TradeHistory {...mockProps} />);
    
    expect(await screen.findByText(/no trades yet/i)).toBeInTheDocument();
  });

  test('handles pagination', async () => {
    render(<TradeHistory {...mockProps} />);
    
    const nextButton = await screen.findByText(/next/i);
    fireEvent.click(nextButton);
    
    expect(global.fetch).toHaveBeenCalledWith(
      expect.stringContaining('page=2')
    );
  });

  test('filters trades by type', async () => {
    render(<TradeHistory {...mockProps} />);
    
    const filterSelect = await screen.findByLabelText(/filter by type/i);
    fireEvent.change(filterSelect, { target: { value: 'buy' } });
    
    expect(global.fetch).toHaveBeenCalledWith(
      expect.stringContaining('type=buy')
    );
  });
});

describe('ErrorBoundary Component', () => {
  const ThrowError = () => {
    throw new Error('Test error');
  };

  test('catches and displays errors', () => {
    const { container } = render(
      <ErrorBoundary>
        <ThrowError />
      </ErrorBoundary>
    );
    
    expect(screen.getByText(/something went wrong/i)).toBeInTheDocument();
    expect(screen.getByText(/test error/i)).toBeInTheDocument();
  });

  test('renders children when no error', () => {
    render(
      <ErrorBoundary>
        <div>Test Content</div>
      </ErrorBoundary>
    );
    
    expect(screen.getByText(/test content/i)).toBeInTheDocument();
  });

  test('handles retry action', () => {
    const onRetry = jest.fn();
    
    render(
      <ErrorBoundary onRetry={onRetry}>
        <ThrowError />
      </ErrorBoundary>
    );
    
    const retryButton = screen.getByText(/retry/i);
    fireEvent.click(retryButton);
    
    expect(onRetry).toHaveBeenCalled();
  });
}); 