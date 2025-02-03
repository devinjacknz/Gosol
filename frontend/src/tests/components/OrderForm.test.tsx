import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import OrderForm from '@/components/Trading/OrderForm';
import React from 'react';
import '@testing-library/jest-dom';

describe('OrderForm', () => {
  it('renders order form components', () => {
    render(<OrderForm />);
    
    expect(screen.getByTestId('order-form')).toBeInTheDocument();
    expect(screen.getByText('Place Order')).toBeInTheDocument();
    expect(screen.getByText('Order Type')).toBeInTheDocument();
    expect(screen.getByRole('combobox')).toBeInTheDocument();
  });
});
