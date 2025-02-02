import { describe, it, expect, beforeEach, vi } from 'vitest'
import { io } from 'socket.io-client'
import { wsService } from './websocket'
import { store } from '@/store'
import { updateMarketData } from '@/store/trading/tradingSlice'
import { addAlert } from '@/store/monitoring/monitoringSlice'

vi.mock('socket.io-client')
vi.mock('@/store')

describe('WebSocket Service', () => {
  interface MockSocket {
    on: jest.Mock;
    emit: jest.Mock;
    disconnect: jest.Mock;
  }

  interface MockStore {
    dispatch: jest.Mock;
  }

  let mockSocket: MockSocket
  let mockStore: MockStore

  beforeEach(() => {
    mockSocket = {
      on: vi.fn(),
      emit: vi.fn(),
      disconnect: vi.fn(),
    }

    mockStore = {
      dispatch: vi.fn(),
    }

    ;(io as any).mockReturnValue(mockSocket)
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

      const connectHandler = mockSocket.on.mock.calls.find(
        (call: [string, ...any[]]) => call[0] === 'connect'
      )[1]
      connectHandler()

      expect(wsService['reconnectAttempts']).toBe(0)
    })

    it('handles disconnection', () => {
      wsService.initialize()

      const disconnectHandler = mockSocket.on.mock.calls.find(
        (call: [string, ...any[]]) => call[0] === 'disconnect'
      )[1]
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

      const reconnectHandler = mockSocket.on.mock.calls.find(
        (call: [string, ...any[]]) => call[0] === 'reconnect_attempt'
      )[1]
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

      const marketDataHandler = mockSocket.on.mock.calls.find(
        (call: [string, ...any[]]) => call[0] === 'marketData'
      )[1]

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

      const errorHandler = mockSocket.on.mock.calls.find(
        (call: [string, ...any[]]) => call[0] === 'error'
      )[1]

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
        const reconnectHandler = mockSocket.on.mock.calls.find(
          (call: [string, ...any[]]) => call[0] === 'reconnect_attempt'
        )[1]
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