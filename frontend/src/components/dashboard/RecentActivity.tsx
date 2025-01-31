'use client';

import { useTradingStore } from '@/store';
import { formatNumber, formatCurrency } from '@/utils/format';

export default function RecentActivity() {
  const { recentTrades } = useTradingStore();

  return (
    <div className="space-y-4">
      {recentTrades.length > 0 ? (
        recentTrades.slice(0, 5).map((trade) => (
          <div
            key={trade.id}
            className="flex items-center justify-between py-2 border-b border-gray-100 last:border-0"
          >
            <div>
              <div className="flex items-center space-x-2">
                <span className="font-medium">{trade.symbol}</span>
                <span className={`text-xs ${
                  trade.direction === 'buy' ? 'text-green-500' : 'text-red-500'
                }`}>
                  {trade.type === 'open'
                    ? (trade.direction === 'buy' ? '开多' : '开空')
                    : (trade.direction === 'buy' ? '平空' : '平多')}
                </span>
              </div>
              <div className="text-sm text-gray-500">
                {new Date(trade.timestamp).toLocaleString()}
              </div>
            </div>
            <div className="text-right">
              <div className="font-medium">
                {formatNumber(trade.size, 4)} {trade.symbol.split('/')[0]}
              </div>
              <div className={`text-sm ${
                trade.pnl > 0
                  ? 'text-green-500'
                  : trade.pnl < 0
                  ? 'text-red-500'
                  : 'text-gray-500'
              }`}>
                {trade.type === 'close' && (
                  <>
                    {trade.pnl > 0 ? '+' : ''}{formatCurrency(trade.pnl)}
                  </>
                )}
              </div>
            </div>
          </div>
        ))
      ) : (
        <div className="text-center text-gray-500 py-4">
          暂无活动记录
        </div>
      )}

      {recentTrades.length > 5 && (
        <div className="text-center">
          <button className="text-sm text-blue-500 hover:text-blue-600">
            查看更多
          </button>
        </div>
      )}
    </div>
  );
} 