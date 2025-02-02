import React, { useState, useEffect } from 'react'
import { mockDataService } from '@/services/mockData'
import { debugService } from '@/services/debug'
import { Card, Table, Typography, Row, Col, Statistic, Space, Button } from 'antd'
import { ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

const { Title } = Typography

interface SignalData {
  symbol: string
  signal: 'buy' | 'sell' | 'hold'
  confidence: number
  timestamp: string
  reason: string
  indicators: {
    rsi: number
    macd: number
    volume: number
  }
}

interface PerformanceData {
  symbol: string
  winRate: number
  profitLoss: number
  totalTrades: number
  avgHoldingTime: string
  sharpeRatio: number
}

const Analysis = () => {
  const signalColumns: ColumnsType<SignalData> = [
    {
      title: '交易对',
      dataIndex: 'symbol',
      key: 'symbol',
    },
    {
      title: '信号',
      dataIndex: 'signal',
      key: 'signal',
      render: (signal) => (
        <span style={{
          color: signal === 'buy' ? '#52c41a' : signal === 'sell' ? '#f5222d' : '#8c8c8c'
        }}>
          {signal === 'buy' ? '买入' : signal === 'sell' ? '卖出' : '观望'}
        </span>
      ),
    },
    {
      title: '置信度',
      dataIndex: 'confidence',
      key: 'confidence',
      render: (confidence) => `${(confidence * 100).toFixed(1)}%`,
      sorter: (a, b) => a.confidence - b.confidence,
    },
    {
      title: '指标',
      dataIndex: 'indicators',
      key: 'indicators',
      render: (indicators) => (
        <Space direction="vertical" size="small">
          <span>RSI: {indicators.rsi.toFixed(2)}</span>
          <span>MACD: {indicators.macd.toFixed(2)}</span>
          <span>成交量: {indicators.volume.toLocaleString()}</span>
        </Space>
      ),
    },
    {
      title: '原因',
      dataIndex: 'reason',
      key: 'reason',
      width: 300,
    },
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      render: (timestamp) => new Date(timestamp).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space>
          <Button
            type={record.signal === 'buy' ? 'primary' : 'default'}
            size="small"
            disabled={record.signal === 'hold'}
          >
            执行
          </Button>
          <Button size="small">忽略</Button>
        </Space>
      ),
    },
  ]

  const performanceColumns: ColumnsType<PerformanceData> = [
    {
      title: '交易对',
      dataIndex: 'symbol',
      key: 'symbol',
    },
    {
      title: '胜率',
      dataIndex: 'winRate',
      key: 'winRate',
      render: (rate) => `${(rate * 100).toFixed(1)}%`,
      sorter: (a, b) => a.winRate - b.winRate,
    },
    {
      title: '盈亏',
      dataIndex: 'profitLoss',
      key: 'profitLoss',
      render: (pl) => (
        <span style={{ color: pl >= 0 ? '#52c41a' : '#f5222d' }}>
          {pl >= 0 ? '+' : ''}{pl.toFixed(2)} USD
        </span>
      ),
      sorter: (a, b) => a.profitLoss - b.profitLoss,
    },
    {
      title: '交易次数',
      dataIndex: 'totalTrades',
      key: 'totalTrades',
      sorter: (a, b) => a.totalTrades - b.totalTrades,
    },
    {
      title: '平均持仓时间',
      dataIndex: 'avgHoldingTime',
      key: 'avgHoldingTime',
    },
    {
      title: '夏普比率',
      dataIndex: 'sharpeRatio',
      key: 'sharpeRatio',
      render: (ratio) => ratio.toFixed(2),
      sorter: (a, b) => a.sharpeRatio - b.sharpeRatio,
    },
  ]

  const [signals, setSignals] = useState<SignalData[]>([])
  const [performance, setPerformance] = useState<PerformanceData[]>([])

  useEffect(() => {
    debugService.info('Analysis component mounted')

    // 模拟信号生成
    const generateSignal = () => {
      const symbols = ['BTC-USD', 'ETH-USD']
      const signalTypes: ('buy' | 'sell' | 'hold')[] = ['buy', 'sell', 'hold']
      
      return symbols.map(symbol => ({
        symbol,
        signal: signalTypes[Math.floor(Math.random() * signalTypes.length)],
        confidence: Math.random() * 0.5 + 0.5, // 0.5 - 1.0
        timestamp: new Date().toISOString(),
        reason: '基于市场数据分析',
        indicators: {
          rsi: Math.random() * 100,
          macd: (Math.random() - 0.5) * 2,
          volume: Math.random() * 1000000 + 500000,
        },
      }))
    }

    // 模拟性能数据
    const generatePerformance = () => {
      const symbols = ['BTC-USD', 'ETH-USD']
      return symbols.map(symbol => ({
        symbol,
        winRate: Math.random() * 0.3 + 0.5, // 0.5 - 0.8
        profitLoss: (Math.random() - 0.3) * 2000, // -600 to 1400
        totalTrades: Math.floor(Math.random() * 50) + 20,
        avgHoldingTime: `${Math.floor(Math.random() * 4)}h ${Math.floor(Math.random() * 60)}m`,
        sharpeRatio: Math.random() * 2 + 0.5,
      }))
    }

    // 初始数据
    setSignals(generateSignal())
    setPerformance(generatePerformance())

    // 定期更新数据
    const signalInterval = setInterval(() => {
      const newSignals = generateSignal()
      setSignals(newSignals)
      debugService.debug('Signals updated', newSignals)
    }, 5000)

    const performanceInterval = setInterval(() => {
      const newPerformance = generatePerformance()
      setPerformance(newPerformance)
      debugService.debug('Performance updated', newPerformance)
    }, 10000)

    return () => {
      clearInterval(signalInterval)
      clearInterval(performanceInterval)
      debugService.info('Analysis component unmounted')
    }
  }, [])

  return (
    <div>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Row gutter={[16, 16]}>
          <Col span={6}>
            <Card>
              <Statistic
                title="总盈亏"
                value={929.75}
                precision={2}
                valueStyle={{ color: '#52c41a' }}
                prefix="+"
                suffix="USD"
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="总交易次数"
                value={77}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="平均胜率"
                value={61.5}
                suffix="%"
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="活跃信号"
                value={2}
              />
            </Card>
          </Col>
        </Row>

        <Card title="交易信号">
          <Table<SignalData>
            columns={signalColumns}
            dataSource={signals}
            rowKey="symbol"
            pagination={false}
          />
        </Card>

        <Card title="策略表现">
          <Table<PerformanceData>
            columns={performanceColumns}
            dataSource={performance}
            rowKey="symbol"
            pagination={false}
          />
        </Card>
      </Space>
    </div>
  )
}

export default Analysis
