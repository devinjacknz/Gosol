import { useEffect } from 'react'
import { Layout, Card, Select } from 'antd'
import { useAppDispatch, useAppSelector } from '@/hooks/store'
import type { RootState } from '@/store'
import { setSelectedSymbol, fetchOrders, fetchPositions } from '@/store/trading/tradingSlice'
import { TradingChart } from '@/components/Chart/TradingChart'
import { OrderForm } from '@/components/trading/OrderForm'
import { OrderTable } from '@/components/trading/OrderTable'
import { PositionTable } from '@/components/trading/PositionTable'
import { useWebSocket } from '@/hooks/useWebSocket'

const { Sider } = Layout

const SYMBOLS = [
  { label: 'BTC/USDT', value: 'BTC/USDT' },
  { label: 'ETH/USDT', value: 'ETH/USDT' },
]

const TradingView = () => {
  const dispatch = useAppDispatch()
  const {
    marketData,
    orders,
    positions,
    selectedSymbol,
    loading,
  } = useAppSelector((state: RootState) => state.trading)

  // 初始化 WebSocket 连接
  useWebSocket(selectedSymbol)

  // 获取初始数据
  useEffect(() => {
    dispatch(fetchOrders())
    dispatch(fetchPositions())
  }, [dispatch])

  // 当前交易对的最新价格
  const lastPrice = marketData[selectedSymbol]?.lastPrice || 0

  return (
    <Layout className="trading-layout">
      <Layout className="trading-content">
        <div className="mb-4 flex items-center justify-between">
          <Select
            value={selectedSymbol}
            onChange={(value) => dispatch(setSelectedSymbol(value))}
            options={SYMBOLS}
            style={{ width: 200 }}
          />
          <div className="text-2xl font-bold">
            {lastPrice.toLocaleString('en-US', {
              style: 'currency',
              currency: 'USD',
            })}
          </div>
        </div>

        <TradingChart
          data={marketData[selectedSymbol]?.klines || []}
          height={600}
        />

        <OrderForm />
      </Layout>

      <Sider width={400} className="bg-white dark:bg-gray-800 p-4">
        <Card title="当前订单" className="mb-4">
          <OrderTable />
        </Card>

        <Card title="当前持仓">
          <PositionTable />
        </Card>
      </Sider>
    </Layout>
  )
}

export default TradingView      