'use client';

import { useState } from 'react';
import { useTradingStore } from '@/store';
import { formatNumber, formatCurrency, formatDateTime } from '@/utils/format';

export default function FundingHistory() {
  const [timeRange, setTimeRange] = useState<'1d' | '1w' | '1m' | '3m' | 'all'>('1w');
  const [type, setType] = useState<'all' | 'deposit' | 'withdrawal'>('all');

  // 模拟资金记录数据
  const fundingHistory = [
    {
      id: '1',
      type: 'deposit',
      asset: 'BTC',
      amount: 0.5,
      status: 'completed',
      timestamp: new Date('2024-03-10T10:30:00'),
      txHash: '0x1234...5678',
      network: 'Bitcoin',
      fee: 0.0001,
    },
    {
      id: '2',
      type: 'withdrawal',
      asset: 'ETH',
      amount: 2.0,
      status: 'completed',
      timestamp: new Date('2024-03-09T15:45:00'),
      txHash: '0x9876...4321',
      network: 'Ethereum',
      fee: 0.002,
    },
    {
      id: '3',
      type: 'deposit',
      asset: 'USDT',
      amount: 1000,
      status: 'pending',
      timestamp: new Date('2024-03-08T09:15:00'),
      txHash: '0xabcd...efgh',
      network: 'Tron',
      fee: 1,
    },
  ];

  // 过滤记录
  const filteredHistory = fundingHistory.filter((record) => {
    const matchesType = type === 'all' || record.type === type;
    
    const recordDate = record.timestamp.getTime();
    const now = new Date().getTime();
    const timeRangeFilter = {
      '1d': now - 24 * 60 * 60 * 1000,
      '1w': now - 7 * 24 * 60 * 60 * 1000,
      '1m': now - 30 * 24 * 60 * 60 * 1000,
      '3m': now - 90 * 24 * 60 * 60 * 1000,
      'all': 0,
    };

    return matchesType && recordDate >= timeRangeFilter[timeRange];
  });

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-green-100 text-green-800';
      case 'pending':
        return 'bg-yellow-100 text-yellow-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div>
      {/* 筛选控件 */}
      <div className="flex flex-wrap gap-4 mb-6">
        <div className="flex space-x-2">
          {[
            { label: '全部', value: 'all' },
            { label: '充值', value: 'deposit' },
            { label: '提现', value: 'withdrawal' },
          ].map((t) => (
            <button
              key={t.value}
              onClick={() => setType(t.value as 'all' | 'deposit' | 'withdrawal')}
              className={`px-4 py-2 rounded-lg ${
                type === t.value
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>
        <div className="flex space-x-2">
          {[
            { label: '1天', value: '1d' },
            { label: '1周', value: '1w' },
            { label: '1月', value: '1m' },
            { label: '3月', value: '3m' },
            { label: '全部', value: 'all' },
          ].map((t) => (
            <button
              key={t.value}
              onClick={() => setTimeRange(t.value as '1d' | '1w' | '1m' | '3m' | 'all')}
              className={`px-4 py-2 rounded-lg ${
                timeRange === t.value
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>
      </div>

      {/* 资金记录列表 */}
      <div className="overflow-x-auto">
        <table className="min-w-full">
          <thead>
            <tr className="text-left text-sm text-gray-500">
              <th className="pb-4">时间</th>
              <th className="pb-4">类型</th>
              <th className="pb-4">币种</th>
              <th className="pb-4">数量</th>
              <th className="pb-4">网络</th>
              <th className="pb-4">手续费</th>
              <th className="pb-4">状态</th>
              <th className="pb-4">交易哈希</th>
            </tr>
          </thead>
          <tbody className="text-sm">
            {filteredHistory.map((record) => (
              <tr key={record.id} className="border-t border-gray-100">
                <td className="py-4">{formatDateTime(record.timestamp)}</td>
                <td className="py-4">
                  <span className={`${
                    record.type === 'deposit' ? 'text-green-500' : 'text-red-500'
                  }`}>
                    {record.type === 'deposit' ? '充值' : '提现'}
                  </span>
                </td>
                <td className="py-4">{record.asset}</td>
                <td className="py-4">{formatNumber(record.amount, 4)}</td>
                <td className="py-4">{record.network}</td>
                <td className="py-4">{formatNumber(record.fee, 4)} {record.asset}</td>
                <td className="py-4">
                  <span className={`px-2 py-1 rounded-full text-xs ${getStatusColor(record.status)}`}>
                    {record.status === 'completed' ? '已完成' :
                     record.status === 'pending' ? '处理中' : '失败'}
                  </span>
                </td>
                <td className="py-4">
                  <a
                    href={`https://etherscan.io/tx/${record.txHash}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-500 hover:text-blue-600"
                  >
                    {record.txHash.slice(0, 6)}...{record.txHash.slice(-4)}
                  </a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {filteredHistory.length === 0 && (
        <div className="text-center py-8 text-gray-500">
          暂无资金记录
        </div>
      )}
    </div>
  );
}     