import { useEffect, useRef, useCallback } from 'react'
import { useDispatch } from 'react-redux'
import { io, Socket } from 'socket.io-client'
import { updateMarketData } from '@/store/trading/tradingSlice'
import { addAlert } from '@/store/monitoring/monitoringSlice'

export const useWebSocket = (symbol: string) => {
  const dispatch = useDispatch()
  const socketRef = useRef<Socket>()
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>()
  const reconnectAttemptsRef = useRef(0)
  const maxReconnectAttempts = 5
  const reconnectDelay = 1000

  const handleConnectionError = useCallback(() => {
    dispatch(addAlert({
      level: 'error',
      message: '网络连接已断开',
      source: 'websocket',
    }))

    if (reconnectAttemptsRef.current < maxReconnectAttempts) {
      reconnectTimeoutRef.current = setTimeout(() => {
        reconnectAttemptsRef.current++
        socketRef.current?.disconnect()
        initializeSocket()
      }, reconnectDelay * Math.pow(2, reconnectAttemptsRef.current))
    }
  }, [dispatch])

  const handleConnectionSuccess = useCallback(() => {
    reconnectAttemptsRef.current = 0
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }

    dispatch(addAlert({
      level: 'success',
      message: '网络已重新连接',
      source: 'websocket',
    }))
  }, [dispatch])

  const initializeSocket = useCallback(() => {
    if (socketRef.current?.connected) return

    socketRef.current = io(process.env.VITE_WS_URL || 'ws://localhost:8080/ws', {
      reconnection: true,
      reconnectionDelay: reconnectDelay,
      reconnectionAttempts: maxReconnectAttempts,
    })

    socketRef.current.on('connect', () => {
      console.log('WebSocket connected')
      handleConnectionSuccess()
      socketRef.current?.emit('subscribe', { channel: 'marketData', symbol })
    })

    socketRef.current.on('disconnect', (reason) => {
      console.log('WebSocket disconnected:', reason)
      handleConnectionError()
    })

    socketRef.current.on('error', (error) => {
      console.error('WebSocket error:', error)
      dispatch(addAlert({
        level: 'error',
        message: `WebSocket error: ${error.message}`,
        source: 'websocket',
      }))
    })

    socketRef.current.on('marketData', (data) => {
      dispatch(updateMarketData({ symbol, data }))
    })
  }, [dispatch, handleConnectionError, handleConnectionSuccess, symbol])

  useEffect(() => {
    initializeSocket()

    // 添加网络状态监听
    window.addEventListener('online', handleConnectionSuccess)
    window.addEventListener('offline', handleConnectionError)

    return () => {
      // 取消订阅
      socketRef.current?.emit('unsubscribe', { channel: 'marketData', symbol })

      // 移除监听器
      window.removeEventListener('online', handleConnectionSuccess)
      window.removeEventListener('offline', handleConnectionError)

      // 清理重连定时器
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }

      // 关闭连接
      socketRef.current?.disconnect()
    }
  }, [symbol, handleConnectionSuccess, handleConnectionError, initializeSocket])

  return socketRef.current
} 