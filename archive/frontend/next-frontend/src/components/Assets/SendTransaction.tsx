'use client';

import { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Button,
  Alert,
  CircularProgress,
  Typography,
  Box,
} from '@mui/material';
import { TokenInfo } from '@/config/tokens';
import { useAssets } from '@/contexts/AssetsContext';

interface SendTransactionProps {
  open: boolean;
  onClose: () => void;
  token: TokenInfo;
}

export default function SendTransaction({ open, onClose, token }: SendTransactionProps) {
  const { state, sendTransaction } = useAssets();
  const [recipient, setRecipient] = useState('');
  const [amount, setAmount] = useState('');
  const [error, setError] = useState('');

  const handleSend = async () => {
    try {
      setError('');
      await sendTransaction(recipient, amount, token);
      onClose();
      setRecipient('');
      setAmount('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to send transaction');
    }
  };

  const balance = state.balances[token.address]?.balance || '0';
  const maxAmount = parseFloat(balance) / Math.pow(10, token.decimals);

  const handleSetMaxAmount = () => {
    setAmount(maxAmount.toString());
  };

  const isValidAmount = () => {
    const value = parseFloat(amount);
    return !isNaN(value) && value > 0 && value <= maxAmount;
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Send {token.symbol}</DialogTitle>
      <DialogContent>
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        <Box sx={{ mb: 2 }}>
          <Typography variant="body2" color="text.secondary">
            Available Balance: {maxAmount.toLocaleString()} {token.symbol}
          </Typography>
        </Box>

        <TextField
          margin="normal"
          fullWidth
          label="Recipient Address"
          value={recipient}
          onChange={(e) => setRecipient(e.target.value)}
          disabled={state.isLoading}
        />

        <Box sx={{ position: 'relative' }}>
          <TextField
            margin="normal"
            fullWidth
            label="Amount"
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            disabled={state.isLoading}
            InputProps={{
              endAdornment: (
                <Button
                  variant="text"
                  size="small"
                  onClick={handleSetMaxAmount}
                  sx={{ position: 'absolute', right: 8 }}
                >
                  MAX
                </Button>
              ),
            }}
          />
        </Box>

        <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
          Network Fee: Calculating...
        </Typography>
      </DialogContent>

      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button
          onClick={handleSend}
          variant="contained"
          disabled={
            state.isLoading ||
            !recipient ||
            !amount ||
            !isValidAmount()
          }
        >
          {state.isLoading ? <CircularProgress size={24} /> : 'Send'}
        </Button>
      </DialogActions>
    </Dialog>
  );
} 