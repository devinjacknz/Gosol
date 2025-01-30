'use client';

import { useState } from 'react';
import {
  Box,
  Button,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Alert,
  Paper,
  Typography,
  InputAdornment,
} from '@mui/material';
import { useWebSocketStore } from '@/lib/websocket';

interface TradeFormProps {
  onOrderSubmit: (order: OrderData) => Promise<void>;
}

interface OrderData {
  symbol: string;
  side: 'BUY' | 'SELL';
  quantity: number;
  price: number;
}

export default function TradeForm({ onOrderSubmit }: TradeFormProps) {
  const { marketData } = useWebSocketStore();
  const [order, setOrder] = useState<OrderData>({
    symbol: '',
    side: 'BUY',
    quantity: 0,
    price: 0,
  });
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState<string>('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    try {
      await onOrderSubmit(order);
      setSuccess('Order placed successfully!');
      // Reset form
      setOrder({
        ...order,
        quantity: 0,
        price: 0,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to place order');
    }
  };

  const handleChange = (field: keyof OrderData) => (
    e: React.ChangeEvent<HTMLInputElement | { value: unknown }>
  ) => {
    const value = e.target.value;
    setOrder((prev) => ({
      ...prev,
      [field]: field === 'side' ? value : Number(value),
    }));
  };

  const symbols = Object.keys(marketData);
  const currentPrice = order.symbol ? marketData[order.symbol]?.price : 0;

  return (
    <Paper elevation={3} sx={{ p: 3, mb: 3 }}>
      <Typography variant="h6" gutterBottom>
        Place Order
      </Typography>
      <Box component="form" onSubmit={handleSubmit} sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
        <FormControl fullWidth>
          <InputLabel>Symbol</InputLabel>
          <Select
            value={order.symbol}
            label="Symbol"
            onChange={(e) => {
              setOrder((prev) => ({
                ...prev,
                symbol: e.target.value as string,
                price: marketData[e.target.value as string]?.price || 0,
              }));
            }}
          >
            {symbols.map((symbol) => (
              <MenuItem key={symbol} value={symbol}>
                {symbol} - ${marketData[symbol].price.toFixed(2)}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        <FormControl fullWidth>
          <InputLabel>Side</InputLabel>
          <Select
            value={order.side}
            label="Side"
            onChange={(e) => setOrder((prev) => ({ ...prev, side: e.target.value as 'BUY' | 'SELL' }))}
          >
            <MenuItem value="BUY">Buy</MenuItem>
            <MenuItem value="SELL">Sell</MenuItem>
          </Select>
        </FormControl>

        <TextField
          label="Quantity"
          type="number"
          value={order.quantity || ''}
          onChange={handleChange('quantity')}
          InputProps={{
            inputProps: { min: 0, step: 0.01 },
          }}
        />

        <TextField
          label="Price"
          type="number"
          value={order.price || ''}
          onChange={handleChange('price')}
          InputProps={{
            startAdornment: <InputAdornment position="start">$</InputAdornment>,
            inputProps: { min: 0, step: 0.01 },
          }}
          helperText={currentPrice ? `Current market price: $${currentPrice.toFixed(2)}` : ''}
        />

        {error && <Alert severity="error">{error}</Alert>}
        {success && <Alert severity="success">{success}</Alert>}

        <Button
          type="submit"
          variant="contained"
          color={order.side === 'BUY' ? 'success' : 'error'}
          disabled={!order.symbol || !order.quantity || !order.price}
        >
          Place {order.side.toLowerCase()} Order
        </Button>
      </Box>
    </Paper>
  );
} 