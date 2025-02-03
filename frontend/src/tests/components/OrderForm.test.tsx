import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import OrderForm from '@/components/Trading/OrderForm';
import React from 'react';
import '@testing-library/jest-dom';

describe('OrderForm', () => {
  const mockSubmit = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders order form components', () => {
    render(<OrderForm onSubmit={mockSubmit} />);
    
    expect(screen.getByTestId('order-form')).toBeInTheDocument();
    expect(screen.getByTestId('submit-order-button')).toBeInTheDocument();
    expect(screen.getByTestId('order-type-select')).toBeInTheDocument();
    expect(screen.getByTestId('order-side-select')).toBeInTheDocument();
    expect(screen.getByTestId('order-size-input')).toBeInTheDocument();
  });

  it('handles market order submission', () => {
    render(<OrderForm onSubmit={mockSubmit} />);
    
    fireEvent.change(screen.getByTestId('order-size-input'), { target: { value: '1.5' } });
    fireEvent.click(screen.getByTestId('submit-order-button'));

    expect(mockSubmit).toHaveBeenCalledWith({
      type: 'market',
      side: 'buy',
      symbol: 'BTC-USD',
      size: 1.5,
      price: undefined
    });
  });

  it('handles limit order submission', () => {
    render(<OrderForm onSubmit={mockSubmit} />);
    
    fireEvent.change(screen.getByTestId('order-type-select'), { target: { value: 'limit' } });
    fireEvent.change(screen.getByTestId('order-size-input'), { target: { value: '2.0' } });
    const priceInput = screen.getByLabelText('Price');
    fireEvent.change(priceInput, { target: { value: '50000' } });
    fireEvent.click(screen.getByTestId('submit-order-button'));

    expect(mockSubmit).toHaveBeenCalledWith({
      type: 'limit',
      side: 'buy',
      symbol: 'BTC-USD',
      size: 2.0,
      price: 50000
    });
  });
});
