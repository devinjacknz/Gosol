'use client';

import { useState } from 'react';
import {
  Box,
  Tabs,
  Tab,
  TextField,
  Button,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Typography,
  Alert,
  Grid,
  InputAdornment,
  Slider,
} from '@mui/material';
import { useTrading } from '@/contexts/TradingContext';
import { ORDER_TYPES, ORDER_SIDES } from '@/config/trading';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`trade-tabpanel-${index}`}
      aria-labelledby={`trade-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 2 }}>{children}</Box>}
    </div>
  );
}

export default function TradeForm() {
  const { state, placeOrder } = useTrading();
  const [activeTab, setActiveTab] = useState(0);
  const [orderType, setOrderType] = useState(ORDER_TYPES.LIMIT);
  const [side, setSide] = useState(ORDER_SIDES.BUY);
  const [price, setPrice] = useState('');
  const [amount, setAmount] = useState('');
  const [total, setTotal] = useState('');
  const [stopPrice, setStopPrice] = useState('');
  const [error, setError] = useState('');
  const [sliderValue, setSliderValue] = useState(0);

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setActiveTab(newValue);
    setSide(newValue === 0 ? ORDER_SIDES.BUY : ORDER_SIDES.SELL);
  };

  const handlePriceChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const newPrice = event.target.value;
    setPrice(newPrice);
    if (amount && newPrice) {
      setTotal((parseFloat(amount) * parseFloat(newPrice)).toString());
    }
  };

  const handleAmountChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const newAmount = event.target.value;
    setAmount(newAmount);
    if (price && newAmount) {
      setTotal((parseFloat(newAmount) * parseFloat(price)).toString());
    }
  };

  const handleTotalChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const newTotal = event.target.value;
    setTotal(newTotal);
    if (price && newTotal) {
      setAmount((parseFloat(newTotal) / parseFloat(price)).toString());
    }
  };

  const handleSliderChange = (event: Event, newValue: number | number[]) => {
    const value = newValue as number;
    setSliderValue(value);
    
    // Calculate amount based on available balance and slider percentage
    if (state.selectedPair) {
      const balance = side === ORDER_SIDES.BUY
        ? state.balances[state.selectedPair.quoteToken]?.balance || '0'
        : state.balances[state.selectedPair.baseToken]?.balance || '0';
      
      const maxAmount = parseFloat(balance);
      const calculatedAmount = (maxAmount * value) / 100;
      
      if (side === ORDER_SIDES.BUY && price) {
        setTotal(calculatedAmount.toString());
        setAmount((calculatedAmount / parseFloat(price)).toString());
      } else {
        setAmount(calculatedAmount.toString());
        if (price) {
          setTotal((calculatedAmount * parseFloat(price)).toString());
        }
      }
    }
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!state.selectedPair) {
      setError('Please select a trading pair');
      return;
    }

    try {
      setError('');
      await placeOrder({
        pair: state.selectedPair,
        type: orderType,
        side,
        price: orderType === ORDER_TYPES.MARKET ? undefined : price,
        amount,
        stopPrice: orderType.includes('stop') ? stopPrice : undefined,
      });

      // Reset form
      setAmount('');
      setTotal('');
      setStopPrice('');
      setSliderValue(0);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to place order');
    }
  };

  const getAvailableBalance = () => {
    if (!state.selectedPair) return '0';
    
    return side === ORDER_SIDES.BUY
      ? state.balances[state.selectedPair.quoteToken]?.balance || '0'
      : state.balances[state.selectedPair.baseToken]?.balance || '0';
  };

  return (
    <Box>
      <Tabs
        value={activeTab}
        onChange={handleTabChange}
        aria-label="trading tabs"
        sx={{ borderBottom: 1, borderColor: 'divider' }}
      >
        <Tab
          label="Buy"
          sx={{
            flex: 1,
            color: 'success.main',
            '&.Mui-selected': { color: 'success.main' },
          }}
        />
        <Tab
          label="Sell"
          sx={{
            flex: 1,
            color: 'error.main',
            '&.Mui-selected': { color: 'error.main' },
          }}
        />
      </Tabs>

      {error && (
        <Alert severity="error" sx={{ mt: 2 }}>
          {error}
        </Alert>
      )}

      <form onSubmit={handleSubmit}>
        <TabPanel value={activeTab} index={0}>
          <Grid container spacing={2}>
            <Grid item xs={12}>
              <FormControl fullWidth size="small">
                <InputLabel>Order Type</InputLabel>
                <Select
                  value={orderType}
                  label="Order Type"
                  onChange={(e) => setOrderType(e.target.value as any)}
                >
                  <MenuItem value={ORDER_TYPES.LIMIT}>Limit</MenuItem>
                  <MenuItem value={ORDER_TYPES.MARKET}>Market</MenuItem>
                  <MenuItem value={ORDER_TYPES.STOP_LIMIT}>Stop Limit</MenuItem>
                  <MenuItem value={ORDER_TYPES.STOP_MARKET}>Stop Market</MenuItem>
                  <MenuItem value={ORDER_TYPES.TRAILING_STOP}>
                    Trailing Stop
                  </MenuItem>
                </Select>
              </FormControl>
            </Grid>

            {orderType !== ORDER_TYPES.MARKET && (
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  size="small"
                  label="Price"
                  type="number"
                  value={price}
                  onChange={handlePriceChange}
                  InputProps={{
                    endAdornment: state.selectedPair && (
                      <InputAdornment position="end">
                        {state.selectedPair.quoteToken}
                      </InputAdornment>
                    ),
                  }}
                />
              </Grid>
            )}

            {orderType.includes('stop') && (
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  size="small"
                  label="Stop Price"
                  type="number"
                  value={stopPrice}
                  onChange={(e) => setStopPrice(e.target.value)}
                  InputProps={{
                    endAdornment: state.selectedPair && (
                      <InputAdornment position="end">
                        {state.selectedPair.quoteToken}
                      </InputAdornment>
                    ),
                  }}
                />
              </Grid>
            )}

            <Grid item xs={12}>
              <TextField
                fullWidth
                size="small"
                label="Amount"
                type="number"
                value={amount}
                onChange={handleAmountChange}
                InputProps={{
                  endAdornment: state.selectedPair && (
                    <InputAdornment position="end">
                      {state.selectedPair.baseToken}
                    </InputAdornment>
                  ),
                }}
              />
            </Grid>

            <Grid item xs={12}>
              <TextField
                fullWidth
                size="small"
                label="Total"
                type="number"
                value={total}
                onChange={handleTotalChange}
                InputProps={{
                  endAdornment: state.selectedPair && (
                    <InputAdornment position="end">
                      {state.selectedPair.quoteToken}
                    </InputAdornment>
                  ),
                }}
              />
            </Grid>

            <Grid item xs={12}>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                Available: {getAvailableBalance()}
              </Typography>
              <Slider
                value={sliderValue}
                onChange={handleSliderChange}
                aria-labelledby="amount-slider"
                valueLabelDisplay="auto"
                marks={[
                  { value: 0, label: '0%' },
                  { value: 25, label: '25%' },
                  { value: 50, label: '50%' },
                  { value: 75, label: '75%' },
                  { value: 100, label: '100%' },
                ]}
              />
            </Grid>

            <Grid item xs={12}>
              <Button
                fullWidth
                variant="contained"
                color="success"
                type="submit"
                disabled={state.isLoading || !amount || (orderType !== ORDER_TYPES.MARKET && !price)}
              >
                Buy {state.selectedPair?.baseToken}
              </Button>
            </Grid>
          </Grid>
        </TabPanel>

        <TabPanel value={activeTab} index={1}>
          <Grid container spacing={2}>
            {/* Same form fields as Buy tab, but with Sell-specific styling */}
            {/* ... */}
          </Grid>
        </TabPanel>
      </form>
    </Box>
  );
} 