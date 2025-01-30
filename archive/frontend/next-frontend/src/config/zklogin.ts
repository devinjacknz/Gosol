import { type SuiClient } from '@mysten/sui.js/client';
import { type ZkLoginProvider } from '@mysten/zklogin';

export const ZKLOGIN_CONFIG = {
  // Google OAuth client ID
  CLIENT_ID: process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID || '',
  // Redirect URL after successful OAuth login
  REDIRECT_URL: process.env.NEXT_PUBLIC_ZKLOGIN_REDIRECT_URL || 'http://localhost:3000/auth/zklogin/callback',
  // Sui network to use (testnet, mainnet, etc.)
  NETWORK: process.env.NEXT_PUBLIC_SUI_NETWORK || 'testnet',
  // Sui RPC endpoint
  RPC_URL: process.env.NEXT_PUBLIC_SUI_RPC_URL || 'https://fullnode.testnet.sui.io',
};

export interface ZkLoginState {
  provider: ZkLoginProvider | null;
  suiClient: SuiClient | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  address: string | null;
}

export const INITIAL_STATE: ZkLoginState = {
  provider: null,
  suiClient: null,
  isAuthenticated: false,
  isLoading: false,
  error: null,
  address: null,
}; 