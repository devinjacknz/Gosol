import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { Provider } from 'react-redux'
import { configureStore } from '@reduxjs/toolkit'
import TradingView from '@/pages/TradingView'
import tradingReducer from '@/store/trading/tradingSlice'

// Mock store
const createTestStore = (initialState = {}) => {
  return configureStore({
    reducer: {
      trading: tradingReducer,
    },
    preloadedState: {
      trading: {
        marketData: {},
        orders: [],
        positions: [],
        selectedSymbol: 'BTC/USDT',
        loading: false,
        error: null,
        ...initialState,
      },
    },
  })
}

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}

describe('TradingView Component', () => {
  it('renders trading view component', () => {
    const store = createTestStore()
    render(
      <Provider store={store}>
        <TradingView />
      </Provider>
    )

    expect(screen.getByText('下单')).toBeInTheDocument()
    expect(screen.getByText('当前订单')).toBeInTheDocument()
    expect(screen.getByText('当前持仓')).toBeInTheDocument()
  })

  it('handles order placement', async () => {
    const store = createTestStore()
    render(
      <Provider store={store}>
        <TradingView />
      </Provider>
    )

    // 填写订单表单
    fireEvent.change(screen.getByPlaceholderText('价格'), {
      target: { value: '50000' },
    })
    fireEvent.change(screen.getByPlaceholderText('数量'), {
      target: { value: '1' },
    })

    // 选择交易对
    const symbolSelect = screen.getByRole('combobox', { name: /symbol/i })
    fireEvent.mouseDown(symbolSelect)
    fireEvent.click(screen.getByText('BTC/USDT'))

    // 选择方向
    const sideSelect = screen.getByRole('combobox', { name: /side/i })
    fireEvent.mouseDown(sideSelect)
    fireEvent.click(screen.getByText('买入'))

    // 提交订单
    fireEvent.click(screen.getByText('下单'))

    // 验证订单是否被提交
    await waitFor(() => {
      const state = store.getState()
      expect(state.trading.orders.length).toBe(1)
    })
  })

  it('displays market data correctly', () => {
    const marketData = {
      'BTC/USDT': {
        klines: [
          {
            time: '2024-02-20T00:00:00Z',
            open: 50000,
            high: 51000,
            low: 49000,
            close: 50500,
            volume: 1000,
          },
        ],
      },
    }

    const store = createTestStore({ marketData })
    render(
      <Provider store={store}>
        <TradingView />
      </Provider>
    )

    // 验证K线图是否正确渲染
    expect(screen.getByTestId('chart-container')).toBeInTheDocument()
  })

  it('handles order cancellation', async () => {
    const orders = [
      {
        id: '1',
        symbol: 'BTC/USDT',
        side: 'buy',
        price: 50000,
        size: 1,
        timestamp: '2024-02-20T00:00:00Z',
      },
    ]

    const store = createTestStore({ orders })
    render(
      <Provider store={store}>
        <TradingView />
      </Provider>
    )

    // 点击取消按钮
    fireEvent.click(screen.getByText('取消'))

    // 验证订单是否被取消
    await waitFor(() => {
      const state = store.getState()
      expect(state.trading.orders.length).toBe(0)
    })
  })

  it('updates when symbol changes', async () => {
    const store = createTestStore()
    render(
      <Provider store={store}>
        <TradingView />
      </Provider>
    )

    // 切换交易对
    const symbolSelect = screen.getByRole('combobox', { name: /symbol/i })
    fireEvent.mouseDown(symbolSelect)
    fireEvent.click(screen.getByText('ETH/USDT'))

    // 验证交易对是否更新
    await waitFor(() => {
      const state = store.getState()
      expect(state.trading.selectedSymbol).toBe('ETH/USDT')
    })
  })
})  