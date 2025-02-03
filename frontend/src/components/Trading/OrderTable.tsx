import React from 'react';

interface Order {
  id: string;
  type: 'market' | 'limit';
  side: 'buy' | 'sell';
  symbol: string;
  size: number;
  price?: number;
  status: 'open' | 'filled' | 'cancelled';
  timestamp: number;
}

interface OrderTableProps {
  orders: Order[];
  onCancelOrder?: (orderId: string) => void;
  loading?: boolean;
  'data-testid'?: string;
}

const OrderTable: React.FC<OrderTableProps> = ({ orders, onCancelOrder, loading = false, 'data-testid': testId = 'order-table' }) => {
  if (loading) {
    return (
      <div className="order-table" data-testid={testId}>
        <h3>Orders</h3>
        <div className="loading" data-testid="loading-message">Loading orders...</div>
      </div>
    );
  }
  return (
    <div className="order-table" data-testid={testId}>
      <h3>Orders</h3>
      <table>
        <thead>
          <tr>
            <th>Symbol</th>
            <th>Type</th>
            <th>Side</th>
            <th>Size</th>
            <th>Price</th>
            <th>Status</th>
            {onCancelOrder && <th>Action</th>}
          </tr>
        </thead>
        <tbody>
          {orders.map((order, index) => (
            <tr key={order.id} data-testid={`order-row-${index}`}>
              <td>{order.symbol}</td>
              <td>{order.type}</td>
              <td>{order.side}</td>
              <td>{order.size}</td>
              <td>{order.price || 'Market'}</td>
              <td>{order.status}</td>
              {onCancelOrder && order.status === 'open' && (
                <td>
                  <button 
                    onClick={() => onCancelOrder(order.id)}
                    disabled={loading}
                    data-testid={`cancel-order-${order.id}`}
                  >
                    取消
                  </button>
                </td>
              )}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default OrderTable;
