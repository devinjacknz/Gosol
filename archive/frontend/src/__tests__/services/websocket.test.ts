import { WebSocketManager, WebSocketMessageType } from '../../services/websocket';

describe('WebSocketManager', () => {
  let wsManager: WebSocketManager;
  let mockWebSocket: any;

  beforeEach(() => {
    // Mock WebSocket
    mockWebSocket = {
      readyState: WebSocket.CLOSED,
      close: jest.fn(),
      send: jest.fn(),
      onopen: null,
      onclose: null,
      onerror: null,
      onmessage: null,
    };

    // Mock WebSocket constructor
    (global as any).WebSocket = jest.fn(() => mockWebSocket);

    wsManager = new WebSocketManager();
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('Connection Management', () => {
    it('establishes connection successfully', () => {
      wsManager.connect('test-token');
      expect(WebSocket).toHaveBeenCalledWith(expect.stringContaining('test-token'));
      
      // Simulate successful connection
      mockWebSocket.onopen();
      expect(mockWebSocket.onopen).toBeDefined();
    });

    it('handles disconnection', () => {
      wsManager.connect('test-token');
      mockWebSocket.onclose();
      expect(mockWebSocket.onclose).toBeDefined();
    });

    it('attempts reconnection after disconnect', () => {
      jest.useFakeTimers();
      wsManager.connect('test-token');
      mockWebSocket.onclose();
      
      jest.advanceTimersByTime(5000);
      expect(WebSocket).toHaveBeenCalledTimes(2);
      
      jest.useRealTimers();
    });

    it('disconnects properly', () => {
      wsManager.connect('test-token');
      wsManager.disconnect();
      expect(mockWebSocket.close).toHaveBeenCalled();
    });
  });

  describe('Message Handling', () => {
    it('processes market data messages', () => {
      const mockHandler = jest.fn();
      wsManager.onMessage(mockHandler);
      wsManager.connect('test-token');

      const mockData = {
        type: WebSocketMessageType.MARKET_DATA,
        data: { price: 100, volume: 1000 }
      };

      mockWebSocket.onmessage({ data: JSON.stringify(mockData) });
      expect(mockHandler).toHaveBeenCalledWith(mockData);
    });

    it('processes trade update messages', () => {
      const mockHandler = jest.fn();
      wsManager.onMessage(mockHandler);
      wsManager.connect('test-token');

      const mockData = {
        type: WebSocketMessageType.TRADE_UPDATE,
        data: { id: '123', type: 'buy', amount: 100 }
      };

      mockWebSocket.onmessage({ data: JSON.stringify(mockData) });
      expect(mockHandler).toHaveBeenCalledWith(mockData);
    });

    it('handles malformed messages gracefully', () => {
      const mockHandler = jest.fn();
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      
      wsManager.onMessage(mockHandler);
      wsManager.connect('test-token');
      
      mockWebSocket.onmessage({ data: 'invalid-json' });
      
      expect(consoleSpy).toHaveBeenCalled();
      expect(mockHandler).not.toHaveBeenCalled();
      
      consoleSpy.mockRestore();
    });
  });

  describe('State Change Notifications', () => {
    it('notifies state change handlers on connect', () => {
      const mockHandler = jest.fn();
      wsManager.onStateChange(mockHandler);
      wsManager.connect('test-token');
      
      mockWebSocket.onopen();
      expect(mockHandler).toHaveBeenCalledWith(true);
    });

    it('notifies state change handlers on disconnect', () => {
      const mockHandler = jest.fn();
      wsManager.onStateChange(mockHandler);
      wsManager.connect('test-token');
      
      mockWebSocket.onclose();
      expect(mockHandler).toHaveBeenCalledWith(false);
    });

    it('removes state change handler correctly', () => {
      const mockHandler = jest.fn();
      const removeHandler = wsManager.onStateChange(mockHandler);
      
      removeHandler();
      wsManager.connect('test-token');
      mockWebSocket.onopen();
      
      expect(mockHandler).not.toHaveBeenCalled();
    });
  });

  describe('Error Handling', () => {
    it('handles connection errors', () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      wsManager.connect('test-token');
      
      mockWebSocket.onerror(new Error('Connection failed'));
      expect(consoleSpy).toHaveBeenCalled();
      
      consoleSpy.mockRestore();
    });

    it('handles message errors', () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      wsManager.connect('test-token');
      
      mockWebSocket.onmessage({ data: null });
      expect(consoleSpy).toHaveBeenCalled();
      
      consoleSpy.mockRestore();
    });
  });
}); 