import { Schema, model } from 'mongoose';

const accountSchema = new Schema({
  userId: { type: Schema.Types.ObjectId, ref: 'User', required: true },
  // 现货资产
  spotAssets: [{
    asset: { type: String, required: true },
    total: { type: Number, default: 0 },
    available: { type: Number, default: 0 },
    frozen: { type: Number, default: 0 },
    value: { type: Number, default: 0 }, // USD价值
    updatedAt: { type: Date, default: Date.now }
  }],
  // 合约资产
  futuresAssets: [{
    symbol: { type: String, required: true },
    margin: { type: Number, default: 0 },
    marginType: { type: String, enum: ['isolated', 'cross'], default: 'isolated' },
    leverage: { type: Number, default: 1 },
    position: { type: Number, default: 0 }, // 正数为多头,负数为空头
    entryPrice: { type: Number },
    liquidationPrice: { type: Number },
    unrealizedPnL: { type: Number, default: 0 },
    realizedPnL: { type: Number, default: 0 },
    updatedAt: { type: Date, default: Date.now }
  }],
  // 账户汇总
  summary: {
    totalEquity: { type: Number, default: 0 }, // 总权益
    availableBalance: { type: Number, default: 0 }, // 可用余额
    usedMargin: { type: Number, default: 0 }, // 已用保证金
    marginLevel: { type: Number, default: 0 }, // 保证金率
    unrealizedPnL: { type: Number, default: 0 }, // 未实现盈亏
    realizedPnL: { type: Number, default: 0 }, // 已实现盈亏
    updatedAt: { type: Date, default: Date.now }
  }
}, {
  timestamps: true
});

// 索引
accountSchema.index({ userId: 1 });
accountSchema.index({ 'spotAssets.asset': 1 });
accountSchema.index({ 'futuresAssets.symbol': 1 });

export const Account = model('Account', accountSchema); 