'use client';

import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  Box,
  Chip,
  IconButton,
  Tooltip,
} from '@mui/material';
import {
  CallMade,
  CallReceived,
  SwapHoriz,
  ContentCopy,
  OpenInNew,
} from '@mui/icons-material';
import { useAssets } from '@/contexts/AssetsContext';
import { TRANSACTION_STATUS } from '@/config/tokens';

export default function TransactionHistory() {
  const { state } = useAssets();

  const handleCopyAddress = (address: string) => {
    navigator.clipboard.writeText(address);
  };

  const handleOpenExplorer = (txId: string) => {
    window.open(`https://explorer.sui.io/transaction/${txId}`, '_blank');
  };

  const formatAddress = (address: string) => {
    return `${address.slice(0, 6)}...${address.slice(-4)}`;
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case TRANSACTION_STATUS.SUCCESS:
        return 'success';
      case TRANSACTION_STATUS.PENDING:
        return 'warning';
      case TRANSACTION_STATUS.FAILED:
        return 'error';
      default:
        return 'default';
    }
  };

  const getTransactionIcon = (type: string) => {
    switch (type) {
      case 'send':
        return <CallMade color="error" />;
      case 'receive':
        return <CallReceived color="success" />;
      case 'swap':
        return <SwapHoriz color="primary" />;
      default:
        return null;
    }
  };

  return (
    <Paper elevation={3} sx={{ p: 3 }}>
      <Typography variant="h6" gutterBottom>
        Transaction History
      </Typography>

      <TableContainer>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Type</TableCell>
              <TableCell>From/To</TableCell>
              <TableCell align="right">Amount</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Time</TableCell>
              <TableCell align="center">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {state.transactions.map((tx) => (
              <TableRow key={tx.id}>
                <TableCell>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    {getTransactionIcon(tx.type)}
                    <Typography
                      variant="body2"
                      color={tx.type === 'send' ? 'error' : 'success'}
                    >
                      {tx.type.toUpperCase()}
                    </Typography>
                  </Box>
                </TableCell>
                <TableCell>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Typography variant="body2">
                      {tx.type === 'send' ? 'To: ' : 'From: '}
                      {formatAddress(tx.type === 'send' ? tx.to : tx.from)}
                    </Typography>
                    <IconButton
                      size="small"
                      onClick={() =>
                        handleCopyAddress(tx.type === 'send' ? tx.to : tx.from)
                      }
                    >
                      <ContentCopy fontSize="small" />
                    </IconButton>
                  </Box>
                </TableCell>
                <TableCell align="right">
                  <Typography variant="body2">
                    {tx.amount} {tx.token.symbol}
                  </Typography>
                  {tx.fee && (
                    <Typography variant="caption" color="text.secondary">
                      Fee: {tx.fee} SUI
                    </Typography>
                  )}
                </TableCell>
                <TableCell>
                  <Chip
                    label={tx.status}
                    color={getStatusColor(tx.status)}
                    size="small"
                  />
                </TableCell>
                <TableCell>
                  <Tooltip
                    title={new Date(tx.timestamp).toLocaleString()}
                    placement="top"
                  >
                    <Typography variant="body2">
                      {new Date(tx.timestamp).toLocaleDateString()}
                    </Typography>
                  </Tooltip>
                </TableCell>
                <TableCell align="center">
                  <IconButton
                    size="small"
                    onClick={() => handleOpenExplorer(tx.id)}
                  >
                    <OpenInNew fontSize="small" />
                  </IconButton>
                </TableCell>
              </TableRow>
            ))}
            {state.transactions.length === 0 && (
              <TableRow>
                <TableCell colSpan={6} align="center">
                  No transactions found
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </Paper>
  );
} 