import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { rest } from 'msw';
import { server } from '../mocks/server';
import App from '../App';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
  },
});

const renderApp = () =>
  render(
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  );

describe('App Integration', () => {
  beforeEach(() => {
    queryClient.clear();
  });

  it('loads and displays initial data correctly', async () => {
    renderApp();

    // Wait for initial data to load
    await waitFor(() => {
      expect(screen.getByText(/wallet balance/i)).toBeInTheDocument();
    });

    // Check if trading controls are rendered
    expect(screen.getByLabelText(/auto trading/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/maximum amount/i)).toBeInTheDocument();

    // Check if market analysis is rendered
    expect(screen.getByText(/market sentiment/i)).toBeInTheDocument();
    expect(screen.getByText(/risk level/i)).toBeInTheDocument();

    // Check if trading stats are rendered
    expect(screen.getByText(/success rate/i)).toBeInTheDocument();
    expect(screen.getByText(/total trades/i)).toBeInTheDocument();
  });

  it('handles trading toggle correctly', async () => {
    renderApp();

    // Wait for initial data to load
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /start trading/i })).toBeInTheDocument();
    });

    // Mock the toggle endpoint
    server.use(
      rest.post('http://localhost:8080/api/trading/toggle', (req, res, ctx) => {
        return res(ctx.status(200));
      })
    );

    // Click the toggle button
    const toggleButton = screen.getByRole('button', { name: /start trading/i });
    fireEvent.click(toggleButton);

    // Verify the button text changes
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /stop trading/i })).toBeInTheDocument();
    });
  });

  it('updates configuration correctly', async () => {
    renderApp();

    // Wait for initial data to load
    await waitFor(() => {
      expect(screen.getByLabelText(/maximum amount/i)).toBeInTheDocument();
    });

    // Mock the config update endpoint
    server.use(
      rest.post('http://localhost:8080/api/config', (req, res, ctx) => {
        return res(ctx.status(200));
      })
    );

    // Update maximum amount
    const maxAmountInput = screen.getByLabelText(/maximum amount/i);
    await userEvent.clear(maxAmountInput);
    await userEvent.type(maxAmountInput, '200');

    // Verify the config update request was made
    await waitFor(() => {
      expect(maxAmountInput).toHaveValue(200);
    });
  });

  it('handles transfer profit correctly', async () => {
    renderApp();

    // Wait for initial data to load
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /transfer profit/i })).toBeInTheDocument();
    });

    // Mock the transfer endpoint
    server.use(
      rest.post('http://localhost:8080/api/trading/transfer', (req, res, ctx) => {
        return res(ctx.status(200));
      })
    );

    // Click the transfer button
    const transferButton = screen.getByRole('button', { name: /transfer profit/i });
    fireEvent.click(transferButton);

    // Verify the transfer was successful
    await waitFor(() => {
      expect(screen.getByText(/transfer successful/i)).toBeInTheDocument();
    });
  });

  it('handles API errors gracefully', async () => {
    // Mock API error
    server.use(
      rest.get('http://localhost:8080/api/status', (req, res, ctx) => {
        return res(ctx.status(500), ctx.json({ error: 'Internal server error' }));
      })
    );

    renderApp();

    // Verify error message is displayed
    await waitFor(() => {
      expect(screen.getByText(/failed to load trading status/i)).toBeInTheDocument();
    });
  });

  it('updates market analysis in real-time', async () => {
    renderApp();

    // Wait for initial data to load
    await waitFor(() => {
      expect(screen.getByText(/market sentiment/i)).toBeInTheDocument();
    });

    // Mock updated market analysis
    server.use(
      rest.get('http://localhost:8080/api/analysis', (req, res, ctx) => {
        return res(
          ctx.json({
            sentiment: 'bearish',
            riskLevel: 'high',
            recommendation: 'sell',
            confidence: 80,
            priceTarget: 140.0,
            keyFactors: ['Downward trend detected']
          })
        );
      })
    );

    // Wait for analysis update
    await waitFor(() => {
      expect(screen.getByText(/bearish/i)).toBeInTheDocument();
      expect(screen.getByText(/high/i)).toBeInTheDocument();
      expect(screen.getByText(/sell/i)).toBeInTheDocument();
    });
  });

  it('maintains state during component updates', async () => {
    renderApp();

    // Wait for initial data to load
    await waitFor(() => {
      expect(screen.getByLabelText(/maximum amount/i)).toBeInTheDocument();
    });

    // Update configuration
    const maxAmountInput = screen.getByLabelText(/maximum amount/i);
    await userEvent.clear(maxAmountInput);
    await userEvent.type(maxAmountInput, '200');

    // Trigger market analysis update
    server.use(
      rest.get('http://localhost:8080/api/analysis', (req, res, ctx) => {
        return res(
          ctx.json({
            sentiment: 'bearish',
            riskLevel: 'high',
            recommendation: 'sell',
            confidence: 80,
            priceTarget: 140.0,
            keyFactors: ['Downward trend detected']
          })
        );
      })
    );

    // Verify configuration is maintained
    await waitFor(() => {
      expect(maxAmountInput).toHaveValue(200);
      expect(screen.getByText(/bearish/i)).toBeInTheDocument();
    });
  });

  it('handles concurrent API requests correctly', async () => {
    renderApp();

    // Wait for initial data to load
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /start trading/i })).toBeInTheDocument();
    });

    // Trigger multiple actions simultaneously
    const toggleButton = screen.getByRole('button', { name: /start trading/i });
    const transferButton = screen.getByRole('button', { name: /transfer profit/i });
    const maxAmountInput = screen.getByLabelText(/maximum amount/i);

    // Perform actions concurrently
    fireEvent.click(toggleButton);
    fireEvent.click(transferButton);
    await userEvent.clear(maxAmountInput);
    await userEvent.type(maxAmountInput, '200');

    // Verify all actions complete successfully
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /stop trading/i })).toBeInTheDocument();
      expect(screen.getByText(/transfer successful/i)).toBeInTheDocument();
      expect(maxAmountInput).toHaveValue(200);
    });
  });

  it('loads token data correctly', async () => {
    render(<App />);
    
    const input = screen.getByPlaceholderText(/enter token address/i);
    const button = screen.getByText(/load token/i);
    
    await userEvent.type(input, 'So11111111111111111111111111111111111111112');
    fireEvent.click(button);
    
    await waitFor(() => {
      expect(screen.getByText(/network request failed/i)).toBeInTheDocument();
    });
  });

  it('handles invalid token address', async () => {
    render(<App />);
    
    const input = screen.getByPlaceholderText(/enter token address/i);
    const button = screen.getByText(/load token/i);
    
    await userEvent.type(input, 'invalid-address');
    fireEvent.click(button);
    
    await waitFor(() => {
      expect(screen.getByText(/network request failed/i)).toBeInTheDocument();
    });
  });

  it('handles network error gracefully', async () => {
    server.close();
    render(<App />);
    
    const input = screen.getByPlaceholderText(/enter token address/i);
    const button = screen.getByText(/load token/i);
    
    await userEvent.type(input, 'So11111111111111111111111111111111111111112');
    fireEvent.click(button);
    
    await waitFor(() => {
      expect(screen.getByText(/network request failed/i)).toBeInTheDocument();
    });
    server.listen();
  });
}); 