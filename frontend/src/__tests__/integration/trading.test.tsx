import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { server } from '../../mocks/server';
import App from '../../App';

// Increase timeout for all tests
jest.setTimeout(10000);

describe('Trading Integration Tests', () => {
  it('loads token data correctly', async () => {
    render(<App />);
    const input = screen.getByPlaceholderText(/enter token address/i);
    const button = screen.getByText(/load token/i);

    await userEvent.type(input, 'So11111111111111111111111111111111111111111');
    fireEvent.click(button);

    await waitFor(() => {
      expect(screen.getByText(/trading controls/i)).toBeInTheDocument();
      expect(screen.getByText(/test token/i)).toBeInTheDocument();
      expect(screen.getByText(/\$1\.50/)).toBeInTheDocument();
    });
  });

  it('handles invalid token address', async () => {
    render(<App />);
    const input = screen.getByPlaceholderText(/enter token address/i);
    const button = screen.getByText(/load token/i);

    await userEvent.type(input, 'So11111111111111111111111111111111111111112');
    fireEvent.click(button);

    await waitFor(() => {
      expect(screen.getByText(/error/i)).toBeInTheDocument();
    });
  });

  it('handles network error', async () => {
    server.close();
    render(<App />);
    const input = screen.getByPlaceholderText(/enter token address/i);
    const button = screen.getByText(/load token/i);

    await userEvent.type(input, 'So11111111111111111111111111111111111111111');
    fireEvent.click(button);

    await waitFor(() => {
      expect(screen.getByText(/network request failed/i)).toBeInTheDocument();
    });
    server.listen();
  });

  it('updates market analysis in real-time', async () => {
    render(<App />);
    const input = screen.getByPlaceholderText(/enter token address/i);
    const button = screen.getByText(/load token/i);

    await userEvent.type(input, 'So11111111111111111111111111111111111111111');
    fireEvent.click(button);

    await waitFor(() => {
      expect(screen.getByText(/test token/i)).toBeInTheDocument();
    });

    // Simulate WebSocket message
    const mockMessage = {
      type: 'market-update',
      data: {
        price: 1.75,
        volume: 1200000,
        change24h: 6.5
      }
    };
    
    // Wait for the update to be reflected in the UI
    await waitFor(() => {
      expect(screen.getByText(/\$1\.75/)).toBeInTheDocument();
      expect(screen.getByText(/6\.5%/)).toBeInTheDocument();
    });
  });
}); 