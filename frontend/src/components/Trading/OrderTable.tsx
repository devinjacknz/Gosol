import { Table } from 'antd';
import { useAppSelector } from '@/hooks/store';

export const OrderTable: React.FC = () => {
  const { orders, loading } = useAppSelector(state => state.trading);

  const columns = [
    { title: 'Symbol', dataIndex: 'symbol', key: 'symbol' },
    { title: 'Type', dataIndex: 'type', key: 'type' },
    { title: 'Side', dataIndex: 'side', key: 'side' },
    { title: 'Amount', dataIndex: 'amount', key: 'amount' },
    { title: 'Status', dataIndex: 'status', key: 'status' },
  ];

  return <Table dataSource={orders} columns={columns} loading={loading} />;
};
