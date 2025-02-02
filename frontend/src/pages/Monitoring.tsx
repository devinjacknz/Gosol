import React, { useState, useEffect } from 'react'
import { mockDataService } from '@/services/mockData'
import { debugService } from '@/services/debug'
import { Card, Table, Typography, Row, Col, Statistic, Space, Tag, Button, Select } from 'antd'
import { CheckCircleOutlined, CloseCircleOutlined, SyncOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import type { WebSocketStatus, AlertConfig } from '@/types/status'
import type { RootState } from '@/store'

const { Title } = Typography

interface SystemMetric {
  name: string
  value: number
  status: 'normal' | 'warning' | 'error'
  unit: string
  timestamp: string
}

interface TradingMetric {
  exchange: string
  status: 'online' | 'offline' | 'error'
  latency: number
  ordersPerSecond: number
  errorRate: number
  lastSync: string
}

interface SystemEvent {
  id: string
  type: 'info' | 'warning' | 'error'
  message: string
  timestamp: string
  details: string
}

const Monitoring = () => {
  const [timeRange, setTimeRange] = useState('1h')

  const [systemMetrics, setSystemMetrics] = useState(mockDataService.getSystemMetrics())
  const [tradingMetrics, setTradingMetrics] = useState(mockDataService.getTradingMetrics())
  const [events, setEvents] = useState<SystemEvent[]>([])

  useEffect(() => {
    debugService.info('Monitoring component mounted')

    const systemMetricsCleanup = mockDataService.simulateSystemMetricsUpdate((data) => {
      setSystemMetrics(data)
    })

    const tradingMetricsCleanup = mockDataService.simulateTradingMetricsUpdate((data) => {
      setTradingMetrics(data)
    })

    // 模拟系统事件
    const generateEvent = () => {
      const types: SystemEvent['type'][] = ['info', 'warning', 'error']
      const type = types[Math.floor(Math.random() * types.length)]
      
      const messages = {
        error: [
          '内存使用率超过阈值',
          '系统延迟异常',
          'API 请求失败',
        ],
        warning: [
          '交易所连接不稳定',
          '订单执行延迟',
          '系统负载较高',
        ],
        info: [
          '系统自动平仓',
          '数据同步完成',
          '配置更新成功',
        ],
      }

      const message = messages[type][Math.floor(Math.random() * messages[type].length)]

      return {
        id: Date.now().toString(),
        type,
        message,
        timestamp: new Date().toISOString(),
        details: `${message}的详细信息`,
      }
    }

    // 每30秒生成一个新事件
    const eventInterval = setInterval(() => {
      const newEvent = generateEvent()
      setEvents(prev => [newEvent, ...prev].slice(0, 10)) // 保留最新的10条
      debugService.debug('New system event', newEvent)
    }, 30000)

    // 初始事件
    setEvents([
      {
        id: '1',
        type: 'error',
        message: '内存使用率超过阈值',
        timestamp: new Date().toISOString(),
        details: '内存使用率达到72%, 超过警告阈值70%',
      },
      {
        id: '2',
        type: 'info',
        message: '系统自动平仓',
        timestamp: new Date().toISOString(),
        details: 'BTC-USD 多仓触发止损, 平仓价格43250',
      },
    ])

    return () => {
      systemMetricsCleanup()
      tradingMetricsCleanup()
      clearInterval(eventInterval)
      debugService.info('Monitoring component unmounted')
    }
  }, [])

  const tradingColumns: ColumnsType<TradingMetric> = [
    {
      title: '交易所',
      dataIndex: 'exchange',
      key: 'exchange',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => {
        const config: Record<string, WebSocketStatus> = {
          online: { color: '#52c41a', icon: <CheckCircleOutlined />, status: 'online' },
          offline: { color: '#f5222d', icon: <CloseCircleOutlined />, status: 'offline' },
          error: { color: '#faad14', icon: <SyncOutlined spin />, status: 'error' },
        }
        return (
          <Tag color={config[status].color}>
            {config[status].icon} {status.toUpperCase()}
          </Tag>
        )
      },
    },
    {
      title: '延迟',
      dataIndex: 'latency',
      key: 'latency',
      render: (latency) => `${latency}ms`,
      sorter: (a, b) => a.latency - b.latency,
    },
    {
      title: '每秒订单',
      dataIndex: 'ordersPerSecond',
      key: 'ordersPerSecond',
      render: (ops) => ops.toFixed(1),
      sorter: (a, b) => a.ordersPerSecond - b.ordersPerSecond,
    },
    {
      title: '错误率',
      dataIndex: 'errorRate',
      key: 'errorRate',
      render: (rate) => `${(rate * 100).toFixed(1)}%`,
      sorter: (a, b) => a.errorRate - b.errorRate,
    },
    {
      title: '最后同步',
      dataIndex: 'lastSync',
      key: 'lastSync',
      render: (time) => new Date(time).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      render: () => (
        <Space>
          <Button size="small">重连</Button>
          <Button size="small">查看日志</Button>
        </Space>
      ),
    },
  ]

  const eventColumns: ColumnsType<SystemEvent> = [
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type) => {
        const config: Record<string, AlertConfig> = {
          info: { color: '#1890ff', text: '信息', type: 'info' },
          warning: { color: '#faad14', text: '警告', type: 'warning' },
          error: { color: '#f5222d', text: '错误', type: 'error' },
        }
        return <Tag color={config[type].color}>{config[type].text}</Tag>
      },
    },
    {
      title: '消息',
      dataIndex: 'message',
      key: 'message',
    },
    {
      title: '详情',
      dataIndex: 'details',
      key: 'details',
      width: 300,
    },
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      render: (time) => new Date(time).toLocaleString(),
    },
  ]

  return (
    <div>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Space style={{ marginBottom: 16 }}>
          <Select
            value={timeRange}
            onChange={setTimeRange}
            style={{ width: 120 }}
            options={[
              { value: '1h', label: '1小时' },
              { value: '4h', label: '4小时' },
              { value: '1d', label: '1天' },
              { value: '1w', label: '1周' },
            ]}
          />
          <Button type="primary">刷新</Button>
        </Space>

        <Row gutter={[16, 16]}>
          {systemMetrics.map((metric) => (
            <Col span={6} key={metric.name}>
              <Card>
                <Statistic
                  title={metric.name}
                  value={metric.value}
                  suffix={metric.unit}
                  valueStyle={{
                    color:
                      metric.status === 'normal'
                        ? '#52c41a'
                        : metric.status === 'warning'
                        ? '#faad14'
                        : '#f5222d',
                  }}
                />
              </Card>
            </Col>
          ))}
        </Row>

        <Card title="交易所状态">
          <Table<TradingMetric>
            columns={tradingColumns}
            dataSource={tradingMetrics}
            rowKey="exchange"
            pagination={false}
          />
        </Card>

        <Card title="系统事件">
          <Table<SystemEvent>
            columns={eventColumns}
            dataSource={events}
            rowKey="id"
            pagination={{ pageSize: 5 }}
          />
        </Card>
      </Space>
    </div>
  )
}

export default Monitoring
