'use client';

import { useTradingStore } from '@/store';
import { formatCurrency, formatPercent } from '@/utils/format';
import {
  ArrowTrendingUpIcon,
  ArrowTrendingDownIcon,
  ScaleIcon,
  BanknotesIcon,
} from '@heroicons/react/24/outline';

export default function AccountSummary() {
  const { accountInfo } = useTradingStore();

  const stats = [
    {
      name: '总权益',
      value: accountInfo?.totalEquity || 0,
      change: ((accountInfo?.totalEquity || 0) - (accountInfo?.totalEquity || 0)) / (accountInfo?.totalEquity || 1),
      icon: BanknotesIcon,
    },
    {
      name: '已用保证金',
      value: accountInfo?.usedMargin || 0,
      change: null,
      icon: ScaleIcon,
    },
    {
      name: '未实现盈亏',
      value: accountInfo?.unrealizedPnL || 0,
      change: null,
      icon: ArrowTrendingUpIcon,
      highlight: true,
    },
    {
      name: '已实现盈亏',
      value: accountInfo?.realizedPnL || 0,
      change: null,
      icon: ArrowTrendingDownIcon,
      highlight: true,
    },
  ];

  return (
    <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
      {stats.map((item) => (
        <div
          key={item.name}
          className="bg-white rounded-lg p-6 flex items-start space-x-4"
        >
          <div className="rounded-lg bg-blue-50 p-3">
            <item.icon className="h-6 w-6 text-blue-600" />
          </div>
          <div>
            <p className="text-sm text-gray-500">{item.name}</p>
            <p className={`text-2xl font-semibold ${
              item.highlight
                ? item.value > 0
                  ? 'text-green-500'
                  : item.value < 0
                  ? 'text-red-500'
                  : 'text-gray-900'
                : 'text-gray-900'
            }`}>
              {formatCurrency(item.value)}
            </p>
            {item.change !== null && (
              <p className={`text-sm ${
                item.change > 0
                  ? 'text-green-500'
                  : item.change < 0
                  ? 'text-red-500'
                  : 'text-gray-500'
              }`}>
                {item.change > 0 ? '↑' : '↓'} {formatPercent(Math.abs(item.change))}
              </p>
            )}
          </div>
        </div>
      ))}
    </div>
  );
} 