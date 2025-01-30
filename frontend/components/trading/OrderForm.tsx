'use client';

import { useState } from 'react';
import { Tab } from '@headlessui/react';
import { useTradingStore } from '@/store';
import { formatNumber, formatCurrency } from '@/utils/format';

interface OrderFormProps {
  symbol: string;
  type: 'limit' | 'market';
  onTypeChange: (type: 'limit' | 'market') => void;
}

export default function OrderForm({ symbol, type, onTypeChange }: OrderFormProps) {
  const { marketData, accountInfo } = useTradingStore();
  const [direction, setDirection] = useState<'buy' | 'sell'>('buy');
  const [price, setPrice] = useState('');
  const [amount, setAmount] = useState('');
  const [leverage, setLeverage] = useState(1);
  const [marginType, setMarginType] = useState<'isolated' | 'cross'>('isolated');

  const currentPrice = marketData[symbol]?.price || 0;
  const availableBalance = accountInfo?.availableBalance || 0;

  // 计算订单价值
  const orderValue = Number(price || currentPrice) * Number(amount || 0);
  
  // 计算所需保证金
  const requiredMargin = orderValue / leverage;

  // 计算预估强平价格
  const liquidationPrice = direction === 'buy'
    ? Number(price || currentPrice) * (1 - 1/leverage + 0.005)
    : Number(price || currentPrice) * (1 + 1/leverage - 0.005);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    // 这里处理下单逻辑
    console.log({
      symbol,
      type,
      direction,
      price: type === 'limit' ? Number(price) : currentPrice,
      amount: Number(amount),
      leverage,
      marginType,
    });
  };

  return (
    <div>
      {/* 订单类型选择 */}
      <div className="mb-4">
        <Tab.Group>
          <Tab.List className="flex space-x-1 rounded-xl bg-gray-100 p-1">
            <Tab
              className={({ selected }) =>
                `w-full rounded-lg py-2.5 text-sm font-medium leading-5
                ${selected
                  ? 'bg-white text-blue-600 shadow'
                  : 'text-gray-600 hover:bg-white/[0.12] hover:text-blue-600'
                }`
              }
              onClick={() => onTypeChange('limit')}
            >
              限价单
            </Tab>
            <Tab
              className={({ selected }) =>
                `w-full rounded-lg py-2.5 text-sm font-medium leading-5
                ${selected
                  ? 'bg-white text-blue-600 shadow'
                  : 'text-gray-600 hover:bg-white/[0.12] hover:text-blue-600'
                }`
              }
              onClick={() => onTypeChange('market')}
            >
              市价单
            </Tab>
          </Tab.List>
        </Tab.Group>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        {/* 买卖方向 */}
        <div className="grid grid-cols-2 gap-2">
          <button
            type="button"
            className={`py-2 rounded-lg font-medium ${
              direction === 'buy'
                ? 'bg-green-500 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            }`}
            onClick={() => setDirection('buy')}
          >
            买入做多
          </button>
          <button
            type="button"
            className={`py-2 rounded-lg font-medium ${
              direction === 'sell'
                ? 'bg-red-500 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            }`}
            onClick={() => setDirection('sell')}
          >
            卖出做空
          </button>
        </div>

        {/* 杠杆和保证金类型 */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              杠杆倍数
            </label>
            <select
              value={leverage}
              onChange={(e) => setLeverage(Number(e.target.value))}
              className="w-full rounded-lg border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
            >
              {[1, 2, 3, 5, 10].map((x) => (
                <option key={x} value={x}>{x}x</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              保证金模式
            </label>
            <select
              value={marginType}
              onChange={(e) => setMarginType(e.target.value as 'isolated' | 'cross')}
              className="w-full rounded-lg border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
            >
              <option value="isolated">逐仓</option>
              <option value="cross">全仓</option>
            </select>
          </div>
        </div>

        {/* 价格输入 */}
        {type === 'limit' && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              价格
            </label>
            <div className="relative">
              <input
                type="number"
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                className="w-full rounded-lg border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                placeholder={`${currentPrice}`}
                step="0.1"
                min="0"
              />
              <span className="absolute right-3 top-2 text-gray-500">USDT</span>
            </div>
          </div>
        )}

        {/* 数量输入 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            数量
          </label>
          <div className="relative">
            <input
              type="number"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              className="w-full rounded-lg border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
              placeholder="0.00"
              step="0.001"
              min="0"
            />
            <span className="absolute right-3 top-2 text-gray-500">
              {symbol.split('/')[0]}
            </span>
          </div>
        </div>

        {/* 订单信息 */}
        <div className="bg-gray-50 rounded-lg p-4 space-y-2 text-sm">
          <div className="flex justify-between">
            <span className="text-gray-600">订单价值</span>
            <span className="font-medium">{formatCurrency(orderValue)}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-gray-600">所需保证金</span>
            <span className="font-medium">{formatCurrency(requiredMargin)}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-gray-600">强平价格</span>
            <span className="font-medium">{formatNumber(liquidationPrice)}</span>
          </div>
        </div>

        {/* 提交按钮 */}
        <button
          type="submit"
          className={`w-full py-3 rounded-lg font-medium text-white ${
            direction === 'buy'
              ? 'bg-green-500 hover:bg-green-600'
              : 'bg-red-500 hover:bg-red-600'
          }`}
        >
          {direction === 'buy' ? '买入做多' : '卖出做空'} {symbol}
        </button>
      </form>
    </div>
  );
} 