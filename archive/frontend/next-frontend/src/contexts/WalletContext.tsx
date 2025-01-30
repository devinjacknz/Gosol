'use client';

import React, { createContext, useContext, useReducer, useCallback, useEffect } from 'react';
import { SuiClient, getFullnodeUrl } from '@mysten/sui.js/client';
import { WALLET_CONFIG, WalletState, INITIAL_WALLET_STATE, SupportedWallets } from '@/config/wallet';

type Action =
  | { type: 'SET_CONNECTING'; payload: boolean }
  | { type: 'SET_CONNECTED'; payload: boolean }
  | { type: 'SET_ACCOUNT'; payload: any }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'SET_BALANCE'; payload: string }
  | { type: 'SET_NETWORK'; payload: string }
  | { type: 'RESET' };

interface WalletContextType {
  state: WalletState;
  connect: (walletType: SupportedWallets) => Promise<void>;
  disconnect: () => void;
  getBalance: () => Promise<void>;
}

const WalletContext = createContext<WalletContextType | null>(null);

function reducer(state: WalletState, action: Action): WalletState {
  switch (action.type) {
    case 'SET_CONNECTING':
      return { ...state, isConnecting: action.payload };
    case 'SET_CONNECTED':
      return { ...state, isConnected: action.payload };
    case 'SET_ACCOUNT':
      return { ...state, account: action.payload };
    case 'SET_ERROR':
      return { ...state, error: action.payload };
    case 'SET_BALANCE':
      return { ...state, balance: action.payload };
    case 'SET_NETWORK':
      return { ...state, network: action.payload };
    case 'RESET':
      return INITIAL_WALLET_STATE;
    default:
      return state;
  }
}

export function WalletProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(reducer, INITIAL_WALLET_STATE);
  const suiClient = new SuiClient({ url: getFullnodeUrl(WALLET_CONFIG.NETWORK) });

  const connect = useCallback(async (walletType: SupportedWallets) => {
    try {
      dispatch({ type: 'SET_CONNECTING', payload: true });
      dispatch({ type: 'SET_ERROR', payload: null });

      // Check if wallet is installed
      if (typeof window === 'undefined' || !window.suiWallet) {
        throw new Error('Please install Sui Wallet');
      }

      // Request wallet connection
      const response = await window.suiWallet.requestPermissions();
      const accounts = await window.suiWallet.getAccounts();

      if (accounts.length === 0) {
        throw new Error('No accounts found');
      }

      dispatch({ type: 'SET_ACCOUNT', payload: accounts[0] });
      dispatch({ type: 'SET_CONNECTED', payload: true });

      // Get initial balance
      await getBalance();
    } catch (error) {
      console.error('Wallet connection error:', error);
      dispatch({
        type: 'SET_ERROR',
        payload: error instanceof Error ? error.message : 'Failed to connect wallet',
      });
    } finally {
      dispatch({ type: 'SET_CONNECTING', payload: false });
    }
  }, []);

  const disconnect = useCallback(() => {
    dispatch({ type: 'RESET' });
  }, []);

  const getBalance = useCallback(async () => {
    if (!state.account?.address) return;

    try {
      const balance = await suiClient.getBalance({
        owner: state.account.address,
      });

      dispatch({ type: 'SET_BALANCE', payload: balance.totalBalance });
    } catch (error) {
      console.error('Balance fetch error:', error);
    }
  }, [state.account?.address, suiClient]);

  // Auto-refresh balance
  useEffect(() => {
    if (state.isConnected) {
      const interval = setInterval(getBalance, 10000);
      return () => clearInterval(interval);
    }
  }, [state.isConnected, getBalance]);

  return (
    <WalletContext.Provider value={{ state, connect, disconnect, getBalance }}>
      {children}
    </WalletContext.Provider>
  );
}

export function useWallet() {
  const context = useContext(WalletContext);
  if (!context) {
    throw new Error('useWallet must be used within a WalletProvider');
  }
  return context;
} 