import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { server } from '../../mocks/server';
import RiskSettings from '../../components/RiskSettings';

const mockToken = 'So11111111111111111111111111111111111111111';

describe('RiskSettings', () => {
  const mockSettings = {
    maxPositionSize: 1000,
    maxDrawdown: 20,
    maxDailyTrades: 10,
    maxLeverage: 3,
    stopLossPercent: 5,
    takeProfitPercent: 10,
    riskPerTrade: 2,
    timeoutAfterLoss: 300 // 5 minutes
  };

  beforeEach(() => {
    server.use(
      rest.get('*/api/risk-settings/*', (req, res, ctx) => {
        return res(ctx.json(mockSettings));
      })
    );
  });

  describe('Initial Render', () => {
    it('shows loading state initially', () => {
      render(<RiskSettings tokenAddress={mockToken} onUpdate={jest.fn()} />);
      expect(screen.getByText(/loading settings/i)).toBeInTheDocument();
    });

    it('renders all risk setting controls', async () => {
      render(<RiskSettings tokenAddress={mockToken} onUpdate={jest.fn()} />);

      await waitFor(() => {
        expect(screen.getByLabelText(/maximum position size/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/maximum drawdown/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/maximum daily trades/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/maximum leverage/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/stop loss/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/take profit/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/risk per trade/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/timeout after loss/i)).toBeInTheDocument();
      });
    });

    it('displays initial values correctly', async () => {
      render(<RiskSettings tokenAddress={mockToken} onUpdate={jest.fn()} />);

      await waitFor(() => {
        expect(screen.getByLabelText(/maximum position size/i)).toHaveValue(1000);
        expect(screen.getByLabelText(/maximum drawdown/i)).toHaveValue(20);
        expect(screen.getByLabelText(/maximum daily trades/i)).toHaveValue(10);
        expect(screen.getByLabelText(/maximum leverage/i)).toHaveValue(3);
        expect(screen.getByLabelText(/stop loss/i)).toHaveValue(5);
        expect(screen.getByLabelText(/take profit/i)).toHaveValue(10);
        expect(screen.getByLabelText(/risk per trade/i)).toHaveValue(2);
        expect(screen.getByLabelText(/timeout after loss/i)).toHaveValue(300);
      });
    });
  });

  describe('User Interactions', () => {
    it('handles position size input changes', async () => {
      const onUpdate = jest.fn();
      render(<RiskSettings tokenAddress={mockToken} onUpdate={onUpdate} />);

      await waitFor(() => {
        const input = screen.getByLabelText(/maximum position size/i);
        fireEvent.change(input, { target: { value: '2000' } });
      });

      expect(onUpdate).toHaveBeenCalledWith({
        ...mockSettings,
        maxPositionSize: 2000
      });
    });

    it('handles drawdown input changes', async () => {
      const onUpdate = jest.fn();
      render(<RiskSettings tokenAddress={mockToken} onUpdate={onUpdate} />);

      await waitFor(() => {
        const input = screen.getByLabelText(/maximum drawdown/i);
        fireEvent.change(input, { target: { value: '25' } });
      });

      expect(onUpdate).toHaveBeenCalledWith({
        ...mockSettings,
        maxDrawdown: 25
      });
    });

    it('handles daily trades limit changes', async () => {
      const onUpdate = jest.fn();
      render(<RiskSettings tokenAddress={mockToken} onUpdate={onUpdate} />);

      await waitFor(() => {
        const input = screen.getByLabelText(/maximum daily trades/i);
        fireEvent.change(input, { target: { value: '15' } });
      });

      expect(onUpdate).toHaveBeenCalledWith({
        ...mockSettings,
        maxDailyTrades: 15
      });
    });
  });

  describe('Validation', () => {
    it('prevents negative values', async () => {
      const onUpdate = jest.fn();
      render(<RiskSettings tokenAddress={mockToken} onUpdate={onUpdate} />);

      await waitFor(() => {
        const input = screen.getByLabelText(/maximum position size/i);
        fireEvent.change(input, { target: { value: '-1000' } });
      });

      expect(onUpdate).not.toHaveBeenCalled();
      expect(screen.getByText(/value must be positive/i)).toBeInTheDocument();
    });

    it('enforces maximum leverage limit', async () => {
      const onUpdate = jest.fn();
      render(<RiskSettings tokenAddress={mockToken} onUpdate={onUpdate} />);

      await waitFor(() => {
        const input = screen.getByLabelText(/maximum leverage/i);
        fireEvent.change(input, { target: { value: '11' } });
      });

      expect(onUpdate).not.toHaveBeenCalled();
      expect(screen.getByText(/maximum leverage cannot exceed 10/i)).toBeInTheDocument();
    });

    it('validates percentage inputs', async () => {
      const onUpdate = jest.fn();
      render(<RiskSettings tokenAddress={mockToken} onUpdate={onUpdate} />);

      await waitFor(() => {
        const input = screen.getByLabelText(/stop loss/i);
        fireEvent.change(input, { target: { value: '101' } });
      });

      expect(onUpdate).not.toHaveBeenCalled();
      expect(screen.getByText(/percentage must be between 0 and 100/i)).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('displays error message on fetch failure', async () => {
      server.use(
        rest.get('*/api/risk-settings/*', (req, res, ctx) => {
          return res(ctx.status(500));
        })
      );

      render(<RiskSettings tokenAddress={mockToken} onUpdate={jest.fn()} />);

      await waitFor(() => {
        expect(screen.getByText(/failed to load risk settings/i)).toBeInTheDocument();
      });
    });

    it('handles network errors gracefully', async () => {
      server.use(
        rest.get('*/api/risk-settings/*', (req, res) => {
          return res.networkError('Failed to connect');
        })
      );

      render(<RiskSettings tokenAddress={mockToken} onUpdate={jest.fn()} />);

      await waitFor(() => {
        expect(screen.getByText(/network error/i)).toBeInTheDocument();
      });
    });
  });

  describe('Save Functionality', () => {
    it('saves settings successfully', async () => {
      const onUpdate = jest.fn();
      render(<RiskSettings tokenAddress={mockToken} onUpdate={onUpdate} />);

      await waitFor(() => {
        const saveButton = screen.getByRole('button', { name: /save settings/i });
        fireEvent.click(saveButton);
      });

      expect(onUpdate).toHaveBeenCalledWith(mockSettings);
      expect(screen.getByText(/settings saved successfully/i)).toBeInTheDocument();
    });

    it('shows error message on save failure', async () => {
      server.use(
        rest.post('*/api/risk-settings/*', (req, res, ctx) => {
          return res(ctx.status(500));
        })
      );

      render(<RiskSettings tokenAddress={mockToken} onUpdate={jest.fn()} />);

      await waitFor(() => {
        const saveButton = screen.getByRole('button', { name: /save settings/i });
        fireEvent.click(saveButton);
      });

      expect(screen.getByText(/failed to save settings/i)).toBeInTheDocument();
    });
  });

  describe('Reset Functionality', () => {
    it('resets settings to defaults', async () => {
      const onUpdate = jest.fn();
      render(<RiskSettings tokenAddress={mockToken} onUpdate={onUpdate} />);

      await waitFor(() => {
        const resetButton = screen.getByRole('button', { name: /reset to defaults/i });
        fireEvent.click(resetButton);
      });

      expect(onUpdate).toHaveBeenCalledWith({
        maxPositionSize: 1000,
        maxDrawdown: 20,
        maxDailyTrades: 10,
        maxLeverage: 3,
        stopLossPercent: 5,
        takeProfitPercent: 10,
        riskPerTrade: 2,
        timeoutAfterLoss: 300
      });
    });

    it('shows confirmation dialog before reset', async () => {
      render(<RiskSettings tokenAddress={mockToken} onUpdate={jest.fn()} />);

      await waitFor(() => {
        const resetButton = screen.getByRole('button', { name: /reset to defaults/i });
        fireEvent.click(resetButton);
      });

      expect(screen.getByText(/are you sure/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /confirm/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
    });
  });
}); 