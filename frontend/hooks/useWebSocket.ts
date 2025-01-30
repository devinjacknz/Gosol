import { useEffect } from 'react';
import { io, Socket } from 'socket.io-client';
import { WSMessageType } from '../types';
import { useTradingStore } from '../store';

const SOCKET_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8000';

export function useWebSocket() {
  const {
    setMarketData,
    setAccountInfo,
    addTrade,
    addRiskAlert,
    setSystemStatus
  } = useTradingStore();

  useEffect(() => {
    const socket: Socket = io(SOCKET_URL, {
      transports: ['websocket'],
      reconnection: true,
      reconnectionAttempts: Infinity,
      reconnectionDelay: 1000,
    });

    socket.on('connect', () => {
      console.log('WebSocket connected');
    });

    socket.on('disconnect', () => {
      console.log('WebSocket disconnected');
    });

    socket.on('message', (message: WSMessageType) => {
      try {
        switch (message.type) {
          case 'market':
            setMarketData(message.data.symbol, message.data);
            break;
          case 'account':
            setAccountInfo(message.data);
            break;
          case 'trade':
            addTrade(message.data);
            break;
          case 'risk':
            addRiskAlert(message.data);
            break;
          case 'system':
            setSystemStatus(message.data);
            break;
          default:
            console.warn('Unknown message type:', message);
        }
      } catch (error) {
        console.error('Error processing WebSocket message:', error);
      }
    });

    // 心跳检测
    const heartbeat = setInterval(() => {
      if (socket.connected) {
        socket.emit('ping');
      }
    }, 30000);

    return () => {
      clearInterval(heartbeat);
      socket.close();
    };
  }, []);
}

export default useWebSocket; 