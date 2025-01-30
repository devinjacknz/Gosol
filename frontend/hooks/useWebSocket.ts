import { useEffect, useRef, useState, useCallback } from 'react';

interface WebSocketHookResult<T> {
  data: T | null;
  isConnected: boolean;
  error: Error | null;
  send: (message: any) => void;
  ws: WebSocket | null;
}

export function useWebSocket<T = any>(url: string): WebSocketHookResult<T> {
  const [data, setData] = useState<T | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>();

  const connect = useCallback(() => {
    try {
      const ws = new WebSocket(url);

      ws.onopen = () => {
        setIsConnected(true);
        setError(null);
      };

      ws.onmessage = (event) => {
        try {
          const parsed = JSON.parse(event.data);
          setData(parsed);
        } catch (e) {
          setError(new Error('Failed to parse message'));
        }
      };

      ws.onclose = () => {
        setIsConnected(false);
        // 尝试重连
        reconnectTimeoutRef.current = setTimeout(() => {
          connect();
        }, 5000);
      };

      ws.onerror = (event) => {
        setError(new Error('WebSocket error'));
        ws.close();
      };

      wsRef.current = ws;
    } catch (e) {
      setError(e as Error);
    }
  }, [url]);

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

  const send = useCallback((message: any) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    } else {
      setError(new Error('WebSocket is not connected'));
    }
  }, []);

  return {
    data,
    isConnected,
    error,
    send,
    ws: wsRef.current,
  };
}

export default useWebSocket; 