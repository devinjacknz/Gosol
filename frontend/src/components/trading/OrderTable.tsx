import { Table, Button, Space } from 'antd'
import { useAppDispatch } from '@/hooks/store'
import { cancelOrder } from '@/store/trading/tradingSlice'

interface OrderTableProps {
  data: any[]
  loading?: boolean
}

export const OrderTable = ({ data, loading }: OrderTableProps) => {
  const dispatch = useAppDispatch()

  const columns = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      render: (timestamp: string) => new Date(timestamp).toLocaleString(),
    },
    {
      title: '交易对',
      dataIndex: 'symbol',
      key: 'symbol',
    },
    {
      title: '方向',
      dataIndex: 'side',
      key: 'side',
      render: (side: string) => {
        const color = side === 'buy' ? 'green' : 'red'
        return <span style={{ color }}>{side === 'buy' ? '买入' : '卖出'}</span>
      },
    },
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
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const color =
          status === 'NEW'
            ? 'blue'
            : status === 'FILLED'
            ? 'green'
            : status === 'CANCELED'
            ? 'gray'
            : 'red'
        return <span style={{ color }}>{status}</span>
      },
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: any) => (
        <Space>
          <Button
            size="small"
            onClick={() => dispatch(cancelOrder(record.id))}
            disabled={record.status !== 'NEW'}
          >
            取消
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <Table
      columns={columns}
      dataSource={data}
      rowKey="id"
      size="small"
      pagination={false}
      loading={loading}
      scroll={{ y: 200 }}
    />
  )
} 