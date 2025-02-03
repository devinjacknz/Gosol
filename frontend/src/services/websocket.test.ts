import { describe, it, expect, beforeEach, vi } from 'vitest';
import { WebSocketService } from './websocket';

describe('WebSocketService', () => {
  let wsService: WebSocketService;
  let mockWebSocket: any;

  beforeEach(() => {
    vi.resetAllMocks();
    mockWebSocket = {
      close: vi.fn(),
      send: vi.fn(),
      readyState: WebSocket.OPEN,
    };
    global.WebSocket = vi.fn().mockImplementation(() => mockWebSocket);
    wsService = new WebSocketService();
  });

  it('connects to WebSocket server', () => {
    wsService.connect();
    expect(WebSocket).toHaveBeenCalledWith(expect.stringContaining('/ws'), ["13"]);
  });

  it('handles connection events', () => {
    const onConnectedMock = vi.fn();
    wsService.on('connected', onConnectedMock);
    wsService.connect();
    mockWebSocket.onopen();
    expect(onConnectedMock).toHaveBeenCalled();
  });

  it('handles message events', () => {
    const mockData = { type: 'market', data: { price: 50000 } };
    const onMessageMock = vi.fn();
    wsService.on('market', onMessageMock);
    wsService.connect();
    mockWebSocket.onmessage({ data: JSON.stringify(mockData) });
    expect(onMessageMock).toHaveBeenCalledWith(mockData.data);
  });

  it('handles disconnection', () => {
    const onDisconnectedMock = vi.fn();
    wsService.on('disconnected', onDisconnectedMock);
    wsService.connect();
    mockWebSocket.onclose();
    expect(onDisconnectedMock).toHaveBeenCalled();
  });

  it('sends messages when connected', () => {
    wsService.connect();
    const message = { type: 'subscribe', channel: 'market:BTC-USD' };
    wsService.send(message);
    expect(mockWebSocket.send).toHaveBeenCalledWith(JSON.stringify(message));
  });

  it('subscribes to channels', () => {
    wsService.connect();
    wsService.subscribe('market:BTC-USD');
    expect(mockWebSocket.send).toHaveBeenCalledWith(JSON.stringify({
      type: 'subscribe',
      channel: 'market:BTC-USD'
    }));
  });

  it('unsubscribes from channels', () => {
    wsService.connect();
    wsService.unsubscribe('market:BTC-USD');
    expect(mockWebSocket.send).toHaveBeenCalledWith(JSON.stringify({
      type: 'unsubscribe',
      channel: 'market:BTC-USD'
    }));
  });

  it('handles heartbeat', () => {
    vi.useFakeTimers();
    wsService.connect();
    vi.advanceTimersByTime(30000);
    expect(mockWebSocket.send).toHaveBeenCalledWith(JSON.stringify({ type: 'ping' }));
    vi.useRealTimers();
  });
});
