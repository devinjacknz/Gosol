import { debugService } from './debug'

export interface MarketData {
  symbol: string
  price: number
  volume: number
  change24h: number
  high24h: number
  low24h: number
  timestamp: string
}

export interface Position {
  symbol: string
  size: number
  entryPrice: number
  markPrice: number
  pnl: number
  roe: number
}

export interface SystemMetric {
  name: string
  value: number
  status: 'normal' | 'warning' | 'error'
  unit: string
  timestamp: string
}

export interface TradingMetric {
  exchange: string
  status: 'online' | 'offline' | 'error'
  latency: number
  ordersPerSecond: number
  errorRate: number
  lastSync: string
}

class MockDataService {
  private static instance: MockDataService
  private updateInterval: number = 2000 // 2 seconds
  private intervalIds: NodeJS.Timeout[] = []

  private constructor() {}

  static getInstance(): MockDataService {
    if (!MockDataService.instance) {
      MockDataService.instance = new MockDataService()
    }
    return MockDataService.instance
  }

  private randomChange(base: number, maxPercent: number): number {
    const change = (Math.random() - 0.5) * 2 * (base * maxPercent)
    return Number((base + change).toFixed(2))
  }

  getMarketData(): MarketData[] {
    return [
      {
        symbol: 'BTC-USD',
        price: 43250.5,
        volume: 1234567,
        change24h: 2.5,
        high24h: 44000,
        low24h: 42800,
        timestamp: new Date().toISOString(),
      },
      {
        symbol: 'ETH-USD',
        price: 2250.75,
        volume: 987654,
        change24h: -1.2,
        high24h: 2300,
        low24h: 2200,
        timestamp: new Date().toISOString(),
      },
    ]
  }

  getPositions(): Position[] {
    return [
      {
        symbol: 'BTC-USD',
        size: 0.1,
        entryPrice: 42000,
        markPrice: 43250.5,
        pnl: 125.5,
        roe: 2.98,
      },
    ]
  }

  getSystemMetrics(): SystemMetric[] {
    return [
      {
        name: 'CPU 使用率',
        value: 45,
        status: 'normal',
        unit: '%',
        timestamp: new Date().toISOString(),
      },
      {
        name: '内存使用率',
        value: 72,
        status: 'warning',
        unit: '%',
        timestamp: new Date().toISOString(),
      },
      {
        name: '系统延迟',
        value: 150,
        status: 'normal',
        unit: 'ms',
        timestamp: new Date().toISOString(),
      },
      {
        name: '错误率',
        value: 0.5,
        status: 'normal',
        unit: '%',
        timestamp: new Date().toISOString(),
      },
    ]
  }

  getTradingMetrics(): TradingMetric[] {
    return [
      {
        exchange: 'dYdX',
        status: 'online',
        latency: 120,
        ordersPerSecond: 5.2,
        errorRate: 0.1,
        lastSync: new Date().toISOString(),
      },
      {
        exchange: 'HyperLiquid',
        status: 'online',
        latency: 180,
        ordersPerSecond: 4.8,
        errorRate: 0.2,
        lastSync: new Date().toISOString(),
      },
    ]
  }

  simulateMarketDataUpdate(callback: (data: MarketData[]) => void) {
    const intervalId = setInterval(() => {
      const data = this.getMarketData().map(item => ({
        ...item,
        price: this.randomChange(item.price, 0.002),
        volume: this.randomChange(item.volume, 0.01),
        timestamp: new Date().toISOString(),
      }))
      callback(data)
      debugService.debug('Market data updated', data)
    }, this.updateInterval)
    
    this.intervalIds.push(intervalId)
    return () => clearInterval(intervalId)
  }

  simulateSystemMetricsUpdate(callback: (data: SystemMetric[]) => void) {
    const intervalId = setInterval(() => {
      const data = this.getSystemMetrics().map(item => ({
        ...item,
        value: this.randomChange(item.value, 0.05),
        timestamp: new Date().toISOString(),
      }))
      callback(data)
      debugService.debug('System metrics updated', data)
    }, this.updateInterval)
    
    this.intervalIds.push(intervalId)
    return () => clearInterval(intervalId)
  }

  simulateTradingMetricsUpdate(callback: (data: TradingMetric[]) => void) {
    const intervalId = setInterval(() => {
      const data = this.getTradingMetrics().map(item => ({
        ...item,
        latency: this.randomChange(item.latency, 0.1),
        ordersPerSecond: this.randomChange(item.ordersPerSecond, 0.05),
        errorRate: this.randomChange(item.errorRate, 0.1),
        lastSync: new Date().toISOString(),
      }))
      callback(data)
      debugService.debug('Trading metrics updated', data)
    }, this.updateInterval)
    
    this.intervalIds.push(intervalId)
    return () => clearInterval(intervalId)
  }

  stopAllSimulations() {
    this.intervalIds.forEach(clearInterval)
    this.intervalIds = []
    debugService.info('All simulations stopped')
  }
}

export const mockDataService = MockDataService.getInstance()
