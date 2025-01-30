import { Schema, model } from 'mongoose';

const orderSchema = new Schema({
  userId: { type: Schema.Types.ObjectId, ref: 'User', required: true },
  symbol: { type: String, required: true },
  type: { type: String, enum: ['spot', 'futures'], required: true },
  // 订单类型
  orderType: { 
    type: String, 
    enum: ['limit', 'market', 'stopLimit', 'stopMarket', 'takeProfitLimit', 'takeProfitMarket'], 
    required: true 
  },
  side: { type: String, enum: ['buy', 'sell'], required: true },
  // 价格和数量
  price: { type: Number },
  stopPrice: { type: Number },
  quantity: { type: Number, required: true },
  executedQuantity: { type: Number, default: 0 },
  remainingQuantity: { type: Number },
  // 订单状态
  status: { 
    type: String, 
    enum: ['new', 'partiallyFilled', 'filled', 'cancelled', 'rejected', 'expired'],
    default: 'new'
  },
  // 成交明细
  fills: [{
    price: Number,
    quantity: Number,
    fee: Number,
    feeAsset: String,
    timestamp: Date
  }],
  // 合约特有字段
  marginType: { type: String, enum: ['isolated', 'cross'] },
  leverage: { type: Number },
  // 订单选项
  timeInForce: { type: String, enum: ['GTC', 'IOC', 'FOK'], default: 'GTC' },
  postOnly: { type: Boolean, default: false },
  reduceOnly: { type: Boolean, default: false },
  // 手续费
  makerFeeRate: { type: Number },
  takerFeeRate: { type: Number },
  totalFee: { type: Number, default: 0 },
  // 订单来源
  source: { type: String, enum: ['web', 'app', 'api'], required: true },
  clientOrderId: { type: String },
  // 错误信息
  error: {
    code: String,
    message: String
  },
  // 更新历史
  statusHistory: [{
    status: String,
    timestamp: Date,
    reason: String
  }]
}, {
  timestamps: true
});

// 索引
orderSchema.index({ userId: 1 });
orderSchema.index({ symbol: 1 });
orderSchema.index({ type: 1 });
orderSchema.index({ status: 1 });
orderSchema.index({ createdAt: -1 });
orderSchema.index({ clientOrderId: 1 }, { sparse: true });

export const Order = model('Order', orderSchema); 