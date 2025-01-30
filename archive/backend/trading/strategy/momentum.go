package strategy

import (
	"context"
	"math"

	"github.com/leonzhao/trading-system/backend/models"
	"github.com/leonzhao/trading-system/backend/trading/analysis"
)

type MomentumStrategy struct {
	RSIPeriod      int
	MACDFastPeriod int
	MACDSlowPeriod int
	SignalPeriod   int
	RiskLevel      float64
}

func NewMomentumStrategy(rsiPeriod, macdFast, macdSlow, signal int, risk float64) *MomentumStrategy {
	return &MomentumStrategy{
		RSIPeriod:      rsiPeriod,
		MACDFastPeriod: macdFast,
		MACDSlowPeriod: macdSlow,
		SignalPeriod:   signal,
		RiskLevel:      risk,
	}
}

func (s *MomentumStrategy) GenerateSignals(ctx context.Context, marketData []models.MarketData) []models.TradeSignal {
	var signals []models.TradeSignal

	prices := extractClosingPrices(marketData)
	rsi := analysis.RSI(prices, s.RSIPeriod)
	macdLine, signalLine, _ := analysis.MACD(prices, s.MACDFastPeriod, s.MACDSlowPeriod, s.SignalPeriod)

	for i := 1; i < len(prices); i++ {
		if i >= len(rsi) || i >= len(macdLine) {
			continue
		}

		var signalType models.SignalType
		currentPrice := prices[i]

		// RSI超卖且MACD金叉
		if rsi[i-1] < 30 && macdLine[i] > signalLine[i] && macdLine[i-1] <= signalLine[i-1] {
			signalType = models.Buy
		// RSI超买且MACD死叉
		} else if rsi[i-1] > 70 && macdLine[i] < signalLine[i] && macdLine[i-1] >= signalLine[i-1] {
			signalType = models.Sell
		} else {
			continue
		}

		positionSize := s.calculatePositionSize(currentPrice)
		if positionSize <= 0 {
			continue
		}

		signals = append(signals, models.TradeSignal{
			Symbol:      marketData[i].Symbol,
			SignalType:  signalType,
			Price:       currentPrice,
			Size:        positionSize,
			Timestamp:   marketData[i].Timestamp,
			Confidence:  s.calculateConfidence(rsi[i-1], macdLine[i]-signalLine[i]),
			Description: "Momentum based trading signal",
		})
	}

	return signals
}

func extractClosingPrices(marketData []models.MarketData) []float64 {
	prices := make([]float64, len(marketData))
	for i, data := range marketData {
		prices[i] = data.ClosePrice
	}
	return prices
}

func (s *MomentumStrategy) calculatePositionSize(price float64) float64 {
	if price == 0 {
		return 0
	}
	return math.Round(s.RiskLevel*100/price*100) / 100
}

func (s *MomentumStrategy) calculateConfidence(rsi, macdDiff float64) float64 {
	rsiConf := 1 - math.Abs(rsi-50)/50
	macdConf := math.Abs(macdDiff) / 0.1
	return math.Min(0.9, (rsiConf+macdConf)/2)
}
