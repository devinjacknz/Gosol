import React from 'react';
import {
  Card,
  CardContent,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  TablePagination,
  Chip,
  Box,
} from '@mui/material';
import { format } from 'date-fns';

interface Trade {
  id: string;
  type: 'buy' | 'sell';
  amount: number;
  price: number;
  timestamp: string;
  profit?: number;
}

interface TradeHistoryProps {
  tokenAddress: string;
  trades?: Trade[];
  pageSize?: number;
}

const TradeHistory: React.FC<TradeHistoryProps> = ({
  tokenAddress,
  trades = [],
  pageSize = 10,
}) => {
  const [page, setPage] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(pageSize);

  const handleChangePage = (_: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price);
  };

  const formatAmount = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      maximumFractionDigits: 6,
    }).format(amount);
  };

  const formatProfit = (profit: number | undefined) => {
    if (profit === undefined) return '-';
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      signDisplay: 'always',
    }).format(profit);
  };

  const getTypeColor = (type: 'buy' | 'sell') => {
    return type === 'buy' ? 'success' : 'error';
  };

  if (trades.length === 0) {
    return (
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Trade History
          </Typography>
          <Box sx={{ textAlign: 'center', py: 4 }}>
            <Typography color="textSecondary">
              No trades have been executed yet
            </Typography>
          </Box>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          Trade History
        </Typography>
        <TableContainer component={Paper} variant="outlined">
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Type</TableCell>
                <TableCell align="right">Amount</TableCell>
                <TableCell align="right">Price</TableCell>
                <TableCell align="right">Profit/Loss</TableCell>
                <TableCell>Time</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {trades
                .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
                .map((trade) => (
                  <TableRow key={trade.id}>
                    <TableCell>
                      <Chip
                        label={trade.type.toUpperCase()}
                        color={getTypeColor(trade.type)}
                        size="small"
                      />
                    </TableCell>
                    <TableCell align="right">
                      {formatAmount(trade.amount)}
                    </TableCell>
                    <TableCell align="right">
                      {formatPrice(trade.price)}
                    </TableCell>
                    <TableCell
                      align="right"
                      sx={{
                        color: trade.profit
                          ? trade.profit > 0
                            ? 'success.main'
                            : 'error.main'
                          : 'inherit',
                      }}
                    >
                      {formatProfit(trade.profit)}
                    </TableCell>
                    <TableCell>
                      {format(new Date(trade.timestamp), 'MMM d, HH:mm:ss')}
                    </TableCell>
                  </TableRow>
                ))}
            </TableBody>
          </Table>
        </TableContainer>
        <TablePagination
          rowsPerPageOptions={[5, 10, 25]}
          component="div"
          count={trades.length}
          rowsPerPage={rowsPerPage}
          page={page}
          onPageChange={handleChangePage}
          onRowsPerPageChange={handleChangeRowsPerPage}
        />
      </CardContent>
    </Card>
  );
};

export default TradeHistory;
