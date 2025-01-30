'use client';

import { useState, useEffect } from 'react';
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
  Tabs,
  Tab,
} from '@mui/material';
import { useTrading } from '@/contexts/TradingContext';

interface OrderBookEntry {
  price: string;
  amount: string;
  total: string;
  accumulated: string;
}

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
      id={`orderbook-tabpanel-${index}`}
      aria-labelledby={`orderbook-tab-${index}`}
      {...other}
    >
      {value === index && <Box>{children}</Box>}
    </div>
  );
}

export default function OrderBook() {
  const { state } = useTrading();
  const [activeTab, setActiveTab] = useState(0);
  const [asks, setAsks] = useState<OrderBookEntry[]>([]);
  const [bids, setBids] = useState<OrderBookEntry[]>([]);
  const [spread, setSpread] = useState({ amount: '0', percentage: '0' });

  useEffect(() => {
    if (!state.selectedPair) return;

    // Simulate WebSocket connection for order book updates
    const interval = setInterval(() => {
      // In a real application, this would be replaced with WebSocket data
      const mockOrderBook = generateMockOrderBook();
      setAsks(mockOrderBook.asks);
      setBids(mockOrderBook.bids);
      calculateSpread(mockOrderBook.asks[0]?.price, mockOrderBook.bids[0]?.price);
    }, 1000);

    return () => clearInterval(interval);
  }, [state.selectedPair]);

  const generateMockOrderBook = () => {
    const asks: OrderBookEntry[] = [];
    const bids: OrderBookEntry[] = [];
    const basePrice = 100;

    let accumulatedAsks = 0;
    let accumulatedBids = 0;

    for (let i = 0; i < 15; i++) {
      const askPrice = (basePrice + i * 0.1).toFixed(2);
      const bidPrice = (basePrice - i * 0.1).toFixed(2);
      const amount = (Math.random() * 10).toFixed(4);
      
      accumulatedAsks += parseFloat(amount);
      asks.push({
        price: askPrice,
        amount,
        total: (parseFloat(askPrice) * parseFloat(amount)).toFixed(4),
        accumulated: accumulatedAsks.toFixed(4),
      });

      accumulatedBids += parseFloat(amount);
      bids.push({
        price: bidPrice,
        amount,
        total: (parseFloat(bidPrice) * parseFloat(amount)).toFixed(4),
        accumulated: accumulatedBids.toFixed(4),
      });
    }

    return { asks, bids };
  };

  const calculateSpread = (askPrice: string, bidPrice: string) => {
    if (!askPrice || !bidPrice) return;

    const spread = parseFloat(askPrice) - parseFloat(bidPrice);
    const percentage = (spread / parseFloat(askPrice)) * 100;

    setSpread({
      amount: spread.toFixed(2),
      percentage: percentage.toFixed(2),
    });
  };

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setActiveTab(newValue);
  };

  const renderOrderTable = (orders: OrderBookEntry[], type: 'ask' | 'bid') => (
    <TableContainer component={Paper} variant="outlined">
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Price</TableCell>
            <TableCell align="right">Amount</TableCell>
            <TableCell align="right">Total</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {orders.map((order, index) => (
            <TableRow
              key={index}
              sx={{
                '&:hover': { backgroundColor: 'action.hover' },
                position: 'relative',
                '&::after': {
                  content: '""',
                  position: 'absolute',
                  right: 0,
                  top: 0,
                  height: '100%',
                  width: `${(parseFloat(order.accumulated) / parseFloat(orders[orders.length - 1].accumulated)) * 100}%`,
                  backgroundColor:
                    type === 'ask'
                      ? 'error.main'
                      : 'success.main',
                  opacity: 0.1,
                  zIndex: 0,
                },
                '& > *': {
                  position: 'relative',
                  zIndex: 1,
                },
              }}
            >
              <TableCell
                sx={{
                  color: type === 'ask' ? 'error.main' : 'success.main',
                }}
              >
                {order.price}
              </TableCell>
              <TableCell align="right">{order.amount}</TableCell>
              <TableCell align="right">{order.total}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Order Book
      </Typography>

      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
        <Tabs value={activeTab} onChange={handleTabChange} aria-label="order book tabs">
          <Tab label="All" />
          <Tab label="Bids" />
          <Tab label="Asks" />
        </Tabs>
      </Box>

      {state.selectedPair && (
        <Box sx={{ mb: 2 }}>
          <Typography variant="body2" color="text.secondary">
            Spread: {spread.amount} ({spread.percentage}%)
          </Typography>
        </Box>
      )}

      <TabPanel value={activeTab} index={0}>
        <Box sx={{ height: 'calc(100vh - 400px)', overflow: 'auto' }}>
          {renderOrderTable(asks, 'ask')}
          {renderOrderTable(bids, 'bid')}
        </Box>
      </TabPanel>

      <TabPanel value={activeTab} index={1}>
        <Box sx={{ height: 'calc(100vh - 400px)', overflow: 'auto' }}>
          {renderOrderTable(bids, 'bid')}
        </Box>
      </TabPanel>

      <TabPanel value={activeTab} index={2}>
        <Box sx={{ height: 'calc(100vh - 400px)', overflow: 'auto' }}>
          {renderOrderTable(asks, 'ask')}
        </Box>
      </TabPanel>
    </Box>
  );
} 