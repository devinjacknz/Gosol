package models

import "time"

type TradeSignalType = SignalType // 统一使用SignalType

const (
	SignalTypeBullish    TradeSignalType = Buy
	SignalTypeBearish    TradeSignalType = Sell
	SignalTypeOverbought TradeSignalType = 4
	SignalTypeOversold   TradeSignalType = 5
	SignalTypeBuy        TradeSignalType = Buy
	SignalTypeSell       TradeSignalType = Sell
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
