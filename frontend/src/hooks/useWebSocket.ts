import { useEffect } from 'react';
import { useTradingStore } from '@/store';

export function useWebSocket(wsUrl: string): void {
  const { setMarketData, setAccountInfo, setSystemStatus } = useTradingStore();

  useEffect(() => {
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      setSystemStatus({
        isConnected: true,
        lastUpdate: new Date().toISOString(),
        status: 'online',
        dataDelay: 0,
      });
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        
        switch (data.type) {
          case 'market_data':
            setMarketData(data.symbol, data.data);
            break;
          case 'account_info':
            setAccountInfo(data.data);
            break;
          case 'system_status':
            setSystemStatus(data.data);
            break;
        }
      } catch (error) {
        console.error('WebSocket message parsing error:', error);
      }
    };

    ws.onclose = () => {
      setSystemStatus({
        isConnected: false,
        lastUpdate: new Date().toISOString(),
        status: 'offline',
        message: 'WebSocket connection closed',
        dataDelay: 0,
      });
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
    };

    return () => {
      ws.close();
    };
  }, [wsUrl, setMarketData, setAccountInfo, setSystemStatus]);
}
