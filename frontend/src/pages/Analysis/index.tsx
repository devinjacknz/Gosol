import { useEffect } from 'react'
import { Layout, Card, Select, Table, Button, Input, Space, Spin } from 'antd'
import { useAppDispatch, useAppSelector } from '@/hooks/store'
import { analyzeMarket, generateReport } from '@/store/analysis/analysisSlice'
import { setSelectedSymbol } from '@/store/trading/tradingSlice'
import { TradingChart } from '@/components/Chart/TradingChart'
import { useWebSocket } from '@/hooks/useWebSocket'

const { Content, Sider } = Layout
const { TextArea } = Input

const SYMBOLS = [
  { label: 'BTC/USDT', value: 'BTC/USDT' },
  { label: 'ETH/USDT', value: 'ETH/USDT' },
]

const TIMEFRAMES = [
  { label: '1分钟', value: '1m' },
  { label: '5分钟', value: '5m' },
  { label: '15分钟', value: '15m' },
  { label: '1小时', value: '1h' },
  { label: '4小时', value: '4h' },
  { label: '1天', value: '1d' },
]

const Analysis = () => {
  const dispatch = useAppDispatch()
  const {
    selectedSymbol,
    marketData,
  } = useAppSelector((state) => state.trading)
  const {
    technicalIndicators,
    marketAnalysis,
    llmAnalysis,
    loading,
    error,
  } = useAppSelector((state) => state.analysis)

  // 初始化 WebSocket 连接
  useWebSocket(selectedSymbol)

  // 获取初始分析数据
  useEffect(() => {
    dispatch(analyzeMarket({ symbol: selectedSymbol, timeframe: '1h' }))
  }, [dispatch, selectedSymbol])

  const indicatorColumns = [
    {
      title: '指标',
      dataIndex: 'indicator',
      key: 'indicator',
    },
    {
      title: '值',
      dataIndex: 'value',
      key: 'value',
    },
    {
      title: '信号',
      dataIndex: 'signal',
      key: 'signal',
      render: (signal: string) => {
        const color = signal === 'buy' ? 'green' : signal === 'sell' ? 'red' : 'gray'
        return <span style={{ color }}>{signal.toUpperCase()}</span>
      },
    },
  ]

  return (
    <Layout className="analysis-layout">
      <Content className="analysis-main">
        <div className="mb-4 flex items-center justify-between">
          <Space>
            <Select
              value={selectedSymbol}
              onChange={(value) => dispatch(setSelectedSymbol(value))}
              options={SYMBOLS}
              style={{ width: 200 }}
            />
            <Select
              defaultValue="1h"
              options={TIMEFRAMES}
              style={{ width: 120 }}
              onChange={(timeframe) =>
                dispatch(analyzeMarket({ symbol: selectedSymbol, timeframe }))
              }
            />
          </Space>
        </div>

        <TradingChart
          data={marketData[selectedSymbol]?.klines || []}
          height={400}
        />

        <Card title="技术指标分析" className="mt-4">
          <Table
            columns={indicatorColumns}
            dataSource={technicalIndicators}
            pagination={false}
            loading={loading}
          />
        </Card>

        <Card title="LLM 分析" className="mt-4">
          <Space direction="vertical" style={{ width: '100%' }}>
            <TextArea
              rows={4}
              placeholder="输入分析需求（可选）"
              className="mb-4"
            />
            <Button
              type="primary"
              onClick={() =>
                dispatch(
                  generateReport({
                    symbol: selectedSymbol,
                    timeframe: '1h',
                  })
                )
              }
              loading={loading}
            >
              生成分析报告
            </Button>
            {loading ? (
              <div className="text-center py-4">
                <Spin />
                <div className="mt-2">正在分析市场数据...</div>
              </div>
            ) : (
              <pre className="whitespace-pre-wrap">{llmAnalysis}</pre>
            )}
          </Space>
        </Card>
      </Content>

      <Sider width={300} className="analysis-sidebar">
        <Card title="市场概览">
          {marketAnalysis && (
            <>
              <div className="mb-4">
                <div className="text-sm text-gray-500">趋势</div>
                <div>{marketAnalysis.trend}</div>
              </div>
              <div className="mb-4">
                <div className="text-sm text-gray-500">支撑位</div>
                <div>{marketAnalysis.support}</div>
              </div>
              <div className="mb-4">
                <div className="text-sm text-gray-500">阻力位</div>
                <div>{marketAnalysis.resistance}</div>
              </div>
              <div className="mb-4">
                <div className="text-sm text-gray-500">波动率</div>
                <div>{marketAnalysis.volatility}</div>
              </div>
              <div className="mb-4">
                <div className="text-sm text-gray-500">成交量</div>
                <div>{marketAnalysis.volume}</div>
              </div>
              <div>
                <div className="text-sm text-gray-500">建议</div>
                <div>{marketAnalysis.recommendation}</div>
              </div>
            </>
          )}
        </Card>
      </Sider>
    </Layout>
  )
}

export default Analysis 