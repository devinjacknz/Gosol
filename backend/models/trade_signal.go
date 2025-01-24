package models

import "time"

type TradeSignalType string

const (
	SignalTypeBullish    TradeSignalType = "BULLISH"
	SignalTypeBearish    TradeSignalType = "BEARISH"
	SignalTypeOverbought TradeSignalType = "OVERBOUGHT"
	SignalTypeOversold   TradeSignalType = "OVERSOLD"
	SignalTypeBuy        TradeSignalType = "BUY"
	SignalTypeSell       TradeSignalType = "SELL"
)

type SignalStrength string

const (
	SignalStrengthWeak   SignalStrength = "WEAK"
	SignalStrengthMedium SignalStrength = "MEDIUM"
	SignalStrengthStrong SignalStrength = "STRONG"
)

type TradeSignal struct {
	Symbol        string
	SignalType    TradeSignalType
	Price         float64
	Size          float64
	Timestamp     time.Time
	Strength      SignalStrength
	Confidence    float64
	Description   string
	IndicatorType string
}
