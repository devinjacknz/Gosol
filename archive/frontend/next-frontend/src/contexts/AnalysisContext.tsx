'use client';

import React, { createContext, useContext, useReducer, useCallback } from 'react';
import {
  Indicator,
  BacktestConfig,
  BacktestResult,
  DEFAULT_INDICATORS,
} from '@/config/analysis';
import { useWallet } from './WalletContext';

interface AnalysisState {
  indicators: Indicator[];
  backtests: BacktestConfig[];
  results: BacktestResult[];
  selectedBacktest?: string;
  isLoading: boolean;
  error: string | null;
}

type Action =
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'SET_INDICATORS'; payload: Indicator[] }
  | { type: 'ADD_INDICATOR'; payload: Indicator }
  | { type: 'UPDATE_INDICATOR'; payload: Partial<Indicator> & { id: string } }
  | { type: 'REMOVE_INDICATOR'; payload: string }
  | { type: 'SET_BACKTESTS'; payload: BacktestConfig[] }
  | { type: 'ADD_BACKTEST'; payload: BacktestConfig }
  | { type: 'UPDATE_BACKTEST'; payload: Partial<BacktestConfig> & { id: string } }
  | { type: 'REMOVE_BACKTEST'; payload: string }
  | { type: 'SET_RESULTS'; payload: BacktestResult[] }
  | { type: 'ADD_RESULT'; payload: BacktestResult }
  | { type: 'SELECT_BACKTEST'; payload: string };

const INITIAL_STATE: AnalysisState = {
  indicators: DEFAULT_INDICATORS,
  backtests: [],
  results: [],
  isLoading: false,
  error: null,
};

const AnalysisContext = createContext<{
  state: AnalysisState;
  addIndicator: (indicator: Omit<Indicator, 'id'>) => Promise<void>;
  updateIndicator: (id: string, updates: Partial<Indicator>) => Promise<void>;
  removeIndicator: (id: string) => Promise<void>;
  createBacktest: (config: Omit<BacktestConfig, 'id'>) => Promise<void>;
  updateBacktest: (id: string, updates: Partial<BacktestConfig>) => Promise<void>;
  removeBacktest: (id: string) => Promise<void>;
  runBacktest: (id: string) => Promise<void>;
  fetchIndicatorData: (pair: string, timeframe: string) => Promise<void>;
} | null>(null);

function reducer(state: AnalysisState, action: Action): AnalysisState {
  switch (action.type) {
    case 'SET_LOADING':
      return { ...state, isLoading: action.payload };
    case 'SET_ERROR':
      return { ...state, error: action.payload };
    case 'SET_INDICATORS':
      return { ...state, indicators: action.payload };
    case 'ADD_INDICATOR':
      return { ...state, indicators: [...state.indicators, action.payload] };
    case 'UPDATE_INDICATOR':
      return {
        ...state,
        indicators: state.indicators.map((indicator) =>
          indicator.id === action.payload.id
            ? { ...indicator, ...action.payload }
            : indicator
        ),
      };
    case 'REMOVE_INDICATOR':
      return {
        ...state,
        indicators: state.indicators.filter(
          (indicator) => indicator.id !== action.payload
        ),
      };
    case 'SET_BACKTESTS':
      return { ...state, backtests: action.payload };
    case 'ADD_BACKTEST':
      return { ...state, backtests: [...state.backtests, action.payload] };
    case 'UPDATE_BACKTEST':
      return {
        ...state,
        backtests: state.backtests.map((backtest) =>
          backtest.id === action.payload.id
            ? { ...backtest, ...action.payload }
            : backtest
        ),
      };
    case 'REMOVE_BACKTEST':
      return {
        ...state,
        backtests: state.backtests.filter(
          (backtest) => backtest.id !== action.payload
        ),
      };
    case 'SET_RESULTS':
      return { ...state, results: action.payload };
    case 'ADD_RESULT':
      return { ...state, results: [...state.results, action.payload] };
    case 'SELECT_BACKTEST':
      return { ...state, selectedBacktest: action.payload };
    default:
      return state;
  }
}

export function AnalysisProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(reducer, INITIAL_STATE);
  const { state: walletState } = useWallet();

  const addIndicator = useCallback(
    async (indicator: Omit<Indicator, 'id'>) => {
      try {
        dispatch({ type: 'SET_LOADING', payload: true });

        const response = await fetch('/api/analysis/indicators', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(indicator),
        });

        const data = await response.json();

        if (!response.ok) {
          throw new Error(data.message || 'Failed to add indicator');
        }

        dispatch({ type: 'ADD_INDICATOR', payload: data.indicator });
      } catch (error) {
        dispatch({
          type: 'SET_ERROR',
          payload: error instanceof Error ? error.message : 'Failed to add indicator',
        });
        throw error;
      } finally {
        dispatch({ type: 'SET_LOADING', payload: false });
      }
    },
    []
  );

  const updateIndicator = useCallback(
    async (id: string, updates: Partial<Indicator>) => {
      try {
        dispatch({ type: 'SET_LOADING', payload: true });

        const response = await fetch(`/api/analysis/indicators/${id}`, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(updates),
        });

        const data = await response.json();

        if (!response.ok) {
          throw new Error(data.message || 'Failed to update indicator');
        }

        dispatch({
          type: 'UPDATE_INDICATOR',
          payload: { id, ...updates },
        });
      } catch (error) {
        dispatch({
          type: 'SET_ERROR',
          payload: error instanceof Error ? error.message : 'Failed to update indicator',
        });
        throw error;
      } finally {
        dispatch({ type: 'SET_LOADING', payload: false });
      }
    },
    []
  );

  const removeIndicator = useCallback(async (id: string) => {
    try {
      dispatch({ type: 'SET_LOADING', payload: true });

      const response = await fetch(`/api/analysis/indicators/${id}`, {
        method: 'DELETE',
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || 'Failed to remove indicator');
      }

      dispatch({ type: 'REMOVE_INDICATOR', payload: id });
    } catch (error) {
      dispatch({
        type: 'SET_ERROR',
        payload: error instanceof Error ? error.message : 'Failed to remove indicator',
      });
      throw error;
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  const createBacktest = useCallback(
    async (config: Omit<BacktestConfig, 'id'>) => {
      if (!walletState.account?.address) {
        throw new Error('Wallet not connected');
      }

      try {
        dispatch({ type: 'SET_LOADING', payload: true });

        const response = await fetch('/api/analysis/backtests', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(config),
        });

        const data = await response.json();

        if (!response.ok) {
          throw new Error(data.message || 'Failed to create backtest');
        }

        dispatch({ type: 'ADD_BACKTEST', payload: data.backtest });
      } catch (error) {
        dispatch({
          type: 'SET_ERROR',
          payload: error instanceof Error ? error.message : 'Failed to create backtest',
        });
        throw error;
      } finally {
        dispatch({ type: 'SET_LOADING', payload: false });
      }
    },
    [walletState.account?.address]
  );

  const updateBacktest = useCallback(
    async (id: string, updates: Partial<BacktestConfig>) => {
      try {
        dispatch({ type: 'SET_LOADING', payload: true });

        const response = await fetch(`/api/analysis/backtests/${id}`, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(updates),
        });

        const data = await response.json();

        if (!response.ok) {
          throw new Error(data.message || 'Failed to update backtest');
        }

        dispatch({
          type: 'UPDATE_BACKTEST',
          payload: { id, ...updates },
        });
      } catch (error) {
        dispatch({
          type: 'SET_ERROR',
          payload: error instanceof Error ? error.message : 'Failed to update backtest',
        });
        throw error;
      } finally {
        dispatch({ type: 'SET_LOADING', payload: false });
      }
    },
    []
  );

  const removeBacktest = useCallback(async (id: string) => {
    try {
      dispatch({ type: 'SET_LOADING', payload: true });

      const response = await fetch(`/api/analysis/backtests/${id}`, {
        method: 'DELETE',
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || 'Failed to remove backtest');
      }

      dispatch({ type: 'REMOVE_BACKTEST', payload: id });
    } catch (error) {
      dispatch({
        type: 'SET_ERROR',
        payload: error instanceof Error ? error.message : 'Failed to remove backtest',
      });
      throw error;
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  const runBacktest = useCallback(async (id: string) => {
    try {
      dispatch({ type: 'SET_LOADING', payload: true });

      const response = await fetch(`/api/analysis/backtests/${id}/run`, {
        method: 'POST',
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || 'Failed to run backtest');
      }

      dispatch({ type: 'ADD_RESULT', payload: data.result });
    } catch (error) {
      dispatch({
        type: 'SET_ERROR',
        payload: error instanceof Error ? error.message : 'Failed to run backtest',
      });
      throw error;
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  const fetchIndicatorData = useCallback(
    async (pair: string, timeframe: string) => {
      try {
        dispatch({ type: 'SET_LOADING', payload: true });

        const response = await fetch(
          `/api/analysis/indicators/data?pair=${pair}&timeframe=${timeframe}`
        );

        const data = await response.json();

        if (!response.ok) {
          throw new Error(data.message || 'Failed to fetch indicator data');
        }

        // Update relevant state with indicator data
      } catch (error) {
        dispatch({
          type: 'SET_ERROR',
          payload: error instanceof Error ? error.message : 'Failed to fetch indicator data',
        });
      } finally {
        dispatch({ type: 'SET_LOADING', payload: false });
      }
    },
    []
  );

  return (
    <AnalysisContext.Provider
      value={{
        state,
        addIndicator,
        updateIndicator,
        removeIndicator,
        createBacktest,
        updateBacktest,
        removeBacktest,
        runBacktest,
        fetchIndicatorData,
      }}
    >
      {children}
    </AnalysisContext.Provider>
  );
}

export function useAnalysis() {
  const context = useContext(AnalysisContext);
  if (!context) {
    throw new Error('useAnalysis must be used within an AnalysisProvider');
  }
  return context;
} 