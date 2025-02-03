import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('axios', async () => {
  const actual = await vi.importActual('axios')
  return {
    ...actual,
    default: {
      create: vi.fn().mockReturnValue({
        get: vi.fn().mockResolvedValue({ data: {} }),
        post: vi.fn().mockResolvedValue({ data: {} }),
        put: vi.fn().mockResolvedValue({ data: {} }),
        delete: vi.fn().mockResolvedValue({ data: {} }),
        interceptors: {
          request: { use: vi.fn() },
          response: { use: vi.fn() }
        }
      })
    }
  }
})

import { marketApi, tradingApi, analysisApi, monitoringApi } from './api'

describe('API Services', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockAxiosInstance.get.mockResolvedValue({ data: {} })
    mockAxiosInstance.post.mockResolvedValue({ data: {} })
    mockAxiosInstance.put.mockResolvedValue({ data: {} })
    mockAxiosInstance.delete.mockResolvedValue({ data: {} })
  })

  describe('Market API', () => {
    it('fetches klines data', async () => {
      const mockData = {
        data: [
          {
            time: '2024-02-20T00:00:00Z',
            open: 50000,
            high: 51000,
            low: 49000,
            close: 50500,
            volume: 1000,
          },
        ],
      }

      mockAxiosInstance.get.mockResolvedValueOnce(mockData)

      const result = await marketApi.getKlines('BTC/USDT', '1h')
      expect(result).toEqual(mockData.data)
      expect(mockAxiosInstance.get).toHaveBeenCalledWith(
        '/market/klines?symbol=BTC/USDT&interval=1h&limit=1000'
      )
    })

    it('fetches order book', async () => {
      const mockData = {
        data: {
          bids: [[50000, 1]],
          asks: [[50100, 1]],
        },
      }

      mockAxiosInstance.get.mockResolvedValueOnce(mockData)

      const result = await marketApi.getOrderBook('BTC/USDT')
      expect(result).toEqual(mockData.data)
      expect(mockAxiosInstance.get).toHaveBeenCalledWith(
        '/market/depth?symbol=BTC/USDT&limit=20'
      )
    })
  })

  describe('Trading API', () => {
    it('places order', async () => {
      const order = {
        symbol: 'BTC/USDT',
        side: 'buy',
        price: 50000,
        size: 1,
      }

      const mockData = {
        data: {
          orderId: '123',
          status: 'NEW',
        },
      }

      mockAxiosInstance.post.mockResolvedValueOnce(mockData)

      const result = await tradingApi.placeOrder(order)
      expect(result).toEqual(mockData.data)
      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/trading/orders', order)
    })

    it('cancels order', async () => {
      const mockData = {
        data: {
          orderId: '123',
          status: 'CANCELED',
        },
      }

      mockAxiosInstance.delete.mockResolvedValueOnce(mockData)

      const result = await tradingApi.cancelOrder('123')
      expect(result).toEqual(mockData.data)
      expect(mockAxiosInstance.delete).toHaveBeenCalledWith('/trading/orders/123')
    })
  })

  describe('Analysis API', () => {
    it('gets technical indicators', async () => {
      const mockData = {
        data: {
          rsi: 65,
          macd: {
            macd: 100,
            signal: 50,
            histogram: 50,
          },
        },
      }

      mockAxiosInstance.get.mockResolvedValueOnce(mockData)

      const result = await analysisApi.getIndicators('BTC/USDT', {
        timeframe: '1h',
      })
      expect(result).toEqual(mockData.data)
      expect(mockAxiosInstance.get).toHaveBeenCalledWith(
        '/analysis/indicators/BTC/USDT',
        { params: { timeframe: '1h' } }
      )
    })

    it('gets LLM analysis', async () => {
      const mockData = {
        data: {
          analysis: 'Market shows bullish trend',
          confidence: 0.8,
        },
      }

      mockAxiosInstance.post.mockResolvedValueOnce(mockData)

      const result = await analysisApi.getLLMAnalysis({
        symbol: 'BTC/USDT',
        timeframe: '1h',
      })
      expect(result).toEqual(mockData.data)
    })
  })

  describe('Monitoring API', () => {
    it('gets system metrics', async () => {
      const mockData = {
        data: {
          cpuUsage: 50,
          memoryUsage: 60,
          requestCount: 1000,
        },
      }

      mockAxiosInstance.get.mockResolvedValueOnce(mockData)

      const result = await monitoringApi.getMetrics()
      expect(result).toEqual(mockData.data)
      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/monitoring/metrics')
    })

    it('gets alerts', async () => {
      const mockData = {
        data: [
          {
            level: 'warning',
            message: 'High CPU usage',
            timestamp: '2024-02-20T00:00:00Z',
          },
        ],
      }

      mockAxiosInstance.get.mockResolvedValueOnce(mockData)

      const result = await monitoringApi.getAlerts()
      expect(result).toEqual(mockData.data)
      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/monitoring/alerts')
    })
  })

  describe('Error Handling', () => {
    it('handles network errors', async () => {
      const error = new Error('Network Error')
      mockAxiosInstance.get.mockRejectedValueOnce(error)

      await expect(marketApi.getKlines('BTC/USDT', '1h')).rejects.toThrow(
        'Network Error'
      )
    })

    it('handles API errors', async () => {
      const error = {
        response: {
          data: {
            message: 'Invalid symbol',
          },
          status: 400,
        },
      }
      mockAxiosInstance.get.mockRejectedValueOnce(error)

      await expect(marketApi.getKlines('INVALID', '1h')).rejects.toThrow(
        'Invalid symbol'
      )
    })
  })
})                      