import React, { useEffect, useState } from 'react';
import {
  Box,
  Grid,
  Paper,
  Typography,
  CircularProgress,
  Alert,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
} from '@mui/material';
import { fetchDashboardData } from '@/utils/api';
import { formatPrice } from '@/utils/trading';

interface DashboardData {
  portfolio?: {
    totalValue: number;
    dailyPnL: number;
    totalPnL: number;
    positions: Array<{
      symbol: string;
      amount: number;
      value: number;
      pnl: number;
    }>;
  };
  trades?: Array<{
    id: string;
    symbol: string;
    side: 'buy' | 'sell';
    price: number;
    amount: number;
    timestamp: string;
  }>;
  performance?: {
    winRate: number;
    profitFactor: number;
    sharpeRatio: number;
    maxDrawdown: number;
  };
}

export default function Dashboard() {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [data, setData] = useState<DashboardData | null>(null);
  const [timeRange, setTimeRange] = useState('24h');

  const fetchData = async () => {
    try {
      setLoading(true);
      const response = await fetchDashboardData({ timeRange });
      setData(response.data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 30000); // 每30秒更新一次
    return () => clearInterval(interval);
  }, [timeRange]);

  if (loading && !data) {
    return (
      <Box display="flex" justifyContent="center" p={4}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">{error}</Alert>;
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ mb: 3, display: 'flex', justifyContent: 'space-between' }}>
        <Typography variant="h4">Dashboard</Typography>
        <FormControl sx={{ minWidth: 120 }}>
          <InputLabel>Time Range</InputLabel>
          <Select
            value={timeRange}
            label="Time Range"
            onChange={(e) => setTimeRange(e.target.value)}
          >
            <MenuItem value="24h">24 Hours</MenuItem>
            <MenuItem value="7d">7 Days</MenuItem>
            <MenuItem value="30d">30 Days</MenuItem>
          </Select>
        </FormControl>
      </Box>

      <Grid container spacing={3}>
        {/* Portfolio Summary */}
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 2 }}>
            <Typography variant="h6" gutterBottom>
              Portfolio Summary
            </Typography>
            {data?.portfolio && (
              <>
                <Typography variant="h4">
                  {formatPrice(data.portfolio.totalValue, 'USD')}
                </Typography>
                <Typography
                  color={data.portfolio.dailyPnL >= 0 ? 'success.main' : 'error.main'}
                >
                  {data.portfolio.dailyPnL >= 0 ? '+' : ''}
                  {formatPrice(data.portfolio.dailyPnL, 'USD')}
                </Typography>
              </>
            )}
          </Paper>
        </Grid>

        {/* Performance Metrics */}
        <Grid item xs={12} md={8}>
          <Paper sx={{ p: 2 }}>
            <Typography variant="h6" gutterBottom>
              Performance Metrics
            </Typography>
            {data?.performance && (
              <Grid container spacing={2}>
                <Grid item xs={3}>
                  <Typography variant="body2" color="text.secondary">
                    Win Rate
                  </Typography>
                  <Typography variant="h6">
                    {(data.performance.winRate * 100).toFixed(1)}%
                  </Typography>
                </Grid>
                <Grid item xs={3}>
                  <Typography variant="body2" color="text.secondary">
                    Profit Factor
                  </Typography>
                  <Typography variant="h6">
                    {data.performance.profitFactor.toFixed(1)}
                  </Typography>
                </Grid>
                <Grid item xs={3}>
                  <Typography variant="body2" color="text.secondary">
                    Sharpe Ratio
                  </Typography>
                  <Typography variant="h6">
                    {data.performance.sharpeRatio.toFixed(1)}
                  </Typography>
                </Grid>
                <Grid item xs={3}>
                  <Typography variant="body2" color="text.secondary">
                    Max Drawdown
                  </Typography>
                  <Typography variant="h6" color="error.main">
                    {(data.performance.maxDrawdown * 100).toFixed(1)}%
                  </Typography>
                </Grid>
              </Grid>
            )}
          </Paper>
        </Grid>

        {/* Positions */}
        <Grid item xs={12}>
          <Paper sx={{ p: 2 }}>
            <Typography variant="h6" gutterBottom>
              Active Positions
            </Typography>
            <Grid container spacing={2}>
              {data?.portfolio?.positions.map((position) => (
                <Grid item xs={12} sm={6} md={4} key={position.symbol}>
                  <Paper
                    variant="outlined"
                    sx={{ p: 2, backgroundColor: 'background.default' }}
                  >
                    <Typography variant="subtitle1">{position.symbol}</Typography>
                    <Typography variant="h6">
                      {formatPrice(position.value, 'USD')}
                    </Typography>
                    <Typography
                      color={position.pnl >= 0 ? 'success.main' : 'error.main'}
                    >
                      {position.pnl >= 0 ? '+' : ''}
                      {formatPrice(position.pnl, 'USD')}
                    </Typography>
                  </Paper>
                </Grid>
              ))}
            </Grid>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
} 