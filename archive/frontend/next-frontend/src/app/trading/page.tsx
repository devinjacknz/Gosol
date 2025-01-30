'use client';

import { useEffect } from 'react';
import {
  Container,
  Grid,
  Paper,
  Box,
} from '@mui/material';
import { useRouter } from 'next/navigation';
import { useWallet } from '@/contexts/WalletContext';
import TradingChart from '@/components/Trading/TradingChart';
import OrderBook from '@/components/Trading/OrderBook';
import TradeForm from '@/components/Trading/TradeForm';
import OpenOrders from '@/components/Trading/OpenOrders';
import Positions from '@/components/Trading/Positions';
import TradingPairs from '@/components/Trading/TradingPairs';

export default function TradingPage() {
  const { state: walletState } = useWallet();
  const router = useRouter();

  useEffect(() => {
    if (!walletState.isConnected) {
      router.push('/');
    }
  }, [walletState.isConnected, router]);

  if (!walletState.isConnected) {
    return null;
  }

  return (
    <Container maxWidth={false} sx={{ mt: 4, mb: 4 }}>
      <Grid container spacing={2}>
        {/* Left Column - Trading Pairs */}
        <Grid item xs={12} md={2}>
          <Paper sx={{ p: 2, height: '100%' }}>
            <TradingPairs />
          </Paper>
        </Grid>

        {/* Middle Column - Chart and Order Form */}
        <Grid item xs={12} md={7}>
          <Grid container spacing={2}>
            {/* Trading Chart */}
            <Grid item xs={12}>
              <Paper sx={{ p: 2, height: '500px' }}>
                <TradingChart />
              </Paper>
            </Grid>

            {/* Trade Form */}
            <Grid item xs={12}>
              <Paper sx={{ p: 2 }}>
                <TradeForm />
              </Paper>
            </Grid>

            {/* Open Orders and Positions */}
            <Grid item xs={12}>
              <Paper sx={{ p: 2 }}>
                <Box sx={{ mb: 2 }}>
                  <OpenOrders />
                </Box>
                <Positions />
              </Paper>
            </Grid>
          </Grid>
        </Grid>

        {/* Right Column - Order Book */}
        <Grid item xs={12} md={3}>
          <Paper sx={{ p: 2, height: '100%' }}>
            <OrderBook />
          </Paper>
        </Grid>
      </Grid>
    </Container>
  );
} 