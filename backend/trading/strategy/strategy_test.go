package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/leonzhao/demo/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestMomentumStrategy_GenerateSignals(t *testing.T) {
	// 生成测试数据
	testData := []models.MarketData{
		{Symbol: "BTC/USDT", ClosePrice: 45000, Timestamp: time.Now().Add(-24 * time.Hour)},
		{Symbol: "BTC/USDT", ClosePrice: 46000, Timestamp: time.Now().Add(-23 * time.Hour)},
		{Symbol: "BTC/USDT", ClosePrice: 47000, Timestamp: time.Now().Add(-22 * time.Hour)},
		{Symbol: "BTC/USDT", ClosePrice: 48000, Timestamp: time.Now().Add(-21 * time.Hour)},
		{Symbol: "BTC/USDT", ClosePrice: 49000, Timestamp: time.Now().Add(-20 * time.Hour)},
	}

	strategy := NewMomentumStrategy(14, 12, 26, 9, 1000)
	signals := strategy.GenerateSignals(context.Background(), testData)

	assert.NotEmpty(t, signals, "应该生成交易信号")
	
	for _, signal := range signals {
		assert.Contains(t, []models.SignalType{models.Buy, models.Sell}, signal.SignalType)
		assert.Greater(t, signal.Price, 0.0)
		assert.Greater(t, signal.Size, 0.0)
		assert.InDelta(t, 0.5, signal.Confidence, 0.4)
		assert.NotEmpty(t, signal.Description)
	}
}
