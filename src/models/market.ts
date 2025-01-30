import { Schema, model } from 'mongoose';

// 交易对信息
const symbolSchema = new Schema({
  symbol: { type: String, required: true, unique: true },
  baseAsset: { type: String, required: true },
  quoteAsset: { type: String, required: true },
  type: { type: String, enum: ['spot', 'futures'], required: true },
  status: { type: String, enum: ['trading', 'halt', 'delisted'], default: 'trading' },
  // 价格精度和数量精度
  pricePrecision: { type: Number, required: true },
  quantityPrecision: { type: Number, required: true },
  // 交易规则
  minQuantity: { type: Number, required: true },
  maxQuantity: { type: Number },
  minPrice: { type: Number },
  maxPrice: { type: Number },
  // 合约特有字段
  contractSize: { type: Number }, // 合约面值
  maxLeverage: { type: Number }, // 最大杠杆
  maintenanceMarginRate: { type: Number }, // 维持保证金率
  liquidationFeeRate: { type: Number }, // 强平手续费率
  // 手续费
  makerFeeRate: { type: Number, required: true },
  takerFeeRate: { type: Number, required: true },
}, {
  timestamps: true
});

// 市场行情
const tickerSchema = new Schema({
  symbol: { type: String, required: true },
  price: { type: Number, required: true },
  priceChange: { type: Number, default: 0 },
  priceChangePercent: { type: Number, default: 0 },
  high24h: { type: Number },
  low24h: { type: Number },
  volume24h: { type: Number, default: 0 },
  quoteVolume24h: { type: Number, default: 0 },
  openPrice: { type: Number },
  closePrice: { type: Number },
  // 合约特有字段
  openInterest: { type: Number }, // 未平仓合约数量
  fundingRate: { type: Number }, // 资金费率
  nextFundingTime: { type: Date }, // 下次结算时间
  // 最新更新时间
  updatedAt: { type: Date, default: Date.now }
}, {
  timestamps: true
});

// 深度数据
const orderBookSchema = new Schema({
  symbol: { type: String, required: true },
  bids: [{
    price: Number,
    quantity: Number
  }],
  asks: [{
    price: Number,
    quantity: Number
  }],
  lastUpdateId: { type: Number, required: true },
  updatedAt: { type: Date, default: Date.now }
});

// 最新成交
const tradeSchema = new Schema({
  symbol: { type: String, required: true },
  price: { type: Number, required: true },
  quantity: { type: Number, required: true },
  side: { type: String, enum: ['buy', 'sell'], required: true },
  time: { type: Date, default: Date.now }
});

// 索引
symbolSchema.index({ symbol: 1 }, { unique: true });
symbolSchema.index({ type: 1 });
symbolSchema.index({ status: 1 });

tickerSchema.index({ symbol: 1 }, { unique: true });
tickerSchema.index({ updatedAt: 1 });

orderBookSchema.index({ symbol: 1 }, { unique: true });
orderBookSchema.index({ updatedAt: 1 });

tradeSchema.index({ symbol: 1 });
tradeSchema.index({ time: -1 });

// 导出模型
export const Symbol = model('Symbol', symbolSchema);
export const Ticker = model('Ticker', tickerSchema);
export const OrderBook = model('OrderBook', orderBookSchema);
export const Trade = model('Trade', tradeSchema); 