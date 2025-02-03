import { describe, it, expect, vi } from 'vitest';
import axios from 'axios';
import { ApiService } from '@/services/api';

vi.mock('axios');

const mockGet = vi.fn();
const mockPost = vi.fn();
const mockDelete = vi.fn();

vi.mocked(axios.create).mockReturnValue({
  get: mockGet,
  post: mockPost,
  delete: mockDelete
} as any);

describe('ApiService', () => {
  let apiService: ApiService;
  let mockAxiosInstance: any;

  beforeEach(() => {
    vi.clearAllMocks();
    apiService = new ApiService();
    mockAxiosInstance = (axios.create as any)();
  });

  it('should fetch market data', async () => {
    const mockData = { symbol: 'BTC-USD', price: '50000' };
    mockGet.mockResolvedValueOnce({ data: mockData });

    const result = await apiService.getMarketData('BTC-USD');
    expect(result).toEqual(mockData);
    expect(mockAxiosInstance.get).toHaveBeenCalledWith('/api/market/BTC-USD');
  });

  it('should handle errors when fetching market data', async () => {
    const mockError = new Error('Network error');
    mockGet.mockRejectedValueOnce(mockError);

    await expect(apiService.getMarketData('BTC-USD')).rejects.toThrow('Network error');
    expect(mockAxiosInstance.get).toHaveBeenCalledWith('/api/market/BTC-USD');
  });
});
