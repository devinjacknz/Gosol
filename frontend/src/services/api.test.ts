import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ApiService } from './api';
import axios from 'axios';

vi.mock('axios');

describe('API Services', () => {
  let apiService: ApiService;

  beforeEach(() => {
    vi.clearAllMocks();
    axios.create().get.mockResolvedValue({ data: {} });
    axios.create().post.mockResolvedValue({ data: {} });
    axios.create().delete.mockResolvedValue({ data: {} });
    apiService = new ApiService();
  });

  describe('Market API', () => {
    it('fetches market data', async () => {
      const mockData = { price: 50000, volume: 100 };
      axios.create().get.mockResolvedValueOnce({ data: mockData });
      
      const result = await apiService.getMarketData('BTC-USD');
      expect(result).toEqual(mockData);
    });
  });

  describe('Trading API', () => {
    it('places order', async () => {
      const mockOrder = {
        symbol: 'BTC-USD',
        side: 'buy',
        type: 'limit',
        price: 50000,
        size: 1
      };
      const mockResponse = { success: true };
      axios.create().post.mockResolvedValueOnce({ data: mockResponse });
      
      const result = await apiService.placeOrder(mockOrder);
      expect(result).toEqual(mockResponse);
    });

    it('cancels order', async () => {
      const orderId = '123';
      const mockResponse = { success: true };
      axios.create().delete.mockResolvedValueOnce({ data: mockResponse });
      
      const result = await apiService.cancelOrder(orderId);
      expect(result).toEqual(mockResponse);
    });
  });

  describe('Error Handling', () => {
    it('handles API errors', async () => {
      const errorMessage = 'API Error';
      axios.create().get.mockRejectedValueOnce({ 
        response: { 
          status: 400, 
          data: { message: errorMessage } 
        } 
      });

      await expect(apiService.getMarketData('BTC-USD')).rejects.toThrow(`API Error: 400 - ${errorMessage}`);
    });

    it('handles network errors', async () => {
      const errorMessage = 'Network Error';
      axios.create().get.mockRejectedValueOnce(new Error(errorMessage));

      await expect(apiService.getMarketData('BTC-USD')).rejects.toThrow(`Network Error: ${errorMessage}`);
    });
  });
});
