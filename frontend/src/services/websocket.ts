import { store } from '@/store'
import { updateMarketData } from '@/store/trading/tradingSlice'
import { addAlert } from '@/store/monitoring/monitoringSlice'

class WebSocketService {
  private socket: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private symbol: string | null = null

  constructor() {
    this.initialize()
  }

  public initialize() {
    if (this.socket?.readyState === WebSocket.OPEN) {
      return
    }

    const wsUrl = `${import.meta.env.VITE_WS_URL || 'ws://localhost:8080'}/ws/${this.formatSymbol(this.symbol)}`
    this.socket = new WebSocket(wsUrl)
    this.setupEventListeners()
  }

  private formatSymbol(symbol: string): string {
    return symbol?.replace('/', '-') || 'BTC-USDT'
  }

  private setupEventListeners() {
    if (!this.socket) return

    this.socket.onopen = () => {
      console.log('WebSocket connected')
      this.reconnectAttempts = 0
      
      // Subscribe to market data
      if (this.symbol) {
        this.socket?.send(JSON.stringify({
          type: 'subscribe',
          symbol: this.symbol
        }))
      }
    }

    this.socket.onclose = (event) => {
      console.log('WebSocket disconnected:', event.reason)
      store.dispatch(addAlert({
        level: 'warning',
        message: `WebSocket disconnected: ${event.reason}`,
        source: 'websocket',
      }))

      // Attempt to reconnect
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++
        console.log(`WebSocket reconnection attempt ${this.reconnectAttempts}`)
        setTimeout(() => this.initialize(), this.reconnectDelay)
      }
    }

    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error)
      store.dispatch(addAlert({
        level: 'error',
        message: 'WebSocket connection error',
        source: 'websocket',
      }))
    }

    this.socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 'market_data') {
          store.dispatch(updateMarketData(data.data))
        }
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error)
      }
    }
  }

  public subscribeMarketData(symbol: string) {
    this.symbol = symbol
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify({
        type: 'subscribe',
        symbol: this.formatSymbol(symbol)
      }))
    } else {
      this.initialize()
    }
  }

  public unsubscribeMarketData(symbol: string) {
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify({
        type: 'unsubscribe',
        symbol: this.formatSymbol(symbol)
      }))
    }
  }

  public sendOrder(order: any) {
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify({
        type: 'placeOrder',
        data: order
      }))
    }
  }

  public cancelOrder(orderId: string) {
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify({
        type: 'cancelOrder',
        data: { orderId }
      }))
    }
  }

  public disconnect() {
    this.socket?.close()
    this.socket = null
  }
}

export const wsService = new WebSocketService()        