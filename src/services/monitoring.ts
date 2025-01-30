import { Gauge, Counter, Histogram } from 'prom-client';
import { logger } from '../lib/logger';
import { config } from '../config';
import { dbManager } from '../lib/database';
import { cacheManager } from '../lib/cache';

class MonitoringService {
  private static instance: MonitoringService;

  // 系统指标
  private metrics = {
    // 系统负载
    systemLoad: new Gauge({
      name: 'system_load_average',
      help: 'System load average'
    }),

    // API请求
    apiRequests: new Counter({
      name: 'api_requests_total',
      help: 'Total number of API requests',
      labelNames: ['method', 'endpoint', 'status']
    }),

    // 响应时间
    responseTime: new Histogram({
      name: 'api_response_time_seconds',
      help: 'API response time in seconds',
      labelNames: ['method', 'endpoint']
    }),

    // 数据库连接
    dbConnections: new Gauge({
      name: 'database_connections',
      help: 'Number of active database connections',
      labelNames: ['database']
    }),

    // 缓存命中率
    cacheHitRate: new Gauge({
      name: 'cache_hit_rate',
      help: 'Cache hit rate percentage',
      labelNames: ['cache']
    }),

    // 内存使用
    memoryUsage: new Gauge({
      name: 'memory_usage_bytes',
      help: 'Process memory usage in bytes'
    })
  };

  // 业务指标
  private businessMetrics = {
    // 活跃用户
    activeUsers: new Gauge({
      name: 'active_users',
      help: 'Number of active users'
    }),

    // 订单统计
    orders: new Counter({
      name: 'orders_total',
      help: 'Total number of orders',
      labelNames: ['type', 'status']
    }),

    // 交易量
    tradingVolume: new Counter({
      name: 'trading_volume_total',
      help: 'Total trading volume',
      labelNames: ['symbol']
    }),

    // 系统盈亏
    systemPnL: new Gauge({
      name: 'system_pnl',
      help: 'System profit and loss',
      labelNames: ['type']
    })
  };

  private constructor() {
    this.startMetricsCollection();
  }

  public static getInstance(): MonitoringService {
    if (!MonitoringService.instance) {
      MonitoringService.instance = new MonitoringService();
    }
    return MonitoringService.instance;
  }

  // 启动指标收集
  private startMetricsCollection(): void {
    setInterval(() => {
      this.collectSystemMetrics();
      this.collectBusinessMetrics();
    }, config.monitoring.metrics.interval);
  }

  // 收集系统指标
  private async collectSystemMetrics(): Promise<void> {
    try {
      // 系统负载
      const load = process.cpuUsage();
      this.metrics.systemLoad.set(load.user / load.system);

      // 内存使用
      const memory = process.memoryUsage();
      this.metrics.memoryUsage.set(memory.heapUsed);

      // 数据库连接
      const dbHealth = await dbManager.healthCheck();
      Object.entries(dbHealth).forEach(([db, status]) => {
        this.metrics.dbConnections.set({ database: db }, status ? 1 : 0);
      });

      logger.debug('System metrics collected successfully');
    } catch (error) {
      logger.error('Failed to collect system metrics:', error);
    }
  }

  // 收集业务指标
  private async collectBusinessMetrics(): Promise<void> {
    try {
      // TODO: 实现业务指标收集逻辑
      logger.debug('Business metrics collected successfully');
    } catch (error) {
      logger.error('Failed to collect business metrics:', error);
    }
  }

  // 记录API请求
  public recordApiRequest(method: string, endpoint: string, status: number, duration: number): void {
    this.metrics.apiRequests.inc({ method, endpoint, status });
    this.metrics.responseTime.observe({ method, endpoint }, duration);
  }

  // 记录缓存命中
  public recordCacheHit(cache: string, hit: boolean): void {
    const current = this.metrics.cacheHitRate.get({ cache }) || 0;
    const newRate = (current * 0.9) + (hit ? 0.1 : 0);
    this.metrics.cacheHitRate.set({ cache }, newRate);
  }

  // 记录订单
  public recordOrder(type: string, status: string): void {
    this.businessMetrics.orders.inc({ type, status });
  }

  // 记录交易量
  public recordTrading(symbol: string, volume: number): void {
    this.businessMetrics.tradingVolume.inc({ symbol }, volume);
  }

  // 健康检查
  public async healthCheck(): Promise<{
    status: 'healthy' | 'unhealthy';
    details: any;
  }> {
    try {
      const dbHealth = await dbManager.healthCheck();
      const memory = process.memoryUsage();
      const load = process.cpuUsage();

      const healthy = Object.values(dbHealth).every(status => status) &&
        memory.heapUsed < memory.heapTotal * 0.9 &&
        load.user / load.system < 0.9;

      return {
        status: healthy ? 'healthy' : 'unhealthy',
        details: {
          database: dbHealth,
          memory: {
            used: memory.heapUsed,
            total: memory.heapTotal,
            external: memory.external
          },
          cpu: {
            user: load.user,
            system: load.system
          }
        }
      };
    } catch (error) {
      logger.error('Health check failed:', error);
      return {
        status: 'unhealthy',
        details: { error: error.message }
      };
    }
  }

  // 告警检查
  public async checkAlerts(): Promise<void> {
    try {
      const health = await this.healthCheck();
      
      if (health.status === 'unhealthy') {
        logger.warn('System health check failed', health.details);
        // TODO: 实现告警通知逻辑
      }

      // 检查响应时间
      const p95ResponseTime = this.metrics.responseTime.get().values[0].value;
      if (p95ResponseTime > config.monitoring.alerts.highLatency) {
        logger.warn('High API latency detected', { p95ResponseTime });
        // TODO: 实现告警通知逻辑
      }

      // 检查错误率
      const totalRequests = this.metrics.apiRequests.get().values
        .reduce((acc, curr) => acc + curr.value, 0);
      const errorRequests = this.metrics.apiRequests.get().values
        .filter(v => v.labels.status >= 500)
        .reduce((acc, curr) => acc + curr.value, 0);
      
      const errorRate = errorRequests / totalRequests;
      if (errorRate > config.monitoring.alerts.errorRate) {
        logger.warn('High error rate detected', { errorRate });
        // TODO: 实现告警通知逻辑
      }
    } catch (error) {
      logger.error('Alert check failed:', error);
    }
  }
}

export const monitoringService = MonitoringService.getInstance(); 