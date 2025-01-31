import axios from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001';

interface ApiResponse<T> {
  data: T;
  status: number;
  message?: string;
}

interface DashboardParams {
  timeRange: string;
}

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add request interceptor to handle authentication
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Add response interceptor to handle errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized access
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export const fetchDashboardData = async (params: DashboardParams) => {
  try {
    const response = await api.get<ApiResponse<any>>('/api/dashboard', {
      params,
    });
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      throw new Error(error.response?.data?.message || 'Failed to fetch dashboard data');
    }
    throw error;
  }
};

export const login = async (email: string, password: string) => {
  try {
    const response = await api.post<ApiResponse<{ token: string }>>('/api/auth/login', {
      email,
      password,
    });
    const { token } = response.data.data;
    localStorage.setItem('token', token);
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      throw new Error(error.response?.data?.message || 'Login failed');
    }
    throw error;
  }
};

export const logout = () => {
  localStorage.removeItem('token');
  window.location.href = '/login';
};

export const placeOrder = async (order: {
  symbol: string;
  side: 'buy' | 'sell';
  type: 'market' | 'limit';
  price?: number;
  amount: number;
}) => {
  try {
    const response = await api.post<ApiResponse<any>>('/api/orders', order);
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      throw new Error(error.response?.data?.message || 'Failed to place order');
    }
    throw error;
  }
};

export const getOrderBook = async (symbol: string) => {
  try {
    const response = await api.get<ApiResponse<any>>(`/api/orderbook/${symbol}`);
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      throw new Error(error.response?.data?.message || 'Failed to fetch order book');
    }
    throw error;
  }
};

export const getTrades = async (symbol: string) => {
  try {
    const response = await api.get<ApiResponse<any>>(`/api/trades/${symbol}`);
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      throw new Error(error.response?.data?.message || 'Failed to fetch trades');
    }
    throw error;
  }
}; 