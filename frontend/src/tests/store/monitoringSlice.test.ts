import { describe, it, expect } from 'vitest';
import monitoringReducer, { addAlert, updateMetrics } from '@/store/monitoring/monitoringSlice';

describe('monitoringSlice', () => {
  const initialState = {
    alerts: [],
    metrics: {
      systemStatus: 'healthy',
      cpuUsage: 0,
      memoryUsage: 0,
      wsConnections: 0,
      activeOrders: 0,
    },
  };

  it('should handle initial state', () => {
    expect(monitoringReducer(undefined, { type: 'unknown' })).toEqual(initialState);
  });

  it('should handle updateMetrics', () => {
    const mockMetrics = { cpuUsage: 50, memoryUsage: 60 };
    const state = monitoringReducer(initialState, updateMetrics(mockMetrics));
    expect(state.metrics).toEqual({ ...initialState.metrics, ...mockMetrics });
  });

  it('should handle addAlert', () => {
    const mockAlert = { level: 'warning' as const, message: 'Test alert', source: 'test' };
    const state = monitoringReducer(initialState, addAlert(mockAlert));
    expect(state.alerts[0]).toMatchObject(mockAlert);
    expect(state.alerts[0].timestamp).toBeDefined();
  });
});
