import mongoose from 'mongoose';
import Redis from 'ioredis';
import { Pool } from 'pg';
import { config } from '../config';
import { logger } from './logger';

class DatabaseManager {
  private static instance: DatabaseManager;
  private mongoConnection: typeof mongoose | null = null;
  private redisClient: Redis | null = null;
  private timescaleClient: Pool | null = null;

  private constructor() {}

  public static getInstance(): DatabaseManager {
    if (!DatabaseManager.instance) {
      DatabaseManager.instance = new DatabaseManager();
    }
    return DatabaseManager.instance;
  }

  // MongoDB连接
  public async connectMongo(): Promise<typeof mongoose> {
    try {
      if (!this.mongoConnection) {
        this.mongoConnection = await mongoose.connect(
          config.database.mongodb.uri,
          config.database.mongodb.options
        );

        mongoose.connection.on('error', (error) => {
          logger.error('MongoDB connection error:', error);
        });

        mongoose.connection.on('disconnected', () => {
          logger.warn('MongoDB disconnected. Attempting to reconnect...');
          this.connectMongo();
        });

        logger.info('MongoDB connected successfully');
      }
      return this.mongoConnection;
    } catch (error) {
      logger.error('Failed to connect to MongoDB:', error);
      throw error;
    }
  }

  // Redis连接
  public async connectRedis(): Promise<Redis> {
    try {
      if (!this.redisClient) {
        this.redisClient = new Redis({
          host: config.database.redis.host,
          port: config.database.redis.port,
          password: config.database.redis.password,
          db: config.database.redis.db,
          retryStrategy: (times) => {
            const delay = Math.min(times * 50, 2000);
            return delay;
          }
        });

        this.redisClient.on('error', (error) => {
          logger.error('Redis connection error:', error);
        });

        this.redisClient.on('connect', () => {
          logger.info('Redis connected successfully');
        });
      }
      return this.redisClient;
    } catch (error) {
      logger.error('Failed to connect to Redis:', error);
      throw error;
    }
  }

  // TimescaleDB连接
  public async connectTimescale(): Promise<Pool> {
    try {
      if (!this.timescaleClient) {
        this.timescaleClient = new Pool({
          host: config.database.timescaledb.host,
          port: config.database.timescaledb.port,
          database: config.database.timescaledb.database,
          user: config.database.timescaledb.username,
          password: config.database.timescaledb.password,
          max: 20, // 连接池最大连接数
          idleTimeoutMillis: 30000,
          connectionTimeoutMillis: 2000,
        });

        // 测试连接
        const client = await this.timescaleClient.connect();
        client.release();
        
        logger.info('TimescaleDB connected successfully');
      }
      return this.timescaleClient;
    } catch (error) {
      logger.error('Failed to connect to TimescaleDB:', error);
      throw error;
    }
  }

  // 关闭所有数据库连接
  public async closeAll(): Promise<void> {
    try {
      if (this.mongoConnection) {
        await mongoose.disconnect();
        this.mongoConnection = null;
      }

      if (this.redisClient) {
        await this.redisClient.quit();
        this.redisClient = null;
      }

      if (this.timescaleClient) {
        await this.timescaleClient.end();
        this.timescaleClient = null;
      }

      logger.info('All database connections closed successfully');
    } catch (error) {
      logger.error('Error closing database connections:', error);
      throw error;
    }
  }

  // 健康检查
  public async healthCheck(): Promise<{
    mongodb: boolean;
    redis: boolean;
    timescaledb: boolean;
  }> {
    const health = {
      mongodb: false,
      redis: false,
      timescaledb: false
    };

    try {
      // 检查MongoDB
      if (this.mongoConnection && mongoose.connection.readyState === 1) {
        health.mongodb = true;
      }

      // 检查Redis
      if (this.redisClient) {
        const pingResult = await this.redisClient.ping();
        health.redis = pingResult === 'PONG';
      }

      // 检查TimescaleDB
      if (this.timescaleClient) {
        const client = await this.timescaleClient.connect();
        await client.query('SELECT 1');
        client.release();
        health.timescaledb = true;
      }
    } catch (error) {
      logger.error('Health check failed:', error);
    }

    return health;
  }
}

export const dbManager = DatabaseManager.getInstance(); 