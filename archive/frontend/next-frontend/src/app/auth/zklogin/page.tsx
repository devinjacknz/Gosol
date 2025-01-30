'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import {
  Container,
  Paper,
  Box,
  Typography,
  Button,
  Alert,
  CircularProgress,
} from '@mui/material';
import { Google } from '@mui/icons-material';
import { useZkLogin } from '@/contexts/ZkLoginContext';

export default function ZkLogin() {
  const router = useRouter();
  const { state, login } = useZkLogin();

  useEffect(() => {
    if (state.isAuthenticated) {
      router.push('/profile');
    }
  }, [state.isAuthenticated, router]);

  return (
    <Container maxWidth="sm">
      <Box sx={{ mt: 8, mb: 4 }}>
        <Paper elevation={3} sx={{ p: 4 }}>
          <Typography component="h1" variant="h5" align="center" gutterBottom>
            Sign in with zkLogin
          </Typography>

          <Typography variant="body2" color="text.secondary" align="center" sx={{ mb: 3 }}>
            Secure and private authentication using zero-knowledge proofs
          </Typography>

          {state.error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {state.error}
            </Alert>
          )}

          <Button
            fullWidth
            variant="contained"
            startIcon={<Google />}
            onClick={login}
            disabled={state.isLoading}
            sx={{ mt: 2 }}
          >
            {state.isLoading ? (
              <CircularProgress size={24} color="inherit" />
            ) : (
              'Continue with Google'
            )}
          </Button>

          <Typography variant="body2" color="text.secondary" align="center" sx={{ mt: 3 }}>
            By continuing, you agree to our Terms of Service and Privacy Policy
          </Typography>
        </Paper>
      </Box>
    </Container>
  );
} 