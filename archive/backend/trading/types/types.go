package types

type RiskConfig struct {
	MaxPositionSize  float64
	DailyLossLimit   float64
	LeverageLimit    int
	AllowedAssets    []string
	VolatilityWindow int
}

type RiskState struct {
	DailyStats map[string]*DailyStats
}

type DailyStats struct {
	ProfitLoss   float64
	TradesCount  int
	VolumeTraded float64
}

type TradeExecutor interface {
	ExecuteOrder(order interface{}) error
}

type Position struct {
	Asset     string
	Size      float64
	EntryPrice float64
	Timestamp int64
}
