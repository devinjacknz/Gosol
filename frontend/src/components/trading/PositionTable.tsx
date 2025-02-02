import { Table } from 'antd'

interface PositionTableProps {
  data: any[]
  loading?: boolean
}

export const PositionTable = ({ data, loading }: PositionTableProps) => {
  const columns = [
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
        const color = side === 'long' ? 'green' : 'red'
        return <span style={{ color }}>{side === 'long' ? '多' : '空'}</span>
      },
    },
    {
      title: '数量',
      dataIndex: 'size',
      key: 'size',
      render: (size: number) => size.toFixed(4),
    },
    {
      title: '开仓价',
      dataIndex: 'entryPrice',
      key: 'entryPrice',
      render: (price: number) =>
        price.toLocaleString('en-US', {
          style: 'currency',
          currency: 'USD',
        }),
    },
    {
      title: '当前价',
      dataIndex: 'markPrice',
      key: 'markPrice',
      render: (price: number) =>
        price.toLocaleString('en-US', {
          style: 'currency',
          currency: 'USD',
        }),
    },
    {
      title: '未实现盈亏',
      dataIndex: 'unrealizedPnl',
      key: 'unrealizedPnl',
      render: (pnl: number) => {
        const color = pnl >= 0 ? 'green' : 'red'
        return (
          <span style={{ color }}>
            {pnl.toLocaleString('en-US', {
              style: 'currency',
              currency: 'USD',
              signDisplay: 'always',
            })}
          </span>
        )
      },
    },
    {
      title: '收益率',
      dataIndex: 'roe',
      key: 'roe',
      render: (roe: number) => {
        const color = roe >= 0 ? 'green' : 'red'
        return (
          <span style={{ color }}>
            {(roe * 100).toFixed(2)}%
          </span>
        )
      },
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