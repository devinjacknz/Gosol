'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useWebSocket } from '../hooks/useWebSocket';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

export function Providers({ children }: { children: React.ReactNode }) {
  // 初始化WebSocket连接
  useWebSocket(process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws');

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}  