import React, { useEffect } from 'react';
import { MarketData } from '../types';

export enum WebSocketMessageType {
  MARKET_DATA = 'market_data',
  TRADE_UPDATE = 'trade_update',
  SYSTEM_ALERT = 'system_alert',
}

interface WebSocketMessage {
  type: WebSocketMessageType;
  data: any;
}

interface WebSocketConfig {
  reconnectAttempts?: number;
  reconnectInterval?: number;
  onMessage?: (message: WebSocketMessage) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: Event) => void;
}

export class WebSocketManager {
  private ws: WebSocket | null = null;
  private messageHandlers: ((data: any) => void)[] = [];
  private stateChangeHandlers: ((connected: boolean) => void)[] = [];
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private baseUrl: string;

  constructor() {
    this.baseUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/ws`;
  }

  connect(token: string) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    this.ws = new WebSocket(`${this.baseUrl}?token=${token}`);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.notifyStateChange(true);
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.notifyStateChange(false);
      this.attemptReconnect(token);
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        this.messageHandlers.forEach(handler => handler(data));
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err);
      }
    };
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  onMessage(handler: (data: any) => void) {
    this.messageHandlers.push(handler);
    return () => {
      const index = this.messageHandlers.indexOf(handler);
      if (index !== -1) {
        this.messageHandlers.splice(index, 1);
      }
    };
  }

  onStateChange(handler: (connected: boolean) => void) {
    this.stateChangeHandlers.push(handler);
    return () => {
      const index = this.stateChangeHandlers.indexOf(handler);
      if (index !== -1) {
        this.stateChangeHandlers.splice(index, 1);
      }
    };
  }

  private notifyStateChange(connected: boolean) {
    this.stateChangeHandlers.forEach(handler => handler(connected));
  }

  private attemptReconnect(token: string) {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }

    this.reconnectTimer = setTimeout(() => {
      console.log('Attempting to reconnect...');
      this.connect(token);
    }, 5000);
  }
}

export const wsManager = new WebSocketManager();
export default wsManager;

// Helper hooks for React components
export function useMarketDataSubscription(tokenAddress: string, onData: (data: MarketData) => void) {
  useEffect(() => {
    const channel = `market_data:${tokenAddress}`;
    
    const handler = (data: MarketData) => {
      onData(data);
    };

    wsManager.onMessage(handler);
    wsManager.connect(tokenAddress);

    return () => {
      wsManager.onMessage(handler);
    };
  }, [tokenAddress, onData]);
}
