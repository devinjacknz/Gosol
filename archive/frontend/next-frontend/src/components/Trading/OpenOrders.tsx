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
  IconButton,
  Tooltip,
} from '@mui/material';
import { Delete as DeleteIcon } from '@mui/icons-material';
import { useTrading } from '@/contexts/TradingContext';

export default function OpenOrders() {
  const { state, cancelOrder } = useTrading();

  const handleCancelOrder = async (orderId: string) => {
    try {
      await cancelOrder(orderId);
    } catch (error) {
      console.error('Failed to cancel order:', error);
    }
  };

  const formatPrice = (price: string) => {
    return parseFloat(price).toLocaleString(undefined, {
      minimumFractionDigits: 2,
      maximumFractionDigits: 8,
    });
  };

  const getOrderTypeChip = (type: string) => {
    let color:
      | 'default'
      | 'primary'
      | 'secondary'
      | 'error'
      | 'info'
      | 'success'
      | 'warning' = 'default';

    switch (type) {
      case 'limit':
        color = 'primary';
        break;
      case 'market':
        color = 'secondary';
        break;
      case 'stop_limit':
        color = 'warning';
        break;
      case 'stop_market':
        color = 'error';
        break;
      case 'trailing_stop':
        color = 'info';
        break;
    }

    return (
      <Chip
        label={type.replace('_', ' ').toUpperCase()}
        size="small"
        color={color}
      />
    );
  };

  const getProgressPercentage = (filled: string, amount: string) => {
    return (parseFloat(filled) / parseFloat(amount)) * 100;
  };

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Open Orders
      </Typography>

      <TableContainer component={Paper} variant="outlined">
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Date</TableCell>
              <TableCell>Pair</TableCell>
              <TableCell>Type</TableCell>
              <TableCell>Side</TableCell>
              <TableCell align="right">Price</TableCell>
              <TableCell align="right">Amount</TableCell>
              <TableCell align="right">Filled</TableCell>
              <TableCell align="right">Total</TableCell>
              <TableCell align="center">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {state.orders
              .filter((order) => order.status === 'open' || order.status === 'partially_filled')
              .map((order) => (
                <TableRow key={order.id}>
                  <TableCell>
                    {new Date(order.timestamp).toLocaleString()}
                  </TableCell>
                  <TableCell>
                    {order.pair.baseToken}/{order.pair.quoteToken}
                  </TableCell>
                  <TableCell>{getOrderTypeChip(order.type)}</TableCell>
                  <TableCell>
                    <Chip
                      label={order.side.toUpperCase()}
                      size="small"
                      color={order.side === 'buy' ? 'success' : 'error'}
                    />
                  </TableCell>
                  <TableCell align="right">{formatPrice(order.price)}</TableCell>
                  <TableCell align="right">{formatPrice(order.amount)}</TableCell>
                  <TableCell align="right">
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Box
                        sx={{
                          width: '60px',
                          height: '4px',
                          bgcolor: 'grey.300',
                          borderRadius: '2px',
                          overflow: 'hidden',
                        }}
                      >
                        <Box
                          sx={{
                            width: `${getProgressPercentage(
                              order.filled,
                              order.amount
                            )}%`,
                            height: '100%',
                            bgcolor: 'primary.main',
                          }}
                        />
                      </Box>
                      <Typography variant="body2">
                        {formatPrice(order.filled)}
                      </Typography>
                    </Box>
                  </TableCell>
                  <TableCell align="right">
                    {formatPrice(
                      (
                        parseFloat(order.price) * parseFloat(order.amount)
                      ).toString()
                    )}
                  </TableCell>
                  <TableCell align="center">
                    <Tooltip title="Cancel Order">
                      <IconButton
                        size="small"
                        onClick={() => handleCancelOrder(order.id)}
                        color="error"
                      >
                        <DeleteIcon />
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              ))}
            {state.orders.filter(
              (order) =>
                order.status === 'open' || order.status === 'partially_filled'
            ).length === 0 && (
              <TableRow>
                <TableCell colSpan={9} align="center">
                  No open orders
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
} 