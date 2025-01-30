'use client';

import React, { createContext, useContext, useReducer, useCallback } from 'react';
import {
  TradingPair,
  Order,
  Position,
  Strategy,
  OrderType,
  OrderSide,
  ORDER_TYPES,
  ORDER_SIDES,
} from '@/config/trading';
import { useWallet } from './WalletContext';

interface TradingState {
  pairs: TradingPair[];
  orders: Order[];
  positions: Position[];
  strategies: Strategy[];
  selectedPair?: TradingPair;
  isLoading: boolean;
  error: string | null;
}

type Action =
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'SET_PAIRS'; payload: TradingPair[] }
  | { type: 'SET_ORDERS'; payload: Order[] }
  | { type: 'SET_POSITIONS'; payload: Position[] }
  | { type: 'SET_STRATEGIES'; payload: Strategy[] }
  | { type: 'SELECT_PAIR'; payload: TradingPair }
  | { type: 'ADD_ORDER'; payload: Order }
  | { type: 'UPDATE_ORDER'; payload: Partial<Order> & { id: string } }
  | { type: 'UPDATE_POSITION'; payload: Partial<Position> & { id: string } }
  | { type: 'UPDATE_STRATEGY'; payload: Partial<Strategy> & { id: string } };

const INITIAL_STATE: TradingState = {
  pairs: [],
  orders: [],
  positions: [],
  strategies: [],
  isLoading: false,
  error: null,
};

const TradingContext = createContext<{
  state: TradingState;
  placeOrder: (params: PlaceOrderParams) => Promise<void>;
  cancelOrder: (orderId: string) => Promise<void>;
  createStrategy: (strategy: Omit<Strategy, 'id' | 'userId'>) => Promise<void>;
  updateStrategy: (strategyId: string, updates: Partial<Strategy>) => Promise<void>;
  fetchMarketData: (pairId: string) => Promise<void>;
  selectPair: (pair: TradingPair) => void;
} | null>(null);

interface PlaceOrderParams {
  pair: TradingPair;
  type: OrderType[keyof OrderType];
  side: OrderSide[keyof OrderSide];
  price?: string;
  amount: string;
  stopPrice?: string;
  trailingDistance?: string;
}

function reducer(state: TradingState, action: Action): TradingState {
  switch (action.type) {
    case 'SET_LOADING':
      return { ...state, isLoading: action.payload };
    case 'SET_ERROR':
      return { ...state, error: action.payload };
    case 'SET_PAIRS':
      return { ...state, pairs: action.payload };
    case 'SET_ORDERS':
      return { ...state, orders: action.payload };
    case 'SET_POSITIONS':
      return { ...state, positions: action.payload };
    case 'SET_STRATEGIES':
      return { ...state, strategies: action.payload };
    case 'SELECT_PAIR':
      return { ...state, selectedPair: action.payload };
    case 'ADD_ORDER':
      return { ...state, orders: [action.payload, ...state.orders] };
    case 'UPDATE_ORDER':
      return {
        ...state,
        orders: state.orders.map((order) =>
          order.id === action.payload.id ? { ...order, ...action.payload } : order
        ),
      };
    case 'UPDATE_POSITION':
      return {
        ...state,
        positions: state.positions.map((position) =>
          position.id === action.payload.id
            ? { ...position, ...action.payload }
            : position
        ),
      };
    case 'UPDATE_STRATEGY':
      return {
        ...state,
        strategies: state.strategies.map((strategy) =>
          strategy.id === action.payload.id
            ? { ...strategy, ...action.payload }
            : strategy
        ),
      };
    default:
      return state;
  }
}

export function TradingProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(reducer, INITIAL_STATE);
  const { state: walletState } = useWallet();

  const placeOrder = useCallback(
    async ({ pair, type, side, price, amount, stopPrice, trailingDistance }: PlaceOrderParams) => {
      if (!walletState.account?.address) {
        throw new Error('Wallet not connected');
      }

      try {
        dispatch({ type: 'SET_LOADING', payload: true });

        const response = await fetch('/api/trading/orders', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            pair: pair.id,
            type,
            side,
            price,
            amount,
            stopPrice,
            trailingDistance,
          }),
        });

        const data = await response.json();

        if (!response.ok) {
          throw new Error(data.message || 'Failed to place order');
        }

        dispatch({ type: 'ADD_ORDER', payload: data.order });
      } catch (error) {
        dispatch({
          type: 'SET_ERROR',
          payload: error instanceof Error ? error.message : 'Failed to place order',
        });
        throw error;
      } finally {
        dispatch({ type: 'SET_LOADING', payload: false });
      }
    },
    [walletState.account?.address]
  );

  const cancelOrder = useCallback(async (orderId: string) => {
    try {
      dispatch({ type: 'SET_LOADING', payload: true });

      const response = await fetch(`/api/trading/orders/${orderId}`, {
        method: 'DELETE',
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || 'Failed to cancel order');
      }

      dispatch({
        type: 'UPDATE_ORDER',
        payload: { id: orderId, status: 'canceled' },
      });
    } catch (error) {
      dispatch({
        type: 'SET_ERROR',
        payload: error instanceof Error ? error.message : 'Failed to cancel order',
      });
      throw error;
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  const createStrategy = useCallback(
    async (strategy: Omit<Strategy, 'id' | 'userId'>) => {
      if (!walletState.account?.address) {
        throw new Error('Wallet not connected');
      }

      try {
        dispatch({ type: 'SET_LOADING', payload: true });

        const response = await fetch('/api/trading/strategies', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(strategy),
        });

        const data = await response.json();

        if (!response.ok) {
          throw new Error(data.message || 'Failed to create strategy');
        }

        dispatch({
          type: 'SET_STRATEGIES',
          payload: [...state.strategies, data.strategy],
        });
      } catch (error) {
        dispatch({
          type: 'SET_ERROR',
          payload: error instanceof Error ? error.message : 'Failed to create strategy',
        });
        throw error;
      } finally {
        dispatch({ type: 'SET_LOADING', payload: false });
      }
    },
    [walletState.account?.address, state.strategies]
  );

  const updateStrategy = useCallback(
    async (strategyId: string, updates: Partial<Strategy>) => {
      try {
        dispatch({ type: 'SET_LOADING', payload: true });

        const response = await fetch(`/api/trading/strategies/${strategyId}`, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(updates),
        });

        const data = await response.json();

        if (!response.ok) {
          throw new Error(data.message || 'Failed to update strategy');
        }

        dispatch({
          type: 'UPDATE_STRATEGY',
          payload: { id: strategyId, ...updates },
        });
      } catch (error) {
        dispatch({
          type: 'SET_ERROR',
          payload: error instanceof Error ? error.message : 'Failed to update strategy',
        });
        throw error;
      } finally {
        dispatch({ type: 'SET_LOADING', payload: false });
      }
    },
    []
  );

  const fetchMarketData = useCallback(async (pairId: string) => {
    try {
      dispatch({ type: 'SET_LOADING', payload: true });

      const response = await fetch(`/api/trading/market-data/${pairId}`);
      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || 'Failed to fetch market data');
      }

      // Update relevant state with market data
      // This could include order book, recent trades, etc.
    } catch (error) {
      dispatch({
        type: 'SET_ERROR',
        payload: error instanceof Error ? error.message : 'Failed to fetch market data',
      });
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  const selectPair = useCallback((pair: TradingPair) => {
    dispatch({ type: 'SELECT_PAIR', payload: pair });
  }, []);

  return (
    <TradingContext.Provider
      value={{
        state,
        placeOrder,
        cancelOrder,
        createStrategy,
        updateStrategy,
        fetchMarketData,
        selectPair,
      }}
    >
      {children}
    </TradingContext.Provider>
  );
}

export function useTrading() {
  const context = useContext(TradingContext);
  if (!context) {
    throw new Error('useTrading must be used within a TradingProvider');
  }
  return context;
} 