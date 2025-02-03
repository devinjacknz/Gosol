import { createSlice, PayloadAction } from '@reduxjs/toolkit';

interface Alert {
  level: 'info' | 'warning' | 'error';
  message: string;
  source: string;
  timestamp?: number;
}

interface Metrics {
  systemStatus: string;
  cpuUsage: number;
  memoryUsage: number;
  wsConnections: number;
  activeOrders: number;
}

interface MonitoringState {
  alerts: Alert[];
  metrics: Metrics;
  loading: boolean;
  error: string | null;
}

const initialState: MonitoringState = {
  alerts: [],
  metrics: {
    systemStatus: 'healthy',
    cpuUsage: 0,
    memoryUsage: 0,
    wsConnections: 0,
    activeOrders: 0,
  },
  loading: false,
  error: null
};

const monitoringSlice = createSlice({
  name: 'monitoring',
  initialState,
  reducers: {
    addAlert: (state, action: PayloadAction<Alert>) => {
      state.alerts.unshift({
        ...action.payload,
        timestamp: Date.now(),
      });
    },
    updateMetrics: (state, action: PayloadAction<Partial<Metrics>>) => {
      state.metrics = { ...state.metrics, ...action.payload };
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
  },
});

export const { addAlert, updateMetrics } = monitoringSlice.actions;
export default monitoringSlice.reducer;
