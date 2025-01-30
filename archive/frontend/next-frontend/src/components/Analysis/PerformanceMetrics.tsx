'use client';

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
  Divider,
} from '@mui/material';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts';
import { useAnalysis } from '@/contexts/AnalysisContext';

const COLORS = ['#4caf50', '#f44336', '#2196f3', '#ff9800'];

export default function PerformanceMetrics() {
  const { state } = useAnalysis();
  const selectedResult = state.results.find(
    (result) => result.configId === state.selectedBacktest
  );

  if (!selectedResult) {
    return (
      <Box sx={{ p: 2 }}>
        <Typography color="text.secondary">
          No performance data available
        </Typography>
      </Box>
    );
  }

  const {
    totalPnl,
    winRate,
    sharpeRatio,
    maxDrawdown,
    averageTrade,
    profitFactor,
    recoveryFactor,
    expectancy,
    trades,
    winningTrades,
    losingTrades,
  } = selectedResult.metrics;

  const tradeDistribution = [
    { name: 'Winning', value: winningTrades },
    { name: 'Losing', value: losingTrades },
  ];

  const monthlyPerformance = calculateMonthlyPerformance(selectedResult.trades);

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Performance Analysis
      </Typography>

      <Grid container spacing={2}>
        {/* Key Metrics */}
        <Grid item xs={12}>
          <Paper sx={{ p: 2 }}>
            <Typography variant="subtitle1" gutterBottom>
              Key Metrics
            </Typography>
            <Grid container spacing={2}>
              <Grid item xs={12} md={3}>
                <Box>
                  <Typography variant="body2" color="text.secondary">
                    Total Return
                  </Typography>
                  <Typography
                    variant="h6"
                    color={
                      parseFloat(totalPnl) >= 0 ? 'success.main' : 'error.main'
                    }
                  >
                    ${parseFloat(totalPnl).toLocaleString()}
                  </Typography>
                </Box>
              </Grid>
              <Grid item xs={12} md={3}>
                <Box>
                  <Typography variant="body2" color="text.secondary">
                    Win Rate
                  </Typography>
                  <Typography variant="h6">
                    {(winRate * 100).toFixed(2)}%
                  </Typography>
                </Box>
              </Grid>
              <Grid item xs={12} md={3}>
                <Box>
                  <Typography variant="body2" color="text.secondary">
                    Profit Factor
                  </Typography>
                  <Typography variant="h6">{profitFactor.toFixed(2)}</Typography>
                </Box>
              </Grid>
              <Grid item xs={12} md={3}>
                <Box>
                  <Typography variant="body2" color="text.secondary">
                    Sharpe Ratio
                  </Typography>
                  <Typography variant="h6">{sharpeRatio.toFixed(2)}</Typography>
                </Box>
              </Grid>
            </Grid>
          </Paper>
        </Grid>

        {/* Trade Statistics */}
        <Grid item xs={12} md={6}>
          <Paper sx={{ p: 2, height: '100%' }}>
            <Typography variant="subtitle1" gutterBottom>
              Trade Statistics
            </Typography>
            <TableContainer>
              <Table size="small">
                <TableBody>
                  <TableRow>
                    <TableCell>Total Trades</TableCell>
                    <TableCell align="right">{trades}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Winning Trades</TableCell>
                    <TableCell align="right">{winningTrades}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Losing Trades</TableCell>
                    <TableCell align="right">{losingTrades}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Average Trade</TableCell>
                    <TableCell align="right">
                      ${parseFloat(averageTrade).toLocaleString()}
                    </TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Max Drawdown</TableCell>
                    <TableCell align="right">
                      {parseFloat(maxDrawdown).toFixed(2)}%
                    </TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Recovery Factor</TableCell>
                    <TableCell align="right">
                      {recoveryFactor.toFixed(2)}
                    </TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Expectancy</TableCell>
                    <TableCell align="right">${expectancy.toFixed(2)}</TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </TableContainer>
          </Paper>
        </Grid>

        {/* Trade Distribution */}
        <Grid item xs={12} md={6}>
          <Paper sx={{ p: 2, height: '100%' }}>
            <Typography variant="subtitle1" gutterBottom>
              Trade Distribution
            </Typography>
            <Box sx={{ height: 300 }}>
              <ResponsiveContainer>
                <PieChart>
                  <Pie
                    data={tradeDistribution}
                    dataKey="value"
                    nameKey="name"
                    cx="50%"
                    cy="50%"
                    outerRadius={100}
                    label={({
                      cx,
                      cy,
                      midAngle,
                      innerRadius,
                      outerRadius,
                      value,
                      name,
                    }) => {
                      const RADIAN = Math.PI / 180;
                      const radius = 25 + innerRadius + (outerRadius - innerRadius);
                      const x = cx + radius * Math.cos(-midAngle * RADIAN);
                      const y = cy + radius * Math.sin(-midAngle * RADIAN);

                      return (
                        <text
                          x={x}
                          y={y}
                          fill="#666"
                          textAnchor={x > cx ? 'start' : 'end'}
                          dominantBaseline="central"
                        >
                          {name} ({value})
                        </text>
                      );
                    }}
                  >
                    {tradeDistribution.map((entry, index) => (
                      <Cell
                        key={`cell-${index}`}
                        fill={COLORS[index % COLORS.length]}
                      />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            </Box>
          </Paper>
        </Grid>

        {/* Monthly Performance */}
        <Grid item xs={12}>
          <Paper sx={{ p: 2 }}>
            <Typography variant="subtitle1" gutterBottom>
              Monthly Performance
            </Typography>
            <Box sx={{ height: 300 }}>
              <ResponsiveContainer>
                <BarChart data={monthlyPerformance}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="month" />
                  <YAxis />
                  <Tooltip
                    formatter={(value) => `$${value.toLocaleString()}`}
                  />
                  <Legend />
                  <Bar
                    dataKey="profit"
                    name="Profit"
                    fill="#4caf50"
                    stackId="a"
                  />
                  <Bar
                    dataKey="loss"
                    name="Loss"
                    fill="#f44336"
                    stackId="a"
                  />
                </BarChart>
              </ResponsiveContainer>
            </Box>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
}

function calculateMonthlyPerformance(trades: any[]) {
  const monthlyData: Record<
    string,
    { month: string; profit: number; loss: number }
  > = {};

  trades.forEach((trade) => {
    const date = new Date(trade.exitTime);
    const month = `${date.getFullYear()}-${(date.getMonth() + 1)
      .toString()
      .padStart(2, '0')}`;

    if (!monthlyData[month]) {
      monthlyData[month] = {
        month,
        profit: 0,
        loss: 0,
      };
    }

    const pnl = parseFloat(trade.pnl);
    if (pnl >= 0) {
      monthlyData[month].profit += pnl;
    } else {
      monthlyData[month].loss += Math.abs(pnl);
    }
  });

  return Object.values(monthlyData).sort((a, b) => a.month.localeCompare(b.month));
} 