'use client';

import { Container, Grid } from '@mui/material';
import AssetsList from '@/components/Assets/AssetsList';
import TransactionHistory from '@/components/Assets/TransactionHistory';
import { useWallet } from '@/contexts/WalletContext';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function AssetsPage() {
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
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Grid container spacing={3}>
        <Grid item xs={12}>
          <AssetsList />
        </Grid>
        <Grid item xs={12}>
          <TransactionHistory />
        </Grid>
      </Grid>
    </Container>
  );
} 