import { Card, Statistic } from 'antd'
import { ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons'

interface MetricCardProps {
  title: string
  value: number
  unit?: string
  precision?: number
  prefix?: React.ReactNode
  suffix?: React.ReactNode
  trend?: {
    value: number
    isUpGood?: boolean
  }
  loading?: boolean
  style?: React.CSSProperties
}

export const MetricCard = ({
  title,
  value,
  unit,
  precision = 2,
  prefix,
  suffix,
  trend,
  loading,
  style,
}: MetricCardProps) => {
  const getTrendColor = (trendValue: number, isUpGood = true) => {
    if (trendValue > 0) {
      return isUpGood ? '#52c41a' : '#ff4d4f'
    }
    if (trendValue < 0) {
      return isUpGood ? '#ff4d4f' : '#52c41a'
    }
    return '#8c8c8c'
  }

  const renderTrend = () => {
    if (!trend) return null

    const { value: trendValue, isUpGood = true } = trend
    const color = getTrendColor(trendValue, isUpGood)
    const icon =
      trendValue > 0 ? (
        <ArrowUpOutlined style={{ color }} />
      ) : (
        <ArrowDownOutlined style={{ color }} />
      )

    return (
      <div className="metric-trend" style={{ color }}>
        {icon} {Math.abs(trendValue)}%
      </div>
    )
  }

  return (
    <Card className="metric-card" loading={loading} style={style}>
      <Statistic
        title={title}
        value={value}
        precision={precision}
        prefix={prefix}
        suffix={unit || suffix}
        valueStyle={{
          color: trend ? getTrendColor(trend.value, trend.isUpGood) : undefined,
        }}
      />
      {renderTrend()}
    </Card>
  )
}  