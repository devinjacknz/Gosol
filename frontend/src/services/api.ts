import axios from 'axios'
import { store } from '@/store'
import { addAlert } from '@/store/monitoring/monitoringSlice'

// 创建 axios 实例
const api = axios.create({
  baseURL: process.env.VITE_API_URL || 'http://localhost:8080/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 添加认证信息
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    // 处理错误
    const message = error.response?.data?.message || error.message
    store.dispatch(addAlert({
      level: 'error',
      message: `API Error: ${message}`,
      source: 'api',
    }))
    return Promise.reject(error)
  }
)

// 市场数据 API
export const marketApi = {
  // 获取K线数据
  getKlines: (symbol: string, interval: string, limit = 1000) =>
    api.get(`/market/klines?symbol=${symbol}&interval=${interval}&limit=${limit}`),

  // 获取行情深度
  getOrderBook: (symbol: string, limit = 20) =>
    api.get(`/market/depth?symbol=${symbol}&limit=${limit}`),

  // 获取最新成交
  getTrades: (symbol: string, limit = 50) =>
    api.get(`/market/trades?symbol=${symbol}&limit=${limit}`),
}

// 交易 API
export const tradingApi = {
  // 下单
  placeOrder: (order: any) =>
    api.post('/trading/orders', order),

  // 取消订单
  cancelOrder: (orderId: string) =>
    api.delete(`/trading/orders/${orderId}`),

  // 获取订单列表
  getOrders: (params: any) =>
    api.get('/trading/orders', { params }),

  // 获取持仓列表
  getPositions: () =>
    api.get('/trading/positions'),
}

// 分析 API
export const analysisApi = {
  // 获取技术指标
  getIndicators: (symbol: string, params: any) =>
    api.get(`/analysis/indicators/${symbol}`, { params }),

  // 获取市场分析
  getAnalysis: (symbol: string) =>
    api.get(`/analysis/market/${symbol}`),

  // LLM 分析
  getLLMAnalysis: (params: any) =>
    api.post('/analysis/llm', params),
}

// 监控 API
export const monitoringApi = {
  // 获取系统指标
  getMetrics: () =>
    api.get('/monitoring/metrics'),

  // 获取告警信息
  getAlerts: () =>
    api.get('/monitoring/alerts'),

  // 获取系统状态
  getStatus: () =>
    api.get('/monitoring/status'),
}

// 用户 API
export const userApi = {
  // 登录
  login: (credentials: any) =>
    api.post('/auth/login', credentials),

  // 注册
  register: (userData: any) =>
    api.post('/auth/register', userData),

  // 获取用户信息
  getProfile: () =>
    api.get('/users/profile'),

  // 更新用户信息
  updateProfile: (data: any) =>
    api.put('/users/profile', data),
}

export default api 