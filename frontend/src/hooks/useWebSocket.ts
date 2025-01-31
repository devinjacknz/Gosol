import { useEffect, useCallback, useRef } from 'react';
import { useTradingStore } from '@/store';

import type { MarketDataEntry, SystemStatus } from '@/types/trading';

interface WebSocketMessage {
  type: string;
  data: MarketDataEntry | string;
}

export function useWebSocket(symbol: string): void {
  const { setMarketData, setAccountInfo, setSystemStatus } = useTradingStore();
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>();
  const maxReconnectDelay = 5000;
  const wsUrl = `${process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'}/ws/${symbol}`;

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;
    
    const handleSubscribe = () => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ 
          type: 'subscribe', 
          symbol,
          timestamp: new Date().toISOString()
        }));
      }
    };

    const handleMessage = (event: MessageEvent) => {
      try {
        const message = JSON.parse(event.data) as WebSocketMessage;
        const currentTime = Date.now();
        const timestamp = new Date().toISOString();

        switch (message.type) {
          case 'market_data':
            if (typeof message.data !== 'string') {
              const marketData = message.data;
              setMarketData(symbol, marketData);
              setSystemStatus({
                isConnected: true,
                lastUpdate: timestamp,
                status: 'online',
                dataDelay: marketData.trades?.[0]?.timestamp 
                  ? currentTime - new Date(marketData.trades[0].timestamp).getTime()
                  : 0,
              });
            } else {
              setSystemStatus({
                isConnected: true,
                lastUpdate: timestamp,
                status: 'online',
                dataDelay: 0,
              });
            }
            break;
          case 'subscribed':
            console.log(`Subscribed to ${symbol}`);
            break;
          case 'error':
            console.error(`WebSocket error: ${message.data}`);
            setSystemStatus({
              isConnected: false,
              lastUpdate: new Date().toISOString(),
              status: 'offline',
              message: `WebSocket error: ${message.data}`,
              dataDelay: 0,
            });
            break;
        }
      } catch (error) {
        console.error('WebSocket message parsing error:', error);
      }
    };

    ws.onopen = () => {
      setSystemStatus({
        isConnected: true,
        lastUpdate: new Date().toISOString(),
        status: 'online',
        dataDelay: 0,
      });
      handleSubscribe();
    };

    ws.onmessage = handleMessage;

    ws.onclose = (event) => {
      if (event.wasClean) {
        console.log(`WebSocket closed cleanly, code=${event.code}`);
      } else {
        console.warn('WebSocket connection died');
      }

      setSystemStatus({
        isConnected: false,
        lastUpdate: new Date().toISOString(),
        status: 'offline',
        message: 'WebSocket connection closed',
        dataDelay: 0,
      });

      // Implement reconnection with exponential backoff
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      
      reconnectTimeoutRef.current = setTimeout(() => {
        connect();
      }, Math.min(1000 + Math.random() * 4000, maxReconnectDelay));
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      setSystemStatus({
        isConnected: false,
        lastUpdate: new Date().toISOString(),
        status: 'offline',
        message: 'WebSocket connection error',
        dataDelay: 0,
      });
      
      // Attempt to reconnect after error
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      reconnectTimeoutRef.current = setTimeout(() => {
        connect();
      }, Math.min(1000 + Math.random() * 4000, maxReconnectDelay));
    };
  }, [symbol, setMarketData, setAccountInfo, setSystemStatus]);

  useEffect(() => {
    connect();
    
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [connect]);
}
