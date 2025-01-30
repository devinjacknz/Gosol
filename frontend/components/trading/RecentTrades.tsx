'use client';

import { useTradingStore } from '@/store';
import { formatNumber, formatDate } from '@/utils/format';

interface RecentTradesProps {
  symbol: string;
}

export default function RecentTrades({ symbol }: RecentTradesProps) {
  const { recentTrades } = useTradingStore();

  // 过滤出当前交易对的成交记录
  const trades = recentTrades.filter(trade => trade.symbol === symbol);

  return (
    <div className="bg-white rounded-lg p-4">
      <h3 className="font-medium mb-4">最近成交</h3>

      <div className="text-sm">
        {/* 列标题 */}
        <div className="grid grid-cols-4 text-gray-500 mb-2">
          <div>时间</div>
          <div>方向</div>
          <div className="text-right">价格</div>
          <div className="text-right">数量</div>
        </div>

        {/* 成交记录 */}
        <div className="space-y-1">
          {trades.length > 0 ? (
            trades.map((trade) => (
              <div key={trade.id} className="grid grid-cols-4">
                <div className="text-gray-500">
                  {new Date(trade.timestamp).toLocaleTimeString()}
                </div>
                <div className={trade.direction === 'buy' ? 'text-green-500' : 'text-red-500'}>
                  {trade.direction === 'buy' ? '买入' : '卖出'}
                </div>
                <div className="text-right">{formatNumber(trade.price)}</div>
                <div className="text-right">{formatNumber(trade.size, 4)}</div>
              </div>
            ))
          ) : (
            <div className="text-center text-gray-500 py-4">
              暂无成交记录
            </div>
          )}
        </div>
      </div>
    </div>
  );
} 