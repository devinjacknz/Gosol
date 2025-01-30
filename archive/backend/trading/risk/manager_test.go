package risk

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/leonzhao/trading-system/backend/models"
)

func TestRiskManager_CanOpenPosition(t *testing.T) {
	config := RiskConfig{
		MaxPositions:    3,
		MaxPositionSize: 1000,
		MaxLeverage:     5,
		MaxDrawdown:     0.2,
		InitialCapital:  10000,
		RiskPerTrade:    0.02,
	}

	rm := NewRiskManager(config)

	// Test basic position validation
	marketData := &models.MarketData{
		TokenAddress: "token1",
		Price:       100,
		Volume24h:   10000,
		Liquidity:   50000,
		PriceImpact: 0.001,
		Timestamp:   time.Now(),
	}

	// Test valid position
	err := rm.CanOpenPosition(context.Background(), "token1", 500, marketData)
	if err != nil {
		t.Errorf("Expected no error for valid position, got %v", err)
	}

	// Test position size too large
	err = rm.CanOpenPosition(context.Background(), "token1", 2000, marketData)
	if err == nil {
		t.Error("Expected error for position size too large")
	}

	// Test max positions limit
	for i := 0; i < 3; i++ {
		pos := &models.Position{
			TokenAddress: fmt.Sprintf("token%d", i),
			Size:        100,
			EntryPrice:  100,
		}
		if err := rm.OpenPosition(context.Background(), pos, marketData); err != nil {
			t.Errorf("Failed to open position: %v", err)
		}
	}

	err = rm.CanOpenPosition(context.Background(), "token4", 100, marketData)
	if err == nil {
		t.Error("Expected error for max positions reached")
	}
}

func TestRiskManager_OpenPosition(t *testing.T) {
	config := RiskConfig{
		MaxPositions:    5,
		MaxPositionSize: 1000,
		MaxLeverage:     5,
		MaxDrawdown:     0.2,
		InitialCapital:  10000,
		RiskPerTrade:    0.02,
	}

	rm := NewRiskManager(config)

	marketData := &models.MarketData{
		TokenAddress: "token1",
		Price:       100,
		Volume24h:   10000,
		Liquidity:   50000,
		PriceImpact: 0.001,
		Timestamp:   time.Now(),
	}

	position := &models.Position{
		TokenAddress: "token1",
		Size:        500,
		Side:        models.PositionSideLong,
	}

	// Test opening position
	err := rm.OpenPosition(context.Background(), position, marketData)
	if err != nil {
		t.Errorf("Failed to open position: %v", err)
	}

	// Verify position was added
	pos := rm.GetPosition("token1")
	if pos == nil {
		t.Error("Position not found after opening")
	}

	// Verify position details
	if pos.EntryPrice != marketData.Price {
		t.Errorf("Expected entry price %f, got %f", marketData.Price, pos.EntryPrice)
	}
	if pos.Status != models.PositionStatusOpen {
		t.Errorf("Expected status %s, got %s", models.PositionStatusOpen, pos.Status)
	}
}

func TestRiskManager_ClosePosition(t *testing.T) {
	config := RiskConfig{
		MaxPositions:    5,
		MaxPositionSize: 1000,
		MaxLeverage:     5,
		MaxDrawdown:     0.2,
		InitialCapital:  10000,
		RiskPerTrade:    0.02,
	}

	rm := NewRiskManager(config)

	marketData := &models.MarketData{
		TokenAddress: "token1",
		Price:       100,
		Volume24h:   10000,
		Liquidity:   50000,
		PriceImpact: 0.001,
		Timestamp:   time.Now(),
	}

	position := &models.Position{
		TokenAddress: "token1",
		Size:        500,
		Side:        models.PositionSideLong,
	}

	// Open position
	err := rm.OpenPosition(context.Background(), position, marketData)
	if err != nil {
		t.Errorf("Failed to open position: %v", err)
	}

	// Close position with profit
	closePrice := 110.0
	err = rm.ClosePosition(context.Background(), "token1", closePrice)
	if err != nil {
		t.Errorf("Failed to close position: %v", err)
	}

	// Verify position was removed
	pos := rm.GetPosition("token1")
	if pos != nil {
		t.Error("Position still exists after closing")
	}

	// Verify portfolio stats
	stats := rm.GetStats().ToDailyStats()
	if stats.TotalTrades != 1 {
		t.Errorf("Expected 1 total trade, got %d", stats.TotalTrades)
	}
	if stats.WinningTrades != 1 {
		t.Errorf("Expected 1 winning trade, got %d", stats.WinningTrades)
	}

func TestRiskManager_UpdatePositions(t *testing.T) {
	config := RiskConfig{
		MaxPositions:    5,
		MaxPositionSize: 1000,
		MaxLeverage:     5,
		MaxDrawdown:     0.2,
		InitialCapital:  10000,
		RiskPerTrade:    0.02,
	}

	rm := NewRiskManager(config)

	// Open two positions
	positions := []*models.Position{
		{
			TokenAddress: "token1",
			Size:        500,
			Side:        models.PositionSideLong,
		},
		{
			TokenAddress: "token2",
			Size:        300,
			Side:        models.PositionSideLong,
		},
	}

	marketData := map[string]*models.MarketData{
		"token1": {
			TokenAddress: "token1",
			Price:       100,
			Volume24h:   10000,
			Liquidity:   50000,
			PriceImpact: 0.001,
			Timestamp:   time.Now(),
		},
		"token2": {
			TokenAddress: "token2",
			Price:       200,
			Volume24h:   20000,
			Liquidity:   100000,
			PriceImpact: 0.002,
			Timestamp:   time.Now(),
		},
	}

	for _, pos := range positions {
		err := rm.OpenPosition(context.Background(), pos, marketData[pos.TokenAddress])
		if err != nil {
			t.Errorf("Failed to open position: %v", err)
		}
	}

	// Update prices
	marketData["token1"].Price = 110  // Profit
	marketData["token2"].Price = 180  // Loss

	rm.UpdatePositions(context.Background(), marketData)

	// Verify position updates
	pos1 := rm.GetPosition("token1")
	if pos1.CurrentPrice != 110 {
		t.Errorf("Expected price 110, got %f", pos1.CurrentPrice)
	}
	if pos1.UnrealizedPnL <= 0 {
		t.Error("Expected positive PnL for token1")
	}

	pos2 := rm.GetPosition("token2")
	if pos2.CurrentPrice != 180 {
		t.Errorf("Expected price 180, got %f", pos2.CurrentPrice)
	}
	if pos2.UnrealizedPnL >= 0 {
		t.Error("Expected negative PnL for token2")
	}
}
