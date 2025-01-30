'use client';

import { useSearchParams } from 'next/navigation';
import { Container, Paper, Typography, Button, Box } from '@mui/material';
import { useRouter } from 'next/navigation';

export default function AuthError() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const error = searchParams.get('error');

  const getErrorMessage = (errorCode: string) => {
    switch (errorCode) {
      case 'Configuration':
        return 'There is a problem with the server configuration.';
      case 'AccessDenied':
        return 'You do not have permission to sign in.';
      case 'Verification':
        return 'The verification code has expired or has already been used.';
      default:
        return 'An error occurred during authentication.';
    }
  };

  return (
    <Container maxWidth="sm">
      <Box sx={{ mt: 8, mb: 4 }}>
        <Paper elevation={3} sx={{ p: 4 }}>
          <Typography variant="h5" component="h1" gutterBottom color="error">
            Authentication Error
          </Typography>
          <Typography variant="body1" sx={{ mb: 3 }}>
            {error ? getErrorMessage(error) : 'An unknown error occurred.'}
          </Typography>
          <Box sx={{ display: 'flex', gap: 2 }}>
            <Button
              variant="contained"
              onClick={() => router.push('/auth/signin')}
            >
              Try Again
            </Button>
            <Button
              variant="outlined"
              onClick={() => router.push('/')}
            >
              Return Home
            </Button>
          </Box>
        </Paper>
      </Box>
    </Container>
  );
} 