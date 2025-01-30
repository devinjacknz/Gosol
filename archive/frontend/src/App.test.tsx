import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { server } from './mocks/server';
import App from './App';

const mockToken = 'So11111111111111111111111111111111111111111';

describe('App Component', () => {
  beforeEach(() => {
    server.resetHandlers();
  });

  describe('Initial Render', () => {
    it('renders app header and welcome message', () => {
      render(<App />);
      expect(screen.getByRole('heading', { level: 1, name: /solmeme trader/i })).toBeInTheDocument();
      expect(screen.getByRole('heading', { level: 5, name: /welcome to solmeme trader/i })).toBeInTheDocument();
      expect(screen.getByText(/enter a solana token address above to start trading/i)).toBeInTheDocument();
    });

    it('renders token input form', () => {
      render(<App />);
      expect(screen.getByPlaceholderText(/enter token address/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /load token/i })).toBeInTheDocument();
    });
  });

  describe('Token Loading', () => {
    it('loads token data successfully', async () => {
      render(<App />);
      const input = screen.getByPlaceholderText(/enter token address/i);
      const button = screen.getByRole('button', { name: /load token/i });

      await userEvent.type(input, mockToken);
      fireEvent.click(button);

      await waitFor(() => {
        expect(screen.getByText(/trading controls/i)).toBeInTheDocument();
        expect(screen.getByText(/wallet balance/i)).toBeInTheDocument();
        expect(screen.getByText(/market analysis/i)).toBeInTheDocument();
      });
    });

    it('handles invalid token address', async () => {
      server.use(
        rest.get('*/api/token/*', (req, res, ctx) => {
          return res(ctx.status(404), ctx.json({ error: 'Token not found' }));
        })
      );

      render(<App />);
      const input = screen.getByPlaceholderText(/enter token address/i);
      const button = screen.getByRole('button', { name: /load token/i });

      await userEvent.type(input, 'invalid-token');
      fireEvent.click(button);

      await waitFor(() => {
        expect(screen.getByText(/token not found/i)).toBeInTheDocument();
      });
    });

    it('handles network error', async () => {
      server.use(
        rest.get('*/api/token/*', (req, res) => {
          return res.networkError('Failed to connect');
        })
      );

      render(<App />);
      const input = screen.getByPlaceholderText(/enter token address/i);
      const button = screen.getByRole('button', { name: /load token/i });

      await userEvent.type(input, mockToken);
      fireEvent.click(button);

      await waitFor(() => {
        expect(screen.getByText(/failed to connect/i)).toBeInTheDocument();
      });
    });
  });

  describe('Trading Controls', () => {
    it('toggles trading successfully', async () => {
      render(<App />);
      await loadToken();

      const toggleButton = screen.getByRole('button', { name: /start trading/i });
      fireEvent.click(toggleButton);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /stop trading/i })).toBeInTheDocument();
      });
    });

    it('handles trading configuration updates', async () => {
      render(<App />);
      await loadToken();

      const maxAmountInput = screen.getByLabelText(/maximum amount/i);
      await userEvent.clear(maxAmountInput);
      await userEvent.type(maxAmountInput, '200');

      await waitFor(() => {
        expect(maxAmountInput).toHaveValue(200);
      });
    });

    it('handles profit transfer', async () => {
      render(<App />);
      await loadToken();

      const transferButton = screen.getByRole('button', { name: /transfer profit/i });
      fireEvent.click(transferButton);

      await waitFor(() => {
        expect(screen.getByText(/profit transferred successfully/i)).toBeInTheDocument();
      });
    });
  });

  describe('Market Analysis', () => {
    it('displays market analysis data', async () => {
      render(<App />);
      await loadToken();

      await waitFor(() => {
        expect(screen.getByText(/market sentiment/i)).toBeInTheDocument();
        expect(screen.getByText(/risk level/i)).toBeInTheDocument();
        expect(screen.getByText(/recommendation/i)).toBeInTheDocument();
      });
    });

    it('updates market data in real-time', async () => {
      render(<App />);
      await loadToken();

      // Simulate WebSocket message
      const mockMessage = {
        type: 'market_data',
        data: {
          price: 1.75,
          volume: 1200000,
          change24h: 6.5
        }
      };

      // TODO: Implement WebSocket message simulation
      
      await waitFor(() => {
        expect(screen.getByText(/\$1\.75/)).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('displays error messages', async () => {
      server.use(
        rest.get('*/api/status', (req, res, ctx) => {
          return res(ctx.status(500), ctx.json({ error: 'Internal server error' }));
        })
      );

      render(<App />);
      await loadToken();

      await waitFor(() => {
        expect(screen.getByText(/internal server error/i)).toBeInTheDocument();
      });
    });

    it('handles connection errors gracefully', async () => {
      server.use(
        rest.get('*/api/status', (req, res) => {
          return res.networkError('Failed to connect');
        })
      );

      render(<App />);
      await loadToken();

      await waitFor(() => {
        expect(screen.getByText(/failed to connect/i)).toBeInTheDocument();
      });
    });
  });
});

// Helper function to load token
async function loadToken() {
  const input = screen.getByPlaceholderText(/enter token address/i);
  const button = screen.getByRole('button', { name: /load token/i });

  await userEvent.type(input, mockToken);
  fireEvent.click(button);

  await waitFor(() => {
    expect(screen.getByText(/trading controls/i)).toBeInTheDocument();
  });
}
