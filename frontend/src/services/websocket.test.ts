import { describe, it, expect, beforeEach, vi } from 'vitest'
import { io, Socket } from 'socket.io-client'
import { wsService } from './websocket'
import { store } from '@/store'
import { updateMarketData } from '@/store/trading/tradingSlice'
import { addAlert } from '@/store/monitoring/monitoringSlice';

vi.mock('socket.io-client', () => ({
  io: vi.fn(),
}));

vi.mock('@/store', () => ({
  store: {
    dispatch: vi.fn(),
  },
}));

describe('WebSocket Service', () => {
  type MockSocket = {
    on: ReturnType<typeof vi.fn>;
    emit: ReturnType<typeof vi.fn>;
    disconnect: ReturnType<typeof vi.fn>;
  }

  type MockStore = {
    dispatch: ReturnType<typeof vi.fn>;
  }

  let mockSocket: MockSocket
  let mockStore: MockStore

  beforeEach(() => {
    vi.resetAllMocks()
    
    mockSocket = {
      on: vi.fn().mockReturnThis(),
      emit: vi.fn().mockReturnThis(),
      disconnect: vi.fn().mockReturnThis(),
    }

    mockStore = {
      dispatch: vi.fn(),
    }

    vi.mocked(io).mockReturnValue(mockSocket as unknown as Socket)

    ;(store as any).mockReturnValue(mockStore)

    // Reset service
    wsService.disconnect()
  })

  describe('Connection Management', () => {
    it('initializes connection', () => {
      wsService.initialize()

      expect(io).toHaveBeenCalledWith(expect.any(String), {
        reconnection: true,
        reconnectionDelay: expect.any(Number),
        reconnectionAttempts: expect.any(Number),
      })
    })

    it('handles connection success', () => {
      wsService.initialize()

      const connectCall = mockSocket.on.mock.calls.find(
        (call): call is [string, (...args: any[]) => void] => 
          Array.isArray(call) && call[0] === 'connect'
      )
      const connectHandler = connectCall?.[1]
      if (!connectHandler) throw new Error('Connect handler not found')
      connectHandler()

      expect(wsService['reconnectAttempts']).toBe(0)
    })

    it('handles disconnection', () => {
      wsService.initialize()

      const disconnectCall = mockSocket.on.mock.calls.find(
        (call): call is [string, (...args: any[]) => void] => 
          Array.isArray(call) && call[0] === 'disconnect'
      )
      const disconnectHandler = disconnectCall?.[1]
      if (!disconnectHandler) throw new Error('Disconnect handler not found')
      disconnectHandler('transport close')

      expect(store.dispatch).toHaveBeenCalledWith(
        addAlert({
          level: 'warning',
          message: expect.stringContaining('disconnected'),
          source: 'websocket',
        })
      )
    })

    it('handles reconnection attempts', () => {
      wsService.initialize()

      const reconnectCall = mockSocket.on.mock.calls.find(
        (call): call is [string, (...args: any[]) => void] => 
          Array.isArray(call) && call[0] === 'reconnect_attempt'
      )
      const reconnectHandler = reconnectCall?.[1]
      if (!reconnectHandler) throw new Error('Reconnect handler not found')
      reconnectHandler(1)

      expect(wsService['reconnectAttempts']).toBe(1)
    })
  })

  describe('Market Data Handling', () => {
    it('subscribes to market data', () => {
      wsService.initialize()
      wsService.subscribeMarketData('BTC/USDT')

      expect(mockSocket.emit).toHaveBeenCalledWith('subscribe', {
        channel: 'marketData',
        symbol: 'BTC/USDT',
      })
    })

    it('unsubscribes from market data', () => {
      wsService.initialize()
      wsService.unsubscribeMarketData('BTC/USDT')

      expect(mockSocket.emit).toHaveBeenCalledWith('unsubscribe', {
        channel: 'marketData',
        symbol: 'BTC/USDT',
      })
    })

    it('handles market data updates', () => {
      wsService.initialize()

      const marketDataCall = mockSocket.on.mock.calls.find(
        (call): call is [string, (...args: any[]) => void] => 
          Array.isArray(call) && call[0] === 'marketData'
      )
      const marketDataHandler = marketDataCall?.[1]
      if (!marketDataHandler) throw new Error('Market data handler not found')

      const mockData = {
        symbol: 'BTC/USDT',
        price: 50000,
      }

      marketDataHandler(mockData)

      expect(store.dispatch).toHaveBeenCalledWith(updateMarketData(mockData))
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

      expect(mockSocket.emit).toHaveBeenCalledWith('placeOrder', order)
    })

    it('cancels order', () => {
      wsService.initialize()
      wsService.cancelOrder('123')

      expect(mockSocket.emit).toHaveBeenCalledWith('cancelOrder', {
        orderId: '123',
      })
    })
  })

  describe('Error Handling', () => {
    it('handles connection errors', () => {
      wsService.initialize()

      const errorCall = mockSocket.on.mock.calls.find(
        (call): call is [string, (...args: any[]) => void] => 
          Array.isArray(call) && call[0] === 'error'
      )
      const errorHandler = errorCall?.[1]
      if (!errorHandler) throw new Error('Error handler not found')

      const error = new Error('Connection failed')
      errorHandler(error)

      expect(store.dispatch).toHaveBeenCalledWith(
        addAlert({
          level: 'error',
          message: expect.stringContaining('error'),
          source: 'websocket',
        })
      )
    })

    it('handles reconnection failure', () => {
      wsService.initialize()

      // Simulate max reconnection attempts
      for (let i = 0; i <= wsService['maxReconnectAttempts']; i++) {
        const reconnectCall = mockSocket.on.mock.calls.find(
          (call): call is [string, (...args: any[]) => void] => 
            Array.isArray(call) && call[0] === 'reconnect_attempt'
        )
        const reconnectHandler = reconnectCall?.[1]
        if (!reconnectHandler) throw new Error('Reconnect handler not found')
        reconnectHandler(i)
      }

      expect(wsService['reconnectAttempts']).toBe(wsService['maxReconnectAttempts'])
    })
  })

  describe('Cleanup', () => {
    it('disconnects properly', () => {
      wsService.initialize()
      wsService.disconnect()

      expect(mockSocket.disconnect).toHaveBeenCalled()
    })
  })
})                