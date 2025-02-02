import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { tradingApi } from '@/services/api'

export interface TradingState {
  marketData: Record<string, any>
  orders: any[]
  positions: any[]
  selectedSymbol: string
  loading: boolean
  error: string | null
}

const initialState: TradingState = {
  marketData: {},
  orders: [],
  positions: [],
  selectedSymbol: 'BTC/USDT',
  loading: false,
  error: null,
}

export const placeOrder = createAsyncThunk(
  'trading/placeOrder',
  async (order: any) => {
    const response = await tradingApi.placeOrder(order)
    return response
  }
)

export const cancelOrder = createAsyncThunk(
  'trading/cancelOrder',
  async (orderId: string) => {
    await tradingApi.cancelOrder(orderId)
    return orderId
  }
)

export const fetchOrders = createAsyncThunk(
  'trading/fetchOrders',
  async () => {
    const response = await tradingApi.getOrders({})
    return response
  }
)

export const fetchPositions = createAsyncThunk(
  'trading/fetchPositions',
  async () => {
    const response = await tradingApi.getPositions()
    return response
  }
)

const tradingSlice = createSlice({
  name: 'trading',
  initialState,
  reducers: {
    updateMarketData: (state, action) => {
      const { symbol, data } = action.payload
      state.marketData[symbol] = {
        ...state.marketData[symbol],
        ...data,
      }
    },
    setSelectedSymbol: (state, action) => {
      state.selectedSymbol = action.payload
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(placeOrder.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(placeOrder.fulfilled, (state, action) => {
        state.loading = false
        state.orders.unshift(action.payload)
      })
      .addCase(placeOrder.rejected, (state, action) => {
        state.loading = false
        state.error = action.error.message || 'Failed to place order'
      })
      .addCase(cancelOrder.fulfilled, (state, action) => {
        state.orders = state.orders.filter(
          (order) => order.id !== action.payload
        )
      })
      .addCase(fetchOrders.fulfilled, (state, action) => {
        state.orders = action.payload.data
      })
      .addCase(fetchPositions.fulfilled, (state, action) => {
        state.positions = action.payload.data
      })
  },
})

export const { updateMarketData, setSelectedSymbol } = tradingSlice.actions
export default tradingSlice.reducer  