import React, { Component, ErrorInfo, createContext, useContext, useState } from 'react';
import { Box, Typography, Button, Alert } from '@mui/material';

interface ErrorContextType {
  error: string | null;
  setError: (error: string | null) => void;
}

const ErrorContext = createContext<ErrorContextType | undefined>(undefined);

export const useError = () => {
  const context = useContext(ErrorContext);
  if (!context) {
    throw new Error('useError must be used within an ErrorProvider');
  }
  return context;
};

export const ErrorProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [error, setError] = useState<string | null>(null);

  return (
    <ErrorContext.Provider value={{ error, setError }}>
      {children}
      {error && (
        <Box
          sx={{
            position: 'fixed',
            bottom: '20px',
            right: '20px',
            maxWidth: '400px',
            zIndex: 9999
          }}
        >
          <Alert
            severity="error"
            onClose={() => setError(null)}
          >
            <Typography variant="body1">
              {error}
            </Typography>
          </Alert>
        </Box>
      )}
    </ErrorContext.Provider>
  );
};

interface Props {
  onRetry?: () => void;
  children: React.ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null
    };
  }

  static getDerivedStateFromError(error: Error): State {
    return {
      hasError: true,
      error
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error caught by error boundary:', error, errorInfo);
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null });
    if (this.props.onRetry) {
      this.props.onRetry();
    }
  };

  render() {
    if (this.state.hasError) {
      return (
        <Box
          sx={{
            position: 'fixed',
            bottom: '20px',
            right: '20px',
            maxWidth: '400px',
            zIndex: 9999
          }}
        >
          <Alert
            severity="error"
            action={
              this.props.onRetry && (
                <Button color="inherit" size="small" onClick={this.handleRetry}>
                  Retry
                </Button>
              )
            }
          >
            <Typography variant="body1">
              {this.state.error?.message || 'An error occurred'}
            </Typography>
          </Alert>
        </Box>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
