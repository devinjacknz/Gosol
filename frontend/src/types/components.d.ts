import { ReactNode } from 'react';

export interface MetricCardProps {
  title: string;
  value: number | string;
  precision?: number;
  suffix?: string;
  loading?: boolean;
  style?: React.CSSProperties;
}

export interface SystemMetrics {
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  llmRequestCount: number;
  activeConnections: number;
  responseTime: number;
  errorRate: number;
  throughput: number;
  queueLength: number;
  modelLoadTime: number;
  inferenceTime: number;
  batchSize: number;
}

export interface Alert {
  id: string;
  type: 'info' | 'warning' | 'error';
  message: string;
  timestamp: string;
}

export interface TradingState {
  loading: boolean;
  error: string | null;
  marketData: Record<string, any>;
  orders: any[];
  positions: any[];
  selectedSymbol: string;
}
