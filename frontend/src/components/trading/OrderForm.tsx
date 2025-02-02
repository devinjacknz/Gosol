import { Form, Select, InputNumber, Button, Space, Tooltip } from 'antd'
import { InfoCircleOutlined } from '@ant-design/icons'
import { useAppDispatch } from '@/hooks/store'
import { placeOrder } from '@/store/trading/tradingSlice'

interface OrderFormProps {
  symbol: string
  lastPrice: number
  loading?: boolean
}

export const OrderForm = ({ symbol, lastPrice, loading }: OrderFormProps) => {
  const [form] = Form.useForm()
  const dispatch = useAppDispatch()

  const handleSubmit = async (values: any) => {
    try {
      await dispatch(placeOrder({
        ...values,
        symbol,
      })).unwrap()
      form.resetFields()
    } catch (error) {
      console.error('Place order failed:', error)
    }
  }

  const handleQuickAmount = (percent: number) => {
    const balance = 10000 // TODO: 从 store 获取余额
    const price = form.getFieldValue('price') || lastPrice
    const amount = (balance * percent) / price
    form.setFieldsValue({ size: amount })
  }

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={handleSubmit}
      className="order-form"
      initialValues={{
        side: 'buy',
        price: lastPrice,
      }}
    >
      <Form.Item
        name="side"
        label="交易方向"
        rules={[{ required: true, message: '请选择交易方向' }]}
      >
        <Select>
          <Select.Option value="buy">买入</Select.Option>
          <Select.Option value="sell">卖出</Select.Option>
        </Select>
      </Form.Item>

      <Form.Item
        name="price"
        label="价格"
        rules={[
          { required: true, message: '请输入价格' },
          { type: 'number', min: 0, message: '价格必须大于0' },
        ]}
      >
        <InputNumber
          style={{ width: '100%' }}
          precision={2}
          step={0.01}
          placeholder="输入价格"
          addonAfter="USDT"
        />
      </Form.Item>

      <Form.Item
        name="size"
        label={
          <Space>
            <span>数量</span>
            <Tooltip title="可用余额: 10000 USDT">
              <InfoCircleOutlined />
            </Tooltip>
          </Space>
        }
        rules={[
          { required: true, message: '请输入数量' },
          { type: 'number', min: 0, message: '数量必须大于0' },
        ]}
      >
        <InputNumber
          style={{ width: '100%' }}
          precision={4}
          step={0.0001}
          placeholder="输入数量"
          addonAfter={symbol.split('/')[0]}
        />
      </Form.Item>

      <Space className="mb-4">
        <Button size="small" onClick={() => handleQuickAmount(0.25)}>
          25%
        </Button>
        <Button size="small" onClick={() => handleQuickAmount(0.5)}>
          50%
        </Button>
        <Button size="small" onClick={() => handleQuickAmount(0.75)}>
          75%
        </Button>
        <Button size="small" onClick={() => handleQuickAmount(1)}>
          100%
        </Button>
      </Space>

      <Form.Item>
        <Button
          type="primary"
          htmlType="submit"
          loading={loading}
          style={{ width: '100%' }}
        >
          {form.getFieldValue('side') === 'buy' ? '买入' : '卖出'}
          {symbol.split('/')[0]}
        </Button>
      </Form.Item>
    </Form>
  )
} 