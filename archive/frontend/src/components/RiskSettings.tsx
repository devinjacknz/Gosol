import React from 'react';
import {
  Card,
  CardContent,
  Typography,
  TextField,
  Button,
  Grid,
  Box,
  Slider,
  FormControlLabel,
  Switch,
} from '@mui/material';

interface RiskConfig {
  maxLossPerTrade: number;
  maxDailyLoss: number;
  maxPositionSize: number;
  stopLossEnabled: boolean;
  stopLossPercentage: number;
  takeProfitEnabled: boolean;
  takeProfitPercentage: number;
}

interface RiskSettingsProps {
  tokenAddress: string;
  onUpdate: (config: RiskConfig) => void;
}

const RiskSettings: React.FC<RiskSettingsProps> = ({ tokenAddress, onUpdate }) => {
  const [config, setConfig] = React.useState<RiskConfig>({
    maxLossPerTrade: 1,
    maxDailyLoss: 5,
    maxPositionSize: 10,
    stopLossEnabled: true,
    stopLossPercentage: 2,
    takeProfitEnabled: true,
    takeProfitPercentage: 5,
  });

  const handleTextFieldChange = (field: keyof RiskConfig) => (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const value = parseFloat(event.target.value);
    if (!isNaN(value)) {
      setConfig((prev) => ({ ...prev, [field]: value }));
    }
  };

  const handleSliderChange = (_: Event, value: number | number[]) => {
    setConfig((prev) => ({ ...prev, maxPositionSize: value as number }));
  };

  const handleSwitchChange = (field: keyof RiskConfig) => (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    setConfig((prev) => ({ ...prev, [field]: event.target.checked }));
  };

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    onUpdate(config);
  };

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          Risk Management Settings
        </Typography>
        <Box component="form" onSubmit={handleSubmit}>
          <Grid container spacing={3}>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Max Loss Per Trade (%)"
                type="number"
                value={config.maxLossPerTrade}
                onChange={handleTextFieldChange('maxLossPerTrade')}
                inputProps={{ min: 0, max: 100, step: 0.1 }}
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Max Daily Loss (%)"
                type="number"
                value={config.maxDailyLoss}
                onChange={handleTextFieldChange('maxDailyLoss')}
                inputProps={{ min: 0, max: 100, step: 0.1 }}
              />
            </Grid>
            <Grid item xs={12}>
              <Typography gutterBottom>Max Position Size (%)</Typography>
              <Slider
                value={config.maxPositionSize}
                onChange={handleSliderChange}
                min={1}
                max={100}
                step={1}
                valueLabelDisplay="auto"
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <FormControlLabel
                control={
                  <Switch
                    checked={config.stopLossEnabled}
                    onChange={handleSwitchChange('stopLossEnabled')}
                  />
                }
                label="Enable Stop Loss"
              />
              {config.stopLossEnabled && (
                <TextField
                  fullWidth
                  label="Stop Loss (%)"
                  type="number"
                  value={config.stopLossPercentage}
                  onChange={handleTextFieldChange('stopLossPercentage')}
                  inputProps={{ min: 0, max: 100, step: 0.1 }}
                />
              )}
            </Grid>
            <Grid item xs={12} md={6}>
              <FormControlLabel
                control={
                  <Switch
                    checked={config.takeProfitEnabled}
                    onChange={handleSwitchChange('takeProfitEnabled')}
                  />
                }
                label="Enable Take Profit"
              />
              {config.takeProfitEnabled && (
                <TextField
                  fullWidth
                  label="Take Profit (%)"
                  type="number"
                  value={config.takeProfitPercentage}
                  onChange={handleTextFieldChange('takeProfitPercentage')}
                  inputProps={{ min: 0, max: 100, step: 0.1 }}
                />
              )}
            </Grid>
            <Grid item xs={12}>
              <Button
                type="submit"
                variant="contained"
                color="primary"
                fullWidth
              >
                Update Risk Settings
              </Button>
            </Grid>
          </Grid>
        </Box>
      </CardContent>
    </Card>
  );
};

export default RiskSettings;
