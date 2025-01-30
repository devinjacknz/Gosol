'use client';

import { useState } from 'react';
import { useTradingStore } from '@/store';
import { formatNumber, formatCurrency, formatPercent } from '@/utils/format';

export default function AssetDetails() {
  const { accountInfo, marketData } = useTradingStore();
  const [searchTerm, setSearchTerm] = useState('');
  const [assetType, setAssetType] = useState<'all' | 'spot' | 'contract'>('all');

  // 模拟资产数据
  const assets = [
    {
      symbol: 'BTC',
      type: 'spot',
      total: 1.5,
      available: 1.2,
      frozen: 0.3,
      value: 45000,
      price: 30000,
      change24h: 0.05,
    },
    {
      symbol: 'ETH',
      type: 'spot',
      total: 10,
      available: 8,
      frozen: 2,
      value: 20000,
      price: 2000,
      change24h: -0.02,
    },
    {
      symbol: 'BTC/USDT',
      type: 'contract',
      total: 0.5,
      available: 0.5,
      frozen: 0,
      value: 15000,
      price: 30000,
      change24h: 0.05,
      leverage: 3,
      marginType: 'isolated',
    },
  ];

  // 过滤资产
  const filteredAssets = assets.filter((asset) => {
    const matchesSearch = asset.symbol.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesType = assetType === 'all' || asset.type === assetType;
    return matchesSearch && matchesType;
  });

  return (
    <div>
      {/* 搜索和筛选 */}
      <div className="flex flex-wrap gap-4 mb-6">
        <div className="flex-1">
          <input
            type="text"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="搜索币种..."
            className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>
        <div className="flex space-x-2">
          {[
            { label: '全部', value: 'all' },
            { label: '现货', value: 'spot' },
            { label: '合约', value: 'contract' },
          ].map((type) => (
            <button
              key={type.value}
              onClick={() => setAssetType(type.value as any)}
              className={`px-4 py-2 rounded-lg ${
                assetType === type.value
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              {type.label}
            </button>
          ))}
        </div>
      </div>

      {/* 资产列表 */}
      <div className="overflow-x-auto">
        <table className="min-w-full">
          <thead>
            <tr className="text-left text-sm text-gray-500">
              <th className="pb-4">币种</th>
              <th className="pb-4">类型</th>
              <th className="pb-4">总数量</th>
              <th className="pb-4">可用</th>
              <th className="pb-4">冻结</th>
              <th className="pb-4">最新价格</th>
              <th className="pb-4">24h涨跌</th>
              <th className="pb-4">估值</th>
              {assetType === 'contract' && (
                <>
                  <th className="pb-4">杠杆</th>
                  <th className="pb-4">保证金模式</th>
                </>
              )}
            </tr>
          </thead>
          <tbody className="text-sm">
            {filteredAssets.map((asset) => (
              <tr key={asset.symbol} className="border-t border-gray-100">
                <td className="py-4">
                  <div className="font-medium">{asset.symbol}</div>
                </td>
                <td className="py-4">
                  <span className="px-2 py-1 rounded-full text-xs bg-gray-100">
                    {asset.type === 'spot' ? '现货' : '合约'}
                  </span>
                </td>
                <td className="py-4">{formatNumber(asset.total, 4)}</td>
                <td className="py-4">{formatNumber(asset.available, 4)}</td>
                <td className="py-4">{formatNumber(asset.frozen, 4)}</td>
                <td className="py-4">{formatCurrency(asset.price)}</td>
                <td className="py-4">
                  <span className={asset.change24h > 0 ? 'text-green-500' : 'text-red-500'}>
                    {formatPercent(asset.change24h)}
                  </span>
                </td>
                <td className="py-4">{formatCurrency(asset.value)}</td>
                {asset.type === 'contract' && (
                  <>
                    <td className="py-4">{asset.leverage}x</td>
                    <td className="py-4">
                      {asset.marginType === 'isolated' ? '逐仓' : '全仓'}
                    </td>
                  </>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
} 