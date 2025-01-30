import { create } from 'zustand';

interface MarketData {
  price: number;
  volume: number;
  timestamp: number;
}

interface WebSocketStore {
  marketData: Record<string, MarketData>;
  isConnected: boolean;
  connect: () => void;
  disconnect: () => void;
  updateMarketData: (symbol: string, data: MarketData) => void;
}

export const useWebSocketStore = create<WebSocketStore>((set) => ({
  marketData: {},
  isConnected: false,
  connect: () => {
    if (typeof window === 'undefined') return; // Skip on server-side

    const ws = new WebSocket('ws://localhost:8080/ws/market');
    
    ws.onopen = () => {
      set({ isConnected: true });
      console.log('WebSocket connected');
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        set((state) => ({
          marketData: {
            ...state.marketData,
            [data.symbol]: {
              price: data.price,
              volume: data.volume,
              timestamp: data.timestamp,
            },
          },
        }));
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    ws.onclose = () => {
      set({ isConnected: false });
      console.log('WebSocket disconnected');
      // Attempt to reconnect after 5 seconds
      setTimeout(() => {
        set((state) => {
          if (!state.isConnected) {
            state.connect();
          }
          return state;
        });
      }, 5000);
    };
  },
  disconnect: () => {
    set({ isConnected: false });
  },
  updateMarketData: (symbol, data) => {
    set((state) => ({
      marketData: {
        ...state.marketData,
        [symbol]: data,
      },
    }));
  },
})); 