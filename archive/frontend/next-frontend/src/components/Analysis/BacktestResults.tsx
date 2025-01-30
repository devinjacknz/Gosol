'use client';

import { useState, useEffect, useRef } from 'react';
import {
  Box,
  Typography,
  Grid,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Tabs,
  Tab,
} from '@mui/material';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { useAnalysis } from '@/contexts/AnalysisContext';

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
      id={`backtest-tabpanel-${index}`}
      aria-labelledby={`backtest-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 2 }}>{children}</Box>}
    </div>
  );
}

export default function BacktestResults() {
  const { state } = useAnalysis();
  const [activeTab, setActiveTab] = useState(0);
  const selectedResult = state.results.find(
    (result) => result.configId === state.selectedBacktest
  );

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setActiveTab(newValue);
  };

  if (!selectedResult) {
    return (
      <Box sx={{ p: 2 }}>
        <Typography color="text.secondary">No backtest results available</Typography>
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Backtest Results
      </Typography>

      <Grid container spacing={2} sx={{ mb: 2 }}>
        <Grid item xs={12} md={3}>
          <Paper sx={{ p: 2 }}>
            <Typography variant="subtitle2" color="text.secondary" gutterBottom>
              Total PnL
            </Typography>
            <Typography
              variant="h6"
              color={
                parseFloat(selectedResult.metrics.totalPnl) >= 0
                  ? 'success.main'
                  : 'error.main'
              }
            >
              ${parseFloat(selectedResult.metrics.totalPnl).toLocaleString()}
            </Typography>
          </Paper>
        </Grid>

        <Grid item xs={12} md={3}>
          <Paper sx={{ p: 2 }}>
            <Typography variant="subtitle2" color="text.secondary" gutterBottom>
              Win Rate
            </Typography>
            <Typography variant="h6">
              {(selectedResult.metrics.winRate * 100).toFixed(2)}%
            </Typography>
          </Paper>
        </Grid>

        <Grid item xs={12} md={3}>
          <Paper sx={{ p: 2 }}>
            <Typography variant="subtitle2" color="text.secondary" gutterBottom>
              Sharpe Ratio
            </Typography>
            <Typography variant="h6">
              {selectedResult.metrics.sharpeRatio.toFixed(2)}
            </Typography>
          </Paper>
        </Grid>

        <Grid item xs={12} md={3}>
          <Paper sx={{ p: 2 }}>
            <Typography variant="subtitle2" color="text.secondary" gutterBottom>
              Max Drawdown
            </Typography>
            <Typography variant="h6" color="error.main">
              {parseFloat(selectedResult.metrics.maxDrawdown).toFixed(2)}%
            </Typography>
          </Paper>
        </Grid>
      </Grid>

      <Paper>
        <Tabs
          value={activeTab}
          onChange={handleTabChange}
          aria-label="backtest results tabs"
        >
          <Tab label="Equity Curve" />
          <Tab label="Drawdown" />
          <Tab label="Trades" />
        </Tabs>

        <TabPanel value={activeTab} index={0}>
          <Box sx={{ height: 400 }}>
            <ResponsiveContainer>
              <LineChart data={selectedResult.equity}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis
                  dataKey="time"
                  tickFormatter={(time) =>
                    new Date(time).toLocaleDateString()
                  }
                />
                <YAxis />
                <Tooltip
                  labelFormatter={(time) =>
                    new Date(time).toLocaleString()
                  }
                  formatter={(value) =>
                    `$${parseFloat(value as string).toLocaleString()}`
                  }
                />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="value"
                  name="Equity"
                  stroke="#2196f3"
                  dot={false}
                />
              </LineChart>
            </ResponsiveContainer>
          </Box>
        </TabPanel>

        <TabPanel value={activeTab} index={1}>
          <Box sx={{ height: 400 }}>
            <ResponsiveContainer>
              <LineChart data={selectedResult.drawdown}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis
                  dataKey="time"
                  tickFormatter={(time) =>
                    new Date(time).toLocaleDateString()
                  }
                />
                <YAxis />
                <Tooltip
                  labelFormatter={(time) =>
                    new Date(time).toLocaleString()
                  }
                  formatter={(value) => `${value}%`}
                />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="value"
                  name="Drawdown"
                  stroke="#f44336"
                  dot={false}
                />
              </LineChart>
            </ResponsiveContainer>
          </Box>
        </TabPanel>

        <TabPanel value={activeTab} index={2}>
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Date</TableCell>
                  <TableCell>Pair</TableCell>
                  <TableCell>Side</TableCell>
                  <TableCell align="right">Entry Price</TableCell>
                  <TableCell align="right">Exit Price</TableCell>
                  <TableCell align="right">Amount</TableCell>
                  <TableCell align="right">PnL</TableCell>
                  <TableCell align="right">Fee</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {selectedResult.trades.map((trade) => (
                  <TableRow key={trade.id}>
                    <TableCell>
                      {new Date(trade.entryTime).toLocaleString()}
                    </TableCell>
                    <TableCell>{trade.pair}</TableCell>
                    <TableCell>
                      <Typography
                        color={
                          trade.side === 'buy' ? 'success.main' : 'error.main'
                        }
                      >
                        {trade.side.toUpperCase()}
                      </Typography>
                    </TableCell>
                    <TableCell align="right">
                      ${parseFloat(trade.entryPrice).toLocaleString()}
                    </TableCell>
                    <TableCell align="right">
                      ${parseFloat(trade.exitPrice).toLocaleString()}
                    </TableCell>
                    <TableCell align="right">
                      {parseFloat(trade.amount).toLocaleString()}
                    </TableCell>
                    <TableCell align="right">
                      <Typography
                        color={
                          parseFloat(trade.pnl) >= 0
                            ? 'success.main'
                            : 'error.main'
                        }
                      >
                        ${parseFloat(trade.pnl).toLocaleString()}
                        {' '}
                        ({trade.pnlPercent.toFixed(2)}%)
                      </Typography>
                    </TableCell>
                    <TableCell align="right">
                      ${parseFloat(trade.fee).toLocaleString()}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </TabPanel>
      </Paper>
    </Box>
  );
} 