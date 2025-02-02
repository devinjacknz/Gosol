import { useEffect } from 'react'
import { Layout, Card, Table, Alert } from 'antd'
import { useAppDispatch, useAppSelector } from '@/hooks/store'
import { fetchMetrics, fetchAlerts } from '@/store/monitoring/monitoringSlice'
import { MetricCard } from '@/components/Monitoring/MetricCard'

const { Content } = Layout

const Monitoring = () => {
  const dispatch = useAppDispatch()
  const { metrics, alerts, loading, error } = useAppSelector(
    (state) => state.monitoring
  )

  useEffect(() => {
    const fetchData = async () => {
      await dispatch(fetchMetrics())
      await dispatch(fetchAlerts())
    }
    fetchData()

    const interval = setInterval(fetchData, 30000) // 每30秒更新一次
    return () => clearInterval(interval)
  }, [dispatch])

  const alertColumns = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      render: (timestamp: string) =>
        new Date(timestamp).toLocaleString(),
    },
    {
      title: '级别',
      dataIndex: 'level',
      key: 'level',
      render: (level: string) => {
        const color =
          level === 'error'
            ? 'red'
            : level === 'warning'
            ? 'orange'
            : 'blue'
        return <span style={{ color }}>{level.toUpperCase()}</span>
      },
    },
    {
      title: '消息',
      dataIndex: 'message',
      key: 'message',
    },
    {
      title: '来源',
      dataIndex: 'source',
      key: 'source',
    },
  ]

  return (
    <Layout>
      <Content style={{ padding: '0 24px' }}>
        {error && (
          <Alert
            message="监控错误"
            description={error}
            type="error"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        <div className="monitoring-grid">
          <MetricCard
            title="LLM 请求数"
            value={metrics?.llmRequestCount || 0}
            precision={0}
            suffix="次/分钟"
            loading={loading}
          />
          <MetricCard
            title="平均响应时间"
            value={metrics?.avgResponseTime || 0}
            precision={2}
            suffix="ms"
            trend={{
              value:
                ((metrics?.avgResponseTime || 0) - 1000) / 10,
              isUpGood: false,
            }}
            loading={loading}
          />
          <MetricCard
            title="错误率"
            value={metrics?.errorRate || 0}
            precision={2}
            suffix="%"
            trend={{
              value: metrics?.errorRate || 0,
              isUpGood: false,
            }}
            loading={loading}
          />
          <MetricCard
            title="Token 使用量"
            value={metrics?.tokenUsage || 0}
            precision={0}
            suffix="个/分钟"
            loading={loading}
          />
        </div>

        <Card title="系统指标" style={{ marginTop: 16 }}>
          <div className="grid grid-cols-3 gap-4">
            <MetricCard
              title="CPU 使用率"
              value={metrics?.cpuUsage || 0}
              precision={2}
              suffix="%"
              trend={{
                value: ((metrics?.cpuUsage || 0) - 50) / 50 * 100,
                isUpGood: false,
              }}
              loading={loading}
            />
            <MetricCard
              title="内存使用率"
              value={metrics?.memoryUsage || 0}
              precision={2}
              suffix="%"
              trend={{
                value: ((metrics?.memoryUsage || 0) - 50) / 50 * 100,
                isUpGood: false,
              }}
              loading={loading}
            />
            <MetricCard
              title="磁盘使用率"
              value={metrics?.diskUsage || 0}
              precision={2}
              suffix="%"
              trend={{
                value: ((metrics?.diskUsage || 0) - 50) / 50 * 100,
                isUpGood: false,
              }}
              loading={loading}
            />
          </div>
        </Card>

        <Card title="告警信息" style={{ marginTop: 16 }}>
          <Table
            columns={alertColumns}
            dataSource={alerts}
            pagination={{ pageSize: 10 }}
            loading={loading}
          />
        </Card>

        <Card title="性能指标" style={{ marginTop: 16 }}>
          <div className="grid grid-cols-3 gap-4">
            <Card title="LLM 性能">
              <MetricCard
                title="平均生成时间"
                value={metrics?.llmGenerationTime || 0}
                precision={2}
                suffix="ms"
                loading={loading}
              />
              <MetricCard
                title="Token 生成速率"
                value={metrics?.tokenGenerationRate || 0}
                precision={2}
                suffix="tokens/s"
                loading={loading}
                style={{ marginTop: 16 }}
              />
            </Card>
            <Card title="API 性能">
              <MetricCard
                title="请求成功率"
                value={metrics?.apiSuccessRate || 0}
                precision={2}
                suffix="%"
                loading={loading}
              />
              <MetricCard
                title="平均延迟"
                value={metrics?.apiLatency || 0}
                precision={2}
                suffix="ms"
                loading={loading}
                style={{ marginTop: 16 }}
              />
            </Card>
            <Card title="系统健康度">
              <MetricCard
                title="系统可用性"
                value={metrics?.systemAvailability || 0}
                precision={2}
                suffix="%"
                loading={loading}
              />
              <MetricCard
                title="错误数量"
                value={metrics?.errorCount || 0}
                precision={0}
                loading={loading}
                style={{ marginTop: 16 }}
              />
            </Card>
          </div>
        </Card>
      </Content>
    </Layout>
  )
}

export default Monitoring 