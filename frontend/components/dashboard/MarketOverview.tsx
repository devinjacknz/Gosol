'use client';

import { useTradingStore } from '@/store';
import { formatNumber, formatPercent, getPriceChangeColor } from '@/utils/format';
import Link from 'next/link';

export default function MarketOverview() {
  const { marketData } = useTradingStore();

  // 将市场数据转换为数组并按24h涨跅排序
  const markets = Object.values(marketData).sort((a, b) => b.change24h - a.change24h);

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full">
        <thead>
          <tr className="text-left text-sm text-gray-500">
            <th className="pb-4">交易对</th>
            <th className="pb-4">最新价格</th>
            <th className="pb-4">24h涨跌</th>
            <th className="pb-4">24h最高</th>
            <th className="pb-4">24h最低</th>
            <th className="pb-4">24h成交量</th>
            <th className="pb-4">操作</th>
          </tr>
        </thead>
        <tbody className="text-sm">
          {markets.length > 0 ? (
            markets.map((market) => (
              <tr key={market.symbol} className="border-t border-gray-100">
                <td className="py-4">
                  <div className="font-medium">{market.symbol}</div>
                </td>
                <td className="py-4">
                  <div className="font-medium">{formatNumber(market.price)}</div>
                </td>
                <td className="py-4">
                  <div className={getPriceChangeColor(market.change24h)}>
                    {formatPercent(market.change24h)}
                  </div>
                </td>
                <td className="py-4">{formatNumber(market.high24h)}</td>
                <td className="py-4">{formatNumber(market.low24h)}</td>
                <td className="py-4">{formatNumber(market.volume)}</td>
                <td className="py-4">
                  <Link
                    href={`/trading?symbol=${market.symbol}`}
                    className="text-blue-500 hover:text-blue-600"
                  >
                    交易
                  </Link>
                </td>
              </tr>
            ))
          ) : (
            <tr>
              <td colSpan={7} className="py-8 text-center text-gray-500">
                暂无市场数据
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
} 