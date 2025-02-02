import { ReactNode } from 'react';

export interface WebSocketStatus {
  status: 'online' | 'offline' | 'error';
  icon: ReactNode;
  color: string;
}

export interface AlertConfig {
  type: 'info' | 'warning' | 'error';
  color: string;
  text: string;
}
