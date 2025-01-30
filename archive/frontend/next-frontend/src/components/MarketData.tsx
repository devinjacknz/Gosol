'use client';

import { useEffect } from 'react';
import { useWebSocketStore } from '@/lib/websocket';
import { Card, CardContent, Typography, Grid, Box } from '@mui/material';
import { TrendingUp, TrendingDown } from '@mui/icons-material';

interface PriceCardProps {
  symbol: string;
  price: number;
  volume: number;
  timestamp: number;
}

const PriceCard = ({ symbol, price, volume, timestamp }: PriceCardProps) => {
  return (
    <Card sx={{ minWidth: 275, m: 1 }}>
      <CardContent>
        <Typography variant="h5" component="div">
          {symbol}
        </Typography>
        <Box sx={{ display: 'flex', alignItems: 'center', mt: 1 }}>
          <Typography variant="h4" component="div" color="primary">
            ${price.toFixed(2)}
          </Typography>
          {price > 0 ? (
            <TrendingUp color="success" sx={{ ml: 1 }} />
          ) : (
            <TrendingDown color="error" sx={{ ml: 1 }} />
          )}
        </Box>
        <Typography sx={{ mb: 1.5 }} color="text.secondary">
          Volume: {volume.toLocaleString()}
        </Typography>
        <Typography variant="body2">
          Last Update: {new Date(timestamp).toLocaleTimeString()}
        </Typography>
      </CardContent>
    </Card>
  );
};

export default function MarketData() {
  const { marketData, connect, isConnected } = useWebSocketStore();

  useEffect(() => {
    if (!isConnected) {
      connect();
    }
    return () => {
      useWebSocketStore.getState().disconnect();
    };
  }, [connect, isConnected]);

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" sx={{ mb: 3 }}>
        Market Data {isConnected ? '🟢' : '🔴'}
      </Typography>
      <Grid container spacing={2}>
        {Object.entries(marketData).map(([symbol, data]) => (
          <Grid item xs={12} sm={6} md={4} key={symbol}>
            <PriceCard
              symbol={symbol}
              price={data.price}
              volume={data.volume}
              timestamp={data.timestamp}
            />
          </Grid>
        ))}
      </Grid>
    </Box>
  );
} 