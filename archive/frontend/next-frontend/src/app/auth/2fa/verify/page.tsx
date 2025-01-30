'use client';

import { useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import {
  Container,
  Paper,
  Box,
  Typography,
  TextField,
  Button,
  Alert,
  CircularProgress,
} from '@mui/material';
import { Security } from '@mui/icons-material';

export default function Verify2FA() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const callbackUrl = searchParams.get('callbackUrl') || '/';
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [code, setCode] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const response = await fetch('/api/auth/2fa/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || 'Invalid verification code');
      }

      router.push(callbackUrl);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to verify code');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxWidth="sm">
      <Box sx={{ mt: 8, mb: 4 }}>
        <Paper elevation={3} sx={{ p: 4 }}>
          <Box sx={{ textAlign: 'center', mb: 3 }}>
            <Security sx={{ fontSize: 48, color: 'primary.main' }} />
          </Box>

          <Typography component="h1" variant="h5" align="center" gutterBottom>
            Two-Factor Authentication
          </Typography>

          <Typography variant="body2" color="text.secondary" align="center" sx={{ mb: 3 }}>
            Enter the verification code from your authenticator app
          </Typography>

          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          <Box component="form" onSubmit={handleSubmit}>
            <TextField
              margin="normal"
              required
              fullWidth
              label="Verification Code"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              disabled={loading}
              inputProps={{
                maxLength: 6,
                pattern: '[0-9]*',
              }}
              autoComplete="one-time-code"
              autoFocus
            />

            <Button
              type="submit"
              fullWidth
              variant="contained"
              sx={{ mt: 3, mb: 2 }}
              disabled={code.length !== 6 || loading}
            >
              {loading ? <CircularProgress size={24} /> : 'Verify'}
            </Button>

            <Typography variant="body2" color="text.secondary" align="center">
              Lost access to your authenticator app?{' '}
              <Button
                variant="text"
                size="small"
                onClick={() => router.push('/auth/2fa/recovery')}
              >
                Use a recovery code
              </Button>
            </Typography>
          </Box>
        </Paper>
      </Box>
    </Container>
  );
} 