'use client';

import { useState, useEffect } from 'react';
import {
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  Typography,
  TextField,
  Box,
  Chip,
} from '@mui/material';
import { useTrading } from '@/contexts/TradingContext';

export default function TradingPairs() {
  const { state, selectPair } = useTrading();
  const [search, setSearch] = useState('');
  const [filteredPairs, setFilteredPairs] = useState(state.pairs);

  useEffect(() => {
    const filtered = state.pairs.filter(
      (pair) =>
        pair.baseToken.toLowerCase().includes(search.toLowerCase()) ||
        pair.quoteToken.toLowerCase().includes(search.toLowerCase())
    );
    setFilteredPairs(filtered);
  }, [search, state.pairs]);

  const handlePairSelect = (pair: any) => {
    selectPair(pair);
  };

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Trading Pairs
      </Typography>

      <TextField
        fullWidth
        size="small"
        placeholder="Search pairs..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        sx={{ mb: 2 }}
      />

      <List dense sx={{ maxHeight: 'calc(100vh - 200px)', overflow: 'auto' }}>
        {filteredPairs.map((pair) => (
          <ListItem
            key={pair.id}
            disablePadding
            selected={state.selectedPair?.id === pair.id}
          >
            <ListItemButton onClick={() => handlePairSelect(pair)}>
              <ListItemText
                primary={
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Typography variant="body1">
                      {pair.baseToken}/{pair.quoteToken}
                    </Typography>
                    <Chip
                      label={`Vol: ${Number(pair.volume24h).toLocaleString()}`}
                      size="small"
                      color="primary"
                      variant="outlined"
                    />
                  </Box>
                }
                secondary={
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography variant="body2" color="text.secondary">
                      ${Number(pair.lastPrice).toLocaleString()}
                    </Typography>
                    <Typography
                      variant="body2"
                      color={
                        Number(pair.priceChange24h) >= 0
                          ? 'success.main'
                          : 'error.main'
                      }
                    >
                      {Number(pair.priceChange24h).toFixed(2)}%
                    </Typography>
                  </Box>
                }
              />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
    </Box>
  );
} 