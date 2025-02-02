import { describe, it, expect, beforeEach, vi } from 'vitest'
import { wsService } from './websocket'
import { store } from '@/store'
import { updateMarketData } from '@/store/trading/tradingSlice'
import { addAlert } from '@/store/monitoring/monitoringSlice'

const mockDispatch = vi.fn();

vi.mock('@/store', () => ({
  store: {
    dispatch: mockDispatch,
  },
}));

describe('WebSocket Service', () => {
  let mockWebSocket: Partial<WebSocket>
  let mockWebSocketCtor: ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.resetAllMocks()
    
    mockWebSocket = {
      send: vi.fn(),
      close: vi.fn(),
      readyState: WebSocket.OPEN,
      onopen: null,
      onclose: null,
      onmessage: null,
      onerror: null,
    }

    mockWebSocketCtor = vi.fn().mockImplementation(() => mockWebSocket)
    vi.stubGlobal('WebSocket', mockWebSocketCtor)

    wsService.disconnect()
  })

  describe('Connection Management', () => {
    it('initializes connection', () => {
      wsService.initialize()

      expect(WebSocket).toHaveBeenCalledWith(expect.stringContaining('/ws/'))
    })

    it('handles connection success', () => {
      wsService.initialize()

      const ws = mockWebSocketCtor.mock.results[0].value as WebSocket
      ws.onopen?.({} as Event)

      expect(wsService['reconnectAttempts']).toBe(0)
    })

    it('handles disconnection', () => {
      wsService.initialize()

      const ws = mockWebSocketCtor.mock.results[0].value as WebSocket
      ws.onclose?.({ reason: 'test close' } as CloseEvent)

      expect(mockDispatch).toHaveBeenCalledWith(
        addAlert({
          level: 'warning',
          message: expect.stringContaining('disconnected'),
          source: 'websocket',
        })
      )
    })

    it('handles reconnection attempts', () => {
      wsService.initialize()
      const ws = mockWebSocketCtor.mock.results[0].value as WebSocket
      ws.onclose?.({ reason: 'test close' } as CloseEvent)
      expect(wsService['reconnectAttempts']).toBe(1)
    })
  })

  describe('Market Data Handling', () => {
    it('subscribes to market data', () => {
      wsService.initialize()
      wsService.subscribeMarketData('BTC/USDT')

      expect(mockWebSocket.send).toHaveBeenCalledWith(
        JSON.stringify({
          type: 'subscribe',
          symbol: 'BTC-USDT'
        })
      )
    })

    it('unsubscribes from market data', () => {
      wsService.initialize()
      wsService.unsubscribeMarketData('BTC/USDT')

      expect(mockWebSocket.send).toHaveBeenCalledWith(
        JSON.stringify({
          type: 'unsubscribe',
          symbol: 'BTC-USDT'
        })
      )
    })

    it('handles market data updates', () => {
      wsService.initialize()

      const ws = mockWebSocketCtor.mock.results[0].value as WebSocket
      const mockData = {
        type: 'market_data',
        data: {
          symbol: 'BTC-USDT',
          price: 50000,
        }
      }
      ws.onmessage?.({ data: JSON.stringify(mockData) } as MessageEvent)

      expect(mockDispatch).toHaveBeenCalledWith(updateMarketData(mockData.data))
    })
  })

  describe('Order Management', () => {
    it('sends order', () => {
      wsService.initialize()

      const order = {
        symbol: 'BTC/USDT',
        side: 'buy',
        price: 50000,
        size: 1,
      }

      wsService.sendOrder(order)

      expect(mockWebSocket.send).toHaveBeenCalledWith(
        JSON.stringify({
          type: 'placeOrder',
          data: order
        })
      )
    })

    it('cancels order', () => {
      wsService.initialize()
      wsService.cancelOrder('123')

      expect(mockWebSocket.send).toHaveBeenCalledWith(
        JSON.stringify({
          type: 'cancelOrder',
          data: { orderId: '123' }
        })
      )
    })
  })

  describe('Error Handling', () => {
    it('handles connection errors', () => {
      wsService.initialize()

      const ws = mockWebSocketCtor.mock.results[0].value as WebSocket
      ws.onerror?.({ error: new Error('test error') } as ErrorEvent)

      expect(mockDispatch).toHaveBeenCalledWith(
        addAlert({
          level: 'error',
          message: expect.stringContaining('error'),
          source: 'websocket',
        })
      )
    })

    it('handles reconnection failure', () => {
      wsService.initialize()
      const ws = mockWebSocketCtor.mock.results[0].value as WebSocket
      
      // Simulate max reconnection attempts
      for (let i = 0; i <= wsService['maxReconnectAttempts']; i++) {
        ws.onclose?.({ reason: 'test close' } as CloseEvent)
      }

      expect(wsService['reconnectAttempts']).toBe(wsService['maxReconnectAttempts'])
    })
  })

  describe('Cleanup', () => {
    it('disconnects properly', () => {
      wsService.initialize()
      wsService.disconnect()

      expect(mockWebSocket.close).toHaveBeenCalled()
    })
  })
})                        