import { useEffect, useRef } from 'react'
import { createChart, IChartApi, ISeriesApi, ColorType } from 'lightweight-charts'
import { useTheme } from '@/hooks/useTheme'

interface TradingChartProps {
  data: {
    time: string
    open: number
    high: number
    low: number
    close: number
    volume: number
  }[]
  height?: number
  onCrosshairMove?: (price: number, time: string) => void
}

export const TradingChart = ({
  data,
  height = 600,
  onCrosshairMove,
}: TradingChartProps) => {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi>()
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'>>()
  const volumeSeriesRef = useRef<ISeriesApi<'Histogram'>>()
  const { isDarkMode } = useTheme()

  useEffect(() => {
    if (!chartContainerRef.current) return

    // 创建图表
    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: {
          type: ColorType.Solid,
          color: isDarkMode ? '#141414' : '#ffffff',
        },
        textColor: isDarkMode ? '#d9d9d9' : '#191919',
      },
      grid: {
        vertLines: {
          color: isDarkMode ? '#303030' : '#f0f0f0',
        },
        horzLines: {
          color: isDarkMode ? '#303030' : '#f0f0f0',
        },
      },
      width: chartContainerRef.current.clientWidth,
      height,
    })

    // 添加K线图
    const candlestickSeries = chart.addCandlestickSeries({
      upColor: '#26a69a',
      downColor: '#ef5350',
      borderVisible: false,
      wickUpColor: '#26a69a',
      wickDownColor: '#ef5350',
    })

    // 添加成交量图
    const volumeSeries = chart.addHistogramSeries({
      color: '#26a69a',
      priceFormat: {
        type: 'volume',
      },
      priceScaleId: '', // 在右边创建一个新的价格轴
    })

    // 设置数据
    candlestickSeries.setData(data)
    volumeSeries.setData(
      data.map(item => ({
        time: item.time,
        value: item.volume,
        color: item.close > item.open ? '#26a69a' : '#ef5350',
      }))
    )

    // 添加十字线移动事件
    chart.subscribeCrosshairMove(param => {
      if (
        param.time &&
        param.point &&
        param.seriesData.get(candlestickSeries) &&
        onCrosshairMove
      ) {
        const price = (param.seriesData.get(candlestickSeries) as any).close
        onCrosshairMove(price, param.time as string)
      }
    })

    // 保存引用
    chartRef.current = chart
    candlestickSeriesRef.current = candlestickSeries
    volumeSeriesRef.current = volumeSeries

    // 处理窗口大小变化
    const handleResize = () => {
      if (chartContainerRef.current && chartRef.current) {
        chartRef.current.applyOptions({
          width: chartContainerRef.current.clientWidth,
        })
      }
    }

    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      chart.remove()
    }
  }, [data, height, isDarkMode, onCrosshairMove])

  return <div ref={chartContainerRef} className="chart-container" />
} 