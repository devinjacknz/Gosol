import axios, { AxiosInstance } from 'axios';
import { Order, Position } from '@/pages/TradingView';

export class ApiService {
  private api: AxiosInstance;

  constructor() {
    this.api = axios.create({
      baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
      headers: {
        'Content-Type': 'application/json',
      },
      timeout: 5000,
    });

    if (this.api?.interceptors) {
      this.api.interceptors.response.use(
      response => response,
      error => {
        if (error.response) {
          throw new Error(`API Error: ${error.response.status} - ${error.response.data?.message || error.message}`);
        }
        throw new Error(`Network Error: ${error.message}`);
      }
    );
    }
  }

  async getMarketData(symbol: string) {
    const response = await this.api.get(`/api/market/${symbol}`);
    return response.data;
  }

  async getPositions(): Promise<Position[]> {
    const response = await this.api.get('/api/positions');
    return response.data;
  }

  async getOrders(): Promise<Order[]> {
    const response = await this.api.get('/api/orders');
    return response.data;
  }

  async placeOrder(order: Omit<Order, 'id' | 'status' | 'timestamp'>): Promise<{ success: boolean }> {
    const response = await this.api.post('/api/orders', order);
    return response.data;
  }

  async cancelOrder(orderId: string): Promise<{ success: boolean }> {
    const response = await this.api.delete(`/api/orders/${orderId}`);
    return response.data;
  }
}

export const apiService = new ApiService();
