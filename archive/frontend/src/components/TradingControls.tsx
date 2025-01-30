import React from 'react';
import {
  Card,
  CardContent,
  Typography,
  Switch,
  Button,
  TextField,
  Grid,
  Box,
  FormControlLabel,
} from '@mui/material';

interface TradingConfig {
  maxAmount: number;
  stopLoss: number;
  takeProfit: number;
  walletAddress: string;
}

interface TradingControlsProps {
  config: TradingConfig;
  isTrading: boolean;
  onConfigChange: (config: TradingConfig) => void;
  onToggleTrading: () => void;
  onTransferProfit: () => void;
}

export const TradingControls: React.FC<TradingControlsProps> = ({
  config,
  isTrading,
  onConfigChange,
  onToggleTrading,
  onTransferProfit,
}) => {
  const handleConfigChange = (field: keyof TradingConfig) => (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const value = field === 'walletAddress' ? event.target.value : Number(event.target.value);
    onConfigChange({
      ...config,
      [field]: value,
    });
  };

  return (
    <Card>
      <CardContent>
        <Box mb={2}>
          <Typography variant="h6" gutterBottom>
            Trading Controls
          </Typography>
          <FormControlLabel
            control={
              <Switch
                checked={isTrading}
                onChange={onToggleTrading}
                color="primary"
                inputProps={{ 'aria-label': 'auto trading' }}
              />
            }
            label="Auto Trading"
          />
        </Box>

        <Grid container spacing={2}>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="Maximum Amount"
              type="number"
              value={config.maxAmount}
              onChange={handleConfigChange('maxAmount')}
              disabled={isTrading}
              inputProps={{
                'aria-label': 'maximum amount',
                min: 0,
                step: 0.1,
              }}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="Stop Loss"
              type="number"
              value={config.stopLoss}
              onChange={handleConfigChange('stopLoss')}
              disabled={isTrading}
              inputProps={{
                'aria-label': 'stop loss',
                min: 0,
                step: 0.1,
              }}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="Take Profit"
              type="number"
              value={config.takeProfit}
              onChange={handleConfigChange('takeProfit')}
              disabled={isTrading}
              inputProps={{
                'aria-label': 'take profit',
                min: 0,
                step: 0.1,
              }}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="Wallet Address"
              value={config.walletAddress}
              onChange={handleConfigChange('walletAddress')}
              disabled={isTrading}
              inputProps={{
                'aria-label': 'wallet address',
              }}
            />
          </Grid>
        </Grid>

        <Box mt={2} display="flex" justifyContent="space-between">
          <Button
            variant="contained"
            color={isTrading ? 'secondary' : 'primary'}
            onClick={onToggleTrading}
          >
            {isTrading ? 'Stop Trading' : 'Start Trading'}
          </Button>
          <Button
            variant="outlined"
            color="primary"
            onClick={onTransferProfit}
          >
            Transfer Profit
          </Button>
        </Box>
      </CardContent>
    </Card>
  );
};
