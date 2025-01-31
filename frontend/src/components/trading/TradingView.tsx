import React, { useState } from 'react';
import {
  Box,
  Grid,
  Paper,
  Typography,
  TextField,
  Button,
  Alert,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from '@mui/material';
import { useWebSocket } from '@/hooks/useWebSocket';
import { formatPrice, formatAmount, validateOrder } from '@/utils/trading';

interface TradingViewProps {
  symbol: string;
  onPlaceOrder?: (order: {
    symbol: string;
    side: 'buy' | 'sell';
    price: number;
    amount: number;
  }) => void;
}

export default function TradingView({ symbol, onPlaceOrder }: TradingViewProps) {
  const [price, setPrice] = useState('');
  const [amount, setAmount] = useState('');
  const [errors, setErrors] = useState<string[]>([]);

  const { data: marketData, isConnected, error: wsError } = useWebSocket(
    `wss://api.exchange.com/ws/market/${symbol}`
  );

  const handleSubmit = (side: 'buy' | 'sell') => {
    const order = {
      symbol,
      side,
      type: 'limit' as const,
      price: parseFloat(price),
      amount: parseFloat(amount),
    };

    const validation = validateOrder(order);
    if (!validation.isValid) {
      setErrors(validation.errors || []);
      return;
    }

    onPlaceOrder?.(order);
    setPrice('');
    setAmount('');
    setErrors([]);
  };

  if (!isConnected) {
    return <Alert severity="error">Connection lost</Alert>;
  }

  if (wsError) {
    return <Alert severity="error">{wsError.message}</Alert>;
  }

  return (
    <Grid container spacing={2}>
      <Grid item xs={12} md={8}>
        <Paper sx={{ p: 2 }}>
          <Typography variant="h6" gutterBottom>
            {symbol} Market
          </Typography>
          {marketData && (
            <Box>
              <Typography variant="h4">{formatPrice(marketData.price)}</Typography>
              <Typography
                color={marketData.change >= 0 ? 'success.main' : 'error.main'}
              >
                {marketData.change >= 0 ? '+' : ''}
                {marketData.change}%
              </Typography>
            </Box>
          )}
        </Paper>
      </Grid>

      <Grid item xs={12} md={4}>
        <Paper sx={{ p: 2 }}>
          <Typography variant="h6" gutterBottom>
            Place Order
          </Typography>
          <Box component="form" sx={{ '& > *': { mb: 2 } }}>
            <TextField
              label="Price"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
              fullWidth
              type="number"
              inputProps={{ step: 'any' }}
            />
            <TextField
              label="Amount"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              fullWidth
              type="number"
              inputProps={{ step: 'any' }}
            />
            {errors.map((error, index) => (
              <Alert key={index} severity="error">
                {error}
              </Alert>
            ))}
            <Box sx={{ display: 'flex', gap: 1 }}>
              <Button
                variant="contained"
                color="success"
                onClick={() => handleSubmit('buy')}
                fullWidth
              >
                Buy
              </Button>
              <Button
                variant="contained"
                color="error"
                onClick={() => handleSubmit('sell')}
                fullWidth
              >
                Sell
              </Button>
            </Box>
          </Box>
        </Paper>
      </Grid>

      <Grid item xs={12} md={6}>
        <Paper sx={{ p: 2 }}>
          <Typography variant="h6" gutterBottom>
            Order Book
          </Typography>
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Price</TableCell>
                  <TableCell align="right">Amount</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {marketData?.orderBook?.asks.map(([price, amount]) => (
                  <TableRow key={price}>
                    <TableCell sx={{ color: 'error.main' }}>
                      {formatPrice(price)}
                    </TableCell>
                    <TableCell align="right">{formatAmount(amount)}</TableCell>
                  </TableRow>
                ))}
                {marketData?.orderBook?.bids.map(([price, amount]) => (
                  <TableRow key={price}>
                    <TableCell sx={{ color: 'success.main' }}>
                      {formatPrice(price)}
                    </TableCell>
                    <TableCell align="right">{formatAmount(amount)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </Paper>
      </Grid>

      <Grid item xs={12} md={6}>
        <Paper sx={{ p: 2 }}>
          <Typography variant="h6" gutterBottom>
            Recent Trades
          </Typography>
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Price</TableCell>
                  <TableCell align="right">Amount</TableCell>
                  <TableCell align="right">Time</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {marketData?.trades?.map((trade) => (
                  <TableRow key={trade.id}>
                    <TableCell
                      sx={{
                        color: trade.side === 'buy' ? 'success.main' : 'error.main',
                      }}
                    >
                      {formatPrice(trade.price)}
                    </TableCell>
                    <TableCell align="right">
                      {formatAmount(trade.amount)}
                    </TableCell>
                    <TableCell align="right">
                      {new Date(trade.timestamp).toLocaleTimeString()}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </Paper>
      </Grid>
    </Grid>
  );
} 