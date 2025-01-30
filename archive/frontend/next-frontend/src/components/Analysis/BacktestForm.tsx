'use client';

import { useState } from 'react';
import {
  Box,
  Typography,
  TextField,
  Button,
  Grid,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Alert,
  Slider,
  InputAdornment,
  Switch,
  FormControlLabel,
} from '@mui/material';
import { DateTimePicker } from '@mui/x-date-pickers/DateTimePicker';
import { useAnalysis } from '@/contexts/AnalysisContext';
import { ANALYSIS_TIMEFRAMES } from '@/config/analysis';

export default function BacktestForm() {
  const { state, createBacktest, runBacktest } = useAnalysis();
  const [name, setName] = useState('');
  const [startTime, setStartTime] = useState<Date | null>(
    new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)
  );
  const [endTime, setEndTime] = useState<Date | null>(new Date());
  const [initialCapital, setInitialCapital] = useState('10000');
  const [leverage, setLeverage] = useState(1);
  const [timeframe, setTimeframe] = useState(ANALYSIS_TIMEFRAMES[4].value); // 1h default
  const [stopLoss, setStopLoss] = useState(2);
  const [takeProfit, setTakeProfit] = useState(4);
  const [trailingStop, setTrailingStop] = useState(false);
  const [maxDrawdown, setMaxDrawdown] = useState(10);
  const [positionSize, setPositionSize] = useState(1);
  const [error, setError] = useState('');

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();

    try {
      setError('');
      if (!startTime || !endTime) {
        setError('Start time and end time are required');
        return;
      }

      if (!state.selectedPair) {
        setError('Please select a trading pair');
        return;
      }

      const config = {
        name,
        startTime: startTime.getTime(),
        endTime: endTime.getTime(),
        initialCapital,
        leverage,
        pairs: [state.selectedPair.id],
        strategy: {
          id: state.selectedBacktest || 'default',
          params: {
            timeframe,
          },
        },
        riskManagement: {
          stopLoss,
          takeProfit,
          trailingStop,
          maxDrawdown,
          positionSize,
        },
      };

      const backtest = await createBacktest(config);
      await runBacktest(backtest.id);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to run backtest');
    }
  };

  return (
    <Box component="form" onSubmit={handleSubmit}>
      <Typography variant="h6" gutterBottom>
        Backtest Configuration
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <Grid container spacing={2}>
        <Grid item xs={12}>
          <TextField
            fullWidth
            label="Backtest Name"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
        </Grid>

        <Grid item xs={12} md={6}>
          <DateTimePicker
            label="Start Time"
            value={startTime}
            onChange={setStartTime}
            sx={{ width: '100%' }}
          />
        </Grid>

        <Grid item xs={12} md={6}>
          <DateTimePicker
            label="End Time"
            value={endTime}
            onChange={setEndTime}
            sx={{ width: '100%' }}
          />
        </Grid>

        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Initial Capital"
            type="number"
            value={initialCapital}
            onChange={(e) => setInitialCapital(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">$</InputAdornment>
              ),
            }}
          />
        </Grid>

        <Grid item xs={12} md={6}>
          <FormControl fullWidth>
            <InputLabel>Timeframe</InputLabel>
            <Select
              value={timeframe}
              label="Timeframe"
              onChange={(e) => setTimeframe(e.target.value)}
            >
              {ANALYSIS_TIMEFRAMES.map((tf) => (
                <MenuItem key={tf.value} value={tf.value}>
                  {tf.label}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </Grid>

        <Grid item xs={12}>
          <Typography gutterBottom>Leverage</Typography>
          <Slider
            value={leverage}
            onChange={(_, value) => setLeverage(value as number)}
            min={1}
            max={20}
            marks={[
              { value: 1, label: '1x' },
              { value: 5, label: '5x' },
              { value: 10, label: '10x' },
              { value: 20, label: '20x' },
            ]}
            valueLabelDisplay="auto"
            valueLabelFormat={(value) => `${value}x`}
          />
        </Grid>

        <Grid item xs={12}>
          <Typography variant="subtitle1" gutterBottom>
            Risk Management
          </Typography>
        </Grid>

        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Stop Loss"
            type="number"
            value={stopLoss}
            onChange={(e) => setStopLoss(Number(e.target.value))}
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
          />
        </Grid>

        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Take Profit"
            type="number"
            value={takeProfit}
            onChange={(e) => setTakeProfit(Number(e.target.value))}
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
          />
        </Grid>

        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Max Drawdown"
            type="number"
            value={maxDrawdown}
            onChange={(e) => setMaxDrawdown(Number(e.target.value))}
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
          />
        </Grid>

        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Position Size"
            type="number"
            value={positionSize}
            onChange={(e) => setPositionSize(Number(e.target.value))}
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
          />
        </Grid>

        <Grid item xs={12}>
          <FormControlLabel
            control={
              <Switch
                checked={trailingStop}
                onChange={(e) => setTrailingStop(e.target.checked)}
              />
            }
            label="Enable Trailing Stop"
          />
        </Grid>

        <Grid item xs={12}>
          <Button
            fullWidth
            variant="contained"
            color="primary"
            type="submit"
            disabled={state.isLoading}
          >
            Run Backtest
          </Button>
        </Grid>
      </Grid>
    </Box>
  );
} 