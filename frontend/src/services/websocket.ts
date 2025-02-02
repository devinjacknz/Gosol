import { io, Socket } from 'socket.io-client'
import { store } from '@/store'
import { updateMarketData } from '@/store/trading/tradingSlice'
import { addAlert } from '@/store/monitoring/monitoringSlice'

class WebSocketService {
  private socket: Socket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000

  constructor() {
    this.initialize()
  }

  private initialize() {
    this.socket = io(process.env.VITE_WS_URL || 'ws://localhost:8080/ws', {
      reconnection: true,
      reconnectionDelay: this.reconnectDelay,
      reconnectionAttempts: this.maxReconnectAttempts,
    })

    this.setupEventListeners()
  }

  private setupEventListeners() {
    if (!this.socket) return

    // 连接事件
    this.socket.on('connect', () => {
      console.log('WebSocket connected')
      this.reconnectAttempts = 0
    })

    // 断开连接事件
    this.socket.on('disconnect', (reason) => {
      console.log('WebSocket disconnected:', reason)
      store.dispatch(addAlert({
        level: 'warning',
        message: `WebSocket disconnected: ${reason}`,
        source: 'websocket',
      }))
    })

    // 重连事件
    this.socket.on('reconnect_attempt', (attempt) => {
      this.reconnectAttempts = attempt
      console.log(`WebSocket reconnection attempt ${attempt}`)
    })

    // 错误事件
    this.socket.on('error', (error) => {
      console.error('WebSocket error:', error)
      store.dispatch(addAlert({
        level: 'error',
        message: `WebSocket error: ${error.message}`,
        source: 'websocket',
      }))
    })

    // 市场数据更新
    this.socket.on('marketData', (data) => {
      store.dispatch(updateMarketData(data))
    })

    // 订单更新
    this.socket.on('orderUpdate', (data) => {
      // 处理订单更新
    })

    // 持仓更新
    this.socket.on('positionUpdate', (data) => {
      // 处理持仓更新
    })
  }

  // 订阅市场数据
  public subscribeMarketData(symbol: string) {
    this.socket?.emit('subscribe', { channel: 'marketData', symbol })
  }

  // 取消订阅市场数据
  public unsubscribeMarketData(symbol: string) {
    this.socket?.emit('unsubscribe', { channel: 'marketData', symbol })
  }

  // 发送订单
  public sendOrder(order: any) {
    this.socket?.emit('placeOrder', order)
  }

  // 取消订单
  public cancelOrder(orderId: string) {
    this.socket?.emit('cancelOrder', { orderId })
  }

  // 关闭连接
  public disconnect() {
    this.socket?.disconnect()
  }
}

export const wsService = new WebSocketService() 