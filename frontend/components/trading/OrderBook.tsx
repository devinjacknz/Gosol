'use client';

import { useState } from 'react';
import { formatNumber } from '@/utils/format';

interface OrderBookProps {
  symbol: string;
}

interface OrderBookLevel {
  price: number;
  amount: number;
  total: number;
  percentage: number;
}

// 模拟数据
const mockOrderBook = {
  asks: Array.from({ length: 15 }, (_, i) => ({
    price: 50000 + i * 10,
    amount: Math.random() * 2,
    total: 0,
    percentage: 0,
  })),
  bids: Array.from({ length: 15 }, (_, i) => ({
    price: 49990 - i * 10,
    amount: Math.random() * 2,
    total: 0,
    percentage: 0,
  })),
};

// 计算累计数量和百分比
const calculateTotals = (levels: OrderBookLevel[]): OrderBookLevel[] => {
  let maxTotal = 0;
  const result = levels.map((level, index, array) => {
    const total = array
      .slice(0, index + 1)
      .reduce((sum, item) => sum + item.amount, 0);
    maxTotal = Math.max(maxTotal, total);
    return { ...level, total };
  });

  return result.map(level => ({
    ...level,
    percentage: (level.total / maxTotal) * 100,
  }));
};

mockOrderBook.asks = calculateTotals(mockOrderBook.asks.reverse()).reverse();
mockOrderBook.bids = calculateTotals(mockOrderBook.bids);

export default function OrderBook({ symbol }: OrderBookProps) {
  const [precision, setPrecision] = useState(1);

  return (
    <div className="bg-white rounded-lg p-4">
      <div className="flex items-center justify-between mb-4">
        <h3 className="font-medium">订单簿</h3>
        <div className="flex space-x-1">
          <button
            onClick={() => setPrecision(0.1)}
            className={`px-2 py-1 text-xs rounded ${
              precision === 0.1 ? 'bg-blue-500 text-white' : 'bg-gray-100'
            }`}
          >
            0.1
          </button>
          <button
            onClick={() => setPrecision(1)}
            className={`px-2 py-1 text-xs rounded ${
              precision === 1 ? 'bg-blue-500 text-white' : 'bg-gray-100'
            }`}
          >
            1.0
          </button>
          <button
            onClick={() => setPrecision(10)}
            className={`px-2 py-1 text-xs rounded ${
              precision === 10 ? 'bg-blue-500 text-white' : 'bg-gray-100'
            }`}
          >
            10.0
          </button>
        </div>
      </div>

      <div className="text-sm">
        {/* 列标题 */}
        <div className="grid grid-cols-3 text-gray-500 mb-2">
          <div>价格(USDT)</div>
          <div className="text-right">数量(BTC)</div>
          <div className="text-right">累计(BTC)</div>
        </div>

        {/* 卖单 */}
        <div className="space-y-1">
          {mockOrderBook.asks.map((level) => (
            <div
              key={level.price}
              className="grid grid-cols-3 relative"
              style={{ height: '20px' }}
            >
              <div
                className="absolute inset-0 bg-red-50"
                style={{ width: `${level.percentage}%`, zIndex: 0 }}
              />
              <div className="text-red-500 z-10">{formatNumber(level.price)}</div>
              <div className="text-right z-10">{formatNumber(level.amount, 4)}</div>
              <div className="text-right text-gray-500 z-10">
                {formatNumber(level.total, 4)}
              </div>
            </div>
          ))}
        </div>

        {/* 最新价格 */}
        <div className="text-center py-2 font-medium text-lg">
          50000.00
          <span className="text-green-500 text-sm ml-2">+1.2%</span>
        </div>

        {/* 买单 */}
        <div className="space-y-1">
          {mockOrderBook.bids.map((level) => (
            <div
              key={level.price}
              className="grid grid-cols-3 relative"
              style={{ height: '20px' }}
            >
              <div
                className="absolute inset-0 bg-green-50"
                style={{ width: `${level.percentage}%`, zIndex: 0 }}
              />
              <div className="text-green-500 z-10">{formatNumber(level.price)}</div>
              <div className="text-right z-10">{formatNumber(level.amount, 4)}</div>
              <div className="text-right text-gray-500 z-10">
                {formatNumber(level.total, 4)}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
} 