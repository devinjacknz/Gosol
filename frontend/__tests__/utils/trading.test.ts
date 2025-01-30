import {
  calculateOrderValue,
  formatPrice,
  formatAmount,
  calculatePnL,
  validateOrder,
  calculateLeverage,
  estimateSlippage,
  calculateLiquidationPrice,
} from '@/utils/trading';

describe('Trading Utilities', () => {
  describe('calculateOrderValue', () => {
    it('calculates correct order value', () => {
      expect(calculateOrderValue(50000, 1.5)).toBe(75000);
      expect(calculateOrderValue(3000, 10)).toBe(30000);
    });

    it('handles zero values', () => {
      expect(calculateOrderValue(50000, 0)).toBe(0);
      expect(calculateOrderValue(0, 1.5)).toBe(0);
    });

    it('handles decimal precision', () => {
      expect(calculateOrderValue(50000.123, 1.123456)).toBeCloseTo(56175.138);
    });
  });

  describe('formatPrice', () => {
    it('formats price with correct decimals', () => {
      expect(formatPrice(50000.123456)).toBe('50,000.12');
      expect(formatPrice(3000.5)).toBe('3,000.50');
    });

    it('handles zero and negative values', () => {
      expect(formatPrice(0)).toBe('0.00');
      expect(formatPrice(-1000)).toBe('-1,000.00');
    });

    it('formats different currencies correctly', () => {
      expect(formatPrice(50000.123456, 'USD')).toBe('$50,000.12');
      expect(formatPrice(50000.123456, 'EUR')).toBe('€50,000.12');
    });
  });

  describe('formatAmount', () => {
    it('formats amount with correct decimals', () => {
      expect(formatAmount(1.123456)).toBe('1.1235');
      expect(formatAmount(0.000123)).toBe('0.0001');
    });

    it('handles zero values', () => {
      expect(formatAmount(0)).toBe('0.0000');
    });

    it('handles different decimal places', () => {
      expect(formatAmount(1.123456, 2)).toBe('1.12');
      expect(formatAmount(1.123456, 6)).toBe('1.123456');
    });
  });

  describe('calculatePnL', () => {
    it('calculates profit correctly', () => {
      const position = {
        entryPrice: 50000,
        amount: 1.5,
        side: 'long',
        currentPrice: 55000,
      };
      expect(calculatePnL(position)).toBe(7500); // (55000 - 50000) * 1.5
    });

    it('calculates loss correctly', () => {
      const position = {
        entryPrice: 50000,
        amount: 1.5,
        side: 'long',
        currentPrice: 45000,
      };
      expect(calculatePnL(position)).toBe(-7500); // (45000 - 50000) * 1.5
    });

    it('handles short positions', () => {
      const position = {
        entryPrice: 50000,
        amount: 1.5,
        side: 'short',
        currentPrice: 45000,
      };
      expect(calculatePnL(position)).toBe(7500); // (50000 - 45000) * 1.5
    });
  });

  describe('validateOrder', () => {
    it('validates valid order', () => {
      const order = {
        symbol: 'BTC/USDT',
        side: 'buy',
        type: 'limit',
        price: 50000,
        amount: 1.5,
      };
      expect(validateOrder(order)).toEqual({ isValid: true });
    });

    it('detects missing required fields', () => {
      const order = {
        symbol: 'BTC/USDT',
        side: 'buy',
        type: 'limit',
        // missing price
        amount: 1.5,
      };
      expect(validateOrder(order)).toEqual({
        isValid: false,
        errors: ['Price is required for limit orders'],
      });
    });

    it('validates price and amount ranges', () => {
      const order = {
        symbol: 'BTC/USDT',
        side: 'buy',
        type: 'limit',
        price: -50000,
        amount: 0,
      };
      expect(validateOrder(order)).toEqual({
        isValid: false,
        errors: ['Price must be positive', 'Amount must be positive'],
      });
    });
  });

  describe('calculateLeverage', () => {
    it('calculates leverage correctly', () => {
      const position = {
        value: 100000,
        margin: 10000,
      };
      expect(calculateLeverage(position)).toBe(10);
    });

    it('handles zero margin', () => {
      const position = {
        value: 100000,
        margin: 0,
      };
      expect(calculateLeverage(position)).toBe(Infinity);
    });

    it('handles zero position value', () => {
      const position = {
        value: 0,
        margin: 10000,
      };
      expect(calculateLeverage(position)).toBe(0);
    });
  });

  describe('estimateSlippage', () => {
    it('estimates slippage based on order book', () => {
      const orderBook = {
        bids: [
          [49900, 1.0],
          [49800, 2.0],
        ],
        asks: [
          [50100, 1.0],
          [50200, 2.0],
        ],
      };
      
      expect(estimateSlippage(orderBook, 'buy', 1.5, 50000)).toBeCloseTo(0.003); // 0.3%
    });

    it('handles insufficient liquidity', () => {
      const orderBook = {
        bids: [[49900, 0.5]],
        asks: [[50100, 0.5]],
      };
      
      expect(estimateSlippage(orderBook, 'buy', 2.0, 50000)).toBeGreaterThan(0.01); // >1%
    });
  });

  describe('calculateLiquidationPrice', () => {
    it('calculates long position liquidation price', () => {
      const position = {
        entryPrice: 50000,
        leverage: 10,
        maintenanceMargin: 0.005,
        side: 'long',
      };
      
      // 预期清算价格应该在入场价格下方
      const liquidationPrice = calculateLiquidationPrice(position);
      expect(liquidationPrice).toBeLessThan(position.entryPrice);
    });

    it('calculates short position liquidation price', () => {
      const position = {
        entryPrice: 50000,
        leverage: 10,
        maintenanceMargin: 0.005,
        side: 'short',
      };
      
      // 预期清算价格应该在入场价格上方
      const liquidationPrice = calculateLiquidationPrice(position);
      expect(liquidationPrice).toBeGreaterThan(position.entryPrice);
    });

    it('handles different leverage values', () => {
      const position = {
        entryPrice: 50000,
        leverage: 20,
        maintenanceMargin: 0.005,
        side: 'long',
      };
      
      const liquidationPrice1 = calculateLiquidationPrice(position);
      
      position.leverage = 5;
      const liquidationPrice2 = calculateLiquidationPrice(position);
      
      // 更高的杠杆应该有更高的清算风险
      expect(liquidationPrice1).toBeGreaterThan(liquidationPrice2);
    });
  });
}); 