'use client';

import { useTradingStore } from '@/store';
import { formatNumber, formatCurrency, formatPercent, formatMarginType } from '@/utils/format';

export default function Positions() {
  const { accountInfo, marketData } = useTradingStore();

  const positions = accountInfo?.positions || [];

  return (
    <div className="bg-white rounded-lg p-4">
      <h3 className="font-medium mb-4">当前持仓</h3>

      {positions.length > 0 ? (
        <div className="space-y-4">
          {positions.map((position) => {
            const currentPrice = marketData[position.symbol]?.price || position.entryPrice;
            const pnlPercent = ((currentPrice - position.entryPrice) / position.entryPrice) * 
              (position.direction === 'long' ? 1 : -1);

            return (
              <div
                key={`${position.symbol}-${position.direction}`}
                className="border border-gray-200 rounded-lg p-4"
              >
                {/* 持仓标题 */}
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center space-x-2">
                    <span className="font-medium">{position.symbol}</span>
                    <span className={`text-sm ${
                      position.direction === 'long' ? 'text-green-500' : 'text-red-500'
                    }`}>
                      {position.direction === 'long' ? '多' : '空'} {position.leverage}x
                    </span>
                    <span className="text-sm text-gray-500">
                      {formatMarginType(position.marginType)}
                    </span>
                  </div>
                  <div className="flex space-x-2">
                    <button className="px-3 py-1 text-sm rounded-lg bg-gray-100 hover:bg-gray-200">
                      调整
                    </button>
                    <button className="px-3 py-1 text-sm rounded-lg bg-red-500 text-white hover:bg-red-600">
                      平仓
                    </button>
                  </div>
                </div>

                {/* 持仓信息 */}
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <div className="text-gray-500 mb-1">持仓数量</div>
                    <div className="font-medium">
                      {formatNumber(position.size, 4)} {position.symbol.split('/')[0]}
                    </div>
                  </div>
                  <div>
                    <div className="text-gray-500 mb-1">开仓价格</div>
                    <div className="font-medium">
                      {formatNumber(position.entryPrice)}
                    </div>
                  </div>
                  <div>
                    <div className="text-gray-500 mb-1">标记价格</div>
                    <div className="font-medium">
                      {formatNumber(currentPrice)}
                    </div>
                  </div>
                  <div>
                    <div className="text-gray-500 mb-1">强平价格</div>
                    <div className="font-medium">
                      {formatNumber(position.liquidationPrice)}
                    </div>
                  </div>
                  <div>
                    <div className="text-gray-500 mb-1">保证金率</div>
                    <div className={`font-medium ${
                      position.marginRatio < 0.1 ? 'text-red-500' : ''
                    }`}>
                      {formatPercent(position.marginRatio)}
                    </div>
                  </div>
                  <div>
                    <div className="text-gray-500 mb-1">未实现盈亏</div>
                    <div className={`font-medium ${
                      position.unrealizedPnL > 0 
                        ? 'text-green-500' 
                        : position.unrealizedPnL < 0 
                        ? 'text-red-500' 
                        : ''
                    }`}>
                      {formatCurrency(position.unrealizedPnL)}
                      <span className="ml-1 text-xs">
                        ({formatPercent(pnlPercent)})
                      </span>
                    </div>
                  </div>
                </div>

                {/* 风险警告 */}
                {position.marginRatio < 0.1 && (
                  <div className="mt-4 p-2 bg-red-50 text-red-600 text-sm rounded-lg">
                    ⚠️ 保证金率过低，请注意风险
                  </div>
                )}
              </div>
            );
          })}
        </div>
      ) : (
        <div className="text-center text-gray-500 py-8">
          暂无持仓
        </div>
      )}
    </div>
  );
} 