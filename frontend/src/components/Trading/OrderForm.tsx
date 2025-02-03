import React from 'react';
import { Order } from '@/pages/TradingView';

interface OrderFormProps {
  onSubmit: (order: Omit<Order, 'id' | 'status' | 'timestamp'>) => void;
  disabled?: boolean;
}

const OrderForm: React.FC<OrderFormProps> = ({ onSubmit, disabled = false }) => {
  const [orderType, setOrderType] = React.useState<'market' | 'limit'>('market');
  const [side, setSide] = React.useState<'buy' | 'sell'>('buy');
  const [size, setSize] = React.useState('');
  const [price, setPrice] = React.useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (disabled) return;
    
    onSubmit({
      type: orderType,
      side,
      symbol: 'BTC-USD',
      size: parseFloat(size),
      price: orderType === 'market' ? undefined : parseFloat(price)
    });
    setSize('');
    setPrice('');
  };

  return (
    <div className="order-form" data-testid="order-form">
      <h3>Place Order</h3>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="order-type">Order Type</label>
          <select 
            id="order-type"
            data-testid="order-type-select"
            value={orderType} 
            onChange={(e) => setOrderType(e.target.value as 'market' | 'limit')}
            disabled={disabled}
          >
            <option value="market">Market</option>
            <option value="limit">Limit</option>
          </select>
        </div>
        <div className="form-group">
          <label htmlFor="order-side">Side</label>
          <select 
            id="order-side"
            data-testid="order-side-select"
            value={side} 
            onChange={(e) => setSide(e.target.value as 'buy' | 'sell')}
            disabled={disabled}
          >
            <option value="buy">Buy</option>
            <option value="sell">Sell</option>
          </select>
        </div>
        <div className="form-group">
          <label htmlFor="order-size">Size</label>
          <input
            id="order-size"
            data-testid="order-size-input"
            type="number"
            value={size}
            onChange={(e) => setSize(e.target.value)}
            disabled={disabled}
            required
            min="0"
            step="0.0001"
          />
        </div>
        {orderType === 'limit' && (
          <div className="form-group">
            <label htmlFor="order-price">Price</label>
            <input
              id="order-price"
              type="number"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
              disabled={disabled}
              required
              min="0"
              step="0.01"
            />
          </div>
        )}
        <button 
          type="submit" 
          disabled={disabled}
          data-testid="submit-order-button"
        >
          Place Order
        </button>
      </form>
    </div>
  );
};

export default OrderForm;
