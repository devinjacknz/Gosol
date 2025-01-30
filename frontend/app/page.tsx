'use client';

import AccountSummary from '@/components/dashboard/AccountSummary';
import PortfolioChart from '@/components/dashboard/PortfolioChart';
import MarketOverview from '@/components/dashboard/MarketOverview';
import PositionOverview from '@/components/dashboard/PositionOverview';
import RiskMetrics from '@/components/dashboard/RiskMetrics';
import RecentActivity from '@/components/dashboard/RecentActivity';

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      {/* 账户概览 */}
      <AccountSummary />

      <div className="grid grid-cols-12 gap-6">
        {/* 左侧 - 图表和市场概览 */}
        <div className="col-span-8 space-y-6">
          {/* 资产曲线 */}
          <div className="bg-white rounded-lg p-6">
            <h2 className="text-lg font-medium mb-4">资产曲线</h2>
            <div className="h-[400px]">
              <PortfolioChart />
            </div>
          </div>

          {/* 市场概览 */}
          <div className="bg-white rounded-lg p-6">
            <h2 className="text-lg font-medium mb-4">市场概览</h2>
            <MarketOverview />
          </div>
        </div>

        {/* 右侧 - 持仓和风险概览 */}
        <div className="col-span-4 space-y-6">
          {/* 持仓概览 */}
          <div className="bg-white rounded-lg p-6">
            <h2 className="text-lg font-medium mb-4">持仓概览</h2>
            <PositionOverview />
          </div>

          {/* 风险指标 */}
          <div className="bg-white rounded-lg p-6">
            <h2 className="text-lg font-medium mb-4">风险指标</h2>
            <RiskMetrics />
          </div>

          {/* 最近活动 */}
          <div className="bg-white rounded-lg p-6">
            <h2 className="text-lg font-medium mb-4">最近活动</h2>
            <RecentActivity />
          </div>
        </div>
      </div>
    </div>
  );
} 