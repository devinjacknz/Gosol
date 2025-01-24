import { rest } from 'msw';

const mockToken = 'So11111111111111111111111111111111111111111';

export const handlers = [
  // Token info
  rest.get('*/api/token/*', (req, res, ctx) => {
    const { tokenAddress } = req.params;
    if (tokenAddress === mockToken) {
      return res(
        ctx.json({
          address: mockToken,
          name: 'Test Token',
          symbol: 'TEST',
          decimals: 9
        })
      );
    }
    return res(
      ctx.status(404),
      ctx.json({ error: 'Token not found' })
    );
  }),

  // Trading status
  rest.get('*/api/status', (req, res, ctx) => {
    return res(
      ctx.json({
        is_trading: false,
        last_update: '2024-03-15T10:30:00Z'
      })
    );
  }),

  // Trading toggle
  rest.post('*/api/trading/toggle', (req, res, ctx) => {
    return res(
      ctx.json({
        success: true,
        is_trading: true
      })
    );
  }),

  // Profit transfer
  rest.post('*/api/trading/transfer', (req, res, ctx) => {
    return res(
      ctx.json({
        success: true,
        message: 'Profit transferred successfully'
      })
    );
  }),

  // Market data
  rest.get('*/api/market-data/*', (req, res, ctx) => {
    return res(
      ctx.json({
        price: 1.5,
        volume_24h: 1000000,
        market_cap: 10000000,
        liquidity: 500000,
        price_change_24h: 5.5
      })
    );
  }),

  // Analysis
  rest.get('*/api/analysis', (req, res, ctx) => {
    return res(
      ctx.json({
        sentiment: 'bullish',
        risk_level: 'medium',
        confidence: 75,
        deepseek_analysis: JSON.stringify({
          recommendation: {
            action: 'buy',
            confidence: 0.85
          },
          analysis: {
            technical: 'Strong upward trend',
            fundamental: 'Good tokenomics'
          }
        })
      })
    );
  }),

  // Wallet balance
  rest.get('*/api/wallet/balance', (req, res, ctx) => {
    return res(
      ctx.json({
        balance: 1000
      })
    );
  }),

  // Risk settings
  rest.get('*/api/risk-settings/*', (req, res, ctx) => {
    return res(
      ctx.json({
        maxPositionSize: 1000,
        maxDrawdown: 20,
        maxDailyTrades: 10,
        maxLeverage: 3,
        stopLossPercent: 5,
        takeProfitPercent: 10,
        riskPerTrade: 2,
        timeoutAfterLoss: 300
      })
    );
  }),

  // Trade history
  rest.get('*/api/trades/*', (req, res, ctx) => {
    return res(
      ctx.json([
        {
          id: '1',
          type: 'buy',
          amount: 100,
          price: 1.5,
          timestamp: '2024-03-15T10:30:00Z',
          profit: 0,
          status: 'completed'
        },
        {
          id: '2',
          type: 'sell',
          amount: 50,
          price: 1.8,
          timestamp: '2024-03-15T11:00:00Z',
          profit: 15,
          status: 'completed'
        },
        {
          id: '3',
          type: 'buy',
          amount: 75,
          price: 1.6,
          timestamp: '2024-03-15T11:30:00Z',
          profit: 0,
          status: 'pending'
        }
      ])
    );
  })
]; 