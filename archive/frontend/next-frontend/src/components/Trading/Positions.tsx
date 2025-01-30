'use client';

import {
  Box,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Button,
  Chip,
} from '@mui/material';
import { useTrading } from '@/contexts/TradingContext';

export default function Positions() {
  const { state } = useTrading();

  const formatPnL = (pnl: string) => {
    const value = parseFloat(pnl);
    const color = value >= 0 ? 'success.main' : 'error.main';
    return (
      <Typography
        variant="body2"
        color={color}
        sx={{ display: 'flex', alignItems: 'center' }}
      >
        {value >= 0 ? '+' : ''}
        {value.toFixed(2)}%
      </Typography>
    );
  };

  const formatPrice = (price: string) => {
    return parseFloat(price).toLocaleString(undefined, {
      minimumFractionDigits: 2,
      maximumFractionDigits: 8,
    });
  };

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Positions
      </Typography>

      <TableContainer component={Paper} variant="outlined">
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Pair</TableCell>
              <TableCell align="right">Side</TableCell>
              <TableCell align="right">Size</TableCell>
              <TableCell align="right">Entry Price</TableCell>
              <TableCell align="right">Mark Price</TableCell>
              <TableCell align="right">Liq. Price</TableCell>
              <TableCell align="right">Margin</TableCell>
              <TableCell align="right">PnL</TableCell>
              <TableCell align="center">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {state.positions.map((position) => (
              <TableRow key={position.id}>
                <TableCell>
                  {position.pair.baseToken}/{position.pair.quoteToken}
                </TableCell>
                <TableCell align="right">
                  <Chip
                    label={position.side}
                    size="small"
                    color={position.side === 'buy' ? 'success' : 'error'}
                  />
                </TableCell>
                <TableCell align="right">
                  {formatPrice(position.amount)}×{position.leverage}
                </TableCell>
                <TableCell align="right">
                  {formatPrice(position.entryPrice)}
                </TableCell>
                <TableCell align="right">
                  {formatPrice(position.markPrice || position.entryPrice)}
                </TableCell>
                <TableCell align="right">
                  {formatPrice(position.liquidationPrice)}
                </TableCell>
                <TableCell align="right">{formatPrice(position.margin)}</TableCell>
                <TableCell align="right">
                  {formatPnL(position.unrealizedPnl)}
                </TableCell>
                <TableCell align="center">
                  <Box sx={{ display: 'flex', gap: 1, justifyContent: 'center' }}>
                    <Button
                      variant="outlined"
                      size="small"
                      color={position.side === 'buy' ? 'error' : 'success'}
                    >
                      Close
                    </Button>
                    <Button variant="outlined" size="small">
                      TP/SL
                    </Button>
                  </Box>
                </TableCell>
              </TableRow>
            ))}
            {state.positions.length === 0 && (
              <TableRow>
                <TableCell colSpan={9} align="center">
                  No open positions
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
} 