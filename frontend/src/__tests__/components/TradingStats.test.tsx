import React from 'react';
import { render, screen } from '@testing-library/react';
import { TradingStats } from '../../components/TradingStats';
import { server } from '../../mocks/server';

describe('TradingStats', () => {
  const mockState = {
    totalTrades: 100,
    successfulTrades: 75,
    totalProfit: 1500.50,
    lastTradeTime: '2024-03-15T10:30:00Z',
    winRate: 75,
    averageProfit: 25.5,
    averageLoss: 15.2,
    profitFactor: 1.68,
    maxDrawdown: 12.5,
    sharpeRatio: 2.1
  };

  describe('Rendering', () => {
    it('renders all trading statistics correctly', () => {
      render(<TradingStats state={mockState} balance={1000} />);

      expect(screen.getByText(/total trades/i)).toBeInTheDocument();
      expect(screen.getByText('100')).toBeInTheDocument();

      expect(screen.getByText(/successful trades/i)).toBeInTheDocument();
      expect(screen.getByText('75')).toBeInTheDocument();

      expect(screen.getByText(/total profit/i)).toBeInTheDocument();
      expect(screen.getByText(/1,500\.50/)).toBeInTheDocument();

      expect(screen.getByText(/win rate/i)).toBeInTheDocument();
      expect(screen.getByText(/75%/)).toBeInTheDocument();

      expect(screen.getByText(/average profit/i)).toBeInTheDocument();
      expect(screen.getByText(/25\.5/)).toBeInTheDocument();

      expect(screen.getByText(/average loss/i)).toBeInTheDocument();
      expect(screen.getByText(/15\.2/)).toBeInTheDocument();

      expect(screen.getByText(/profit factor/i)).toBeInTheDocument();
      expect(screen.getByText(/1\.68/)).toBeInTheDocument();

      expect(screen.getByText(/max drawdown/i)).toBeInTheDocument();
      expect(screen.getByText(/12\.5%/)).toBeInTheDocument();

      expect(screen.getByText(/sharpe ratio/i)).toBeInTheDocument();
      expect(screen.getByText(/2\.1/)).toBeInTheDocument();
    });

    it('displays wallet balance correctly', () => {
      render(<TradingStats state={mockState} balance={1000} />);
      expect(screen.getByText(/1,000/)).toBeInTheDocument();
    });

    it('formats last trade time correctly', () => {
      render(<TradingStats state={mockState} balance={1000} />);
      expect(screen.getByText(/march 15, 2024/i)).toBeInTheDocument();
      expect(screen.getByText(/10:30/i)).toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('displays loading skeleton when isLoading is true', () => {
      render(<TradingStats state={mockState} balance={1000} isLoading={true} />);
      expect(screen.getByTestId('stats-loading-skeleton')).toBeInTheDocument();
    });

    it('hides loading skeleton when isLoading is false', () => {
      render(<TradingStats state={mockState} balance={1000} isLoading={false} />);
      expect(screen.queryByTestId('stats-loading-skeleton')).not.toBeInTheDocument();
    });
  });

  describe('Error State', () => {
    it('displays error message when error prop is provided', () => {
      const errorMessage = 'Failed to load trading stats';
      render(<TradingStats state={mockState} balance={1000} error={errorMessage} />);
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });

    it('hides stats when error is present', () => {
      const errorMessage = 'Failed to load trading stats';
      render(<TradingStats state={mockState} balance={1000} error={errorMessage} />);
      expect(screen.queryByText(/total trades/i)).not.toBeInTheDocument();
    });
  });

  describe('Empty State', () => {
    const emptyState = {
      totalTrades: 0,
      successfulTrades: 0,
      totalProfit: 0,
      lastTradeTime: '',
      winRate: 0,
      averageProfit: 0,
      averageLoss: 0,
      profitFactor: 0,
      maxDrawdown: 0,
      sharpeRatio: 0
    };

    it('handles empty state gracefully', () => {
      render(<TradingStats state={emptyState} balance={0} />);
      expect(screen.getByText(/no trades yet/i)).toBeInTheDocument();
    });

    it('displays zero values correctly', () => {
      render(<TradingStats state={emptyState} balance={0} />);
      expect(screen.getByText('0')).toBeInTheDocument();
      expect(screen.getByText('0%')).toBeInTheDocument();
    });
  });

  describe('Conditional Styling', () => {
    it('applies positive styling to profit values', () => {
      render(<TradingStats state={mockState} balance={1000} />);
      const profitElement = screen.getByText(/1,500\.50/).closest('div');
      expect(profitElement).toHaveStyle({ color: expect.stringContaining('success') });
    });

    it('applies negative styling to loss values', () => {
      const stateWithLoss = {
        ...mockState,
        totalProfit: -500.25
      };
      render(<TradingStats state={stateWithLoss} balance={1000} />);
      const lossElement = screen.getByText(/-500\.25/).closest('div');
      expect(lossElement).toHaveStyle({ color: expect.stringContaining('error') });
    });

    it('applies warning styling to high drawdown', () => {
      const stateWithHighDrawdown = {
        ...mockState,
        maxDrawdown: 25.5
      };
      render(<TradingStats state={stateWithHighDrawdown} balance={1000} />);
      const drawdownElement = screen.getByText(/25\.5%/).closest('div');
      expect(drawdownElement).toHaveStyle({ color: expect.stringContaining('warning') });
    });
  });

  describe('Number Formatting', () => {
    it('formats large numbers with commas', () => {
      const stateWithLargeNumbers = {
        ...mockState,
        totalTrades: 1000000,
        totalProfit: 1234567.89
      };
      render(<TradingStats state={stateWithLargeNumbers} balance={1000000} />);
      expect(screen.getByText(/1,000,000/)).toBeInTheDocument();
      expect(screen.getByText(/1,234,567\.89/)).toBeInTheDocument();
    });

    it('formats percentages correctly', () => {
      const stateWithDecimalPercentages = {
        ...mockState,
        winRate: 66.67,
        maxDrawdown: 12.34
      };
      render(<TradingStats state={stateWithDecimalPercentages} balance={1000} />);
      expect(screen.getByText(/66\.67%/)).toBeInTheDocument();
      expect(screen.getByText(/12\.34%/)).toBeInTheDocument();
    });
  });
}); 