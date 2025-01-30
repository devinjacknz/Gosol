import { Schema, model } from 'mongoose';

const userSchema = new Schema({
  email: { type: String, required: true, unique: true },
  password: { type: String, required: true },
  nickname: { type: String },
  avatar: { type: String },
  isVerified: { type: Boolean, default: false },
  kycLevel: { type: Number, default: 0 },
  twoFactorEnabled: { type: Boolean, default: false },
  twoFactorSecret: { type: String },
  apiKeys: [{
    key: String,
    secret: String,
    name: String,
    permissions: [String],
    createdAt: Date,
    lastUsedAt: Date
  }],
  preferences: {
    theme: { type: String, default: 'light' },
    language: { type: String, default: 'zh-CN' },
    notifications: {
      email: { type: Boolean, default: true },
      browser: { type: Boolean, default: true },
      mobile: { type: Boolean, default: true }
    }
  },
  createdAt: { type: Date, default: Date.now },
  updatedAt: { type: Date, default: Date.now }
}, {
  timestamps: true
});

export const User = model('User', userSchema); 