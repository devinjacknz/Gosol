import { configureStore } from '@reduxjs/toolkit'
import tradingReducer from './trading/tradingSlice'
import analysisReducer from './analysis/analysisSlice'
import monitoringReducer from './monitoring/monitoringSlice'

export const store = configureStore({
  reducer: {
    trading: tradingReducer,
    analysis: analysisReducer,
    monitoring: monitoringReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: false,
    }),
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch 