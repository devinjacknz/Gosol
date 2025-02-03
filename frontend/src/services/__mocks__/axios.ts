import { vi } from 'vitest';
import axios, { 
  AxiosResponse, 
  AxiosError, 
  AxiosRequestConfig, 
  AxiosInstance,
  InternalAxiosRequestConfig
} from 'axios';
import type { Order, Position } from '@/types/trading';
import type { MarketData, HealthCheck } from '@/types/api';

type MockFn = ReturnType<typeof vi.fn>;
interface MockAxiosFn extends MockFn {
  mockResolvedValue: (value: AxiosResponse) => MockAxiosFn;
  mockRejectedValue: (error: AxiosError) => MockAxiosFn;
  mockResolvedValueOnce: (value: AxiosResponse) => MockAxiosFn;
  mockRejectedValueOnce: (error: AxiosError) => MockAxiosFn;
  mockReset: () => MockAxiosFn;
  mockImplementation: (fn: (...args: any[]) => Promise<AxiosResponse>) => MockAxiosFn;
}

const headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json'
} as const;

const defaultConfig: InternalAxiosRequestConfig = {
  headers: axios.AxiosHeaders.from(headers),
  method: 'get',
  url: '',
  baseURL: 'http://localhost:8080'
};

type MockFn = ReturnType<typeof vi.fn>;
interface MockAxiosFn extends MockFn {
  mockResolvedValue: (value: any) => MockAxiosFn;
  mockRejectedValue: (error: any) => MockAxiosFn;
  mockResolvedValueOnce: (value: any) => MockAxiosFn;
  mockRejectedValueOnce: (error: any) => MockAxiosFn;
  mockReset: () => void;
}

const getMock = vi.fn() as MockAxiosFn;
const postMock = vi.fn() as MockAxiosFn;
const putMock = vi.fn() as MockAxiosFn;
const deleteMock = vi.fn() as MockAxiosFn;
const requestMock = vi.fn() as MockAxiosFn;

getMock.mockImplementation((url: string, config?: AxiosRequestConfig) => {
  const baseResponse = { 
    status: 200, 
    statusText: 'OK', 
    headers,
    config: { ...defaultConfig, ...config, url },
    request: {},
    data: { success: true, data: {} }
  };

  if (url.includes('/health')) {
    return Promise.resolve({ ...baseResponse, data: { success: true, data: mockData.health } });
  }
  if (url.includes('/market')) {
    return Promise.resolve({ ...baseResponse, data: { success: true, data: mockData.market } });
  }
  if (url.includes('/orders')) {
    return Promise.resolve({ ...baseResponse, data: { success: true, data: [mockData.order] } });
  }
  if (url.includes('/positions')) {
    return Promise.resolve({ ...baseResponse, data: { success: true, data: [mockData.position] } });
  }
  return Promise.reject(createErrorResponse(404, 'Resource not found'));
});

postMock.mockImplementation((url: string, data?: any, config?: AxiosRequestConfig) => {
  const baseResponse = { 
    status: 200, 
    statusText: 'OK', 
    headers,
    config: { ...defaultConfig, ...config, url },
    request: {},
    data: { success: true, data: {} }
  };

  if (url.includes('/orders')) {
    return Promise.resolve({ ...baseResponse, data: { success: true, data: { ...mockData.order, ...data } } });
  }
  return Promise.reject(createErrorResponse(404, 'Resource not found'));
});

putMock.mockImplementation((url: string, data?: any, config?: AxiosRequestConfig) => {
  const baseResponse = { 
    status: 200, 
    statusText: 'OK', 
    headers,
    config: { ...defaultConfig, ...config, url },
    request: {},
    data: { success: true, data: {} }
  };

  if (url.includes('/positions')) {
    return Promise.resolve({ ...baseResponse, data: { success: true, data: { ...mockData.position, ...data } } });
  }
  return Promise.reject(createErrorResponse(404, 'Resource not found'));
});

deleteMock.mockImplementation((url: string, config?: AxiosRequestConfig) => {
  const baseResponse = { 
    status: 200, 
    statusText: 'OK', 
    headers,
    config: { ...defaultConfig, ...config, url },
    request: {},
    data: { success: true, data: {} }
  };

  if (url.includes('/orders/')) {
    return Promise.resolve({ ...baseResponse, data: { success: true, data: { success: true } } });
  }
  return Promise.reject(createErrorResponse(404, 'Resource not found'));
});

const mockData = {
  market: {
    price: 50000,
    volume: 100,
    timestamp: new Date().toISOString(),
    high: 51000,
    low: 49000,
    open: 49500,
    close: 50000
  } satisfies MarketData,
  order: {
    id: '1',
    symbol: 'BTC/USDT',
    type: 'limit' as const,
    side: 'buy' as const,
    size: 1,
    price: 50000,
    status: 'open' as const,
    timestamp: Date.now()
  } satisfies Order,
  position: {
    symbol: 'BTC/USDT',
    size: 1,
    entryPrice: 50000,
    markPrice: 51000,
    pnl: 1000,
    status: 'open' as const,
    lastUpdated: Date.now()
  } satisfies Position,
  health: {
    status: 'ok' as const,
    services: {
      database: { status: 'up', latency: 5 },
      redis: { status: 'up', latency: 2 },
      websocket: { status: 'up', latency: 2 }
    }
  } satisfies HealthCheck
};

const createErrorResponse = (status: number, message: string): AxiosError => {
  const error = new Error(message) as AxiosError;
  error.isAxiosError = true;
  error.name = 'AxiosError';
  error.message = message;
  error.config = defaultConfig;
  error.response = {
    data: { success: false, data: null, error: message },
    status,
    statusText: message,
    headers,
    config: defaultConfig,
    request: {}
  };
  return error;
};

requestMock.mockImplementation((config: AxiosRequestConfig) => {
  const method = (config.method || 'get').toLowerCase();
  switch (method) {
    case 'get': return getMock(config.url || '', config);
    case 'post': return postMock(config.url || '', config.data, config);
    case 'put': return putMock(config.url || '', config.data, config);
    case 'delete': return deleteMock(config.url || '', config);
    default: return Promise.reject(createErrorResponse(405, `Method ${method} not implemented`));
  }
});

const axiosInstance = {
  defaults: {
    baseURL: 'http://localhost:8080',
    headers: axios.AxiosHeaders.from(headers)
  },
  get: getMock,
  post: postMock,
  put: putMock,
  delete: deleteMock,
  request: requestMock,
  interceptors: {
    request: { use: vi.fn(), eject: vi.fn(), clear: vi.fn() },
    response: { use: vi.fn(), eject: vi.fn(), clear: vi.fn() }
  }
} as unknown as AxiosInstance;

const mockAxios = {
  create: vi.fn().mockReturnValue(mockAxiosInstance)
};

const mockAxiosInstance = {
  defaults: {
    baseURL: 'http://localhost:8080',
    headers
  },
  get: getMock,
  post: postMock,
  put: putMock,
  delete: deleteMock,
  request: requestMock,
  interceptors: {
    request: { use: vi.fn(), eject: vi.fn(), clear: vi.fn() },
    response: { use: vi.fn(), eject: vi.fn(), clear: vi.fn() }
  }
} as unknown as AxiosInstance;

const mockAxios = {
  create: vi.fn().mockReturnValue(axiosInstance),
  get: getMock,
  post: postMock,
  put: putMock,
  delete: deleteMock,
  request: requestMock,
  interceptors: {
    request: { use: vi.fn(), eject: vi.fn(), clear: vi.fn() },
    response: { use: vi.fn(), eject: vi.fn(), clear: vi.fn() }
  },
  isAxiosError: axios.isAxiosError
} as unknown as AxiosInstance & { create: typeof axios.create };

export default mockAxios;
