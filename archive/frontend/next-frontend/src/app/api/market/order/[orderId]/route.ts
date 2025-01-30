import { NextResponse } from 'next/server';

export async function DELETE(
  request: Request,
  { params }: { params: { orderId: string } }
) {
  try {
    const response = await fetch(
      `http://localhost:8080/api/market/order/${params.orderId}`,
      {
        method: 'DELETE',
      }
    );
    
    if (!response.ok) {
      const error = await response.json();
      return NextResponse.json(
        { error: error.message || 'Failed to cancel order' },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('Failed to cancel order:', error);
    return NextResponse.json(
      { error: 'Failed to cancel order' },
      { status: 500 }
    );
  }
} 