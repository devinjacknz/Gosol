'use client';

import { useEffect, useRef } from 'react';
import { createChart, IChartApi, ColorType } from 'lightweight-charts';
import { useTradingStore } from '@/store';

interface TradingChartProps {
  symbol: string;
}

export default function TradingChart({ symbol }: TradingChartProps) {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const { marketData } = useTradingStore();

  useEffect(() => {
    if (!chartContainerRef.current) return;

    // 创建图表
    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { color: '#ffffff' },
        textColor: '#333',
      },
      grid: {
        vertLines: { color: '#f0f0f0' },
        horzLines: { color: '#f0f0f0' },
      },
      crosshair: {
        mode: 1,
        vertLine: {
          width: 1,
          color: '#2962FF',
          style: 0,
        },
        horzLine: {
          width: 1,
          color: '#2962FF',
          style: 0,
        },
      },
      timeScale: {
        borderColor: '#f0f0f0',
      },
      rightPriceScale: {
        borderColor: '#f0f0f0',
      },
      handleScroll: {
        mouseWheel: true,
        pressedMouseMove: true,
      },
      handleScale: {
        axisPressedMouseMove: true,
        mouseWheel: true,
        pinch: true,
      },
    });

    // 创建K线图系列
    const candlestickSeries = chart.addCandlestickSeries({
      upColor: '#26a69a',
      downColor: '#ef5350',
      borderVisible: false,
      wickUpColor: '#26a69a',
      wickDownColor: '#ef5350',
    });

    // 创建成交量系列
    const volumeSeries = chart.addHistogramSeries({
      color: '#26a69a',
      priceFormat: {
        type: 'volume',
      },
      priceScaleId: '',
      scaleMargins: {
        top: 0.8,
        bottom: 0,
      },
    });

    // 设置初始数据
    // 这里应该从后端API获取历史K线数据
    // 暂时使用模拟数据
    const initialData = [
      { time: '2024-01-01', open: 50000, high: 51000, low: 49000, close: 50500 },
      { time: '2024-01-02', open: 50500, high: 52000, low: 50000, close: 51500 },
      // ... 更多数据
    ];

    candlestickSeries.setData(initialData);

    // 自适应容器大小
    const handleResize = () => {
      if (chartContainerRef.current) {
        chart.applyOptions({
          width: chartContainerRef.current.clientWidth,
          height: chartContainerRef.current.clientHeight,
        });
      }
    };

    window.addEventListener('resize', handleResize);
    handleResize();

    // 保存图表引用
    chartRef.current = chart;

    return () => {
      window.removeEventListener('resize', handleResize);
      chart.remove();
    };
  }, []);

  // 更新实时数据
  useEffect(() => {
    if (!chartRef.current || !marketData[symbol]) return;

    const data = marketData[symbol];
    // 更新最新价格
    // 这里需要根据实际数据格式进行调整
  }, [marketData, symbol]);

  return (
    <div ref={chartContainerRef} className="w-full h-full" />
  );
} 