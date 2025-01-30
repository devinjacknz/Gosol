import { renderHook, act } from '@testing-library/react';
import { useWebSocket } from '@/hooks/useWebSocket';

// 模拟WebSocket
class MockWebSocket {
  onopen: (() => void) | null = null;
  onclose: (() => void) | null = null;
  onmessage: ((event: any) => void) | null = null;
  onerror: ((error: any) => void) | null = null;
  readyState: number = WebSocket.CONNECTING;
  send = jest.fn();
  close = jest.fn();

  constructor(url: string) {
    setTimeout(() => {
      this.readyState = WebSocket.OPEN;
      if (this.onopen) this.onopen();
    }, 0);
  }
}

// 替换全局WebSocket
(global as any).WebSocket = MockWebSocket;

describe('useWebSocket Hook', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('connects to WebSocket successfully', async () => {
    const { result } = renderHook(() => useWebSocket('wss://test.com'));

    // 初始状态
    expect(result.current.isConnected).toBe(false);
    expect(result.current.error).toBeNull();

    // 等待连接建立
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 0));
    });

    expect(result.current.isConnected).toBe(true);
  });

  it('handles incoming messages', async () => {
    const { result } = renderHook(() => useWebSocket('wss://test.com'));

    // 等待连接建立
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 0));
    });

    // 模拟接收消息
    const mockMessage = { type: 'ticker', data: { price: 50000 } };
    act(() => {
      const ws = result.current.ws as any;
      ws.onmessage({ data: JSON.stringify(mockMessage) });
    });

    expect(result.current.data).toEqual(mockMessage);
  });

  it('handles connection errors', async () => {
    const { result } = renderHook(() => useWebSocket('wss://test.com'));

    act(() => {
      const ws = result.current.ws as any;
      ws.onerror(new Error('Connection failed'));
    });

    expect(result.current.error).toBeTruthy();
    expect(result.current.isConnected).toBe(false);
  });

  it('handles disconnection', async () => {
    const { result } = renderHook(() => useWebSocket('wss://test.com'));

    // 等待连接建立
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 0));
    });

    // 模拟断开连接
    act(() => {
      const ws = result.current.ws as any;
      ws.onclose();
    });

    expect(result.current.isConnected).toBe(false);
  });

  it('sends messages correctly', async () => {
    const { result } = renderHook(() => useWebSocket('wss://test.com'));

    // 等待连接建立
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 0));
    });

    // 发送消息
    const message = { type: 'subscribe', symbol: 'BTC/USDT' };
    act(() => {
      result.current.send(message);
    });

    const ws = result.current.ws as any;
    expect(ws.send).toHaveBeenCalledWith(JSON.stringify(message));
  });

  it('reconnects on connection loss', async () => {
    jest.useFakeTimers();

    const { result } = renderHook(() => useWebSocket('wss://test.com'));

    // 等待连接建立
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 0));
    });

    // 模拟断开连接
    act(() => {
      const ws = result.current.ws as any;
      ws.onclose();
    });

    // 前进重连间隔时间
    act(() => {
      jest.advanceTimersByTime(5000);
    });

    expect(result.current.isConnected).toBe(true);

    jest.useRealTimers();
  });

  it('cleans up on unmount', () => {
    const { result, unmount } = renderHook(() => useWebSocket('wss://test.com'));

    const ws = result.current.ws as any;
    unmount();

    expect(ws.close).toHaveBeenCalled();
  });
}); 