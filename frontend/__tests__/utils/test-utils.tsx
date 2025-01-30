import React from 'react';
import { render as rtlRender } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { ThemeProvider } from '@mui/material/styles';
import { theme } from '@/utils/theme';
import { rootReducer } from '@/store/rootReducer';

function render(
  ui: React.ReactElement,
  {
    preloadedState = {},
    store = configureStore({
      reducer: rootReducer,
      preloadedState,
    }),
    ...renderOptions
  } = {}
) {
  function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <Provider store={store}>
        <ThemeProvider theme={theme}>
          {children}
        </ThemeProvider>
      </Provider>
    );
  }
  return rtlRender(ui, { wrapper: Wrapper, ...renderOptions });
}

// 重新导出所有testing-library的工具
export * from '@testing-library/react';
export { render };

// 创建模拟响应数据
export const mockMarketData = {
  symbol: 'BTC/USDT',
  price: 50000,
  volume: 1000,
  change: 2.5,
  high: 51000,
  low: 49000,
};

export const mockOrderBook = {
  bids: [
    [49900, 1.5],
    [49800, 2.0],
    [49700, 2.5],
  ],
  asks: [
    [50100, 1.0],
    [50200, 1.5],
    [50300, 2.0],
  ],
};

export const mockTrades = [
  {
    id: '1',
    symbol: 'BTC/USDT',
    side: 'buy',
    price: 50000,
    amount: 1.0,
    timestamp: new Date().toISOString(),
  },
  {
    id: '2',
    symbol: 'BTC/USDT',
    side: 'sell',
    price: 50100,
    amount: 0.5,
    timestamp: new Date().toISOString(),
  },
];

// 模拟API响应
export const mockApiResponse = {
  success: true,
  data: null,
  error: null,
};

// 模拟错误响应
export const mockApiError = {
  success: false,
  data: null,
  error: 'Error message',
};

// 等待组件更新的工具函数
export const waitForComponentUpdate = () => new Promise(resolve => setTimeout(resolve, 0)); 