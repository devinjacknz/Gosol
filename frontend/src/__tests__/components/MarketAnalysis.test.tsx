import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { rest } from 'msw';
import { server } from '../../mocks/server';
import MarketAnalysis from '../../components/MarketAnalysis';

const mockToken = 'So11111111111111111111111111111111111111111';

describe('MarketAnalysis', () => {
  const mockMarketData = {
    price: 1.5,
    volume_24h: 1000000,
    market_cap: 10000000,
    liquidity: 500000,
  };

  const mockAnalysis = {
    sentiment: 'bullish',
    risk_level: 'medium',
    confidence: 75,
    deepseek_analysis: JSON.stringify({
      recommendation: {
        action: 'buy',
        confidence: 0.85
      },
      analysis: {
        technical: 'Strong upward trend',
        fundamental: 'Good tokenomics'
      }
    })
  };

  beforeEach(() => {
    server.use(
      rest.get('*/api/market-data/*', (req, res, ctx) => {
        return res(ctx.json(mockMarketData));
      }),
      rest.get('*/api/analysis', (req, res, ctx) => {
        return res(ctx.json(mockAnalysis));
      })
    );
  });

  describe('Initial Render', () => {
    it('shows loading state initially', () => {
      render(<MarketAnalysis tokenAddress={mockToken} />);
      expect(screen.getByText(/loading market data/i)).toBeInTheDocument();
    });

    it('renders market analysis sections', async () => {
      render(<MarketAnalysis tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/market analysis/i)).toBeInTheDocument();
        expect(screen.getByText(/price analysis/i)).toBeInTheDocument();
      });
    });
  });

  describe('Market Data Display', () => {
    it('displays current price correctly', async () => {
      render(<MarketAnalysis tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/1\.5 SOL/)).toBeInTheDocument();
      });
    });

    it('displays volume data correctly', async () => {
      render(<MarketAnalysis tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/1000000/)).toBeInTheDocument();
      });
    });

    it('displays market cap correctly', async () => {
      render(<MarketAnalysis tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/10000000/)).toBeInTheDocument();
      });
    });

    it('displays liquidity correctly', async () => {
      render(<MarketAnalysis tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/500000/)).toBeInTheDocument();
      });
    });
  });

  describe('Analysis Display', () => {
    it('displays market sentiment correctly', async () => {
      render(<MarketAnalysis tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/bullish/i)).toBeInTheDocument();
      });
    });

    it('displays risk level correctly', async () => {
      render(<MarketAnalysis tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/medium/i)).toBeInTheDocument();
      });
    });

    it('displays confidence level correctly', async () => {
      render(<MarketAnalysis tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/75%/)).toBeInTheDocument();
      });
    });

    it('displays AI recommendation correctly', async () => {
      render(<MarketAnalysis tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/buy/i)).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('handles market data fetch error', async () => {
      server.use(
        rest.get('*/api/market-data/*', (req, res, ctx) => {
          return res(ctx.status(500));
        })
      );

      render(<MarketAnalysis tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/failed to fetch market data/i)).toBeInTheDocument();
      });
    });

    it('handles analysis fetch error', async () => {
      server.use(
        rest.get('*/api/analysis', (req, res, ctx) => {
          return res(ctx.status(500));
        })
      );

      render(<MarketAnalysis tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/failed to fetch analysis/i)).toBeInTheDocument();
      });
    });

    it('handles network errors gracefully', async () => {
      server.use(
        rest.get('*/api/market-data/*', (req, res) => {
          return res.networkError('Failed to connect');
        })
      );

      render(<MarketAnalysis tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByText(/network error/i)).toBeInTheDocument();
      });
    });
  });

  describe('Data Updates', () => {
    it('updates data periodically', async () => {
      jest.useFakeTimers();
      
      const updatedMarketData = {
        ...mockMarketData,
        price: 2.0
      };

      render(<MarketAnalysis tokenAddress={mockToken} />);

      // Wait for initial data
      await waitFor(() => {
        expect(screen.getByText(/1\.5 SOL/)).toBeInTheDocument();
      });

      // Update mock data
      server.use(
        rest.get('*/api/market-data/*', (req, res, ctx) => {
          return res(ctx.json(updatedMarketData));
        })
      );

      // Fast forward 30 seconds
      jest.advanceTimersByTime(30000);

      await waitFor(() => {
        expect(screen.getByText(/2\.0 SOL/)).toBeInTheDocument();
      });

      jest.useRealTimers();
    });
  });

  describe('Chart Display', () => {
    it('renders price chart', async () => {
      render(<MarketAnalysis tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByRole('img', { name: /price chart/i })).toBeInTheDocument();
      });
    });

    it('updates chart with new data', async () => {
      const { rerender } = render(<MarketAnalysis tokenAddress={mockToken} />);

      await waitFor(() => {
        expect(screen.getByRole('img', { name: /price chart/i })).toBeInTheDocument();
      });

      // Update with new token address
      rerender(<MarketAnalysis tokenAddress="newToken" />);

      await waitFor(() => {
        expect(screen.getByText(/loading market data/i)).toBeInTheDocument();
      });
    });
  });
}); 