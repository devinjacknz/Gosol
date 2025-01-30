import mongoose from 'mongoose';
import Redis from 'ioredis';
import { Pool } from 'pg';
import { dbManager } from '../../lib/database';
import { config } from '../../config';

// Mock外部依赖
jest.mock('mongoose');
jest.mock('ioredis');
jest.mock('pg');

describe('DatabaseManager', () => {
  beforeEach(() => {
    // 清除所有mock
    jest.clearAllMocks();
  });

  describe('MongoDB Connection', () => {
    it('should connect to MongoDB successfully', async () => {
      const connectSpy = jest.spyOn(mongoose, 'connect');
      connectSpy.mockResolvedValueOnce(mongoose);

      const connection = await dbManager.connectMongo();
      
      expect(connectSpy).toHaveBeenCalledWith(
        config.database.mongodb.uri,
        config.database.mongodb.options
      );
      expect(connection).toBe(mongoose);
    });

    it('should handle MongoDB connection error', async () => {
      const error = new Error('Connection failed');
      const connectSpy = jest.spyOn(mongoose, 'connect');
      connectSpy.mockRejectedValueOnce(error);

      await expect(dbManager.connectMongo()).rejects.toThrow('Connection failed');
    });

    it('should reuse existing MongoDB connection', async () => {
      const connectSpy = jest.spyOn(mongoose, 'connect');
      connectSpy.mockResolvedValueOnce(mongoose);

      await dbManager.connectMongo();
      await dbManager.connectMongo();

      expect(connectSpy).toHaveBeenCalledTimes(1);
    });
  });

  describe('Redis Connection', () => {
    it('should connect to Redis successfully', async () => {
      const mockRedis = new Redis();
      const connection = await dbManager.connectRedis();

      expect(Redis).toHaveBeenCalledWith({
        host: config.database.redis.host,
        port: config.database.redis.port,
        password: config.database.redis.password,
        db: config.database.redis.db,
        retryStrategy: expect.any(Function)
      });
      expect(connection).toBeInstanceOf(Redis);
    });

    it('should handle Redis connection error', async () => {
      const error = new Error('Redis connection failed');
      jest.spyOn(Redis.prototype, 'on').mockImplementation((event, callback) => {
        if (event === 'error') {
          callback(error);
        }
        return mockRedis;
      });

      const mockRedis = new Redis();
      mockRedis.connect = jest.fn().mockRejectedValue(error);

      await expect(dbManager.connectRedis()).rejects.toThrow('Redis connection failed');
    });

    it('should reuse existing Redis connection', async () => {
      const constructorSpy = jest.spyOn(Redis.prototype, 'constructor');

      await dbManager.connectRedis();
      await dbManager.connectRedis();

      expect(constructorSpy).toHaveBeenCalledTimes(1);
    });
  });

  describe('TimescaleDB Connection', () => {
    it('should connect to TimescaleDB successfully', async () => {
      const mockPool = new Pool();
      const mockClient = {
        query: jest.fn().mockResolvedValue({}),
        release: jest.fn()
      };
      mockPool.connect = jest.fn().mockResolvedValue(mockClient);

      const connection = await dbManager.connectTimescale();

      expect(Pool).toHaveBeenCalledWith({
        host: config.database.timescaledb.host,
        port: config.database.timescaledb.port,
        database: config.database.timescaledb.database,
        user: config.database.timescaledb.username,
        password: config.database.timescaledb.password,
        max: 20,
        idleTimeoutMillis: 30000,
        connectionTimeoutMillis: 2000
      });
      expect(connection).toBeInstanceOf(Pool);
    });

    it('should handle TimescaleDB connection error', async () => {
      const error = new Error('TimescaleDB connection failed');
      const mockPool = new Pool();
      mockPool.connect = jest.fn().mockRejectedValue(error);

      await expect(dbManager.connectTimescale()).rejects.toThrow('TimescaleDB connection failed');
    });

    it('should reuse existing TimescaleDB connection', async () => {
      const constructorSpy = jest.spyOn(Pool.prototype, 'constructor');

      await dbManager.connectTimescale();
      await dbManager.connectTimescale();

      expect(constructorSpy).toHaveBeenCalledTimes(1);
    });
  });

  describe('Connection Management', () => {
    it('should close all connections successfully', async () => {
      const disconnectSpy = jest.spyOn(mongoose, 'disconnect');
      const redisQuitSpy = jest.spyOn(Redis.prototype, 'quit');
      const poolEndSpy = jest.spyOn(Pool.prototype, 'end');

      await dbManager.closeAll();

      expect(disconnectSpy).toHaveBeenCalled();
      expect(redisQuitSpy).toHaveBeenCalled();
      expect(poolEndSpy).toHaveBeenCalled();
    });

    it('should handle errors when closing connections', async () => {
      const error = new Error('Close failed');
      jest.spyOn(mongoose, 'disconnect').mockRejectedValueOnce(error);

      await expect(dbManager.closeAll()).rejects.toThrow('Close failed');
    });
  });

  describe('Health Check', () => {
    it('should return correct health status for all databases', async () => {
      // Mock successful connections
      jest.spyOn(mongoose.connection, 'readyState', 'get').mockReturnValue(1);
      jest.spyOn(Redis.prototype, 'ping').mockResolvedValue('PONG');
      const mockPool = new Pool();
      const mockClient = {
        query: jest.fn().mockResolvedValue({}),
        release: jest.fn()
      };
      mockPool.connect = jest.fn().mockResolvedValue(mockClient);

      const health = await dbManager.healthCheck();

      expect(health).toEqual({
        mongodb: true,
        redis: true,
        timescaledb: true
      });
    });

    it('should handle database health check errors', async () => {
      // Mock failed connections
      jest.spyOn(mongoose.connection, 'readyState', 'get').mockReturnValue(0);
      jest.spyOn(Redis.prototype, 'ping').mockRejectedValue(new Error());
      const mockPool = new Pool();
      mockPool.connect = jest.fn().mockRejectedValue(new Error());

      const health = await dbManager.healthCheck();

      expect(health).toEqual({
        mongodb: false,
        redis: false,
        timescaledb: false
      });
    });
  });
}); 