import React from 'react';
import { render, screen, waitFor } from '../../utils/test-utils';
import { mockApiResponse } from '../../utils/test-utils';
import Dashboard from '@/components/dashboard/Dashboard';

// 模拟API调用
jest.mock('@/utils/api', () => ({
  fetchDashboardData: jest.fn(() => Promise.resolve(mockApiResponse)),
}));

describe('Dashboard Component', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders dashboard with loading state', () => {
    render(<Dashboard />);
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it('displays portfolio summary', async () => {
    const mockData = {
      portfolio: {
        totalValue: 100000,
        dailyPnL: 5000,
        totalPnL: 15000,
        positions: [
          {
            symbol: 'BTC/USDT',
            amount: 1.5,
            value: 75000,
            pnl: 3000,
          },
        ],
      },
    };

    jest.mock('@/utils/api', () => ({
      fetchDashboardData: jest.fn(() => Promise.resolve({ ...mockApiResponse, data: mockData })),
    }));

    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('$100,000')).toBeInTheDocument();
      expect(screen.getByText('+$5,000')).toBeInTheDocument();
    });
  });

  it('displays trading history', async () => {
    const mockData = {
      trades: [
        {
          id: '1',
          symbol: 'BTC/USDT',
          side: 'buy',
          price: 50000,
          amount: 1.0,
          timestamp: new Date().toISOString(),
        },
      ],
    };

    jest.mock('@/utils/api', () => ({
      fetchDashboardData: jest.fn(() => Promise.resolve({ ...mockApiResponse, data: mockData })),
    }));

    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('BTC/USDT')).toBeInTheDocument();
      expect(screen.getByText('50000')).toBeInTheDocument();
    });
  });

  it('displays performance metrics', async () => {
    const mockData = {
      performance: {
        winRate: 0.65,
        profitFactor: 2.1,
        sharpeRatio: 1.8,
        maxDrawdown: -0.15,
      },
    };

    jest.mock('@/utils/api', () => ({
      fetchDashboardData: jest.fn(() => Promise.resolve({ ...mockApiResponse, data: mockData })),
    }));

    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('65%')).toBeInTheDocument();
      expect(screen.getByText('2.1')).toBeInTheDocument();
      expect(screen.getByText('1.8')).toBeInTheDocument();
      expect(screen.getByText('-15%')).toBeInTheDocument();
    });
  });

  it('handles error state', async () => {
    const mockError = {
      success: false,
      error: 'Failed to fetch dashboard data',
    };

    jest.mock('@/utils/api', () => ({
      fetchDashboardData: jest.fn(() => Promise.reject(mockError)),
    }));

    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText(/failed to fetch dashboard data/i)).toBeInTheDocument();
    });
  });

  it('updates data periodically', async () => {
    const mockData1 = {
      portfolio: {
        totalValue: 100000,
      },
    };

    const mockData2 = {
      portfolio: {
        totalValue: 101000,
      },
    };

    const fetchDashboardData = jest.fn()
      .mockResolvedValueOnce({ ...mockApiResponse, data: mockData1 })
      .mockResolvedValueOnce({ ...mockApiResponse, data: mockData2 });

    jest.mock('@/utils/api', () => ({
      fetchDashboardData,
    }));

    // 设置定时器模拟
    jest.useFakeTimers();

    render(<Dashboard />);

    // 验证初始数据
    await waitFor(() => {
      expect(screen.getByText('$100,000')).toBeInTheDocument();
    });

    // 前进30秒
    jest.advanceTimersByTime(30000);

    // 验证更新后的数据
    await waitFor(() => {
      expect(screen.getByText('$101,000')).toBeInTheDocument();
    });

    // 恢复真实定时器
    jest.useRealTimers();
  });

  it('filters data by time range', async () => {
    render(<Dashboard />);

    // 选择时间范围
    const rangeSelect = screen.getByRole('combobox', { name: /time range/i });
    fireEvent.change(rangeSelect, { target: { value: '7d' } });

    // 验证API调用包含正确的时间范围参数
    await waitFor(() => {
      expect(fetchDashboardData).toHaveBeenCalledWith(expect.objectContaining({
        timeRange: '7d',
      }));
    });
  });
}); 