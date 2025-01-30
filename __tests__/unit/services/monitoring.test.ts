import { Gauge, Counter, Histogram } from 'prom-client';
import { monitoringService } from '../../services/monitoring';
import { dbManager } from '../../lib/database';
import { config } from '../../config';

// Mock外部依赖
jest.mock('prom-client');
jest.mock('../../lib/database');
jest.mock('../../lib/logger');

describe('MonitoringService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  describe('Metrics Collection', () => {
    it('should start collecting metrics on initialization', () => {
      expect(setInterval).toHaveBeenCalledWith(
        expect.any(Function),
        config.monitoring.metrics.interval
      );
    });

    it('should collect system metrics successfully', async () => {
      const mockCpuUsage = { user: 100, system: 200 };
      const mockMemoryUsage = {
        heapUsed: 1000,
        heapTotal: 2000,
        external: 500
      };

      jest.spyOn(process, 'cpuUsage').mockReturnValue(mockCpuUsage);
      jest.spyOn(process, 'memoryUsage').mockReturnValue(mockMemoryUsage as any);
      (dbManager.healthCheck as jest.Mock).mockResolvedValue({
        mongodb: true,
        redis: true,
        timescaledb: true
      });

      await monitoringService['collectSystemMetrics']();

      expect(monitoringService['metrics'].systemLoad.set)
        .toHaveBeenCalledWith(mockCpuUsage.user / mockCpuUsage.system);
      expect(monitoringService['metrics'].memoryUsage.set)
        .toHaveBeenCalledWith(mockMemoryUsage.heapUsed);
      expect(monitoringService['metrics'].dbConnections.set)
        .toHaveBeenCalledTimes(3);
    });

    it('should handle system metrics collection errors', async () => {
      const error = new Error('Collection failed');
      jest.spyOn(process, 'cpuUsage').mockImplementation(() => {
        throw error;
      });

      await expect(monitoringService['collectSystemMetrics']())
        .resolves.not.toThrow();
    });
  });

  describe('API Request Tracking', () => {
    it('should record API request metrics', () => {
      const method = 'GET';
      const endpoint = '/api/test';
      const status = 200;
      const duration = 100;

      monitoringService.recordApiRequest(method, endpoint, status, duration);

      expect(monitoringService['metrics'].apiRequests.inc)
        .toHaveBeenCalledWith({ method, endpoint, status });
      expect(monitoringService['metrics'].responseTime.observe)
        .toHaveBeenCalledWith({ method, endpoint }, duration);
    });
  });

  describe('Cache Hit Rate Tracking', () => {
    it('should record cache hit rate', () => {
      const cache = 'market';
      const hit = true;

      monitoringService.recordCacheHit(cache, hit);

      expect(monitoringService['metrics'].cacheHitRate.set)
        .toHaveBeenCalledWith({ cache }, expect.any(Number));
    });
  });

  describe('Business Metrics', () => {
    it('should record order metrics', () => {
      const type = 'limit';
      const status = 'filled';

      monitoringService.recordOrder(type, status);

      expect(monitoringService['businessMetrics'].orders.inc)
        .toHaveBeenCalledWith({ type, status });
    });

    it('should record trading volume', () => {
      const symbol = 'BTC/USDT';
      const volume = 1.5;

      monitoringService.recordTrading(symbol, volume);

      expect(monitoringService['businessMetrics'].tradingVolume.inc)
        .toHaveBeenCalledWith({ symbol }, volume);
    });
  });

  describe('Health Check', () => {
    it('should return healthy status when all checks pass', async () => {
      const mockMemoryUsage = {
        heapUsed: 1000,
        heapTotal: 2000,
        external: 500
      };
      const mockCpuUsage = { user: 100, system: 200 };

      jest.spyOn(process, 'memoryUsage').mockReturnValue(mockMemoryUsage as any);
      jest.spyOn(process, 'cpuUsage').mockReturnValue(mockCpuUsage);
      (dbManager.healthCheck as jest.Mock).mockResolvedValue({
        mongodb: true,
        redis: true,
        timescaledb: true
      });

      const health = await monitoringService.healthCheck();

      expect(health.status).toBe('healthy');
      expect(health.details).toEqual({
        database: {
          mongodb: true,
          redis: true,
          timescaledb: true
        },
        memory: mockMemoryUsage,
        cpu: mockCpuUsage
      });
    });

    it('should return unhealthy status when checks fail', async () => {
      (dbManager.healthCheck as jest.Mock).mockResolvedValue({
        mongodb: false,
        redis: false,
        timescaledb: false
      });

      const health = await monitoringService.healthCheck();

      expect(health.status).toBe('unhealthy');
    });
  });

  describe('Alert Checking', () => {
    it('should trigger alerts when thresholds are exceeded', async () => {
      // Mock unhealthy system status
      jest.spyOn(monitoringService, 'healthCheck').mockResolvedValue({
        status: 'unhealthy',
        details: {}
      });

      // Mock high latency
      jest.spyOn(monitoringService['metrics'].responseTime, 'get')
        .mockReturnValue({
          values: [{ value: config.monitoring.alerts.highLatency + 100 }]
        } as any);

      // Mock high error rate
      jest.spyOn(monitoringService['metrics'].apiRequests, 'get')
        .mockReturnValue({
          values: [
            { labels: { status: 500 }, value: 50 },
            { labels: { status: 200 }, value: 950 }
          ]
        } as any);

      await monitoringService.checkAlerts();

      // Verify that appropriate warnings were logged
      expect(console.warn).toHaveBeenCalledTimes(3);
    });

    it('should not trigger alerts when metrics are within thresholds', async () => {
      // Mock healthy system status
      jest.spyOn(monitoringService, 'healthCheck').mockResolvedValue({
        status: 'healthy',
        details: {}
      });

      // Mock normal latency
      jest.spyOn(monitoringService['metrics'].responseTime, 'get')
        .mockReturnValue({
          values: [{ value: config.monitoring.alerts.highLatency - 100 }]
        } as any);

      // Mock low error rate
      jest.spyOn(monitoringService['metrics'].apiRequests, 'get')
        .mockReturnValue({
          values: [
            { labels: { status: 500 }, value: 1 },
            { labels: { status: 200 }, value: 999 }
          ]
        } as any);

      await monitoringService.checkAlerts();

      // Verify that no warnings were logged
      expect(console.warn).not.toHaveBeenCalled();
    });
  });
}); 