import { Schema, model } from 'mongoose';

const transactionSchema = new Schema({
  userId: { type: Schema.Types.ObjectId, ref: 'User', required: true },
  type: { 
    type: String, 
    enum: ['deposit', 'withdrawal', 'transfer', 'fee', 'commission'], 
    required: true 
  },
  asset: { type: String, required: true },
  amount: { type: Number, required: true },
  fee: { type: Number, default: 0 },
  feeAsset: { type: String },
  network: { type: String }, // 区块链网络
  address: { type: String }, // 充提地址
  txHash: { type: String }, // 交易哈希
  status: { 
    type: String, 
    enum: ['pending', 'processing', 'completed', 'failed', 'cancelled'],
    default: 'pending'
  },
  statusHistory: [{
    status: String,
    timestamp: Date,
    reason: String
  }],
  // 内部转账相关
  fromAccount: { type: String }, // spot, futures, funding等
  toAccount: { type: String },
  // 关联订单
  orderId: { type: Schema.Types.ObjectId, ref: 'Order' },
  // 备注
  remarks: { type: String }
}, {
  timestamps: true
});

// 索引
transactionSchema.index({ userId: 1 });
transactionSchema.index({ type: 1 });
transactionSchema.index({ asset: 1 });
transactionSchema.index({ status: 1 });
transactionSchema.index({ createdAt: -1 });
transactionSchema.index({ txHash: 1 }, { sparse: true });

export const Transaction = model('Transaction', transactionSchema); 