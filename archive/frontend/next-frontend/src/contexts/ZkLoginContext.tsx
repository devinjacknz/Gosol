'use client';

import React, { createContext, useContext, useReducer, useCallback } from 'react';
import { SuiClient, getFullnodeUrl } from '@mysten/sui.js/client';
import { ZkLoginProvider } from '@mysten/zklogin';
import { ZKLOGIN_CONFIG, ZkLoginState, INITIAL_STATE } from '@/config/zklogin';

type Action =
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'SET_AUTHENTICATED'; payload: boolean }
  | { type: 'SET_ADDRESS'; payload: string }
  | { type: 'RESET' };

const ZkLoginContext = createContext<{
  state: ZkLoginState;
  login: () => Promise<void>;
  logout: () => void;
} | null>(null);

function reducer(state: ZkLoginState, action: Action): ZkLoginState {
  switch (action.type) {
    case 'SET_LOADING':
      return { ...state, isLoading: action.payload };
    case 'SET_ERROR':
      return { ...state, error: action.payload };
    case 'SET_AUTHENTICATED':
      return { ...state, isAuthenticated: action.payload };
    case 'SET_ADDRESS':
      return { ...state, address: action.payload };
    case 'RESET':
      return INITIAL_STATE;
    default:
      return state;
  }
}

export function ZkLoginProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(reducer, INITIAL_STATE);

  const login = useCallback(async () => {
    try {
      dispatch({ type: 'SET_LOADING', payload: true });
      dispatch({ type: 'SET_ERROR', payload: null });

      // Initialize Sui client
      const suiClient = new SuiClient({ url: getFullnodeUrl(ZKLOGIN_CONFIG.NETWORK) });

      // Initialize zkLogin provider
      const provider = new ZkLoginProvider({
        clientId: ZKLOGIN_CONFIG.CLIENT_ID,
        redirectUrl: ZKLOGIN_CONFIG.REDIRECT_URL,
      });

      // Begin login flow
      await provider.login();

      // Get the zkLogin address
      const address = await provider.getAddress();

      dispatch({ type: 'SET_ADDRESS', payload: address });
      dispatch({ type: 'SET_AUTHENTICATED', payload: true });
    } catch (error) {
      console.error('zkLogin error:', error);
      dispatch({ 
        type: 'SET_ERROR', 
        payload: error instanceof Error ? error.message : 'Failed to login' 
      });
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  const logout = useCallback(() => {
    dispatch({ type: 'RESET' });
  }, []);

  return (
    <ZkLoginContext.Provider value={{ state, login, logout }}>
      {children}
    </ZkLoginContext.Provider>
  );
}

export function useZkLogin() {
  const context = useContext(ZkLoginContext);
  if (!context) {
    throw new Error('useZkLogin must be used within a ZkLoginProvider');
  }
  return context;
} 