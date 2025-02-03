import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { WebSocketService } from '../services/websocket';

describe('WebSocketService', () => {
  let wsService: WebSocketService;
  let mockWebSocket: any;

  beforeEach(() => {
    mockWebSocket = {
      send: vi.fn(),
      close: vi.fn(),
      readyState: WebSocket.OPEN,
      onopen: null,
      onmessage: null,
      onclose: null,
      onerror: null,
    };

    const MockWebSocket = vi.fn(() => mockWebSocket) as any;
    MockWebSocket.CONNECTING = 0;
    MockWebSocket.OPEN = 1;
    MockWebSocket.CLOSING = 2;
    MockWebSocket.CLOSED = 3;
    global.WebSocket = MockWebSocket;
    wsService = new WebSocketService();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should connect to WebSocket server', () => {
    wsService.connect();
    expect(global.WebSocket).toHaveBeenCalledWith('ws://localhost:8081/ws', ['13']);
  });

  it('should send messages when connected', () => {
    wsService.connect();
    wsService.send({ type: 'test', data: { message: 'test' } });
    expect(mockWebSocket.send).toHaveBeenCalledWith(JSON.stringify({ type: 'test', data: { message: 'test' } }));
  });

  it('should handle ping/pong messages', () => {
    wsService.connect();
    const message = { type: 'ping' };
    wsService.send(message);
    expect(mockWebSocket.send).toHaveBeenCalledWith(JSON.stringify(message));
  });

  it('should handle subscription messages', () => {
    wsService.connect();
    wsService.subscribe('marketData');
    expect(mockWebSocket.send).toHaveBeenCalledWith(JSON.stringify({ type: 'subscribe', channel: 'marketData' }));
  });
});
