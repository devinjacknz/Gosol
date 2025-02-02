import { Form, Input, Button, Select } from 'antd';
import { useAppDispatch, useAppSelector } from '@/hooks/store';
import { placeOrder } from '@/store/trading/tradingSlice';

export const OrderForm: React.FC = () => {
  const dispatch = useAppDispatch();
  const { selectedSymbol } = useAppSelector(state => state.trading);

  const onFinish = (values: any) => {
    dispatch(placeOrder({ ...values, symbol: selectedSymbol }));
  };

  return (
    <Form onFinish={onFinish}>
      <Form.Item name="type" rules={[{ required: true }]}>
        <Select>
          <Select.Option value="market">Market</Select.Option>
          <Select.Option value="limit">Limit</Select.Option>
        </Select>
      </Form.Item>
      <Form.Item name="side" rules={[{ required: true }]}>
        <Select>
          <Select.Option value="buy">Buy</Select.Option>
          <Select.Option value="sell">Sell</Select.Option>
        </Select>
      </Form.Item>
      <Form.Item name="amount" rules={[{ required: true }]}>
        <Input type="number" placeholder="Amount" />
      </Form.Item>
      <Form.Item>
        <Button type="primary" htmlType="submit">Place Order</Button>
      </Form.Item>
    </Form>
  );
};
