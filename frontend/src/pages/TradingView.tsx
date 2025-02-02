import React, { useState, useEffect } from 'react'
import { mockDataService } from '@/services/mockData'
import { debugService } from '@/services/debug'
import { Card, Table, Select, Space, Button, InputNumber, Typography, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'

const { Title } = Typography

interface MarketData {
  symbol: string
  price: number
  volume: number
  change24h: number
  high24h: number
  low24h: number
  timestamp: string
}

interface Position {
  symbol: string
  size: number
  entryPrice: number
  markPrice: number
  pnl: number
  roe: number
}

const TradingView = () => {
  const [selectedSymbol, setSelectedSymbol] = useState('BTC-USD')
  const [orderSize, setOrderSize] = useState(0.01)
  const [leverage, setLeverage] = useState(1)

  const marketColumns: ColumnsType<MarketData> = [
    {
      title: '交易对',
      dataIndex: 'symbol',
      key: 'symbol',
    },
    {
      title: '最新价',
      dataIndex: 'price',
      key: 'price',
      render: (price) => price.toLocaleString(),
    },
    {
      title: '24h涨跌',
      dataIndex: 'change24h',
      key: 'change24h',
      render: (change) => (
        <span style={{ color: change >= 0 ? '#52c41a' : '#f5222d' }}>
          {change >= 0 ? '+' : ''}{change.toFixed(2)}%
        </span>
      ),
    },
    {
      title: '24h成交量',
      dataIndex: 'volume',
      key: 'volume',
      render: (volume) => volume.toLocaleString(),
    },
    {
      title: '24h最高',
      dataIndex: 'high24h',
      key: 'high24h',
      render: (price) => price.toLocaleString(),
    },
    {
      title: '24h最低',
      dataIndex: 'low24h',
      key: 'low24h',
      render: (price) => price.toLocaleString(),
    },
  ]

  const positionColumns: ColumnsType<Position> = [
    {
      title: '交易对',
      dataIndex: 'symbol',
      key: 'symbol',
    },
    {
      title: '仓位大小',
      dataIndex: 'size',
      key: 'size',
    },
    {
      title: '开仓价格',
      dataIndex: 'entryPrice',
      key: 'entryPrice',
      render: (price) => price.toLocaleString(),
    },
    {
      title: '标记价格',
      dataIndex: 'markPrice',
      key: 'markPrice',
      render: (price) => price.toLocaleString(),
    },
    {
      title: '未实现盈亏',
      dataIndex: 'pnl',
      key: 'pnl',
      render: (pnl) => (
        <span style={{ color: pnl >= 0 ? '#52c41a' : '#f5222d' }}>
          {pnl >= 0 ? '+' : ''}{pnl.toFixed(2)} USD
        </span>
      ),
    },
    {
      title: 'ROE',
      dataIndex: 'roe',
      key: 'roe',
      render: (roe) => (
        <span style={{ color: roe >= 0 ? '#52c41a' : '#f5222d' }}>
          {roe >= 0 ? '+' : ''}{roe.toFixed(2)}%
        </span>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: () => (
        <Space>
          <Button size="small" danger>平仓</Button>
          <Button size="small">调整杠杆</Button>
        </Space>
      ),
    },
  ]

  const [marketData, setMarketData] = useState(mockDataService.getMarketData())
  const [positions, setPositions] = useState(mockDataService.getPositions())

  useEffect(() => {
    debugService.info('TradingView component mounted')
    
    const cleanup = mockDataService.simulateMarketDataUpdate((data) => {
      setMarketData(data)
    })

    return () => {
      cleanup()
      debugService.info('TradingView component unmounted')
    }
  }, [])

  const handleTrade = (direction: 'long' | 'short') => {
    debugService.info('Trade executed', {
      direction,
      symbol: selectedSymbol,
      size: orderSize,
      leverage,
    })

    // 模拟下单延迟
    setTimeout(() => {
      const newPosition = {
        symbol: selectedSymbol,
        size: orderSize * (direction === 'long' ? 1 : -1),
        entryPrice: marketData.find(m => m.symbol === selectedSymbol)?.price || 0,
        markPrice: marketData.find(m => m.symbol === selectedSymbol)?.price || 0,
        pnl: 0,
        roe: 0,
      }

      setPositions([...positions, newPosition])
      debugService.info('Position opened', newPosition)
    }, 500)
  }

  return (
    <div>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Card title="市场数据">
          <Table<MarketData>
            columns={marketColumns}
            dataSource={marketData}
            rowKey="symbol"
            pagination={false}
          />
        </Card>

        <Card title="交易操作">
          <Space direction="vertical" style={{ width: '100%' }}>
            <Space>
              <Select
                value={selectedSymbol}
                onChange={setSelectedSymbol}
                style={{ width: 120 }}
                options={[
                  { value: 'BTC-USD', label: 'BTC/USD' },
                  { value: 'ETH-USD', label: 'ETH/USD' },
                  { value: 'SOL-USD', label: 'SOL/USD' },
                ]}
              />
              <InputNumber
                value={orderSize}
                onChange={(value) => setOrderSize(value || 0.01)}
                style={{ width: 120 }}
                min={0.001}
                step={0.001}
                precision={3}
                addonBefore="数量"
              />
              <InputNumber
                value={leverage}
                onChange={(value) => setLeverage(value || 1)}
                style={{ width: 120 }}
                min={1}
                max={100}
                step={1}
                addonBefore="杠杆"
              />
              <Button type="primary" onClick={() => handleTrade('long')}>做多</Button>
              <Button danger onClick={() => handleTrade('short')}>做空</Button>
            </Space>
            <Space>
              <Tag>可用保证金: 10000 USD</Tag>
              <Tag>仓位价值: 4325.05 USD</Tag>
              <Tag>维持保证金率: 0.5%</Tag>
            </Space>
          </Space>
        </Card>

        <Card title="当前持仓">
          <Table<Position>
            columns={positionColumns}
            dataSource={positions}
            rowKey="symbol"
            pagination={false}
          />
        </Card>
      </Space>
    </div>
  )
}

export default TradingView
