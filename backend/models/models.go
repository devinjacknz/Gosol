package models

import (
	"time"
)

// TradingState represents the current trading state
type TradingState struct {
	IsTrading        bool         `json:"is_trading"`
	LastTrade        time.Time    `json:"last_trade"`
	CurrentProfit    float64      `json:"current_profit"`
	TotalTrades      int          `json:"total_trades"`
	SuccessfulTrades int          `json:"successful_trades"`
	MarketData       *MarketData  `json:"market_data"`
}

// Position represents a trading position
type Position struct {
	ID            string    `json:"id"`
	TokenAddress  string    `json:"token_address"`
	WalletAddress string    `json:"wallet_address"`
	Side          string    `json:"side"`  // See PositionSide* constants
	Size          float64   `json:"size"`
	EntryPrice    float64   `json:"entry_price"`
	CurrentPrice  float64   `json:"current_price"`
	StopLoss      float64   `json:"stop_loss"`
	TakeProfit    float64   `json:"take_profit"`
	UnrealizedPnL float64   `json:"unrealized_pnl"`
	RealizedPnL   float64   `json:"realized_pnl"`
	OpenTime      time.Time `json:"open_time"`
	CloseTime     *time.Time `json:"close_time,omitempty"`
	Status        string    `json:"status"` // See PositionStatus* constants
	TxHash        string    `json:"tx_hash"`
	Commission    float64   `json:"commission"`
	Leverage      float64   `json:"leverage"`
	LiqPrice      float64   `json:"liq_price"`
	UpdateTime    time.Time `json:"update_time"`
}

// Trade represents a trade execution
type Trade struct {
	ID            string    `json:"id"`
	TokenAddress  string    `json:"token_address"`
	WalletAddress string    `json:"wallet_address"`
	Type          string    `json:"type"`  // See OrderType* constants
	Side          string    `json:"side"`  // See OrderSide* constants
	Amount        float64   `json:"amount"`
	Price         float64   `json:"price"`
	Status        string    `json:"status"` // See TradeStatus* constants
	TxHash        string    `json:"tx_hash"`
	Commission    float64   `json:"commission"`
	Timestamp     time.Time `json:"timestamp"`
	UpdateTime    time.Time `json:"update_time"`
}

// DailyStats tracks daily trading statistics
type DailyStats struct {
	Date               time.Time `json:"date"`
	TotalTrades        int       `json:"total_trades"`
	WinningTrades      int       `json:"winning_trades"`
	LosingTrades       int       `json:"losing_trades"`
	RealizedPnL        float64   `json:"realized_pnl"`
	MaxDrawdown        float64   `json:"max_drawdown"`
	HighWaterMark      float64   `json:"high_water_mark"`
	Volume             float64   `json:"volume"`
	Commissions        float64   `json:"commissions"`
	StartBalance       float64   `json:"start_balance"`
	EndBalance         float64   `json:"end_balance"`
	LargestWin         float64   `json:"largest_win"`
	LargestLoss        float64   `json:"largest_loss"`
	AverageWin         float64   `json:"average_win"`
	AverageLoss        float64   `json:"average_loss"`
	WinRate            float64   `json:"win_rate"`
	ProfitFactor       float64   `json:"profit_factor"`
	SharpeRatio        float64   `json:"sharpe_ratio"`
	MaxConsecWins      int       `json:"max_consec_wins"`
	MaxConsecLosses    int       `json:"max_consec_losses"`
	CurrentConsecWins  int       `json:"current_consec_wins"`
	CurrentConsecLosses int      `json:"current_consec_losses"`
}

// TradeSignalMessage represents a trade signal message
type TradeSignalMessage struct {
	TokenAddress string    `json:"token_address"`
	Action       string    `json:"action"` // "buy", "sell", "hold"
	TargetPrice  float64   `json:"target_price"`
	Amount       float64   `json:"amount"`
	Confidence   float64   `json:"confidence"` // 0-1
	Reason       string    `json:"reason"`
	Timestamp    time.Time `json:"timestamp"`
}

// MarketData represents market data for a token
type MarketData struct {
	TokenAddress  string    `json:"token_address"`
	Price         float64   `json:"price"`
	Volume24h     float64   `json:"volume_24h"`
	MarketCap     float64   `json:"market_cap"`
	Liquidity     float64   `json:"liquidity"`
	PriceImpact   float64   `json:"price_impact"`
	Timestamp     time.Time `json:"timestamp"`
}

// TechnicalIndicators represents technical analysis indicators
type TechnicalIndicators struct {
	RSI        float64 `json:"rsi"`
	MACD       float64 `json:"macd"`
	Signal     float64 `json:"signal"`
	BBUpper    float64 `json:"bb_upper"`
	BBLower    float64 `json:"bb_lower"`
	EMA20      float64 `json:"ema_20"`
	Volume     float64 `json:"volume"`
	Volatility float64 `json:"volatility"`
}

// DeepseekAnalysis represents AI-powered market analysis
type DeepseekAnalysis struct {
	Sentiment  string `json:"sentiment"`
	Confidence float64 `json:"confidence"`
	KeyFactors []string `json:"key_factors"`
	RiskAnalysis struct {
		ManipulationRisk string `json:"manipulation_risk"`
		LiquidityRisk    string `json:"liquidity_risk"`
		VolatilityRisk   string `json:"volatility_risk"`
	} `json:"risk_analysis"`
	Recommendation struct {
		Action      string    `json:"action"`
		EntryPoints []float64 `json:"entry_points"`
		ExitPoints  []float64 `json:"exit_points"`
		StopLoss    float64   `json:"stop_loss"`
	} `json:"recommendation"`
}

// AnalysisResult represents a complete market analysis result
type AnalysisResult struct {
	TokenAddress        string    `json:"token_address"`
	Prediction         float64   `json:"prediction"`
	Confidence         float64   `json:"confidence"`
	Sentiment         string    `json:"sentiment"`
	RiskLevel         string    `json:"risk_level"`
	TechnicalIndicators []byte    `json:"technical_indicators"` // JSON encoded
	DeepseekAnalysis   []byte    `json:"deepseek_analysis"`    // JSON encoded
	Timestamp         time.Time `json:"timestamp"`
}

// Position methods

func (p *Position) CalculatePnL() float64 {
	if p.Side == PositionSideLong {
		return p.Size * (p.CurrentPrice - p.EntryPrice)
	}
	return p.Size * (p.EntryPrice - p.CurrentPrice)
}

func (p *Position) UpdatePrice(price float64) {
	p.CurrentPrice = price
	p.UnrealizedPnL = p.CalculatePnL()
	p.UpdateTime = time.Now()
}

func (p *Position) Close(closePrice float64, closeTime time.Time) {
	p.CurrentPrice = closePrice
	p.RealizedPnL = p.CalculatePnL()
	p.Status = PositionStatusClosed
	p.CloseTime = &closeTime
	p.UpdateTime = closeTime
}

func (p *Position) IsActive() bool {
	return p.Status == PositionStatusOpen
}

func (p *Position) Duration() time.Duration {
	if p.CloseTime != nil {
		return p.CloseTime.Sub(p.OpenTime)
	}
	return time.Since(p.OpenTime)
}

func (p *Position) ROI() float64 {
	investment := p.Size * p.EntryPrice
	if investment == 0 {
		return 0
	}
	
	var pnl float64
	if p.Status == PositionStatusClosed {
		pnl = p.RealizedPnL
	} else {
		pnl = p.UnrealizedPnL
	}
	
	return (pnl / investment) * 100
}

func (p *Position) Value() float64 {
	return p.Size * p.CurrentPrice
}

func (p *Position) Cost() float64 {
	return p.Size * p.EntryPrice
}

func (p *Position) NetPnL() float64 {
	if p.Status == PositionStatusClosed {
		return p.RealizedPnL - p.Commission
	}
	return p.UnrealizedPnL - p.Commission
}

func (p *Position) UpdateLiquidationPrice() {
	if p.Leverage <= 1 {
		p.LiqPrice = 0
		return
	}

	maintenanceMargin := 0.025 // 2.5% maintenance margin requirement
	if p.Side == PositionSideLong {
		p.LiqPrice = p.EntryPrice * (1 - 1/p.Leverage + maintenanceMargin)
	} else {
		p.LiqPrice = p.EntryPrice * (1 + 1/p.Leverage - maintenanceMargin)
	}
}

func (p *Position) ShouldLiquidate() bool {
	if p.LiqPrice == 0 {
		return false
	}

	if p.Side == PositionSideLong {
		return p.CurrentPrice <= p.LiqPrice
	}
	return p.CurrentPrice >= p.LiqPrice
}
