'use client';

import { useTradingStore } from '@/store';
import { formatPercent, formatNumber } from '@/utils/format';

export default function RiskMetrics() {
  const { accountInfo } = useTradingStore();

  // 模拟风险指标数据
  const metrics = [
    {
      name: '账户杠杆',
      value: 2.5,
      threshold: 3,
      format: (v: number) => `${v}x`,
      status: 'normal', // 'normal' | 'warning' | 'danger'
    },
    {
      name: '保证金使用率',
      value: 0.75,
      threshold: 0.8,
      format: formatPercent,
      status: 'warning',
    },
    {
      name: '未实现亏损',
      value: -0.15,
      threshold: -0.2,
      format: formatPercent,
      status: 'normal',
    },
    {
      name: '最大持仓占比',
      value: 0.4,
      threshold: 0.5,
      format: formatPercent,
      status: 'normal',
    },
    {
      name: '相关性暴露',
      value: 0.85,
      threshold: 0.9,
      format: formatNumber,
      status: 'warning',
    },
  ];

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'danger':
        return 'text-red-500';
      case 'warning':
        return 'text-yellow-500';
      default:
        return 'text-green-500';
    }
  };

  const getProgressColor = (status: string) => {
    switch (status) {
      case 'danger':
        return 'bg-red-500';
      case 'warning':
        return 'bg-yellow-500';
      default:
        return 'bg-green-500';
    }
  };

  return (
    <div className="space-y-4">
      {metrics.map((metric) => (
        <div key={metric.name}>
          <div className="flex justify-between mb-1">
            <span className="text-sm text-gray-500">{metric.name}</span>
            <span className={`text-sm font-medium ${getStatusColor(metric.status)}`}>
              {metric.format(metric.value)}
            </span>
          </div>
          <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
            <div
              className={`h-full ${getProgressColor(metric.status)} transition-all duration-300`}
              style={{
                width: `${(Math.abs(metric.value) / metric.threshold) * 100}%`,
              }}
            />
          </div>
        </div>
      ))}

      {/* 风险提示 */}
      {metrics.some((m) => m.status !== 'normal') && (
        <div className="mt-6 p-3 bg-yellow-50 rounded-lg">
          <div className="flex items-start">
            <svg
              className="w-5 h-5 text-yellow-500 mt-0.5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-yellow-800">
                风险提示
              </h3>
              <div className="mt-1 text-sm text-yellow-700">
                <ul className="list-disc list-inside">
                  {metrics
                    .filter((m) => m.status !== 'normal')
                    .map((m) => (
                      <li key={m.name}>
                        {m.name}已接近阈值({m.format(m.threshold)})，请注意风险
                      </li>
                    ))}
                </ul>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
} 