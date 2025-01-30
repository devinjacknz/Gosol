import { Redis } from 'ioredis';
import { dbManager } from './database';
import { logger } from './logger';
import { config } from '../config';

class CacheManager {
  private static instance: CacheManager;
  private redis: Redis | null = null;

  private constructor() {}

  public static getInstance(): CacheManager {
    if (!CacheManager.instance) {
      CacheManager.instance = new CacheManager();
    }
    return CacheManager.instance;
  }

  // 初始化Redis连接
  public async init(): Promise<void> {
    try {
      this.redis = await dbManager.connectRedis();
    } catch (error) {
      logger.error('Failed to initialize cache manager:', error);
      throw error;
    }
  }

  // 市场数据缓存
  public market = {
    // 缓存行情数据
    async setTicker(symbol: string, data: any): Promise<void> {
      try {
        const key = `market:ticker:${symbol}`;
        await this.redis?.setex(
          key,
          config.cache.market.ticker,
          JSON.stringify(data)
        );
      } catch (error) {
        logger.error('Failed to cache ticker:', error);
        throw error;
      }
    },

    // 获取行情数据
    async getTicker(symbol: string): Promise<any> {
      try {
        const key = `market:ticker:${symbol}`;
        const data = await this.redis?.get(key);
        return data ? JSON.parse(data) : null;
      } catch (error) {
        logger.error('Failed to get ticker from cache:', error);
        throw error;
      }
    },

    // 缓存深度数据
    async setOrderBook(symbol: string, data: any): Promise<void> {
      try {
        const key = `market:orderbook:${symbol}`;
        await this.redis?.setex(
          key,
          config.cache.market.orderBook,
          JSON.stringify(data)
        );
      } catch (error) {
        logger.error('Failed to cache order book:', error);
        throw error;
      }
    },

    // 获取深度数据
    async getOrderBook(symbol: string): Promise<any> {
      try {
        const key = `market:orderbook:${symbol}`;
        const data = await this.redis?.get(key);
        return data ? JSON.parse(data) : null;
      } catch (error) {
        logger.error('Failed to get order book from cache:', error);
        throw error;
      }
    }
  };

  // 用户数据缓存
  public user = {
    // 缓存账户信息
    async setAccount(userId: string, data: any): Promise<void> {
      try {
        const key = `user:account:${userId}`;
        await this.redis?.setex(
          key,
          config.cache.user.account,
          JSON.stringify(data)
        );
      } catch (error) {
        logger.error('Failed to cache account:', error);
        throw error;
      }
    },

    // 获取账户信息
    async getAccount(userId: string): Promise<any> {
      try {
        const key = `user:account:${userId}`;
        const data = await this.redis?.get(key);
        return data ? JSON.parse(data) : null;
      } catch (error) {
        logger.error('Failed to get account from cache:', error);
        throw error;
      }
    },

    // 缓存持仓信息
    async setPositions(userId: string, data: any): Promise<void> {
      try {
        const key = `user:positions:${userId}`;
        await this.redis?.setex(
          key,
          config.cache.user.positions,
          JSON.stringify(data)
        );
      } catch (error) {
        logger.error('Failed to cache positions:', error);
        throw error;
      }
    },

    // 获取持仓信息
    async getPositions(userId: string): Promise<any> {
      try {
        const key = `user:positions:${userId}`;
        const data = await this.redis?.get(key);
        return data ? JSON.parse(data) : null;
      } catch (error) {
        logger.error('Failed to get positions from cache:', error);
        throw error;
      }
    }
  };

  // 会话管理
  public session = {
    // 设置会话
    async set(sessionId: string, data: any, ttl: number): Promise<void> {
      try {
        const key = `session:${sessionId}`;
        await this.redis?.setex(key, ttl, JSON.stringify(data));
      } catch (error) {
        logger.error('Failed to set session:', error);
        throw error;
      }
    },

    // 获取会话
    async get(sessionId: string): Promise<any> {
      try {
        const key = `session:${sessionId}`;
        const data = await this.redis?.get(key);
        return data ? JSON.parse(data) : null;
      } catch (error) {
        logger.error('Failed to get session:', error);
        throw error;
      }
    },

    // 删除会话
    async delete(sessionId: string): Promise<void> {
      try {
        const key = `session:${sessionId}`;
        await this.redis?.del(key);
      } catch (error) {
        logger.error('Failed to delete session:', error);
        throw error;
      }
    }
  };

  // 限流控制
  public rateLimit = {
    // 增加计数
    async increment(key: string, ttl: number): Promise<number> {
      try {
        const count = await this.redis?.incr(key);
        if (count === 1) {
          await this.redis?.expire(key, ttl);
        }
        return count || 0;
      } catch (error) {
        logger.error('Failed to increment rate limit:', error);
        throw error;
      }
    },

    // 获取计数
    async get(key: string): Promise<number> {
      try {
        const count = await this.redis?.get(key);
        return parseInt(count || '0');
      } catch (error) {
        logger.error('Failed to get rate limit:', error);
        throw error;
      }
    }
  };

  // 清理过期缓存
  public async cleanup(): Promise<void> {
    try {
      // 实现缓存清理逻辑
      logger.info('Cache cleanup completed');
    } catch (error) {
      logger.error('Failed to cleanup cache:', error);
      throw error;
    }
  }
}

export const cacheManager = CacheManager.getInstance(); 