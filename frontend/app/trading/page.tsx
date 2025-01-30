'use client';

import { useState } from 'react';
import TradingChart from '@/components/trading/TradingChart';
import OrderForm from '@/components/trading/OrderForm';
import MarketInfo from '@/components/trading/MarketInfo';
import OrderBook from '@/components/trading/OrderBook';
import RecentTrades from '@/components/trading/RecentTrades';
import Positions from '@/components/trading/Positions';
import { useTradingStore } from '@/store';

export default function TradingPage() {
  const { selectedSymbol } = useTradingStore();
  const [orderType, setOrderType] = useState<'limit' | 'market'>('limit');

  return (
    <div className="h-full grid grid-cols-12 gap-4">
      {/* 左侧 - 市场信息和订单簿 */}
      <div className="col-span-3 space-y-4">
        <MarketInfo symbol={selectedSymbol} />
        <OrderBook symbol={selectedSymbol} />
        <RecentTrades symbol={selectedSymbol} />
      </div>

      {/* 中间 - 图表和交易表单 */}
      <div className="col-span-6 space-y-4">
        <div className="bg-white rounded-lg p-4 h-[600px]">
          <TradingChart symbol={selectedSymbol} />
        </div>
        <div className="bg-white rounded-lg p-4">
          <OrderForm
            symbol={selectedSymbol}
            type={orderType}
            onTypeChange={setOrderType}
          />
        </div>
      </div>

      {/* 右侧 - 持仓信息 */}
      <div className="col-span-3">
        <Positions />
      </div>
    </div>
  );
} 