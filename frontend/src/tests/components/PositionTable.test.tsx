import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import PositionTable from '@/components/Trading/PositionTable';
import React from 'react';
import '@testing-library/jest-dom';

describe('PositionTable', () => {
  const mockPositions = [
    { symbol: 'BTC-USD', size: '1.0', entryPrice: '50000' },
    { symbol: 'ETH-USD', size: '10.0', entryPrice: '3000' }
  ];

  it('renders position table with data', () => {
    render(<PositionTable positions={mockPositions} />);
    
    expect(screen.getByTestId('position-table')).toBeInTheDocument();
    expect(screen.getByText('Positions')).toBeInTheDocument();
    expect(screen.getByText('Symbol')).toBeInTheDocument();
    expect(screen.getByText('Size')).toBeInTheDocument();
    expect(screen.getByText('Entry Price')).toBeInTheDocument();
    expect(screen.getByText('BTC-USD')).toBeInTheDocument();
    expect(screen.getByText('ETH-USD')).toBeInTheDocument();
  });
});
