import { Table, Row, Col, Typography } from 'antd'
import { useAppSelector } from '@/hooks/store'

const { Text } = Typography

interface OrderBookProps {
  symbol: string
  depth?: number
}

export const OrderBook = ({ symbol, depth = 20 }: OrderBookProps) => {
  const { marketData } = useAppSelector((state) => state.trading)
  const orderBook = marketData[symbol]?.orderBook || { bids: [], asks: [] }

  const columns = [
    {
      title: '价格',
      dataIndex: 'price',
      key: 'price',
      render: (price: number) =>
        price.toLocaleString('en-US', {
          style: 'currency',
          currency: 'USD',
        }),
    },
    {
      title: '数量',
      dataIndex: 'size',
      key: 'size',
      render: (size: number) => size.toFixed(4),
    },
    {
      title: '累计',
      dataIndex: 'total',
      key: 'total',
      render: (total: number) => total.toFixed(4),
    },
  ]

  // 计算累计数量
  const processOrders = (orders: [number, number][]) => {
    let total = 0
    return orders.slice(0, depth).map(([price, size]) => {
      total += size
      return {
        price,
        size,
        total,
      }
    })
  }

  const bids = processOrders(orderBook.bids)
  const asks = processOrders(orderBook.asks).reverse()

  return (
    <Row gutter={16}>
      <Col span={12}>
        <Text type="success">买盘</Text>
        <Table
          columns={columns}
          dataSource={bids}
          size="small"
          pagination={false}
          rowKey="price"
          scroll={{ y: 400 }}
        />
      </Col>
      <Col span={12}>
        <Text type="danger">卖盘</Text>
        <Table
          columns={columns}
          dataSource={asks}
          size="small"
          pagination={false}
          rowKey="price"
          scroll={{ y: 400 }}
        />
      </Col>
    </Row>
  )
} 