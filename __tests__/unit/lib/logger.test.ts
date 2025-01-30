import winston from 'winston';
import DailyRotateFile from 'winston-daily-rotate-file';
import { logger, requestLogger, errorLogger, auditLogger } from '../../lib/logger';

// Mock外部依赖
jest.mock('winston', () => ({
  createLogger: jest.fn().mockReturnValue({
    info: jest.fn(),
    error: jest.fn(),
    warn: jest.fn()
  }),
  format: {
    combine: jest.fn(),
    timestamp: jest.fn(),
    errors: jest.fn(),
    splat: jest.fn(),
    json: jest.fn(),
    colorize: jest.fn(),
    simple: jest.fn()
  },
  transports: {
    Console: jest.fn(),
  }
}));
jest.mock('winston-daily-rotate-file');

describe('Logger', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Logger Configuration', () => {
    it('should create logger with correct configuration', () => {
      expect(winston.createLogger).toHaveBeenCalled();
      expect(DailyRotateFile).toHaveBeenCalledTimes(3); // app, error, exceptions
    });
  });

  describe('Request Logger Middleware', () => {
    let mockReq: any;
    let mockRes: any;
    let mockNext: jest.Mock;

    beforeEach(() => {
      mockReq = {
        method: 'GET',
        originalUrl: '/api/test',
        ip: '127.0.0.1',
        get: jest.fn().mockReturnValue('test-agent'),
        user: { id: 'user123' }
      };
      mockRes = {
        statusCode: 200,
        on: jest.fn().mockImplementation((event, callback) => {
          if (event === 'finish') callback();
        })
      };
      mockNext = jest.fn();
    });

    it('should log successful requests', () => {
      requestLogger(mockReq, mockRes, mockNext);

      expect(mockNext).toHaveBeenCalled();
      expect(logger.info).toHaveBeenCalledWith('Request completed', expect.any(Object));
    });

    it('should log warning for 4xx status codes', () => {
      mockRes.statusCode = 400;
      requestLogger(mockReq, mockRes, mockNext);

      expect(logger.warn).toHaveBeenCalledWith('Request warning', expect.any(Object));
    });

    it('should log error for 5xx status codes', () => {
      mockRes.statusCode = 500;
      requestLogger(mockReq, mockRes, mockNext);

      expect(logger.error).toHaveBeenCalledWith('Request failed', expect.any(Object));
    });
  });

  describe('Error Logger Middleware', () => {
    it('should log errors with full context', () => {
      const error = new Error('Test error');
      const mockReq = {
        method: 'GET',
        originalUrl: '/api/test',
        headers: { 'user-agent': 'test-agent' },
        query: { test: 'value' },
        body: { data: 'test' },
        user: { id: 'user123' }
      };
      const mockRes = {};
      const mockNext = jest.fn();

      errorLogger(error, mockReq, mockRes, mockNext);

      expect(logger.error).toHaveBeenCalledWith('Unhandled error', expect.any(Object));
      expect(mockNext).toHaveBeenCalledWith(error);
    });
  });

  describe('Audit Logger', () => {
    describe('User Actions', () => {
      it('should log user login', () => {
        auditLogger.user.login('user123', '127.0.0.1');
        expect(logger.info).toHaveBeenCalledWith('Audit log', {
          action: 'user.login',
          details: { userId: 'user123', ip: '127.0.0.1' },
          timestamp: expect.any(String)
        });
      });

      it('should log user logout', () => {
        auditLogger.user.logout('user123');
        expect(logger.info).toHaveBeenCalledWith('Audit log', {
          action: 'user.logout',
          details: { userId: 'user123' },
          timestamp: expect.any(String)
        });
      });

      it('should log password change', () => {
        auditLogger.user.passwordChange('user123');
        expect(logger.info).toHaveBeenCalledWith('Audit log', {
          action: 'user.passwordChange',
          details: { userId: 'user123' },
          timestamp: expect.any(String)
        });
      });
    });

    describe('Trading Actions', () => {
      it('should log order creation', () => {
        const orderDetails = { symbol: 'BTC/USDT', amount: 1 };
        auditLogger.trading.orderCreated('user123', 'order123', orderDetails);
        expect(logger.info).toHaveBeenCalledWith('Audit log', {
          action: 'trading.orderCreated',
          details: { userId: 'user123', orderId: 'order123', details: orderDetails },
          timestamp: expect.any(String)
        });
      });

      it('should log order cancellation', () => {
        auditLogger.trading.orderCancelled('user123', 'order123', 'user_requested');
        expect(logger.info).toHaveBeenCalledWith('Audit log', {
          action: 'trading.orderCancelled',
          details: { userId: 'user123', orderId: 'order123', reason: 'user_requested' },
          timestamp: expect.any(String)
        });
      });

      it('should log position closure', () => {
        const positionDetails = { pnl: 100 };
        auditLogger.trading.positionClosed('user123', 'position123', positionDetails);
        expect(logger.info).toHaveBeenCalledWith('Audit log', {
          action: 'trading.positionClosed',
          details: { userId: 'user123', positionId: 'position123', details: positionDetails },
          timestamp: expect.any(String)
        });
      });
    });

    describe('Finance Actions', () => {
      it('should log deposit', () => {
        auditLogger.finance.deposit('user123', 1000, 'USDT');
        expect(logger.info).toHaveBeenCalledWith('Audit log', {
          action: 'finance.deposit',
          details: { userId: 'user123', amount: 1000, currency: 'USDT' },
          timestamp: expect.any(String)
        });
      });

      it('should log withdrawal', () => {
        auditLogger.finance.withdrawal('user123', 500, 'BTC');
        expect(logger.info).toHaveBeenCalledWith('Audit log', {
          action: 'finance.withdrawal',
          details: { userId: 'user123', amount: 500, currency: 'BTC' },
          timestamp: expect.any(String)
        });
      });

      it('should log transfer', () => {
        auditLogger.finance.transfer('user123', 'spot', 'futures', 1000);
        expect(logger.info).toHaveBeenCalledWith('Audit log', {
          action: 'finance.transfer',
          details: { userId: 'user123', from: 'spot', to: 'futures', amount: 1000 },
          timestamp: expect.any(String)
        });
      });
    });
  });
}); 