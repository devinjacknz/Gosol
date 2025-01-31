'use client';

import { useTradingStore } from '@/store';
import { formatNumber, formatPercent, getPriceChangeColor } from '@/utils/format';

interface MarketInfoProps {
  symbol: string;
}

export default function MarketInfo({ symbol }: MarketInfoProps) {
  const { marketData } = useTradingStore();
  const data = marketData[symbol];

  if (!data) {
    return (
      <div className="bg-white rounded-lg p-4">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="h-8 bg-gray-200 rounded w-1/2 mb-4"></div>
          <div className="space-y-3">
            <div className="h-4 bg-gray-200 rounded"></div>
            <div className="h-4 bg-gray-200 rounded"></div>
            <div className="h-4 bg-gray-200 rounded"></div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg p-4">
      {/* 交易对信息 */}
      <div className="mb-4">
        <h2 className="text-lg font-semibold">{symbol}</h2>
        <div className="flex items-baseline space-x-2">
          <span className="text-2xl font-bold">
            {formatNumber(data.price)}
          </span>
          <span className={`text-sm ${getPriceChangeColor(data.change24h)}`}>
            {formatPercent(data.change24h)}
          </span>
        </div>
      </div>

      {/* 24小时统计 */}
      <div className="space-y-2 text-sm">
        <div className="flex justify-between">
          <span className="text-gray-500">24h最高</span>
          <span className="font-medium">{formatNumber(data.high24h)}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-500">24h最低</span>
          <span className="font-medium">{formatNumber(data.low24h)}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-500">24h成交量</span>
          <span className="font-medium">{formatNumber(data.volume)}</span>
        </div>
      </div>

      {/* 资金费率信息 */}
      {data.fundingRate && (
        <div className="mt-4 pt-4 border-t border-gray-100">
          <div className="flex justify-between text-sm">
            <span className="text-gray-500">当前资金费率</span>
            <span className={getPriceChangeColor(data.fundingRate)}>
              {formatPercent(data.fundingRate)}
            </span>
          </div>
          {data.nextFundingTime && (
            <div className="flex justify-between text-sm mt-2">
              <span className="text-gray-500">下次收取时间</span>
              <span className="font-medium">
                {new Date(data.nextFundingTime).toLocaleTimeString()}
              </span>
            </div>
          )}
        </div>
      )}
    </div>
  );
} 