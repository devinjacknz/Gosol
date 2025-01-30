'use client';

import { useEffect, useState } from 'react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { formatCurrency } from '@/utils/format';

// 模拟数据
const generateData = () => {
  const data = [];
  const startDate = new Date('2024-01-01');
  let value = 100000;

  for (let i = 0; i < 30; i++) {
    const date = new Date(startDate);
    date.setDate(startDate.getDate() + i);
    
    // 模拟每日波动
    const change = (Math.random() - 0.45) * 0.02;
    value = value * (1 + change);

    data.push({
      date: date.toISOString().split('T')[0],
      value: value,
      unrealizedPnL: value * (Math.random() - 0.5) * 0.1,
      realizedPnL: value * (Math.random() - 0.5) * 0.05,
    });
  }

  return data;
};

const CustomTooltip = ({ active, payload, label }: any) => {
  if (active && payload && payload.length) {
    return (
      <div className="bg-white p-4 border border-gray-200 rounded-lg shadow-lg">
        <p className="text-sm text-gray-500">{label}</p>
        <p className="font-medium">
          总权益: {formatCurrency(payload[0].value)}
        </p>
        <p className={`text-sm ${payload[1].value > 0 ? 'text-green-500' : 'text-red-500'}`}>
          未实现盈亏: {formatCurrency(payload[1].value)}
        </p>
        <p className={`text-sm ${payload[2].value > 0 ? 'text-green-500' : 'text-red-500'}`}>
          已实现盈亏: {formatCurrency(payload[2].value)}
        </p>
      </div>
    );
  }
  return null;
};

export default function PortfolioChart() {
  const [data, setData] = useState(generateData());
  const [timeRange, setTimeRange] = useState<'1d' | '7d' | '30d' | '90d'>('30d');

  return (
    <div>
      {/* 时间范围选择 */}
      <div className="flex space-x-2 mb-4">
        {[
          { label: '1天', value: '1d' },
          { label: '7天', value: '7d' },
          { label: '30天', value: '30d' },
          { label: '90天', value: '90d' },
        ].map((range) => (
          <button
            key={range.value}
            onClick={() => setTimeRange(range.value as any)}
            className={`px-3 py-1 rounded-lg text-sm ${
              timeRange === range.value
                ? 'bg-blue-500 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            }`}
          >
            {range.label}
          </button>
        ))}
      </div>

      {/* 图表 */}
      <div className="h-[300px]">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart
            data={data}
            margin={{
              top: 10,
              right: 30,
              left: 0,
              bottom: 0,
            }}
          >
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis
              dataKey="date"
              tick={{ fontSize: 12 }}
              tickFormatter={(value) => value.slice(5)}
            />
            <YAxis
              tick={{ fontSize: 12 }}
              tickFormatter={(value) => `$${(value / 1000).toFixed(1)}k`}
            />
            <Tooltip content={<CustomTooltip />} />
            <Area
              type="monotone"
              dataKey="value"
              stroke="#3b82f6"
              fill="#3b82f6"
              fillOpacity={0.1}
            />
            <Area
              type="monotone"
              dataKey="unrealizedPnL"
              stroke="#10b981"
              fill="#10b981"
              fillOpacity={0.1}
            />
            <Area
              type="monotone"
              dataKey="realizedPnL"
              stroke="#f59e0b"
              fill="#f59e0b"
              fillOpacity={0.1}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      {/* 图例 */}
      <div className="flex justify-center space-x-6 mt-4">
        <div className="flex items-center">
          <div className="w-3 h-3 rounded-full bg-blue-500 mr-2" />
          <span className="text-sm text-gray-600">总权益</span>
        </div>
        <div className="flex items-center">
          <div className="w-3 h-3 rounded-full bg-green-500 mr-2" />
          <span className="text-sm text-gray-600">未实现盈亏</span>
        </div>
        <div className="flex items-center">
          <div className="w-3 h-3 rounded-full bg-yellow-500 mr-2" />
          <span className="text-sm text-gray-600">已实现盈亏</span>
        </div>
      </div>
    </div>
  );
} 