import { config } from '../config';

// 设置测试环境变量
process.env.NODE_ENV = 'test';

// 禁用控制台输出
global.console = {
  ...console,
  log: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  error: jest.fn(),
  debug: jest.fn(),
};

// 扩展Jest匹配器
expect.extend({
  toBeWithinRange(received: number, floor: number, ceiling: number) {
    const pass = received >= floor && received <= ceiling;
    if (pass) {
      return {
        message: () =>
          `expected ${received} not to be within range ${floor} - ${ceiling}`,
        pass: true,
      };
    } else {
      return {
        message: () =>
          `expected ${received} to be within range ${floor} - ${ceiling}`,
        pass: false,
      };
    }
  },
});

// 全局测试超时设置
jest.setTimeout(10000);

// 清理函数
afterAll(async () => {
  // 清理测试数据库
  // 关闭连接等
}); 