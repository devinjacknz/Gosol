import React, { useState } from 'react';
import { Container, Box, Typography, TextField, Button, Paper, Grid } from '@mui/material';
import MarketAnalysis from './components/MarketAnalysis';
import { TradingControls } from './components/TradingControls';
import { TradingStats } from './components/TradingStats';
import TradeHistory from './components/TradeHistory';
import RiskSettings from './components/RiskSettings';
import ErrorBoundary, { ErrorProvider } from './components/ErrorBoundary';
import wsManager from './services/websocket';

const AppContent: React.FC = () => {
  const [tokenAddress, setTokenAddress] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isTrading, setIsTrading] = useState(false);
  const [config, setConfig] = useState({
    maxAmount: 1000,
    stopLoss: 5,
    takeProfit: 10,
    walletAddress: ''
  });
  const [transferSuccess, setTransferSuccess] = useState(false);
  const [walletBalance, setWalletBalance] = useState(0);
  const [loadingStatus, setLoadingStatus] = useState(false);
  const [statusError, setStatusError] = useState<string | null>(null);

  // Add effect to fetch trading status
  React.useEffect(() => {
    if (tokenAddress) {
      setLoadingStatus(true);
      fetch('http://localhost:8080/api/status')
        .then(res => {
          if (!res.ok) {
            throw new Error('Failed to load trading status');
          }
          return res.json();
        })
        .then(data => {
          setIsTrading(data.is_trading);
          setLoadingStatus(false);
        })
        .catch(err => {
          setStatusError(err.message);
          setLoadingStatus(false);
        });
    }
  }, [tokenAddress]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setStatusError(null);
    
    try {
      const response = await fetch(`http://localhost:8080/api/token/${tokenAddress}`);
      if (!response.ok) {
        throw new Error('Network request failed');
      }
      const data = await response.json();
      // Handle successful response
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Network request failed');
    }
  };

  const handleConfigChange = (newConfig: typeof config) => {
    setConfig(newConfig);
  };

  const handleToggleTrading = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/trading/toggle', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled: !isTrading })
      });
      if (!response.ok) {
        throw new Error('Network request failed');
      }
      setIsTrading(!isTrading);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Network request failed');
    }
  };

  const handleTransferProfit = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/trading/transfer', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ action: 'transfer' })
      });
      if (!response.ok) {
        throw new Error('Network request failed');
      }
      setTransferSuccess(true);
      setTimeout(() => setTransferSuccess(false), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Network request failed');
    }
  };

  // Cleanup WebSocket connection on unmount
  React.useEffect(() => {
    return () => {
      wsManager.disconnect();
    };
  }, []);

  // Add effect to fetch wallet balance
  React.useEffect(() => {
    if (tokenAddress) {
      fetch('http://localhost:8080/api/wallet/balance')
        .then(res => res.json())
        .then(data => setWalletBalance(data.balance))
        .catch(err => setError(err.message));
    }
  }, [tokenAddress]);

  return (
    <Container maxWidth="lg">
      <Box sx={{ py: 4 }}>
        <Box sx={{ mb: 4, textAlign: 'center' }}>
          <Typography variant="h4" component="h1" gutterBottom>
            Solmeme Trader
          </Typography>
          <Paper component="form" onSubmit={handleSubmit} sx={{ p: 2, maxWidth: 600, mx: 'auto' }}>
            <Box sx={{ display: 'flex', gap: 2 }}>
              <TextField
                fullWidth
                value={tokenAddress}
                onChange={(e) => setTokenAddress(e.target.value)}
                placeholder="Enter token address"
                error={!!error}
                helperText={error}
              />
              <Button type="submit" variant="contained" color="primary">
                Load Token
              </Button>
            </Box>
          </Paper>
        </Box>

        {statusError && (
          <Box sx={{ mb: 2 }}>
            <Typography color="error">
              {statusError}
            </Typography>
          </Box>
        )}

        {error ? (
          <div className="error">Error: {error}</div>
        ) : !tokenAddress ? (
          <Box sx={{ textAlign: 'center', py: 8 }}>
            <Typography variant="h5" gutterBottom>
              Welcome to Solmeme Trader
            </Typography>
            <Typography variant="body1" color="text.secondary">
              Enter a Solana token address above to start trading
            </Typography>
          </Box>
        ) : (
          <Grid container spacing={3}>
            <Grid item xs={12}>
              <Box sx={{ mb: 2 }}>
                <Typography variant="h6">
                  Wallet Balance: {walletBalance} SOL
                </Typography>
                {transferSuccess && (
                  <Typography color="success.main">
                    Transfer successful
                  </Typography>
                )}
              </Box>
            </Grid>
            <Grid item xs={12} md={8}>
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
                <MarketAnalysis tokenAddress={tokenAddress} />
                <TradingControls
                  config={config}
                  isTrading={isTrading}
                  onConfigChange={handleConfigChange}
                  onToggleTrading={handleToggleTrading}
                  onTransferProfit={handleTransferProfit}
                />
                <RiskSettings
                  tokenAddress={tokenAddress}
                  onUpdate={(config) => console.log('Risk settings updated:', config)}
                />
              </Box>
            </Grid>
            <Grid item xs={12} md={4}>
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
                <TradingStats 
                  state={{
                    totalTrades: 0,
                    successfulTrades: 0,
                    totalProfit: 0,
                    lastTradeTime: '',
                    winRate: 0,
                    averageProfit: 0,
                    averageLoss: 0,
                    profitFactor: 0,
                    maxDrawdown: 0,
                    sharpeRatio: 0
                  }}
                  balance={walletBalance}
                />
                <TradeHistory tokenAddress={tokenAddress} />
              </Box>
            </Grid>
          </Grid>
        )}
      </Box>
    </Container>
  );
};

const App: React.FC = () => {
  return (
    <ErrorProvider>
      <ErrorBoundary>
        <AppContent />
      </ErrorBoundary>
    </ErrorProvider>
  );
};

export default App;
