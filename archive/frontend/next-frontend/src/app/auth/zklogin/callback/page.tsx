'use client';

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import {
  Container,
  Paper,
  Box,
  Typography,
  CircularProgress,
  Alert,
} from '@mui/material';
import { useZkLogin } from '@/contexts/ZkLoginContext';

export default function ZkLoginCallback() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { state } = useZkLogin();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const handleCallback = async () => {
      try {
        // Check for error in URL parameters
        const errorParam = searchParams.get('error');
        if (errorParam) {
          throw new Error(errorParam);
        }

        // Wait for authentication to complete
        if (state.isAuthenticated) {
          router.push('/profile');
        }
      } catch (err) {
        console.error('zkLogin callback error:', err);
        setError(err instanceof Error ? err.message : 'Authentication failed');
        setTimeout(() => router.push('/auth/zklogin'), 3000);
      }
    };

    handleCallback();
  }, [router, searchParams, state.isAuthenticated]);

  return (
    <Container maxWidth="sm">
      <Box sx={{ mt: 8, mb: 4 }}>
        <Paper elevation={3} sx={{ p: 4 }}>
          {error ? (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
              <Typography variant="body2" sx={{ mt: 1 }}>
                Redirecting back to login...
              </Typography>
            </Alert>
          ) : (
            <Box sx={{ textAlign: 'center' }}>
              <CircularProgress sx={{ mb: 2 }} />
              <Typography>
                Completing authentication...
              </Typography>
            </Box>
          )}
        </Paper>
      </Box>
    </Container>
  );
} 