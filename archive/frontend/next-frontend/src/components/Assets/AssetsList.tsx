'use client';

import { useState, useEffect } from 'react';
import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  IconButton,
  Box,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Alert,
  CircularProgress,
} from '@mui/material';
import {
  Send,
  Add as AddIcon,
  Delete as DeleteIcon,
} from '@mui/icons-material';
import Image from 'next/image';
import { useAssets } from '@/contexts/AssetsContext';
import { TokenInfo } from '@/config/tokens';

export default function AssetsList() {
  const { state, addToken, removeToken, refreshBalances } = useAssets();
  const [openAddToken, setOpenAddToken] = useState(false);
  const [openSend, setOpenSend] = useState(false);
  const [selectedToken, setSelectedToken] = useState<TokenInfo | null>(null);
  const [newToken, setNewToken] = useState({
    address: '',
    symbol: '',
    name: '',
    decimals: '9',
  });
  const [error, setError] = useState('');

  useEffect(() => {
    refreshBalances();
    // Set up polling for balance updates
    const interval = setInterval(refreshBalances, 30000);
    return () => clearInterval(interval);
  }, [refreshBalances]);

  const handleAddToken = async () => {
    try {
      setError('');
      await addToken({
        address: newToken.address,
        symbol: newToken.symbol,
        name: newToken.name,
        decimals: parseInt(newToken.decimals),
      });
      setOpenAddToken(false);
      setNewToken({ address: '', symbol: '', name: '', decimals: '9' });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add token');
    }
  };

  const handleRemoveToken = async (token: TokenInfo) => {
    if (token.symbol === 'SUI') return; // Prevent removing native token
    removeToken(token.address);
  };

  const formatBalance = (balance: string, decimals: number) => {
    const value = parseFloat(balance) / Math.pow(10, decimals);
    return value.toLocaleString(undefined, {
      minimumFractionDigits: 2,
      maximumFractionDigits: 6,
    });
  };

  return (
    <>
      <Paper elevation={3} sx={{ p: 3, mb: 3 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Typography variant="h6">Assets</Typography>
          <Button
            variant="outlined"
            startIcon={<AddIcon />}
            onClick={() => setOpenAddToken(true)}
          >
            Add Token
          </Button>
        </Box>

        {state.error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {state.error}
          </Alert>
        )}

        <TableContainer>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Token</TableCell>
                <TableCell align="right">Balance</TableCell>
                <TableCell align="right">Value (USD)</TableCell>
                <TableCell align="center">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {Object.values(state.balances).map((balance) => (
                <TableRow key={balance.token.address}>
                  <TableCell>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      {balance.token.icon && (
                        <Image
                          src={balance.token.icon}
                          alt={balance.token.symbol}
                          width={24}
                          height={24}
                        />
                      )}
                      <Box>
                        <Typography variant="body1">
                          {balance.token.symbol}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          {balance.token.name}
                        </Typography>
                      </Box>
                    </Box>
                  </TableCell>
                  <TableCell align="right">
                    {formatBalance(balance.balance, balance.token.decimals)}
                  </TableCell>
                  <TableCell align="right">
                    ${balance.value.toLocaleString()}
                  </TableCell>
                  <TableCell align="center">
                    <IconButton
                      onClick={() => {
                        setSelectedToken(balance.token);
                        setOpenSend(true);
                      }}
                    >
                      <Send />
                    </IconButton>
                    {balance.token.symbol !== 'SUI' && (
                      <IconButton
                        onClick={() => handleRemoveToken(balance.token)}
                        color="error"
                      >
                        <DeleteIcon />
                      </IconButton>
                    )}
                  </TableCell>
                </TableRow>
              ))}
              {Object.keys(state.balances).length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} align="center">
                    {state.isLoading ? (
                      <CircularProgress size={24} />
                    ) : (
                      'No assets found'
                    )}
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      </Paper>

      {/* Add Token Dialog */}
      <Dialog
        open={openAddToken}
        onClose={() => setOpenAddToken(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>Add Token</DialogTitle>
        <DialogContent>
          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}
          <TextField
            margin="normal"
            fullWidth
            label="Contract Address"
            value={newToken.address}
            onChange={(e) => setNewToken({ ...newToken, address: e.target.value })}
          />
          <TextField
            margin="normal"
            fullWidth
            label="Symbol"
            value={newToken.symbol}
            onChange={(e) => setNewToken({ ...newToken, symbol: e.target.value })}
          />
          <TextField
            margin="normal"
            fullWidth
            label="Name"
            value={newToken.name}
            onChange={(e) => setNewToken({ ...newToken, name: e.target.value })}
          />
          <TextField
            margin="normal"
            fullWidth
            label="Decimals"
            type="number"
            value={newToken.decimals}
            onChange={(e) => setNewToken({ ...newToken, decimals: e.target.value })}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenAddToken(false)}>Cancel</Button>
          <Button
            onClick={handleAddToken}
            variant="contained"
            disabled={state.isLoading}
          >
            {state.isLoading ? <CircularProgress size={24} /> : 'Add Token'}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
} 