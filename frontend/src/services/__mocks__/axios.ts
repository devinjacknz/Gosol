import { vi } from 'vitest';
import type { AxiosError, AxiosResponse, InternalAxiosRequestConfig } from 'axios';
import type { Order, Position, MarketData, HealthCheck } from '@/types/api';

const createApiError = (status: number, message: string, code = 'ERROR'): AxiosError => {
  const error = new Error(message) as AxiosError;
  error.isAxiosError = true;
  error.code = code;
  error.config = {
    url: '',
    method: 'get',
    baseURL: 'http://localhost:8080',
    timeout: 5000,
    headers: { 'Content-Type': 'application/json' }
  } as InternalAxiosRequestConfig;
  error.response = {
    status,
    statusText: message,
    headers: { 'content-type': 'application/json' },
    config: error.config,
    data: { 
      success: false, 
      error: message,
      code,
      details: {
        timestamp: new Date().toISOString(),
        path: error.config.url,
        method: error.config.method?.toUpperCase()
      }
    }
  } as AxiosResponse;
  return error;
};

const createNetworkError = (message: string, code = 'NETWORK_ERROR'): AxiosError => {
  const error = new Error(message) as AxiosError;
  error.isAxiosError = true;
  error.code = code;
  error.config = {
    url: '',
    method: 'get',
    baseURL: 'http://localhost:8080',
    timeout: 5000
  } as InternalAxiosRequestConfig;
  error.request = { status: 0, statusText: message };
  return error;
};

const mockAxiosInstance = {
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
  request: vi.fn(),
  getUri: vi.fn(),
  defaults: {
    baseURL: 'http://localhost:8080',
    headers: { common: {} }
  },
  interceptors: {
    request: { use: vi.fn(), eject: vi.fn() },
    response: { use: vi.fn(), eject: vi.fn() }
  }
};

const createMockResponse = <T>(data: T): AxiosResponse => ({
  data: { success: true, data },
  status: 200,
  statusText: 'OK',
  headers: {},
  config: {} as InternalAxiosRequestConfig
});

const mockMarketData: MarketData = {
  price: 50000,
  volume: 1000,
  timestamp: new Date().toISOString(),
  high: 51000,
  low: 49000,
  open: 49500,
  close: 50000
};

const mockPosition: Position = {
  symbol: 'BTC/USDT',
  size: 1.0,
  entryPrice: 50000,
  markPrice: 51000,
  pnl: 1000,
  status: 'open',
  lastUpdated: new Date().toISOString()
};

const mockOrder: Order = {
  id: '1',
  symbol: 'BTC/USDT',
  type: 'limit',
  side: 'buy',
  size: 1.0,
  price: 50000,
  status: 'open',
  timestamp: new Date().toISOString()
};

mockAxiosInstance.get.mockImplementation((url: string) => {
  if (url.includes('timeout')) return Promise.reject(createNetworkError('Request timeout', 'ECONNABORTED'));
  if (url.includes('network')) return Promise.reject(createNetworkError('Connection refused', 'ECONNREFUSED'));
  if (url.includes('error')) return Promise.reject(createApiError(500, 'Internal server error'));
  if (url.includes('rate-limit')) return Promise.reject(createApiError(429, 'Too Many Requests'));
  if (url.includes('service-unavailable')) return Promise.reject(createApiError(503, 'Service unavailable'));
  if (url.includes('unauthorized')) return Promise.reject(createApiError(401, 'Unauthorized access'));
  if (url.includes('insufficient-margin')) return Promise.reject(createApiError(400, 'Insufficient margin'));
  if (url.includes('invalid')) return Promise.reject(createApiError(400, 'Invalid request'));

  switch (url) {
    case '/api/market/BTC-USD':
      return Promise.resolve(createMockResponse(mockMarketData));
    case '/api/positions':
      return Promise.resolve(createMockResponse([mockPosition]));
    case '/api/orders':
      return Promise.resolve(createMockResponse([mockOrder]));
    case '/health':
      return Promise.resolve(createMockResponse({
        status: 'ok',
        services: {
          backend: { status: 'up', latency: 50 },
          database: { status: 'up', latency: 20 }
        }
      } as HealthCheck));
    default:
      return Promise.reject(createApiError(404, `Resource not found: ${url}`));
  }
});

mockAxiosInstance.post.mockImplementation((url: string, data: unknown) => {
  if (url.includes('timeout')) return Promise.reject(createNetworkError('Request timeout', 'ECONNABORTED'));
  if (url.includes('network')) return Promise.reject(createNetworkError('Connection refused', 'ECONNREFUSED'));
  if (url.includes('error')) return Promise.reject(createApiError(500, 'Internal server error'));
  if (url.includes('rate-limit')) return Promise.reject(createApiError(429, 'Too Many Requests'));
  if (url.includes('invalid-order')) return Promise.reject(createApiError(400, 'Invalid order data'));

  switch (url) {
    case '/api/orders':
      return Promise.resolve(createMockResponse({
        ...mockOrder,
        ...data,
        id: '123',
        timestamp: new Date().toISOString()
      }));
    case '/api/orders/batch':
      return Promise.resolve(createMockResponse(
        (data as Array<unknown>).map((order, index) => ({
          id: `batch-${index + 1}`,
          timestamp: new Date().toISOString(),
          status: 'open',
          ...order
        }))
      ));
    default:
      return Promise.reject(createApiError(404, 'Invalid endpoint'));
  }
});

mockAxiosInstance.put.mockImplementation((url: string, data: unknown) => {
  if (url.includes('timeout')) return Promise.reject(createNetworkError('Request timeout', 'ECONNABORTED'));
  if (url.includes('network')) return Promise.reject(createNetworkError('Connection refused', 'ECONNREFUSED'));
  if (url.includes('error')) return Promise.reject(createApiError(500, 'Internal server error'));

  switch (url) {
    case '/api/positions':
      return Promise.resolve(createMockResponse(Array.isArray(data) ? data : [data]));
    default:
      return Promise.reject(createApiError(404, 'Invalid endpoint'));
  }
});

mockAxiosInstance.delete.mockImplementation((url: string) => {
  if (url.includes('timeout')) return Promise.reject(createNetworkError('Request timeout', 'ECONNABORTED'));
  if (url.includes('network')) return Promise.reject(createNetworkError('Connection refused', 'ECONNREFUSED'));
  if (url.includes('error')) return Promise.reject(createApiError(500, 'Internal server error'));

  if (url.includes('/api/orders/')) {
    return Promise.resolve(createMockResponse({ success: true }));
  }

  return Promise.reject(createApiError(404, 'Order not found'));
});

const mockAxios = {
  create: vi.fn(() => mockAxiosInstance),
  isAxiosError: (error: unknown): error is AxiosError => 
    error instanceof Error && 'isAxiosError' in error,
  defaults: {
    headers: { common: {} },
    transformRequest: [],
    transformResponse: [],
    timeout: 0,
    withCredentials: false,
    adapter: vi.fn(),
    responseType: 'json',
    xsrfCookieName: 'XSRF-TOKEN',
    xsrfHeaderName: 'X-XSRF-TOKEN',
    validateStatus: (status: number) => status >= 200 && status < 300
  },
  get: mockAxiosInstance.get,
  post: mockAxiosInstance.post,
  put: mockAxiosInstance.put,
  delete: mockAxiosInstance.delete,
  request: mockAxiosInstance.request,
  getUri: mockAxiosInstance.getUri,
  interceptors: mockAxiosInstance.interceptors
};

export default mockAxios;
mockAxiosInstance.get.mockImplementation((url: string): Promise<AxiosResponse> => {
  if (url.includes('timeout')) return Promise.reject(createNetworkError('Request timeout', 'ECONNABORTED'));
  if (url.includes('network')) return Promise.reject(createNetworkError('Connection failed', 'ERR_NETWORK'));
  if (url.includes('error')) return Promise.reject(createApiError(500, 'Internal server error'));
  if (url.includes('invalid')) return Promise.reject(createApiError(400, 'Invalid request'));
  if (url.includes('rate-limit')) return Promise.reject(createApiError(429, 'Too Many Requests'));
  if (url.includes('service-unavailable')) return Promise.reject(createApiError(503, 'Service unavailable'));
  if (url.includes('unauthorized')) return Promise.reject(createApiError(401, 'Unauthorized access'));
  if (url.includes('insufficient-margin')) return Promise.reject(createApiError(400, 'Insufficient margin'));

  const mockMarketData: MarketData = {
    price: 50000,
    volume: 1000,
    timestamp: new Date().toISOString(),
    high: 51000,
    low: 49000,
    open: 49500,
    close: 50000
  };

  const mockPosition: Position = {
    symbol: 'BTC/USDT',
    size: 1.0,
    entryPrice: 50000,
    markPrice: 51000,
    pnl: 1000,
    status: 'open',
    lastUpdated: new Date().toISOString()
  };

  switch (url) {
    case '/api/market/BTC-USD':
      return Promise.resolve({
        data: { success: true, data: mockMarketData },
        status: 200,
        statusText: 'OK',
        headers: {},
        config: {} as InternalAxiosRequestConfig
      });

      switch (url) {
        case '/api/market/BTC-USD':
          return Promise.resolve({
            data: {
              success: true,
              data: {
                price: 50000,
                volume: 1000,
                timestamp: new Date().toISOString(),
                high: 51000,
                low: 49000,
                open: 49500,
                close: 50000
              }
            }
          });
        case '/api/positions':
          return Promise.resolve({
            data: {
              success: true,
              data: [{
                symbol: 'BTC/USDT',
                size: 1.0,
                entryPrice: 50000,
                markPrice: 51000,
                pnl: 1000,
                status: 'open',
                lastUpdated: new Date().toISOString()
              }]
            }
          });
        default:
          return Promise.reject(createApiError(404, 'Resource not found'));
      }
    });

mockAxiosInstance.post.mockImplementation((url: string, data: unknown) => {
      if (url.includes('timeout')) return Promise.reject(createNetworkError('Request timeout', 'ECONNABORTED'));
      if (url.includes('network')) return Promise.reject(createNetworkError('Connection failed', 'ERR_NETWORK'));
      if (url.includes('error')) return Promise.reject(createApiError(500, 'Internal server error'));
      if (url.includes('invalid-order')) return Promise.reject(createApiError(400, 'Invalid order data'));
      if (url.includes('rate-limit')) return Promise.reject(createApiError(429, 'Too Many Requests'));

      switch (url) {
        case '/api/orders':
          return Promise.resolve({
            data: {
              success: true,
              data: {
                id: '123',
                symbol: 'BTC/USDT',
                type: 'limit',
                side: 'buy',
                size: 1.0,
                price: 50000,
                status: 'open',
                timestamp: new Date().toISOString()
              }
            }
          });
        default:
          return Promise.reject(createApiError(404, 'Invalid endpoint'));
      }
    });

    return mockAxiosInstance;
  },
  isAxiosError: (error: unknown): error is AxiosError => 
    error instanceof Error && 'isAxiosError' in error
};
    if (url.includes('timeout')) return Promise.reject(createNetworkError('Request timeout', 'ECONNABORTED'));
    if (url.includes('network')) return Promise.reject(createNetworkError('Connection refused', 'ECONNREFUSED'));
    if (url.includes('error')) return Promise.reject(createApiError(500, 'Internal server error'));
    if (url.includes('rate-limit')) return Promise.reject(createApiError(429, 'Too Many Requests'));
    if (url.includes('service-unavailable')) return Promise.reject(createApiError(503, 'Service unavailable'));
    if (url.includes('cancelled')) return Promise.reject(createNetworkError('Request cancelled', 'ERR_CANCELED'));
    if (url.includes('insufficient-margin')) return Promise.reject(createApiError(400, 'Insufficient margin'));
    if (url.includes('order-not-found')) return Promise.reject(createApiError(404, 'Order not found'));
    if (url.includes('invalid')) return Promise.reject(createApiError(400, 'Invalid request'));

    switch (url) {
      case '/api/market/BTC-USD':
        return Promise.resolve({
          data: {
            success: true,
            data: {
              price: 50000,
              volume: 1000,
              timestamp: new Date().toISOString(),
              high: 51000,
              low: 49000,
              open: 49500,
              close: 50000
            } as MarketData
          }
        });
      case '/api/positions':
        return Promise.resolve({
          data: {
            success: true,
            data: [{
              symbol: 'BTC/USDT',
              size: 1.0,
              entryPrice: 50000,
              markPrice: 51000,
              pnl: 1000,
              status: 'open',
              lastUpdated: new Date().toISOString()
            }] as Position[]
          }
        });
      case '/api/orders':
        return Promise.resolve({
          data: {
            success: true,
            data: [{
              id: '1',
              symbol: 'BTC/USDT',
              type: 'market',
              side: 'buy',
              size: 1.0,
              status: 'open',
              timestamp: new Date().toISOString()
            }] as Order[]
          }
        });
      case '/health':
        return Promise.resolve({
          data: {
            success: true,
            data: {
              status: 'ok',
              services: {
                backend: { status: 'up', latency: 50 },
                database: { status: 'up', latency: 20 }
              }
            } as HealthCheck
          }
        });
      default:
        return Promise.reject(createApiError(404, `Resource not found: ${url}`, 'NOT_FOUND'));
    }
  }),
  post: vi.fn().mockImplementation((url: string, data: unknown) => {
    if (url.includes('timeout')) return Promise.reject(createNetworkError('Request timeout', 'ECONNABORTED'));
    if (url.includes('network')) return Promise.reject(createNetworkError('Connection refused', 'ECONNREFUSED'));
    if (url.includes('error')) return Promise.reject(createApiError(500, 'Internal server error'));
    if (url.includes('rate-limit')) return Promise.reject(createApiError(429, 'Too Many Requests'));
    if (url.includes('invalid-order')) return Promise.reject(createApiError(400, 'Invalid order data'));
    if (url.includes('cancelled')) return Promise.reject(createNetworkError('Request cancelled', 'ERR_CANCELED'));

    switch (url) {
      case '/api/orders':
        return Promise.resolve({
          data: {
            success: true,
            data: {
              id: '123',
              ...(data as object),
              status: 'open',
              timestamp: new Date().toISOString()
            } as Order
          }
        });
      case '/api/orders/batch':
        return Promise.resolve({
          data: {
            success: true,
            data: (data as Array<unknown>).map((order, index) => ({
              id: `batch-${index + 1}`,
              status: 'open',
              timestamp: new Date().toISOString(),
              ...order
            }))
          }
        });
      default:
        return Promise.reject(createApiError(404, 'Invalid endpoint', 'INVALID_ENDPOINT'));
    }
  }),
  delete: vi.fn().mockImplementation((url: string) => {
    if (url.includes('timeout')) return Promise.reject(createNetworkError('Request timeout', 'ECONNABORTED'));
    if (url.includes('network')) return Promise.reject(createNetworkError('Connection refused', 'ECONNREFUSED'));
    if (url.includes('error')) return Promise.reject(createApiError(500, 'Internal server error'));
    if (url.includes('rate-limit')) return Promise.reject(createApiError(429, 'Too Many Requests'));
    if (url.includes('cancelled')) return Promise.reject(createNetworkError('Request cancelled', 'ERR_CANCELED'));
    
    if (url.startsWith('/api/orders/')) {
      const orderId = url.split('/').pop();
      if (!orderId) return Promise.reject(createApiError(400, 'Invalid order ID'));
      if (orderId === 'not-found') return Promise.reject(createApiError(404, 'Order not found'));
      if (orderId === 'permission-denied') return Promise.reject(createApiError(403, 'Permission denied'));
      return Promise.resolve({
        data: {
          success: true,
          data: {
            id: orderId,
            status: 'cancelled',
            timestamp: new Date().toISOString()
          }
        }
      });
    }
    throw createApiError(404, 'Order not found', 'ORDER_NOT_FOUND');
  }),
  isAxiosError: (error: unknown): error is AxiosError => 
    error instanceof Error && 'isAxiosError' in error
};

export default mockAxios;
