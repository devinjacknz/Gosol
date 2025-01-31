'use client';

import { BellIcon } from '@heroicons/react/24/outline';
import { useTradingStore } from '../../store';
import { formatCurrency, formatDelay, getRiskLevelColor } from '../../utils/format';

export default function Header() {
  const { accountInfo, systemStatus, riskAlerts } = useTradingStore();

  return (
    <header className="bg-white border-b border-gray-200">
      <div className="flex h-16 items-center justify-between px-4">
        {/* 账户信息 */}
        <div className="flex items-center space-x-8">
          <div>
            <div className="text-sm text-gray-500">总权益</div>
            <div className="font-semibold">
              {accountInfo ? formatCurrency(accountInfo.totalEquity) : '-'}
            </div>
          </div>
          <div>
            <div className="text-sm text-gray-500">可用保证金</div>
            <div className="font-semibold">
              {accountInfo ? formatCurrency(accountInfo.availableBalance) : '-'}
            </div>
          </div>
          <div>
            <div className="text-sm text-gray-500">已用保证金</div>
            <div className="font-semibold">
              {accountInfo ? formatCurrency(accountInfo.usedMargin) : '-'}
            </div>
          </div>
          <div>
            <div className="text-sm text-gray-500">日盈亏</div>
            <div className={`font-semibold ${
              accountInfo?.dailyPnL && accountInfo.dailyPnL > 0 
                ? 'text-green-500' 
                : accountInfo?.dailyPnL && accountInfo.dailyPnL < 0 
                ? 'text-red-500' 
                : ''
            }`}>
              {accountInfo ? formatCurrency(accountInfo.dailyPnL) : '-'}
            </div>
          </div>
        </div>

        {/* 系统状态和通知 */}
        <div className="flex items-center space-x-4">
          {/* 系统状态 */}
          {systemStatus && (
            <div className="flex items-center text-sm text-gray-500">
              <span className="mr-2">延迟:</span>
              <span className={`font-medium ${
                systemStatus.dataDelay > 1000 ? 'text-red-500' : 'text-green-500'
              }`}>
                {formatDelay(systemStatus.dataDelay)}
              </span>
            </div>
          )}

          {/* 风险警告 */}
          <div className="relative">
            <button className="p-2 rounded-full hover:bg-gray-100 relative">
              <BellIcon className="w-6 h-6 text-gray-500" />
              {riskAlerts.length > 0 && (
                <span className="absolute top-0 right-0 w-4 h-4 bg-red-500 rounded-full text-white text-xs flex items-center justify-center">
                  {riskAlerts.length}
                </span>
              )}
            </button>

            {/* 风险警告下拉菜单 */}
            {riskAlerts.length > 0 && (
              <div className="absolute right-0 mt-2 w-80 bg-white rounded-lg shadow-lg border border-gray-200 z-50">
                <div className="p-2">
                  <h3 className="text-sm font-medium text-gray-900 mb-2">风险警告</h3>
                  <div className="space-y-2">
                    {riskAlerts.map((alert) => (
                      <div
                        key={alert.timestamp}
                        className="text-sm p-2 rounded-lg bg-gray-50"
                      >
                        <div className="flex items-center justify-between">
                          <span className={getRiskLevelColor(alert.severity)}>
                            {alert.type}
                          </span>
                          <span className="text-gray-400 text-xs">
                            {new Date(alert.timestamp).toLocaleTimeString()}
                          </span>
                        </div>
                        <p className="text-gray-600 mt-1">{alert.message}</p>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}  