'use client';

import { useTradingStore } from '@/store';
import { formatNumber, formatCurrency, formatPercent } from '@/utils/format';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from 'recharts';

interface Position {
  symbol: string;
  size: number;
  entryPrice: number;
  direction: 'long' | 'short';
}

interface MarketData {
  [symbol: string]: {
    price: number;
  };
}

interface AccountInfo {
  positions: Position[];
}

interface CustomTooltipProps {
  active?: boolean;
  payload?: Array<{
    payload: {
      symbol: string;
      value: number;
      percentage: number;
      direction: 'long' | 'short';
    };
  }>;
}

export default function PositionOverview() {
  const { accountInfo, marketData } = useTradingStore() as { accountInfo: AccountInfo | null; marketData: MarketData };

  const positions = accountInfo?.positions || [];
  const totalValue = positions.reduce((sum: number, pos: Position) => {
    const currentPrice = marketData[pos.symbol]?.price || pos.entryPrice;
    return sum + Math.abs(pos.size * currentPrice);
  }, 0);

  // 计算持仓分布数据
  const positionData = positions.map((position) => {
    const currentPrice = marketData[position.symbol]?.price || position.entryPrice;
    const value = Math.abs(position.size * currentPrice);
    const percentage = (value / totalValue) * 100;

    return {
      symbol: position.symbol,
      value,
      percentage,
      direction: position.direction,
    };
  });

  // 饼图颜色
  const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];

  const CustomTooltip = ({ active, payload }: CustomTooltipProps) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload;
      return (
        <div className="bg-white p-3 border border-gray-200 rounded-lg shadow-lg">
          <div className="font-medium">{data.symbol}</div>
          <div className="text-sm text-gray-500">
            {formatCurrency(data.value)} ({formatPercent(data.percentage / 100)})
          </div>
          <div className="text-sm text-gray-500">
            {data.direction === 'long' ? '多头' : '空头'}
          </div>
        </div>
      );
    }
    return null;
  };

  return (
    <div>
      {positions.length > 0 ? (
        <div className="space-y-6">
          {/* 持仓分布图表 */}
          <div className="h-[200px]">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={positionData}
                  dataKey="value"
                  nameKey="symbol"
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={80}
                >
                  {positionData.map((entry, index) => (
                    <Cell
                      key={entry.symbol}
                      fill={COLORS[index % COLORS.length]}
                    />
                  ))}
                </Pie>
                <Tooltip content={<CustomTooltip />} />
              </PieChart>
            </ResponsiveContainer>
          </div>

          {/* 持仓列表 */}
          <div className="space-y-3">
            {positionData.map((position, index) => (
              <div
                key={position.symbol}
                className="flex items-center justify-between"
              >
                <div className="flex items-center space-x-2">
                  <div
                    className="w-3 h-3 rounded-full"
                    style={{ backgroundColor: COLORS[index % COLORS.length] }}
                  />
                  <span className="text-sm font-medium">{position.symbol}</span>
                  <span className={`text-xs ${
                    position.direction === 'long'
                      ? 'text-green-500'
                      : 'text-red-500'
                  }`}>
                    {position.direction === 'long' ? '多' : '空'}
                  </span>
                </div>
                <div className="text-sm text-gray-500">
                  {formatPercent(position.percentage / 100)}
                </div>
              </div>
            ))}
          </div>
        </div>
      ) : (
        <div className="text-center text-gray-500 py-8">
          暂无持仓
        </div>
      )}
    </div>
  );
}  