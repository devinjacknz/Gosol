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
import { useTradingStore } from '@/store';
import type { MarketData } from '@/types/trading';

interface TradingViewProps {
  symbol: string;
  onPlaceOrder?: (order: {
    symbol: string;
    side: 'buy' | 'sell';
    type: 'limit';
    price: number;
    amount: number;
  }) => void;
}

type OrderBookEntry = [number, number];

interface MarketTrade {
  id: string;
  price: number;
  amount: number;
  side: 'buy' | 'sell';
  timestamp: string;
}

export default function TradingView({ symbol, onPlaceOrder }: TradingViewProps) {
  const [price, setPrice] = useState('');
  const [amount, setAmount] = useState('');
  const [errors, setErrors] = useState<string[]>([]);

  useWebSocket(process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws');
  const { marketData, systemStatus } = useTradingStore();
  const currentMarket = marketData[symbol];
  const isConnected = systemStatus?.isConnected || false;

  const handleSubmit = (side: 'buy' | 'sell') => {
    if (!currentMarket) return;

    const order = {
      symbol,
      side,
      type: 'limit' as const,
      price: parseFloat(price || currentMarket.price.toString()),
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

  if (!currentMarket) {
    return <Alert severity="warning">Loading market data...</Alert>;
  }

  return (
    <Grid container spacing={2}>
      <Grid item xs={12} md={8}>
        <Paper sx={{ p: 2 }}>
          <Typography variant="h6" gutterBottom>
            {symbol} Market
          </Typography>
          {currentMarket && (
            <Box>
              <Typography variant="h4">{formatPrice(currentMarket.price)}</Typography>
              <Typography
                color={currentMarket.change24h >= 0 ? 'success.main' : 'error.main'}
              >
                {currentMarket.change24h >= 0 ? '+' : ''}
                {currentMarket.change24h}%
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
                {currentMarket?.orderBook?.asks?.map(([price, amount]) => (
                  <TableRow key={price}>
                    <TableCell sx={{ color: 'error.main' }}>
                      {formatPrice(price)}
                    </TableCell>
                    <TableCell align="right">{formatAmount(amount)}</TableCell>
                  </TableRow>
                ))}
                {currentMarket?.orderBook?.bids?.map(([price, amount]) => (
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
                {currentMarket?.trades?.map((trade) => (
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