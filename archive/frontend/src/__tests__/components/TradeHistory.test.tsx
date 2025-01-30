import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { rest } from 'msw';
import { server } from '../../mocks/server';
import TradeHistory from '../../components/TradeHistory';

const mockToken = 'So11111111111111111111111111111111111111111';

describe('TradeHistory', () => {
  const mockTrades = [
    {
      id: '1',
      type: 'buy',
      amount: 100,
      price: 1.5,
      timestamp: '2024-03-15T10:30:00Z',
      profit: 0,
      status: 'completed'
    },
    {
      id: '2',
      type: 'sell',
      amount: 50,
      price: 1.8,
      timestamp: '2024-03-15T11:00:00Z',
      profit: 15,
      status: 'completed'
    },
    {
      id: '3',
      type: 'buy',
      amount: 75,
      price: 1.6,
      timestamp: '2024-03-15T11:30:00Z',
      profit: 0,
      status: 'pending'
    }
  ];

  beforeEach(() => {
    server.use(
      rest.get('*/api/trades/*', (req, res, ctx) => {
        return res(ctx.json(mockTrades));
      })
    );
  });

  describe('Initial Render', () => {
    it('shows loading state initially', () => {
      render(<TradeHistory tokenAddress={mockToken} />);
      expect(screen.getByText(/loading trades/i)).toBeInTheDocument();
    });

    it('renders trade history table', async () => {
      render(<TradeHistory tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByRole('table')).toBeInTheDocument();
        expect(screen.getByText(/trade history/i)).toBeInTheDocument();
      });
    });

    it('displays table headers correctly', async () => {
      render(<TradeHistory tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/type/i)).toBeInTheDocument();
        expect(screen.getByText(/amount/i)).toBeInTheDocument();
        expect(screen.getByText(/price/i)).toBeInTheDocument();
        expect(screen.getByText(/profit/i)).toBeInTheDocument();
        expect(screen.getByText(/status/i)).toBeInTheDocument();
        expect(screen.getByText(/time/i)).toBeInTheDocument();
      });
    });
  });

  describe('Trade Display', () => {
    it('displays trade details correctly', async () => {
      render(<TradeHistory tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText('100')).toBeInTheDocument();
        expect(screen.getByText('1.5')).toBeInTheDocument();
        expect(screen.getByText(/completed/i)).toBeInTheDocument();
      });
    });

    it('formats trade type with appropriate styling', async () => {
      render(<TradeHistory tokenAddress={mockToken} />);

      await waitFor(() => {
        const buyCell = screen.getByText(/buy/i);
        const sellCell = screen.getByText(/sell/i);
        
        expect(buyCell).toHaveStyle({ color: expect.stringContaining('success') });
        expect(sellCell).toHaveStyle({ color: expect.stringContaining('error') });
      });
    });

    it('formats timestamps correctly', async () => {
      render(<TradeHistory tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/10:30/i)).toBeInTheDocument();
        expect(screen.getByText(/11:00/i)).toBeInTheDocument();
      });
    });

    it('displays profit with appropriate styling', async () => {
      render(<TradeHistory tokenAddress={mockToken} />);

      await waitFor(() => {
        const profitCell = screen.getByText('15');
        expect(profitCell).toHaveStyle({ color: expect.stringContaining('success') });
      });
    });
  });

  describe('Pagination', () => {
    it('displays pagination controls', async () => {
      render(<TradeHistory tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /next/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /previous/i })).toBeInTheDocument();
      });
    });

    it('handles page changes', async () => {
      render(<TradeHistory tokenAddress={mockToken} />);

      await waitFor(() => {
        const nextButton = screen.getByRole('button', { name: /next/i });
        fireEvent.click(nextButton);
      });

      // Verify that the page changed
      await waitFor(() => {
        expect(screen.getByText(/page 2/i)).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('displays error message on fetch failure', async () => {
      server.use(
        rest.get('*/api/trades/*', (req, res, ctx) => {
          return res(ctx.status(500));
        })
      );

      render(<TradeHistory tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/failed to load trades/i)).toBeInTheDocument();
      });
    });

    it('handles network errors gracefully', async () => {
      server.use(
        rest.get('*/api/trades/*', (req, res) => {
          return res.networkError('Failed to connect');
        })
      );

      render(<TradeHistory tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/network error/i)).toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('displays message when no trades exist', async () => {
      server.use(
        rest.get('*/api/trades/*', (req, res, ctx) => {
          return res(ctx.json([]));
        })
      );

      render(<TradeHistory tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/no trades found/i)).toBeInTheDocument();
      });
    });
  });

  describe('Filtering and Sorting', () => {
    it('allows filtering by trade type', async () => {
      render(<TradeHistory tokenAddress={mockToken} />);

      await waitFor(() => {
        const filterSelect = screen.getByLabelText(/filter by type/i);
        fireEvent.change(filterSelect, { target: { value: 'buy' } });
      });

      await waitFor(() => {
        expect(screen.getAllByText(/buy/i)).toHaveLength(2);
        expect(screen.queryByText(/sell/i)).not.toBeInTheDocument();
      });
    });

    it('allows sorting by timestamp', async () => {
      render(<TradeHistory tokenAddress={mockToken} />);

      await waitFor(() => {
        const sortButton = screen.getByRole('button', { name: /sort by time/i });
        fireEvent.click(sortButton);
      });

      await waitFor(() => {
        const timestamps = screen.getAllByText(/\d{2}:\d{2}/);
        expect(timestamps[0]).toHaveTextContent('11:30');
        expect(timestamps[timestamps.length - 1]).toHaveTextContent('10:30');
      });
    });
  });

  describe('Real-time Updates', () => {
    it('updates trade list when new trade arrives', async () => {
      render(<TradeHistory tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getAllByRole('row')).toHaveLength(4); // 3 trades + header row
      });

      // Simulate new trade arrival
      const newTrade = {
        id: '4',
        type: 'sell',
        amount: 25,
        price: 1.9,
        timestamp: '2024-03-15T12:00:00Z',
        profit: 7.5,
        status: 'completed'
      };

      server.use(
        rest.get('*/api/trades/*', (req, res, ctx) => {
          return res(ctx.json([newTrade, ...mockTrades]));
        })
      );

      // Trigger refresh
      fireEvent.click(screen.getByRole('button', { name: /refresh/i }));

      await waitFor(() => {
        expect(screen.getAllByRole('row')).toHaveLength(5); // 4 trades + header row
        expect(screen.getByText('25')).toBeInTheDocument();
      });
    });
  });
}); 