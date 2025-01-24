import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { TradingControls } from '../../components/TradingControls';

describe('TradingControls', () => {
  const defaultConfig = {
    maxAmount: 1000,
    stopLoss: 5,
    takeProfit: 10,
    walletAddress: ''
  };

  const mockProps = {
    config: defaultConfig,
    isTrading: false,
    onConfigChange: jest.fn(),
    onToggleTrading: jest.fn(),
    onTransferProfit: jest.fn()
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Rendering', () => {
    it('renders all input fields with correct initial values', () => {
      render(<TradingControls {...mockProps} />);

      expect(screen.getByLabelText(/maximum amount/i)).toHaveValue(1000);
      expect(screen.getByLabelText(/stop loss/i)).toHaveValue(5);
      expect(screen.getByLabelText(/take profit/i)).toHaveValue(10);
      expect(screen.getByLabelText(/wallet address/i)).toHaveValue('');
    });

    it('renders trading control buttons', () => {
      render(<TradingControls {...mockProps} />);

      expect(screen.getByRole('button', { name: /start trading/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /transfer profit/i })).toBeInTheDocument();
    });

    it('shows correct button text based on trading state', () => {
      const { rerender } = render(<TradingControls {...mockProps} />);
      expect(screen.getByRole('button', { name: /start trading/i })).toBeInTheDocument();

      rerender(<TradingControls {...mockProps} isTrading={true} />);
      expect(screen.getByRole('button', { name: /stop trading/i })).toBeInTheDocument();
    });
  });

  describe('User Interactions', () => {
    it('handles maximum amount input changes', async () => {
      render(<TradingControls {...mockProps} />);
      const input = screen.getByLabelText(/maximum amount/i);

      await userEvent.clear(input);
      await userEvent.type(input, '2000');

      expect(mockProps.onConfigChange).toHaveBeenCalledWith({
        ...defaultConfig,
        maxAmount: 2000
      });
    });

    it('handles stop loss input changes', async () => {
      render(<TradingControls {...mockProps} />);
      const input = screen.getByLabelText(/stop loss/i);

      await userEvent.clear(input);
      await userEvent.type(input, '8');

      expect(mockProps.onConfigChange).toHaveBeenCalledWith({
        ...defaultConfig,
        stopLoss: 8
      });
    });

    it('handles take profit input changes', async () => {
      render(<TradingControls {...mockProps} />);
      const input = screen.getByLabelText(/take profit/i);

      await userEvent.clear(input);
      await userEvent.type(input, '15');

      expect(mockProps.onConfigChange).toHaveBeenCalledWith({
        ...defaultConfig,
        takeProfit: 15
      });
    });

    it('handles wallet address input changes', async () => {
      render(<TradingControls {...mockProps} />);
      const input = screen.getByLabelText(/wallet address/i);
      const testAddress = 'test-wallet-address';

      await userEvent.clear(input);
      await userEvent.type(input, testAddress);

      expect(mockProps.onConfigChange).toHaveBeenCalledWith({
        ...defaultConfig,
        walletAddress: testAddress
      });
    });

    it('calls onToggleTrading when trading button is clicked', () => {
      render(<TradingControls {...mockProps} />);
      const button = screen.getByRole('button', { name: /start trading/i });

      fireEvent.click(button);
      expect(mockProps.onToggleTrading).toHaveBeenCalled();
    });

    it('calls onTransferProfit when transfer button is clicked', () => {
      render(<TradingControls {...mockProps} />);
      const button = screen.getByRole('button', { name: /transfer profit/i });

      fireEvent.click(button);
      expect(mockProps.onTransferProfit).toHaveBeenCalled();
    });
  });

  describe('Validation', () => {
    it('prevents negative values in maximum amount input', async () => {
      render(<TradingControls {...mockProps} />);
      const input = screen.getByLabelText(/maximum amount/i);

      await userEvent.clear(input);
      await userEvent.type(input, '-100');

      expect(mockProps.onConfigChange).not.toHaveBeenCalledWith({
        ...defaultConfig,
        maxAmount: -100
      });
    });

    it('prevents negative values in stop loss input', async () => {
      render(<TradingControls {...mockProps} />);
      const input = screen.getByLabelText(/stop loss/i);

      await userEvent.clear(input);
      await userEvent.type(input, '-5');

      expect(mockProps.onConfigChange).not.toHaveBeenCalledWith({
        ...defaultConfig,
        stopLoss: -5
      });
    });

    it('prevents negative values in take profit input', async () => {
      render(<TradingControls {...mockProps} />);
      const input = screen.getByLabelText(/take profit/i);

      await userEvent.clear(input);
      await userEvent.type(input, '-10');

      expect(mockProps.onConfigChange).not.toHaveBeenCalledWith({
        ...defaultConfig,
        takeProfit: -10
      });
    });

    it('validates wallet address format', async () => {
      render(<TradingControls {...mockProps} />);
      const input = screen.getByLabelText(/wallet address/i);
      const invalidAddress = 'invalid!@#';

      await userEvent.clear(input);
      await userEvent.type(input, invalidAddress);

      expect(screen.getByText(/invalid wallet address/i)).toBeInTheDocument();
    });
  });

  describe('Disabled States', () => {
    it('disables inputs while trading is active', () => {
      render(<TradingControls {...mockProps} isTrading={true} />);

      expect(screen.getByLabelText(/maximum amount/i)).toBeDisabled();
      expect(screen.getByLabelText(/stop loss/i)).toBeDisabled();
      expect(screen.getByLabelText(/take profit/i)).toBeDisabled();
      expect(screen.getByLabelText(/wallet address/i)).toBeDisabled();
    });

    it('enables inputs while trading is inactive', () => {
      render(<TradingControls {...mockProps} isTrading={false} />);

      expect(screen.getByLabelText(/maximum amount/i)).toBeEnabled();
      expect(screen.getByLabelText(/stop loss/i)).toBeEnabled();
      expect(screen.getByLabelText(/take profit/i)).toBeEnabled();
      expect(screen.getByLabelText(/wallet address/i)).toBeEnabled();
    });
  });
}); 