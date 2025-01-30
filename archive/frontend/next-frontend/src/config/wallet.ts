import { type WalletAccount } from '@mysten/sui.js/client';

export const WALLET_CONFIG = {
  APP_NAME: 'Trading Platform',
  CHAIN_ID: process.env.NEXT_PUBLIC_SUI_CHAIN_ID || '1',
  NETWORK: process.env.NEXT_PUBLIC_SUI_NETWORK || 'testnet',
  RPC_URL: process.env.NEXT_PUBLIC_SUI_RPC_URL || 'https://fullnode.testnet.sui.io',
};

export interface WalletState {
  account: WalletAccount | null;
  isConnected: boolean;
  isConnecting: boolean;
  error: string | null;
  balance: string;
  network: string;
}

export const INITIAL_WALLET_STATE: WalletState = {
  account: null,
  isConnected: false,
  isConnecting: false,
  error: null,
  balance: '0',
  network: WALLET_CONFIG.NETWORK,
};

export type SupportedWallets = 'sui' | 'martian' | 'suiet' | 'ethos';

export interface WalletInfo {
  name: string;
  icon: string;
  downloadUrl: string;
}

export const SUPPORTED_WALLETS: Record<SupportedWallets, WalletInfo> = {
  sui: {
    name: 'Sui Wallet',
    icon: '/images/wallets/sui-wallet.svg',
    downloadUrl: 'https://chrome.google.com/webstore/detail/sui-wallet/opcgpfmipidbgpenhmajoajpbobppdil',
  },
  martian: {
    name: 'Martian Wallet',
    icon: '/images/wallets/martian-wallet.svg',
    downloadUrl: 'https://chrome.google.com/webstore/detail/martian-wallet-sui-wallet/efbglgofoippbgcjepnhiblaibcnclgk',
  },
  suiet: {
    name: 'Suiet Wallet',
    icon: '/images/wallets/suiet-wallet.svg',
    downloadUrl: 'https://chrome.google.com/webstore/detail/suiet-sui-wallet/khpkpbbcccdmmclmpigdgddabeilkdpd',
  },
  ethos: {
    name: 'Ethos Wallet',
    icon: '/images/wallets/ethos-wallet.svg',
    downloadUrl: 'https://chrome.google.com/webstore/detail/ethos-sui-wallet/mcbigmjiafegjnnogedioegffbooigli',
  },
}; 