import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { monitoringApi } from '@/services/api'

export interface Alert {
  level: 'info' | 'warning' | 'error'
  message: string
  source: string
  timestamp?: string
}

export interface SystemMetrics {
  cpuUsage: number
  memoryUsage: number
  diskUsage: number
  llmRequestCount: number
  avgResponseTime: number
  errorRate: number
  tokenUsage: number
  llmGenerationTime: number
  tokenGenerationRate: number
  apiSuccessRate: number
  apiLatency: number
  systemAvailability: number
  errorCount: number
}

export interface MonitoringState {
  metrics: SystemMetrics | null
  alerts: Alert[]
  loading: boolean
  error: string | null
}

const initialState: MonitoringState = {
  metrics: null,
  alerts: [],
  loading: false,
  error: null,
}

export const fetchMetrics = createAsyncThunk(
  'monitoring/fetchMetrics',
  async () => {
    const response = await monitoringApi.getMetrics()
    return response.data
  }
)

export const fetchAlerts = createAsyncThunk(
  'monitoring/fetchAlerts',
  async () => {
    const response = await monitoringApi.getAlerts()
    return response.data
  }
)

const monitoringSlice = createSlice({
  name: 'monitoring',
  initialState,
  reducers: {
    addAlert: (state, action) => {
      state.alerts.unshift({
        ...action.payload,
        timestamp: new Date().toISOString(),
      })
      // 保持最近100条告警
      if (state.alerts.length > 100) {
        state.alerts.pop()
      }
    },
    clearAlerts: (state) => {
      state.alerts = []
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchMetrics.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchMetrics.fulfilled, (state, action) => {
        state.loading = false
        state.metrics = action.payload
      })
      .addCase(fetchMetrics.rejected, (state, action) => {
        state.loading = false
        state.error = action.error.message || 'Failed to fetch metrics'
      })
      .addCase(fetchAlerts.fulfilled, (state, action) => {
        state.alerts = action.payload
      })
  },
})

export const { addAlert, clearAlerts } = monitoringSlice.actions
export default monitoringSlice.reducer  