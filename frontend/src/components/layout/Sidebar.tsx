import { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  HomeIcon,
  ChartBarIcon,
  CurrencyDollarIcon,
  ClockIcon,
  CogIcon,
} from '@heroicons/react/24/outline';

const navigation = [
  { name: '仪表盘', href: '/', icon: HomeIcon },
  { name: '交易', href: '/trading', icon: ChartBarIcon },
  { name: '资产', href: '/assets', icon: CurrencyDollarIcon },
  { name: '历史', href: '/history', icon: ClockIcon },
  { name: '设置', href: '/settings', icon: CogIcon },
];

export default function Sidebar() {
  const pathname = usePathname();
  const [collapsed, setCollapsed] = useState(false);

  return (
    <div className={`bg-gray-900 text-white transition-all duration-300 ${
      collapsed ? 'w-16' : 'w-64'
    }`}>
      <div className="flex h-16 items-center justify-between px-4">
        {!collapsed && <span className="text-xl font-bold">Trading System</span>}
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="p-2 hover:bg-gray-800 rounded-lg"
        >
          <svg
            className="w-6 h-6"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d={collapsed ? 'M13 5l7 7-7 7M5 5l7 7-7 7' : 'M11 19l-7-7 7-7m8 14l-7-7 7-7'}
            />
          </svg>
        </button>
      </div>

      <nav className="mt-5 px-2">
        {navigation.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.name}
              href={item.href}
              className={`flex items-center px-4 py-2 mt-2 text-sm rounded-lg transition-colors duration-200 ${
                isActive
                  ? 'bg-gray-800 text-white'
                  : 'text-gray-400 hover:bg-gray-800 hover:text-white'
              }`}
            >
              <item.icon className="w-6 h-6" />
              {!collapsed && <span className="ml-3">{item.name}</span>}
            </Link>
          );
        })}
      </nav>

      <div className="absolute bottom-0 w-full p-4">
        <div className="flex items-center px-4 py-2 text-sm text-gray-400">
          <CogIcon className="w-6 h-6" />
          {!collapsed && <span className="ml-3">v1.0.0</span>}
        </div>
      </div>
    </div>
  );
} 