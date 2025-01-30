'use client';

import { useState } from 'react';
import { Tab } from '@headlessui/react';
import AssetOverview from '@/components/assets/AssetOverview';
import AssetDetails from '@/components/assets/AssetDetails';
import TransactionHistory from '@/components/assets/TransactionHistory';
import FundingHistory from '@/components/assets/FundingHistory';

export default function AssetsPage() {
  const tabs = [
    { name: '资产总览', component: AssetOverview },
    { name: '资产明细', component: AssetDetails },
    { name: '资金流水', component: TransactionHistory },
    { name: '资金费用', component: FundingHistory },
  ];

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg">
        <Tab.Group>
          <Tab.List className="flex border-b border-gray-200">
            {tabs.map((tab) => (
              <Tab
                key={tab.name}
                className={({ selected }) =>
                  `flex-1 py-4 px-6 text-sm font-medium focus:outline-none ${
                    selected
                      ? 'text-blue-600 border-b-2 border-blue-600'
                      : 'text-gray-500 hover:text-gray-700'
                  }`
                }
              >
                {tab.name}
              </Tab>
            ))}
          </Tab.List>
          <Tab.Panels>
            {tabs.map((tab, idx) => (
              <Tab.Panel key={idx} className="p-6">
                <tab.component />
              </Tab.Panel>
            ))}
          </Tab.Panels>
        </Tab.Group>
      </div>
    </div>
  );
} 