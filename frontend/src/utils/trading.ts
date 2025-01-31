/**
 * 计算订单价值
 */
export function calculateOrderValue(price: number, amount: number): number {
  return price * amount;
}

/**
 * 格式化价格
 */
export function formatPrice(price: number, currency: string = ''): string {
  const formatted = new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(price);

  return currency ? `${currency === 'USD' ? '$' : '€'}${formatted}` : formatted;
}

/**
 * 格式化数量
 */
export function formatAmount(amount: number, decimals: number = 4): string {
  return amount.toFixed(decimals);
}

/**
 * 计算盈亏
 */
export function calculatePnL(position: {
  entryPrice: number;
  amount: number;
  side: 'long' | 'short';
  currentPrice: number;
}): number {
  const { entryPrice, amount, side, currentPrice } = position;
  const priceDiff = side === 'long' ? currentPrice - entryPrice : entryPrice - currentPrice;
  return priceDiff * amount;
}

/**
 * 验证订单
 */
export function validateOrder(order: {
  symbol: string;
  side: 'buy' | 'sell';
  type: 'market' | 'limit';
  price?: number;
  amount: number;
}): { isValid: boolean; errors?: string[] } {
  const errors: string[] = [];

  if (order.type === 'limit' && !order.price) {
    errors.push('Price is required for limit orders');
  }

  if (order.price && order.price <= 0) {
    errors.push('Price must be positive');
  }

  if (order.amount <= 0) {
    errors.push('Amount must be positive');
  }

  return {
    isValid: errors.length === 0,
    errors: errors.length > 0 ? errors : undefined,
  };
}

/**
 * 计算杠杆
 */
export function calculateLeverage(position: {
  value: number;
  margin: number;
}): number {
  const { value, margin } = position;
  return margin === 0 ? Infinity : value / margin;
}

/**
 * 估算滑点
 */
export function estimateSlippage(
  orderBook: {
    bids: [number, number][];
    asks: [number, number][];
  },
  side: 'buy' | 'sell',
  amount: number,
  marketPrice: number
): number {
  const levels = side === 'buy' ? orderBook.asks : orderBook.bids;
  let remainingAmount = amount;
  let totalCost = 0;

  for (const [price, quantity] of levels) {
    const fillAmount = Math.min(remainingAmount, quantity);
    totalCost += fillAmount * price;
    remainingAmount -= fillAmount;

    if (remainingAmount <= 0) break;
  }

  const avgPrice = totalCost / amount;
  return Math.abs(avgPrice - marketPrice) / marketPrice;
}

/**
 * 计算清算价格
 */
export function calculateLiquidationPrice(position: {
  entryPrice: number;
  leverage: number;
  maintenanceMargin: number;
  side: 'long' | 'short';
}): number {
  const { entryPrice, leverage, maintenanceMargin, side } = position;
  const maintenanceAmount = entryPrice * maintenanceMargin;
  const leverageEffect = 1 / leverage;

  if (side === 'long') {
    return entryPrice * (1 - leverageEffect - maintenanceMargin);
  } else {
    return entryPrice * (1 + leverageEffect + maintenanceMargin);
  }
} 