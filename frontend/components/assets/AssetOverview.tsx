'use client';

import { useTradingStore } from '@/store';
import { formatCurrency, formatPercent } from '@/utils/format';
import {
  ArrowTrendingUpIcon,
  ArrowTrendingDownIcon,
  ScaleIcon,
  BanknotesIcon,
  ChartBarIcon,
  ShieldCheckIcon,
} from '@heroicons/react/24/outline';

export default function AssetOverview() {
  const { accountInfo } = useTradingStore();

  const stats = [
    {
      name: '总资产',
      value: accountInfo?.totalEquity || 0,
      icon: BanknotesIcon,
      description: '包含已实现和未实现盈亏',
    },
    {
      name: '可用余额',
      value: accountInfo?.availableBalance || 0,
      icon: ChartBarIcon,
      description: '可用于开仓的资金',
    },
    {
      name: '已用保证金',
      value: accountInfo?.usedMargin || 0,
      icon: ScaleIcon,
      description: '当前持仓占用的保证金',
    },
    {
      name: '保证金率',
      value: accountInfo?.marginLevel || 0,
      icon: ShieldCheckIcon,
      description: '账户风险水平指标',
      format: formatPercent,
    },
    {
      name: '未实现盈亏',
      value: accountInfo?.unrealizedPnL || 0,
      icon: ArrowTrendingUpIcon,
      description: '当前持仓的浮动盈亏',
      highlight: true,
    },
    {
      name: '已实现盈亏',
      value: accountInfo?.realizedPnL || 0,
      icon: ArrowTrendingDownIcon,
      description: '历史平仓的累计盈亏',
      highlight: true,
    },
  ];

  return (
    <div>
      {/* 资产统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {stats.map((item) => (
          <div
            key={item.name}
            className="bg-gray-50 rounded-lg p-6 hover:bg-gray-100 transition-colors duration-200"
          >
            <div className="flex items-center space-x-4">
              <div className="rounded-lg bg-white p-3">
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
                  {item.format ? item.format(item.value) : formatCurrency(item.value)}
                </p>
                <p className="text-sm text-gray-500 mt-1">{item.description}</p>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* 资产分布图表 */}
      <div className="mt-8 bg-gray-50 rounded-lg p-6">
        <h3 className="text-lg font-medium mb-4">资产分布</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          {/* 左侧 - 资产类型分布 */}
          <div>
            <h4 className="text-sm font-medium text-gray-500 mb-4">按资产类型</h4>
            <div className="space-y-4">
              {[
                { name: '现货', value: 0.6 },
                { name: '合约', value: 0.3 },
                { name: '资金费用', value: 0.1 },
              ].map((item) => (
                <div key={item.name}>
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-gray-500">{item.name}</span>
                    <span className="font-medium">{formatPercent(item.value)}</span>
                  </div>
                  <div className="h-2 bg-gray-200 rounded-full">
                    <div
                      className="h-full bg-blue-500 rounded-full"
                      style={{ width: `${item.value * 100}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* 右侧 - 币种分布 */}
          <div>
            <h4 className="text-sm font-medium text-gray-500 mb-4">按币种</h4>
            <div className="space-y-4">
              {[
                { name: 'USDT', value: 0.5 },
                { name: 'BTC', value: 0.3 },
                { name: 'ETH', value: 0.2 },
              ].map((item) => (
                <div key={item.name}>
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-gray-500">{item.name}</span>
                    <span className="font-medium">{formatPercent(item.value)}</span>
                  </div>
                  <div className="h-2 bg-gray-200 rounded-full">
                    <div
                      className="h-full bg-green-500 rounded-full"
                      style={{ width: `${item.value * 100}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
} 