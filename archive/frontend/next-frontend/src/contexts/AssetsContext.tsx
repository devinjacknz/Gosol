'use client';

import React, { createContext, useContext, useReducer, useCallback } from 'react';
import { TokenInfo, TokenBalance, Transaction, DEFAULT_TOKENS } from '@/config/tokens';
import { useWallet } from './WalletContext';

interface AssetsState {
  tokens: Record<string, TokenInfo>;
  balances: Record<string, TokenBalance>;
  transactions: Transaction[];
  isLoading: boolean;
  error: string | null;
}

type Action =
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'SET_TOKENS'; payload: Record<string, TokenInfo> }
  | { type: 'SET_BALANCES'; payload: Record<string, TokenBalance> }
  | { type: 'SET_TRANSACTIONS'; payload: Transaction[] }
  | { type: 'ADD_TOKEN'; payload: TokenInfo }
  | { type: 'REMOVE_TOKEN'; payload: string }
  | { type: 'ADD_TRANSACTION'; payload: Transaction }
  | { type: 'UPDATE_TRANSACTION'; payload: { id: string; update: Partial<Transaction> } };

const INITIAL_STATE: AssetsState = {
  tokens: DEFAULT_TOKENS,
  balances: {},
  transactions: [],
  isLoading: false,
  error: null,
};

const AssetsContext = createContext<{
  state: AssetsState;
  addToken: (token: TokenInfo) => Promise<void>;
  removeToken: (address: string) => void;
  refreshBalances: () => Promise<void>;
  sendTransaction: (to: string, amount: string, token: TokenInfo) => Promise<void>;
} | null>(null);

function reducer(state: AssetsState, action: Action): AssetsState {
  switch (action.type) {
    case 'SET_LOADING':
      return { ...state, isLoading: action.payload };
    case 'SET_ERROR':
      return { ...state, error: action.payload };
    case 'SET_TOKENS':
      return { ...state, tokens: action.payload };
    case 'SET_BALANCES':
      return { ...state, balances: action.payload };
    case 'SET_TRANSACTIONS':
      return { ...state, transactions: action.payload };
    case 'ADD_TOKEN':
      return {
        ...state,
        tokens: { ...state.tokens, [action.payload.address]: action.payload },
      };
    case 'REMOVE_TOKEN':
      const { [action.payload]: _, ...remainingTokens } = state.tokens;
      return { ...state, tokens: remainingTokens };
    case 'ADD_TRANSACTION':
      return {
        ...state,
        transactions: [action.payload, ...state.transactions],
      };
    case 'UPDATE_TRANSACTION':
      return {
        ...state,
        transactions: state.transactions.map((tx) =>
          tx.id === action.payload.id ? { ...tx, ...action.payload.update } : tx
        ),
      };
    default:
      return state;
  }
}

export function AssetsProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(reducer, INITIAL_STATE);
  const { state: walletState } = useWallet();

  const addToken = useCallback(async (token: TokenInfo) => {
    try {
      dispatch({ type: 'SET_LOADING', payload: true });
      // Validate token contract
      // Add token to user's token list
      dispatch({ type: 'ADD_TOKEN', payload: token });
      // Fetch initial balance
      await refreshBalances();
    } catch (error) {
      dispatch({
        type: 'SET_ERROR',
        payload: error instanceof Error ? error.message : 'Failed to add token',
      });
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  const removeToken = useCallback((address: string) => {
    dispatch({ type: 'REMOVE_TOKEN', payload: address });
  }, []);

  const refreshBalances = useCallback(async () => {
    if (!walletState.account?.address) return;

    try {
      dispatch({ type: 'SET_LOADING', payload: true });
      const balances: Record<string, TokenBalance> = {};

      // Fetch balances for each token
      for (const token of Object.values(state.tokens)) {
        try {
          const response = await fetch(`/api/tokens/${token.address}/balance?address=${walletState.account.address}`);
          const data = await response.json();
          
          if (response.ok) {
            balances[token.address] = {
              token,
              balance: data.balance,
              value: data.value || 0,
            };
          }
        } catch (error) {
          console.error(`Failed to fetch balance for ${token.symbol}:`, error);
        }
      }

      dispatch({ type: 'SET_BALANCES', payload: balances });
    } catch (error) {
      dispatch({
        type: 'SET_ERROR',
        payload: error instanceof Error ? error.message : 'Failed to refresh balances',
      });
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, [walletState.account?.address, state.tokens]);

  const sendTransaction = useCallback(async (to: string, amount: string, token: TokenInfo) => {
    if (!walletState.account?.address) {
      throw new Error('Wallet not connected');
    }

    try {
      dispatch({ type: 'SET_LOADING', payload: true });

      const tx: Transaction = {
        id: `tx_${Date.now()}`,
        type: 'send',
        status: 'pending',
        timestamp: Date.now(),
        from: walletState.account.address,
        to,
        amount,
        token,
      };

      dispatch({ type: 'ADD_TRANSACTION', payload: tx });

      const response = await fetch('/api/transactions/send', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          to,
          amount,
          tokenAddress: token.address,
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || 'Transaction failed');
      }

      dispatch({
        type: 'UPDATE_TRANSACTION',
        payload: {
          id: tx.id,
          update: {
            status: 'success',
            fee: data.fee,
          },
        },
      });

      // Refresh balances after successful transaction
      await refreshBalances();
    } catch (error) {
      dispatch({
        type: 'SET_ERROR',
        payload: error instanceof Error ? error.message : 'Transaction failed',
      });
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, [walletState.account?.address, refreshBalances]);

  return (
    <AssetsContext.Provider
      value={{
        state,
        addToken,
        removeToken,
        refreshBalances,
        sendTransaction,
      }}
    >
      {children}
    </AssetsContext.Provider>
  );
}

export function useAssets() {
  const context = useContext(AssetsContext);
  if (!context) {
    throw new Error('useAssets must be used within an AssetsProvider');
  }
  return context;
} 