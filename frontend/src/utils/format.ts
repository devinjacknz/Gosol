import { format } from 'date-fns';

export const formatDateTime = (date: string | number | Date): string => {
  return format(new Date(date), 'yyyy-MM-dd HH:mm:ss');
};

// 格式化数字
export const formatNumber = (num: number, decimals = 2): string => {
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(num);
};

// 格式化货币
export const formatCurrency = (num: number, currency = 'USD'): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
  }).format(num);
};

// 格式化百分比
export const formatPercent = (num: number, decimals = 2): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'percent',
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(num);
};

// 格式化时间
export const formatDate = (date: string | number | Date): string => {
  return format(new Date(date), 'yyyy-MM-dd HH:mm:ss');
};

// 格式化持仓方向
export const formatDirection = (direction: 'long' | 'short'): string => {
  return direction === 'long' ? '多' : '空';
};

// 格式化保证金类型
export const formatMarginType = (type: 'isolated' | 'cross'): string => {
  return type === 'isolated' ? '逐仓' : '全仓';
};

// 格式化文件大小
export const formatFileSize = (bytes: number): string => {
  const units = ['B', 'KB', 'MB', 'GB'];
  let size = bytes;
  let unitIndex = 0;
  
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }
  
  return `${formatNumber(size, 2)} ${units[unitIndex]}`;
};

// 格式化延迟时间
export const formatDelay = (ms: number): string => {
  return ms < 1000 ? `${ms}ms` : `${formatNumber(ms / 1000, 1)}s`;
};

// 格式化风险等级
export const formatRiskLevel = (severity: 'low' | 'medium' | 'high'): string => {
  const levels = {
    low: '低',
    medium: '中',
    high: '高',
  };
  return levels[severity];
};

// 获取价格变化的颜色类名
export const getPriceChangeColor = (change: number): string => {
  if (change > 0) return 'text-green-500';
  if (change < 0) return 'text-red-500';
  return 'text-gray-500';
};

// 获取风险等级的颜色类名
export const getRiskLevelColor = (severity: 'low' | 'medium' | 'high' | 'critical'): string => {
  const colors = {
    low: 'text-green-500',
    medium: 'text-yellow-500',
    high: 'text-red-500',
    critical: 'text-red-600',
  };
  return colors[severity];
};        