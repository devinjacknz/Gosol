import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import OrderTable from '@/components/Trading/OrderTable';
import React from 'react';
import '@testing-library/jest-dom';

describe('OrderTable', () => {
  const mockOrders = [
    { type: 'market', size: '1.0', price: '50000' },
    { type: 'limit', size: '0.5', price: '49000' }
  ];

  it('renders order table with data', () => {
    render(<OrderTable orders={mockOrders} />);
    
    expect(screen.getByTestId('order-table')).toBeInTheDocument();
    expect(screen.getByText('Orders')).toBeInTheDocument();
    expect(screen.getByText('Type')).toBeInTheDocument();
    expect(screen.getByText('Size')).toBeInTheDocument();
    expect(screen.getByText('Price')).toBeInTheDocument();
    expect(screen.getByText('market')).toBeInTheDocument();
    expect(screen.getByText('limit')).toBeInTheDocument();
  });
});
