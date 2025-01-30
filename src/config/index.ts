import dotenv from 'dotenv';

// 加载环境变量
dotenv.config();

export const config = {
  // 服务配置
  server: {
    port: process.env.PORT || 3000,
    env: process.env.NODE_ENV || 'development',
    apiVersion: process.env.API_VERSION || 'v1',
    corsOrigins: process.env.CORS_ORIGINS?.split(',') || ['http://localhost:3000']
  },

  // 数据库配置
  database: {
    // MongoDB配置
    mongodb: {
      uri: process.env.MONGODB_URI || 'mongodb://localhost:27017/trading',
      options: {
        useNewUrlParser: true,
        useUnifiedTopology: true,
        maxPoolSize: 10,
        serverSelectionTimeoutMS: 5000,
        socketTimeoutMS: 45000,
      }
    },
    // Redis配置
    redis: {
      host: process.env.REDIS_HOST || 'localhost',
      port: parseInt(process.env.REDIS_PORT || '6379'),
      password: process.env.REDIS_PASSWORD,
      db: parseInt(process.env.REDIS_DB || '0')
    },
    // TimescaleDB配置
    timescaledb: {
      host: process.env.TIMESCALEDB_HOST || 'localhost',
      port: parseInt(process.env.TIMESCALEDB_PORT || '5432'),
      database: process.env.TIMESCALEDB_DATABASE || 'trading',
      username: process.env.TIMESCALEDB_USER || 'postgres',
      password: process.env.TIMESCALEDB_PASSWORD
    }
  },

  // 缓存配置
  cache: {
    // 市场数据缓存时间(秒)
    market: {
      ticker: 60,
      orderBook: 1,
      recentTrades: 5,
      kline: 300
    },
    // 用户数据缓存时间(秒)
    user: {
      account: 5,
      positions: 3,
      orders: 10
    }
  },

  // 安全配置
  security: {
    // JWT配置
    jwt: {
      secret: process.env.JWT_SECRET || 'your-secret-key',
      expiresIn: process.env.JWT_EXPIRES_IN || '24h'
    },
    // 密码加密
    crypto: {
      iterations: 10000,
      keylen: 64,
      digest: 'sha512'
    },
    // 限流配置
    rateLimit: {
      window: 60000, // 1分钟
      max: 100 // 最大请求数
    }
  },

  // 交易配置
  trading: {
    // 风控限制
    riskControl: {
      maxLeverage: 20,
      maxPositionSize: 1000000,
      minMarginLevel: 0.1,
      maxDailyWithdrawal: 100000
    },
    // 手续费率
    fees: {
      maker: 0.001,
      taker: 0.002,
      withdrawal: 0.0005
    },
    // 订单限制
    orderLimits: {
      maxOpenOrders: 200,
      minOrderSize: 10,
      priceDeviation: 0.1 // 价格偏离限制
    }
  },

  // 监控配置
  monitoring: {
    // 日志级别
    logLevel: process.env.LOG_LEVEL || 'info',
    // 性能监控
    metrics: {
      enabled: true,
      interval: 60000 // 收集间隔(毫秒)
    },
    // 告警阈值
    alerts: {
      highLatency: 1000, // 毫秒
      errorRate: 0.01, // 1%
      systemLoad: 0.8 // 80%
    }
  },

  // 外部服务配置
  services: {
    // 邮件服务
    email: {
      host: process.env.EMAIL_HOST,
      port: parseInt(process.env.EMAIL_PORT || '587'),
      secure: process.env.EMAIL_SECURE === 'true',
      auth: {
        user: process.env.EMAIL_USER,
        pass: process.env.EMAIL_PASS
      }
    },
    // 短信服务
    sms: {
      provider: process.env.SMS_PROVIDER,
      apiKey: process.env.SMS_API_KEY,
      apiSecret: process.env.SMS_API_SECRET
    }
  },

  // 系统维护
  maintenance: {
    // 自动备份
    backup: {
      enabled: true,
      interval: '0 0 * * *', // 每天凌晨
      retention: 30 // 保留天数
    },
    // 清理任务
    cleanup: {
      enabled: true,
      interval: '0 1 * * *', // 每天凌晨1点
      olderthan: 90 // 清理90天前的数据
    }
  }
}; 