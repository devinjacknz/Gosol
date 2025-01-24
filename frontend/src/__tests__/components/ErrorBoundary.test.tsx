import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { ErrorBoundary, ErrorProvider, useError } from '../../components/ErrorBoundary';

const ThrowError = () => {
  throw new Error('Test error');
};

const TestComponent = () => {
  const { setError } = useError();
  return (
    <button onClick={() => setError('Test error message')}>
      Set Error
    </button>
  );
};

describe('ErrorBoundary', () => {
  beforeEach(() => {
    jest.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  it('renders children when there is no error', () => {
    render(
      <ErrorBoundary>
        <div>Test content</div>
      </ErrorBoundary>
    );

    expect(screen.getByText('Test content')).toBeInTheDocument();
  });

  it('renders error message when an error occurs', () => {
    render(
      <ErrorBoundary>
        <ThrowError />
      </ErrorBoundary>
    );

    expect(screen.getByText('Test error')).toBeInTheDocument();
  });

  it('provides retry functionality', () => {
    const onRetry = jest.fn();
    render(
      <ErrorBoundary onRetry={onRetry}>
        <ThrowError />
      </ErrorBoundary>
    );

    const retryButton = screen.getByRole('button', { name: /retry/i });
    fireEvent.click(retryButton);

    expect(onRetry).toHaveBeenCalled();
  });

  it('resets error state after retry', () => {
    let shouldThrow = true;
    const TestComponent = () => {
      if (shouldThrow) {
        throw new Error('Test error');
      }
      return <div>Success</div>;
    };

    render(
      <ErrorBoundary onRetry={() => { shouldThrow = false; }}>
        <TestComponent />
      </ErrorBoundary>
    );

    const retryButton = screen.getByRole('button', { name: /retry/i });
    fireEvent.click(retryButton);

    expect(screen.getByText('Success')).toBeInTheDocument();
  });
});

describe('ErrorProvider', () => {
  it('provides error context to children', () => {
    render(
      <ErrorProvider>
        <TestComponent />
      </ErrorProvider>
    );

    const button = screen.getByRole('button', { name: /set error/i });
    fireEvent.click(button);

    expect(screen.getByText('Test error message')).toBeInTheDocument();
  });

  it('allows clearing error message', () => {
    render(
      <ErrorProvider>
        <TestComponent />
      </ErrorProvider>
    );

    const button = screen.getByRole('button', { name: /set error/i });
    fireEvent.click(button);

    const closeButton = screen.getByRole('button', { name: /close/i });
    fireEvent.click(closeButton);

    expect(screen.queryByText('Test error message')).not.toBeInTheDocument();
  });

  it('throws error when useError is used outside provider', () => {
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {});
    
    expect(() => {
      render(<TestComponent />);
    }).toThrow('useError must be used within an ErrorProvider');
    
    consoleSpy.mockRestore();
  });

  it('handles multiple error updates', () => {
    render(
      <ErrorProvider>
        <TestComponent />
      </ErrorProvider>
    );

    const button = screen.getByRole('button', { name: /set error/i });
    
    // Set first error
    fireEvent.click(button);
    expect(screen.getByText('Test error message')).toBeInTheDocument();

    // Clear error
    const closeButton = screen.getByRole('button', { name: /close/i });
    fireEvent.click(closeButton);
    expect(screen.queryByText('Test error message')).not.toBeInTheDocument();

    // Set second error
    fireEvent.click(button);
    expect(screen.getByText('Test error message')).toBeInTheDocument();
  });
}); 