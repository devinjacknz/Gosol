'use client';

import { useEffect, useRef, useState } from 'react';
import { createChart, IChartApi, ISeriesApi, ColorType } from 'lightweight-charts';
import {
  Box,
  ToggleButtonGroup,
  ToggleButton,
  FormControl,
  Select,
  MenuItem,
  Typography,
} from '@mui/material';
import { useTrading } from '@/contexts/TradingContext';
import { useTheme } from '@mui/material/styles';

interface ChartData {
  time: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume?: number;
}

export default function TradingChart() {
  const theme = useTheme();
  const { state } = useTrading();
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const [chart, setChart] = useState<IChartApi | null>(null);
  const [candlestickSeries, setCandlestickSeries] = useState<ISeriesApi<'Candlestick'> | null>(null);
  const [volumeSeries, setVolumeSeries] = useState<ISeriesApi<'Histogram'> | null>(null);
  const [timeframe, setTimeframe] = useState('1h');
  const [chartType, setChartType] = useState('candles');

  useEffect(() => {
    if (!chartContainerRef.current) return;

    const chartInstance = createChart(chartContainerRef.current, {
      layout: {
        background: { color: theme.palette.background.paper },
        textColor: theme.palette.text.primary,
      },
      grid: {
        vertLines: { color: theme.palette.divider },
        horzLines: { color: theme.palette.divider },
      },
      crosshair: {
        mode: 0,
      },
      rightPriceScale: {
        borderColor: theme.palette.divider,
      },
      timeScale: {
        borderColor: theme.palette.divider,
        timeVisible: true,
        secondsVisible: false,
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

    const candlesticks = chartInstance.addCandlestickSeries({
      upColor: theme.palette.success.main,
      downColor: theme.palette.error.main,
      borderVisible: false,
      wickUpColor: theme.palette.success.main,
      wickDownColor: theme.palette.error.main,
    });

    const volumes = chartInstance.addHistogramSeries({
      color: theme.palette.primary.main,
      priceFormat: {
        type: 'volume',
      },
      priceScaleId: '',
      scaleMargins: {
        top: 0.8,
        bottom: 0,
      },
    });

    setChart(chartInstance);
    setCandlestickSeries(candlesticks);
    setVolumeSeries(volumes);

    // Handle resize
    const handleResize = () => {
      if (chartContainerRef.current) {
        chartInstance.applyOptions({
          width: chartContainerRef.current.clientWidth,
          height: chartContainerRef.current.clientHeight,
        });
      }
    };

    window.addEventListener('resize', handleResize);

    // Initial size
    handleResize();

    return () => {
      window.removeEventListener('resize', handleResize);
      chartInstance.remove();
    };
  }, [theme]);

  useEffect(() => {
    if (!state.selectedPair || !candlestickSeries || !volumeSeries) return;

    // Simulate fetching historical data
    const fetchHistoricalData = () => {
      const data: ChartData[] = [];
      const now = new Date();
      const timeframeMinutes = parseTimeframe(timeframe);

      for (let i = 0; i < 100; i++) {
        const time = new Date(now.getTime() - i * timeframeMinutes * 60 * 1000);
        const basePrice = 100;
        const volatility = 2;
        const open = basePrice + (Math.random() - 0.5) * volatility;
        const close = basePrice + (Math.random() - 0.5) * volatility;
        const high = Math.max(open, close) + Math.random() * volatility;
        const low = Math.min(open, close) - Math.random() * volatility;
        const volume = Math.random() * 100;

        data.unshift({
          time: time.toISOString(),
          open,
          high,
          low,
          close,
          volume,
        });
      }

      return data;
    };

    const data = fetchHistoricalData();
    candlestickSeries.setData(data);
    volumeSeries.setData(
      data.map((d) => ({
        time: d.time,
        value: d.volume || 0,
        color:
          d.close >= d.open
            ? theme.palette.success.main
            : theme.palette.error.main,
      }))
    );

    // Simulate real-time updates
    const interval = setInterval(() => {
      const lastData = data[data.length - 1];
      const newPrice = lastData.close + (Math.random() - 0.5) * 0.5;
      const newVolume = Math.random() * 10;

      candlestickSeries.update({
        time: new Date().toISOString(),
        open: lastData.close,
        high: Math.max(lastData.close, newPrice),
        low: Math.min(lastData.close, newPrice),
        close: newPrice,
      });

      volumeSeries.update({
        time: new Date().toISOString(),
        value: newVolume,
        color:
          newPrice >= lastData.close
            ? theme.palette.success.main
            : theme.palette.error.main,
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [state.selectedPair, candlestickSeries, volumeSeries, timeframe, theme]);

  const parseTimeframe = (tf: string): number => {
    const value = parseInt(tf);
    const unit = tf.slice(-1);
    switch (unit) {
      case 'm':
        return value;
      case 'h':
        return value * 60;
      case 'd':
        return value * 60 * 24;
      default:
        return 60;
    }
  };

  const handleTimeframeChange = (event: React.ChangeEvent<{ value: unknown }>) => {
    setTimeframe(event.target.value as string);
  };

  const handleChartTypeChange = (
    event: React.MouseEvent<HTMLElement>,
    newType: string
  ) => {
    if (newType !== null) {
      setChartType(newType);
      // Update chart type logic here
    }
  };

  return (
    <Box sx={{ height: '100%', position: 'relative' }}>
      <Box
        sx={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          p: 1,
          zIndex: 1,
        }}
      >
        <ToggleButtonGroup
          value={chartType}
          exclusive
          onChange={handleChartTypeChange}
          size="small"
        >
          <ToggleButton value="candles">Candles</ToggleButton>
          <ToggleButton value="line">Line</ToggleButton>
        </ToggleButtonGroup>

        <FormControl size="small" sx={{ minWidth: 120 }}>
          <Select value={timeframe} onChange={handleTimeframeChange}>
            <MenuItem value="1m">1m</MenuItem>
            <MenuItem value="5m">5m</MenuItem>
            <MenuItem value="15m">15m</MenuItem>
            <MenuItem value="1h">1h</MenuItem>
            <MenuItem value="4h">4h</MenuItem>
            <MenuItem value="1d">1d</MenuItem>
          </Select>
        </FormControl>
      </Box>

      <Box
        ref={chartContainerRef}
        sx={{
          width: '100%',
          height: '100%',
          '& .tv-lightweight-charts': {
            width: '100% !important',
            height: '100% !important',
          },
        }}
      />
    </Box>
  );
} 