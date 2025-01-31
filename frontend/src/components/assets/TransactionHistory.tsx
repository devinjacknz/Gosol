'use client';

import React, { useState } from 'react';

interface Transaction {
  id: string;
  type: string;
  amount: number;
  timestamp: string;
  status: string;
  asset: string;
}

export default function TransactionHistory() {
  const [transactions] = useState<Transaction[]>([]);

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">时间</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">类型</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">资产</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">数量</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {transactions.map((transaction) => (
            <tr key={transaction.id}>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{transaction.timestamp}</td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{transaction.type}</td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{transaction.asset}</td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{transaction.amount}</td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{transaction.status}</td>
            </tr>
          ))}
          {transactions.length === 0 && (
            <tr>
              <td colSpan={5} className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-center">
                暂无交易记录
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}
