import winston from 'winston';
import DailyRotateFile from 'winston-daily-rotate-file';
import { config } from '../config';

// 定义日志格式
const logFormat = winston.format.combine(
  winston.format.timestamp({
    format: 'YYYY-MM-DD HH:mm:ss'
  }),
  winston.format.errors({ stack: true }),
  winston.format.splat(),
  winston.format.json()
);

// 创建日志传输器
const transports = [
  // 控制台输出
  new winston.transports.Console({
    format: winston.format.combine(
      winston.format.colorize(),
      winston.format.simple()
    )
  }),

  // 普通日志文件
  new DailyRotateFile({
    filename: 'logs/app-%DATE%.log',
    datePattern: 'YYYY-MM-DD',
    zippedArchive: true,
    maxSize: '20m',
    maxFiles: '14d',
    level: 'info'
  }),

  // 错误日志文件
  new DailyRotateFile({
    filename: 'logs/error-%DATE%.log',
    datePattern: 'YYYY-MM-DD',
    zippedArchive: true,
    maxSize: '20m',
    maxFiles: '30d',
    level: 'error'
  })
];

// 创建日志记录器
export const logger = winston.createLogger({
  level: config.monitoring.logLevel,
  format: logFormat,
  transports,
  // 异常处理
  exceptionHandlers: [
    new DailyRotateFile({
      filename: 'logs/exceptions-%DATE%.log',
      datePattern: 'YYYY-MM-DD',
      zippedArchive: true,
      maxSize: '20m',
      maxFiles: '30d'
    })
  ],
  // 退出处理
  exitOnError: false
});

// 创建请求日志中间件
export const requestLogger = (req: any, res: any, next: any) => {
  const start = Date.now();

  // 响应结束时记录日志
  res.on('finish', () => {
    const duration = Date.now() - start;
    const logData = {
      method: req.method,
      url: req.originalUrl,
      status: res.statusCode,
      duration: duration,
      ip: req.ip,
      userAgent: req.get('user-agent'),
      userId: req.user?.id
    };

    // 根据状态码决定日志级别
    if (res.statusCode >= 500) {
      logger.error('Request failed', logData);
    } else if (res.statusCode >= 400) {
      logger.warn('Request warning', logData);
    } else {
      logger.info('Request completed', logData);
    }
  });

  next();
};

// 创建错误日志中间件
export const errorLogger = (err: any, req: any, res: any, next: any) => {
  logger.error('Unhandled error', {
    error: {
      message: err.message,
      stack: err.stack
    },
    request: {
      method: req.method,
      url: req.originalUrl,
      headers: req.headers,
      query: req.query,
      body: req.body
    },
    user: req.user
  });

  next(err);
};

// 创建审计日志记录器
export const auditLogger = {
  log: (action: string, details: any) => {
    logger.info('Audit log', {
      action,
      details,
      timestamp: new Date().toISOString()
    });
  },

  // 用户相关审计
  user: {
    login: (userId: string, ip: string) => {
      auditLogger.log('user.login', { userId, ip });
    },
    logout: (userId: string) => {
      auditLogger.log('user.logout', { userId });
    },
    passwordChange: (userId: string) => {
      auditLogger.log('user.passwordChange', { userId });
    }
  },

  // 交易相关审计
  trading: {
    orderCreated: (userId: string, orderId: string, details: any) => {
      auditLogger.log('trading.orderCreated', { userId, orderId, details });
    },
    orderCancelled: (userId: string, orderId: string, reason: string) => {
      auditLogger.log('trading.orderCancelled', { userId, orderId, reason });
    },
    positionClosed: (userId: string, positionId: string, details: any) => {
      auditLogger.log('trading.positionClosed', { userId, positionId, details });
    }
  },

  // 资金相关审计
  finance: {
    deposit: (userId: string, amount: number, currency: string) => {
      auditLogger.log('finance.deposit', { userId, amount, currency });
    },
    withdrawal: (userId: string, amount: number, currency: string) => {
      auditLogger.log('finance.withdrawal', { userId, amount, currency });
    },
    transfer: (userId: string, from: string, to: string, amount: number) => {
      auditLogger.log('finance.transfer', { userId, from, to, amount });
    }
  }
}; 