'use client';

import { Container, Box, Grid, Typography } from '@mui/material';
import dynamic from 'next/dynamic';
import { placeOrder } from '@/lib/tradeService';
import Navbar from '@/components/Navbar';

// Dynamically import components that use WebSocket to avoid SSR issues
const DynamicMarketData = dynamic(() => import('@/components/MarketData'), {
  ssr: false,
});

const DynamicTradeForm = dynamic(() => import('@/components/TradeForm'), {
  ssr: false,
});

const DynamicOrderHistory = dynamic(() => import('@/components/OrderHistory'), {
  ssr: false,
});

export default function Home() {
  const handleOrderSubmit = async (order: any) => {
    try {
      await placeOrder(order);
    } catch (error) {
      throw new Error(error instanceof Error ? error.message : 'Failed to place order');
    }
  };

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <Navbar />

      <Container maxWidth="xl" sx={{ mt: 4, mb: 4, flexGrow: 1 }}>
        <Grid container spacing={3}>
          {/* Market Data Section */}
          <Grid item xs={12}>
            <DynamicMarketData />
          </Grid>

          {/* Trading Section */}
          <Grid item xs={12} md={6}>
            <DynamicTradeForm onOrderSubmit={handleOrderSubmit} />
          </Grid>

          {/* Order History Section */}
          <Grid item xs={12} md={6}>
            <DynamicOrderHistory />
          </Grid>
        </Grid>
      </Container>

      <Box component="footer" sx={{ py: 3, px: 2, mt: 'auto', backgroundColor: 'background.paper' }}>
        <Container maxWidth="sm">
          <Typography variant="body2" color="text.secondary" align="center">
            © {new Date().getFullYear()} Trading Platform. All rights reserved.
          </Typography>
        </Container>
      </Box>
    </Box>
  );
}
