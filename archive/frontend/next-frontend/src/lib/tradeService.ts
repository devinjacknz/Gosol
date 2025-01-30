interface OrderData {
  symbol: string;
  side: 'BUY' | 'SELL';
  quantity: number;
  price: number;
}

interface OrderResponse {
  orderId: string;
  status: string;
  message?: string;
}

export async function placeOrder(order: OrderData): Promise<OrderResponse> {
  const response = await fetch('/api/market/order', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(order),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'Failed to place order');
  }

  return response.json();
}

export async function getOrderHistory(): Promise<OrderResponse[]> {
  const response = await fetch('/api/market/orders');
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'Failed to fetch order history');
  }

  return response.json();
}

export async function cancelOrder(orderId: string): Promise<OrderResponse> {
  const response = await fetch(`/api/market/order/${orderId}`, {
    method: 'DELETE',
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'Failed to cancel order');
  }

  return response.json();
} 