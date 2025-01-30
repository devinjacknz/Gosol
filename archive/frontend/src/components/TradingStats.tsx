import React from 'react';
import { Card, CardContent, Typography, Grid, Box, CircularProgress } from '@mui/material';

export interface TradingState {
  totalTrades: number;
  successfulTrades: number;
  totalProfit: number;
  lastTradeTime: string;
  winRate: number;
  averageProfit: number;
  averageLoss: number;
  profitFactor: number;
  maxDrawdown: number;
  sharpeRatio: number;
}

interface TradingStatsProps {
  state: TradingState;
  balance: number;
  isLoading?: boolean;
  error?: string;
}

export const TradingStats: React.FC<TradingStatsProps> = ({ state, balance, isLoading, error }) => {
  if (isLoading) {
    return (
      <Box data-testid="loading-skeleton">
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Typography color="error">
        {error}
      </Typography>
    );
  }

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    }).format(value);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          Trading Statistics
        </Typography>
        <Grid container spacing={2}>
          <Grid item xs={12} sm={6}>
            <Typography variant="subtitle2">Wallet Balance</Typography>
            <Typography variant="h6">${formatCurrency(balance)}</Typography>
          </Grid>
          <Grid item xs={12} sm={6}>
            <Typography variant="subtitle2">Total Profit</Typography>
            <Typography variant="h6" color={state.totalProfit >= 0 ? 'success.main' : 'error.main'}>
              ${formatCurrency(state.totalProfit)}
            </Typography>
          </Grid>
          <Grid item xs={12} sm={6}>
            <Typography variant="subtitle2">Success Rate</Typography>
            <Typography variant="h6">
              {((state.successfulTrades / state.totalTrades) * 100 || 0).toFixed(1)}%
            </Typography>
          </Grid>
          <Grid item xs={12} sm={6}>
            <Typography variant="subtitle2">Total Trades</Typography>
            <Typography variant="h6">{state.totalTrades}</Typography>
          </Grid>
          <Grid item xs={12} sm={6}>
            <Typography variant="subtitle2">Successful Trades</Typography>
            <Typography variant="h6">{state.successfulTrades}</Typography>
          </Grid>
          <Grid item xs={12} sm={6}>
            <Typography variant="subtitle2">Last Trade</Typography>
            <Typography variant="h6">{state.lastTradeTime ? formatDate(state.lastTradeTime) : 'No trades yet'}</Typography>
          </Grid>
        </Grid>
      </CardContent>
    </Card>
  );
};

export default TradingStats;
